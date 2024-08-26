// {{ .ListTagsFunc }} lists {{ .ServicePackage }} service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func {{ .ListTagsFunc }}(ctx context.Context, conn {{ .ClientType }}, identifier{{ if .TagResTypeElem }}, resourceType{{ end }} string, optFns ...func(*{{ .AWSService }}.Options)) (tftags.KeyValueTags, error) {
	input := &{{ .TagPackage  }}.{{ .ListTagsOp }}Input{
		{{- if .ListTagsInFiltIDName }}
		Filters: []awstypes.Filter{
			{
				Name:   aws.String("{{ .ListTagsInFiltIDName }}"),
				Values: []string{identifier},
			},
		},
		{{- else }}
		{{- if .ListTagsInIDNeedSlice }}
		{{ .ListTagsInIDElem }}: aws.StringSlice([]string{identifier}),
		{{- else if .ListTagsInIDNeedValueSlice }}
		{{ .ListTagsInIDElem }}: []string{identifier},
		{{- else }}
		{{ .ListTagsInIDElem }}: aws.String(identifier),
		{{- end }}
		{{- if .TagResTypeElem }}
		{{- if .TagResTypeElemType }}
		{{ .TagResTypeElem }}:         awstypes.{{ .TagResTypeElemType }}(resourceType),
		{{- else }}
		{{ .TagResTypeElem }}:         aws.String(resourceType),
		{{- end }}
		{{- end }}
		{{- end }}
	}
{{- if .ListTagsOpPaginated }}
	var output []awstypes.{{ or .TagType2 .TagType }}

	pages := {{ .TagPackage  }}.New{{ .ListTagsOp }}Paginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx, optFns...)

	{{ if and ( .ParentNotFoundErrCode ) ( .ParentNotFoundErrMsg ) }}
			if tfawserr.ErrMessageContains(err, "{{ .ParentNotFoundErrCode }}", "{{ .ParentNotFoundErrMsg }}") {
				return nil, &retry.NotFoundError{
					LastError:   err,
					LastRequest: input,
				}
			}
	{{- else if ( .ParentNotFoundErrCode ) }}
			if tfawserr.ErrCodeEquals(err, "{{ .ParentNotFoundErrCode }}") {
				return nil, &retry.NotFoundError{
					LastError:   err,
					LastRequest: input,
				}
			}
	{{- end }}

		if err != nil {
			return tftags.New(ctx, nil), err
		}

		for _, v := range page.{{ .ListTagsOutTagsElem }} {
			output = append(output, v)
		}
	}

	return {{ .KeyValueTagsFunc }}(ctx, output{{ if .TagTypeIDElem }}, identifier{{ if .TagResTypeElem }}, resourceType{{ end }}{{ end }}), nil
{{- else }}

	output, err := conn.{{ .ListTagsOp }}(ctx, input, optFns...)

	{{ if and ( .ParentNotFoundErrCode ) ( .ParentNotFoundErrMsg ) }}
			if tfawserr.ErrMessageContains(err, "{{ .ParentNotFoundErrCode }}", "{{ .ParentNotFoundErrMsg }}") {
				return nil, &retry.NotFoundError{
					LastError:   err,
					LastRequest: input,
				}
			}
	{{- else if ( .ParentNotFoundErrCode ) }}
			if tfawserr.ErrCodeEquals(err, "{{ .ParentNotFoundErrCode }}") {
				return nil, &retry.NotFoundError{
					LastError:   err,
					LastRequest: input,
				}
			}
	{{- end }}

	if err != nil {
		return tftags.New(ctx, nil), err
	}

	return {{ .KeyValueTagsFunc }}(ctx, output.{{ .ListTagsOutTagsElem }}{{ if .TagTypeIDElem }}, identifier{{ if .TagResTypeElem }}, resourceType{{ end }}{{ end }}), nil
{{- end }}
}

{{- if .IsDefaultListTags }}
// {{ .ListTagsFunc | Title }} lists {{ .ServicePackage }} service tags and set them in Context.
// It is called from outside this package.
func (p *servicePackage) {{ .ListTagsFunc | Title }}(ctx context.Context, meta any, identifier{{ if .TagResTypeElem }}, resourceType{{ end }} string) error {
	tags, err :=  {{ .ListTagsFunc }}(ctx, meta.(*conns.AWSClient).{{ .ProviderNameUpper }}Client(ctx), identifier{{ if .TagResTypeElem }}, resourceType{{ end }})

	if err != nil {
		return err
	}

	if inContext, ok := tftags.FromContext(ctx); ok {
		inContext.TagsOut = option.Some(tags)
	}

	return nil
}
{{- end }}
