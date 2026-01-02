// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package budgets

import (
	"cmp"
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/budgets"
	awstypes "github.com/aws/aws-sdk-go-v2/service/budgets/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_budgets_budget_action", name="Budget Action")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
func resourceBudgetAction() *schema.Resource {
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
	}
}

func resourceBudgetActionCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.BudgetsClient(ctx)

	accountID := cmp.Or(d.Get(names.AttrAccountID).(string), c.AccountID(ctx))
	budgetName := d.Get("budget_name").(string)
	input := budgets.CreateBudgetActionInput{
		AccountId:        aws.String(accountID),
		ActionThreshold:  expandActionThreshold(d.Get("action_threshold").([]any)),
		ActionType:       awstypes.ActionType(d.Get("action_type").(string)),
		ApprovalModel:    awstypes.ApprovalModel(d.Get("approval_model").(string)),
		BudgetName:       aws.String(budgetName),
		Definition:       expandDefinition(d.Get("definition").([]any)),
		ExecutionRoleArn: aws.String(d.Get(names.AttrExecutionRoleARN).(string)),
		NotificationType: awstypes.NotificationType(d.Get("notification_type").(string)),
		Subscribers:      expandBudgetActionSubscribers(d.Get("subscriber").(*schema.Set)),
		ResourceTags:     getTagsIn(ctx),
	}

	output, err := tfresource.RetryWhenIsA[*budgets.CreateBudgetActionOutput, *awstypes.AccessDeniedException](ctx, iamPropagationTimeout, func(ctx context.Context) (*budgets.CreateBudgetActionOutput, error) {
		return conn.CreateBudgetAction(ctx, &input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Budget Action: %s", err)
	}

	actionID := aws.ToString(output.ActionId)
	d.SetId(budgetActionCreateResourceID(accountID, actionID, budgetName))

	_, err = findWithDelay(ctx, func(context.Context) (*awstypes.Action, error) {
		return findBudgetActionByThreePartKey(ctx, conn, accountID, actionID, budgetName)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Budget Action (%s): %s", d.Id(), err)
	}

	return append(diags, resourceBudgetActionRead(ctx, d, meta)...)
}

func resourceBudgetActionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.BudgetsClient(ctx)

	accountID, actionID, budgetName, err := budgetActionParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := findBudgetActionByThreePartKey(ctx, conn, accountID, actionID, budgetName)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Budget Action (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Budget Action (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrAccountID, accountID)
	d.Set("action_id", actionID)
	if err := d.Set("action_threshold", flattenActionThreshold(output.ActionThreshold)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting action_threshold: %s", err)
	}
	d.Set("action_type", output.ActionType)
	d.Set("approval_model", output.ApprovalModel)
	d.Set(names.AttrARN, budgetActionARN(ctx, c, budgetName, actionID))
	d.Set("budget_name", budgetName)
	if err := d.Set("definition", flattenDefinition(output.Definition)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting definition: %s", err)
	}
	d.Set(names.AttrExecutionRoleARN, output.ExecutionRoleArn)
	d.Set("notification_type", output.NotificationType)
	d.Set(names.AttrStatus, output.Status)
	if err := d.Set("subscriber", flattenBudgetActionSubscribers(output.Subscribers)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting subscriber: %s", err)
	}

	return diags
}

func resourceBudgetActionUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BudgetsClient(ctx)

	accountID, actionID, budgetName, err := budgetActionParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := budgets.UpdateBudgetActionInput{
			AccountId:  aws.String(accountID),
			ActionId:   aws.String(actionID),
			BudgetName: aws.String(budgetName),
		}

		if d.HasChange("action_threshold") {
			input.ActionThreshold = expandActionThreshold(d.Get("action_threshold").([]any))
		}

		if d.HasChange("approval_model") {
			input.ApprovalModel = awstypes.ApprovalModel(d.Get("approval_model").(string))
		}

		if d.HasChange("definition") {
			input.Definition = expandDefinition(d.Get("definition").([]any))
		}

		if d.HasChange(names.AttrExecutionRoleARN) {
			input.ExecutionRoleArn = aws.String(d.Get(names.AttrExecutionRoleARN).(string))
		}

		if d.HasChange("notification_type") {
			input.NotificationType = awstypes.NotificationType(d.Get("notification_type").(string))
		}

		if d.HasChange("subscriber") {
			input.Subscribers = expandBudgetActionSubscribers(d.Get("subscriber").(*schema.Set))
		}

		_, err = conn.UpdateBudgetAction(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Budget Action (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceBudgetActionRead(ctx, d, meta)...)
}

func resourceBudgetActionDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BudgetsClient(ctx)

	accountID, actionID, budgetName, err := budgetActionParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Budget Action: %s", d.Id())
	input := budgets.DeleteBudgetActionInput{
		AccountId:  aws.String(accountID),
		ActionId:   aws.String(actionID),
		BudgetName: aws.String(budgetName),
	}
	_, err = tfresource.RetryWhenIsA[any, *awstypes.ResourceLockedException](ctx, d.Timeout(schema.TimeoutDelete), func(ctx context.Context) (any, error) {
		return conn.DeleteBudgetAction(ctx, &input)
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

func budgetActionCreateResourceID(accountID, actionID, budgetName string) string {
	parts := []string{accountID, actionID, budgetName}
	id := strings.Join(parts, budgetActionResourceIDSeparator)

	return id
}

func budgetActionParseResourceID(id string) (string, string, string, error) {
	parts := strings.Split(id, budgetActionResourceIDSeparator)

	if len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != "" {
		return parts[0], parts[1], parts[2], nil
	}

	return "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected AccountID%[2]sActionID%[2]sBudgetName", id, budgetActionResourceIDSeparator)
}

func findBudgetAction(ctx context.Context, conn *budgets.Client, input *budgets.DescribeBudgetActionInput) (*awstypes.Action, error) {
	output, err := conn.DescribeBudgetAction(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
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

func findBudgetActionByThreePartKey(ctx context.Context, conn *budgets.Client, accountID, actionID, budgetName string) (*awstypes.Action, error) {
	input := budgets.DescribeBudgetActionInput{
		AccountId:  aws.String(accountID),
		ActionId:   aws.String(actionID),
		BudgetName: aws.String(budgetName),
	}

	return findBudgetAction(ctx, conn, &input)
}

func expandActionThreshold(tfList []any) *awstypes.ActionThreshold {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.ActionThreshold{}

	if v, ok := tfMap["action_threshold_type"].(string); ok && v != "" {
		apiObject.ActionThresholdType = awstypes.ThresholdType(v)
	}

	if v, ok := tfMap["action_threshold_value"].(float64); ok {
		apiObject.ActionThresholdValue = v
	}

	return apiObject
}

func expandBudgetActionSubscribers(tfSet *schema.Set) []awstypes.Subscriber {
	if tfSet.Len() == 0 {
		return []awstypes.Subscriber{}
	}

	apiObjects := []awstypes.Subscriber{}

	for _, m := range tfSet.List() {
		apiObject := awstypes.Subscriber{}
		tfMap := m.(map[string]any)

		if v, ok := tfMap[names.AttrAddress].(string); ok && v != "" {
			apiObject.Address = aws.String(v)
		}

		if v, ok := tfMap["subscription_type"].(string); ok {
			apiObject.SubscriptionType = awstypes.SubscriptionType(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandDefinition(tfList []any) *awstypes.Definition {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.Definition{}

	if v, ok := tfMap["iam_action_definition"].([]any); ok && len(v) > 0 {
		apiObject.IamActionDefinition = expandIAMActionDefinition(v)
	}

	if v, ok := tfMap["scp_action_definition"].([]any); ok && len(v) > 0 {
		apiObject.ScpActionDefinition = expandSCPActionDefinition(v)
	}

	if v, ok := tfMap["ssm_action_definition"].([]any); ok && len(v) > 0 {
		apiObject.SsmActionDefinition = expandSSMActionDefinition(v)
	}

	return apiObject
}

func expandSCPActionDefinition(tfList []any) *awstypes.ScpActionDefinition {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.ScpActionDefinition{}

	if v, ok := tfMap["policy_id"].(string); ok && v != "" {
		apiObject.PolicyId = aws.String(v)
	}

	if v, ok := tfMap["target_ids"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.TargetIds = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func expandSSMActionDefinition(tfList []any) *awstypes.SsmActionDefinition {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.SsmActionDefinition{}

	if v, ok := tfMap["action_sub_type"].(string); ok && v != "" {
		apiObject.ActionSubType = awstypes.ActionSubType(v)
	}

	if v, ok := tfMap["instance_ids"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.InstanceIds = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap[names.AttrRegion].(string); ok && v != "" {
		apiObject.Region = aws.String(v)
	}

	return apiObject
}

func expandIAMActionDefinition(tfList []any) *awstypes.IamActionDefinition {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.IamActionDefinition{}

	if v, ok := tfMap["groups"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Groups = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["policy_arn"].(string); ok && v != "" {
		apiObject.PolicyArn = aws.String(v)
	}

	if v, ok := tfMap["roles"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Roles = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["users"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Users = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func flattenBudgetActionSubscribers(apiObjects []awstypes.Subscriber) []any {
	tfList := make([]any, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfMap := make(map[string]any)
		tfMap[names.AttrAddress] = aws.ToString(apiObject.Address)
		tfMap["subscription_type"] = apiObject.SubscriptionType

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenActionThreshold(apiObject *awstypes.ActionThreshold) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"action_threshold_type":  apiObject.ActionThresholdType,
		"action_threshold_value": apiObject.ActionThresholdValue,
	}

	return []any{tfMap}
}

func flattenIAMActionDefinition(apiObject *awstypes.IamActionDefinition) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"groups":     apiObject.Groups,
		"policy_arn": aws.ToString(apiObject.PolicyArn),
		"roles":      apiObject.Roles,
		"users":      apiObject.Users,
	}

	return []any{tfMap}
}

func flattenSCPActionDefinition(apiObject *awstypes.ScpActionDefinition) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"policy_id":  aws.ToString(apiObject.PolicyId),
		"target_ids": apiObject.TargetIds,
	}

	return []any{tfMap}
}

func flattenSSMActionDefinition(apiObject *awstypes.SsmActionDefinition) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"action_sub_type": apiObject.ActionSubType,
		"instance_ids":    apiObject.InstanceIds,
		names.AttrRegion:  aws.ToString(apiObject.Region),
	}

	return []any{tfMap}
}

func flattenDefinition(apiObject *awstypes.Definition) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{}

	if apiObject.IamActionDefinition != nil {
		tfMap["iam_action_definition"] = flattenIAMActionDefinition(apiObject.IamActionDefinition)
	}

	if apiObject.ScpActionDefinition != nil {
		tfMap["scp_action_definition"] = flattenSCPActionDefinition(apiObject.ScpActionDefinition)
	}

	if apiObject.SsmActionDefinition != nil {
		tfMap["ssm_action_definition"] = flattenSSMActionDefinition(apiObject.SsmActionDefinition)
	}

	return []any{tfMap}
}

func budgetActionARN(ctx context.Context, c *conns.AWSClient, budgetName, actionID string) string {
	return c.GlobalARN(ctx, "budgets", fmt.Sprintf("budget/%s/action/%s", budgetName, actionID))
}
