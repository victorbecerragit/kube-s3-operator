# kube-s3-operator

[![Go Report Card](https://goreportcard.com/badge/github.com/victorbecerragit/kube-s3-operator/code)](https://goreportcard.com/report/github.com/victorbecerragit/kube-s3-operator/code)
[![Go Version](https://img.shields.io/badge/go%20version-1.25.0-blue)](https://golang.org/dl/)
[![Kubernetes](https://img.shields.io/badge/kubernetes-v0.35.1-blue)](https://kubernetes.io/)
[![AWS SDK](https://img.shields.io/badge/aws--sdk-v2-orange)](https://github.com/aws/aws-sdk-go-v2)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Release](https://img.shields.io/github/release/victorbecerragit/kube-s3-operator.svg)](https://github.com/victorbecerragit/kube-s3-operator/releases)
[![Tests](https://github.com/victorbecerragit/kube-s3-operator/actions/workflows/tests.yml/badge.svg)](https://github.com/victorbecerragit/kube-s3-operator/actions/workflows/tests.yml)

A Kubernetes operator for managing AWS S3 buckets using native Kubernetes resources.

## Features

- 🪣 Declarative S3 bucket management through Kubernetes CRDs
- 🔒 Bucket locking support to prevent accidental deletion
- 🌍 Multi-region bucket creation
- ♻️ Automatic reconciliation and drift detection
- 🏷️ **S3 Lifecycle Policy Support**: Define and manage S3 bucket lifecycle rules (expiration, transition, etc.) via CRD spec
- 🌐 **Region Auto-Detection**: Automatically detects and uses the correct AWS region for each bucket
- 🧹 **Automatic S3 Bucket Cleanup in Tests**: All S3 buckets created during tests are deleted after each test run
- 🛡️ **Secure Commit History**: All AWS credentials and secrets are fully removed from the repository history
- 🎯 Native Kubernetes integration with kubectl
- 🚀 AWS SDK v2 for enhanced security and performance
- 🔐 Optional AWS credentials — explicit K8s secrets **or** IAM roles / IRSA (no secret required)
- ✅ Go 1.25.0 with latest Kubernetes v0.35.1 compatibility
- 📝 Comprehensive testing with Ginkgo v2.28.1 and Gomega v1.39.0

## S3 Ecosystem Weekly Trends

This section is updated weekly from CI to highlight recent S3-compatible storage news and competitor tooling.

<!-- S3_TRENDS_START -->
Last updated: 2026-04-13 (UTC)

- [Object Storage Comparison 2026: 21 S3 Providers ...](https://mixpeek.com/blog/object-storage-comparison-2026) — Side-by-side comparison of 21 S3-compatible storage providers — AWS, GCS, Azure, R2, B2, Wasabi, MinIO, and more. Real pricing, escape costs ...
- [S3 Compatible Object Storage Solutions](https://www.cloudflare.com/developer-platform/use-cases/s3-compatible-object-storage/) — Cloudflare R2 is compatible with S3. R2's S3-compatible API allows developers to access a wide range of S3 tools, libraries, and extensions.
- [Choosing the Right S3 Alternatives for Artifact Storage](https://blog.inedo.com/proget/s3-alternatives) — Cloud and on-premises S3-compatible providers like Cloudflare R2, Wasabi, Backblaze B2, and MinIO let teams reduce storage expenses, eliminate ...
- [What s3 compatible object store has the mainstream ...](https://www.reddit.com/r/selfhosted/comments/1s6t9rf/what_s3_compatible_object_store_has_the/) — What s3 compatible object store has the mainstream community moved on to from minio ... Minio removed admin features from the web ui in latest ...
- [S3-Compatible Object Storage: The Best Solutions at a ...](https://lowcloud.io/en/blog/s3-compatible-object-storage) — MinIO is the most well-known S3-compatible self-hosted object storage. Written in Go, it is performant and relatively easy to operate. MinIO has ...
<!-- S3_TRENDS_END -->

## Quick Start

### Prerequisites

- Kubernetes cluster (v1.28+, tested with v0.35.1)
- Go 1.25.0 or later
- AWS credentials configured
- kubectl installed

### Installation

#### Using Helm (Recommended)

helm repo add kube-s3-operator https://victorbecerragit.github.io/kube-s3-operator

helm install kube-s3-operator kube-s3-operator/kube-s3-operator

helm upgrade --install s3-operator --namespace s3-acme code/charts/kube-s3-operator/ --values charts/kube-s3-operator/default-values.yaml --dry-run

#### Using kubectl

kubectl apply -f https://github.com/victorbecerragit/kube-s3-operator/code/config/default/install.yaml

### Usage

Create an sample S3 bucket:

```yaml
apiVersion: s3.acme.io/v1alpha1
kind: S3Bucket
metadata:
name: my-app-bucket
spec:
bucketName: my-unique-bucket-name-12345
region: us-west-2
locked: false
```

### AWS Credentials

The operator supports two authentication modes, controlled by the `awsCredentials.enabled` Helm value.

#### Option 1: IAM roles / IRSA — no secret needed (default)

By default (`awsCredentials.enabled: false`) the operator uses the AWS **default credential chain**:
- IAM role attached to the node or pod (EKS IRSA, EC2 instance profile)
- `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` environment variables
- `~/.aws/credentials` file
- EC2 instance metadata service

```bash
# Install with default settings — no secret required
helm install kube-s3-operator kube-s3-operator/kube-s3-operator
```

#### Option 2: Explicit credentials via Kubernetes Secret

Enable this mode to read credentials from a K8s Secret:

```bash
# 1. Create the secret BEFORE deploying the operator
kubectl create secret generic aws-secret \
  --from-literal=aws-access-key-id=YOUR_AWS_ACCESS_KEY_ID \
  --from-literal=aws-secret-access-key=YOUR_AWS_SECRET_ACCESS_KEY \
  -n <namespace>

# 2. Install with credentials enabled
helm install kube-s3-operator kube-s3-operator/kube-s3-operator \
  --set awsCredentials.enabled=true
```

> ⚠️ **Important**: The secret must exist **before** the operator pod starts. If the secret is missing the pod will fail with `CreateContainerConfigError`.

## Testing

The project uses a three-layer testing strategy, each layer catching a different class of problem:

| Layer | Tool | What it validates |
|-------|------|-------------------|
| **Unit** | Go `testing` + Ginkgo/Gomega | Controller logic, reconciliation loops, error handling — no cluster needed |
| **KUTTL** | [KUTTL](https://kuttl.dev) (declarative YAML) | Kubernetes operator behavior — CRD lifecycle, status conditions, drift detection against a real `kind` cluster |
| **E2E** | Ginkgo + `kind` | Full lifecycle — deploy operator via `make deploy`, create CRs, verify AWS API interactions, test teardown |

### Why KUTTL alongside E2E?

KUTTL tests are plain YAML — no Go code required. Each test is an `apply` + assertion file that anyone can read and write without knowing the codebase. This makes them ideal for validating *Kubernetes-native behavior* (does the `S3Bucket` CR reach `Ready` status? does the controller reconcile after a spec change?) quickly and declaratively.

E2E tests (Ginkgo/Go) handle what KUTTL can't: complex multi-step scenarios, programmatic setup/teardown, real AWS API call verification, and nuanced error-path assertions.

### Run tests locally

```bash
# Unit tests
cd code && make test

# KUTTL integration tests (requires a kind cluster named kuttl-test)
kind create cluster --name kuttl-test
cd code && make kuttl-test

# E2E tests (cluster is created and destroyed automatically)
cd code && make test-e2e
```

## Recent Updates

### Upcoming Release (Architecture Refactoring)
- **Modular Architecture**: Extracted AWS SDK operations into a dedicated `s3client` package, achieving strict separation of concerns from the K8s reconciler.
- **Performance Optimization**: Implemented a thread-safe `BucketManager` that caches AWS clients per-region, preventing redundant initialization/config loads on every reconcile loop.
- **Idiomatic Kubernetes Reconciliation**: Eliminated blocking `time.Sleep` wait loops from the reconciler, substituting them with non-blocking `RequeueAfter` polls to prevent worker thread starvation.
- **Enhanced Reliability**: Reworked ConfigMap lifecycle routines to rely on `controllerutil.CreateOrUpdate` for flawless idempotent state syncing.
- **Accelerated Testing**: Substituted live AWS dependencies with a complete mock memory client (`mockBucketManager`) within the suite, drastically increasing test velocity and guaranteeing test stability without actual credentials.

### Latest Release (v0.2.6)
- **AWS SDK Refresh**: Rebased on aws-sdk-go-v2 v1.41.5, including the lifecycle filter fix so controller lifecycle rules continue to compile and reconcile with the newest SDK definitions
- **Chart + App Version**: Published as Helm `0.2.6` to stay aligned with the controller release train and make the SDK upgrade available to Helm consumers
- **Go & Kubernetes Compatibility**: Maintains Go 1.25.0 and Kubernetes v0.35.1 compatibility while keeping controller-runtime at v0.23.1
- **Tests & Tooling**: All existing Ginkgo/Gomega/KUTTL suites continue running in CI with the updated dependencies

### Version Compatibility Matrix

| Component | Version | Status |
|-----------|---------|--------|
| Go        | 1.25.0  | ✅ Latest |
| Kubernetes| v0.35.1 | ✅ Latest |
| AWS SDK   | v2      | ✅ Current |
| Controller Runtime | v0.23.1 | ✅ Compatible |
| Ginkgo    | v2.28.1 | ✅ Latest |
| Gomega    | v1.39.0 | ✅ Latest |

## 🚀 Automated Release Pipeline

This project uses GitHub Actions to automate Helm chart publication and validation.

### Release Workflow

When you push a **version tag** (e.g., `v0.2.0`), the automated pipeline:

1. **Validates** the Helm chart
   - Runs `helm lint` for syntax checks
   - Renders templates and validates structure
   - Validates all Kubernetes manifests

2. **Requests Approval**
   - Creates a GitHub issue with release details
   - Provides approval/cancel options via comments
   - Allows code review before publication

3. **Publishes** to Helm Repository
   - Packages the Helm chart
   - Generates/updates `index.yaml`
   - Preserves all previous versions
   - Publishes to [gh-pages branch](https://github.com/victorbecerragit/kube-s3-operator/tree/gh-pages)

4. **Creates Release Notes**
   - Auto-generates GitHub release
   - Includes version information
   - Links to chart documentation

### How to Release a New Version

```bash
# 1. Update chart version in code/charts/kube-s3-operator/Chart.yaml
# 2. Update appVersion if needed
# 3. Commit changes
git add code/charts/kube-s3-operator/Chart.yaml
git commit -m "chore: bump chart version to X.Y.Z"
git push origin main

# 4. Create and push a version tag
git tag -a vX.Y.Z -m "Release vX.Y.Z - description of changes"
git push origin vX.Y.Z

# 5. GitHub Actions automatically:
#    - Validates the chart
#    - Creates approval issue
#    - Publishes to Helm repository
#    - Generates release notes
```

### Helm Repository Configuration

The project maintains a public Helm repository at:
```
https://victorbecerragit.github.io/kube-s3-operator
```

**Add the repository:**
```bash
helm repo add kube-s3-operator https://victorbecerragit.github.io/kube-s3-operator
helm repo update
```

**View available versions:**
```bash
helm search repo kube-s3-operator -A
```

**Install a specific version:**
```bash
# Latest version
helm install my-operator kube-s3-operator/kube-s3-operator

# Specific version
helm install my-operator kube-s3-operator/kube-s3-operator --version 0.2.0

# Previous version (rollback)
helm install my-operator kube-s3-operator/kube-s3-operator --version 0.1.1
```

### Workflow Configuration

Helm chart publication is configured in [`.github/workflows/publish-helm-chart.yml`](.github/workflows/publish-helm-chart.yml).

Key features:
- ✅ Automatic validation on version tags
- ✅ Version history preservation (no data loss)
- ✅ Approval mechanism for production safety
- ✅ Intelligent index merging for multi-version support
- ✅ Comprehensive release documentation

## Contributing

We welcome contributions! Please see our [Contributing Guide](docs/CONTRIBUTING.md) for details.

## License

This project is licensed under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

Based on the LinkedIn course "Extending Kubernetes with Operator Patterns" by Frank P Moley


