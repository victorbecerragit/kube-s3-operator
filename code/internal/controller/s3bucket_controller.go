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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
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
)

const (
	configMapName     = "%s-s3-cm"
	s3BucketFinalizer = "s3bucket.s3.acme.io/finalizer" // Finalizer string to be added to S3Bucket resources
)

// S3BucketReconciler reconciles a S3Bucket object
type S3BucketReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	S3svc  *s3.Client // AWS S3 service client defined in main.go
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
	if !s3bkt.DeletionTimestamp.IsZero() {
		// Resource is being deleted
		log.Info("S3Bucket is being deleted", "BucketName", s3bkt.Spec.Name)

		if controllerutil.ContainsFinalizer(s3bkt, s3BucketFinalizer) {
			// Our finalizer is present, so handle deletion
			if err := r.DeleteResource(ctx, s3bkt); err != nil {
				log.Error(err, "Failed to delete S3 bucket resources")
				return ctrl.Result{}, err
			}
			// Remove finalizer after successful deletion
			controllerutil.RemoveFinalizer(s3bkt, s3BucketFinalizer)
			_ = r.Update(ctx, s3bkt)
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
		// Ensure lifecycle configuration is applied (idempotent call).
		if err := r.applyLifecycleConfiguration(ctx, s3bkt); err != nil {
			if err2 := r.updateBucketStatus(ctx, s3bkt, s3v1alpha1.ERROR_STATE); err2 != nil {
				log.Error(err2, "Failed to update status to ERROR")
			}
			return ctrl.Result{}, err
		}
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
	location, err := r.createS3Bucket(ctx, s3bkt)
	if err != nil {
		if err := r.updateBucketStatus(ctx, s3bkt, s3v1alpha1.ERROR_STATE); err != nil {
			log.Error(err, "Failed to update status to ERROR")
		}
		return fmt.Errorf("failed to create S3 bucket: %w", err)
	}

	// Wait for bucket to be ready
	if err := r.waitForBucketReady(ctx, s3bkt); err != nil {
		if err := r.updateBucketStatus(ctx, s3bkt, s3v1alpha1.ERROR_STATE); err != nil {
			log.Error(err, "Failed to update status to ERROR")
		}
		return fmt.Errorf("bucket creation timeout: %w", err)
	}

	// Apply lifecycle configuration (optional).
	if err := r.applyLifecycleConfiguration(ctx, s3bkt); err != nil {
		if err := r.updateBucketStatus(ctx, s3bkt, s3v1alpha1.ERROR_STATE); err != nil {
			log.Error(err, "Failed to update status to ERROR")
		}
		return fmt.Errorf("failed to apply lifecycle configuration: %w", err)
	}

	// Create ConfigMap with bucket details
	if err := r.createBucketConfigMap(ctx, s3bkt, location); err != nil {
		if err := r.updateBucketStatus(ctx, s3bkt, s3v1alpha1.ERROR_STATE); err != nil {
			log.Error(err, "Failed to update status to ERROR")
		}
		return fmt.Errorf("failed to create ConfigMap: %w", err)
	}

	// Update status to CREATED
	if err := r.updateBucketStatus(ctx, s3bkt, s3v1alpha1.CREATED_STATE); err != nil {
		return fmt.Errorf("failed to update status to CREATED: %w", err)
	}

	log.Info("S3 Bucket created successfully", "BucketName", s3bkt.Spec.Name)
	return nil
}

func (r *S3BucketReconciler) applyLifecycleConfiguration(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket) error {
	if s3bkt.Spec.Lifecycle == nil || len(s3bkt.Spec.Lifecycle.Rules) == 0 {
		return nil
	}

	rules := make([]s3types.LifecycleRule, 0, len(s3bkt.Spec.Lifecycle.Rules))
	for i, rule := range s3bkt.Spec.Lifecycle.Rules {
		status := rule.Status
		if status == "" {
			status = "Enabled"
		}

		awsStatus := s3types.ExpirationStatusEnabled
		if status == "Disabled" {
			awsStatus = s3types.ExpirationStatusDisabled
		}

		awsRule := s3types.LifecycleRule{
			Status: awsStatus,
		}

		if rule.ID != "" {
			awsRule.ID = aws.String(rule.ID)
		} else {
			// Ensure a stable ID for drift/idempotency.
			awsRule.ID = aws.String(fmt.Sprintf("%s-rule-%d", s3bkt.Name, i))
		}

		if rule.Prefix != "" {
			awsRule.Filter = &s3types.LifecycleRuleFilterMemberPrefix{
				Value: rule.Prefix,
			}
		}

		if rule.Expiration != nil && rule.Expiration.Days > 0 {
			awsRule.Expiration = &s3types.LifecycleExpiration{
				Days: aws.Int32(rule.Expiration.Days),
			}
		}

		if len(rule.Transitions) > 0 {
			awsRule.Transitions = make([]s3types.Transition, 0, len(rule.Transitions))
			for _, t := range rule.Transitions {
				if t.Days <= 0 || t.StorageClass == "" {
					continue
				}
				awsRule.Transitions = append(awsRule.Transitions, s3types.Transition{
					Days:         aws.Int32(t.Days),
					StorageClass: s3types.TransitionStorageClass(t.StorageClass),
				})
			}
		}

		if rule.NoncurrentVersionExpiration != nil && rule.NoncurrentVersionExpiration.Days > 0 {
			awsRule.NoncurrentVersionExpiration = &s3types.NoncurrentVersionExpiration{
				NoncurrentDays: aws.Int32(rule.NoncurrentVersionExpiration.Days),
			}
		}

		if len(rule.NoncurrentVersionTransitions) > 0 {
			awsRule.NoncurrentVersionTransitions = make([]s3types.NoncurrentVersionTransition, 0, len(rule.NoncurrentVersionTransitions))
			for _, t := range rule.NoncurrentVersionTransitions {
				if t.Days <= 0 || t.StorageClass == "" {
					continue
				}
				awsRule.NoncurrentVersionTransitions = append(awsRule.NoncurrentVersionTransitions, s3types.NoncurrentVersionTransition{
					NoncurrentDays: aws.Int32(t.Days),
					StorageClass:   s3types.TransitionStorageClass(t.StorageClass),
				})
			}
		}

		rules = append(rules, awsRule)
	}

	input := &s3.PutBucketLifecycleConfigurationInput{
		Bucket: aws.String(s3bkt.Spec.Name),
		LifecycleConfiguration: &s3types.BucketLifecycleConfiguration{
			Rules: rules,
		},
	}
	_, err := r.S3svc.PutBucketLifecycleConfiguration(ctx, input)
	if err != nil {
		return fmt.Errorf("PutBucketLifecycleConfiguration failed: %w", err)
	}
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

// createS3Bucket creates the S3 bucket using AWS SDK v2
func (r *S3BucketReconciler) createS3Bucket(
	ctx context.Context,
	s3bkt *s3v1alpha1.S3Bucket,
) (string, error) {
	log := logf.FromContext(ctx)
	log.Info("Creating S3 bucket", "BucketName", s3bkt.Spec.Name)

	output, err := r.S3svc.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket:                     aws.String(s3bkt.Spec.Name),
		ObjectLockEnabledForBucket: aws.Bool(s3bkt.Spec.Locked),
	})
	if err != nil {
		return "", fmt.Errorf("S3 CreateBucket API call failed: %w", err)
	}

	// Extract the location from the response
	location := ""
	if output.Location != nil {
		location = *output.Location
	}
	return location, nil
}

