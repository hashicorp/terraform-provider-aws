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

// @SDKResource("aws_config_organization_custom_rule", name="Organization Custom Rule")
func resourceOrganizationCustomRule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceOrganizationCustomRuleCreate,
		ReadWithoutTimeout:   resourceOrganizationCustomRuleRead,
		UpdateWithoutTimeout: resourceOrganizationCustomRuleUpdate,
		DeleteWithoutTimeout: resourceOrganizationCustomRuleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
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
			"lambda_function_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
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
					ValidateDiagFunc: enum.Validate[types.OrganizationConfigRuleTriggerType](),
				},
			},
		},
	}
}

func resourceOrganizationCustomRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &configservice.PutOrganizationConfigRuleInput{
		OrganizationConfigRuleName: aws.String(name),
		OrganizationCustomRuleMetadata: &types.OrganizationCustomRuleMetadata{
			LambdaFunctionArn:                  aws.String(d.Get("lambda_function_arn").(string)),
			OrganizationConfigRuleTriggerTypes: flex.ExpandStringyValueSet[types.OrganizationConfigRuleTriggerType](d.Get("trigger_types").(*schema.Set)),
		},
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.OrganizationCustomRuleMetadata.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("excluded_accounts"); ok && v.(*schema.Set).Len() > 0 {
		input.ExcludedAccounts = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("input_parameters"); ok {
		input.OrganizationCustomRuleMetadata.InputParameters = aws.String(v.(string))
	}

	if v, ok := d.GetOk("maximum_execution_frequency"); ok {
		input.OrganizationCustomRuleMetadata.MaximumExecutionFrequency = types.MaximumExecutionFrequency(v.(string))
	}

	if v, ok := d.GetOk("resource_id_scope"); ok {
		input.OrganizationCustomRuleMetadata.ResourceIdScope = aws.String(v.(string))
	}

	if v, ok := d.GetOk("resource_types_scope"); ok && v.(*schema.Set).Len() > 0 {
		input.OrganizationCustomRuleMetadata.ResourceTypesScope = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("tag_key_scope"); ok {
		input.OrganizationCustomRuleMetadata.TagKeyScope = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tag_value_scope"); ok {
		input.OrganizationCustomRuleMetadata.TagValueScope = aws.String(v.(string))
	}

	_, err := tfresource.RetryWhenIsA[*types.OrganizationAccessDeniedException](ctx, organizationsPropagationTimeout, func() (interface{}, error) {
		return conn.PutOrganizationConfigRule(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ConfigService Organization Custom Rule (%s): %s", name, err)
	}

	d.SetId(name)

	if _, err := waitOrganizationConfigRuleCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ConfigService Organization Custom Rule (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceOrganizationCustomRuleRead(ctx, d, meta)...)
}

func resourceOrganizationCustomRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceClient(ctx)

	configRule, err := findOrganizationCustomRuleByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ConfigService Organization Custom Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ConfigService Organization Custom Rule (%s): %s", d.Id(), err)
	}

	customRule := configRule.OrganizationCustomRuleMetadata
	d.Set(names.AttrARN, configRule.OrganizationConfigRuleArn)
	d.Set(names.AttrDescription, customRule.Description)
	d.Set("excluded_accounts", configRule.ExcludedAccounts)
	d.Set("input_parameters", customRule.InputParameters)
	d.Set("lambda_function_arn", customRule.LambdaFunctionArn)
	d.Set("maximum_execution_frequency", customRule.MaximumExecutionFrequency)
	d.Set(names.AttrName, configRule.OrganizationConfigRuleName)
	d.Set("resource_id_scope", customRule.ResourceIdScope)
	d.Set("resource_types_scope", customRule.ResourceTypesScope)
	d.Set("tag_key_scope", customRule.TagKeyScope)
	d.Set("tag_value_scope", customRule.TagValueScope)
	d.Set("trigger_types", customRule.OrganizationConfigRuleTriggerTypes)

	return diags
}

func resourceOrganizationCustomRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceClient(ctx)

	input := &configservice.PutOrganizationConfigRuleInput{
		OrganizationConfigRuleName: aws.String(d.Id()),
		OrganizationCustomRuleMetadata: &types.OrganizationCustomRuleMetadata{
			LambdaFunctionArn:                  aws.String(d.Get("lambda_function_arn").(string)),
			OrganizationConfigRuleTriggerTypes: flex.ExpandStringyValueSet[types.OrganizationConfigRuleTriggerType](d.Get("trigger_types").(*schema.Set)),
		},
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.OrganizationCustomRuleMetadata.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("excluded_accounts"); ok && v.(*schema.Set).Len() > 0 {
		input.ExcludedAccounts = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("input_parameters"); ok {
		input.OrganizationCustomRuleMetadata.InputParameters = aws.String(v.(string))
	}

	if v, ok := d.GetOk("maximum_execution_frequency"); ok {
		input.OrganizationCustomRuleMetadata.MaximumExecutionFrequency = types.MaximumExecutionFrequency(v.(string))
	}

	if v, ok := d.GetOk("resource_id_scope"); ok {
		input.OrganizationCustomRuleMetadata.ResourceIdScope = aws.String(v.(string))
	}

	if v, ok := d.GetOk("resource_types_scope"); ok && v.(*schema.Set).Len() > 0 {
		input.OrganizationCustomRuleMetadata.ResourceTypesScope = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("tag_key_scope"); ok {
		input.OrganizationCustomRuleMetadata.TagKeyScope = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tag_value_scope"); ok {
		input.OrganizationCustomRuleMetadata.TagValueScope = aws.String(v.(string))
	}

	_, err := conn.PutOrganizationConfigRule(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating ConfigService Organization Custom Rule (%s): %s", d.Id(), err)
	}

	if _, err := waitOrganizationConfigRuleUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ConfigService Organization Custom Rule (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceOrganizationCustomRuleRead(ctx, d, meta)...)
}

func resourceOrganizationCustomRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceClient(ctx)

	const (
		timeout = 2 * time.Minute
	)
	log.Printf("[DEBUG] Deleting ConfigService Organization Custom Rule: %s", d.Id())
	_, err := tfresource.RetryWhenIsA[*types.ResourceInUseException](ctx, timeout, func() (interface{}, error) {
		return conn.DeleteOrganizationConfigRule(ctx, &configservice.DeleteOrganizationConfigRuleInput{
			OrganizationConfigRuleName: aws.String(d.Id()),
		})
	})

	if errs.IsA[*types.NoSuchOrganizationConfigRuleException](err) || errs.IsA[*types.OrganizationAccessDeniedException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ConfigService Organization Custom Rule (%s): %s", d.Id(), err)
	}

	if _, err := waitOrganizationConfigRuleDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ConfigService Organization Custom Rule (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findOrganizationCustomRuleByName(ctx context.Context, conn *configservice.Client, name string) (*types.OrganizationConfigRule, error) {
	output, err := findOrganizationConfigRuleByName(ctx, conn, name)

	if err != nil {
		return nil, err
	}

	if output.OrganizationCustomRuleMetadata == nil {
		return nil, tfresource.NewEmptyResultError(nil)
	}

	return output, nil
}
