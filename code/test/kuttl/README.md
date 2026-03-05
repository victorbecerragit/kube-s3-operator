# KUTTL Tests for kube-s3-operator

This directory contains KUTTL (KUbernetes Test TooL) tests for the S3Bucket operator. KUTTL provides a declarative approach to test Kubernetes operators without writing procedural Go code.

## Overview

KUTTL tests are organized in test cases (directories), each containing a sequence of YAML files representing test steps:

- **00-*.yaml**: Initial setup (create resources)
- **01-*.yaml**: Assertions or modifications
- **02-*.yaml**: Cleanup or additional assertions

Each step is processed in order, and all steps must pass for the test case to succeed.

## Test Cases

### 1. s3bucket-basic-create
Tests basic S3Bucket resource creation with minimal configuration.

**Steps:**
- Create namespace (if not exists)
- Create an S3Bucket resource with basic spec (bucketName, region, acl)
- Assert the S3Bucket was created successfully

**Expected Outcome:** S3Bucket resource is created and persisted in the cluster.

### 2. s3bucket-with-labels
Tests S3Bucket creation with metadata labels (Kubernetes best practice).

**Steps:**
- Create an S3Bucket resource with labels and additional features (versioning)
- Assert the S3Bucket was created with correct labels and spec

**Expected Outcome:** S3Bucket resource includes all labels and features specified.

### 3. s3bucket-deletion
Tests S3Bucket resource deletion and cleanup.

**Steps:**
- Create an S3Bucket resource
- Verify creation using assertions
- Delete the S3Bucket resource

**Expected Outcome:** S3Bucket resource is successfully removed from the cluster.

## Running KUTTL Tests

### Prerequisites

1. **Install KUTTL CLI plugin:**
   ```bash
   go install github.com/kudobuilder/kuttl/cmd/kubectl-kuttl@latest
   ```

2. **Have a running Kubernetes cluster** (Kind, minikube, or real cluster)

3. **Install the operator in your cluster:**
   ```bash
   cd /path/to/kube-s3-operator/code
   make install
   make deploy IMG=your-image:tag
   ```

### Run Tests

From the project root:

```bash
# Run all KUTTL tests
make test-kuttl

# Or directly with kubectl plugin
kubectl kuttl test test/kuttl/

# Run specific test case
kubectl kuttl test test/kuttl/s3bucket-basic-create

# Run with verbose output
kubectl kuttl test test/kuttl/ -v
```

### Output

Test output will show:
- Test case name and status (PASSED/FAILED)
- Which step failed (if any)
- Detailed assertion failures with actual vs expected values
- Useful for debugging operator behavior

## Test File Structure

Each KUTTL test file uses one of these special object kinds:

- **TestAssert**: Verify resources exist with expected spec/status
  - Success: Resource matches the defined spec exactly
  - Failure: Resource doesn't exist or spec doesn't match

- **TestDelete**: Remove resources from the cluster
  - Success: Resource is deleted
  - Failure: Resource still exists after timeout

## Writing New KUTTL Tests

To add a new test case:

1. Create a new directory: `test/kuttl/your-test-name/`

2. Create numbered YAML files:
   ```yaml
   # 00-setup.yaml
   apiVersion: s3.kubebuilder.io/v1alpha1
   kind: S3Bucket
   metadata:
     name: my-bucket
   spec:
     bucketName: my-bucket
     region: us-east-1
   
   ---
   # 01-assert.yaml
   apiVersion: kuttl.dev/v1beta1
   kind: TestAssert
   ---
   apiVersion: s3.kubebuilder.io/v1alpha1
   kind: S3Bucket
   metadata:
     name: my-bucket
   ```

3. Run tests to validate: `make test-kuttl`

## Benefits Over Unit/E2E Tests

| Aspect | KUTTL | Traditional E2E |
|--------|-------|-----------------|
| **Readability** | Declarative YAML | Procedural Go code |
| **Maintenance** | Easy to understand | Requires domain knowledge |
| **Boilerplate** | Minimal | Lots of helper functions |
| **Debugging** | Clear assertion failures | Need to parse logs manually |
| **Integration Tests** | Native support | Requires custom setup |

## Troubleshooting

### Tests hang or timeout
- Check if operator is deployed: `kubectl get deployment -A`
- Check operator logs: `kubectl logs -n code-system deployment/code-controller-manager`
- Increase timeout in `kuttl-test.yaml`

### Assertion failures
- Verify the operator is managing resources correctly
- Check the actual resource: `kubectl get s3bucket -o yaml`
- Review operator controller logs for errors

### KUTTL CLI not found
```bash
# Reinstall KUTTL
go install github.com/kudobuilder/kuttl/cmd/kubectl-kuttl@latest

# Verify installation
kubectl kuttl version
```

## Resources

- [KUTTL Documentation](https://kudobuilder.io/docs/)
- [KUTTL GitHub](https://github.com/kudobuilder/kuttl)
- [Kubernetes Testing Best Practices](https://kubernetes.io/docs/concepts/configuration/overview/)

## Integration with Phase 2

As part of Phase 2, we'll gradually migrate the procedural E2E tests in `test/e2e/` to KUTTL format. This allows us to:

- Simplify the E2E test suite
- Reduce Go code maintenance burden
- Improve test clarity for new contributors
- Maintain compatibility with existing unit tests (Ginkgo/Gomega)
