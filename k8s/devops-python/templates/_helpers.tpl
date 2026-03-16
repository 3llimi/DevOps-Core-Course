{{/*
Expand the name of the chart.
*/}}
{{- define "devops-python.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "devops-python.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "devops-python.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "devops-python.labels" -}}
helm.sh/chart: {{ include "devops-python.chart" . }}
{{ include "devops-python.selectorLabels" . }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "devops-python.selectorLabels" -}}
app.kubernetes.io/name: {{ include "devops-python.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
```

**6. Replace `k8s/devops-python/templates/NOTES.txt`:**
```
DevOps Python Info Service has been deployed!

Release: {{ .Release.Name }}
Namespace: {{ .Release.Namespace }}
Image: {{ .Values.image.repository }}:{{ .Values.image.tag }}
Replicas: {{ .Values.replicaCount }}

To access the service:
  minikube service {{ include "devops-python.fullname" . }} --url

Then test:
  curl <URL>/health
  curl <URL>/