```yaml

# kube-s3-operator
Based on linkedin course "Extending Kubernetes with Operator Patterns" by Frank P Moley

# Navigate to your code directory
cd /path/to/kube-s3-operator/code

# Check if go.mod exists and is correct
cat go.mod

# If go.mod doesn't exist or has issues, reinitialize
rm -f go.mod go.sum
go mod init github.com/victorbecerragit/kube-s3-operator

# Ensure kubebuilder project is properly initialized
kubebuilder init --domain acme.io --repo github.com/victorbecerragit/kube-s3-operator

# Install controller-runtime and other dependencies
go get sigs.k8s.io/controller-runtime@v0.21.0

# Tidy up dependencies
go mod tidy

# Verify dependencies are properly installed
go mod verify

# Create the S3 API and controller
kubebuilder create api --group s3 --version v1alpha1 --kind S3Bucket --resource --controller


This will create:

Kubernetes kind object "S3Bucket"

api/v1alpha1/s3bucket_types.go - Define your S3Bucket custom resource

controllers/s3bucket_controller.go - Implement your reconciliation logic

# Define Your Custom Resource (api/v1alpha1/s3bucket_types.go)

sample: 
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

# Implement the controller logic (controllers/s3bucket_controller.go)


# Generate/regenerate code like CRD, RBAC, and other manifests like rbac for the group "s3.acme.io" (From --group "s3"  and --domain "acme.io")
make manifests

# Verify generated files
ls -la config/crd/bases/
ls -la config/rbac/


# Build and test 
# Install CRDs in your cluster
make install

# Run the operator locally for testing
make run

# Sample resource
# Create a sample S3Bucket resource as from config/samples/s3_v1alpha1_s3bucket.yaml

cat <<EOF | kubectl apply -f -
apiVersion: s3.acme.io/v1alpha1
kind: S3Bucket
metadata:
  name: my-test-bucket
  namespace: default
spec:
  bucketName: my-unique-test-bucket-12345
  region: us-west-2
  accessControl: private
EOF

# Check the resource
kubectl get s3buckets
kubectl describe s3bucket my-test-bucket

# Build container image

# Build the operator image
make docker-build IMG=your-registry/kube-s3-operator:v1.0.0

# Push to your registry
make docker-push IMG=your-registry/kube-s3-operator:v1.0.0

# Deploy to Kubernetes
# Deploy the operator to your cluster
make deploy IMG=your-registry/kube-s3-operator:v1.0.0

# Verify deployment
kubectl get pods -n kube-s3-operator-system
kubectl logs -n kube-s3-operator-system deployment/kube-s3-operator-controller-manager

# Test Deployed Operator

# Create test resources
kubectl apply -f config/samples/

# Monitor the operator logs
kubectl logs -f -n kube-s3-operator-system deployment/kube-s3-operator-controller-manager

# Check resource status
kubectl get s3buckets -o wide

# Project structure for code/

kube-s3-operator/
├── api/
│   └── v1alpha1/
│       ├── groupversion_info.go
│       ├── s3bucket_types.go
│       └── zz_generated.deepcopy.go
├── bin/
├── config/
│   ├── crd/
│   ├── default/
│   ├── manager/
│   ├── rbac/
│   └── samples/
├── controllers/
│   ├── s3bucket_controller.go
│   └── suite_test.go
├── Dockerfile
├── go.mod
├── go.sum
├── main.go
├── Makefile
├── PROJECT
└── README.md

# Project structure for code-module/
code-module/
└── internal/
    ├── controller/
    │   └── s3bucket_controller.go         # Controller wiring only
    ├── s3client/
    │   └── s3-service.go                     # S3 API operations
    ├── status/
    │   └── updater.go                     # Status patching helpers
    └── recorder/
        └── event-recorder.go                       # Event-recording abstraction

```


