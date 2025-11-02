
#

helm repo index . --url https://victorbecerragit.github.io/kube-s3-operator

helm package charts/kube-s3-operator
