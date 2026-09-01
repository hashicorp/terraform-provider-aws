// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package organizations

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
)

// @FrameworkListResource("aws_organizations_aws_service_access")
func newAWSServiceAccessResourceAsListResource() list.ListResourceWithConfigure { // nosemgrep:ci.aws-in-func-name
	return &awsServiceAccessListResource{}
}

var _ list.ListResource = &awsServiceAccessListResource{}

type awsServiceAccessListResource struct {
	awsServiceAccessResource
	framework.WithList
}

func (l *awsServiceAccessListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().OrganizationsClient(ctx)

	var query listAWSServiceAccessModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		var input organizations.ListAWSServiceAccessForOrganizationInput
		for item, err := range listEnabledServicePrincipals(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			sp := aws.ToString(item.ServicePrincipal)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("service_principal"), sp)

			result := request.NewListResult(ctx)

			var data awsServiceAccessResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				result.Diagnostics.Append(l.flatten(ctx, &item, &data)...)
				if result.Diagnostics.HasError() {
					return
				}

				result.DisplayName = sp
			})

			if !yield(result) {
				return
			}
		}
	}
}

type listAWSServiceAccessModel struct{}
