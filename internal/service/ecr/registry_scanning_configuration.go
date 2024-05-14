// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecr

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ecr_registry_scanning_configuration", name="Registry Scanning Configuration")
func resourceRegistryScanningConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRegistryScanningConfigurationPut,
		ReadWithoutTimeout:   resourceRegistryScanningConfigurationRead,
		UpdateWithoutTimeout: resourceRegistryScanningConfigurationPut,
		DeleteWithoutTimeout: resourceRegistryScanningConfigurationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"registry_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrRule: {
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: 0,
				MaxItems: 100,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"repository_filter": {
							Type:     schema.TypeSet,
							MinItems: 1,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrFilter: {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 256),
											validation.StringMatch(regexache.MustCompile(`^[0-9a-z*](?:[0-9a-z_./*-]?[0-9a-z*]+)*$`), "must contain only lowercase alphanumeric, dot, underscore, hyphen, wildcard, and colon characters"),
										),
									},
									"filter_type": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[types.ScanningRepositoryFilterType](),
									},
								},
							},
						},
						"scan_frequency": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.ScanFrequency](),
						},
					},
				},
			},
			"scan_type": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[types.ScanType](),
			},
		},
	}
}

func resourceRegistryScanningConfigurationPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRClient(ctx)

	input := ecr.PutRegistryScanningConfigurationInput{
		ScanType: types.ScanType(d.Get("scan_type").(string)),
		Rules:    expandScanningRegistryRules(d.Get(names.AttrRule).(*schema.Set).List()),
	}

	_, err := conn.PutRegistryScanningConfiguration(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting ECR Registry Scanning Configuration: %s", err)
	}

	if d.IsNewResource() {
		d.SetId(meta.(*conns.AWSClient).AccountID)
	}

	return append(diags, resourceRegistryScanningConfigurationRead(ctx, d, meta)...)
}

func resourceRegistryScanningConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRClient(ctx)

	output, err := findRegistryScanningConfiguration(ctx, conn)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ECR Registry Scanning Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECR Registry Scanning Configuration (%s): %s", d.Id(), err)
	}

	d.Set("registry_id", output.RegistryId)
	if err := d.Set(names.AttrRule, flattenScanningConfigurationRules(output.ScanningConfiguration.Rules)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting rule: %s", err)
	}
	d.Set("scan_type", output.ScanningConfiguration.ScanType)

	return diags
}

func resourceRegistryScanningConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRClient(ctx)

	log.Printf("[DEBUG] Deleting ECR Registry Scanning Configuration: %s", d.Id())
	_, err := conn.PutRegistryScanningConfiguration(ctx, &ecr.PutRegistryScanningConfigurationInput{
		Rules:    []types.RegistryScanningRule{},
		ScanType: types.ScanTypeBasic,
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ECR Registry Scanning Configuration (%s): %s", d.Id(), err)
	}

	return diags
}

func findRegistryScanningConfiguration(ctx context.Context, conn *ecr.Client) (*ecr.GetRegistryScanningConfigurationOutput, error) {
	input := &ecr.GetRegistryScanningConfigurationInput{}

	output, err := conn.GetRegistryScanningConfiguration(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

// Helper functions

func expandScanningRegistryRules(l []interface{}) []types.RegistryScanningRule {
	rules := make([]types.RegistryScanningRule, 0)

	for _, rule := range l {
		if rule == nil {
			continue
		}
		rules = append(rules, expandScanningRegistryRule(rule.(map[string]interface{})))
	}

	return rules
}

func expandScanningRegistryRule(m map[string]interface{}) types.RegistryScanningRule {
	if m == nil {
		return types.RegistryScanningRule{}
	}

	rule := types.RegistryScanningRule{
		RepositoryFilters: expandScanningRegistryRuleRepositoryFilters(m["repository_filter"].(*schema.Set).List()),
		ScanFrequency:     types.ScanFrequency((m["scan_frequency"].(string))),
	}

	return rule
}

func expandScanningRegistryRuleRepositoryFilters(l []interface{}) []types.ScanningRepositoryFilter {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	filters := make([]types.ScanningRepositoryFilter, 0)

	for _, f := range l {
		if f == nil {
			continue
		}
		m := f.(map[string]interface{})
		filters = append(filters, types.ScanningRepositoryFilter{
			Filter:     aws.String(m[names.AttrFilter].(string)),
			FilterType: types.ScanningRepositoryFilterType((m["filter_type"].(string))),
		})
	}

	return filters
}

func flattenScanningConfigurationRules(r []types.RegistryScanningRule) interface{} {
	out := make([]map[string]interface{}, len(r))
	for i, rule := range r {
		m := make(map[string]interface{})
		m["scan_frequency"] = rule.ScanFrequency
		m["repository_filter"] = flattenScanningConfigurationFilters(rule.RepositoryFilters)
		out[i] = m
	}
	return out
}

func flattenScanningConfigurationFilters(l []types.ScanningRepositoryFilter) []interface{} {
	if len(l) == 0 {
		return nil
	}

	out := make([]interface{}, len(l))
	for i, filter := range l {
		out[i] = map[string]interface{}{
			names.AttrFilter: aws.ToString(filter.Filter),
			"filter_type":    filter.FilterType,
		}
	}

	return out
}
