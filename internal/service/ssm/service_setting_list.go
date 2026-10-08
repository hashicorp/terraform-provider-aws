// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_ssm_service_setting")
func newServiceSettingResourceAsListResource() inttypes.ListResourceForSDK {
	l := serviceSettingListResource{}
	l.SetResourceSchema(resourceServiceSetting())
	return &l
}

var _ list.ListResource = &serviceSettingListResource{}
var _ list.ListResourceWithRawV5Schemas = &serviceSettingListResource{}

type serviceSettingListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type serviceSettingListResourceModel struct {
	framework.WithRegionModel
	ARN types.String `tfsdk:"arn"`
}

func (l *serviceSettingListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			names.AttrARN: listschema.StringAttribute{
				Required:    true,
				Description: "ARN of the SSM Service Setting.",
				Validators: []validator.String{
					fwvalidators.ARN(),
				},
			},
		},
	}
}

func (l *serviceSettingListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	var query serviceSettingListResourceModel
	if diags := request.Config.Get(ctx, &query); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	awsClient := l.Meta()
	conn := awsClient.SSMClient(ctx)

	settingARN := query.ARN.ValueString()

	tflog.Info(ctx, "Listing SSM Service Setting", map[string]any{
		names.AttrARN: settingARN,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		// Service Settings have no List API. This is a per-account, per-Region singleton looked up by ARN.
		output, err := findServiceSettingByID(ctx, conn, settingARN)
		if err != nil {
			result := fwdiag.NewListResultErrorDiagnostic(fmt.Errorf("reading SSM Service Setting (%s): %w", settingARN, err))
			yield(result)
			return
		}

		arn := aws.ToString(output.ARN)
		ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

		result := request.NewListResult(ctx)

		rd := l.ResourceData()
		rd.SetId(arn)
		rd.Set(names.AttrARN, arn)

		if request.IncludeResource {
			resourceServiceSettingFlatten(rd, output)
		}

		result.DisplayName = aws.ToString(output.SettingId)

		l.SetResult(ctx, awsClient, request.IncludeResource, rd, &result)
		yield(result)
	}
}
