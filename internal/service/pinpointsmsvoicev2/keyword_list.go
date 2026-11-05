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
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
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
			"origination_identity_arn": listschema.StringAttribute{
				CustomType:  fwtypes.ARNType,
				Required:    true,
				Description: "ARN of the origination identity (phone number or pool) to list keywords for.",
			},
		},
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
		OriginationIdentity: query.OriginationIdentityARN.ValueStringPointer(),
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

			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("origination_identity_arn"), aws.ToString(originationIdentityARN))
			ctx = tflog.SetField(ctx, logging.ResourceAttributeKey("keyword"), keyword)

			result := request.NewListResult(ctx)

			var data keywordResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				// KeywordInformation does not carry the origination identity; its ARN
				// comes from the DescribeKeywords response envelope.
				smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, item, originationIdentityARN, &data))
				if result.Diagnostics.HasError() {
					return
				}

				// keyword_action is AWS-managed for the mandatory keywords HELP and STOP
				// and cannot be set (see the resource ValidateConfig). Null it so any
				// generated configuration omits it and remains valid on apply.
				if isMandatoryKeyword(keyword) {
					data.KeywordAction = fwtypes.StringEnumNull[awstypes.KeywordAction]()
				}

				result.DisplayName = keyword
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

type keywordListModel struct {
	framework.WithRegionModel
	OriginationIdentityARN fwtypes.ARN `tfsdk:"origination_identity_arn"`
}
