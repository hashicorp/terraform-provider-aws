package budgets

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/budgets"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceBudgetAction() *schema.Resource {
	return &schema.Resource{
		Create: resourceBudgetActionCreate,
		Read:   resourceBudgetActionRead,
		Update: resourceBudgetActionUpdate,
		Delete: resourceBudgetActionDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"account_id": {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"action_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"action_threshold": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action_threshold_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(budgets.ThresholdType_Values(), false),
						},
						"action_threshold_value": {
							Type:         schema.TypeFloat,
							Required:     true,
							ValidateFunc: validation.FloatBetween(0, 40000000000),
						},
					},
				},
			},
			"action_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(budgets.ActionType_Values(), false),
			},
			"approval_model": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(budgets.ApprovalModel_Values(), false),
			},
			"budget_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 100),
					validation.StringMatch(regexp.MustCompile(`[^:\\]+`), "The ':' and '\\' characters aren't allowed."),
				),
			},
			"definition": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"iam_action_definition": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"policy_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
									"groups": {
										Type:     schema.TypeSet,
										Optional: true,
										MaxItems: 100,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"roles": {
										Type:     schema.TypeSet,
										Optional: true,
										MaxItems: 100,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"users": {
										Type:     schema.TypeSet,
										Optional: true,
										MaxItems: 100,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"ssm_action_definition": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"action_sub_type": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(budgets.ActionSubType_Values(), false),
									},
									"instance_ids": {
										Type:     schema.TypeSet,
										Required: true,
										MaxItems: 100,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"region": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"scp_action_definition": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"policy_id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"target_ids": {
										Type:     schema.TypeSet,
										Required: true,
										MaxItems: 100,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
					},
				},
			},
			"execution_role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"notification_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(budgets.NotificationType_Values(), false),
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subscriber": {
				Type:     schema.TypeSet,
				Required: true,
				MaxItems: 11,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 2147483647),
								validation.StringMatch(regexp.MustCompile(`(.*[\n\r\t\f\ ]?)*`), "Can't contain line breaks."),
							)},
						"subscription_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(budgets.SubscriptionType_Values(), false),
						},
					},
				},
			},
		},
	}
}

func resourceBudgetActionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BudgetsConn

	accountID := d.Get("account_id").(string)
	if accountID == "" {
		accountID = meta.(*conns.AWSClient).AccountID
	}

	input := &budgets.CreateBudgetActionInput{
		AccountId:        aws.String(accountID),
		ActionThreshold:  expandBudgetActionActionThreshold(d.Get("action_threshold").([]interface{})),
		ActionType:       aws.String(d.Get("action_type").(string)),
		ApprovalModel:    aws.String(d.Get("approval_model").(string)),
		BudgetName:       aws.String(d.Get("budget_name").(string)),
		Definition:       expandBudgetActionActionDefinition(d.Get("definition").([]interface{})),
		ExecutionRoleArn: aws.String(d.Get("execution_role_arn").(string)),
		NotificationType: aws.String(d.Get("notification_type").(string)),
		Subscribers:      expandBudgetActionSubscriber(d.Get("subscriber").(*schema.Set)),
	}

	log.Printf("[DEBUG] Creating Budget Action: %s", input)
	outputRaw, err := tfresource.RetryWhenAWSErrCodeEquals(propagationTimeout, func() (interface{}, error) {
		return conn.CreateBudgetAction(input)
	}, budgets.ErrCodeAccessDeniedException)

	if err != nil {
		return fmt.Errorf("error creating Budget Action: %w", err)
	}

	output := outputRaw.(*budgets.CreateBudgetActionOutput)
	actionID := aws.StringValue(output.ActionId)
	budgetName := aws.StringValue(output.BudgetName)

	d.SetId(BudgetActionCreateResourceID(accountID, actionID, budgetName))

	if _, err := waitActionAvailable(conn, accountID, actionID, budgetName); err != nil {
		return fmt.Errorf("error waiting for Budget Action (%s) to create: %w", d.Id(), err)
	}

	return resourceBudgetActionRead(d, meta)
}

func resourceBudgetActionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BudgetsConn

	accountID, actionID, budgetName, err := BudgetActionParseResourceID(d.Id())

	if err != nil {
		return err
	}

	output, err := FindActionByAccountIDActionIDAndBudgetName(conn, accountID, actionID, budgetName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Budget Action (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Budget Action (%s): %w", d.Id(), err)
	}

	d.Set("account_id", accountID)
	d.Set("action_id", actionID)

	if err := d.Set("action_threshold", flattenBudgetActionActionThreshold(output.ActionThreshold)); err != nil {
		return fmt.Errorf("error setting action_threshold: %w", err)
	}

	d.Set("action_type", output.ActionType)
	d.Set("approval_model", output.ApprovalModel)
	d.Set("budget_name", budgetName)

	if err := d.Set("definition", flattenBudgetActionDefinition(output.Definition)); err != nil {
		return fmt.Errorf("error setting definition: %w", err)
	}

	d.Set("execution_role_arn", output.ExecutionRoleArn)
	d.Set("notification_type", output.NotificationType)
	d.Set("status", output.Status)

	if err := d.Set("subscriber", flattenBudgetActionSubscriber(output.Subscribers)); err != nil {
		return fmt.Errorf("error setting subscriber: %w", err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "budgets",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("budget/%s/action/%s", budgetName, actionID),
	}
	d.Set("arn", arn.String())

	return nil
}

func resourceBudgetActionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BudgetsConn

	accountID, actionID, budgetName, err := BudgetActionParseResourceID(d.Id())

	if err != nil {
		return err
	}

	input := &budgets.UpdateBudgetActionInput{
		AccountId:  aws.String(accountID),
		ActionId:   aws.String(actionID),
		BudgetName: aws.String(budgetName),
	}

	if d.HasChange("action_threshold") {
		input.ActionThreshold = expandBudgetActionActionThreshold(d.Get("action_threshold").([]interface{}))
	}

	if d.HasChange("approval_model") {
		input.ApprovalModel = aws.String(d.Get("approval_model").(string))
	}

	if d.HasChange("definition") {
		input.Definition = expandBudgetActionActionDefinition(d.Get("definition").([]interface{}))
	}

	if d.HasChange("execution_role_arn") {
		input.ExecutionRoleArn = aws.String(d.Get("execution_role_arn").(string))
	}

	if d.HasChange("notification_type") {
		input.NotificationType = aws.String(d.Get("notification_type").(string))
	}

	if d.HasChange("subscriber") {
		input.Subscribers = expandBudgetActionSubscriber(d.Get("subscriber").(*schema.Set))
	}

	log.Printf("[DEBUG] Updating Budget Action: %s", input)
	_, err = conn.UpdateBudgetAction(input)

	if err != nil {
		return fmt.Errorf("error updating Budget Action (%s): %w", d.Id(), err)
	}

	if _, err := waitActionAvailable(conn, accountID, actionID, budgetName); err != nil {
		return fmt.Errorf("error waiting for Budget Action (%s) to update: %w", d.Id(), err)
	}

	return resourceBudgetActionRead(d, meta)
}

func resourceBudgetActionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BudgetsConn

	accountID, actionID, budgetName, err := BudgetActionParseResourceID(d.Id())

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Deleting Budget Action: %s", d.Id())
	_, err = conn.DeleteBudgetAction(&budgets.DeleteBudgetActionInput{
		AccountId:  aws.String(accountID),
		ActionId:   aws.String(actionID),
		BudgetName: aws.String(budgetName),
	})

	if tfawserr.ErrCodeEquals(err, budgets.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Budget Action (%s): %w", d.Id(), err)
	}

	return nil
}

func expandBudgetActionActionThreshold(l []interface{}) *budgets.ActionThreshold {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &budgets.ActionThreshold{}

	if v, ok := m["action_threshold_type"].(string); ok && v != "" {
		config.ActionThresholdType = aws.String(v)
	}

	if v, ok := m["action_threshold_value"].(float64); ok {
		config.ActionThresholdValue = aws.Float64(v)
	}

	return config
}

func expandBudgetActionSubscriber(l *schema.Set) []*budgets.Subscriber {
	if l.Len() == 0 {
		return []*budgets.Subscriber{}
	}

	items := []*budgets.Subscriber{}

	for _, m := range l.List() {
		config := &budgets.Subscriber{}
		raw := m.(map[string]interface{})

		if v, ok := raw["address"].(string); ok && v != "" {
			config.Address = aws.String(v)
		}

		if v, ok := raw["subscription_type"].(string); ok {
			config.SubscriptionType = aws.String(v)
		}

		items = append(items, config)
	}

	return items
}

func expandBudgetActionActionDefinition(l []interface{}) *budgets.Definition {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &budgets.Definition{}

	if v, ok := m["ssm_action_definition"].([]interface{}); ok && len(v) > 0 {
		config.SsmActionDefinition = expandBudgetActionActionSSMActionDefinition(v)
	}

	if v, ok := m["scp_action_definition"].([]interface{}); ok && len(v) > 0 {
		config.ScpActionDefinition = expandBudgetActionActionScpActionDefinition(v)
	}

	if v, ok := m["iam_action_definition"].([]interface{}); ok && len(v) > 0 {
		config.IamActionDefinition = expandBudgetActionActionIAMActionDefinition(v)
	}

	return config
}

func expandBudgetActionActionScpActionDefinition(l []interface{}) *budgets.ScpActionDefinition {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &budgets.ScpActionDefinition{}

	if v, ok := m["policy_id"].(string); ok && v != "" {
		config.PolicyId = aws.String(v)
	}

	if v, ok := m["target_ids"].(*schema.Set); ok && v.Len() > 0 {
		config.TargetIds = flex.ExpandStringSet(v)
	}

	return config
}

