# kube-s3-operator

[![Go Report Card](https://goreportcard.com/badge/github.com/victorbecerragit/kube-s3-operator)](https://goreportcard.com/report/github.com/victorbecerragit/kube-s3-operator)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Release](https://img.shields.io/github/release/victorbecerragit/kube-s3-operator.svg)](https://github.com/victorbecerragit/kube-s3-operator/releases)

A Kubernetes operator for managing AWS S3 buckets using native Kubernetes resources.

## Features

- 🪣 Declarative S3 bucket management through Kubernetes CRDs
- 🔒 Bucket locking support to prevent accidental deletion
- 🌍 Multi-region bucket creation
- ♻️ Automatic reconciliation and drift detection
- 🎯 Native Kubernetes integration with kubectl

## Quick Start

### Prerequisites

- Kubernetes cluster (v1.24+)
- AWS credentials configured
- kubectl installed

### Installation

#### Using Helm (Recommended)

helm repo add kube-s3-operator https://victorbecerragit.github.io/kube-s3-operator

helm install kube-s3-operator kube-s3-operator/kube-s3-operator


#### Using kubectl


kubectl apply -f https://github.com/victorbecerragit/kube-s3-operator/releases/latest/download/install.yaml



### Usage

Create an S3 bucket:


apiVersion: s3.acme.io/v1alpha1
kind: S3Bucket
metadata:
name: my-app-bucket
spec:
bucketName: my-unique-bucket-name-12345
region: us-west-2
locked: false


## Documentation

- [Installation Guide](docs/installation.md)
- [User Guide](docs/user-guide.md)
- [API Reference](docs/api-reference.md)
- [Development Guide](docs/development.md)

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## License

This project is licensed under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

Based on the LinkedIn course "Extending Kubernetes with Operator Patterns" by Frank P Moley


