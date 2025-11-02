{{/*
Expand the name of the chart.
*/}}
{{- define "kube-s3-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
Uses fullnameOverride if set.
Otherwise, uses only the Helm release name to avoid duplicate prefixes.
*/}}
{{- define "kube-s3-operator.fullname" -}}
  {{- if .Values.fullnameOverride }}
    {{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
  {{- else }}
    {{- printf "%s" .Release.Name | trunc 63 | trimSuffix "-" }}
  {{- end }}
{{- end }}


{{/*
Create a fully qualified name for cluster-scoped resources (includes namespace).
This ensures ClusterRoles and ClusterRoleBindings are unique across namespaces.
*/}}
{{- define "kube-s3-operator.clustername" -}}
{{- if .Values.fullnameOverride }}
{{- printf "%s-%s" .Release.Namespace .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- printf "%s-%s" .Release.Namespace .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s-%s" .Release.Namespace .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "kube-s3-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "kube-s3-operator.labels" -}}
helm.sh/chart: {{ include "kube-s3-operator.chart" . }}
{{ include "kube-s3-operator.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "kube-s3-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "kube-s3-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "kube-s3-operator.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the RoleBinding to use
*/}}
{{- define "roleBindingName" -}}
{{- default (printf "%s-leader-election-rolebinding" (include "kube-s3-operator.fullname" .)) .Values.roleBinding.name }}
{{- end }}

{{/*
Create the name of the ClusterRoleBinding to use
*/}}
{{- define "clusterRoleBindingName" -}}
{{- $name := "cluster-role-binding" }}
{{- if and .Values.clusterRoleBinding .Values.clusterRoleBinding.name }}
  {{- $name = .Values.clusterRoleBinding.name }}
{{- end }}
{{- printf "%s-%s" (include "kube-s3-operator.fullname" .) $name }}
{{- end }}

{{/*
Create the name of the ClusterRole to use
*/}}
{{- define "clusterRoleName" -}}
{{- $name := "cluster-role" }}
{{- if and .Values.clusterRole .Values.clusterRole.name }}
  {{- $name = .Values.clusterRole.name }}
{{- end }}
{{- printf "%s-%s" (include "kube-s3-operator.fullname" .) $name }}
{{- end }}
