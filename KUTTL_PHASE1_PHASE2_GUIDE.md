# KUTTL Integration - Phase 1 & 2 Implementation Guide

## Overview

This branch (`feature/kuttl-integration-phase1-2`) implements **Phase 1** (Add KUTTL alongside existing tests) and establishes the foundation for **Phase 2** (Gradually migrate E2E tests to KUTTL).

### What is KUTTL?

**KUbernetes Test TooL** (KUTTL) is a declarative testing framework specifically designed for Kubernetes operators. It allows you to write tests in YAML instead of procedural Go code.

**Key Benefits:**
- ✅ Declarative YAML-based tests (easier to read and maintain)
- ✅ No boilerplate code needed
- ✅ Built for operator testing
- ✅ Clear assertion failures
- ✅ Complements existing Ginkgo unit tests

---

## What Was Done in This Branch

### 1. Directory Structure Created

```
code/test/kuttl/
├── kuttl-test.yaml                 # Main KUTTL configuration
├── README.md                         # Documentation for KUTTL tests
├── s3bucket-basic-create/           # Test Case 1
│   ├── 00-assert-namespace.yaml
│   ├── 01-create-s3bucket.yaml
│   └── 02-assert-s3bucket-created.yaml
├── s3bucket-with-labels/            # Test Case 2
│   ├── 00-create-s3bucket-with-labels.yaml
│   └── 01-assert-bucket-with-labels.yaml
└── s3bucket-deletion/               # Test Case 3
    ├── 00-create-s3bucket-for-deletion.yaml
    ├── 01-assert-bucket-exists.yaml
    └── 02-delete-s3bucket.yaml
```

### 2. Test Cases Implemented

#### Test Case 1: `s3bucket-basic-create`
- Creates a simple S3Bucket with minimal configuration
- Verifies the resource is created successfully
- **Purpose:** Validate basic operator functionality

#### Test Case 2: `s3bucket-with-labels`
- Creates an S3Bucket with Kubernetes labels (best practice)
- Includes additional features (versioning)
- Verifies all labels and features are persisted
- **Purpose:** Test operator handles Kubernetes metadata

#### Test Case 3: `s3bucket-deletion`
- Creates an S3Bucket
- Verifies creation
- Deletes the resource
- **Purpose:** Test cleanup and resource removal

### 3. Makefile Updates

Added new KUTTL-related targets:

```bash
# Install KUTTL CLI
make kuttl-install

# Setup Kind cluster for KUTTL
make setup-test-kuttl

# Run KUTTL tests
make test-kuttl

# Run KUTTL tests (verbose)
make test-kuttl-verbose

# Cleanup Kind cluster
make cleanup-test-kuttl

# Run all tests (unit + e2e)
make test-all
```

### 4. Documentation Created

- **test/kuttl/README.md** - Comprehensive guide including:
  - KUTTL test structure explanation
  - How to run tests
  - Benefits vs traditional E2E tests
  - How to write new KUTTL tests
  - Troubleshooting guide
  - Integration with Phase 2

---

## Running KUTTL Tests

### Quick Start

```bash
# 1. Install KUTTL (one-time)
make kuttl-install

# 2. Setup Kind cluster
make setup-test-kuttl

# 3. Install operator CRDs
make install

# 4. Deploy operator
make deploy

# 5. Run tests
make test-kuttl
```

### Or All-In-One

```bash
# This will do everything above
make test-kuttl
```

### Verbose Output

```bash
# See detailed test progress
make test-kuttl-verbose
```

### Run Specific Test Case

```bash
kubectl kuttl test test/kuttl/s3bucket-basic-create/
```

### Cleanup

```bash
make cleanup-test-kuttl
```

---

## Project Structure Now

Your operator testing now has three layers:

```
┌─────────────────────────────────────────────────────────┐
│             Unit Tests (Ginkgo/Gomega)                 │
│  - Controller logic testing                            │
│  - Using envtest for isolated environment             │
│  - Fast execution (seconds)                           │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│     Integration Tests (KUTTL) ← NEW ✨                │
│  - Declarative operator behavior testing              │
│  - Real Kubernetes API interactions                   │
│  - Medium execution (minutes)                         │
│  - Easy to maintain and extend                        │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│      E2E Tests (Ginkgo + kubectl)                      │
│  - Full operator lifecycle testing                    │
│  - Uses Kind clusters                                 │
│  - Slower but most realistic                          │
└─────────────────────────────────────────────────────────┘
```

---

## KUTTL Test Structure Explained

Each test case is a directory with numbered YAML files. Files are executed in order:

