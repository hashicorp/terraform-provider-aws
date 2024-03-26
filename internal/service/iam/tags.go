// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !generate
// +build !generate

package iam

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/types/option"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Custom IAM tag service update functions using the same format as generated code.

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

func instanceProfileCreateTags(ctx context.Context, conn iamiface.IAMAPI, identifier string, tags []*iam.Tag) error {
	if len(tags) == 0 {
		return nil
	}

	return instanceProfileUpdateTags(ctx, conn, identifier, nil, KeyValueTags(ctx, tags))
}

func instanceProfileKeyValueTags(ctx context.Context, conn *iam.IAM, identifier string) (tftags.KeyValueTags, error) {
	tags, err := instanceProfileTags(ctx, conn, identifier)
	if err != nil {
		return tftags.New(ctx, nil), fmt.Errorf("listing tags for resource (%s): %w", identifier, err)
	}

	return KeyValueTags(ctx, tags), nil
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

func openIDConnectProviderCreateTags(ctx context.Context, conn iamiface.IAMAPI, identifier string, tags []*iam.Tag) error {
	if len(tags) == 0 {
		return nil
	}

	return openIDConnectProviderUpdateTags(ctx, conn, identifier, nil, KeyValueTags(ctx, tags))
}

func openIDConnectProviderKeyValueTags(ctx context.Context, conn *iam.IAM, identifier string) (tftags.KeyValueTags, error) {
	tags, err := openIDConnectProviderTags(ctx, conn, identifier)
	if err != nil {
		return tftags.New(ctx, nil), fmt.Errorf("listing tags for resource (%s): %w", identifier, err)
	}

	return KeyValueTags(ctx, tags), nil
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

func policyCreateTags(ctx context.Context, conn iamiface.IAMAPI, identifier string, tags []*iam.Tag) error {
	if len(tags) == 0 {
		return nil
	}

	return policyUpdateTags(ctx, conn, identifier, nil, KeyValueTags(ctx, tags))
}

func policyKeyValueTags(ctx context.Context, conn *iam.IAM, identifier string) (tftags.KeyValueTags, error) {
	tags, err := policyTags(ctx, conn, identifier)
	if err != nil {
		return tftags.New(ctx, nil), fmt.Errorf("listing tags for resource (%s): %w", identifier, err)
	}

	return KeyValueTags(ctx, tags), nil
}

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

func roleCreateTags(ctx context.Context, conn iamiface.IAMAPI, identifier string, tags []*iam.Tag) error {
	if len(tags) == 0 {
		return nil
	}

	return roleUpdateTags(ctx, conn, identifier, nil, KeyValueTags(ctx, tags))
}

func roleKeyValueTags(ctx context.Context, conn *iam.IAM, identifier string) (tftags.KeyValueTags, error) {
	tags, err := roleTags(ctx, conn, identifier)
	if err != nil {
		return tftags.New(ctx, nil), fmt.Errorf("listing tags for resource (%s): %w", identifier, err)
	}

	return KeyValueTags(ctx, tags), nil
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

func samlProviderCreateTags(ctx context.Context, conn iamiface.IAMAPI, identifier string, tags []*iam.Tag) error {
	if len(tags) == 0 {
		return nil
	}

	return samlProviderUpdateTags(ctx, conn, identifier, nil, KeyValueTags(ctx, tags))
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

func serverCertificateCreateTags(ctx context.Context, conn iamiface.IAMAPI, identifier string, tags []*iam.Tag) error {
	if len(tags) == 0 {
		return nil
	}

	return serverCertificateUpdateTags(ctx, conn, identifier, nil, KeyValueTags(ctx, tags))
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

func userCreateTags(ctx context.Context, conn iamiface.IAMAPI, identifier string, tags []*iam.Tag) error {
	if len(tags) == 0 {
		return nil
	}

	return userUpdateTags(ctx, conn, identifier, nil, KeyValueTags(ctx, tags))
}

func userKeyValueTags(ctx context.Context, conn *iam.IAM, identifier string) (tftags.KeyValueTags, error) {
	tags, err := userTags(ctx, conn, identifier)
	if err != nil {
		return tftags.New(ctx, nil), fmt.Errorf("listing tags for resource (%s): %w", identifier, err)
	}

	return KeyValueTags(ctx, tags), nil
}

// virtualMFADeviceUpdateTags updates IAM Virtual MFA Device tags.
// The identifier is the Virtual MFA Device ARN.
func virtualMFADeviceUpdateTags(ctx context.Context, conn iamiface.IAMAPI, identifier string, oldTagsMap, newTagsMap any) error {
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

func virtualMFADeviceCreateTags(ctx context.Context, conn iamiface.IAMAPI, identifier string, tags []*iam.Tag) error {
	if len(tags) == 0 {
		return nil
	}

	return virtualMFADeviceUpdateTags(ctx, conn, identifier, nil, KeyValueTags(ctx, tags))
}

func virtualMFADeviceKeyValueTags(ctx context.Context, conn *iam.IAM, identifier string) (tftags.KeyValueTags, error) {
	tags, err := virtualMFADeviceTags(ctx, conn, identifier)
	if err != nil {
		return tftags.New(ctx, nil), fmt.Errorf("listing tags for resource (%s): %w", identifier, err)
	}

	return KeyValueTags(ctx, tags), nil
}

// updateTags updates iam service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func updateTags(ctx context.Context, conn iamiface.IAMAPI, identifier, resourceType string, oldTagsMap, newTagsMap any) error {
	switch resourceType {
	case "InstanceProfile":
		return instanceProfileUpdateTags(ctx, conn, identifier, oldTagsMap, newTagsMap)
	case "OIDCProvider":
		return openIDConnectProviderUpdateTags(ctx, conn, identifier, oldTagsMap, newTagsMap)
	case "Policy":
		return policyUpdateTags(ctx, conn, identifier, oldTagsMap, newTagsMap)
	case "Role":
		return roleUpdateTags(ctx, conn, identifier, oldTagsMap, newTagsMap)
	case "ServiceLinkedRole":
		_, roleName, _, err := DecodeServiceLinkedRoleID(identifier)
		if err != nil {
			return err
		}
		return roleUpdateTags(ctx, conn, roleName, oldTagsMap, newTagsMap)
	case "SAMLProvider":
		return samlProviderUpdateTags(ctx, conn, identifier, oldTagsMap, newTagsMap)
	case "ServerCertificate":
		return serverCertificateUpdateTags(ctx, conn, identifier, oldTagsMap, newTagsMap)
	case "User":
		return userUpdateTags(ctx, conn, identifier, oldTagsMap, newTagsMap)
	case "VirtualMFADevice":
		return virtualMFADeviceUpdateTags(ctx, conn, identifier, oldTagsMap, newTagsMap)
	}

	return fmt.Errorf("unsupported resource type: %s", resourceType)
}

// ListTags lists iam service tags and set them in Context.
// It is called from outside this package.
func (p *servicePackage) ListTags(ctx context.Context, meta any, identifier, resourceType string) error {
	var (
		tags tftags.KeyValueTags
		err  error
	)
	switch resourceType {
	case "InstanceProfile":
		tags, err = instanceProfileKeyValueTags(ctx, meta.(*conns.AWSClient).IAMConn(ctx), identifier)

	case "OIDCProvider":
		tags, err = openIDConnectProviderKeyValueTags(ctx, meta.(*conns.AWSClient).IAMConn(ctx), identifier)

	case "Policy":
		tags, err = policyKeyValueTags(ctx, meta.(*conns.AWSClient).IAMConn(ctx), identifier)

	case "Role":
		tags, err = roleKeyValueTags(ctx, meta.(*conns.AWSClient).IAMConn(ctx), identifier)

	case "ServiceLinkedRole":
		var roleName string
		_, roleName, _, err = DecodeServiceLinkedRoleID(identifier)
		if err != nil {
			return err
		}
		tags, err = roleKeyValueTags(ctx, meta.(*conns.AWSClient).IAMConn(ctx), roleName)

	case "User":
		tags, err = userKeyValueTags(ctx, meta.(*conns.AWSClient).IAMConn(ctx), identifier)

	case "VirtualMFADevice":
		tags, err = virtualMFADeviceKeyValueTags(ctx, meta.(*conns.AWSClient).IAMConn(ctx), identifier)

	default:
		return nil
	}

	if err != nil {
		return err
	}

	if inContext, ok := tftags.FromContext(ctx); ok {
		inContext.TagsOut = option.Some(tags)
	}

	return nil
}

// UpdateTags updates iam service tags.
// It is called from outside this package.
func (p *servicePackage) UpdateTags(ctx context.Context, meta any, identifier, resourceType string, oldTags, newTags any) error {
	return updateTags(ctx, meta.(*conns.AWSClient).IAMConn(ctx), identifier, resourceType, oldTags, newTags)
}
