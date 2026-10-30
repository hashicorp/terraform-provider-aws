// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecr

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @FrameworkListResource("aws_ecr_signing_configuration")
func newSigningConfigurationResourceAsListResource() list.ListResourceWithConfigure {
	return &signingConfigurationListResource{}
}

var _ list.ListResource = &signingConfigurationListResource{}

type signingConfigurationListResource struct {
	signingConfigurationResource
	framework.WithList
}

// List returns at most one result: a registry has a single signing configuration per Region.
func (l *signingConfigurationListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().ECRClient(ctx)

	tflog.Info(ctx, "Listing ECR Signing Configurations")

	stream.Results = func(yield func(list.ListResult) bool) {
		output, err := findSigningConfiguration(ctx, conn)
		if retry.NotFound(err) {
			return
		}
		if err != nil {
			yield(fwdiag.NewListResultErrorDiagnostic(err))
			return
		}

		result := request.NewListResult(ctx)
		var data signingConfigurationResourceModel

		l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
			if request.IncludeResource {
				result.Diagnostics.Append(l.flatten(ctx, output, &data)...)
				if result.Diagnostics.HasError() {
					return
				}
			}

			result.DisplayName = aws.ToString(output.RegistryId)
		})

		yield(result)
	}
}