```yaml
# 00-*.yaml: Create resources
apiVersion: s3.kubebuilder.io/v1alpha1
kind: S3Bucket
metadata:
  name: test-bucket
spec:
  bucketName: test-bucket
  region: us-east-1

---
# 01-*.yaml: Assert (verify resources exist with correct spec)
apiVersion: kuttl.dev/v1beta1
kind: TestAssert
---
apiVersion: s3.kubebuilder.io/v1alpha1
kind: S3Bucket
metadata:
  name: test-bucket
spec:
  bucketName: test-bucket
  region: us-east-1

---
# 02-*.yaml: Delete resources (cleanup)
apiVersion: kuttl.dev/v1beta1
kind: TestDelete
---
apiVersion: s3.kubebuilder.io/v1alpha1
kind: S3Bucket
metadata:
  name: test-bucket
```

---

## Next Steps: Phase 2 (Future)

Phase 2 will involve:

1. **Identify procedures in existing tests**
   - Review `test/e2e/e2e_test.go`
   - Extract individual test scenarios

2. **Convert to KUTTL format**
   - Create new test case directories
   - Write YAML-based tests
   - Remove procedural code

3. **New test scenarios**
   - Error handling (invalid specs)
   - Status updates
   - Operator crash recovery
   - Concurrent resources

4. **Migration guidelines**
   - Keep existing E2E tests during transition
   - Gradually replace procedures with KUTTL
   - Ensure all scenarios are covered

---

## Example: Adding a New KUTTL Test

Want to test a new scenario? Here's how:

```bash
# 1. Create test directory
mkdir -p code/test/kuttl/your-test-name

# 2. Create setup file
cat > code/test/kuttl/your-test-name/00-setup.yaml << 'EOF'
apiVersion: s3.kubebuilder.io/v1alpha1
kind: S3Bucket
metadata:
  name: my-bucket
spec:
  bucketName: my-bucket
  region: us-east-1
EOF

# 3. Create assertion file
cat > code/test/kuttl/your-test-name/01-assert.yaml << 'EOF'
apiVersion: kuttl.dev/v1beta1
kind: TestAssert
---
apiVersion: s3.kubebuilder.io/v1alpha1
kind: S3Bucket
metadata:
  name: my-bucket
EOF

# 4. Run test
make test-kuttl
```

---

## Troubleshooting

### KUTTL CLI not found
```bash
go install github.com/kudobuilder/kuttl/cmd/kubectl-kuttl@latest
which kubectl-kuttl
```

### Tests timeout
- Check operator is deployed: `kubectl get deployment -A`
- Check logs: `kubectl logs -n code-system deployment/code-controller-manager`
- Increase timeout in `code/test/kuttl/kuttl-test.yaml`

### Kind cluster issues
```bash
# List clusters
kind get clusters

# Delete stuck cluster
kind delete cluster --name code-test-e2e

# Start fresh
make setup-test-kuttl
```

### Operator not reconciling
- Check operator logs for errors
- Verify RBAC permissions
- Check S3Bucket CRD is installed: `kubectl get crd s3buckets.s3.kubebuilder.io`

---

## Files Changed/Created

### New Files
- `code/test/kuttl/kuttl-test.yaml` - Main configuration
- `code/test/kuttl/README.md` - Documentation
- `code/test/kuttl/s3bucket-basic-create/` - Test case
- `code/test/kuttl/s3bucket-with-labels/` - Test case
- `code/test/kuttl/s3bucket-deletion/` - Test case

### Modified Files
- `code/Makefile` - Added KUTTL test targets

### Unchanged Files
- `code/test/e2e/` - Still exists and functional
- Unit tests - Not affected

---

## Running Tests in CI/CD

For GitHub Actions or other CI/CD, use:

```yaml
jobs:
  test-kuttl:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
      - run: make kuttl-install
      - run: make setup-test-kuttl
      - run: make install
      - run: make deploy IMG=ghcr.io/your-org/kube-s3-operator:latest
      - run: make test-kuttl
```

---

## Key Takeaways

✅ **Phase 1 Complete:**
- KUTTL tests created alongside existing tests
- No disruption to current workflow
- Easy to run and maintain

✅ **Foundation for Phase 2:**
- Clear directory structure for future tests
- Makefile targets ready for expansion
- Documentation in place

✅ **Benefits Realized:**
- Declarative test approach
- Easier to add new test scenarios
- Better for operator-specific testing
- Complements existing Ginkgo tests

---

## Resources

- [KUTTL GitHub](https://github.com/kudobuilder/kuttl)
- [KUTTL Documentation](https://kudobuilder.io/docs/)
- [KUDO Project](https://kudo.dev/)

## Questions?

Refer to `code/test/kuttl/README.md` for detailed usage documentation.
