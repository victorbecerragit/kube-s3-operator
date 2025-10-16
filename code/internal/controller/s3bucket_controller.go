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
	"github.com/aws/aws-sdk-go/aws"               // AWS SDK for Go
	"github.com/aws/aws-sdk-go/aws/awserr"        // For AWS error handling
	"github.com/aws/aws-sdk-go/service/s3"        // S3 service client
	corev1 "k8s.io/api/core/v1"                   // Core Kubernetes API types (like ConfigMap)
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1" // Meta types for Kubernetes resources (like ObjectMeta)
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types" // For NamespacedName
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"time"

	s3v1alpha1 "github.com/victorbecerragit/kube-s3-operator/code/api/v1alpha1"
	"k8s.io/client-go/util/retry"                                                 // For retrying on conflict errors
	controllerutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil" // For managing finalizers
)

const (
	configMapName     = "%s-s3-cm"
	s3BucketFinalizer = "s3bucket.s3.acme.io/finalizer" // Finalizer string to be added to S3Bucket resources
)

// S3BucketReconciler reconciles a S3Bucket object
type S3BucketReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	S3svc  *s3.S3 // AWS S3 service client defined in main.go
}

// +kubebuilder:rbac:groups=s3.acme.io,resources=s3buckets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=s3.acme.io,resources=s3buckets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=s3.acme.io,resources=s3buckets/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the S3Bucket object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile

func (r *S3BucketReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Reconciling S3Bucket", "NamespacedName", req.NamespacedName)

	// Fetch the S3Bucket resource
	s3bkt := &s3v1alpha1.S3Bucket{}
	err := r.Get(ctx, req.NamespacedName, s3bkt)
	if err != nil {
		log.Info("S3Bucket resource not found, ignoring since object must be deleted")
		return ctrl.Result{}, nil
	}

	// Check if the resource is being deleted
	if !s3bkt.ObjectMeta.DeletionTimestamp.IsZero() {
		// Resource is being deleted
		log.Info("S3Bucket is being deleted", "BucketName", s3bkt.Spec.Name)

		if controllerutil.ContainsFinalizer(s3bkt, s3BucketFinalizer) {
			// Run cleanup logic
			if err := r.DeleteResource(ctx, s3bkt); err != nil {
				log.Error(err, "Failed to delete S3 bucket resources")
				return ctrl.Result{}, err
			}
			// DeleteResource handles finalizer removal internally
		}

		return ctrl.Result{}, nil
	}

	// Resource is not being deleted - ensure finalizer is present
	if !controllerutil.ContainsFinalizer(s3bkt, s3BucketFinalizer) {
		log.Info("Adding finalizer to S3Bucket", "BucketName", s3bkt.Spec.Name)
		if err := r.addFinalizer(ctx, s3bkt); err != nil {
			log.Error(err, "Failed to add finalizer")
			return ctrl.Result{}, err
		}
		// Requeue to continue reconciliation with updated resource
		return ctrl.Result{Requeue: true}, nil
	}

	// Handle creation or update logic based on current state
	switch s3bkt.Status.State {
	case "":
		// New resource - create it
		log.Info("Creating new S3 bucket", "BucketName", s3bkt.Spec.Name)
		if err := r.CreateResource(ctx, s3bkt); err != nil {
			log.Error(err, "Failed to create S3 bucket")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil

	case s3v1alpha1.CREATED_STATE:
		// Resource exists and is healthy
		log.Info("S3 bucket is in CREATED state", "BucketName", s3bkt.Spec.Name)
		// Add any update/sync logic here if needed
		return ctrl.Result{}, nil

	case s3v1alpha1.ERROR_STATE:
		// Resource is in error state - might want to retry or alert
		log.Info("S3 bucket is in ERROR state", "BucketName", s3bkt.Spec.Name)
		return ctrl.Result{RequeueAfter: time.Minute * 5}, nil

	case s3v1alpha1.CREATING_STATE, s3v1alpha1.DELETING_STATE:
		// Transitional state - requeue to check later
		log.Info("S3 bucket in transitional state", "BucketName", s3bkt.Spec.Name, "State", s3bkt.Status.State)
		return ctrl.Result{RequeueAfter: time.Second * 30}, nil

	default:
		log.Info("Unknown state for S3 bucket", "BucketName", s3bkt.Spec.Name, "State", s3bkt.Status.State)
		return ctrl.Result{}, nil
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *S3BucketReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&s3v1alpha1.S3Bucket{}).
		Named("s3bucket").
		Complete(r)
}

// CreateResource creates a new S3 bucket resource
func (r *S3BucketReconciler) CreateResource(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket) error {
	log := logf.FromContext(ctx)
	log.Info("Starting creation of S3 Bucket", "BucketName", s3bkt.Spec.Name)

	// Update status to CREATING
	if err := r.updateBucketStatus(ctx, s3bkt, s3v1alpha1.CREATING_STATE); err != nil {
		return fmt.Errorf("failed to update status to CREATING: %w", err)
	}

	// Create the S3 bucket
	bucketOutput, err := r.createS3Bucket(ctx, s3bkt)
	if err != nil {
		r.updateBucketStatus(ctx, s3bkt, s3v1alpha1.ERROR_STATE)
		return fmt.Errorf("failed to create S3 bucket: %w", err)
	}

	// Wait for bucket to be ready
	if err := r.waitForBucketReady(ctx, s3bkt); err != nil {
		r.updateBucketStatus(ctx, s3bkt, s3v1alpha1.ERROR_STATE)
		return fmt.Errorf("bucket creation timeout: %w", err)
	}

	// Create ConfigMap with bucket details
	if err := r.createBucketConfigMap(ctx, s3bkt, bucketOutput); err != nil {
		r.updateBucketStatus(ctx, s3bkt, s3v1alpha1.ERROR_STATE)
		return fmt.Errorf("failed to create ConfigMap: %w", err)
	}

	// Update status to CREATED
	if err := r.updateBucketStatus(ctx, s3bkt, s3v1alpha1.CREATED_STATE); err != nil {
		return fmt.Errorf("failed to update status to CREATED: %w", err)
	}

	log.Info("S3 Bucket created successfully", "BucketName", s3bkt.Spec.Name)
	return nil
}

// updateBucketStatus updates the bucket status with retry logic to handle conflicts
func (r *S3BucketReconciler) updateBucketStatus(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket, state string) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		// Always fetch the latest version to avoid conflicts
		latest := &s3v1alpha1.S3Bucket{}
		if err := r.Get(ctx, types.NamespacedName{
			Name:      s3bkt.Name,
			Namespace: s3bkt.Namespace,
		}, latest); err != nil {
			return err
		}

		// Update the status field
		latest.Status.State = state
		return r.Status().Update(ctx, latest)
	})
}

