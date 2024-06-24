// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ssm_patch_baseline", name="Patch Baseline")
// @Tags(identifierAttribute="id", resourceType="PatchBaseline")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/ssm;ssm.GetPatchBaselineOutput")
func resourcePatchBaseline() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePatchBaselineCreate,
		ReadWithoutTimeout:   resourcePatchBaselineRead,
		UpdateWithoutTimeout: resourcePatchBaselineUpdate,
		DeleteWithoutTimeout: resourcePatchBaselineDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"approval_rule": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"approve_after_days": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 100),
						},
						"approve_until_date": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringMatch(regexache.MustCompile(`([12]\d{3}-(0[1-9]|1[0-2])-(0[1-9]|[12]\d|3[01]))`), "must be formatted YYYY-MM-DD"),
						},
						"compliance_level": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.PatchComplianceLevelUnspecified,
							ValidateDiagFunc: enum.Validate[awstypes.PatchComplianceLevel](),
						},
						"enable_non_security": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"patch_filter": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 10,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrKey: {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[awstypes.PatchFilterKey](),
									},
									names.AttrValues: {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 20,
										MinItems: 1,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 64),
										},
									},
								},
							},
						},
					},
				},
			},
			"approved_patches": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 50,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(1, 100),
				},
			},
			"approved_patches_compliance_level": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.PatchComplianceLevelUnspecified,
				ValidateDiagFunc: enum.Validate[awstypes.PatchComplianceLevel](),
			},
			"approved_patches_enable_non_security": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"global_filter": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 4,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrKey: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.PatchFilterKey](),
						},
						names.AttrValues: {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 20,
							MinItems: 1,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(1, 64),
							},
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
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 128),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]{3,128}$`), "must contain only alphanumeric, underscore, hyphen, or period characters"),
				),
			},
			"operating_system": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          awstypes.OperatingSystemWindows,
				ValidateDiagFunc: enum.Validate[awstypes.OperatingSystem](),
			},
			"rejected_patches": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 50,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(1, 100),
				},
			},
			"rejected_patches_action": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[awstypes.PatchAction](),
			},
			names.AttrSource: {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 20,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrConfiguration: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 1024),
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(3, 50),
								validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]{3,50}$`), "must contain only alphanumeric, underscore, hyphen, or period characters"),
							),
						},
						"products": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 20,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(1, 128),
							},
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: customdiff.Sequence(
			func(_ context.Context, d *schema.ResourceDiff, meta interface{}) error {
				if d.HasChanges(
					names.AttrDescription,
					"global_filter",
					"approval_rule",
					"approved_patches",
					"rejected_patches",
					"operating_system",
					"approved_patches_compliance_level",
					"rejected_patches_action",
					"approved_patches_enable_non_security",
					names.AttrSource,
				) {
					return d.SetNewComputed(names.AttrJSON)
				}

				return nil
			},
			verify.SetTagsDiff,
		),
	}
}

func resourcePatchBaselineCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &ssm.CreatePatchBaselineInput{
		ApprovedPatchesComplianceLevel: awstypes.PatchComplianceLevel(d.Get("approved_patches_compliance_level").(string)),
		Name:                           aws.String(name),
		OperatingSystem:                awstypes.OperatingSystem(d.Get("operating_system").(string)),
		Tags:                           getTagsIn(ctx),
	}

	if _, ok := d.GetOk("approval_rule"); ok {
		input.ApprovalRules = expandPatchRuleGroup(d)
	}

	if v, ok := d.GetOk("approved_patches"); ok && v.(*schema.Set).Len() > 0 {
		input.ApprovedPatches = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("approved_patches_enable_non_security"); ok {
		input.ApprovedPatchesEnableNonSecurity = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if _, ok := d.GetOk("global_filter"); ok {
		input.GlobalFilters = expandPatchFilterGroup(d)
	}

	if v, ok := d.GetOk("rejected_patches"); ok && v.(*schema.Set).Len() > 0 {
		input.RejectedPatches = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("rejected_patches_action"); ok {
		input.RejectedPatchesAction = awstypes.PatchAction(v.(string))
	}

	if _, ok := d.GetOk(names.AttrSource); ok {
		input.Sources = expandPatchSource(d)
	}

	output, err := conn.CreatePatchBaseline(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SSM Patch Baseline (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.BaselineId))

	return append(diags, resourcePatchBaselineRead(ctx, d, meta)...)
}

func resourcePatchBaselineRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	output, err := findPatchBaselineByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SSM Patch Baseline (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SSM Patch Baseline (%s): %s", d.Id(), err)
	}

	jsonDoc, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	jsonString := string(jsonDoc)

	if err := d.Set("approval_rule", flattenPatchRuleGroup(output.ApprovalRules)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting approval_rule: %s", err)
	}
	d.Set("approved_patches", output.ApprovedPatches)
	d.Set("approved_patches_compliance_level", output.ApprovedPatchesComplianceLevel)
	d.Set("approved_patches_enable_non_security", output.ApprovedPatchesEnableNonSecurity)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Service:   "ssm",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  "patchbaseline/" + strings.TrimPrefix(d.Id(), "/"),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrDescription, output.Description)
	if err := d.Set("global_filter", flattenPatchFilterGroup(output.GlobalFilters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting global_filter: %s", err)
	}
	d.Set(names.AttrJSON, jsonString)
	d.Set(names.AttrName, output.Name)
	d.Set("operating_system", output.OperatingSystem)
	d.Set("rejected_patches", output.RejectedPatches)
	d.Set("rejected_patches_action", output.RejectedPatchesAction)
	if err := d.Set(names.AttrSource, flattenPatchSource(output.Sources)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting source: %s", err)
	}

	return diags
}

func resourcePatchBaselineUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &ssm.UpdatePatchBaselineInput{
			BaselineId: aws.String(d.Id()),
		}

		if d.HasChange("approval_rule") {
			input.ApprovalRules = expandPatchRuleGroup(d)
		}

		if d.HasChange("approved_patches") {
			input.ApprovedPatches = flex.ExpandStringValueSet(d.Get("approved_patches").(*schema.Set))
		}

		if d.HasChange("approved_patches_compliance_level") {
			input.ApprovedPatchesComplianceLevel = awstypes.PatchComplianceLevel(d.Get("approved_patches_compliance_level").(string))
		}

		if d.HasChange("approved_patches_enable_non_security") {
			input.ApprovedPatchesEnableNonSecurity = aws.Bool(d.Get("approved_patches_enable_non_security").(bool))
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChange("global_filter") {
			input.GlobalFilters = expandPatchFilterGroup(d)
		}

		if d.HasChange(names.AttrName) {
			input.Name = aws.String(d.Get(names.AttrName).(string))
		}

		if d.HasChange("rejected_patches") {
			input.RejectedPatches = flex.ExpandStringValueSet(d.Get("rejected_patches").(*schema.Set))
		}

		if d.HasChange("rejected_patches_action") {
			input.RejectedPatchesAction = awstypes.PatchAction(d.Get("rejected_patches_action").(string))
		}

		if d.HasChange(names.AttrSource) {
			input.Sources = expandPatchSource(d)
		}

		_, err := conn.UpdatePatchBaseline(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SSM Patch Baseline (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourcePatchBaselineRead(ctx, d, meta)...)
}

func resourcePatchBaselineDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	log.Printf("[INFO] Deleting SSM Patch Baseline: %s", d.Id())
	input := &ssm.DeletePatchBaselineInput{
		BaselineId: aws.String(d.Id()),
	}

	_, err := conn.DeletePatchBaseline(ctx, input)

	if errs.IsA[*awstypes.ResourceInUseException](err) {
		// Reset the default patch baseline before retrying.
		diags = append(diags, defaultPatchBaselineRestoreOSDefault(ctx, meta.(*conns.AWSClient).SSMClient(ctx), awstypes.OperatingSystem(d.Get("operating_system").(string)))...)
		if diags.HasError() {
			return diags
		}

		_, err = conn.DeletePatchBaseline(ctx, input)
	}

	if err != nil {
		diags = sdkdiag.AppendErrorf(diags, "deleting SSM Patch Baseline (%s): %s", d.Id(), err)
	}

	return diags
}

func findPatchBaselineByID(ctx context.Context, conn *ssm.Client, id string) (*ssm.GetPatchBaselineOutput, error) {
	input := &ssm.GetPatchBaselineInput{
		BaselineId: aws.String(id),
	}

	output, err := conn.GetPatchBaseline(ctx, input)

	if errs.IsA[*awstypes.DoesNotExistException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func expandPatchFilterGroup(d *schema.ResourceData) *awstypes.PatchFilterGroup {
	var filters []awstypes.PatchFilter

	tfList := d.Get("global_filter").([]interface{})

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]interface{})

		filter := awstypes.PatchFilter{
			Key:    awstypes.PatchFilterKey(tfMap[names.AttrKey].(string)),
			Values: flex.ExpandStringValueList(tfMap[names.AttrValues].([]interface{})),
		}

		filters = append(filters, filter)
	}

	return &awstypes.PatchFilterGroup{
		PatchFilters: filters,
	}
}

