{{- define "calendar.name" -}}
calendar
{{- end }}

{{- define "calendar.fullname" -}}
{{ .Release.Name }}-calendar
{{- end }}

{{- define "calendar_scheduler.name" -}}
calendar-scheduler
{{- end }}

{{- define "calendar_scheduler.fullname" -}}
{{ .Release.Name }}-calendar-scheduler
{{- end }}

{{- define "calendar_sender.name" -}}
calendar-sender
{{- end }}

{{- define "calendar_sender.fullname" -}}
{{ .Release.Name }}-calendar-sender
{{- end }}

{{- define "postgres.name" -}}
postgres
{{- end }}

{{- define "postgres.fullname" -}}
{{ .Release.Name }}-postgres
{{- end }}

{{- define "rabbitmq.name" -}}
rabbitmq
{{- end }}

{{- define "rabbitmq.fullname" -}}
{{ .Release.Name }}-rabbitmq
{{- end }}

{{- define "goose.name" -}}
goose
{{- end }}

{{- define "goose.fullname" -}}
{{ .Release.Name }}-goose
{{- end }}
