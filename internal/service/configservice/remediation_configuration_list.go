// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package configservice

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/configservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/configservice/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_config_remediation_configuration")
func newRemediationConfigurationResourceAsListResource() inttypes.ListResourceForSDK {
	l := remediationConfigurationListResource{}
	l.SetResourceSchema(resourceRemediationConfiguration())
	return &l
}

var _ list.ListResource = &remediationConfigurationListResource{}

type remediationConfigurationListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *remediationConfigurationListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"config_rule_names": listschema.ListAttribute{
				ElementType: types.StringType,
				CustomType:  fwtypes.ListOfStringType,
				Required:    true,
				Description: "Names of the AWS Config rules.",
			},
		},
	}
}

func (l *remediationConfigurationListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().ConfigServiceClient(ctx)

	var query listRemediationConfigurationModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	configRuleNames := fwflex.ExpandFrameworkStringValueList(ctx, query.ConfigRuleNames)

	stream.Results = func(yield func(list.ListResult) bool) {
		input := configservice.DescribeRemediationConfigurationsInput{
			ConfigRuleNames: configRuleNames,
		}
		for item, err := range listRemediationConfigurations(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			arn := aws.ToString(item.Arn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			result := request.NewListResult(ctx)

			name := aws.ToString(item.ConfigRuleName)
			rd := l.ResourceData()
			rd.SetId(name)
			rd.Set("config_rule_name", name)

			if request.IncludeResource {
				if err := resourceRemediationConfigurationFlatten(ctx, &item, rd); err != nil {
					tflog.Error(ctx, "Reading Config Remediation Configuration", map[string]any{
						"error": err.Error(),
					})
					continue
				}
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

type listRemediationConfigurationModel struct {
	framework.WithRegionModel
	ConfigRuleNames fwtypes.ListOfString `tfsdk:"config_rule_names"`
}

func listRemediationConfigurations(ctx context.Context, conn *configservice.Client, input *configservice.DescribeRemediationConfigurationsInput) iter.Seq2[awstypes.RemediationConfiguration, error] {
	return func(yield func(awstypes.RemediationConfiguration, error) bool) {
		page, err := conn.DescribeRemediationConfigurations(ctx, input)
		if err != nil {
			yield(inttypes.Zero[awstypes.RemediationConfiguration](), fmt.Errorf("listing Config Remediation Configuration resources: %w", err))
			return
		}

		for _, item := range page.RemediationConfigurations {
			if !yield(item, nil) {
				return
			}
		}
	}
}
