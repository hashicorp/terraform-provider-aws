
// {{ .Name }}PaginatorOptions is the paginator options for
// {{ .AWSName }}
type {{ .Name }}PaginatorOptions struct {
	// The maximum number of items to return for this request. To get the next page of
	// items, make another request with the token returned in the output. For more
	// information, see [Pagination].
	//
	// [Pagination]: https://docs.aws.amazon.com/AWSEC2/latest/APIReference/Query-Requests.html#api-pagination
	Limit int32

	// Set to true if pagination should stop if the service returns a pagination token
	// that matches the most recent token provided to the service.
	StopOnDuplicateToken bool
}

// {{ .Name }}Paginator is a paginator for {{ .AWSName }}
type {{ .Name }}Paginator struct {
	options   {{ .Name }}PaginatorOptions
	client    {{ .Name }}APIClient
	params    *ec2.{{ .AWSName }}Input
	nextToken *string
	firstPage bool
}

// new{{ .AWSName }}Paginator returns a new {{ .AWSName }}Paginator
{{ if .NoSemgrep -}}
// nosemgrep:ci.{{ .NoSemgrep }}-in-func-name
{{- end }}
func new{{ .AWSName }}Paginator(client {{ .Name }}APIClient, params *ec2.{{ .AWSName }}Input, optFns ...func(*{{ .Name }}PaginatorOptions)) *{{ .Name }}Paginator {
	if params == nil {
		params = &ec2.{{ .AWSName }}Input{}
	}

	options := {{ .Name }}PaginatorOptions{}
	if params.MaxResults != nil {
		options.Limit = *params.MaxResults // nosemgrep:ci.semgrep.aws.prefer-pointer-conversion-assignment
	}

	for _, fn := range optFns {
		fn(&options)
	}

	return &{{ .Name }}Paginator{
		options:   options,
		client:    client,
		params:    params,
		firstPage: true,
		nextToken: params.NextToken,
	}
}

// HasMorePages returns a boolean indicating whether more pages are available
func (p *{{ .Name }}Paginator) HasMorePages() bool {
	return p.firstPage || (p.nextToken != nil && len(*p.nextToken) != 0)
}

// NextPage retrieves the next {{ .AWSName }} page.
func (p *{{ .Name }}Paginator) NextPage(ctx context.Context, optFns ...func(*ec2.Options)) (*ec2.{{ .AWSName }}Output, error) {
	if !p.HasMorePages() {
		return nil, fmt.Errorf("no more pages available")
	}

	params := *p.params // nosemgrep:ci.semgrep.aws.prefer-pointer-conversion-assignment
	params.NextToken = p.nextToken

	var limit *int32
	if p.options.Limit > 0 {
		limit = &p.options.Limit
	}
	params.MaxResults = limit

	result, err := p.client.{{ .AWSName }}(ctx, &params, optFns...)
	if err != nil {
		return nil, err
	}
	p.firstPage = false

	prevToken := p.nextToken
	p.nextToken = result.NextToken

	if p.options.StopOnDuplicateToken &&
		prevToken != nil &&
		p.nextToken != nil &&
		*prevToken == *p.nextToken { // nosemgrep:ci.semgrep.aws.prefer-pointer-conversion-conditional
		p.nextToken = nil
	}

	return result, nil
}

// {{ .Name }}APIClient is a client that implements the
// {{ .AWSName }} operation.
type {{ .Name }}APIClient interface {
	{{ .AWSName }}(context.Context, *ec2.{{ .AWSName }}Input, ...func(*ec2.Options)) (*ec2.{{ .AWSName }}Output, error)
}

var _ {{ .Name }}APIClient = (*ec2.Client)(nil)
