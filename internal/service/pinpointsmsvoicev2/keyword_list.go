// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
)

// @FrameworkListResource("aws_pinpointsmsvoicev2_keyword")
func newKeywordResourceAsListResource() list.ListResourceWithConfigure {
	return &keywordListResource{}
}

var _ list.ListResource = &keywordListResource{}

type keywordListResource struct {
	keywordResource
	framework.WithList
}

func (l *keywordListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"origination_identity": listschema.StringAttribute{
				Required:    true,
				Description: "Origination identity to list keywords for. Value is the ID or ARN of a phone number or pool.",
			},
		},
		Blocks: map[string]listschema.Block{},
	}
}

func (l *keywordListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().PinpointSMSVoiceV2Client(ctx)

	var query keywordListModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	input := pinpointsmsvoicev2.DescribeKeywordsInput{
		OriginationIdentity: query.OriginationIdentity.ValueStringPointer(),
	}

	tflog.Info(ctx, "Listing End User Messaging SMS Keywords")

	stream.Results = func(yield func(list.ListResult) bool) {
		keywords, originationIdentityARN, err := findKeywords(ctx, conn, &input)
		if err != nil {
			result := fwdiag.NewListResultErrorDiagnostic(err)
			yield(result)
			return
		}

		for i := range keywords {
			item := &keywords[i]
			keyword := aws.ToString(item.Keyword)

			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("origination_identity"), query.OriginationIdentity.ValueString())
			ctx = tflog.SetField(ctx, logging.ResourceAttributeKey("keyword"), keyword)

			result := request.NewListResult(ctx)

			var data keywordResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				smerr.AddEnrich(ctx, &result.Diagnostics, fwflex.Flatten(ctx, item, &data))
				if result.Diagnostics.HasError() {
					return
				}

				data.OriginationIdentity = query.OriginationIdentity
				data.OriginationIdentityARN = fwtypes.ARNValue(aws.ToString(originationIdentityARN))

				// keyword_action is AWS-managed for the mandatory keywords HELP and STOP
				// and cannot be set (see the resource ValidateConfig). Null it so any
				// generated configuration omits it and remains valid on apply.
				if isMandatoryKeyword(keyword) {
					data.KeywordAction = fwtypes.StringEnumNull[awstypes.KeywordAction]()
				}

				result.DisplayName = keyword
			})

			if !yield(result) {
				return
			}
		}
	}
}

type keywordListModel struct {
	framework.WithRegionModel
	OriginationIdentity types.String `tfsdk:"origination_identity"`
}
