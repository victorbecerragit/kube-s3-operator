/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	s3v1alpha1 "github.com/victorbecerragit/kube-s3-operator/code/api/v1alpha1"
	"github.com/victorbecerragit/kube-s3-operator/code/internal/s3client"
)

const (
	configMapName     = "%s-s3-cm"
	s3BucketFinalizer = "s3bucket.s3.acme.io/finalizer"
)

type S3BucketReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	S3Manager s3client.BucketManager
}

// +kubebuilder:rbac:groups=s3.acme.io,resources=s3buckets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=s3.acme.io,resources=s3buckets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=s3.acme.io,resources=s3buckets/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// subreconciler is a function that handles one concern of the reconcile loop.
// Returning a non-zero Result or a non-nil error stops the chain.
type subreconciler func(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket) (ctrl.Result, error)

func (r *S3BucketReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Reconciling S3Bucket", "NamespacedName", req.NamespacedName)

	s3bkt := &s3v1alpha1.S3Bucket{}
	if err := r.Get(ctx, req.NamespacedName, s3bkt); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	for _, sub := range []subreconciler{
		r.handleFinalizer,
		r.reconcileBucket,
	} {
		if result, err := sub(ctx, s3bkt); err != nil || !result.IsZero() {
			return result, err
		}
	}
	return ctrl.Result{}, nil
}

// handleFinalizer manages finalizer registration and deletion teardown.
func (r *S3BucketReconciler) handleFinalizer(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	if !s3bkt.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(s3bkt, s3BucketFinalizer) {
			return r.DeleteResource(ctx, s3bkt)
		}
		return ctrl.Result{}, nil
	}

	if !controllerutil.ContainsFinalizer(s3bkt, s3BucketFinalizer) {
		log.Info("Adding finalizer to S3Bucket", "BucketName", s3bkt.Spec.Name)
		if err := r.addFinalizer(ctx, s3bkt); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, nil
}

// reconcileBucket drives the bucket state machine.
func (r *S3BucketReconciler) reconcileBucket(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	switch s3bkt.Status.State {
	case "":
		return r.CreateResource(ctx, s3bkt)

	case s3v1alpha1.CREATED_STATE:
		if err := r.S3Manager.ApplyLifecycleConfiguration(ctx, s3bkt); err != nil {
			log.Error(err, "Failed to re-apply lifecycle configuration")
			if err2 := r.updateBucketStatus(ctx, s3bkt, s3v1alpha1.ERROR_STATE); err2 != nil {
				log.Error(err2, "Failed to update status to ERROR")
			}
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil

	case s3v1alpha1.ERROR_STATE:
		log.Info("S3 bucket is in ERROR state", "BucketName", s3bkt.Spec.Name)
		return ctrl.Result{RequeueAfter: time.Minute * 5}, nil

	case s3v1alpha1.CREATING_STATE:
		log.Info("S3 bucket in transitional state", "BucketName", s3bkt.Spec.Name, "State", s3bkt.Status.State)
		ready, err := r.S3Manager.IsBucketReady(ctx, s3bkt)
		if err != nil {
			if err2 := r.updateBucketStatus(ctx, s3bkt, s3v1alpha1.ERROR_STATE); err2 != nil {
				log.Error(err2, "Failed to update status to ERROR")
			}
			return ctrl.Result{}, err
		}
		if ready {
			if err := r.finalizeCreation(ctx, s3bkt, s3bkt.Status.Location); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{RequeueAfter: time.Second * 10}, nil

	default:
		return ctrl.Result{}, nil
	}
}

func (r *S3BucketReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Ensure S3Manager is initialized if not provided
	if r.S3Manager == nil {
		r.S3Manager = s3client.NewManager()
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&s3v1alpha1.S3Bucket{}).
		Named("s3bucket").
		Complete(r)
}

func (r *S3BucketReconciler) CreateResource(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Starting creation of S3 Bucket", "BucketName", s3bkt.Spec.Name)

	if err := r.updateBucketStatus(ctx, s3bkt, s3v1alpha1.CREATING_STATE); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to update status to CREATING: %w", err)
	}

	location, err := r.S3Manager.CreateBucket(ctx, s3bkt)
	if err != nil {
		r.updateBucketStatus(ctx, s3bkt, s3v1alpha1.ERROR_STATE)
		return ctrl.Result{}, fmt.Errorf("failed to create S3 bucket: %w", err)
	}
	log.Info("S3 bucket API call succeeded", "BucketName", s3bkt.Spec.Name, "Location", location)

	ready, err := r.S3Manager.IsBucketReady(ctx, s3bkt)
	if err != nil {
		r.updateBucketStatus(ctx, s3bkt, s3v1alpha1.ERROR_STATE)
		return ctrl.Result{}, err
	}

	// If not ready immediately, persist the location and requeue
	if !ready {
		log.Info("Bucket not ready yet, requeuing", "BucketName", s3bkt.Spec.Name)
		if location != "" {
			if err := r.persistLocation(ctx, s3bkt, location); err != nil {
				log.Error(err, "Failed to persist bucket location")
			}
		}
		return ctrl.Result{RequeueAfter: time.Second * 5}, nil
	}

	// Ready immediately
	if err := r.finalizeCreation(ctx, s3bkt, location); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *S3BucketReconciler) finalizeCreation(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket, location string) error {
	log := logf.FromContext(ctx)

	if err := r.S3Manager.ApplyLifecycleConfiguration(ctx, s3bkt); err != nil {
		r.updateBucketStatus(ctx, s3bkt, s3v1alpha1.ERROR_STATE)
		return fmt.Errorf("failed to apply lifecycle configuration: %w", err)
	}

	if err := r.createOrUpdateBucketConfigMap(ctx, s3bkt, location); err != nil {
		r.updateBucketStatus(ctx, s3bkt, s3v1alpha1.ERROR_STATE)
		return fmt.Errorf("failed to create ConfigMap: %w", err)
	}

	if err := r.updateBucketStatusCreated(ctx, s3bkt, location); err != nil {
		return fmt.Errorf("failed to update status to CREATED: %w", err)
	}

	log.Info("S3 Bucket created successfully", "BucketName", s3bkt.Spec.Name, "Location", location)
	return nil
}

func (r *S3BucketReconciler) DeleteResource(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Starting deletion of S3 Bucket", "BucketName", s3bkt.Spec.Name)

	if err := r.updateBucketStatus(ctx, s3bkt, s3v1alpha1.DELETING_STATE); err != nil {
		log.Error(err, "Failed to update status to DELETING, continuing with deletion")
	}

	if err := r.S3Manager.DeleteBucket(ctx, s3bkt); err != nil {
		r.updateBucketStatus(ctx, s3bkt, s3v1alpha1.ERROR_STATE)
		return ctrl.Result{}, fmt.Errorf("cleanup failed: %w", err)
	}

	deleted, err := r.S3Manager.IsBucketDeleted(ctx, s3bkt)
	if err != nil {
		return ctrl.Result{}, err
	}

	if !deleted {
		log.Info("Bucket not fully deleted yet, requeuing", "BucketName", s3bkt.Spec.Name)
		return ctrl.Result{RequeueAfter: time.Second * 5}, nil
	}

	// Fully deleted
	r.deleteBucketConfigMap(ctx, s3bkt)
	if err := r.removeFinalizer(ctx, s3bkt); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to remove finalizer: %w", err)
	}

	log.Info("S3 Bucket deleted successfully and finalizer removed", "BucketName", s3bkt.Spec.Name)
	return ctrl.Result{}, nil
}

