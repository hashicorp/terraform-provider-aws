//go:build !generate
// +build !generate

package iam

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Custom IAM tag service update functions using the same format as generated code.

// roleUpdateTags updates IAM role tags.
// The identifier is the role name.
func roleUpdateTags(ctx context.Context, conn iamiface.IAMAPI, identifier string, oldTagsMap, newTagsMap any) error {
	oldTags := tftags.New(ctx, oldTagsMap)
	newTags := tftags.New(ctx, newTagsMap)

	if removedTags := oldTags.Removed(newTags).IgnoreSystem(names.IAM); len(removedTags) > 0 {
		input := &iam.UntagRoleInput{
			RoleName: aws.String(identifier),
			TagKeys:  aws.StringSlice(removedTags.Keys()),
		}

		_, err := conn.UntagRoleWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("untagging resource (%s): %w", identifier, err)
		}
	}

	if updatedTags := oldTags.Updated(newTags).IgnoreSystem(names.IAM); len(updatedTags) > 0 {
		input := &iam.TagRoleInput{
			RoleName: aws.String(identifier),
			Tags:     Tags(updatedTags),
		}

		_, err := conn.TagRoleWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("tagging resource (%s): %w", identifier, err)
		}
	}

	return nil
}

// userUpdateTags updates IAM user tags.
// The identifier is the user name.
func userUpdateTags(ctx context.Context, conn iamiface.IAMAPI, identifier string, oldTagsMap, newTagsMap any) error {
	oldTags := tftags.New(ctx, oldTagsMap)
	newTags := tftags.New(ctx, newTagsMap)

	if removedTags := oldTags.Removed(newTags).IgnoreSystem(names.IAM); len(removedTags) > 0 {
		input := &iam.UntagUserInput{
			UserName: aws.String(identifier),
			TagKeys:  aws.StringSlice(removedTags.Keys()),
		}

		_, err := conn.UntagUserWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("untagging resource (%s): %w", identifier, err)
		}
	}

	if updatedTags := oldTags.Updated(newTags).IgnoreSystem(names.IAM); len(updatedTags) > 0 {
		input := &iam.TagUserInput{
			UserName: aws.String(identifier),
			Tags:     Tags(updatedTags),
		}

		_, err := conn.TagUserWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("tagging resource (%s): %w", identifier, err)
		}
	}

	return nil
}

// instanceProfileUpdateTags updates IAM Instance Profile tags.
// The identifier is the Instance Profile name.
func instanceProfileUpdateTags(ctx context.Context, conn iamiface.IAMAPI, identifier string, oldTagsMap, newTagsMap any) error {
	oldTags := tftags.New(ctx, oldTagsMap)
	newTags := tftags.New(ctx, newTagsMap)

	if removedTags := oldTags.Removed(newTags).IgnoreSystem(names.IAM); len(removedTags) > 0 {
		input := &iam.UntagInstanceProfileInput{
			InstanceProfileName: aws.String(identifier),
			TagKeys:             aws.StringSlice(removedTags.Keys()),
		}

		_, err := conn.UntagInstanceProfileWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("untagging resource (%s): %w", identifier, err)
		}
	}

	if updatedTags := oldTags.Updated(newTags).IgnoreSystem(names.IAM); len(updatedTags) > 0 {
		input := &iam.TagInstanceProfileInput{
			InstanceProfileName: aws.String(identifier),
			Tags:                Tags(updatedTags),
		}

		_, err := conn.TagInstanceProfileWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("tagging resource (%s): %w", identifier, err)
		}
	}

	return nil
}

// openIDConnectProviderUpdateTags updates IAM OpenID Connect Provider tags.
// The identifier is the OpenID Connect Provider ARN.
func openIDConnectProviderUpdateTags(ctx context.Context, conn iamiface.IAMAPI, identifier string, oldTagsMap, newTagsMap any) error {
	oldTags := tftags.New(ctx, oldTagsMap)
	newTags := tftags.New(ctx, newTagsMap)

	if removedTags := oldTags.Removed(newTags).IgnoreSystem(names.IAM); len(removedTags) > 0 {
		input := &iam.UntagOpenIDConnectProviderInput{
			OpenIDConnectProviderArn: aws.String(identifier),
			TagKeys:                  aws.StringSlice(removedTags.Keys()),
		}

		_, err := conn.UntagOpenIDConnectProviderWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("untagging resource (%s): %w", identifier, err)
		}
	}

	if updatedTags := oldTags.Updated(newTags).IgnoreSystem(names.IAM); len(updatedTags) > 0 {
		input := &iam.TagOpenIDConnectProviderInput{
			OpenIDConnectProviderArn: aws.String(identifier),
			Tags:                     Tags(updatedTags),
		}

		_, err := conn.TagOpenIDConnectProviderWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("tagging resource (%s): %w", identifier, err)
		}
	}

	return nil
}

