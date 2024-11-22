// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ssm_patch_baseline", name="Patch Baseline")
func dataSourcePatchBaseline() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataPatchBaselineRead,

		Schema: map[string]*schema.Schema{
			"approved_patches": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"approved_patches_compliance_level": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"approved_patches_enable_non_security": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"approval_rule": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"approve_after_days": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"approve_until_date": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"compliance_level": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"enable_non_security": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"patch_filter": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrKey: {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrValues: {
										Type:     schema.TypeList,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
					},
				},
			},
			"default_baseline": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"global_filter": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrKey: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrValues: {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			names.AttrJSON: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrNamePrefix: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 255),
			},
			"operating_system": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.OperatingSystem](),
			},
			names.AttrOwner: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"rejected_patches": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"rejected_patches_action": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrSource: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrConfiguration: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"products": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func dataPatchBaselineRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	filters := []awstypes.PatchOrchestratorFilter{
		{
			Key:    aws.String("OWNER"),
			Values: []string{d.Get(names.AttrOwner).(string)},
		},
	}

	if v, ok := d.GetOk(names.AttrNamePrefix); ok {
		filters = append(filters, awstypes.PatchOrchestratorFilter{
			Key:    aws.String("NAME_PREFIX"),
			Values: []string{v.(string)},
		})
	}

	input := &ssm.DescribePatchBaselinesInput{
		Filters: filters,
	}
	var baselines []awstypes.PatchBaselineIdentity

	pages := ssm.NewDescribePatchBaselinesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading SSM Patch Baselines: %s", err)
		}

		baselines = append(baselines, page.BaselineIdentities...)
	}

	if v, ok := d.GetOk("operating_system"); ok {
		operatingSystem := awstypes.OperatingSystem(v.(string))
		baselines = tfslices.Filter(baselines, func(v awstypes.PatchBaselineIdentity) bool {
			return v.OperatingSystem == operatingSystem
		})
	}

	if v, ok := d.GetOk("default_baseline"); ok {
		defaultBaseline := v.(bool)
		for _, v := range baselines {
			if v.DefaultBaseline == defaultBaseline {
				baselines = []awstypes.PatchBaselineIdentity{v}
				break
			}
		}
	}

	baseline, err := tfresource.AssertSingleValueResult(baselines)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("SSM Patch Baseline", err))
	}

	id := aws.ToString(baseline.BaselineId)
	output, err := findPatchBaselineByID(ctx, conn, id)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SSM Patch Baseline (%s): %s", id, err)
	}

	jsonDoc, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	jsonString := string(jsonDoc)

	d.SetId(id)
	d.Set("approved_patches", output.ApprovedPatches)
	d.Set("approved_patches_compliance_level", output.ApprovedPatchesComplianceLevel)
	d.Set("approved_patches_enable_non_security", output.ApprovedPatchesEnableNonSecurity)
	if err := d.Set("approval_rule", flattenPatchRuleGroup(output.ApprovalRules)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting approval_rule: %s", err)
	}
	d.Set("default_baseline", baseline.DefaultBaseline)
	d.Set(names.AttrDescription, baseline.BaselineDescription)
	if err := d.Set("global_filter", flattenPatchFilterGroup(output.GlobalFilters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting global_filter: %s", err)
	}
	d.Set(names.AttrJSON, jsonString)
	d.Set(names.AttrName, baseline.BaselineName)
	d.Set("operating_system", baseline.OperatingSystem)
	d.Set("rejected_patches", output.RejectedPatches)
	d.Set("rejected_patches_action", output.RejectedPatchesAction)
	if err := d.Set(names.AttrSource, flattenPatchSource(output.Sources)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting source: %s", err)
	}

	return diags
}
