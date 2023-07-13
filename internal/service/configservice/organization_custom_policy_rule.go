// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package configservice

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_config_organization_custom_policy_rule")
func ResourceOrganizationCustomPolicyRule() *schema.Resource {
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
			"arn": {
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
			"description": {
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
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 2048),
					validation.StringIsJSON,
				),
			},
			"maximum_execution_frequency": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(configservice.MaximumExecutionFrequency_Values(), false),
			},
			"name": {
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
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						"ConfigurationItemChangeNotification",
						"OversizedConfigurationItemChangeNotification",
					}, false),
				},
			},
		},
	}
}

const (
	ResNameOrganizationCustomPolicyRule = "Organization Custom Policy Rule"
)

func resourceOrganizationCustomPolicyRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConfigServiceConn(ctx)
	name := d.Get("name").(string)

	in := &configservice.PutOrganizationConfigRuleInput{
		OrganizationConfigRuleName: aws.String(name),
		OrganizationCustomPolicyRuleMetadata: &configservice.OrganizationCustomPolicyRuleMetadata{
			PolicyRuntime:                      aws.String(d.Get("policy_runtime").(string)),
			PolicyText:                         aws.String(d.Get("policy_text").(string)),
			OrganizationConfigRuleTriggerTypes: flex.ExpandStringSet(d.Get("trigger_types").(*schema.Set)),
		},
	}

	if v, ok := d.GetOk("debug_log_delivery_accounts"); ok {
		in.OrganizationCustomPolicyRuleMetadata.DebugLogDeliveryAccounts = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("description"); ok {
		in.OrganizationCustomPolicyRuleMetadata.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("excluded_accounts"); ok && v.(*schema.Set).Len() > 0 {
		in.ExcludedAccounts = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("input_parameters"); ok {
		in.OrganizationCustomPolicyRuleMetadata.InputParameters = aws.String(v.(string))
	}

	if v, ok := d.GetOk("maximum_execution_frequency"); ok {
		in.OrganizationCustomPolicyRuleMetadata.MaximumExecutionFrequency = aws.String(v.(string))
	}

	if v, ok := d.GetOk("resource_id_scope"); ok {
		in.OrganizationCustomPolicyRuleMetadata.ResourceIdScope = aws.String(v.(string))
	}

	if v, ok := d.GetOk("resource_types_scope"); ok && v.(*schema.Set).Len() > 0 {
		in.OrganizationCustomPolicyRuleMetadata.ResourceTypesScope = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("tag_key_scope"); ok {
		in.OrganizationCustomPolicyRuleMetadata.TagKeyScope = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tag_value_scope"); ok {
		in.OrganizationCustomPolicyRuleMetadata.TagValueScope = aws.String(v.(string))
	}

	out, err := conn.PutOrganizationConfigRuleWithContext(ctx, in)

	if err != nil {
		return create.DiagError(names.ConfigService, create.ErrActionCreating, ResNameOrganizationCustomPolicyRule, name, err)
	}

	if out == nil || out.OrganizationConfigRuleArn == nil {
		return create.DiagError(names.ConfigService, create.ErrActionCreating, ResNameOrganizationCustomPolicyRule, name, errors.New("empty output"))
	}

	d.SetId(name)

	if err := waitForOrganizationRuleStatusCreateSuccessful(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.DiagError(names.ConfigService, create.ErrActionWaitingForCreation, ResNameOrganizationCustomPolicyRule, d.Id(), err)
	}

	return resourceOrganizationCustomPolicyRuleRead(ctx, d, meta)
}

func resourceOrganizationCustomPolicyRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConfigServiceConn(ctx)

	rule, err := FindOrganizationConfigRule(ctx, conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Config %s (%s) not found, removing from state", ResNameOrganizationCustomPolicyRule, d.Id())
		d.SetId("")
		return nil
	}

	if rule.OrganizationManagedRuleMetadata != nil {
		return create.DiagError(names.ConfigService, create.ErrActionReading, ResNameOrganizationCustomPolicyRule, d.Id(), errors.New("expected ResNameOrganizationCustomPolicy, found Organization Managed Rule"))
	}

	if rule.OrganizationCustomRuleMetadata != nil {
		return create.DiagError(names.ConfigService, create.ErrActionReading, ResNameOrganizationCustomPolicyRule, d.Id(), errors.New("expected ResNameOrganizationCustomPolicy, found Organization Custom Rule"))
	}

	if rule.OrganizationCustomPolicyRuleMetadata == nil {
		return create.DiagError(names.ConfigService, create.ErrActionReading, ResNameOrganizationCustomPolicyRule, d.Id(), errors.New("empty metadata"))
	}

	in := &configservice.GetOrganizationCustomRulePolicyInput{
		OrganizationConfigRuleName: aws.String(d.Id()),
	}
	policy, err := conn.GetOrganizationCustomRulePolicyWithContext(ctx, in)

	if err != nil {
		return create.DiagError(names.ConfigService, create.ErrActionReading, ResNameOrganizationCustomPolicyRule, d.Id(), err)
	}

	d.Set("arn", rule.OrganizationConfigRuleArn)

	if err := d.Set("debug_log_delivery_accounts", aws.StringValueSlice(rule.OrganizationCustomPolicyRuleMetadata.DebugLogDeliveryAccounts)); err != nil {
		return create.DiagError(names.ConfigService, create.ErrActionSetting, ResNameOrganizationCustomPolicyRule, d.Id(), err)
	}

	d.Set("description", rule.OrganizationCustomPolicyRuleMetadata.Description)

	if err := d.Set("excluded_accounts", aws.StringValueSlice(rule.ExcludedAccounts)); err != nil {
		return create.DiagError(names.ConfigService, create.ErrActionSetting, ResNameOrganizationCustomPolicyRule, d.Id(), err)
	}

	d.Set("input_parameters", rule.OrganizationCustomPolicyRuleMetadata.InputParameters)
	d.Set("policy_runtime", rule.OrganizationCustomPolicyRuleMetadata.PolicyRuntime)
	d.Set("policy_text", policy.PolicyText)
	d.Set("maximum_execution_frequency", rule.OrganizationCustomPolicyRuleMetadata.MaximumExecutionFrequency)
	d.Set("name", rule.OrganizationConfigRuleName)
	d.Set("resource_id_scope", rule.OrganizationCustomPolicyRuleMetadata.ResourceIdScope)

	if err := d.Set("resource_types_scope", aws.StringValueSlice(rule.OrganizationCustomPolicyRuleMetadata.ResourceTypesScope)); err != nil {
		return create.DiagError(names.ConfigService, create.ErrActionSetting, ResNameOrganizationCustomPolicyRule, d.Id(), err)
	}

	d.Set("tag_key_scope", rule.OrganizationCustomPolicyRuleMetadata.TagKeyScope)
	d.Set("tag_value_scope", rule.OrganizationCustomPolicyRuleMetadata.TagValueScope)

	if err := d.Set("trigger_types", aws.StringValueSlice(rule.OrganizationCustomPolicyRuleMetadata.OrganizationConfigRuleTriggerTypes)); err != nil {
		return create.DiagError(names.ConfigService, create.ErrActionSetting, ResNameOrganizationCustomPolicyRule, d.Id(), err)
	}

	return nil
}

func resourceOrganizationCustomPolicyRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConfigServiceConn(ctx)

	in := &configservice.PutOrganizationConfigRuleInput{
		OrganizationConfigRuleName: aws.String(d.Id()),
		OrganizationCustomPolicyRuleMetadata: &configservice.OrganizationCustomPolicyRuleMetadata{
			PolicyText:                         aws.String(d.Get("policy_text").(string)),
			PolicyRuntime:                      aws.String(d.Get("policy_runtime").(string)),
			OrganizationConfigRuleTriggerTypes: flex.ExpandStringSet(d.Get("trigger_types").(*schema.Set)),
		},
	}

	if v, ok := d.GetOk("debug_log_delivery_accounts"); ok && v.(*schema.Set).Len() > 0 {
		in.OrganizationCustomPolicyRuleMetadata.DebugLogDeliveryAccounts = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("description"); ok {
		in.OrganizationCustomPolicyRuleMetadata.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("excluded_accounts"); ok && v.(*schema.Set).Len() > 0 {
		in.ExcludedAccounts = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("input_parameters"); ok {
		in.OrganizationCustomPolicyRuleMetadata.InputParameters = aws.String(v.(string))
	}

	if v, ok := d.GetOk("maximum_execution_frequency"); ok {
		in.OrganizationCustomPolicyRuleMetadata.MaximumExecutionFrequency = aws.String(v.(string))
	}

	if v, ok := d.GetOk("resource_id_scope"); ok {
		in.OrganizationCustomPolicyRuleMetadata.ResourceIdScope = aws.String(v.(string))
	}

	if v, ok := d.GetOk("resource_types_scope"); ok && v.(*schema.Set).Len() > 0 {
		in.OrganizationCustomPolicyRuleMetadata.ResourceTypesScope = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("tag_key_scope"); ok {
		in.OrganizationCustomPolicyRuleMetadata.TagKeyScope = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tag_value_scope"); ok {
		in.OrganizationCustomPolicyRuleMetadata.TagValueScope = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Updating ConfigService %s (%s): %#v", ResNameOrganizationCustomPolicyRule, d.Id(), in)
	_, err := conn.PutOrganizationConfigRuleWithContext(ctx, in)
	if err != nil {
		return create.DiagError(names.ConfigService, create.ErrActionUpdating, ResNameOrganizationCustomPolicyRule, d.Id(), err)
	}

	err = waitForOrganizationRuleStatusUpdateSuccessful(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return create.DiagError(names.ConfigService, create.ErrActionWaitingForUpdate, ResNameOrganizationCustomPolicyRule, d.Id(), err)
	}

	return resourceOrganizationCustomPolicyRuleRead(ctx, d, meta)
}

func resourceOrganizationCustomPolicyRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConfigServiceConn(ctx)

	log.Printf("[INFO] Deleting ConfigService %s %s", ResNameOrganizationCustomPolicyRule, d.Id())

	in := &configservice.DeleteOrganizationConfigRuleInput{
		OrganizationConfigRuleName: aws.String(d.Id()),
	}

	_, err := conn.DeleteOrganizationConfigRuleWithContext(ctx, in)

	if tfawserr.ErrCodeEquals(err, configservice.ErrCodeNoSuchOrganizationConfigRuleException) {
		return nil
	}

	if err != nil {
		return create.DiagError(names.ConfigService, create.ErrActionDeleting, ResNameOrganizationCustomPolicyRule, d.Id(), err)
	}

	if err := waitForOrganizationRuleStatusDeleteSuccessful(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.DiagError(names.ConfigService, create.ErrActionWaitingForDeletion, ResNameOrganizationCustomPolicyRule, d.Id(), err)
	}

	return nil
}
