func {{ .Name }}Pages(ctx context.Context, conn *{{ .AWSService }}.Client, input {{ .ParamType }}, fn func({{ .ResultType }}, bool) bool) error {
	for {
		output, err := conn.{{ .AWSName }}(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.ToString(output.{{ .OutputPaginator }}) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.{{ .InputPaginator }} = output.{{ .OutputPaginator }}
	}
	return nil
}
