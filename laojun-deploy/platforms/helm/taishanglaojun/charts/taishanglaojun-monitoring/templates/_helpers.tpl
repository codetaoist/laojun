{{/*
Expand the name of the chart.
*/}}
{{- define "taishanglaojun-monitoring.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "taishanglaojun-monitoring.fullname" -}}
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
{{- define "taishanglaojun-monitoring.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "taishanglaojun-monitoring.labels" -}}
helm.sh/chart: {{ include "taishanglaojun-monitoring.chart" . }}
{{ include "taishanglaojun-monitoring.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: taishanglaojun
{{- end }}

{{/*
Selector labels
*/}}
{{- define "taishanglaojun-monitoring.selectorLabels" -}}
app.kubernetes.io/name: {{ include "taishanglaojun-monitoring.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "taishanglaojun-monitoring.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "taishanglaojun-monitoring.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the namespace name
*/}}
{{- define "taishanglaojun-monitoring.namespace" -}}
{{- if .Values.namespaceOverride }}
{{- .Values.namespaceOverride }}
{{- else if .Values.global.namespace }}
{{- .Values.global.namespace }}
{{- else }}
{{- .Release.Namespace }}
{{- end }}
{{- end }}

{{/*
Get the image pull policy
*/}}
{{- define "taishanglaojun-monitoring.imagePullPolicy" -}}
{{- if .Values.global.imagePullPolicy }}
{{- .Values.global.imagePullPolicy }}
{{- else }}
{{- .Values.image.pullPolicy | default "IfNotPresent" }}
{{- end }}
{{- end }}

{{/*
Get the image repository with registry
*/}}
{{- define "taishanglaojun-monitoring.imageRepository" -}}
{{- $image := . -}}
{{- if .global.imageRegistry }}
{{- printf "%s/%s" .global.imageRegistry $image.repository }}
{{- else }}
{{- $image.repository }}
{{- end }}
{{- end }}

{{/*
Get the storage class
*/}}
{{- define "taishanglaojun-monitoring.storageClass" -}}
{{- if .Values.global.storageClass }}
{{- .Values.global.storageClass }}
{{- else if .Values.persistence.storageClass }}
{{- .Values.persistence.storageClass }}
{{- else }}
{{- "" }}
{{- end }}
{{- end }}