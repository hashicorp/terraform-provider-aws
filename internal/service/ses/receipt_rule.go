// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
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

// @SDKResource("aws_ses_receipt_rule", name="Receipt Rule")
func resourceReceiptRule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceReceiptRuleCreate,
		UpdateWithoutTimeout: resourceReceiptRuleUpdate,
		ReadWithoutTimeout:   resourceReceiptRuleRead,
		DeleteWithoutTimeout: resourceReceiptRuleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceReceiptRuleImport,
		},

		Schema: map[string]*schema.Schema{
			"add_header_action": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"header_name": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 50),
								validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z-]+$`), "must contain only alphanumeric and dash characters"),
							),
						},
						"header_value": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 2048),
						},
						"position": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
			"after": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bounce_action": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrMessage: {
							Type:     schema.TypeString,
							Required: true,
						},
						"position": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"sender": {
							Type:     schema.TypeString,
							Required: true,
						},
						"smtp_reply_code": {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrStatusCode: {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrTopicARN: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			names.AttrEnabled: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"lambda_action": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrFunctionARN: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						"invocation_type": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.InvocationTypeEvent,
							ValidateDiagFunc: enum.Validate[awstypes.InvocationType](),
						},
						"position": {
							Type:     schema.TypeInt,
							Required: true,
						},
						names.AttrTopicARN: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+$`), "must contain only alphanumeric, period, underscore, and hyphen characters"),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z]`), "must begin with a alphanumeric character"),
					validation.StringMatch(regexache.MustCompile(`[0-9A-Za-z]$`), "must end with a alphanumeric character"),
				),
			},
			"recipients": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"rule_set_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"s3_action": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrBucketName: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrIAMRoleARN: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
						names.AttrKMSKeyARN: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
						"object_key_prefix": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"position": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},
						names.AttrTopicARN: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"scan_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"sns_action": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"encoding": {
							Type:             schema.TypeString,
							Default:          awstypes.SNSActionEncodingUtf8,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.SNSActionEncoding](),
						},
						"position": {
							Type:     schema.TypeInt,
							Required: true,
						},
						names.AttrTopicARN: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"stop_action": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrScope: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.StopScope](),
						},
						"position": {
							Type:     schema.TypeInt,
							Required: true,
						},
						names.AttrTopicARN: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"tls_policy": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[awstypes.TlsPolicy](),
			},
			"workmail_action": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"organization_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						"position": {
							Type:     schema.TypeInt,
							Required: true,
						},
						names.AttrTopicARN: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
		},
	}
}

func resourceReceiptRuleCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &ses.CreateReceiptRuleInput{
		Rule:        expandReceiptRule(d),
		RuleSetName: aws.String(d.Get("rule_set_name").(string)),
	}

	if v, ok := d.GetOk("after"); ok {
		input.After = aws.String(v.(string))
	}

	_, err := tfresource.RetryWhen(ctx, d.Timeout(schema.TimeoutCreate),
		func() (any, error) {
			return conn.CreateReceiptRule(ctx, input)
		},
		func(err error) (bool, error) {
			if tfawserr.ErrMessageContains(err, errCodeInvalidLambdaConfiguration, "Could not invoke Lambda function") ||
				tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "Could not assume the provided IAM Role") ||
				tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "Unable to write to S3 bucket") ||
				tfawserr.ErrMessageContains(err, errCodeInvalidS3Configuration, "Could not write to bucket") {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SES Receipt Rule (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceReceiptRuleRead(ctx, d, meta)...)
}

func resourceReceiptRuleRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	ruleSetName := d.Get("rule_set_name").(string)
	rule, err := findReceiptRuleByTwoPartKey(ctx, conn, d.Id(), ruleSetName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SES Receipt Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SES Receipt Rule (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition(ctx),
		Service:   "ses",
		Region:    meta.(*conns.AWSClient).Region(ctx),
		AccountID: meta.(*conns.AWSClient).AccountID(ctx),
		Resource:  fmt.Sprintf("receipt-rule-set/%s:receipt-rule/%s", ruleSetName, d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrEnabled, rule.Enabled)
	d.Set("recipients", rule.Recipients)
	d.Set("scan_enabled", rule.ScanEnabled)
	d.Set("tls_policy", rule.TlsPolicy)

	addHeaderActionList := []map[string]any{}
	bounceActionList := []map[string]any{}
	lambdaActionList := []map[string]any{}
	s3ActionList := []map[string]any{}
	snsActionList := []map[string]any{}
	stopActionList := []map[string]any{}
	workmailActionList := []map[string]any{}

	for i, apiObject := range rule.Actions {
		if apiObject := apiObject.AddHeaderAction; apiObject != nil {
			tfMap := map[string]any{
				"header_name":  aws.ToString(apiObject.HeaderName),
				"header_value": aws.ToString(apiObject.HeaderValue),
				"position":     i + 1,
			}
			addHeaderActionList = append(addHeaderActionList, tfMap)
		}

		if apiObject := apiObject.BounceAction; apiObject != nil {
			tfMap := map[string]any{
				names.AttrMessage: aws.ToString(apiObject.Message),
				"sender":          aws.ToString(apiObject.Sender),
				"smtp_reply_code": aws.ToString(apiObject.SmtpReplyCode),
				"position":        i + 1,
			}

			if v := apiObject.StatusCode; v != nil {
				tfMap[names.AttrStatusCode] = aws.ToString(v)
			}

			if v := apiObject.TopicArn; v != nil {
				tfMap[names.AttrTopicARN] = aws.ToString(v)
			}

			bounceActionList = append(bounceActionList, tfMap)
		}

		if apiObject := apiObject.LambdaAction; apiObject != nil {
			tfMap := map[string]any{
				names.AttrFunctionARN: aws.ToString(apiObject.FunctionArn),
				"invocation_type":     apiObject.InvocationType,
				"position":            i + 1,
			}

			if v := apiObject.TopicArn; v != nil {
				tfMap[names.AttrTopicARN] = aws.ToString(v)
			}

			lambdaActionList = append(lambdaActionList, tfMap)
		}

		if apiObject := apiObject.S3Action; apiObject != nil {
			tfMap := map[string]any{
				names.AttrBucketName: aws.ToString(apiObject.BucketName),
				"position":           i + 1,
			}

			if v := apiObject.IamRoleArn; v != nil {
				tfMap[names.AttrIAMRoleARN] = aws.ToString(v)
			}

			if v := apiObject.KmsKeyArn; v != nil {
				tfMap[names.AttrKMSKeyARN] = aws.ToString(v)
			}

			if v := apiObject.ObjectKeyPrefix; v != nil {
				tfMap["object_key_prefix"] = aws.ToString(v)
			}

			if v := apiObject.TopicArn; v != nil {
				tfMap[names.AttrTopicARN] = aws.ToString(v)
			}

			s3ActionList = append(s3ActionList, tfMap)
		}

		if apiObject := apiObject.SNSAction; apiObject != nil {
			tfMap := map[string]any{
				names.AttrTopicARN: aws.ToString(apiObject.TopicArn),
				"encoding":         apiObject.Encoding,
				"position":         i + 1,
			}

			snsActionList = append(snsActionList, tfMap)
		}

		if apiObject := apiObject.StopAction; apiObject != nil {
			stopAction := map[string]any{
				names.AttrScope: apiObject.Scope,
				"position":      i + 1,
			}

			if v := apiObject.TopicArn; v != nil {
				stopAction[names.AttrTopicARN] = aws.ToString(v)
			}

			stopActionList = append(stopActionList, stopAction)
		}

		if apiObject := apiObject.WorkmailAction; apiObject != nil {
			workmailAction := map[string]any{
				"organization_arn": aws.ToString(apiObject.OrganizationArn),
				"position":         i + 1,
			}

			if v := apiObject.TopicArn; v != nil {
				workmailAction[names.AttrTopicARN] = aws.ToString(v)
			}

			workmailActionList = append(workmailActionList, workmailAction)
		}
	}

	if err := d.Set("add_header_action", addHeaderActionList); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting add_header_action: %s", err)
	}
	if err := d.Set("bounce_action", bounceActionList); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting bounce_action: %s", err)
	}
	if err := d.Set("lambda_action", lambdaActionList); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting lambda_action: %s", err)
	}
	if err := d.Set("s3_action", s3ActionList); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting s3_action: %s", err)
	}
	if err := d.Set("sns_action", snsActionList); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting sns_action: %s", err)
	}
	if err := d.Set("stop_action", stopActionList); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting stop_action: %s", err)
	}
	if err := d.Set("workmail_action", workmailActionList); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting workmail_action: %s", err)
	}

	return diags
}

func resourceReceiptRuleUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	input := &ses.UpdateReceiptRuleInput{
		Rule:        expandReceiptRule(d),
		RuleSetName: aws.String(d.Get("rule_set_name").(string)),
	}

	_, err := tfresource.RetryWhen(ctx, d.Timeout(schema.TimeoutUpdate),
		func() (any, error) {
			return conn.UpdateReceiptRule(ctx, input)
		},
		func(err error) (bool, error) {
			if tfawserr.ErrMessageContains(err, errCodeInvalidLambdaConfiguration, "Could not invoke Lambda function") ||
				tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "Could not assume the provided IAM Role") ||
				tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "Unable to write to S3 bucket") ||
				tfawserr.ErrMessageContains(err, errCodeInvalidS3Configuration, "Could not write to bucket") {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating SES Receipt Rule (%s): %s", d.Id(), err)
	}

	if d.HasChange("after") {
		input := &ses.SetReceiptRulePositionInput{
			After:       aws.String(d.Get("after").(string)),
			RuleName:    aws.String(d.Get(names.AttrName).(string)),
			RuleSetName: aws.String(d.Get("rule_set_name").(string)),
		}

		_, err := conn.SetReceiptRulePosition(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting SES Receipt Rule (%s) position: %s", d.Id(), err)
		}
	}

	return append(diags, resourceReceiptRuleRead(ctx, d, meta)...)
}

func resourceReceiptRuleDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	log.Printf("[DEBUG] Deleting SES Receipt Rule: %s", d.Id())
	_, err := conn.DeleteReceiptRule(ctx, &ses.DeleteReceiptRuleInput{
		RuleName:    aws.String(d.Id()),
		RuleSetName: aws.String(d.Get("rule_set_name").(string)),
	})

	if errs.IsA[*awstypes.RuleSetDoesNotExistException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SES Receipt Rule (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceReceiptRuleImport(_ context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	idParts := strings.Split(d.Id(), ":")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("unexpected format of ID (%q), expected <ruleset-name>:<rule-name>", d.Id())
	}

	ruleSetName := idParts[0]
	ruleName := idParts[1]

	d.Set("rule_set_name", ruleSetName)
	d.Set(names.AttrName, ruleName)
	d.SetId(ruleName)

	return []*schema.ResourceData{d}, nil
}

func findReceiptRuleByTwoPartKey(ctx context.Context, conn *ses.Client, ruleName, ruleSetName string) (*awstypes.ReceiptRule, error) {
	input := &ses.DescribeReceiptRuleInput{
		RuleName:    aws.String(ruleName),
		RuleSetName: aws.String(ruleSetName),
	}

	return findReceiptRule(ctx, conn, input)
}

func findReceiptRule(ctx context.Context, conn *ses.Client, input *ses.DescribeReceiptRuleInput) (*awstypes.ReceiptRule, error) {
	output, err := conn.DescribeReceiptRule(ctx, input)

	if errs.IsA[*awstypes.RuleDoesNotExistException](err) || errs.IsA[*awstypes.RuleSetDoesNotExistException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Rule == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Rule, nil
}

func expandReceiptRule(d *schema.ResourceData) *awstypes.ReceiptRule {
	apiObject := &awstypes.ReceiptRule{
		Name: aws.String(d.Get(names.AttrName).(string)),
	}

	if v, ok := d.GetOk(names.AttrEnabled); ok {
		apiObject.Enabled = v.(bool)
	}

	if v, ok := d.GetOk("recipients"); ok {
		apiObject.Recipients = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("scan_enabled"); ok {
		apiObject.ScanEnabled = v.(bool)
	}

	if v, ok := d.GetOk("tls_policy"); ok {
		apiObject.TlsPolicy = awstypes.TlsPolicy(v.(string))
	}

	actions := make(map[int]awstypes.ReceiptAction)

	if v, ok := d.GetOk("add_header_action"); ok {
		for _, element := range v.(*schema.Set).List() {
			elem := element.(map[string]any)

			actions[elem["position"].(int)] = awstypes.ReceiptAction{
				AddHeaderAction: &awstypes.AddHeaderAction{
					HeaderName:  aws.String(elem["header_name"].(string)),
					HeaderValue: aws.String(elem["header_value"].(string)),
				},
			}
		}
	}

	if v, ok := d.GetOk("bounce_action"); ok {
		for _, element := range v.(*schema.Set).List() {
			elem := element.(map[string]any)

			bounceAction := &awstypes.BounceAction{
				Message:       aws.String(elem[names.AttrMessage].(string)),
				Sender:        aws.String(elem["sender"].(string)),
				SmtpReplyCode: aws.String(elem["smtp_reply_code"].(string)),
			}

			if elem[names.AttrStatusCode] != "" {
				bounceAction.StatusCode = aws.String(elem[names.AttrStatusCode].(string))
			}

			if elem[names.AttrTopicARN] != "" {
				bounceAction.TopicArn = aws.String(elem[names.AttrTopicARN].(string))
			}

			actions[elem["position"].(int)] = awstypes.ReceiptAction{
				BounceAction: bounceAction,
			}
		}
	}

	if v, ok := d.GetOk("lambda_action"); ok {
		for _, element := range v.(*schema.Set).List() {
			elem := element.(map[string]any)

			lambdaAction := &awstypes.LambdaAction{
				FunctionArn: aws.String(elem[names.AttrFunctionARN].(string)),
			}

			if elem["invocation_type"] != "" {
				lambdaAction.InvocationType = awstypes.InvocationType(elem["invocation_type"].(string))
			}

			if elem[names.AttrTopicARN] != "" {
				lambdaAction.TopicArn = aws.String(elem[names.AttrTopicARN].(string))
			}

			actions[elem["position"].(int)] = awstypes.ReceiptAction{
				LambdaAction: lambdaAction,
			}
		}
	}

	if v, ok := d.GetOk("s3_action"); ok {
		for _, element := range v.(*schema.Set).List() {
			elem := element.(map[string]any)

			s3Action := &awstypes.S3Action{
				BucketName: aws.String(elem[names.AttrBucketName].(string)),
			}

			if elem[names.AttrIAMRoleARN] != "" {
				s3Action.IamRoleArn = aws.String(elem[names.AttrIAMRoleARN].(string))
			}

			if elem[names.AttrKMSKeyARN] != "" {
				s3Action.KmsKeyArn = aws.String(elem[names.AttrKMSKeyARN].(string))
			}

			if elem["object_key_prefix"] != "" {
				s3Action.ObjectKeyPrefix = aws.String(elem["object_key_prefix"].(string))
			}

			if elem[names.AttrTopicARN] != "" {
				s3Action.TopicArn = aws.String(elem[names.AttrTopicARN].(string))
			}

			actions[elem["position"].(int)] = awstypes.ReceiptAction{
				S3Action: s3Action,
			}
		}
	}

	if v, ok := d.GetOk("sns_action"); ok {
		for _, element := range v.(*schema.Set).List() {
			elem := element.(map[string]any)

			snsAction := &awstypes.SNSAction{
				TopicArn: aws.String(elem[names.AttrTopicARN].(string)),
				Encoding: awstypes.SNSActionEncoding(elem["encoding"].(string)),
			}

			actions[elem["position"].(int)] = awstypes.ReceiptAction{
				SNSAction: snsAction,
			}
		}
	}

	if v, ok := d.GetOk("stop_action"); ok {
		for _, element := range v.(*schema.Set).List() {
			elem := element.(map[string]any)

			stopAction := &awstypes.StopAction{
				Scope: awstypes.StopScope(elem[names.AttrScope].(string)),
			}

			if elem[names.AttrTopicARN] != "" {
				stopAction.TopicArn = aws.String(elem[names.AttrTopicARN].(string))
			}

			actions[elem["position"].(int)] = awstypes.ReceiptAction{
				StopAction: stopAction,
			}
		}
	}

	if v, ok := d.GetOk("workmail_action"); ok {
		for _, element := range v.(*schema.Set).List() {
			elem := element.(map[string]any)

			workmailAction := &awstypes.WorkmailAction{
				OrganizationArn: aws.String(elem["organization_arn"].(string)),
			}

			if elem[names.AttrTopicARN] != "" {
				workmailAction.TopicArn = aws.String(elem[names.AttrTopicARN].(string))
			}

			actions[elem["position"].(int)] = awstypes.ReceiptAction{
				WorkmailAction: workmailAction,
			}
		}
	}

	var keys []int
	for k := range actions {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	sortedActions := []awstypes.ReceiptAction{}
	for _, k := range keys {
		sortedActions = append(sortedActions, actions[k])
	}

	apiObject.Actions = sortedActions

	return apiObject
}
