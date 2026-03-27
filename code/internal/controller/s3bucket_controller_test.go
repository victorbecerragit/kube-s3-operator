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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	s3v1alpha1 "github.com/victorbecerragit/kube-s3-operator/code/api/v1alpha1"
)

type mockBucketManager struct {
	buckets map[string]bool
	failAWS bool
}

func (m *mockBucketManager) CreateBucket(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket) (string, error) {
	if m.failAWS {
		return "", fmt.Errorf("mock AWS error")
	}
	m.buckets[s3bkt.Spec.Name] = true
	return "http://mock.location", nil
}

func (m *mockBucketManager) DeleteBucket(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket) error {
	if m.failAWS {
		return fmt.Errorf("mock AWS error")
	}
	delete(m.buckets, s3bkt.Spec.Name)
	return nil
}

func (m *mockBucketManager) IsBucketReady(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket) (bool, error) {
	if m.failAWS {
		return false, fmt.Errorf("mock AWS error")
	}
	_, exists := m.buckets[s3bkt.Spec.Name]
	return exists, nil
}

func (m *mockBucketManager) IsBucketDeleted(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket) (bool, error) {
	if m.failAWS {
		return false, fmt.Errorf("mock AWS error")
	}
	_, exists := m.buckets[s3bkt.Spec.Name]
	return !exists, nil
}

func (m *mockBucketManager) ApplyLifecycleConfiguration(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket) error {
	if m.failAWS {
		return fmt.Errorf("mock AWS error")
	}
	return nil
}

var _ = Describe("S3Bucket Controller", func() {
	const testRegion = "us-west-2"
	ctx := context.Background()

	uniqueName := func(prefix string) string {
		return prefix + "-" + fmt.Sprintf("%d", time.Now().UnixNano())
	}

	getBucket := func(name string) *s3v1alpha1.S3Bucket {
		bucket := &s3v1alpha1.S3Bucket{}
		err := k8sClient.Get(ctx, types.NamespacedName{
			Name:      name,
			Namespace: "default",
		}, bucket)
		Expect(err).NotTo(HaveOccurred())
		return bucket
	}

	var mockManager *mockBucketManager

	BeforeEach(func() {
		mockManager = &mockBucketManager{
			buckets: make(map[string]bool),
		}
	})

	It("should reconcile and set status transitions", func() {
		resourceName := uniqueName("test-resource")
		bucketName := uniqueName("test-bucket")

		By("creating the custom resource")
		resource := &s3v1alpha1.S3Bucket{
			ObjectMeta: metav1.ObjectMeta{
				Name:      resourceName,
				Namespace: "default",
			},
			Spec: s3v1alpha1.S3BucketSpec{
				Name:   bucketName,
				Region: testRegion,
				Locked: false,
			},
		}
		Expect(k8sClient.Create(ctx, resource)).To(Succeed())

		controllerReconciler := &S3BucketReconciler{
			Client:    k8sClient,
			Scheme:    k8sClient.Scheme(),
			S3Manager: mockManager,
		}

		// First reconcile: adds finalizer
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      resourceName,
				Namespace: "default",
			},
		})
		Expect(err).NotTo(HaveOccurred())

		bucket := getBucket(resourceName)
		Expect(bucket.Status.State).To(Equal("")) // Still empty state, finalizer added

		// Second reconcile: should create bucket and set state to CREATED
		// Since our mock returns ready=true immediately, it skips CREATING directly to CREATED
		_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      resourceName,
				Namespace: "default",
			},
		})
		Expect(err).NotTo(HaveOccurred())
		bucket = getBucket(resourceName)
		Expect(bucket.Status.State).To(Equal("CREATED"))

		// Cleanup
		Expect(k8sClient.Delete(ctx, bucket)).To(Succeed())

		// Reconcile deletion
		_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      resourceName,
				Namespace: "default",
			},
		})
		Expect(err).NotTo(HaveOccurred())
	})

	It("should handle lifecycle configuration in spec", func() {
		resourceName := uniqueName("test-resource-lifecycle")
		bucketName := uniqueName("test-bucket-lifecycle")

		By("creating S3Bucket with lifecycle rules")
		resource := &s3v1alpha1.S3Bucket{
			ObjectMeta: metav1.ObjectMeta{
				Name:      resourceName,
				Namespace: "default",
			},
			Spec: s3v1alpha1.S3BucketSpec{
				Name:   bucketName,
				Region: testRegion,
				Locked: false,
				Lifecycle: &s3v1alpha1.S3BucketLifecycleConfiguration{
					Rules: []s3v1alpha1.S3BucketLifecycleRule{{
						ID:     "expire-30-days",
						Status: "Enabled",
						Prefix: "",
						Expiration: &s3v1alpha1.S3BucketLifecycleExpiration{
							Days: 30,
						},
					}},
				},
			},
		}
		Expect(k8sClient.Create(ctx, resource)).To(Succeed())

		controllerReconciler := &S3BucketReconciler{
			Client:    k8sClient,
			Scheme:    k8sClient.Scheme(),
			S3Manager: mockManager,
		}

		// First reconcile (add finalizer)
		_, _ = controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      resourceName,
				Namespace: "default",
			},
		})

		// Second reconcile (create & apply lifecycle)
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      resourceName,
				Namespace: "default",
			},
		})
		Expect(err).NotTo(HaveOccurred())
		bucket := getBucket(resourceName)
		Expect(bucket.Status.State).To(Equal("CREATED"))

		// Cleanup
		Expect(k8sClient.Delete(ctx, bucket)).To(Succeed())
	})

	It("should set ERROR state on AWS failure", func() {
		resourceName := uniqueName("test-resource-error")
		bucketName := uniqueName("test-bucket-error")

		By("creating S3Bucket with enforced AWS failure in mock")
		mockManager.failAWS = true

		resource := &s3v1alpha1.S3Bucket{
			ObjectMeta: metav1.ObjectMeta{
				Name:      resourceName,
				Namespace: "default",
			},
			Spec: s3v1alpha1.S3BucketSpec{
				Name:   bucketName,
				Region: testRegion,
				Locked: false,
			},
		}
		Expect(k8sClient.Create(ctx, resource)).To(Succeed())

		controllerReconciler := &S3BucketReconciler{
			Client:    k8sClient,
			Scheme:    k8sClient.Scheme(),
			S3Manager: mockManager,
		}

		// Add finalizer
		_, _ = controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      resourceName,
				Namespace: "default",
			},
		})

		// Next reconcile should return an error due to mock failure, and state becomes ERROR
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      resourceName,
				Namespace: "default",
			},
		})
		Expect(err).To(HaveOccurred())

		bucket := getBucket(resourceName)
		Expect(bucket.Status.State).To(Equal("ERROR"))

		// Cleanup
		mockManager.failAWS = false // Reset mock so deletion succeeds
		Expect(k8sClient.Delete(ctx, bucket)).To(Succeed())

		_, _ = controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      resourceName,
				Namespace: "default",
			},
		})
	})
})
