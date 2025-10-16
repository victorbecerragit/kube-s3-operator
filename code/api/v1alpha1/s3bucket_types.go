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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Create constants for S3Bucket status conditions.
const (
	// PENDING_STATE indicates that the S3 bucket is pending creation.
	PENDING_STATE = "PENDING"
	// CREATED_STATE indicates that the S3 bucket was created.
	CREATED_STATE = "CREATED"
	// CREATING_STATE indicates that the S3 bucket is being created.
	CREATING_STATE = "CREATING"
	// DELETING_STATE indicates that the S3 bucket is being deleted.
	DELETING_STATE = "DELETING"
	// ERROR_STATE  indicates that the S3 bucket creation or deletion has failed.
	ERROR_STATE = "ERROR"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// S3BucketSpec defines the desired state of S3Bucket.
type S3BucketSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Name is the name of the S3 bucket
	Name string `json:"name,omitempty"` // omitempty is used to avoid issues with Terraform when the field is not set, but it is required for the API

	// Region is the AWS region where the bucket will be created
	Region string `json:"region,omitempty"` // omitempty is used to avoid issues with Terraform when the field is not set, but it is required for the API

	// Locked indicates if the bucket is locked for deletion
	Locked bool `json:"locked,omitempty"` // omitempty is used to avoid issues with Terraform when the field is not set, but it is required for the API
}

// S3BucketStatus defines the observed state of S3Bucket.
type S3BucketStatus struct {
	State string `json:"state,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Bucket Name",type="string",JSONPath=".spec.name",description="The name of the S3 bucket"
// +kubebuilder:printcolumn:name="Region",type="string",JSONPath=".spec.region",description="The AWS region of the S3 bucket"
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`,description="The current state of the S3 bucket"

// S3Bucket is the Schema for the s3buckets API.
type S3Bucket struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   S3BucketSpec   `json:"spec,omitempty"`
	Status S3BucketStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// S3BucketList contains a list of S3Bucket.
type S3BucketList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []S3Bucket `json:"items"`
}

func init() {
	SchemeBuilder.Register(&S3Bucket{}, &S3BucketList{})
}