func (r *S3BucketReconciler) createOrUpdateBucketConfigMap(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket, location string) error {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf(configMapName, s3bkt.Name),
			Namespace: s3bkt.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, cm, func() error {
		if cm.Data == nil {
			cm.Data = make(map[string]string)
		}
		cm.Data["BucketName"] = s3bkt.Spec.Name
		cm.Data["Region"] = s3bkt.Spec.Region
		cm.Data["Locked"] = fmt.Sprintf("%t", s3bkt.Spec.Locked)
		if location != "" {
			cm.Data["location"] = location
		}

		return controllerutil.SetControllerReference(s3bkt, cm, r.Scheme)
	})

	return err
}

func (r *S3BucketReconciler) deleteBucketConfigMap(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket) {
	cm := &corev1.ConfigMap{}
	cmName := types.NamespacedName{
		Name:      fmt.Sprintf(configMapName, s3bkt.Name),
		Namespace: s3bkt.Namespace,
	}

	if err := r.Get(ctx, cmName, cm); err == nil {
		_ = r.Delete(ctx, cm)
	}
}

func (r *S3BucketReconciler) updateBucketStatus(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket, state string) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		latest := &s3v1alpha1.S3Bucket{}
		if err := r.Get(ctx, types.NamespacedName{
			Name:      s3bkt.Name,
			Namespace: s3bkt.Namespace,
		}, latest); err != nil {
			return err
		}

		latest.Status.State = state
		return r.Status().Update(ctx, latest)
	})
}

func (r *S3BucketReconciler) updateBucketStatusCreated(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket, location string) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		latest := &s3v1alpha1.S3Bucket{}
		if err := r.Get(ctx, types.NamespacedName{
			Name:      s3bkt.Name,
			Namespace: s3bkt.Namespace,
		}, latest); err != nil {
			return err
		}
		latest.Status.State = s3v1alpha1.CREATED_STATE
		latest.Status.Location = location
		return r.Status().Update(ctx, latest)
	})
}

func (r *S3BucketReconciler) persistLocation(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket, location string) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		latest := &s3v1alpha1.S3Bucket{}
		if err := r.Get(ctx, types.NamespacedName{
			Name:      s3bkt.Name,
			Namespace: s3bkt.Namespace,
		}, latest); err != nil {
			return err
		}
		latest.Status.Location = location
		return r.Status().Update(ctx, latest)
	})
}

func (r *S3BucketReconciler) addFinalizer(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		latest := &s3v1alpha1.S3Bucket{}
		if err := r.Get(ctx, types.NamespacedName{
			Name:      s3bkt.Name,
			Namespace: s3bkt.Namespace,
		}, latest); err != nil {
			return err
		}

		if !controllerutil.ContainsFinalizer(latest, s3BucketFinalizer) {
			controllerutil.AddFinalizer(latest, s3BucketFinalizer)
			return r.Update(ctx, latest)
		}
		return nil
	})
}

func (r *S3BucketReconciler) removeFinalizer(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		latest := &s3v1alpha1.S3Bucket{}
		if err := r.Get(ctx, types.NamespacedName{
			Name:      s3bkt.Name,
			Namespace: s3bkt.Namespace,
		}, latest); err != nil {
			return err
		}

		controllerutil.RemoveFinalizer(latest, s3BucketFinalizer)
		return r.Update(ctx, latest)
	})
}
