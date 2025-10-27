{{/*
Expand the name of the chart.
*/}}
{{- define "taishanglaojun-platform.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "taishanglaojun-platform.fullname" -}}
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
{{- define "taishanglaojun-platform.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "taishanglaojun-platform.labels" -}}
helm.sh/chart: {{ include "taishanglaojun-platform.chart" . }}
{{ include "taishanglaojun-platform.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: taishanglaojun-platform
{{- end }}

{{/*
Selector labels
*/}}
{{- define "taishanglaojun-platform.selectorLabels" -}}
app.kubernetes.io/name: {{ include "taishanglaojun-platform.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "taishanglaojun-platform.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "taishanglaojun-platform.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the namespace to use
*/}}
{{- define "taishanglaojun-platform.namespace" -}}
{{- default "taishanglaojun" .Values.namespace.name }}
{{- end }}

{{/*
Create image pull policy
*/}}
{{- define "taishanglaojun-platform.imagePullPolicy" -}}
{{- default "IfNotPresent" .Values.image.pullPolicy }}
{{- end }}

{{/*
Create image repository with registry
*/}}
{{- define "taishanglaojun-platform.imageRepository" -}}
{{- if .Values.global.imageRegistry }}
{{- printf "%s/%s" .Values.global.imageRegistry .repository }}
{{- else }}
{{- .repository }}
{{- end }}
{{- end }}

{{/*
Create storage class name
*/}}
{{- define "taishanglaojun-platform.storageClass" -}}
{{- if .Values.global.storageClass }}
{{- .Values.global.storageClass }}
{{- else if .Values.persistence.storageClass }}
{{- .Values.persistence.storageClass }}
{{- else }}
{{- "" }}
{{- end }}
{{- end }}