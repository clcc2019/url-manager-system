{{/*
Expand the name of the chart.
*/}}
{{- define "url-manager.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "url-manager.fullname" -}}
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
{{- define "url-manager.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "url-manager.labels" -}}
helm.sh/chart: {{ include "url-manager.chart" . }}
{{ include "url-manager.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "url-manager.selectorLabels" -}}
app.kubernetes.io/name: {{ include "url-manager.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "url-manager.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "url-manager.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Backend labels
*/}}
{{- define "url-manager.backend.labels" -}}
{{ include "url-manager.labels" . }}
app.kubernetes.io/component: backend
{{- end }}

{{/*
Backend selector labels
*/}}
{{- define "url-manager.backend.selectorLabels" -}}
{{ include "url-manager.selectorLabels" . }}
app.kubernetes.io/component: backend
{{- end }}

{{/*
Frontend labels
*/}}
{{- define "url-manager.frontend.labels" -}}
{{ include "url-manager.labels" . }}
app.kubernetes.io/component: frontend
{{- end }}

{{/*
Frontend selector labels
*/}}
{{- define "url-manager.frontend.selectorLabels" -}}
{{ include "url-manager.selectorLabels" . }}
app.kubernetes.io/component: frontend
{{- end }}

{{/*
Database URL
*/}}
{{- define "url-manager.databaseUrl" -}}
{{- if .Values.postgresql.enabled }}
{{- printf "postgres://postgres:%s@%s-postgresql:5432/%s?sslmode=disable" .Values.postgresql.auth.password .Release.Name .Values.postgresql.auth.database }}
{{- else }}
{{- .Values.backend.config.database.url }}
{{- end }}
{{- end }}

{{/*
Redis Address
*/}}
{{- define "url-manager.redisAddress" -}}
{{- if .Values.redis.enabled }}
{{- printf "%s-redis-master:6379" .Release.Name }}
{{- else }}
{{- .Values.backend.config.redis.address }}
{{- end }}
{{- end }}