func expandBudgetActionActionSSMActionDefinition(l []interface{}) *budgets.SsmActionDefinition {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &budgets.SsmActionDefinition{}

	if v, ok := m["action_sub_type"].(string); ok && v != "" {
		config.ActionSubType = aws.String(v)
	}

	if v, ok := m["region"].(string); ok && v != "" {
		config.Region = aws.String(v)
	}

	if v, ok := m["instance_ids"].(*schema.Set); ok && v.Len() > 0 {
		config.InstanceIds = flex.ExpandStringSet(v)
	}

	return config
}

func expandBudgetActionActionIAMActionDefinition(l []interface{}) *budgets.IamActionDefinition {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &budgets.IamActionDefinition{}

	if v, ok := m["policy_arn"].(string); ok && v != "" {
		config.PolicyArn = aws.String(v)
	}

	if v, ok := m["groups"].(*schema.Set); ok && v.Len() > 0 {
		config.Groups = flex.ExpandStringSet(v)
	}

	if v, ok := m["roles"].(*schema.Set); ok && v.Len() > 0 {
		config.Roles = flex.ExpandStringSet(v)
	}

	if v, ok := m["users"].(*schema.Set); ok && v.Len() > 0 {
		config.Users = flex.ExpandStringSet(v)
	}

	return config
}

func flattenBudgetActionSubscriber(configured []*budgets.Subscriber) []map[string]interface{} {
	dataResources := make([]map[string]interface{}, 0, len(configured))

	for _, raw := range configured {
		item := make(map[string]interface{})
		item["address"] = aws.StringValue(raw.Address)
		item["subscription_type"] = aws.StringValue(raw.SubscriptionType)

		dataResources = append(dataResources, item)
	}

	return dataResources
}

func flattenBudgetActionActionThreshold(lt *budgets.ActionThreshold) []map[string]interface{} {
	if lt == nil {
		return []map[string]interface{}{}
	}

	attrs := map[string]interface{}{
		"action_threshold_type":  aws.StringValue(lt.ActionThresholdType),
		"action_threshold_value": aws.Float64Value(lt.ActionThresholdValue),
	}

	return []map[string]interface{}{attrs}
}

func flattenBudgetActionIAMActionDefinition(lt *budgets.IamActionDefinition) []map[string]interface{} {
	if lt == nil {
		return []map[string]interface{}{}
	}

	attrs := map[string]interface{}{
		"policy_arn": aws.StringValue(lt.PolicyArn),
	}

	if lt.Users != nil && len(lt.Users) > 0 {
		attrs["users"] = flex.FlattenStringSet(lt.Users)
	}

	if lt.Roles != nil && len(lt.Roles) > 0 {
		attrs["roles"] = flex.FlattenStringSet(lt.Roles)
	}

	if lt.Groups != nil && len(lt.Groups) > 0 {
		attrs["groups"] = flex.FlattenStringSet(lt.Groups)
	}

	return []map[string]interface{}{attrs}
}

func flattenBudgetActionScpActionDefinition(lt *budgets.ScpActionDefinition) []map[string]interface{} {
	if lt == nil {
		return []map[string]interface{}{}
	}

	attrs := map[string]interface{}{
		"policy_id": aws.StringValue(lt.PolicyId),
	}

	if lt.TargetIds != nil && len(lt.TargetIds) > 0 {
		attrs["target_ids"] = flex.FlattenStringSet(lt.TargetIds)
	}

	return []map[string]interface{}{attrs}
}

func flattenBudgetActionSSMActionDefinition(lt *budgets.SsmActionDefinition) []map[string]interface{} {
	if lt == nil {
		return []map[string]interface{}{}
	}

	attrs := map[string]interface{}{
		"action_sub_type": aws.StringValue(lt.ActionSubType),
		"instance_ids":    flex.FlattenStringSet(lt.InstanceIds),
		"region":          aws.StringValue(lt.Region),
	}

	return []map[string]interface{}{attrs}
}

func flattenBudgetActionDefinition(lt *budgets.Definition) []map[string]interface{} {
	if lt == nil {
		return []map[string]interface{}{}
	}

	attrs := map[string]interface{}{}

	if lt.SsmActionDefinition != nil {
		attrs["ssm_action_definition"] = flattenBudgetActionSSMActionDefinition(lt.SsmActionDefinition)
	}

	if lt.IamActionDefinition != nil {
		attrs["iam_action_definition"] = flattenBudgetActionIAMActionDefinition(lt.IamActionDefinition)
	}

	if lt.ScpActionDefinition != nil {
		attrs["scp_action_definition"] = flattenBudgetActionScpActionDefinition(lt.ScpActionDefinition)
	}

	return []map[string]interface{}{attrs}
}