// createS3Bucket creates the S3 bucket using AWS SDK
func (r *S3BucketReconciler) createS3Bucket(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket) (*s3.CreateBucketOutput, error) {
	log := logf.FromContext(ctx)
	log.Info("Creating S3 bucket", "BucketName", s3bkt.Spec.Name)

	output, err := r.S3svc.CreateBucket(&s3.CreateBucketInput{
		Bucket:                     aws.String(s3bkt.Spec.Name),
		ObjectLockEnabledForBucket: aws.Bool(s3bkt.Spec.Locked),
	})
	if err != nil {
		return nil, fmt.Errorf("S3 CreateBucket API call failed: %w", err)
	}

	return output, nil
}

// waitForBucketReady waits until the bucket exists and is ready
func (r *S3BucketReconciler) waitForBucketReady(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket) error {
	log := logf.FromContext(ctx)
	log.Info("Waiting for bucket to be ready", "BucketName", s3bkt.Spec.Name)

	err := r.S3svc.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: aws.String(s3bkt.Spec.Name),
	})
	if err != nil {
		return fmt.Errorf("bucket did not become ready: %w", err)
	}

	return nil
}

// createBucketConfigMap creates a ConfigMap containing bucket metadata
func (r *S3BucketReconciler) createBucketConfigMap(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket, bucketOutput *s3.CreateBucketOutput) error {
	log := logf.FromContext(ctx)
	log.Info("Creating ConfigMap for bucket", "BucketName", s3bkt.Spec.Name)

	data := map[string]string{
		"BucketName": s3bkt.Spec.Name,
		"Region":     s3bkt.Spec.Region,
		"Locked":     fmt.Sprintf("%t", s3bkt.Spec.Locked),
		"location":   aws.StringValue(bucketOutput.Location),
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf(configMapName, s3bkt.Name),
			Namespace: s3bkt.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(s3bkt, s3v1alpha1.GroupVersion.WithKind("S3Bucket")),
			},
		},
		Data: data,
	}

	if err := r.Create(ctx, cm); err != nil {
		return fmt.Errorf("failed to create ConfigMap: %w", err)
	}

	return nil
}

// DeleteResource handles the complete deletion flow including finalizer management
func (r *S3BucketReconciler) DeleteResource(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket) error {
	log := logf.FromContext(ctx)
	log.Info("Starting deletion of S3 Bucket", "BucketName", s3bkt.Spec.Name)

	// Check if the resource has our finalizer
	if !controllerutil.ContainsFinalizer(s3bkt, s3BucketFinalizer) {
		log.Info("Finalizer not found, resource likely already cleaned up", "BucketName", s3bkt.Spec.Name)
		return nil
	}

	// Update status to DELETING
	if err := r.updateBucketStatus(ctx, s3bkt, s3v1alpha1.DELETING_STATE); err != nil {
		log.Error(err, "Failed to update status to DELETING, continuing with deletion")
		// Don't return error here - we want to proceed with deletion even if status update fails
	}

	// Perform the actual cleanup operations
	if err := r.performCleanup(ctx, s3bkt); err != nil {
		r.updateBucketStatus(ctx, s3bkt, s3v1alpha1.ERROR_STATE)
		return fmt.Errorf("cleanup failed: %w", err)
	}

	// Remove finalizer after successful cleanup
	if err := r.removeFinalizer(ctx, s3bkt); err != nil {
		return fmt.Errorf("failed to remove finalizer: %w", err)
	}

	log.Info("S3 Bucket deleted successfully and finalizer removed", "BucketName", s3bkt.Spec.Name)
	return nil
}

