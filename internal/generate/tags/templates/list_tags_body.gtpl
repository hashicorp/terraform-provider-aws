// {{ .ListTagsFunc }} lists {{ .ServicePackage }} service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func {{ .ListTagsFunc }}(ctx context.Context, conn {{ .ClientType }}, identifier{{ if .TagResTypeElem }}, resourceType{{ end }} string, optFns ...func(*{{ .AWSService }}.Options)) (tftags.KeyValueTags, error) {
	input := {{ .AWSService }}.{{ .ListTagsOp }}Input{
		{{- if .ListTagsInFiltIDName }}
		Filters: []awstypes.Filter{
			{
				Name:   aws.String("{{ .ListTagsInFiltIDName }}"),
				Values: []string{identifier},
			},
		},
		{{- else }}
		{{- if .ListTagsInIDNeedValueSlice }}
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
    {{- if .RetryTagOps }}
		output, err := tfresource.RetryWhenIsAErrorMessageContains[*{{ .AWSService }}.{{ .RetryTagsListTagsType }}, *{{ .RetryErrorCode }}](ctx, {{ .RetryTimeout }},
			func(ctx context.Context) (*{{ .AWSService }}.{{ .RetryTagsListTagsType }}, error) {
				var output []awstypes.{{ or .TagType2 .TagType }}

				pages := {{ .AWSService }}.New{{ .ListTagsOp }}Paginator(conn, &input)
				for pages.HasMorePages() {
					page, err := pages.NextPage(ctx, optFns...)

				{{ if and ( .ParentNotFoundErrCode ) ( .ParentNotFoundErrMsg ) }}
					if tfawserr.ErrMessageContains(err, "{{ .ParentNotFoundErrCode }}", "{{ .ParentNotFoundErrMsg }}") {
						return nil, smarterr.NewError(&retry.NotFoundError{
							LastError: err,
						})
					}
				{{- else if ( .ParentNotFoundErrCode ) }}
					if tfawserr.ErrCodeEquals(err, "{{ .ParentNotFoundErrCode }}") {
						return nil, smarterr.NewError(&retry.NotFoundError{
							LastError: err,
						})
					}
				{{- end }}

					if err != nil {
						return tftags.New(ctx, nil), smarterr.NewError(err)
					}

					output = append(output, page.{{ .ListTagsOutTagsElem }}...)
				}
			},
			"{{ .RetryErrorMessage }}",
		)
	{{ else }}
		{{ if .ServiceTagsMap }}
			output := make(map[string]string)
		{{- else }}
			var output []awstypes.{{ or .TagType2 .TagType }}
		{{- end }}

		{{ if .ListTagsOpPaginatorCustom }}
			err := {{ .ListTagsOp | FirstLower }}Pages(ctx, conn, &input, func(page *{{ .AWSService }}.{{ .ListTagsOp }}Output, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

			{{ if .ServiceTagsMap }}
				maps.Copy(output, page.{{ .ListTagsOutTagsElem }})
			{{- else }}
				output = append(output, page.{{ .ListTagsOutTagsElem }}...)
			{{- end }}

				return !lastPage
			}, optFns...)

			{{ if and ( .ParentNotFoundErrCode ) ( .ParentNotFoundErrMsg ) }}
				if tfawserr.ErrMessageContains(err, "{{ .ParentNotFoundErrCode }}", "{{ .ParentNotFoundErrMsg }}") {
					return nil, &retry.NotFoundError{
						LastError: err,
					}
				}
			{{- else if ( .ParentNotFoundErrCode ) }}
				if tfawserr.ErrCodeEquals(err, "{{ .ParentNotFoundErrCode }}") {
					return nil, &retry.NotFoundError{
						LastError: err,
					}
				}
			{{- end }}

			if err != nil {
				return tftags.New(ctx, nil), err
			}
		{{- else }}
			pages := {{ .AWSService }}.New{{ .ListTagsOp }}Paginator(conn, &input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx, optFns...)

			{{ if and ( .ParentNotFoundErrCode ) ( .ParentNotFoundErrMsg ) }}
				if tfawserr.ErrMessageContains(err, "{{ .ParentNotFoundErrCode }}", "{{ .ParentNotFoundErrMsg }}") {
					return nil, smarterr.NewError(&retry.NotFoundError{
						LastError: err,
					})
				}
			{{- else if ( .ParentNotFoundErrCode ) }}
				if tfawserr.ErrCodeEquals(err, "{{ .ParentNotFoundErrCode }}") {
					return nil, smarterr.NewError(&retry.NotFoundError{
						LastError: err,
					})
				}
			{{- end }}

				if err != nil {
					return tftags.New(ctx, nil), smarterr.NewError(err)
				}

			{{ if .ServiceTagsMap }}
				maps.Copy(output, page.{{ .ListTagsOutTagsElem }})
			{{- else }}
				output = append(output, page.{{ .ListTagsOutTagsElem }}...)
			{{- end }}
			}
		{{- end }}
	{{- end }}

	return {{ .KeyValueTagsFunc }}(ctx, output{{ if .TagTypeIDElem }}, identifier{{ if .TagResTypeElem }}, resourceType{{ end }}{{ end }}), nil
{{- else }}
    {{ if .RetryTagOps }}
		output, err := tfresource.RetryWhenIsAErrorMessageContains[*{{ .AWSService }}.{{ .RetryTagsListTagsType }}, *{{ .RetryErrorCode }}](ctx, {{ .RetryTimeout }},
			func(ctx context.Context) (*{{ .AWSService }}.{{ .RetryTagsListTagsType }}, error) {
				return conn.{{ .ListTagsOp }}(ctx, &input, optFns...)
			},
			"{{ .RetryErrorMessage }}",
		)
	{{- else }}
		output, err := conn.{{ .ListTagsOp }}(ctx, &input, optFns...)
	{{- end }}

	{{ if and ( .ParentNotFoundErrCode ) ( .ParentNotFoundErrMsg ) }}
		if tfawserr.ErrMessageContains(err, "{{ .ParentNotFoundErrCode }}", "{{ .ParentNotFoundErrMsg }}") {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}
	{{- else if ( .ParentNotFoundErrCode ) }}
		if tfawserr.ErrCodeEquals(err, "{{ .ParentNotFoundErrCode }}") {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}
	{{- end }}

	if err != nil {
		return tftags.New(ctx, nil), smarterr.NewError(err)
	}

	return {{ .KeyValueTagsFunc }}(ctx, output.{{ .ListTagsOutTagsElem }}{{ if .TagTypeIDElem }}, identifier{{ if .TagResTypeElem }}, resourceType{{ end }}{{ end }}), nil
{{- end }}
}

