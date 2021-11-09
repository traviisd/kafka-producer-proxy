{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "{{ .Release.Name }}" -}}
{{- $name := default .Release.Name -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "local.selectorLabels" -}}
app: {{ .Release.Name }}
{{- end }}

{{/*
Additional labels
*/}}
{{- define "local.labels" -}}
app: {{ .Release.Name }}
{{- if .Values.labels }}
{{- toYaml .Values.labels | nindent 0 }}
{{- end }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "local.serviceAccountName" -}}
{{- if .Values.serviceAccount }}
{{- default .Release.Name .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}


{{/*
Mount the service account if not default. 
*/}}
{{- define "local.automountServiceAccountToken" -}}
{{ $saname := include "local.serviceAccountName" . }}
{{- if ne $saname "default" }}
automountServiceAccountToken: true
{{- end }}
{{- end }}


{{/*
Default resources
*/}}
{{- define "local.resources" -}}
{{- if .Values.resources }}
{{- toYaml .Values.resources | nindent 0 }}
{{- else }}
requests:
  cpu: 10m
  memory: 128Mi
limits:
  cpu: 1
  memory: 2G
{{- end }}
{{- end }}
{{- end }}

{{/*
Default service ports
*/}}
{{- define "local.service" -}}
{{- if .Values.service }}
{{- toYaml .Values.service | nindent 0 }}
{{- else }}
type: ClusterIP
port: 80
containerPort: 39000
annotations: {}
{{- end }}
{{- end }}
{{- end }}

{{/*
Default volume mounts
*/}}
{{- define "local.volumeMounts" -}}
- name: config-volume
  mountPath: /config
{{- if .Values.volumeMounts }}
{{- toYaml .Values.volumeMounts | nindent 0 }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Default volumes
*/}}
{{- define "local.volumes" -}}
- name: config-volume
  configMap:
    name: {{ .Release.Name }}
{{- if .Values.volumes }}
{{- toYaml .Values.volumes | nindent 0 }}
{{- end }}
{{- end }}
{{- end }}