func flattenPatchFilterGroup(apiObject *awstypes.PatchFilterGroup) []interface{} {
	if len(apiObject.PatchFilters) == 0 {
		return nil
	}

	tfList := make([]interface{}, 0, len(apiObject.PatchFilters))

	for _, apiObject := range apiObject.PatchFilters {
		tfMap := make(map[string]interface{})
		tfMap[names.AttrKey] = apiObject.Key
		tfMap[names.AttrValues] = apiObject.Values

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func expandPatchRuleGroup(d *schema.ResourceData) *awstypes.PatchRuleGroup {
	var rules []awstypes.PatchRule

	tfList := d.Get("approval_rule").([]interface{})

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]interface{})

		var filters []awstypes.PatchFilter
		tfList := tfMap["patch_filter"].([]interface{})

		for _, tfMapRaw := range tfList {
			tfMap := tfMapRaw.(map[string]interface{})

			filter := awstypes.PatchFilter{
				Key:    awstypes.PatchFilterKey(tfMap[names.AttrKey].(string)),
				Values: flex.ExpandStringValueList(tfMap[names.AttrValues].([]interface{})),
			}

			filters = append(filters, filter)
		}

		filterGroup := &awstypes.PatchFilterGroup{
			PatchFilters: filters,
		}

		rule := awstypes.PatchRule{
			ComplianceLevel:   awstypes.PatchComplianceLevel(tfMap["compliance_level"].(string)),
			EnableNonSecurity: aws.Bool(tfMap["enable_non_security"].(bool)),
			PatchFilterGroup:  filterGroup,
		}

		if v, ok := tfMap["approve_until_date"].(string); ok && v != "" {
			rule.ApproveUntilDate = aws.String(v)
		} else if v, ok := tfMap["approve_after_days"].(int); ok {
			rule.ApproveAfterDays = aws.Int32(int32(v))
		}

		rules = append(rules, rule)
	}

	return &awstypes.PatchRuleGroup{
		PatchRules: rules,
	}
}

func flattenPatchRuleGroup(apiObject *awstypes.PatchRuleGroup) []interface{} {
	if len(apiObject.PatchRules) == 0 {
		return nil
	}

	tfList := make([]interface{}, 0, len(apiObject.PatchRules))

	for _, apiObject := range apiObject.PatchRules {
		tfMap := make(map[string]interface{})
		tfMap["compliance_level"] = apiObject.ComplianceLevel
		tfMap["enable_non_security"] = aws.ToBool(apiObject.EnableNonSecurity)
		tfMap["patch_filter"] = flattenPatchFilterGroup(apiObject.PatchFilterGroup)

		if apiObject.ApproveAfterDays != nil {
			tfMap["approve_after_days"] = aws.ToInt32(apiObject.ApproveAfterDays)
		}

		if apiObject.ApproveUntilDate != nil {
			tfMap["approve_until_date"] = aws.ToString(apiObject.ApproveUntilDate)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func expandPatchSource(d *schema.ResourceData) []awstypes.PatchSource {
	var apiObjects []awstypes.PatchSource

	tfList := d.Get(names.AttrSource).([]interface{})

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]interface{})

		apiObject := awstypes.PatchSource{
			Configuration: aws.String(tfMap[names.AttrConfiguration].(string)),
			Name:          aws.String(tfMap[names.AttrName].(string)),
			Products:      flex.ExpandStringValueList(tfMap["products"].([]interface{})),
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenPatchSource(apiObjects []awstypes.PatchSource) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	tfList := make([]interface{}, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfMap := make(map[string]interface{})

		tfMap[names.AttrConfiguration] = aws.ToString(apiObject.Configuration)
		tfMap[names.AttrName] = aws.ToString(apiObject.Name)
		tfMap["products"] = apiObject.Products

		tfList = append(tfList, tfMap)
	}

	return tfList
}
