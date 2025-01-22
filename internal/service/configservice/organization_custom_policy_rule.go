// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package configservice

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/aws/aws-sdk-go-v2/service/configservice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_config_organization_custom_policy_rule", name="Organization Custom Policy Rule")
func resourceOrganizationCustomPolicyRule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceOrganizationCustomPolicyRuleCreate,
		ReadWithoutTimeout:   resourceOrganizationCustomPolicyRuleRead,
		UpdateWithoutTimeout: resourceOrganizationCustomPolicyRuleUpdate,
		DeleteWithoutTimeout: resourceOrganizationCustomPolicyRuleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"debug_log_delivery_accounts": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1000,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidAccountID,
				},
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"excluded_accounts": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1000,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidAccountID,
				},
			},
			"input_parameters": {
				Type:                  schema.TypeString,
				Optional:              true,
				DiffSuppressFunc:      verify.SuppressEquivalentJSONDiffs,
				DiffSuppressOnRefresh: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 2048),
					validation.StringIsJSON,
				),
			},
			"maximum_execution_frequency": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[types.MaximumExecutionFrequency](),
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"policy_runtime": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"policy_text": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 10000),
			},
			"resource_id_scope": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 768),
			},
			"resource_types_scope": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 100,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(0, 256),
				},
			},
			"tag_key_scope": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 128),
			},
			"tag_value_scope": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"trigger_types": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				MaxItems: 3,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: enum.Validate[types.OrganizationConfigRuleTriggerTypeNoSN](),
				},
			},
		},
	}
}

func resourceOrganizationCustomPolicyRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &configservice.PutOrganizationConfigRuleInput{
		OrganizationConfigRuleName: aws.String(name),
		OrganizationCustomPolicyRuleMetadata: &types.OrganizationCustomPolicyRuleMetadata{
			OrganizationConfigRuleTriggerTypes: flex.ExpandStringyValueSet[types.OrganizationConfigRuleTriggerTypeNoSN](d.Get("trigger_types").(*schema.Set)),
			PolicyRuntime:                      aws.String(d.Get("policy_runtime").(string)),
			PolicyText:                         aws.String(d.Get("policy_text").(string)),
		},
	}

	if v, ok := d.GetOk("debug_log_delivery_accounts"); ok && v.(*schema.Set).Len() > 0 {
		input.OrganizationCustomPolicyRuleMetadata.DebugLogDeliveryAccounts = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.OrganizationCustomPolicyRuleMetadata.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("excluded_accounts"); ok && v.(*schema.Set).Len() > 0 {
		input.ExcludedAccounts = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("input_parameters"); ok {
		input.OrganizationCustomPolicyRuleMetadata.InputParameters = aws.String(v.(string))
	}

	if v, ok := d.GetOk("maximum_execution_frequency"); ok {
		input.OrganizationCustomPolicyRuleMetadata.MaximumExecutionFrequency = types.MaximumExecutionFrequency(v.(string))
	}

	if v, ok := d.GetOk("resource_id_scope"); ok {
		input.OrganizationCustomPolicyRuleMetadata.ResourceIdScope = aws.String(v.(string))
	}

	if v, ok := d.GetOk("resource_types_scope"); ok && v.(*schema.Set).Len() > 0 {
		input.OrganizationCustomPolicyRuleMetadata.ResourceTypesScope = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("tag_key_scope"); ok {
		input.OrganizationCustomPolicyRuleMetadata.TagKeyScope = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tag_value_scope"); ok {
		input.OrganizationCustomPolicyRuleMetadata.TagValueScope = aws.String(v.(string))
	}

	_, err := tfresource.RetryWhenIsA[*types.OrganizationAccessDeniedException](ctx, organizationsPropagationTimeout, func() (interface{}, error) {
		return conn.PutOrganizationConfigRule(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ConfigService Organization Custom Policy Rule (%s): %s", name, err)
	}

	d.SetId(name)

	if _, err := waitOrganizationConfigRuleCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ConfigService Organization Custom Policy Rule (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceOrganizationCustomPolicyRuleRead(ctx, d, meta)...)
}

func resourceOrganizationCustomPolicyRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceClient(ctx)

	configRule, err := findOrganizationCustomPolicyRuleByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ConfigService Organization Custom Policy Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ConfigService Organization Custom Policy Rule (%s): %s", d.Id(), err)
	}

	policy, err := findOrganizationCustomRulePolicyByName(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ConfigService Organization Custom Policy Rule (%s) policy: %s", d.Id(), err)
	}

	customPolicyRule := configRule.OrganizationCustomPolicyRuleMetadata
	d.Set(names.AttrARN, configRule.OrganizationConfigRuleArn)
	d.Set("debug_log_delivery_accounts", customPolicyRule.DebugLogDeliveryAccounts)
	d.Set(names.AttrDescription, customPolicyRule.Description)
	d.Set("excluded_accounts", configRule.ExcludedAccounts)
	d.Set("input_parameters", customPolicyRule.InputParameters)
	d.Set("policy_runtime", customPolicyRule.PolicyRuntime)
	d.Set("policy_text", policy)
	d.Set("maximum_execution_frequency", customPolicyRule.MaximumExecutionFrequency)
	d.Set(names.AttrName, configRule.OrganizationConfigRuleName)
	d.Set("resource_id_scope", customPolicyRule.ResourceIdScope)
	d.Set("resource_types_scope", customPolicyRule.ResourceTypesScope)
	d.Set("tag_key_scope", customPolicyRule.TagKeyScope)
	d.Set("tag_value_scope", customPolicyRule.TagValueScope)
	d.Set("trigger_types", customPolicyRule.OrganizationConfigRuleTriggerTypes)

	return diags
}

func resourceOrganizationCustomPolicyRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConfigServiceClient(ctx)

	input := &configservice.PutOrganizationConfigRuleInput{
		OrganizationConfigRuleName: aws.String(d.Id()),
		OrganizationCustomPolicyRuleMetadata: &types.OrganizationCustomPolicyRuleMetadata{
			OrganizationConfigRuleTriggerTypes: flex.ExpandStringyValueSet[types.OrganizationConfigRuleTriggerTypeNoSN](d.Get("trigger_types").(*schema.Set)),
			PolicyRuntime:                      aws.String(d.Get("policy_runtime").(string)),
			PolicyText:                         aws.String(d.Get("policy_text").(string)),
		},
	}

	if v, ok := d.GetOk("debug_log_delivery_accounts"); ok && v.(*schema.Set).Len() > 0 {
		input.OrganizationCustomPolicyRuleMetadata.DebugLogDeliveryAccounts = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.OrganizationCustomPolicyRuleMetadata.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("excluded_accounts"); ok && v.(*schema.Set).Len() > 0 {
		input.ExcludedAccounts = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("input_parameters"); ok {
		input.OrganizationCustomPolicyRuleMetadata.InputParameters = aws.String(v.(string))
	}

	if v, ok := d.GetOk("maximum_execution_frequency"); ok {
		input.OrganizationCustomPolicyRuleMetadata.MaximumExecutionFrequency = types.MaximumExecutionFrequency(v.(string))
	}

	if v, ok := d.GetOk("resource_id_scope"); ok {
		input.OrganizationCustomPolicyRuleMetadata.ResourceIdScope = aws.String(v.(string))
	}

	if v, ok := d.GetOk("resource_types_scope"); ok && v.(*schema.Set).Len() > 0 {
		input.OrganizationCustomPolicyRuleMetadata.ResourceTypesScope = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("tag_key_scope"); ok {
		input.OrganizationCustomPolicyRuleMetadata.TagKeyScope = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tag_value_scope"); ok {
		input.OrganizationCustomPolicyRuleMetadata.TagValueScope = aws.String(v.(string))
	}

	_, err := conn.PutOrganizationConfigRule(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating ConfigService Organization Custom Policy Rule (%s): %s", d.Id(), err)
	}

	if _, err := waitOrganizationConfigRuleUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ConfigService Organization Custom Policy Rule (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceOrganizationCustomPolicyRuleRead(ctx, d, meta)...)
}

func resourceOrganizationCustomPolicyRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceClient(ctx)

	const (
		timeout = 2 * time.Minute
	)
	log.Printf("[DEBUG] Deleting ConfigService Organization Custom Policy Rule: %s", d.Id())
	_, err := tfresource.RetryWhenIsA[*types.ResourceInUseException](ctx, timeout, func() (interface{}, error) {
		return conn.DeleteOrganizationConfigRule(ctx, &configservice.DeleteOrganizationConfigRuleInput{
			OrganizationConfigRuleName: aws.String(d.Id()),
		})
	})

	if errs.IsA[*types.NoSuchOrganizationConfigRuleException](err) || errs.IsA[*types.OrganizationAccessDeniedException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ConfigService Organization Custom Policy Rule (%s): %s", d.Id(), err)
	}

	if _, err := waitOrganizationConfigRuleDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ConfigService Organization Custom Policy Rule (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findOrganizationCustomPolicyRuleByName(ctx context.Context, conn *configservice.Client, name string) (*types.OrganizationConfigRule, error) {
	output, err := findOrganizationConfigRuleByName(ctx, conn, name)

	if err != nil {
		return nil, err
	}

	if output.OrganizationCustomPolicyRuleMetadata == nil {
		return nil, tfresource.NewEmptyResultError(nil)
	}

	return output, nil
}

func findOrganizationCustomRulePolicyByName(ctx context.Context, conn *configservice.Client, name string) (*string, error) {
	input := &configservice.GetOrganizationCustomRulePolicyInput{
		OrganizationConfigRuleName: aws.String(name),
	}

	output, err := conn.GetOrganizationCustomRulePolicy(ctx, input)

	if errs.IsA[*types.NoSuchOrganizationConfigRuleException](err) {
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

	return output.PolicyText, nil
}
