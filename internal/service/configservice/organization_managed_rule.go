package configservice

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceOrganizationManagedRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceOrganizationManagedRuleCreate,
		Delete: resourceOrganizationManagedRuleDelete,
		Read:   resourceOrganizationManagedRuleRead,
		Update: resourceOrganizationManagedRuleUpdate,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
			"rule_identifier": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
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
		},
	}
}

func resourceOrganizationManagedRuleCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ConfigServiceConn
	name := d.Get("name").(string)

	input := &configservice.PutOrganizationConfigRuleInput{
		OrganizationConfigRuleName: aws.String(name),
		OrganizationManagedRuleMetadata: &configservice.OrganizationManagedRuleMetadata{
			RuleIdentifier: aws.String(d.Get("rule_identifier").(string)),
		},
	}

	if v, ok := d.GetOk("description"); ok {
		input.OrganizationManagedRuleMetadata.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("excluded_accounts"); ok && v.(*schema.Set).Len() > 0 {
		input.ExcludedAccounts = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("input_parameters"); ok {
		input.OrganizationManagedRuleMetadata.InputParameters = aws.String(v.(string))
	}

	if v, ok := d.GetOk("maximum_execution_frequency"); ok {
		input.OrganizationManagedRuleMetadata.MaximumExecutionFrequency = aws.String(v.(string))
	}

	if v, ok := d.GetOk("resource_id_scope"); ok {
		input.OrganizationManagedRuleMetadata.ResourceIdScope = aws.String(v.(string))
	}

	if v, ok := d.GetOk("resource_types_scope"); ok && v.(*schema.Set).Len() > 0 {
		input.OrganizationManagedRuleMetadata.ResourceTypesScope = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("tag_key_scope"); ok {
		input.OrganizationManagedRuleMetadata.TagKeyScope = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tag_value_scope"); ok {
		input.OrganizationManagedRuleMetadata.TagValueScope = aws.String(v.(string))
	}

	_, err := conn.PutOrganizationConfigRule(input)

	if err != nil {
		return fmt.Errorf("error creating Config Organization Managed Rule (%s): %s", name, err)
	}

	d.SetId(name)

	if err := waitForOrganizationRuleStatusCreateSuccessful(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for Config Organization Managed Rule (%s) creation: %s", d.Id(), err)
	}

	return resourceOrganizationManagedRuleRead(d, meta)
}

func resourceOrganizationManagedRuleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ConfigServiceConn

	rule, err := DescribeOrganizationConfigRule(conn, d.Id())

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, configservice.ErrCodeNoSuchOrganizationConfigRuleException) {
		log.Printf("[WARN] Config Organization Managed Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing Config Organization Managed Rule (%s): %s", d.Id(), err)
	}

	if !d.IsNewResource() && rule == nil {
		log.Printf("[WARN] Config Organization Managed Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if d.IsNewResource() && rule == nil {
		return names.Error(names.ConfigService, names.ErrActionReading, "Organization Managed Rule", d.Id(), errors.New("empty rule after creation"))
	}

	if rule.OrganizationCustomRuleMetadata != nil {
		return fmt.Errorf("expected Config Organization Managed Rule, found Config Organization Custom Rule: %s", d.Id())
	}

	if rule.OrganizationManagedRuleMetadata == nil {
		return fmt.Errorf("error describing Config Organization Managed Rule (%s): empty metadata", d.Id())
	}

	d.Set("arn", rule.OrganizationConfigRuleArn)
	d.Set("description", rule.OrganizationManagedRuleMetadata.Description)

	if err := d.Set("excluded_accounts", aws.StringValueSlice(rule.ExcludedAccounts)); err != nil {
		return fmt.Errorf("error setting excluded_accounts: %s", err)
	}

	d.Set("input_parameters", rule.OrganizationManagedRuleMetadata.InputParameters)
	d.Set("maximum_execution_frequency", rule.OrganizationManagedRuleMetadata.MaximumExecutionFrequency)
	d.Set("name", rule.OrganizationConfigRuleName)
	d.Set("resource_id_scope", rule.OrganizationManagedRuleMetadata.ResourceIdScope)

	if err := d.Set("resource_types_scope", aws.StringValueSlice(rule.OrganizationManagedRuleMetadata.ResourceTypesScope)); err != nil {
		return fmt.Errorf("error setting resource_types_scope: %s", err)
	}

	d.Set("rule_identifier", rule.OrganizationManagedRuleMetadata.RuleIdentifier)
	d.Set("tag_key_scope", rule.OrganizationManagedRuleMetadata.TagKeyScope)
	d.Set("tag_value_scope", rule.OrganizationManagedRuleMetadata.TagValueScope)

	return nil
}

func resourceOrganizationManagedRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ConfigServiceConn

	input := &configservice.PutOrganizationConfigRuleInput{
		OrganizationConfigRuleName: aws.String(d.Id()),
		OrganizationManagedRuleMetadata: &configservice.OrganizationManagedRuleMetadata{
			RuleIdentifier: aws.String(d.Get("rule_identifier").(string)),
		},
	}

	if v, ok := d.GetOk("description"); ok {
		input.OrganizationManagedRuleMetadata.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("excluded_accounts"); ok && v.(*schema.Set).Len() > 0 {
		input.ExcludedAccounts = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("input_parameters"); ok {
		input.OrganizationManagedRuleMetadata.InputParameters = aws.String(v.(string))
	}

	if v, ok := d.GetOk("maximum_execution_frequency"); ok {
		input.OrganizationManagedRuleMetadata.MaximumExecutionFrequency = aws.String(v.(string))
	}

	if v, ok := d.GetOk("resource_id_scope"); ok {
		input.OrganizationManagedRuleMetadata.ResourceIdScope = aws.String(v.(string))
	}

	if v, ok := d.GetOk("resource_types_scope"); ok && v.(*schema.Set).Len() > 0 {
		input.OrganizationManagedRuleMetadata.ResourceTypesScope = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("tag_key_scope"); ok {
		input.OrganizationManagedRuleMetadata.TagKeyScope = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tag_value_scope"); ok {
		input.OrganizationManagedRuleMetadata.TagValueScope = aws.String(v.(string))
	}

	_, err := conn.PutOrganizationConfigRule(input)

	if err != nil {
		return fmt.Errorf("error updating Config Organization Managed Rule (%s): %s", d.Id(), err)
	}

	if err := waitForOrganizationRuleStatusUpdateSuccessful(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return fmt.Errorf("error waiting for Config Organization Managed Rule (%s) update: %s", d.Id(), err)
	}

	return resourceOrganizationManagedRuleRead(d, meta)
}

func resourceOrganizationManagedRuleDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ConfigServiceConn

	input := &configservice.DeleteOrganizationConfigRuleInput{
		OrganizationConfigRuleName: aws.String(d.Id()),
	}

	_, err := conn.DeleteOrganizationConfigRule(input)

	if err != nil {
		return fmt.Errorf("error deleting Config Organization Managed Rule (%s): %s", d.Id(), err)
	}

	if err := waitForOrganizationRuleStatusDeleteSuccessful(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for Config Organization Managed Rule (%s) deletion: %s", d.Id(), err)
	}

	return nil
}