// waitForBucketReady waits until the bucket exists and is ready
func (r *S3BucketReconciler) waitForBucketReady(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket) error {
	log := logf.FromContext(ctx)
	log.Info("Waiting for bucket to be ready", "BucketName", s3bkt.Spec.Name)

	// Use a simple poll loop to wait for the bucket
	for i := 0; i < 60; i++ {
		_, err := r.S3svc.HeadBucket(ctx, &s3.HeadBucketInput{
			Bucket: aws.String(s3bkt.Spec.Name),
		})
		if err == nil {
			log.Info("Bucket is ready", "BucketName", s3bkt.Spec.Name)
			return nil
		}
		time.Sleep(time.Second)
	}

	return fmt.Errorf("bucket did not become ready within timeout")
}

// createBucketConfigMap creates a ConfigMap containing bucket metadata
func (r *S3BucketReconciler) createBucketConfigMap(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket, location string) error {
	log := logf.FromContext(ctx)
	log.Info("Creating ConfigMap for bucket", "BucketName", s3bkt.Spec.Name)

	data := map[string]string{
		"BucketName": s3bkt.Spec.Name,
		"Region":     s3bkt.Spec.Region,
		"Locked":     fmt.Sprintf("%t", s3bkt.Spec.Locked),
		"location":   location,
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
		if err := r.updateBucketStatus(ctx, s3bkt, s3v1alpha1.ERROR_STATE); err != nil {
			log.Error(err, "Failed to update status to ERROR")
		}
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
	// Delete the S3 bucket
	if err := r.deleteS3Bucket(ctx, s3bkt); err != nil {
		return fmt.Errorf("failed to delete S3 bucket: %w", err)
	}

	// Wait for bucket to be fully deleted
	if err := r.waitForBucketDeleted(ctx, s3bkt); err != nil {
		return fmt.Errorf("bucket deletion timeout: %w", err)
	}

	// Delete the ConfigMap (best effort - don't fail if it doesn't exist)
	r.deleteBucketConfigMap(ctx, s3bkt)

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

// deleteS3Bucket deletes the S3 bucket using AWS SDK v2
func (r *S3BucketReconciler) deleteS3Bucket(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket) error {
	log := logf.FromContext(ctx)
	log.Info("Deleting S3 bucket", "BucketName", s3bkt.Spec.Name)

	_, err := r.S3svc.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(s3bkt.Spec.Name),
	})
	if err != nil {
		var apiErr any
		if err != nil {
			apiErr = err.Error()
		}
		// Check if bucket doesn't exist (already deleted) or other S3 errors
		if err != nil {
			errStr := err.Error()
			if errStr == "NoSuchBucket" || errStr == "<?xml version=\"1.0\" encoding=\"UTF-8\"?>" {
				log.Info("Bucket already deleted or doesn't exist", "BucketName", s3bkt.Spec.Name)
				return nil
			}
			if errStr == "BucketNotEmpty" {
				return fmt.Errorf("bucket is not empty, cannot delete: %w", err)
			}
		}
		_ = apiErr
		return fmt.Errorf("S3 DeleteBucket API call failed: %w", err)
	}

	return nil
}

// waitForBucketDeleted waits until the bucket is fully deleted
func (r *S3BucketReconciler) waitForBucketDeleted(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket) error {
	log := logf.FromContext(ctx)
	log.Info("Waiting for bucket to be deleted", "BucketName", s3bkt.Spec.Name)

	// Use a poll loop to wait for the bucket to be deleted
	for i := 0; i < 60; i++ {
		_, err := r.S3svc.HeadBucket(ctx, &s3.HeadBucketInput{
			Bucket: aws.String(s3bkt.Spec.Name),
		})
		if err != nil {
			// Bucket doesn't exist - deletion is complete
			log.Info("Bucket has been deleted", "BucketName", s3bkt.Spec.Name)
			return nil
		}
		time.Sleep(time.Second)
	}

	return fmt.Errorf("bucket deletion did not complete within timeout")
}

// deleteBucketConfigMap deletes the ConfigMap associated with the bucket.
func (r *S3BucketReconciler) deleteBucketConfigMap(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket) {
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
		return
	}

	if err := r.Delete(ctx, cm); err != nil {
		// Already deleted
		return
	}

	log.Info("ConfigMap deleted successfully", "ConfigMapName", cmName.Name)
}
