// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecr

import (
	"context"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_ecr_registry_scanning_configuration")
func ResourceRegistryScanningConfiguration() *schema.Resource {
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
			"rule": {
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
									"filter": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 256),
											validation.StringMatch(regexp.MustCompile(`^[a-z0-9*](?:[._\-/a-z0-9*]?[a-z0-9*]+)*$`), "must contain only lowercase alphanumeric, dot, underscore, hyphen, wildcard, and colon characters"),
										),
									},
									"filter_type": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(ecr.ScanningRepositoryFilterType_Values(), false),
									},
								},
							},
						},
						"scan_frequency": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(ecr.ScanFrequency_Values(), false),
						},
					},
				},
			},
			"scan_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(ecr.ScanType_Values(), false),
			},
		},
	}
}

func resourceRegistryScanningConfigurationPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRConn(ctx)

	input := ecr.PutRegistryScanningConfigurationInput{
		ScanType: aws.String(d.Get("scan_type").(string)),
		Rules:    expandScanningRegistryRules(d.Get("rule").(*schema.Set).List()),
	}

	_, err := conn.PutRegistryScanningConfigurationWithContext(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ECR Registry Scanning Configuration: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).AccountID)

	return append(diags, resourceRegistryScanningConfigurationRead(ctx, d, meta)...)
}

func resourceRegistryScanningConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRConn(ctx)

	out, err := conn.GetRegistryScanningConfigurationWithContext(ctx, &ecr.GetRegistryScanningConfigurationInput{})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECR Registry Scanning Configuration (%s): %s", d.Id(), err)
	}

	d.Set("registry_id", out.RegistryId)
	d.Set("scan_type", out.ScanningConfiguration.ScanType)
	d.Set("rule", flattenScanningConfigurationRules(out.ScanningConfiguration.Rules))

	return diags
}

func resourceRegistryScanningConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRConn(ctx)

	log.Printf("[DEBUG] Deleting ECR Registry Scanning Configuration: (%s)", d.Id())
	_, err := conn.PutRegistryScanningConfigurationWithContext(ctx, &ecr.PutRegistryScanningConfigurationInput{
		Rules:    []*ecr.RegistryScanningRule{},
		ScanType: aws.String(ecr.ScanTypeBasic),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ECR Registry Scanning Configuration (%s): %s", d.Id(), err)
	}

	return diags
}

// Helper functions

func expandScanningRegistryRules(l []interface{}) []*ecr.RegistryScanningRule {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	rules := make([]*ecr.RegistryScanningRule, 0)

	for _, rule := range l {
		if rule == nil {
			continue
		}
		rules = append(rules, expandScanningRegistryRule(rule.(map[string]interface{})))
	}

	return rules
}

func expandScanningRegistryRule(m map[string]interface{}) *ecr.RegistryScanningRule {
	if m == nil {
		return nil
	}

	rule := &ecr.RegistryScanningRule{
		RepositoryFilters: expandScanningRegistryRuleRepositoryFilters(m["repository_filter"].(*schema.Set).List()),
		ScanFrequency:     aws.String(m["scan_frequency"].(string)),
	}

	return rule
}

func expandScanningRegistryRuleRepositoryFilters(l []interface{}) []*ecr.ScanningRepositoryFilter {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	filters := make([]*ecr.ScanningRepositoryFilter, 0)

	for _, f := range l {
		if f == nil {
			continue
		}
		m := f.(map[string]interface{})
		filters = append(filters, &ecr.ScanningRepositoryFilter{
			Filter:     aws.String(m["filter"].(string)),
			FilterType: aws.String(m["filter_type"].(string)),
		})
	}

	return filters
}

func flattenScanningConfigurationRules(r []*ecr.RegistryScanningRule) interface{} {
	out := make([]map[string]interface{}, len(r))
	for i, rule := range r {
		m := make(map[string]interface{})
		m["scan_frequency"] = aws.StringValue(rule.ScanFrequency)
		m["repository_filter"] = flattenScanningConfigurationFilters(rule.RepositoryFilters)
		out[i] = m
	}
	return out
}

func flattenScanningConfigurationFilters(l []*ecr.ScanningRepositoryFilter) []interface{} {
	if len(l) == 0 {
		return nil
	}

	out := make([]interface{}, len(l))
	for i, filter := range l {
		out[i] = map[string]interface{}{
			"filter":      aws.StringValue(filter.Filter),
			"filter_type": aws.StringValue(filter.FilterType),
		}
	}

	return out
}
