// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package budgets

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/budgets"
	awstypes "github.com/aws/aws-sdk-go-v2/service/budgets/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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

// @SDKResource("aws_budgets_budget_action")
// @Tags(identifierAttribute="arn")
func ResourceBudgetAction() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBudgetActionCreate,
		ReadWithoutTimeout:   resourceBudgetActionRead,
		UpdateWithoutTimeout: resourceBudgetActionUpdate,
		DeleteWithoutTimeout: resourceBudgetActionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute), // unneeded, but a breaking change to remove
			Update: schema.DefaultTimeout(5 * time.Minute), // unneeded, but a breaking change to remove
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrAccountID: {
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
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ThresholdType](),
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
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.ActionType](),
			},
			"approval_model": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.ApprovalModel](),
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"budget_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 100),
					validation.StringMatch(regexache.MustCompile(`[^:\\]+`), "The ':' and '\\' characters aren't allowed."),
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
									"groups": {
										Type:     schema.TypeSet,
										Optional: true,
										MaxItems: 100,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"policy_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
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
						"ssm_action_definition": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"action_sub_type": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[awstypes.ActionSubType](),
									},
									"instance_ids": {
										Type:     schema.TypeSet,
										Required: true,
										MaxItems: 100,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									names.AttrRegion: {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			names.AttrExecutionRoleARN: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"notification_type": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.NotificationType](),
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subscriber": {
				Type:     schema.TypeSet,
				Required: true,
				MaxItems: 11,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrAddress: {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 2147483647),
								validation.StringMatch(regexache.MustCompile(`(.*[\n\r\t\f\ ]?)*`), "Can't contain line breaks."),
							)},
						"subscription_type": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.SubscriptionType](),
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceBudgetActionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BudgetsClient(ctx)

	accountID := d.Get(names.AttrAccountID).(string)
	if accountID == "" {
		accountID = meta.(*conns.AWSClient).AccountID
	}
	input := &budgets.CreateBudgetActionInput{
		AccountId:        aws.String(accountID),
		ActionThreshold:  expandBudgetActionActionThreshold(d.Get("action_threshold").([]interface{})),
		ActionType:       awstypes.ActionType(d.Get("action_type").(string)),
		ApprovalModel:    awstypes.ApprovalModel(d.Get("approval_model").(string)),
		BudgetName:       aws.String(d.Get("budget_name").(string)),
		Definition:       expandBudgetActionActionDefinition(d.Get("definition").([]interface{})),
		ExecutionRoleArn: aws.String(d.Get(names.AttrExecutionRoleARN).(string)),
		NotificationType: awstypes.NotificationType(d.Get("notification_type").(string)),
		Subscribers:      expandBudgetActionSubscriber(d.Get("subscriber").(*schema.Set)),
		ResourceTags:     getTagsIn(ctx),
	}

	outputRaw, err := tfresource.RetryWhenIsA[*awstypes.AccessDeniedException](ctx, propagationTimeout, func() (interface{}, error) {
		return conn.CreateBudgetAction(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Budget Action: %s", err)
	}

	output := outputRaw.(*budgets.CreateBudgetActionOutput)
	actionID := aws.ToString(output.ActionId)
	budgetName := aws.ToString(output.BudgetName)

	d.SetId(BudgetActionCreateResourceID(accountID, actionID, budgetName))

	return append(diags, resourceBudgetActionRead(ctx, d, meta)...)
}

func resourceBudgetActionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BudgetsClient(ctx)

	accountID, actionID, budgetName, err := BudgetActionParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := FindBudgetWithDelay(ctx, func() (*awstypes.Action, error) {
		return FindActionByThreePartKey(ctx, conn, accountID, actionID, budgetName)
	})

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Budget Action (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Budget Action (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrAccountID, accountID)
	d.Set("action_id", actionID)
	if err := d.Set("action_threshold", flattenBudgetActionActionThreshold(output.ActionThreshold)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting action_threshold: %s", err)
	}
	d.Set("action_type", output.ActionType)
	d.Set("approval_model", output.ApprovalModel)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "budgets",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("budget/%s/action/%s", budgetName, actionID),
	}
	d.Set(names.AttrARN, arn.String())
	d.Set("budget_name", budgetName)
	if err := d.Set("definition", flattenBudgetActionDefinition(output.Definition)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting definition: %s", err)
	}
	d.Set(names.AttrExecutionRoleARN, output.ExecutionRoleArn)
	d.Set("notification_type", output.NotificationType)
	d.Set(names.AttrStatus, output.Status)
	if err := d.Set("subscriber", flattenBudgetActionSubscriber(output.Subscribers)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting subscriber: %s", err)
	}

	return diags
}

func resourceBudgetActionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BudgetsClient(ctx)

	accountID, actionID, budgetName, err := BudgetActionParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &budgets.UpdateBudgetActionInput{
			AccountId:  aws.String(accountID),
			ActionId:   aws.String(actionID),
			BudgetName: aws.String(budgetName),
		}

		if d.HasChange("action_threshold") {
			input.ActionThreshold = expandBudgetActionActionThreshold(d.Get("action_threshold").([]interface{}))
		}

		if d.HasChange("approval_model") {
			input.ApprovalModel = awstypes.ApprovalModel(d.Get("approval_model").(string))
		}

		if d.HasChange("definition") {
			input.Definition = expandBudgetActionActionDefinition(d.Get("definition").([]interface{}))
		}

		if d.HasChange(names.AttrExecutionRoleARN) {
			input.ExecutionRoleArn = aws.String(d.Get(names.AttrExecutionRoleARN).(string))
		}

		if d.HasChange("notification_type") {
			input.NotificationType = awstypes.NotificationType(d.Get("notification_type").(string))
		}

		if d.HasChange("subscriber") {
			input.Subscribers = expandBudgetActionSubscriber(d.Get("subscriber").(*schema.Set))
		}

		_, err = conn.UpdateBudgetAction(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Budget Action (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceBudgetActionRead(ctx, d, meta)...)
}

func resourceBudgetActionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BudgetsClient(ctx)

	accountID, actionID, budgetName, err := BudgetActionParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Budget Action: %s", d.Id())
	_, err = tfresource.RetryWhenIsA[*awstypes.ResourceLockedException](ctx, d.Timeout(schema.TimeoutDelete), func() (interface{}, error) {
		return conn.DeleteBudgetAction(ctx, &budgets.DeleteBudgetActionInput{
			AccountId:  aws.String(accountID),
			ActionId:   aws.String(actionID),
			BudgetName: aws.String(budgetName),
		})
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Budget Action (%s): %s", d.Id(), err)
	}

	return diags
}

const budgetActionResourceIDSeparator = ":"

func BudgetActionCreateResourceID(accountID, actionID, budgetName string) string {
	parts := []string{accountID, actionID, budgetName}
	id := strings.Join(parts, budgetActionResourceIDSeparator)

	return id
}

func BudgetActionParseResourceID(id string) (string, string, string, error) {
	parts := strings.Split(id, budgetActionResourceIDSeparator)

	if len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != "" {
		return parts[0], parts[1], parts[2], nil
	}

	return "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected AccountID%[2]sActionID%[2]sBudgetName", id, budgetActionResourceIDSeparator)
}

const (
	propagationTimeout = 2 * time.Minute
)

