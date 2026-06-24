// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
)

// @FrameworkListResource("aws_ssoadmin_application_grant")
func newApplicationGrantResourceAsListResource() list.ListResourceWithConfigure {
	return &applicationGrantListResource{}
}

var _ list.ListResource = &applicationGrantListResource{}

type applicationGrantListResource struct {
	applicationGrantResource
	framework.WithList
}

type listApplicationGrantModel struct {
	framework.WithRegionModel
	ApplicationARN types.String `tfsdk:"application_arn"`
}

func (l *applicationGrantListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"application_arn": listschema.StringAttribute{
				Required:    true,
				Description: "ARN of the application whose grants to list.",
			},
		},
	}
}

func (l *applicationGrantListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().SSOAdminClient(ctx)

	var query listApplicationGrantModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	applicationARN := query.ApplicationARN.ValueString()

	tflog.Info(ctx, "Listing SSO Application Grants", map[string]any{
		logging.ResourceAttributeKey("application_arn"): applicationARN,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		input := ssoadmin.ListApplicationGrantsInput{
			ApplicationArn: aws.String(applicationARN),
		}

		for item, err := range listApplicationGrants(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			grantType := string(item.GrantType)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("grant_type"), grantType)

			result := request.NewListResult(ctx)

			var data applicationGrantResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				idParts := []string{applicationARN, grantType}
				id, err := intflex.FlattenResourceId(idParts, applicationGrantIDPartCount, false)
				if err != nil {
					result.Diagnostics.AddError("Creating resource ID", err.Error())
					return
				}

				data.ID = types.StringValue(id)
				data.ApplicationARN = fwtypes.ARNValue(applicationARN)
				data.GrantType = fwtypes.StringEnumValue(item.GrantType)

				grantValue, diags := flattenApplicationGrant(ctx, item.Grant)
				result.Diagnostics.Append(diags...)
				if result.Diagnostics.HasError() {
					return
				}
				data.Grant = grantValue

				result.DisplayName = grantType
			})

			if result.Diagnostics.HasError() {
				yield(result)
				return
			}

			if !yield(result) {
				return
			}
		}
	}
}

func listApplicationGrants(ctx context.Context, conn *ssoadmin.Client, input *ssoadmin.ListApplicationGrantsInput) iter.Seq2[awstypes.GrantItem, error] {
	return func(yield func(awstypes.GrantItem, error) bool) {
		pages := ssoadmin.NewListApplicationGrantsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.GrantItem{}, fmt.Errorf("listing SSO Application Grants: %w", err))
				return
			}

			for _, item := range page.Grants {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
