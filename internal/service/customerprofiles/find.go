// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package customerprofiles

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/customerprofiles"
	"github.com/aws/aws-sdk-go-v2/service/customerprofiles/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindProfileByIdAndDomain(ctx context.Context, conn *customerprofiles.Client, profileId, domainName string) (*types.Profile, error) {
	input := &customerprofiles.SearchProfilesInput{
		DomainName: aws.String(domainName),
		KeyName:    aws.String("_profileId"),
		Values:     []string{profileId},
	}

	output, err := conn.SearchProfiles(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Items) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.Items); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return &output.Items[0], nil
}

func FindDomainByDomainName(ctx context.Context, conn *customerprofiles.Client, domainName string) (*customerprofiles.GetDomainOutput, error) {
	input := &customerprofiles.GetDomainInput{
		DomainName: aws.String(domainName),
	}

	output, err := conn.GetDomain(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
