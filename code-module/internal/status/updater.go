// StatusUpdater is responsible for updating the status of S3Bucket resources. (Patching helpers)

package status

import (
    "context"
    "sigs.k8s.io/controller-runtime/pkg/client"
    api "github.com/victorbecerragit/kube-s3-operator/api/v1alpha1"
)

type Updater interface {
    UpdateBucketStatus(ctx context.Context, obj *api.S3Bucket, ready bool, reason, message string) error
}

func New(client client.Client) Updater {
    return &updater{client: client}
}

type updater struct {
    client client.Client
}

func (u *updater) UpdateBucketStatus(ctx context.Context, obj *api.S3Bucket, ready bool, reason, message string) error {
    obj.Status.Ready = ready
    obj.Status.Reason = reason
    obj.Status.Message = message
    return u.client.Status().Patch(ctx, obj, client.MergeFrom(obj.DeepCopy()))
}
