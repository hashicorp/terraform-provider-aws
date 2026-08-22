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
	input := {{ .AWSService }}.{{ .ListTagsOp }}Input{
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

	{{ if .RetryTagOps }}
	{{- if eq (len .RetryConditions) 1 }}
	output, err := tfresource.RetryWhenIsAErrorMessageContains[*{{ .AWSService }}.{{ .RetryTagsListTagsType }}, *{{ (index .RetryConditions 0).Code }}](ctx, {{ .RetryTimeout }},
	{{- else if eq (len .RetryConditions) 2 }}
	output, err := tfresource.RetryWhenIsOneOf2ErrorMessageContains[*{{ .AWSService }}.{{ .RetryTagsListTagsType }}, *{{ (index .RetryConditions 0).Code }}, *{{ (index .RetryConditions 1).Code }}](ctx, {{ .RetryTimeout }},
	{{- else if eq (len .RetryConditions) 3 }}
	output, err := tfresource.RetryWhenIsOneOf3ErrorMessageContains[*{{ .AWSService }}.{{ .RetryTagsListTagsType }}, *{{ (index .RetryConditions 0).Code }}, *{{ (index .RetryConditions 1).Code }}, *{{ (index .RetryConditions 2).Code }}](ctx, {{ .RetryTimeout }},
	{{- end }}
		func(ctx context.Context) (*{{ .AWSService }}.{{ .RetryTagsListTagsType }}, error) {
			return conn.{{ .ListTagsOp }}(ctx, &input, optFns...)
		},
	{{- range .RetryConditions }}
		"{{ .Message }}",
	{{- end }}
	)
	{{ else }}
	output, err := conn.{{ .ListTagsOp }}(ctx, &input, optFns...)
	{{- end }}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	listTags := {{ .KeyValueTagsFunc }}(ctx, output.{{ .ListTagsOutTagsElem }}{{ if .TagTypeIDElem }}, identifier{{ if .TagResTypeElem }}, resourceType{{ end }}{{ end }})
	{{- else }}
	listTags, err := {{ .ListTagsFunc }}(ctx, conn, identifier{{ if .TagResTypeElem }}, resourceType{{ end }}, optFns...)

	if err != nil {
		return nil, smarterr.NewError(err)
	}
	{{- end }}

	if !listTags.KeyExists(key) {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	{{ if or ( .TagTypeIDElem ) ( .TagTypeAddBoolElem) }}
	return listTags.KeyTagData(key), nil
	{{- else }}
	return listTags.KeyValue(key), nil
	{{- end }}
}
