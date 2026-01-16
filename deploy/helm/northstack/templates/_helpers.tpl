{{/* NorthStack Helm Helpers */}}

{{/*
Expand the name of the chart.
*/}}
{{- define "northstack.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "northstack.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "northstack.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "northstack.labels" -}}
helm.sh/chart: {{ include "northstack.chart" . }}
{{ include "northstack.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: northstack
{{- end }}

{{/*
Selector labels
*/}}
{{- define "northstack.selectorLabels" -}}
app.kubernetes.io/name: {{ include "northstack.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "northstack.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "northstack.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Cloud provider specific ingress class
*/}}
{{- define "northstack.ingressClassName" -}}
{{- if eq .Values.global.cloudProvider "aws" }}
alb
{{- else if eq .Values.global.cloudProvider "gcp" }}
gce
{{- else if eq .Values.global.cloudProvider "azure" }}
azure/application-gateway
{{- else }}
nginx
{{- end }}
{{- end }}

{{/*
Storage class based on cloud provider
*/}}
{{- define "northstack.storageClass" -}}
{{- if eq .Values.global.cloudProvider "aws" }}
gp3
{{- else if eq .Values.global.cloudProvider "gcp" }}
pd-ssd
{{- else if eq .Values.global.cloudProvider "azure" }}
managed-premium
{{- else if eq .Values.global.cloudProvider "openstack" }}
cinder-ssd
{{- else }}
longhorn
{{- end }}
{{- end }}
