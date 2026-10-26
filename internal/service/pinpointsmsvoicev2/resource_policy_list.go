// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_pinpointsmsvoicev2_resource_policy")
func newResourcePolicyResourceAsListResource() list.ListResourceWithConfigure {
	return &resourcePolicyListResource{}
}

var _ list.ListResource = &resourcePolicyListResource{}

type resourcePolicyListResource struct {
	resourcePolicyResource
	framework.WithList
}

func (l *resourcePolicyListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			names.AttrResourceARN: listschema.StringAttribute{
				CustomType:  fwtypes.ARNType,
				Required:    true,
				Description: "ARN of the End User Messaging SMS resource whose attached policy to list.",
			},
		},
	}
}

// There is no ListResourcePolicies API — a resource policy is retrieved
// per parent ARN via GetResourcePolicy. Listing is therefore scoped to a
// required resource_arn and yields the single attached policy (or nothing
// when none is attached).
func (l *resourcePolicyListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().PinpointSMSVoiceV2Client(ctx)

	var query listResourcePolicyModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	arn := fwflex.StringValueFromFramework(ctx, query.ResourceARN)

	ctx = tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrResourceARN), arn)

	stream.Results = func(yield func(list.ListResult) bool) {
		output, err := findResourcePolicyByARN(ctx, conn, arn)
		if retry.NotFound(err) {
			return
		}
		if err != nil {
			yield(fwdiag.NewListResultErrorDiagnostic(err))
			return
		}

		result := request.NewListResult(ctx)

		var data resourcePolicyResourceModel
		l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
			smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, output, &data))
			if result.Diagnostics.HasError() {
				return
			}

			result.DisplayName = arn
		})

		yield(result)
	}
}

type listResourcePolicyModel struct {
	framework.WithRegionModel
	ResourceARN fwtypes.ARN `tfsdk:"resource_arn"`
}
