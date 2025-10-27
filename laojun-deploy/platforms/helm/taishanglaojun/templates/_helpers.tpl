{{/*
Expand the name of the chart.
*/}}
{{- define "taishanglaojun.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "taishanglaojun.fullname" -}}
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
{{- define "taishanglaojun.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "taishanglaojun.labels" -}}
helm.sh/chart: {{ include "taishanglaojun.chart" . }}
{{ include "taishanglaojun.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: taishanglaojun
{{- end }}

{{/*
Selector labels
*/}}
{{- define "taishanglaojun.selectorLabels" -}}
app.kubernetes.io/name: {{ include "taishanglaojun.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "taishanglaojun.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "taishanglaojun.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Get the global domain
*/}}
{{- define "taishanglaojun.domain" -}}
{{- .Values.global.domain | default "taishanglaojun.local" }}
{{- end }}

{{/*
Get the image registry
*/}}
{{- define "taishanglaojun.imageRegistry" -}}
{{- .Values.global.imageRegistry | default "" }}
{{- end }}

{{/*
Get the storage class
*/}}
{{- define "taishanglaojun.storageClass" -}}
{{- .Values.global.storageClass | default "" }}
{{- end }}