// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecr

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ecr_replication_configuration", name="Replication Configuration")
func resourceReplicationConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceReplicationConfigurationPut,
		ReadWithoutTimeout:   resourceReplicationConfigurationRead,
		UpdateWithoutTimeout: resourceReplicationConfigurationPut,
		DeleteWithoutTimeout: resourceReplicationConfigurationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"registry_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"replication_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrRule: {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 10,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrDestination: {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 25,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrRegion: {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: verify.ValidRegionName,
												},
												"registry_id": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: verify.ValidAccountID,
												},
											},
										},
									},
									"repository_filter": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 1,
										MaxItems: 100,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrFilter: {
													Type:     schema.TypeString,
													Required: true,
												},
												"filter_type": {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[types.RepositoryFilterType](),
												},
											},
										},
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

func resourceReplicationConfigurationPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRClient(ctx)

	input := &ecr.PutReplicationConfigurationInput{
		ReplicationConfiguration: expandReplicationConfigurationReplicationConfiguration(d.Get("replication_configuration").([]interface{})),
	}

	_, err := conn.PutReplicationConfiguration(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting ECR Replication Configuration: %s", err)
	}

	if d.IsNewResource() {
		d.SetId(meta.(*conns.AWSClient).AccountID)
	}

	return append(diags, resourceReplicationConfigurationRead(ctx, d, meta)...)
}

func resourceReplicationConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRClient(ctx)

	output, err := findReplicationConfiguration(ctx, conn)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ECR Replication Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECR Replication Configuration (%s): %s", d.Id(), err)
	}

	d.Set("registry_id", output.RegistryId)
	if err := d.Set("replication_configuration", flattenReplicationConfigurationReplicationConfiguration(output.ReplicationConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting replication_configuration: %s", err)
	}

	return diags
}

func resourceReplicationConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRClient(ctx)

	log.Printf("[DEBUG] Deleting ECR Replication Configuration: %s", d.Id())
	_, err := conn.PutReplicationConfiguration(ctx, &ecr.PutReplicationConfigurationInput{
		ReplicationConfiguration: &types.ReplicationConfiguration{
			Rules: []types.ReplicationRule{},
		},
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ECR Replication Configuration (%s): %s", d.Id(), err)
	}

	return diags
}

func findReplicationConfiguration(ctx context.Context, conn *ecr.Client) (*ecr.DescribeRegistryOutput, error) {
	input := &ecr.DescribeRegistryInput{}

	output, err := conn.DescribeRegistry(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil || output.ReplicationConfiguration == nil || len(output.ReplicationConfiguration.Rules) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func expandReplicationConfigurationReplicationConfiguration(data []interface{}) *types.ReplicationConfiguration {
	if len(data) == 0 || data[0] == nil {
		return nil
	}

	ec := data[0].(map[string]interface{})
	config := &types.ReplicationConfiguration{
		Rules: expandReplicationConfigurationReplicationConfigurationRules(ec[names.AttrRule].([]interface{})),
	}
	return config
}

func flattenReplicationConfigurationReplicationConfiguration(ec *types.ReplicationConfiguration) []map[string]interface{} {
	if ec == nil {
		return nil
	}

	config := map[string]interface{}{
		names.AttrRule: flattenReplicationConfigurationReplicationConfigurationRules(ec.Rules),
	}

	return []map[string]interface{}{
		config,
	}
}

func expandReplicationConfigurationReplicationConfigurationRules(data []interface{}) []types.ReplicationRule {
	if len(data) == 0 || data[0] == nil {
		return nil
	}

	var rules []types.ReplicationRule

	for _, rule := range data {
		ec := rule.(map[string]interface{})
		config := types.ReplicationRule{
			Destinations:      expandReplicationConfigurationReplicationConfigurationRulesDestinations(ec[names.AttrDestination].([]interface{})),
			RepositoryFilters: expandReplicationConfigurationReplicationConfigurationRulesRepositoryFilters(ec["repository_filter"].([]interface{})),
		}

		rules = append(rules, config)
	}
	return rules
}

func flattenReplicationConfigurationReplicationConfigurationRules(ec []types.ReplicationRule) []interface{} {
	if len(ec) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range ec {
		tfMap := map[string]interface{}{
			names.AttrDestination: flattenReplicationConfigurationReplicationConfigurationRulesDestinations(apiObject.Destinations),
			"repository_filter":   flattenReplicationConfigurationReplicationConfigurationRulesRepositoryFilters(apiObject.RepositoryFilters),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func expandReplicationConfigurationReplicationConfigurationRulesDestinations(data []interface{}) []types.ReplicationDestination {
	if len(data) == 0 || data[0] == nil {
		return nil
	}

	var dests []types.ReplicationDestination

	for _, dest := range data {
		ec := dest.(map[string]interface{})
		config := types.ReplicationDestination{
			Region:     aws.String(ec[names.AttrRegion].(string)),
			RegistryId: aws.String(ec["registry_id"].(string)),
		}

		dests = append(dests, config)
	}
	return dests
}

func flattenReplicationConfigurationReplicationConfigurationRulesDestinations(ec []types.ReplicationDestination) []interface{} {
	if len(ec) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range ec {
		tfMap := map[string]interface{}{
			names.AttrRegion: aws.ToString(apiObject.Region),
			"registry_id":    aws.ToString(apiObject.RegistryId),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func expandReplicationConfigurationReplicationConfigurationRulesRepositoryFilters(data []interface{}) []types.RepositoryFilter {
	if len(data) == 0 || data[0] == nil {
		return nil
	}

	var filters []types.RepositoryFilter

	for _, filter := range data {
		ec := filter.(map[string]interface{})
		config := types.RepositoryFilter{
			Filter:     aws.String(ec[names.AttrFilter].(string)),
			FilterType: types.RepositoryFilterType((ec["filter_type"].(string))),
		}

		filters = append(filters, config)
	}
	return filters
}

func flattenReplicationConfigurationReplicationConfigurationRulesRepositoryFilters(ec []types.RepositoryFilter) []interface{} {
	if len(ec) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range ec {
		tfMap := map[string]interface{}{
			names.AttrFilter: aws.ToString(apiObject.Filter),
			"filter_type":    apiObject.FilterType,
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
