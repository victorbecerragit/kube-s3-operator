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

## Contributing

We welcome contributions! Please see our [Contributing Guide](docs/CONTRIBUTING.md) for details.

## License

This project is licensed under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

Based on the LinkedIn course "Extending Kubernetes with Operator Patterns" by Frank P Moley


