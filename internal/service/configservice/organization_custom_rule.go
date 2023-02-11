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
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceOrganizationCustomRule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceOrganizationCustomRuleCreate,
		DeleteWithoutTimeout: resourceOrganizationCustomRuleDelete,
		ReadWithoutTimeout:   resourceOrganizationCustomRuleRead,
		UpdateWithoutTimeout: resourceOrganizationCustomRuleUpdate,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
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
			"lambda_function_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			"maximum_execution_frequency": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					configservice.MaximumExecutionFrequencyOneHour,
					configservice.MaximumExecutionFrequencyThreeHours,
					configservice.MaximumExecutionFrequencySixHours,
					configservice.MaximumExecutionFrequencyTwelveHours,
					configservice.MaximumExecutionFrequencyTwentyFourHours,
				}, false),
			},
			"name": {
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
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						"ConfigurationItemChangeNotification",
						"OversizedConfigurationItemChangeNotification",
						"ScheduledNotification",
					}, false),
				},
			},
		},
	}
}

func resourceOrganizationCustomRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceConn()
	name := d.Get("name").(string)

	input := &configservice.PutOrganizationConfigRuleInput{
		OrganizationConfigRuleName: aws.String(name),
		OrganizationCustomRuleMetadata: &configservice.OrganizationCustomRuleMetadata{
			LambdaFunctionArn:                  aws.String(d.Get("lambda_function_arn").(string)),
			OrganizationConfigRuleTriggerTypes: flex.ExpandStringSet(d.Get("trigger_types").(*schema.Set)),
		},
	}

	if v, ok := d.GetOk("description"); ok {
		input.OrganizationCustomRuleMetadata.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("excluded_accounts"); ok && v.(*schema.Set).Len() > 0 {
		input.ExcludedAccounts = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("input_parameters"); ok {
		input.OrganizationCustomRuleMetadata.InputParameters = aws.String(v.(string))
	}

	if v, ok := d.GetOk("maximum_execution_frequency"); ok {
		input.OrganizationCustomRuleMetadata.MaximumExecutionFrequency = aws.String(v.(string))
	}

	if v, ok := d.GetOk("resource_id_scope"); ok {
		input.OrganizationCustomRuleMetadata.ResourceIdScope = aws.String(v.(string))
	}

	if v, ok := d.GetOk("resource_types_scope"); ok && v.(*schema.Set).Len() > 0 {
		input.OrganizationCustomRuleMetadata.ResourceTypesScope = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("tag_key_scope"); ok {
		input.OrganizationCustomRuleMetadata.TagKeyScope = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tag_value_scope"); ok {
		input.OrganizationCustomRuleMetadata.TagValueScope = aws.String(v.(string))
	}

	_, err := conn.PutOrganizationConfigRuleWithContext(ctx, input)

	if err != nil {
		return create.DiagError(names.ConfigService, create.ErrActionCreating, ResNameOrganizationCustomRule, name, err)
	}

	d.SetId(name)

	if err := waitForOrganizationRuleStatusCreateSuccessful(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.DiagError(names.ConfigService, create.ErrActionWaitingForCreation, ResNameOrganizationCustomRule, d.Id(), err)
	}

	return append(diags, resourceOrganizationCustomRuleRead(ctx, d, meta)...)
}

func resourceOrganizationCustomRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceConn()

	rule, err := DescribeOrganizationConfigRule(ctx, conn, d.Id())

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, configservice.ErrCodeNoSuchOrganizationConfigRuleException) {
		log.Printf("[WARN] Config Organization Custom Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.DiagError(names.ConfigService, create.ErrActionReading, ResNameOrganizationCustomRule, d.Id(), err)
	}

	if !d.IsNewResource() && rule == nil {
		log.Printf("[WARN] Config Organization Custom Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if d.IsNewResource() && rule == nil {
		return create.DiagError(names.ConfigService, create.ErrActionReading, ResNameOrganizationCustomRule, d.Id(), errors.New("empty rule after creation"))
	}

	if rule.OrganizationManagedRuleMetadata != nil {
		return create.DiagError(names.ConfigService, create.ErrActionReading, ResNameOrganizationCustomRule, d.Id(), errors.New("expected Organization Custom Rule, found Organization Managed Rule"))
	}

	if rule.OrganizationCustomRuleMetadata == nil {
		return create.DiagError(names.ConfigService, create.ErrActionReading, ResNameOrganizationCustomRule, d.Id(), errors.New("empty metadata"))
	}

	d.Set("arn", rule.OrganizationConfigRuleArn)
	d.Set("description", rule.OrganizationCustomRuleMetadata.Description)

	if err := d.Set("excluded_accounts", aws.StringValueSlice(rule.ExcludedAccounts)); err != nil {
		return create.DiagError(names.ConfigService, create.ErrActionSetting, ResNameOrganizationCustomRule, d.Id(), err)
	}

	d.Set("input_parameters", rule.OrganizationCustomRuleMetadata.InputParameters)
	d.Set("lambda_function_arn", rule.OrganizationCustomRuleMetadata.LambdaFunctionArn)
	d.Set("maximum_execution_frequency", rule.OrganizationCustomRuleMetadata.MaximumExecutionFrequency)
	d.Set("name", rule.OrganizationConfigRuleName)
	d.Set("resource_id_scope", rule.OrganizationCustomRuleMetadata.ResourceIdScope)

	if err := d.Set("resource_types_scope", aws.StringValueSlice(rule.OrganizationCustomRuleMetadata.ResourceTypesScope)); err != nil {
		return create.DiagError(names.ConfigService, create.ErrActionSetting, ResNameOrganizationCustomRule, d.Id(), err)
	}

	d.Set("tag_key_scope", rule.OrganizationCustomRuleMetadata.TagKeyScope)
	d.Set("tag_value_scope", rule.OrganizationCustomRuleMetadata.TagValueScope)

	if err := d.Set("trigger_types", aws.StringValueSlice(rule.OrganizationCustomRuleMetadata.OrganizationConfigRuleTriggerTypes)); err != nil {
		return create.DiagError(names.ConfigService, create.ErrActionSetting, ResNameOrganizationCustomRule, d.Id(), err)
	}

	return diags
}

func resourceOrganizationCustomRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceConn()

	input := &configservice.PutOrganizationConfigRuleInput{
		OrganizationConfigRuleName: aws.String(d.Id()),
		OrganizationCustomRuleMetadata: &configservice.OrganizationCustomRuleMetadata{
			LambdaFunctionArn:                  aws.String(d.Get("lambda_function_arn").(string)),
			OrganizationConfigRuleTriggerTypes: flex.ExpandStringSet(d.Get("trigger_types").(*schema.Set)),
		},
	}

	if v, ok := d.GetOk("description"); ok {
		input.OrganizationCustomRuleMetadata.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("excluded_accounts"); ok && v.(*schema.Set).Len() > 0 {
		input.ExcludedAccounts = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("input_parameters"); ok {
		input.OrganizationCustomRuleMetadata.InputParameters = aws.String(v.(string))
	}

	if v, ok := d.GetOk("maximum_execution_frequency"); ok {
		input.OrganizationCustomRuleMetadata.MaximumExecutionFrequency = aws.String(v.(string))
	}

	if v, ok := d.GetOk("resource_id_scope"); ok {
		input.OrganizationCustomRuleMetadata.ResourceIdScope = aws.String(v.(string))
	}

	if v, ok := d.GetOk("resource_types_scope"); ok && v.(*schema.Set).Len() > 0 {
		input.OrganizationCustomRuleMetadata.ResourceTypesScope = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("tag_key_scope"); ok {
		input.OrganizationCustomRuleMetadata.TagKeyScope = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tag_value_scope"); ok {
		input.OrganizationCustomRuleMetadata.TagValueScope = aws.String(v.(string))
	}

	_, err := conn.PutOrganizationConfigRuleWithContext(ctx, input)

	if err != nil {
		return create.DiagError(names.ConfigService, create.ErrActionUpdating, ResNameOrganizationCustomRule, d.Id(), err)
	}

	if err := waitForOrganizationRuleStatusUpdateSuccessful(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return create.DiagError(names.ConfigService, create.ErrActionWaitingForUpdate, ResNameOrganizationCustomRule, d.Id(), err)
	}

	return append(diags, resourceOrganizationCustomRuleRead(ctx, d, meta)...)
}

func resourceOrganizationCustomRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceConn()

	input := &configservice.DeleteOrganizationConfigRuleInput{
		OrganizationConfigRuleName: aws.String(d.Id()),
	}

	_, err := conn.DeleteOrganizationConfigRuleWithContext(ctx, input)

	if err != nil {
		return create.DiagError(names.ConfigService, create.ErrActionDeleting, ResNameOrganizationCustomRule, d.Id(), err)
	}

	if err := waitForOrganizationRuleStatusDeleteSuccessful(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.DiagError(names.ConfigService, create.ErrActionWaitingForDeletion, ResNameOrganizationCustomRule, d.Id(), err)
	}

	return diags
}
