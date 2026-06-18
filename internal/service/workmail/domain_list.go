// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package workmail

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/workmail"
	awstypes "github.com/aws/aws-sdk-go-v2/service/workmail/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @FrameworkListResource("aws_workmail_domain")
func newDomainResourceAsListResource() list.ListResourceWithConfigure {
	return &domainListResource{}
}

var _ list.ListResource = &domainListResource{}

type domainListResource struct {
	domainResource
	framework.WithList
}

func (l *domainListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"organization_id": listschema.StringAttribute{
				Required:    true,
				Description: "ID of the WorkMail organization to list domains from.",
			},
		},
	}
}

func (l *domainListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().WorkMailClient(ctx)

	var query listDomainModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	organizationID := query.OrganizationID.ValueString()

	tflog.Info(ctx, "Listing WorkMail Domains", map[string]any{
		"organization_id": organizationID,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		input := workmail.ListMailDomainsInput{
			OrganizationId: aws.String(organizationID),
		}

		for item, err := range listDomains(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			domainName := aws.ToString(item.DomainName)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrDomainName), domainName)

			out, err := findDomainByOrgAndName(ctx, conn, organizationID, domainName)
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			result := request.NewListResult(ctx)

			var data domainResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				result.Diagnostics.Append(l.flatten(ctx, out, &data)...)
				if result.Diagnostics.HasError() {
					return
				}
				data.OrganizationId = flex.StringValueToFramework(ctx, organizationID)
				data.DomainName = flex.StringValueToFramework(ctx, domainName)
				result.DisplayName = domainName
			})

			if result.Diagnostics.HasError() {
				yield(list.ListResult{Diagnostics: result.Diagnostics})
				return
			}

			if !yield(result) {
				return
			}
		}
	}
}

type listDomainModel struct {
	framework.WithRegionModel
	OrganizationID types.String `tfsdk:"organization_id"`
}

func listDomains(ctx context.Context, conn *workmail.Client, input *workmail.ListMailDomainsInput) iter.Seq2[awstypes.MailDomainSummary, error] {
	return func(yield func(awstypes.MailDomainSummary, error) bool) {
		pages := workmail.NewListMailDomainsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.MailDomainSummary{}, fmt.Errorf("listing WorkMail Domain resources: %w", err))
				return
			}

			for _, item := range page.MailDomains {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
