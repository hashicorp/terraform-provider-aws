package validators

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type clusterIdentifierValidator struct{}

func (validator clusterIdentifierValidator) Description(_ context.Context) string {
	return "value must be a valid Cluster Identifier"
}

func (validator clusterIdentifierValidator) MarkdownDescription(ctx context.Context) string {
	return validator.Description(ctx)
}

func (validator clusterIdentifierValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	if err := validateClusterIdentifier(request.ConfigValue.ValueString()); err != nil {
		response.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			request.Path,
			validator.Description(ctx),
			err.Error(),
		))
		return
	}
}

func ClusterIdentifier() validator.String {
	return clusterIdentifierValidator{}
}

type clusterIdentifierPrefixValidator struct{}

func (validator clusterIdentifierPrefixValidator) Description(_ context.Context) string {
	return "value must be a valid Cluster Identifier Prefix"
}

func (validator clusterIdentifierPrefixValidator) MarkdownDescription(ctx context.Context) string {
	return validator.Description(ctx)
}

func (validator clusterIdentifierPrefixValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	if err := validateClusterIdentifierPrefix(request.ConfigValue.ValueString()); err != nil {
		response.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			request.Path,
			validator.Description(ctx),
			err.Error(),
		))
		return
	}
}

func ClusterIdentifierPrefix() validator.String {
	return clusterIdentifierPrefixValidator{}
}

type clusterFinalSnapshotIdentifierValidator struct{}

func (validator clusterFinalSnapshotIdentifierValidator) Description(_ context.Context) string {
	return "value must be a valid Final Snapshot Identifier"
}

func (validator clusterFinalSnapshotIdentifierValidator) MarkdownDescription(ctx context.Context) string {
	return validator.Description(ctx)
}

func (validator clusterFinalSnapshotIdentifierValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	if err := validateFinalSnapshotIdentifier(request.ConfigValue.ValueString()); err != nil {
		response.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			request.Path,
			validator.Description(ctx),
			err.Error(),
		))
		return
	}
}

func ClusterFinalSnapshotIdentifier() validator.String {
	return clusterFinalSnapshotIdentifierValidator{}
}

type evaluate struct {
	regex   func(string) bool
	message string
	isMatch bool
}

func (e evaluate) match(value string) error {
	if e.regex(value) == e.isMatch {
		return fmt.Errorf(e.message, value)
	}

	return nil
}

var (
	firstCharacterIsLetter = evaluate{
		regex:   regexp.MustCompile(`^[a-z]`).MatchString,
		message: "first character of %q must be a letter",
		isMatch: false,
	}
	onlyAlphanumeric = evaluate{
		regex:   regexp.MustCompile(`^[0-9A-Za-z-]+$`).MatchString,
		message: "only alphanumeric characters and hyphens allowed in %q",
		isMatch: false,
	}
	onlyLowercaseAlphanumeric = evaluate{
		regex:   regexp.MustCompile(`^[0-9a-z-]+$`).MatchString,
		message: "only lowercase alphanumeric characters and hyphens allowed in %q",
		isMatch: false,
	}
	noConsecutiveHyphens = evaluate{
		regex:   regexp.MustCompile(`--`).MatchString,
		message: "%q cannot contain two consecutive hyphens",
		isMatch: true,
	}
	cannotEndWithHyphen = evaluate{
		regex:   regexp.MustCompile(`-$`).MatchString,
		message: "%q cannot end with a hyphen",
		isMatch: true,
	}
)

func validateClusterIdentifier(value string) error {
	err := onlyLowercaseAlphanumeric.match(value)
	if err != nil {
		return err
	}
	err = noConsecutiveHyphens.match(value)
	if err != nil {
		return err
	}
	err = cannotEndWithHyphen.match(value)
	if err != nil {
		return err
	}

	return firstCharacterIsLetter.match(value)
}

func validateClusterIdentifierPrefix(value string) error {
	err := onlyLowercaseAlphanumeric.match(value)
	if err != nil {
		return err
	}
	err = noConsecutiveHyphens.match(value)
	if err != nil {
		return err
	}

	return firstCharacterIsLetter.match(value)
}

func validateFinalSnapshotIdentifier(value string) error {
	err := onlyAlphanumeric.match(value)
	if err != nil {
		return err
	}
	err = cannotEndWithHyphen.match(value)
	if err != nil {
		return err
	}

	return noConsecutiveHyphens.match(value)
}
