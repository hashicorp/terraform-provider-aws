// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_wafv2_ip_set")
func newIPSetResourceAsListResource() inttypes.ListResourceForSDK {
	l := ipSetListResource{}
	l.SetResourceSchema(resourceIPSet())

	return &l
}

var _ list.ListResource = &ipSetListResource{}

type ipSetListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type listIPSetModel struct {
	framework.WithRegionModel
	Scope types.String `tfsdk:"scope"`
}

func (l *ipSetListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			names.AttrScope: listschema.StringAttribute{
				Required:    true,
				Description: "Whether this is for a global (CLOUDFRONT) or regional (REGIONAL) application.",
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.Scope](),
				},
			},
		},
	}
}

func (l *ipSetListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().WAFV2Client(ctx)

	var query listIPSetModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	scope := query.Scope.ValueString()

	tflog.Info(ctx, "Listing Resources", map[string]any{
		logging.ResourceAttributeKey(names.AttrScope): scope,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		input := wafv2.ListIPSetsInput{
			Scope: awstypes.Scope(scope),
		}
		for item, err := range listIPSets(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			id, name := aws.ToString(item.Id), aws.ToString(item.Name)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(id)
			rd.Set(names.AttrName, name)
			rd.Set(names.AttrScope, scope)

			if request.IncludeResource {
				output, err := findIPSetByThreePartKey(ctx, conn, id, name, scope)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}

				resourceIPSetFlatten(output, rd)
			}

			result.DisplayName = name

			l.SetResult(ctx, l.Meta(), request.IncludeResource, rd, &result)
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

func listIPSets(ctx context.Context, conn *wafv2.Client, input *wafv2.ListIPSetsInput) iter.Seq2[awstypes.IPSetSummary, error] {
	return func(yield func(awstypes.IPSetSummary, error) bool) {
		var stopped bool
		err := listIPSetsPages(ctx, conn, input, func(page *wafv2.ListIPSetsOutput, lastPage bool) bool {
			for _, v := range page.IPSets {
				if !yield(v, nil) {
					stopped = true
					return false
				}
			}
			return !lastPage
		})
		if err != nil && !stopped {
			yield(awstypes.IPSetSummary{}, fmt.Errorf("listing WAFv2 IP Set resources: %w", err))
		}
	}
}
