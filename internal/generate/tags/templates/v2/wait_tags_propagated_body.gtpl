// {{ .WaitTagsPropagatedFunc }} waits for {{ .ServicePackage }} service tags to be propagated.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func {{ .WaitTagsPropagatedFunc }}(ctx context.Context, tags tftags.KeyValueTags, checkFunc func() (bool, error)) error {
	tflog.Debug(ctx, "Waiting for tag propagation", map[string]any{
		names.AttrTags: tags,
	})

	opts := tfresource.WaitOpts{
		{{- if ne .WaitContinuousOccurence 0 }}
		ContinuousTargetOccurence: {{ .WaitContinuousOccurence }},
		{{- end }}
		{{- if ne .WaitDelay "" }}
		Delay: {{ .WaitDelay }},
		{{- end }}
		{{- if ne .WaitMinTimeout "" }}
		MinTimeout: {{ .WaitMinTimeout }},
		{{- end }}
		{{- if ne .WaitPollInterval "" }}
		PollInterval: {{ .WaitPollInterval }},
		{{- end }}
	}

	return tfresource.WaitUntil(ctx, {{ .WaitTimeout }}, checkFunc, opts)
}

// checkFunc returns a function that checks if the tags are propagated.
func checkFunc(ctx context.Context, conn {{ .ClientType }}, tags tftags.KeyValueTags, id string, optFns ...func(*{{ .AWSService }}.Options)) func() (bool, error) {
    return func() (bool, error) {
        output, err := listTags(ctx, conn, id, optFns...)

        if tfresource.NotFound(err) {
            return false, nil
        }

        if err != nil {
            return false, err
        }

        if inContext, ok := tftags.FromContext(ctx); ok {
            tags = tags.IgnoreConfig(inContext.IgnoreConfig)
            output = output.IgnoreConfig(inContext.IgnoreConfig)
        }

        {{- if .UpdateTagsForResource }}
        return output.ContainsAll(tags), nil
        {{- else }}
        return output.Equal(tags), nil
        {{- end }}
    }
}
