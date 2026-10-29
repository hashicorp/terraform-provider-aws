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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// @SDKListResource("aws_config_remediation_configuration")
func newRemediationConfigurationResourceAsListResource() inttypes.ListResourceForSDK {
	return framework.NewSDKListResource(resourceRemediationConfiguration(), framework.SDKListSpec[awstypes.RemediationConfiguration, listRemediationConfigurationModel]{
		ConfigSchema: listRemediationConfigurationConfigSchema,
		Items: func(ctx context.Context, c *conns.AWSClient, query listRemediationConfigurationModel) (iter.Seq2[awstypes.RemediationConfiguration, error], diag.Diagnostics) {
			input := configservice.DescribeRemediationConfigurationsInput{
				ConfigRuleNames: fwflex.ExpandFrameworkStringValueList(ctx, query.ConfigRuleNames),
			}

			return listRemediationConfigurations(ctx, c.ConfigServiceClient(ctx), &input), nil
		},
		ID: func(rc awstypes.RemediationConfiguration, _ listRemediationConfigurationModel) string {
			return aws.ToString(rc.ConfigRuleName)
		},
		Flatten: resourceRemediationConfigurationFlatten,
		DisplayName: func(rc awstypes.RemediationConfiguration) string {
			return aws.ToString(rc.ConfigRuleName)
		},
	})
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

func listRemediationConfigurationConfigSchema(_ context.Context) listschema.Schema {
	return listschema.Schema{
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
