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
{{- printf "%s/toe-controller:%s" .Values.global.registry.repository .Values.global.version }}
{{- end }}

{{/*
Collector image  
*/}}
{{- define "toe-operator.collector.image" -}}
{{- printf "%s/toe-collector:%s" .Values.global.registry.repository .Values.global.version }}
{{- end }}

{{/*
Aperf image
*/}}
{{- define "toe-operator.aperf.image" -}}
{{- printf "%s/toe-aperf:%s" .Values.global.registry.repository .Values.global.version }}
{{- end }}

{{/*
Tcpdump image
*/}}
{{- define "toe-operator.tcpdump.image" -}}
{{- printf "%s/toe-tcpdump:%s" .Values.global.registry.repository .Values.global.version }}
{{- end }}

{{/*
Chaos image
*/}}
{{- define "toe-operator.chaos.image" -}}
{{- printf "%s/toe-chaos:%s" .Values.global.registry.repository .Values.global.version }}
{{- end }}

{{/*
Namespace
*/}}
{{- define "toe-operator.namespace" -}}
{{- default "toe-system" .Values.global.namespace }}
{{- end }}

{{/*
Image pull secrets - only include if not using IRSA for ECR
*/}}
{{- define "toe-operator.imagePullSecrets" -}}
{{- if and (eq .Values.global.registry.type "ecr") (not .Values.ecr.useIRSA) }}
{{- if .Values.ecr.secretName }}
- name: {{ .Values.ecr.secretName }}
{{- end }}
{{- else if .Values.global.imagePullSecrets }}
{{- range .Values.global.imagePullSecrets }}
- name: {{ . }}
{{- end }}
{{- end }}
{{- end }}
