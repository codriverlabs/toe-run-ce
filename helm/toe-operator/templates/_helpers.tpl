{{/*
Expand the name of the chart.
*/}}
{{- define "toe-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "toe-operator.fullname" -}}
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
{{- define "toe-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "toe-operator.labels" -}}
helm.sh/chart: {{ include "toe-operator.chart" . }}
{{ include "toe-operator.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "toe-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "toe-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Controller image
*/}}
{{- define "toe-operator.controller.image" -}}
{{- printf "%s:%s" .Values.controller.image.repository .Values.controller.image.tag }}
{{- end }}

{{/*
Collector image  
*/}}
{{- define "toe-operator.collector.image" -}}
{{- printf "%s:%s" .Values.collector.image.repository .Values.collector.image.tag }}
{{- end }}

{{/*
Namespace
*/}}
{{- define "toe-operator.namespace" -}}
{{- default "toe-system" .Values.global.namespace }}
{{- end }}