// policyUpdateTags updates IAM Policy tags.
// The identifier is the Policy ARN.
func policyUpdateTags(ctx context.Context, conn iamiface.IAMAPI, identifier string, oldTagsMap, newTagsMap any) error {
	oldTags := tftags.New(ctx, oldTagsMap)
	newTags := tftags.New(ctx, newTagsMap)

	if removedTags := oldTags.Removed(newTags).IgnoreSystem(names.IAM); len(removedTags) > 0 {
		input := &iam.UntagPolicyInput{
			PolicyArn: aws.String(identifier),
			TagKeys:   aws.StringSlice(removedTags.Keys()),
		}

		_, err := conn.UntagPolicyWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("untagging resource (%s): %w", identifier, err)
		}
	}

	if updatedTags := oldTags.Updated(newTags).IgnoreSystem(names.IAM); len(updatedTags) > 0 {
		input := &iam.TagPolicyInput{
			PolicyArn: aws.String(identifier),
			Tags:      Tags(updatedTags),
		}

		_, err := conn.TagPolicyWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("tagging resource (%s): %w", identifier, err)
		}
	}

	return nil
}

// samlProviderUpdateTags updates IAM SAML Provider tags.
// The identifier is the SAML Provider ARN.
func samlProviderUpdateTags(ctx context.Context, conn iamiface.IAMAPI, identifier string, oldTagsMap, newTagsMap any) error {
	oldTags := tftags.New(ctx, oldTagsMap)
	newTags := tftags.New(ctx, newTagsMap)

	if removedTags := oldTags.Removed(newTags).IgnoreSystem(names.IAM); len(removedTags) > 0 {
		input := &iam.UntagSAMLProviderInput{
			SAMLProviderArn: aws.String(identifier),
			TagKeys:         aws.StringSlice(removedTags.Keys()),
		}

		_, err := conn.UntagSAMLProviderWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("untagging resource (%s): %w", identifier, err)
		}
	}

	if updatedTags := oldTags.Updated(newTags).IgnoreSystem(names.IAM); len(updatedTags) > 0 {
		input := &iam.TagSAMLProviderInput{
			SAMLProviderArn: aws.String(identifier),
			Tags:            Tags(updatedTags),
		}

		_, err := conn.TagSAMLProviderWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("tagging resource (%s): %w", identifier, err)
		}
	}

	return nil
}

// serverCertificateUpdateTags updates IAM Server Certificate tags.
// The identifier is the Server Certificate name.
func serverCertificateUpdateTags(ctx context.Context, conn iamiface.IAMAPI, identifier string, oldTagsMap, newTagsMap any) error {
	oldTags := tftags.New(ctx, oldTagsMap)
	newTags := tftags.New(ctx, newTagsMap)

	if removedTags := oldTags.Removed(newTags).IgnoreSystem(names.IAM); len(removedTags) > 0 {
		input := &iam.UntagServerCertificateInput{
			ServerCertificateName: aws.String(identifier),
			TagKeys:               aws.StringSlice(removedTags.Keys()),
		}

		_, err := conn.UntagServerCertificateWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("untagging resource (%s): %w", identifier, err)
		}
	}

	if updatedTags := oldTags.Updated(newTags).IgnoreSystem(names.IAM); len(updatedTags) > 0 {
		input := &iam.TagServerCertificateInput{
			ServerCertificateName: aws.String(identifier),
			Tags:                  Tags(updatedTags),
		}

		_, err := conn.TagServerCertificateWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("tagging resource (%s): %w", identifier, err)
		}
	}

	return nil
}

// virtualMFAUpdateTags updates IAM Virtual MFA Device tags.
// The identifier is the Virtual MFA Device ARN.
func virtualMFAUpdateTags(ctx context.Context, conn iamiface.IAMAPI, identifier string, oldTagsMap, newTagsMap any) error {
	oldTags := tftags.New(ctx, oldTagsMap)
	newTags := tftags.New(ctx, newTagsMap)

	if removedTags := oldTags.Removed(newTags).IgnoreSystem(names.IAM); len(removedTags) > 0 {
		input := &iam.UntagMFADeviceInput{
			SerialNumber: aws.String(identifier),
			TagKeys:      aws.StringSlice(removedTags.Keys()),
		}

		_, err := conn.UntagMFADeviceWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("untagging resource (%s): %w", identifier, err)
		}
	}

	if updatedTags := oldTags.Updated(newTags).IgnoreSystem(names.IAM); len(updatedTags) > 0 {
		input := &iam.TagMFADeviceInput{
			SerialNumber: aws.String(identifier),
			Tags:         Tags(updatedTags),
		}

		_, err := conn.TagMFADeviceWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("tagging resource (%s): %w", identifier, err)
		}
	}

	return nil
}