{{- if .IsDefaultListTags }}
// {{ .ListTagsFunc | Title }} lists {{ .ServicePackage }} service tags and set them in Context.
// It is called from outside this package.
{{- if .TagResTypeElem }}
{{- if .TagResTypeIsAccountID }}
func (p *servicePackage) {{ .ListTagsFunc | Title }}(ctx context.Context, meta any, identifier string) error {
	c := meta.(*conns.AWSClient)
	tags, err :=  {{ .ListTagsFunc }}(ctx, c.{{ .ProviderNameUpper }}Client(ctx), identifier, c.AccountID(ctx))
{{- else }}
func (p *servicePackage) {{ .ListTagsFunc | Title }}(ctx context.Context, meta any, identifier, resourceType string) error {
	tags, err :=  {{ .ListTagsFunc }}(ctx, meta.(*conns.AWSClient).{{ .ProviderNameUpper }}Client(ctx), identifier, resourceType)
{{- end }}
{{- else }}
func (p *servicePackage) {{ .ListTagsFunc | Title }}(ctx context.Context, meta any, identifier string) error {
	tags, err :=  {{ .ListTagsFunc }}(ctx, meta.(*conns.AWSClient).{{ .ProviderNameUpper }}Client(ctx), identifier)
{{- end }}

	if err != nil {
		return smarterr.NewError(err)
	}

	if inContext, ok := tftags.FromContext(ctx); ok {
		inContext.TagsOut = option.Some(tags)
	}

	return nil
}
{{- end }}
