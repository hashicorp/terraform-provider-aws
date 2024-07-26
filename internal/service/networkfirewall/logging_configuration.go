// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkfirewall

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkfirewall/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_networkfirewall_logging_configuration", name="Logging Configuration")
func resourceLoggingConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLoggingConfigurationCreate,
		ReadWithoutTimeout:   resourceLoggingConfigurationRead,
		UpdateWithoutTimeout: resourceLoggingConfigurationUpdate,
		DeleteWithoutTimeout: resourceLoggingConfigurationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"firewall_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrLoggingConfiguration: {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"log_destination_config": {
							// At most 2 configurations can exist,
							// with 1 destination for FLOW logs and 1 for ALERT logs
							Type:     schema.TypeSet,
							Required: true,
							MaxItems: 2,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"log_destination": {
										Type:     schema.TypeMap,
										Required: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"log_destination_type": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[awstypes.LogDestinationType](),
									},
									"log_type": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[awstypes.LogType](),
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceLoggingConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkFirewallClient(ctx)

	firewallARN := d.Get("firewall_arn").(string)
	loggingConfigs := expandLoggingConfigurations(d.Get(names.AttrLoggingConfiguration).([]interface{}))
	if err := addLoggingConfigurations(ctx, conn, firewallARN, loggingConfigs); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.SetId(firewallARN)

	return append(diags, resourceLoggingConfigurationRead(ctx, d, meta)...)
}

func resourceLoggingConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkFirewallClient(ctx)

	output, err := findLoggingConfigurationByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] NetworkFirewall Logging Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading NetworkFirewall Logging Configuration (%s): %s", d.Id(), err)
	}

	d.Set("firewall_arn", output.FirewallArn)
	if err := d.Set(names.AttrLoggingConfiguration, flattenLoggingConfiguration(output.LoggingConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting logging_configuration: %s", err)
	}

	return diags
}

func resourceLoggingConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkFirewallClient(ctx)

	o, n := d.GetChange(names.AttrLoggingConfiguration)

	// Remove destination configs one by one, if any.
	if oldConfig := o.([]interface{}); len(oldConfig) != 0 && oldConfig[0] != nil {
		if loggingConfig := expandLoggingConfigurationOnUpdate(oldConfig); loggingConfig != nil {
			if err := removeLoggingConfiguration(ctx, conn, d.Id(), loggingConfig); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	// Only send new LoggingConfiguration with content.
	if newConfig := n.([]interface{}); len(newConfig) != 0 && newConfig[0] != nil {
		loggingConfigs := expandLoggingConfigurations(d.Get(names.AttrLoggingConfiguration).([]interface{}))
		if err := addLoggingConfigurations(ctx, conn, d.Id(), loggingConfigs); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceLoggingConfigurationRead(ctx, d, meta)...)
}

func resourceLoggingConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkFirewallClient(ctx)

	output, err := findLoggingConfigurationByARN(ctx, conn, d.Id())

	if tfresource.NotFound(err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading NetworkFirewall Logging Configuration (%s): %s", d.Id(), err)
	}

	if output != nil && output.LoggingConfiguration != nil {
		log.Printf("[DEBUG] Deleting NetworkFirewall Logging Configuration: %s", d.Id())
		err := removeLoggingConfiguration(ctx, conn, d.Id(), output.LoggingConfiguration)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return diags
}

func addLoggingConfigurations(ctx context.Context, conn *networkfirewall.Client, arn string, loggingConfigs []*awstypes.LoggingConfiguration) error {
	var errs []error

	for _, loggingConfig := range loggingConfigs {
		input := &networkfirewall.UpdateLoggingConfigurationInput{
			FirewallArn:          aws.String(arn),
			LoggingConfiguration: loggingConfig,
		}

		_, err := conn.UpdateLoggingConfiguration(ctx, input)

		if err != nil {
			errs = append(errs, fmt.Errorf("adding NetworkFirewall Logging Configuration (%s): %w", arn, err))
		}
	}

	return errors.Join(errs...)
}

func removeLoggingConfiguration(ctx context.Context, conn *networkfirewall.Client, arn string, loggingConfig *awstypes.LoggingConfiguration) error {
	if loggingConfig == nil {
		return nil
	}

	var errs []error

	// Must delete destination configs one at a time.
	for i, logDestinationConfig := range loggingConfig.LogDestinationConfigs {
		input := &networkfirewall.UpdateLoggingConfigurationInput{
			FirewallArn: aws.String(arn),
		}

		if i == 0 && len(loggingConfig.LogDestinationConfigs) == 2 {
			loggingConfig := &awstypes.LoggingConfiguration{
				LogDestinationConfigs: []awstypes.LogDestinationConfig{logDestinationConfig},
			}
			input.LoggingConfiguration = loggingConfig
		}

		_, err := conn.UpdateLoggingConfiguration(ctx, input)

		if err != nil {
			errs = append(errs, fmt.Errorf("removing NetworkFirewall Logging Configuration (%s): %w", arn, err))
		}
	}

	return errors.Join(errs...)
}

func findLoggingConfigurationByARN(ctx context.Context, conn *networkfirewall.Client, arn string) (*networkfirewall.DescribeLoggingConfigurationOutput, error) {
	input := &networkfirewall.DescribeLoggingConfigurationInput{
		FirewallArn: aws.String(arn),
	}

	output, err := conn.DescribeLoggingConfiguration(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.LoggingConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func expandLoggingConfigurations(tfList []interface{}) []*awstypes.LoggingConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObjects := make([]*awstypes.LoggingConfiguration, 0)

	if v, ok := tfMap["log_destination_config"].(*schema.Set); ok && v.Len() > 0 {
		for _, tfMapRaw := range v.List() {
			tfMap, ok := tfMapRaw.(map[string]interface{})
			if !ok {
				continue
			}

			logDestinationConfig := awstypes.LogDestinationConfig{}

			if v, ok := tfMap["log_destination"].(map[string]interface{}); ok && len(v) > 0 {
				logDestinationConfig.LogDestination = flex.ExpandStringValueMap(v)
			}
			if v, ok := tfMap["log_destination_type"].(string); ok && v != "" {
				logDestinationConfig.LogDestinationType = awstypes.LogDestinationType(v)
			}
			if v, ok := tfMap["log_type"].(string); ok && v != "" {
				logDestinationConfig.LogType = awstypes.LogType(v)
			}

			// Exclude empty LogDestinationConfig due to TypeMap in TypeSet behavior.
			// Related: https://github.com/hashicorp/terraform-plugin-sdk/issues/588.
			if logDestinationConfig.LogDestination == nil && logDestinationConfig.LogDestinationType == "" && logDestinationConfig.LogType == "" {
				continue
			}

			apiObject := &awstypes.LoggingConfiguration{}
			// Include all (max 2) "log_destination_config" i.e. prepend the already-expanded loggingConfig.
			if len(apiObjects) == 1 && len(apiObjects[0].LogDestinationConfigs) == 1 {
				apiObject.LogDestinationConfigs = append(apiObject.LogDestinationConfigs, apiObjects[0].LogDestinationConfigs[0])
			}
			apiObject.LogDestinationConfigs = append(apiObject.LogDestinationConfigs, logDestinationConfig)

			apiObjects = append(apiObjects, apiObject)
		}
	}

	return apiObjects
}

func expandLoggingConfigurationOnUpdate(tfList []interface{}) *awstypes.LoggingConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.LoggingConfiguration{}

	if v, ok := tfMap["log_destination_config"].(*schema.Set); ok && v.Len() > 0 {
		tfList := v.List()
		logDestinationConfigs := make([]awstypes.LogDestinationConfig, 0, len(tfList))

		for _, tfMapRaw := range tfList {
			tfMap, ok := tfMapRaw.(map[string]interface{})
			if !ok {
				continue
			}

			logDestinationConfig := awstypes.LogDestinationConfig{}

			if v, ok := tfMap["log_destination"].(map[string]interface{}); ok && len(v) > 0 {
				logDestinationConfig.LogDestination = flex.ExpandStringValueMap(v)
			}
			if v, ok := tfMap["log_destination_type"].(string); ok && v != "" {
				logDestinationConfig.LogDestinationType = awstypes.LogDestinationType(v)
			}
			if v, ok := tfMap["log_type"].(string); ok && v != "" {
				logDestinationConfig.LogType = awstypes.LogType(v)
			}

			// Exclude empty LogDestinationConfig due to TypeMap in TypeSet behavior.
			// Related: https://github.com/hashicorp/terraform-plugin-sdk/issues/588.
			if logDestinationConfig.LogDestination == nil && logDestinationConfig.LogDestinationType == "" && logDestinationConfig.LogType == "" {
				continue
			}

			logDestinationConfigs = append(logDestinationConfigs, logDestinationConfig)
		}

		apiObject.LogDestinationConfigs = logDestinationConfigs
	}

	return apiObject
}

func flattenLoggingConfiguration(apiObject *awstypes.LoggingConfiguration) []interface{} {
	if apiObject == nil || apiObject.LogDestinationConfigs == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		"log_destination_config": flattenLoggingConfigurationLogDestinationConfigs(apiObject.LogDestinationConfigs),
	}

	return []interface{}{tfMap}
}

func flattenLoggingConfigurationLogDestinationConfigs(apiObjects []awstypes.LogDestinationConfig) []interface{} {
	tfList := make([]interface{}, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			"log_destination":      apiObject.LogDestination,
			"log_destination_type": apiObject.LogDestinationType,
			"log_type":             apiObject.LogType,
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