// performCleanup performs all cleanup operations (S3 bucket, ConfigMap, etc.)
func (r *S3BucketReconciler) performCleanup(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket) error {
	log := logf.FromContext(ctx)

	// Delete the S3 bucket
	if err := r.deleteS3Bucket(ctx, s3bkt); err != nil {
		return fmt.Errorf("failed to delete S3 bucket: %w", err)
	}

	// Wait for bucket to be fully deleted
	if err := r.waitForBucketDeleted(ctx, s3bkt); err != nil {
		return fmt.Errorf("bucket deletion timeout: %w", err)
	}

	// Delete the ConfigMap (best effort - don't fail if it doesn't exist)
	if err := r.deleteBucketConfigMap(ctx, s3bkt); err != nil {
		log.Error(err, "Failed to delete ConfigMap, but bucket is deleted", "BucketName", s3bkt.Spec.Name)
		// Don't return error - the bucket is already deleted
	}

	return nil
}

// removeFinalizer removes the finalizer from the resource
func (r *S3BucketReconciler) removeFinalizer(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		// Fetch the latest version
		latest := &s3v1alpha1.S3Bucket{}
		if err := r.Get(ctx, types.NamespacedName{
			Name:      s3bkt.Name,
			Namespace: s3bkt.Namespace,
		}, latest); err != nil {
			return err
		}

		// Remove finalizer
		controllerutil.RemoveFinalizer(latest, s3BucketFinalizer)

		// Update the resource
		return r.Update(ctx, latest)
	})
}

// addFinalizer adds the finalizer to the resource
func (r *S3BucketReconciler) addFinalizer(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		// Fetch the latest version
		latest := &s3v1alpha1.S3Bucket{}
		if err := r.Get(ctx, types.NamespacedName{
			Name:      s3bkt.Name,
			Namespace: s3bkt.Namespace,
		}, latest); err != nil {
			return err
		}

		// Add finalizer if not present
		if !controllerutil.ContainsFinalizer(latest, s3BucketFinalizer) {
			controllerutil.AddFinalizer(latest, s3BucketFinalizer)
			return r.Update(ctx, latest)
		}

		return nil
	})
}

// deleteS3Bucket deletes the S3 bucket using AWS SDK
func (r *S3BucketReconciler) deleteS3Bucket(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket) error {
	log := logf.FromContext(ctx)
	log.Info("Deleting S3 bucket", "BucketName", s3bkt.Spec.Name)

	_, err := r.S3svc.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(s3bkt.Spec.Name),
	})
	if err != nil {
		// Check if bucket doesn't exist (already deleted)
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				log.Info("Bucket already deleted or doesn't exist", "BucketName", s3bkt.Spec.Name)
				return nil
			case "BucketNotEmpty":
				return fmt.Errorf("bucket is not empty, cannot delete: %w", err)
			}
		}
		return fmt.Errorf("S3 DeleteBucket API call failed: %w", err)
	}

	return nil
}

// waitForBucketDeleted waits until the bucket is fully deleted
func (r *S3BucketReconciler) waitForBucketDeleted(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket) error {
	log := logf.FromContext(ctx)
	log.Info("Waiting for bucket to be deleted", "BucketName", s3bkt.Spec.Name)

	err := r.S3svc.WaitUntilBucketNotExists(&s3.HeadBucketInput{
		Bucket: aws.String(s3bkt.Spec.Name),
	})
	if err != nil {
		return fmt.Errorf("bucket deletion did not complete: %w", err)
	}

	return nil
}

// deleteBucketConfigMap deletes the ConfigMap associated with the bucket
func (r *S3BucketReconciler) deleteBucketConfigMap(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket) error {
	log := logf.FromContext(ctx)
	log.Info("Deleting ConfigMap for bucket", "BucketName", s3bkt.Spec.Name)

	cm := &corev1.ConfigMap{}
	cmName := types.NamespacedName{
		Name:      fmt.Sprintf(configMapName, s3bkt.Name),
		Namespace: s3bkt.Namespace,
	}

	err := r.Get(ctx, cmName, cm)
	if err != nil {
		log.Info("ConfigMap already deleted or doesn't exist", "ConfigMapName", cmName.Name)
		return nil
	}

	if err := r.Delete(ctx, cm); err != nil {
		// Already deleted
		return nil
	}

	return nil
}
