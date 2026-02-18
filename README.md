# kube-s3-operator

[![Go Report Card](https://goreportcard.com/badge/github.com/victorbecerragit/kube-s3-operator/code)](https://goreportcard.com/report/github.com/victorbecerragit/kube-s3-operator/code)
[![Go Version](https://img.shields.io/badge/go%20version-1.25.0-blue)](https://golang.org/dl/)
[![Kubernetes](https://img.shields.io/badge/kubernetes-v0.35.1-blue)](https://kubernetes.io/)
[![AWS SDK](https://img.shields.io/badge/aws--sdk-v2-orange)](https://github.com/aws/aws-sdk-go-v2)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Release](https://img.shields.io/github/release/victorbecerragit/kube-s3-operator.svg)](https://github.com/victorbecerragit/kube-s3-operator/releases)

A Kubernetes operator for managing AWS S3 buckets using native Kubernetes resources.

## Features

- 🪣 Declarative S3 bucket management through Kubernetes CRDs
- 🔒 Bucket locking support to prevent accidental deletion
- 🌍 Multi-region bucket creation
- ♻️ Automatic reconciliation and drift detection
- 🎯 Native Kubernetes integration with kubectl
- 🚀 AWS SDK v2 for enhanced security and performance
- ✅ Go 1.25.0 with latest Kubernetes v0.35.1 compatibility
- 📝 Comprehensive testing with Ginkgo v2.28.1 and Gomega v1.39.0

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

Create a aws secrets:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: aws-secret
type: Opaque
data:
  # Replace the following base64 encoded strings with your actual AWS credentials
  aws-access-key-id: xxxxxxxxxx
  aws-secret-access-key: xxxxxxxxxxx==
```

## Documentation - TODO

- [Installation Guide](docs/installation.md)
- [User Guide](docs/user-guide.md)
- [API Reference](docs/api-reference.md)
- [Development Guide](docs/development.md)

## Recent Updates

### Latest Release (v2.0.0)
- **AWS SDK Migration**: Upgraded from AWS SDK v1 (EOL: July 31, 2025) to AWS SDK v2
- **Go Update**: Bumped to Go 1.25.0 for latest performance and security improvements
- **Kubernetes Libraries**: Updated to v0.35.1 (k8s.io/api, k8s.io/apimachinery, k8s.io/client-go)
- **Controller Runtime**: Updated to v0.23.1 for better stability
- **Testing Framework**: Ginkgo v2.28.1 and Gomega v1.39.0

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


