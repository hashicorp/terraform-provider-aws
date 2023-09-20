var _ validator.String = arnValidator{}

type arnValidator struct {
}

// Description describes the validation in plain text formatting.
func (validator arnValidator) Description(_ context.Context) string {
	return "String must be a valid arn"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (validator arnValidator) MarkdownDescription(ctx context.Context) string {
	return validator.Description(ctx)
}

func (v arnValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue.ValueString()

    parsedARN, err := arn.Parse(value)

    if err != nil {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			fmt.Sprintf("(%s) is an invalid ARN: %s", value, err),
			fmt.Sprintf("%s", value),
		))
        return
    }

    if parsedARN.Partition == "" {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			fmt.Sprintf("(%s) is an invalid ARN: missing partition value", value),
			fmt.Sprintf("%s", value),
		))
        return
    } else if !partitionRegexp.MatchString(parsedARN.Partition) {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			fmt.Sprintf("(%s) is an invalid ARN: invalid partition value (expecting to match regular expression: %s)", value, partitionRegexp),
			fmt.Sprintf("%s", value),
		))
        return
    }

    if parsedARN.Region != "" && !regionRegexp.MatchString(parsedARN.Region) {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			fmt.Sprintf("(%s) is an invalid ARN: invalid region value (expecting to match regular expression: %s)", value, regionRegexp),
			fmt.Sprintf("%s", value),
		))
        return
    }

    if parsedARN.AccountID != "" && !accountIDRegexp.MatchString(parsedARN.AccountID) {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			fmt.Sprintf("(%s) is an invalid ARN: invalid account ID value (expecting to match regular expression: %s)", value, accountIDRegexp),
			fmt.Sprintf("%s", value),
		))
        return
    }

    if parsedARN.Resource == "" {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			fmt.Sprintf("(%s) is an invalid ARN: missing resource value", value),
			fmt.Sprintf("%s", value),
		))
        return
    }
}

func validateArnFramework() validator.String {
	return arnValidator{}
}
