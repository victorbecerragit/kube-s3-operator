// EventRecorder is responsible for recording events related to S3Bucket resources.

package recorder

import (
    "sigs.k8s.io/controller-runtime/pkg/event"
    "sigs.k8s.io/controller-runtime/pkg/recorder"
    api "github.com/victorbecerragit/kube-s3-operator/api/v1alpha1" // Assuming api package contains the S3Bucket definition
)

type Recorder interface {
    Normal(obj api.Object, reason, message string)
    Warning(obj api.Object, reason, message string)
}

func New(recorder recorder.EventRecorder) Recorder {
    return &rec{recorder: recorder}
}

type rec struct {
    recorder recorder.EventRecorder
}

func (r *rec) Normal(obj api.Object, reason, message string) {
    r.recorder.Event(obj, event.Normal, reason, message)
}

func (r *rec) Warning(obj api.Object, reason, message string) {
    r.recorder.Event(obj, event.Warning, reason, message)
}
