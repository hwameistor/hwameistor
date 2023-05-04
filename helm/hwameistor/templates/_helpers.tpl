{{/* Allow KubeVersion to be overridden. */}}
{{- define "hwameistor.kubeVersion" -}}
  {{- default .Capabilities.KubeVersion.Version .Values.kubeVersion -}}
{{- end -}}

{{/* Allow Scheudler image tag to be overridden. */}}
{{- define "hwameistor.scheudlerImageTag" -}}
  {{- default .Chart.Version .Values.scheduler.tag -}}
{{- end -}}

{{/* Allow Local Storage image tag to be overridden. */}}
{{- define "hwameistor.localstorageImageTag" -}}
  {{- default .Chart.Version .Values.localStorage.member.tag -}}
{{- end -}}

{{/* Allow Scheudler image tag to be overridden. */}}
{{- define "hwameistor.localdiskmanagerImageTag" -}}
  {{- default .Chart.Version .Values.localDiskManager.manager.tag -}}
{{- end -}}

{{/* Allow Admission image tag to be overridden. */}}
{{- define "hwameistor.admissionImageTag" -}}
  {{- default .Chart.Version .Values.admission.tag -}}
{{- end -}}

{{/* Allow Evictor image tag to be overridden. */}}
{{- define "hwameistor.evictorImageTag" -}}
  {{- default .Chart.Version .Values.evictor.tag -}}
{{- end -}}

{{/* Allow Metrics image tag to be overridden. */}}
{{- define "hwameistor.exporterImageTag" -}}
  {{- default .Chart.Version .Values.exporter.tag -}}
{{- end -}}

{{/* Allow APIServer image tag to be overridden. */}}
{{- define "hwameistor.apiserverImageTag" -}}
  {{- default .Chart.Version .Values.apiserver.tag -}}
{{- end -}}

{{/* Allow UI image tag to be overridden. */}}
{{- define "hwameistor.uiImageTag" -}}
  {{- default .Chart.Version .Values.ui.tag -}}
{{- end -}}

{{/* Allow KubeletRootDir to be overridden. */}}
{{- define "hwameistor.kubeletRootDir" -}}
  {{- default "/var/lib/kubelet" .Values.kubeletRootDir -}}
{{- end -}}

{{/* Return scheduler image tag. */}}
{{/*
{{- define "hwameistor.scheduler.tag" -}}
{{- if (semverCompare "> 1.20-0" (include "hwameistor.kubeVersion" .)) }}
{{- printf "%s" .Values.scheduler.tag -}}
{{- else -}}
{{- printf "%s-%s" .Values.scheduler.tag "kube-pre1.20" -}}
{{- end -}}
{{- end -}}
*/}}
