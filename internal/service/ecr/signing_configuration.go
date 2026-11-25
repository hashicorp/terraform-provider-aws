// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

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
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ecr_signing_configuration", name="Signing Configuration")
func resourceSigningConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSigningConfigurationPut,
		ReadWithoutTimeout:   resourceSigningConfigurationRead,
		UpdateWithoutTimeout: resourceSigningConfigurationPut,
		DeleteWithoutTimeout: resourceSigningConfigurationDelete,

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
				Required: true,
				MinItems: 1,
				MaxItems: 10,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"repository_filter": {
							Type:     schema.TypeSet,
							Optional: true,
							MinItems: 1,
							MaxItems: 100,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrFilter: {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 256),
											validation.StringMatch(
												regexache.MustCompile(`^(?:[a-z0-9*]+(?:[._-][a-z0-9*]+)*/)*[a-z0-9*]+(?:[._-][a-z0-9*]+)*$`),
												"must contain only lowercase alphanumeric, dot, underscore, hyphen, slash, and wildcard characters",
											),
										),
									},
									"filter_type": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[types.SigningRepositoryFilterType](),
									},
								},
							},
						},
						"signing_profile_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
		},
	}
}

func resourceSigningConfigurationPut(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRClient(ctx)

	input := &ecr.PutSigningConfigurationInput{
		SigningConfiguration: &types.SigningConfiguration{
			Rules: expandSigningConfigurationRules(d.Get(names.AttrRule).(*schema.Set).List()),
		},
	}

	_, err := conn.PutSigningConfiguration(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting ECR Signing Configuration: %s", err)
	}

	if d.IsNewResource() {
		d.SetId(meta.(*conns.AWSClient).AccountID(ctx))
	}

	return append(diags, resourceSigningConfigurationRead(ctx, d, meta)...)
}

func resourceSigningConfigurationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRClient(ctx)

	output, err := findSigningConfiguration(ctx, conn)
	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] ECR Signing Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECR Signing Configuration (%s): %s", d.Id(), err)
	}

	d.Set("registry_id", output.RegistryId)
	if err := d.Set(names.AttrRule, flattenSigningConfigurationRules(output.SigningConfiguration.Rules)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting rule: %s", err)
	}

	return diags
}

func resourceSigningConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRClient(ctx)

	log.Printf("[DEBUG] Deleting ECR Signing Configuration: %s", d.Id())
	_, err := conn.DeleteSigningConfiguration(ctx, &ecr.DeleteSigningConfigurationInput{})
	if errs.IsA[*types.SigningConfigurationNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ECR Signing Configuration (%s): %s", d.Id(), err)
	}

	return diags
}

func findSigningConfiguration(ctx context.Context, conn *ecr.Client) (*ecr.GetSigningConfigurationOutput, error) {
	input := &ecr.GetSigningConfigurationInput{}

	output, err := conn.GetSigningConfiguration(ctx, input)
	if errs.IsA[*types.SigningConfigurationNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.SigningConfiguration == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func expandSigningConfigurationRules(l []any) []types.SigningRule {
	if len(l) == 0 {
		return nil
	}

	rules := make([]types.SigningRule, 0, len(l))

	for _, rule := range l {
		if rule == nil {
			continue
		}

		rules = append(rules, expandSigningConfigurationRule(rule.(map[string]any)))
	}

	return rules
}

func expandSigningConfigurationRule(m map[string]any) types.SigningRule {
	if m == nil {
		return types.SigningRule{}
	}

	rule := types.SigningRule{
		SigningProfileArn: aws.String(m["signing_profile_arn"].(string)),
	}

	if v, ok := m["repository_filter"]; ok {
		rule.RepositoryFilters = expandSigningConfigurationRepositoryFilters(v.(*schema.Set).List())
	}

	return rule
}

func expandSigningConfigurationRepositoryFilters(l []any) []types.SigningRepositoryFilter {
	if len(l) == 0 {
		return nil
	}

	filters := make([]types.SigningRepositoryFilter, 0, len(l))

	for _, filter := range l {
		if filter == nil {
			continue
		}

		m := filter.(map[string]any)
		filters = append(filters, types.SigningRepositoryFilter{
			Filter:     aws.String(m[names.AttrFilter].(string)),
			FilterType: types.SigningRepositoryFilterType(m["filter_type"].(string)),
		})
	}

	return filters
}

func flattenSigningConfigurationRules(r []types.SigningRule) []map[string]any {
	if len(r) == 0 {
		return nil
	}

	out := make([]map[string]any, 0, len(r))

	for _, rule := range r {
		out = append(out, map[string]any{
			"repository_filter":   flattenSigningConfigurationRepositoryFilters(rule.RepositoryFilters),
			"signing_profile_arn": aws.ToString(rule.SigningProfileArn),
		})
	}

	return out
}

func flattenSigningConfigurationRepositoryFilters(filters []types.SigningRepositoryFilter) []map[string]any {
	if len(filters) == 0 {
		return nil
	}

	out := make([]map[string]any, 0, len(filters))

	for _, filter := range filters {
		out = append(out, map[string]any{
			names.AttrFilter: aws.ToString(filter.Filter),
			"filter_type":    filter.FilterType,
		})
	}

	return out
}
