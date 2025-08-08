// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkfirewall

import (
	"context"
	"fmt"
	"log"
	"slices"

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
							Type:     schema.TypeSet,
							Required: true,
							MaxItems: len(enum.Values[awstypes.LogType]()),
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

		CustomizeDiff: func(ctx context.Context, d *schema.ResourceDiff, meta any) error {
			// Ensure distinct logging_configuration.log_destination_config.log_type values.
			if v, ok := d.GetOk(names.AttrLoggingConfiguration); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				tfMap := v.([]any)[0].(map[string]any)

				if v, ok := tfMap["log_destination_config"].(*schema.Set); ok && v.Len() > 0 {
					logTypes := make(map[string]struct{})

					for _, tfMapRaw := range v.List() {
						tfMap, ok := tfMapRaw.(map[string]any)
						if !ok {
							continue
						}

						if v, ok := tfMap["log_type"].(string); ok && v != "" {
							if _, ok := logTypes[v]; ok {
								return fmt.Errorf("duplicate logging_configuration.log_destination_config.log_type value: %s", v)
							}
							logTypes[v] = struct{}{}
						}
					}
				}
			}

			return nil
		},
	}
}

func resourceLoggingConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkFirewallClient(ctx)

	firewallARN := d.Get("firewall_arn").(string)

	if v, ok := d.GetOk(names.AttrLoggingConfiguration); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		tfMap := v.([]any)[0].(map[string]any)

		if v, ok := tfMap["log_destination_config"].(*schema.Set); ok && v.Len() > 0 {
			if err := addLogDestinationConfigs(ctx, conn, firewallARN, &awstypes.LoggingConfiguration{}, expandLogDestinationConfigs(v.List())); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	d.SetId(firewallARN)

	return append(diags, resourceLoggingConfigurationRead(ctx, d, meta)...)
}

func resourceLoggingConfigurationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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

func resourceLoggingConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkFirewallClient(ctx)

	output, err := findLoggingConfigurationByARN(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading NetworkFirewall Logging Configuration (%s): %s", d.Id(), err)
	}

	o, n := d.GetChange("logging_configuration.0.log_destination_config")
	os, ns := o.(*schema.Set), n.(*schema.Set)
	add, del := ns.Difference(os), os.Difference(ns)

	if err := deleteLogDestinationConfigs(ctx, conn, d.Id(), output.LoggingConfiguration, expandLogDestinationConfigs(del.List())); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if err := addLogDestinationConfigs(ctx, conn, d.Id(), output.LoggingConfiguration, expandLogDestinationConfigs(add.List())); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return append(diags, resourceLoggingConfigurationRead(ctx, d, meta)...)
}

func resourceLoggingConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkFirewallClient(ctx)

	output, err := findLoggingConfigurationByARN(ctx, conn, d.Id())

	if tfresource.NotFound(err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading NetworkFirewall Logging Configuration (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Deleting NetworkFirewall Logging Configuration: %s", d.Id())
	if v, ok := d.GetOk(names.AttrLoggingConfiguration); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		tfMap := v.([]any)[0].(map[string]any)

		if v, ok := tfMap["log_destination_config"].(*schema.Set); ok && v.Len() > 0 {
			if err := deleteLogDestinationConfigs(ctx, conn, d.Id(), output.LoggingConfiguration, expandLogDestinationConfigs(v.List())); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	return diags
}

// See https://docs.aws.amazon.com/network-firewall/latest/APIReference/API_UpdateLoggingConfiguration.html.
// The logging configuration is changed one LogDestinationConfig at a time.

func addLogDestinationConfigs(ctx context.Context, conn *networkfirewall.Client, firewallARN string, loggingConfiguration *awstypes.LoggingConfiguration, logDestinationConfigs []awstypes.LogDestinationConfig) error {
	for _, logDestinationConfig := range logDestinationConfigs {
		loggingConfiguration.LogDestinationConfigs = append(loggingConfiguration.LogDestinationConfigs, logDestinationConfig)

		input := &networkfirewall.UpdateLoggingConfigurationInput{
			FirewallArn:          aws.String(firewallARN),
			LoggingConfiguration: loggingConfiguration,
		}

		_, err := conn.UpdateLoggingConfiguration(ctx, input)

		if err != nil {
			return fmt.Errorf("adding NetworkFirewall Logging Configuration (%s): %w", firewallARN, err)
		}
	}

	return nil
}

func deleteLogDestinationConfigs(ctx context.Context, conn *networkfirewall.Client, firewallARN string, loggingConfiguration *awstypes.LoggingConfiguration, logDestinationConfigs []awstypes.LogDestinationConfig) error {
	for _, logDestinationConfig := range logDestinationConfigs {
		loggingConfiguration.LogDestinationConfigs = slices.DeleteFunc(loggingConfiguration.LogDestinationConfigs, func(v awstypes.LogDestinationConfig) bool {
			return v.LogType == logDestinationConfig.LogType
		})

		input := &networkfirewall.UpdateLoggingConfigurationInput{
			FirewallArn:          aws.String(firewallARN),
			LoggingConfiguration: loggingConfiguration,
		}

		_, err := conn.UpdateLoggingConfiguration(ctx, input)

		if err != nil {
			return fmt.Errorf("deleting NetworkFirewall Logging Configuration (%s): %w", firewallARN, err)
		}
	}

	return nil
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

	if output == nil || output.LoggingConfiguration == nil || len(output.LoggingConfiguration.LogDestinationConfigs) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func expandLogDestinationConfigs(tfList []any) []awstypes.LogDestinationConfig {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.LogDestinationConfig

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := awstypes.LogDestinationConfig{}

		if v, ok := tfMap["log_destination"].(map[string]any); ok && len(v) > 0 {
			apiObject.LogDestination = flex.ExpandStringValueMap(v)
		}

		if v, ok := tfMap["log_destination_type"].(string); ok && v != "" {
			apiObject.LogDestinationType = awstypes.LogDestinationType(v)
		}

		if v, ok := tfMap["log_type"].(string); ok && v != "" {
			apiObject.LogType = awstypes.LogType(v)
		} else {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenLoggingConfiguration(apiObject *awstypes.LoggingConfiguration) []any {
	if apiObject == nil || apiObject.LogDestinationConfigs == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"log_destination_config": flattenLogDestinationConfigs(apiObject.LogDestinationConfigs),
	}

	return []any{tfMap}
}

func flattenLogDestinationConfigs(apiObjects []awstypes.LogDestinationConfig) []any {
	tfList := make([]any, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			"log_destination":      apiObject.LogDestination,
			"log_destination_type": apiObject.LogDestinationType,
			"log_type":             apiObject.LogType,
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
