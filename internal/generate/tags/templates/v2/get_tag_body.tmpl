// {{ .GetTagFunc }} fetches an individual {{ .ServicePackage }} service tag for a resource.
// Returns whether the key value and any errors. A NotFoundError is used to signal that no value was found.
// This function will optimise the handling over {{ .ListTagsFunc }}, if possible.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
{{- if or ( .TagTypeIDElem ) ( .TagTypeAddBoolElem ) }}
func {{ .GetTagFunc }}(ctx context.Context, conn {{ .ClientType }}, identifier{{ if .TagResTypeElem }}, resourceType{{ end }}, key string, optFns ...func(*{{ .AWSService }}.Options)) (*tftags.TagData, error) {
{{- else }}
func {{ .GetTagFunc }}(ctx context.Context, conn {{ .ClientType }}, identifier{{ if .TagResTypeElem }}, resourceType{{ end }}, key string, optFns ...func(*{{ .AWSService }}.Options)) (*string, error) {
{{- end }}
	{{- if .ListTagsInFiltIDName }}
	input := &{{ .AWSService  }}.{{ .ListTagsOp }}Input{
		Filters: []awstypes.Filter{
			{
				Name:   aws.String("{{ .ListTagsInFiltIDName }}"),
				Values: []string{identifier},
			},
			{
				Name:   aws.String(names.AttrKey),
				Values: []string{key},
			},
		},
	}

	output, err := conn.{{ .ListTagsOp }}(ctx, input, optFns...)

	if err != nil {
		return nil, err
	}

	listTags := {{ .KeyValueTagsFunc }}(ctx, output.{{ .ListTagsOutTagsElem }}{{ if .TagTypeIDElem }}, identifier{{ if .TagResTypeElem }}, resourceType{{ end }}{{ end }})
	{{- else }}
	listTags, err := {{ .ListTagsFunc }}(ctx, conn, identifier{{ if .TagResTypeElem }}, resourceType{{ end }}, optFns...)

	if err != nil {
		return nil, err
	}
	{{- end }}

	if !listTags.KeyExists(key) {
		return nil, tfresource.NewEmptyResultError(nil)
	}

	{{ if or ( .TagTypeIDElem ) ( .TagTypeAddBoolElem) }}
	return listTags.KeyTagData(key), nil
	{{- else }}
	return listTags.KeyValue(key), nil
	{{- end }}
}
