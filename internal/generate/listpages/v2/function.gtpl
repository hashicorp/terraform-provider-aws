func {{ .Name }}Pages(ctx context.Context, conn *{{ .AWSService }}.Client, input {{ .ParamType }}, fn func({{ .ResultType }}, bool) bool, optFns ...func(*{{ .AWSService }}.Options)) error {
	for {
		output, err := conn.{{ .AWSName }}(ctx, input, optFns...)
		if err != nil {
			return smarterr.NewError(err)
		}

		lastPage := aws.ToString(output.{{ .OutputPaginator }}) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.{{ .InputPaginator }} = output.{{ .OutputPaginator }}
	}
	return nil
}