func FindActionByThreePartKey(ctx context.Context, conn *budgets.Client, accountID, actionID, budgetName string) (*awstypes.Action, error) {
	input := &budgets.DescribeBudgetActionInput{
		AccountId:  aws.String(accountID),
		ActionId:   aws.String(actionID),
		BudgetName: aws.String(budgetName),
	}

	output, err := conn.DescribeBudgetAction(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Action == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Action, nil
}

func expandBudgetActionActionThreshold(l []interface{}) *awstypes.ActionThreshold {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &awstypes.ActionThreshold{}

	if v, ok := m["action_threshold_type"].(string); ok && v != "" {
		config.ActionThresholdType = awstypes.ThresholdType(v)
	}

	if v, ok := m["action_threshold_value"].(float64); ok {
		config.ActionThresholdValue = v
	}

	return config
}

func expandBudgetActionSubscriber(l *schema.Set) []awstypes.Subscriber {
	if l.Len() == 0 {
		return []awstypes.Subscriber{}
	}

	items := []awstypes.Subscriber{}

	for _, m := range l.List() {
		config := awstypes.Subscriber{}
		raw := m.(map[string]interface{})

		if v, ok := raw[names.AttrAddress].(string); ok && v != "" {
			config.Address = aws.String(v)
		}

		if v, ok := raw["subscription_type"].(string); ok {
			config.SubscriptionType = awstypes.SubscriptionType(v)
		}

		items = append(items, config)
	}

	return items
}

func expandBudgetActionActionDefinition(l []interface{}) *awstypes.Definition {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &awstypes.Definition{}

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

func expandBudgetActionActionScpActionDefinition(l []interface{}) *awstypes.ScpActionDefinition {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &awstypes.ScpActionDefinition{}

	if v, ok := m["policy_id"].(string); ok && v != "" {
		config.PolicyId = aws.String(v)
	}

	if v, ok := m["target_ids"].(*schema.Set); ok && v.Len() > 0 {
		config.TargetIds = flex.ExpandStringValueSet(v)
	}

	return config
}

func expandBudgetActionActionSSMActionDefinition(l []interface{}) *awstypes.SsmActionDefinition {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &awstypes.SsmActionDefinition{}

	if v, ok := m["action_sub_type"].(string); ok && v != "" {
		config.ActionSubType = awstypes.ActionSubType(v)
	}

	if v, ok := m[names.AttrRegion].(string); ok && v != "" {
		config.Region = aws.String(v)
	}

	if v, ok := m["instance_ids"].(*schema.Set); ok && v.Len() > 0 {
		config.InstanceIds = flex.ExpandStringValueSet(v)
	}

	return config
}

func expandBudgetActionActionIAMActionDefinition(l []interface{}) *awstypes.IamActionDefinition {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &awstypes.IamActionDefinition{}

	if v, ok := m["policy_arn"].(string); ok && v != "" {
		config.PolicyArn = aws.String(v)
	}

	if v, ok := m["groups"].(*schema.Set); ok && v.Len() > 0 {
		config.Groups = flex.ExpandStringValueSet(v)
	}

	if v, ok := m["roles"].(*schema.Set); ok && v.Len() > 0 {
		config.Roles = flex.ExpandStringValueSet(v)
	}

	if v, ok := m["users"].(*schema.Set); ok && v.Len() > 0 {
		config.Users = flex.ExpandStringValueSet(v)
	}

	return config
}

func flattenBudgetActionSubscriber(configured []awstypes.Subscriber) []map[string]interface{} {
	dataResources := make([]map[string]interface{}, 0, len(configured))

	for _, raw := range configured {
		item := make(map[string]interface{})
		item[names.AttrAddress] = aws.ToString(raw.Address)
		item["subscription_type"] = string(raw.SubscriptionType)

		dataResources = append(dataResources, item)
	}

	return dataResources
}

func flattenBudgetActionActionThreshold(lt *awstypes.ActionThreshold) []map[string]interface{} {
	if lt == nil {
		return []map[string]interface{}{}
	}

	attrs := map[string]interface{}{
		"action_threshold_type":  string(lt.ActionThresholdType),
		"action_threshold_value": lt.ActionThresholdValue,
	}

	return []map[string]interface{}{attrs}
}

func flattenBudgetActionIAMActionDefinition(lt *awstypes.IamActionDefinition) []map[string]interface{} {
	if lt == nil {
		return []map[string]interface{}{}
	}

	attrs := map[string]interface{}{
		"policy_arn": aws.ToString(lt.PolicyArn),
	}

	if lt.Users != nil && len(lt.Users) > 0 {
		attrs["users"] = flex.FlattenStringValueSet(lt.Users)
	}

	if lt.Roles != nil && len(lt.Roles) > 0 {
		attrs["roles"] = flex.FlattenStringValueSet(lt.Roles)
	}

	if lt.Groups != nil && len(lt.Groups) > 0 {
		attrs["groups"] = flex.FlattenStringValueSet(lt.Groups)
	}

	return []map[string]interface{}{attrs}
}

func flattenBudgetActionScpActionDefinition(lt *awstypes.ScpActionDefinition) []map[string]interface{} {
	if lt == nil {
		return []map[string]interface{}{}
	}

	attrs := map[string]interface{}{
		"policy_id": aws.ToString(lt.PolicyId),
	}

	if lt.TargetIds != nil && len(lt.TargetIds) > 0 {
		attrs["target_ids"] = flex.FlattenStringValueSet(lt.TargetIds)
	}

	return []map[string]interface{}{attrs}
}

func flattenBudgetActionSSMActionDefinition(lt *awstypes.SsmActionDefinition) []map[string]interface{} {
	if lt == nil {
		return []map[string]interface{}{}
	}

	attrs := map[string]interface{}{
		"action_sub_type": string(lt.ActionSubType),
		"instance_ids":    flex.FlattenStringValueSet(lt.InstanceIds),
		names.AttrRegion:  aws.ToString(lt.Region),
	}

	return []map[string]interface{}{attrs}
}

func flattenBudgetActionDefinition(lt *awstypes.Definition) []map[string]interface{} {
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
