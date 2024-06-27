// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package grafana

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/managedgrafana"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindSamlConfigurationByID(ctx context.Context, conn *managedgrafana.ManagedGrafana, id string) (*managedgrafana.SamlAuthentication, error) {
	input := &managedgrafana.DescribeWorkspaceAuthenticationInput{
		WorkspaceId: aws.String(id),
	}

	output, err := conn.DescribeWorkspaceAuthenticationWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, managedgrafana.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Authentication == nil || output.Authentication.Saml == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if status := aws.StringValue(output.Authentication.Saml.Status); status == managedgrafana.SamlConfigurationStatusNotConfigured {
		return nil, &retry.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return output.Authentication.Saml, nil
}
