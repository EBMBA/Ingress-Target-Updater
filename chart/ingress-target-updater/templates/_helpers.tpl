{{/*
Expand the name of the chart.
*/}}
{{- define "ingress-target-updater.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "ingress-target-updater.fullname" -}}
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
{{- define "ingress-target-updater.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "ingress-target-updater.labels" -}}
helm.sh/chart: {{ include "ingress-target-updater.chart" . }}
{{ include "ingress-target-updater.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "ingress-target-updater.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ingress-target-updater.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "ingress-target-updater.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "ingress-target-updater.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Define the namespace name to use
*/}}
{{- define "ingress-target-updater.namespace" -}}
{{- default .Release.Namespace .Values.namespaceOverride | trunc 63 | trimSuffix "-" }}
{{- end }}


{{/*
Define the ClusterRole name to use
*/}}
{{- define "ingress-target-updater.clusterRoleName" -}}
{{- printf "%s-%s" (include "ingress-target-updater.fullname" .) "clusterrole" }}
{{- end }}

{{/* 
Define the ClusterRoleBinding name to use
*/}}
{{- define "ingress-target-updater.clusterRoleBindingName" -}}
{{- printf "%s-%s" (include "ingress-target-updater.fullname" .) "clusterrolebinding" }}
{{- end }}



