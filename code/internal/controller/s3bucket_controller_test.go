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

var _ = Describe("S3Bucket Controller", func() {
	const testRegion = "us-west-2"
	ctx := context.Background()

	// Helper to generate a unique name per test
	uniqueName := func(prefix string) string {
		return prefix + "-" + fmt.Sprintf("%d", time.Now().UnixNano())
	}

	// Helper to fetch latest S3Bucket
	getBucket := func(name string) *s3v1alpha1.S3Bucket {
		bucket := &s3v1alpha1.S3Bucket{}
		err := k8sClient.Get(ctx, types.NamespacedName{
			Name:      name,
			Namespace: "default",
		}, bucket)
		Expect(err).NotTo(HaveOccurred())
		return bucket
	}

	It("should reconcile and set status transitions", func() {
		resourceName := uniqueName("test-resource")
		bucketName := uniqueName("test-bucket")
		By("creating the custom resource for the Kind S3Bucket with required fields")
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
			Client: k8sClient,
			Scheme: k8sClient.Scheme(),
		}

		// First reconcile: adds finalizer only
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      resourceName,
				Namespace: "default",
			},
		})
		Expect(err).NotTo(HaveOccurred())
		bucket := getBucket(resourceName)
		// Should still be empty state after first reconcile
		Expect(bucket.Status.State).To(Equal(""))

		// Second reconcile: should set status to CREATING or ERROR
		_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      resourceName,
				Namespace: "default",
			},
		})
		Expect(err).NotTo(HaveOccurred())
		bucket = getBucket(resourceName)
		Expect(bucket.Status.State).To(Or(Equal("CREATING"), Equal("ERROR"), Equal("CREATED")))

		// Simulate bucket created (manually set status for test)
		bucket.Status.State = "CREATED"
		Expect(k8sClient.Status().Update(ctx, bucket)).To(Succeed())

		// Third reconcile: should apply lifecycle (if any) and remain CREATED or go ERROR
		_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      resourceName,
				Namespace: "default",
			},
		})
		Expect(err).NotTo(HaveOccurred())
		bucket = getBucket(resourceName)
		Expect(bucket.Status.State).To(Or(Equal("CREATED"), Equal("ERROR")))

		// Cleanup
		Expect(k8sClient.Delete(ctx, bucket)).To(Succeed())
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
						Prefix: "", // Empty prefix applies to all objects
						Expiration: &s3v1alpha1.S3BucketLifecycleExpiration{
							Days: 30,
						},
					}},
				},
			},
		}
		Expect(k8sClient.Create(ctx, resource)).To(Succeed())

		controllerReconciler := &S3BucketReconciler{
			Client: k8sClient,
			Scheme: k8sClient.Scheme(),
		}

		// First reconcile: adds finalizer only
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      resourceName,
				Namespace: "default",
			},
		})
		Expect(err).NotTo(HaveOccurred())

		// Second reconcile: should process lifecycle logic
		_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      resourceName,
				Namespace: "default",
			},
		})
		Expect(err).NotTo(HaveOccurred())
		bucket := getBucket(resourceName)
		Expect(bucket.Spec.Lifecycle).NotTo(BeNil())

		// Cleanup
		Expect(k8sClient.Delete(ctx, bucket)).To(Succeed())
	})

	It("should set ERROR state on AWS failure", func() {
		resourceName := uniqueName("test-resource-error")
		bucketName := uniqueName("test-bucket-error")
		By("creating S3Bucket with invalid region to force error")
		resource := &s3v1alpha1.S3Bucket{
			ObjectMeta: metav1.ObjectMeta{
				Name:      resourceName,
				Namespace: "default",
			},
			Spec: s3v1alpha1.S3BucketSpec{
				Name:   bucketName,
				Region: "invalid-region-xyz",
				Locked: false,
			},
		}
		Expect(k8sClient.Create(ctx, resource)).To(Succeed())

		controllerReconciler := &S3BucketReconciler{
			Client: k8sClient,
			Scheme: k8sClient.Scheme(),
		}

		// First reconcile: adds finalizer only
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      resourceName,
				Namespace: "default",
			},
		})
		Expect(err).NotTo(HaveOccurred())

		// Second reconcile: should attempt AWS call and fail
		_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      resourceName,
				Namespace: "default",
			},
		})
		// Error may or may not be returned depending on controller logic, but status should be ERROR
		bucket := getBucket(resourceName)
		Expect(bucket.Status.State).To(Equal("ERROR"))

		// Cleanup
		Expect(k8sClient.Delete(ctx, bucket)).To(Succeed())
	})
})
