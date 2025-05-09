{{ $CRD := .CRD }}
{{ $SDKAPI := .SDKAPI }}
{{/* This is a better way to generate JobUpdate, where the API requires you to provide Update fields */}}
{{/* Generate helper methods for Job */}}
{{- $operation := (index $SDKAPI.API.Operations "UpdateJob" ) -}}
{{- range $jobFieldName, $jobFieldMemberRef := $operation.InputRef.Shape.MemberRefs -}}
{{- if eq $jobFieldName "JobUpdate" }}
func (rm *resourceManager) new{{ $jobFieldName }}(
	    r *resource,
        res *svcsdk.{{ $operation.InputRef.ShapeName }},
) (interface{}, error) {
    jobUpdate := &svcsdktypes.{{ $jobFieldName }}{}
{{ GoCodeSetSDKForStruct $CRD "" "jobUpdate" $jobFieldMemberRef "" "r.ko.Spec" 1 }}
    res.JobUpdate = jobUpdate
	return nil, nil
}
{{- end }}
{{- end }}

