
# Respect this order, package should be generated first, otherwise, the index-yaml will be empty.

helm package charts/kube-s3-operator


helm repo index . --url https://victorbecerragit.github.io/kube-s3-operator

