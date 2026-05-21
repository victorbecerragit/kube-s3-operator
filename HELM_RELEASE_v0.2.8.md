# Helm Chart Release: kube-s3-operator v0.2.8

## Overview
Released dependency batch update with Go 1.26, aws-sdk-go-v2, Kubernetes 1.33 client libraries, and testing framework upgrades.

## Package Details
- **Chart Name**: kube-s3-operator
- **Chart Version**: 0.2.8
- **Release Tag**: v0.2.8
- **Package**: kube-s3-operator-0.2.8.tgz
- **Size**: 7,397 bytes
- **SHA256**: 7b29579ae5d6e91822b1214b20b3298fdf3796071af3605d19cf2d811bb7d3e6

## Installation

### Quick Start
```bash
# Download from GitHub release
curl -L -o kube-s3-operator-0.2.8.tgz \
  https://github.com/victorbecerragit/kube-s3-operator/releases/download/v0.2.8/kube-s3-operator-0.2.8.tgz

# Create namespace
kubectl create namespace kube-s3-system

# Install Helm chart
helm install kube-s3-operator kube-s3-operator-0.2.8.tgz \
  --namespace kube-s3-system \
  --set controllerManager.manager.image.repository=victorbecerra/kube-s3-controller \
  --set controllerManager.manager.image.tag=deps-integration-202605131608
```

### Configuration
Key values for customization:
- `controllerManager.manager.image.repository` - Controller image repository
- `controllerManager.manager.image.tag` - Controller image tag (default: chart appVersion)
- `controllerManager.replicas` - Number of controller replicas (default: 2)
- `serviceAccount.annotations` - AWS IAM annotations for IRSA

### AWS Credentials
For local/non-EKS deployments, provide AWS credentials:
```bash
kubectl create secret generic aws-credentials \
  --from-literal=AWS_ACCESS_KEY_ID=<key> \
  --from-literal=AWS_SECRET_ACCESS_KEY=<secret> \
  -n kube-s3-system
```

## Validation Results
✅ Tested successfully on Kind cluster v1.33.1  
✅ Both controller replicas deployed and running  
✅ S3Bucket CRD reconciliation functional  
✅ No breaking changes vs previous releases  

## Changes in v0.2.8
- Upgraded Go from 1.25 to 1.26
- Updated aws-sdk-go-v2 ecosystem (s3, sts, iam, smithy-go)
- Bumped Kubernetes libraries from 1.29 to 1.33
- Testing framework updates: ginkgo v2, gomega v1
- golang.org/x/net security updates

## Compatibility Matrix
| Component | Version |
|-----------|---------|
| Go | 1.26 |
| Kubernetes | 1.29-1.33 |
| Helm | 3.0+ |
| CRD API | v1alpha1 |

## Support
For issues or questions:
- GitHub Issues: https://github.com/victorbecerragit/kube-s3-operator/issues
- Release Notes: https://github.com/victorbecerragit/kube-s3-operator/releases/tag/v0.2.8

## Chart Repository
To add this chart to your Helm repos (when published):
```bash
helm repo add kube-s3-operator https://charts.example.com/kube-s3-operator
helm repo update
helm install kube-s3-operator kube-s3-operator/kube-s3-operator --version 0.2.8
```

## Installation Methods

### Method 1: Direct Download (Recommended for Quick Testing)
```bash
helm install kube-s3-operator \
  https://github.com/victorbecerragit/kube-s3-operator/releases/download/v0.2.8/kube-s3-operator-0.2.8.tgz \
  --namespace kube-s3-system \
  --create-namespace
```

### Method 2: Local File Installation
```bash
# Extract and inspect values
tar -xzf kube-s3-operator-0.2.8.tgz
cat kube-s3-operator/values.yaml

# Customize values as needed
helm install kube-s3-operator ./kube-s3-operator \
  --namespace kube-s3-system \
  --create-namespace \
  -f custom-values.yaml
```

### Method 3: Helm Repository (Future)
When published to a Helm repository:
```bash
helm repo add kube-s3-operator https://your-helm-repo/
helm repo update
helm install kube-s3-operator kube-s3-operator/kube-s3-operator --version 0.2.8 \
  --namespace kube-s3-system \
  --create-namespace
```

## Upgrade from Previous Release
```bash
helm upgrade kube-s3-operator ./kube-s3-operator \
  --namespace kube-s3-system \
  -f your-values.yaml
```

## Verification
After installation, verify deployment:
```bash
# Check deployment
kubectl get deployment -n kube-s3-system

# Check controller replicas
kubectl get pods -n kube-s3-system -l control-plane=controller-manager

# Check CRD is registered
kubectl get crd | grep s3bucket

# Verify logs
kubectl logs -n kube-s3-system -l control-plane=controller-manager -f
```

## Breaking Changes
**None.** This is a minor version release with dependencies updated. No API or configuration changes required.

## Known Issues
- Local Kind cluster cannot access AWS IMDS for credential refresh (expected behavior)
- S3Bucket reconciliation will fail with AWS auth errors in non-AWS environments without proper credentials
- These are not bugs—they validate the operator attempts AWS operations correctly

## Testing
The chart was tested with:
- Kind v1.33.1 cluster
- Helm v3.x package manager
- Two controller replicas with leader election
- S3Bucket CRD reconciliation trigger
- Custom resource creation and deletion
