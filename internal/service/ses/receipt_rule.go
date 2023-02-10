package ses

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceReceiptRule() *schema.Resource {
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
								validation.StringMatch(regexp.MustCompile(`^[0-9a-zA-Z-]+$`), "must contain only alphanumeric and dash characters"),
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bounce_action": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"message": {
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
						"status_code": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"topic_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"lambda_action": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"function_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						"invocation_type": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      ses.InvocationTypeEvent,
							ValidateFunc: validation.StringInSlice(ses.InvocationType_Values(), false),
						},
						"position": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"topic_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[0-9a-zA-Z._-]+$`), "must contain only alphanumeric, period, underscore, and hyphen characters"),
					validation.StringMatch(regexp.MustCompile(`^[0-9a-zA-Z]`), "must begin with a alphanumeric character"),
					validation.StringMatch(regexp.MustCompile(`[0-9a-zA-Z]$`), "must end with a alphanumeric character"),
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
						"bucket_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"kms_key_arn": {
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
						"topic_arn": {
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
							Type:         schema.TypeString,
							Default:      ses.SNSActionEncodingUtf8,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(ses.SNSActionEncoding_Values(), false),
						},
						"position": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"topic_arn": {
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
						"scope": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(ses.StopScope_Values(), false),
						},
						"position": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"topic_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"tls_policy": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(ses.TlsPolicy_Values(), false),
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
						"topic_arn": {
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

func resourceReceiptRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESConn()

	name := d.Get("name").(string)
	input := &ses.CreateReceiptRuleInput{
		Rule:        buildReceiptRule(d),
		RuleSetName: aws.String(d.Get("rule_set_name").(string)),
	}

	if v, ok := d.GetOk("after"); ok {
		input.After = aws.String(v.(string))
	}

	_, err := conn.CreateReceiptRuleWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SES Receipt Rule (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceReceiptRuleRead(ctx, d, meta)...)
}

func resourceReceiptRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESConn()

	ruleSetName := d.Get("rule_set_name").(string)
	rule, err := FindReceiptRuleByTwoPartKey(ctx, conn, d.Id(), ruleSetName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SES Receipt Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SES Receipt Rule (%s): %s", d.Id(), err)
	}

	d.Set("enabled", rule.Enabled)
	d.Set("recipients", flex.FlattenStringSet(rule.Recipients))
	d.Set("scan_enabled", rule.ScanEnabled)
	d.Set("tls_policy", rule.TlsPolicy)

	addHeaderActionList := []map[string]interface{}{}
	bounceActionList := []map[string]interface{}{}
	lambdaActionList := []map[string]interface{}{}
	s3ActionList := []map[string]interface{}{}
	snsActionList := []map[string]interface{}{}
	stopActionList := []map[string]interface{}{}
	workmailActionList := []map[string]interface{}{}

	for i, element := range rule.Actions {
		if element.AddHeaderAction != nil {
			addHeaderAction := map[string]interface{}{
				"header_name":  aws.StringValue(element.AddHeaderAction.HeaderName),
				"header_value": aws.StringValue(element.AddHeaderAction.HeaderValue),
				"position":     i + 1,
			}
			addHeaderActionList = append(addHeaderActionList, addHeaderAction)
		}

		if element.BounceAction != nil {
			bounceAction := map[string]interface{}{
				"message":         aws.StringValue(element.BounceAction.Message),
				"sender":          aws.StringValue(element.BounceAction.Sender),
				"smtp_reply_code": aws.StringValue(element.BounceAction.SmtpReplyCode),
				"position":        i + 1,
			}

			if element.BounceAction.StatusCode != nil {
				bounceAction["status_code"] = aws.StringValue(element.BounceAction.StatusCode)
			}

			if element.BounceAction.TopicArn != nil {
				bounceAction["topic_arn"] = aws.StringValue(element.BounceAction.TopicArn)
			}

			bounceActionList = append(bounceActionList, bounceAction)
		}

		if element.LambdaAction != nil {
			lambdaAction := map[string]interface{}{
				"function_arn": aws.StringValue(element.LambdaAction.FunctionArn),
				"position":     i + 1,
			}

			if element.LambdaAction.InvocationType != nil {
				lambdaAction["invocation_type"] = aws.StringValue(element.LambdaAction.InvocationType)
			}

			if element.LambdaAction.TopicArn != nil {
				lambdaAction["topic_arn"] = aws.StringValue(element.LambdaAction.TopicArn)
			}

			lambdaActionList = append(lambdaActionList, lambdaAction)
		}

		if element.S3Action != nil {
			s3Action := map[string]interface{}{
				"bucket_name": aws.StringValue(element.S3Action.BucketName),
				"position":    i + 1,
			}

			if element.S3Action.KmsKeyArn != nil {
				s3Action["kms_key_arn"] = aws.StringValue(element.S3Action.KmsKeyArn)
			}

			if element.S3Action.ObjectKeyPrefix != nil {
				s3Action["object_key_prefix"] = aws.StringValue(element.S3Action.ObjectKeyPrefix)
			}

			if element.S3Action.TopicArn != nil {
				s3Action["topic_arn"] = aws.StringValue(element.S3Action.TopicArn)
			}

			s3ActionList = append(s3ActionList, s3Action)
		}

		if element.SNSAction != nil {
			snsAction := map[string]interface{}{
				"topic_arn": aws.StringValue(element.SNSAction.TopicArn),
				"encoding":  aws.StringValue(element.SNSAction.Encoding),
				"position":  i + 1,
			}

			snsActionList = append(snsActionList, snsAction)
		}

		if element.StopAction != nil {
			stopAction := map[string]interface{}{
				"scope":    aws.StringValue(element.StopAction.Scope),
				"position": i + 1,
			}

			if element.StopAction.TopicArn != nil {
				stopAction["topic_arn"] = aws.StringValue(element.StopAction.TopicArn)
			}

			stopActionList = append(stopActionList, stopAction)
		}

		if element.WorkmailAction != nil {
			workmailAction := map[string]interface{}{
				"organization_arn": aws.StringValue(element.WorkmailAction.OrganizationArn),
				"position":         i + 1,
			}

			if element.WorkmailAction.TopicArn != nil {
				workmailAction["topic_arn"] = aws.StringValue(element.WorkmailAction.TopicArn)
			}

			workmailActionList = append(workmailActionList, workmailAction)
		}
	}

	err = d.Set("add_header_action", addHeaderActionList)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting add_header_action: %s", err)
	}

	err = d.Set("bounce_action", bounceActionList)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting bounce_action: %s", err)
	}

	err = d.Set("lambda_action", lambdaActionList)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting lambda_action: %s", err)
	}

	err = d.Set("s3_action", s3ActionList)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting s3_action: %s", err)
	}

	err = d.Set("sns_action", snsActionList)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting sns_action: %s", err)
	}

	err = d.Set("stop_action", stopActionList)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting stop_action: %s", err)
	}

	err = d.Set("workmail_action", workmailActionList)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting workmail_action: %s", err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "ses",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("receipt-rule-set/%s:receipt-rule/%s", ruleSetName, d.Id()),
	}.String()
	d.Set("arn", arn)

	return diags
}

func resourceReceiptRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESConn()

	input := &ses.UpdateReceiptRuleInput{
		Rule:        buildReceiptRule(d),
		RuleSetName: aws.String(d.Get("rule_set_name").(string)),
	}

	_, err := conn.UpdateReceiptRuleWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating SES Receipt Rule (%s): %s", d.Id(), err)
	}

	if d.HasChange("after") {
		input := &ses.SetReceiptRulePositionInput{
			After:       aws.String(d.Get("after").(string)),
			RuleName:    aws.String(d.Get("name").(string)),
			RuleSetName: aws.String(d.Get("rule_set_name").(string)),
		}

		_, err := conn.SetReceiptRulePositionWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting SES Receipt Rule (%s) position: %s", d.Id(), err)
		}
	}

	return append(diags, resourceReceiptRuleRead(ctx, d, meta)...)
}

func resourceReceiptRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESConn()

	log.Printf("[DEBUG] Deleting SES Receipt Rule: %s", d.Id())
	_, err := conn.DeleteReceiptRuleWithContext(ctx, &ses.DeleteReceiptRuleInput{
		RuleName:    aws.String(d.Id()),
		RuleSetName: aws.String(d.Get("rule_set_name").(string)),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SES Receipt Rule (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceReceiptRuleImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.Split(d.Id(), ":")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("unexpected format of ID (%q), expected <ruleset-name>:<rule-name>", d.Id())
	}

	ruleSetName := idParts[0]
	ruleName := idParts[1]

	d.Set("rule_set_name", ruleSetName)
	d.Set("name", ruleName)
	d.SetId(ruleName)

	return []*schema.ResourceData{d}, nil
}

func FindReceiptRuleByTwoPartKey(ctx context.Context, conn *ses.SES, ruleName, ruleSetName string) (*ses.ReceiptRule, error) {
	input := &ses.DescribeReceiptRuleInput{
		RuleName:    aws.String(ruleName),
		RuleSetName: aws.String(ruleSetName),
	}

	output, err := conn.DescribeReceiptRuleWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, ses.ErrCodeRuleDoesNotExistException, ses.ErrCodeRuleSetDoesNotExistException) {
		return nil, &resource.NotFoundError{
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

func buildReceiptRule(d *schema.ResourceData) *ses.ReceiptRule {
	receiptRule := &ses.ReceiptRule{
		Name: aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("enabled"); ok {
		receiptRule.Enabled = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("recipients"); ok {
		receiptRule.Recipients = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("scan_enabled"); ok {
		receiptRule.ScanEnabled = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("tls_policy"); ok {
		receiptRule.TlsPolicy = aws.String(v.(string))
	}

	actions := make(map[int]*ses.ReceiptAction)

	if v, ok := d.GetOk("add_header_action"); ok {
		for _, element := range v.(*schema.Set).List() {
			elem := element.(map[string]interface{})

			actions[elem["position"].(int)] = &ses.ReceiptAction{
				AddHeaderAction: &ses.AddHeaderAction{
					HeaderName:  aws.String(elem["header_name"].(string)),
					HeaderValue: aws.String(elem["header_value"].(string)),
				},
			}
		}
	}

	if v, ok := d.GetOk("bounce_action"); ok {
		for _, element := range v.(*schema.Set).List() {
			elem := element.(map[string]interface{})

			bounceAction := &ses.BounceAction{
				Message:       aws.String(elem["message"].(string)),
				Sender:        aws.String(elem["sender"].(string)),
				SmtpReplyCode: aws.String(elem["smtp_reply_code"].(string)),
			}

			if elem["status_code"] != "" {
				bounceAction.StatusCode = aws.String(elem["status_code"].(string))
			}

			if elem["topic_arn"] != "" {
				bounceAction.TopicArn = aws.String(elem["topic_arn"].(string))
			}

			actions[elem["position"].(int)] = &ses.ReceiptAction{
				BounceAction: bounceAction,
			}
		}
	}

	if v, ok := d.GetOk("lambda_action"); ok {
		for _, element := range v.(*schema.Set).List() {
			elem := element.(map[string]interface{})

			lambdaAction := &ses.LambdaAction{
				FunctionArn: aws.String(elem["function_arn"].(string)),
			}

			if elem["invocation_type"] != "" {
				lambdaAction.InvocationType = aws.String(elem["invocation_type"].(string))
			}

			if elem["topic_arn"] != "" {
				lambdaAction.TopicArn = aws.String(elem["topic_arn"].(string))
			}

			actions[elem["position"].(int)] = &ses.ReceiptAction{
				LambdaAction: lambdaAction,
			}
		}
	}

	if v, ok := d.GetOk("s3_action"); ok {
		for _, element := range v.(*schema.Set).List() {
			elem := element.(map[string]interface{})

			s3Action := &ses.S3Action{
				BucketName: aws.String(elem["bucket_name"].(string)),
			}

			if elem["kms_key_arn"] != "" {
				s3Action.KmsKeyArn = aws.String(elem["kms_key_arn"].(string))
			}

			if elem["object_key_prefix"] != "" {
				s3Action.ObjectKeyPrefix = aws.String(elem["object_key_prefix"].(string))
			}

			if elem["topic_arn"] != "" {
				s3Action.TopicArn = aws.String(elem["topic_arn"].(string))
			}

			actions[elem["position"].(int)] = &ses.ReceiptAction{
				S3Action: s3Action,
			}
		}
	}

	if v, ok := d.GetOk("sns_action"); ok {
		for _, element := range v.(*schema.Set).List() {
			elem := element.(map[string]interface{})

			snsAction := &ses.SNSAction{
				TopicArn: aws.String(elem["topic_arn"].(string)),
				Encoding: aws.String(elem["encoding"].(string)),
			}

			actions[elem["position"].(int)] = &ses.ReceiptAction{
				SNSAction: snsAction,
			}
		}
	}

	if v, ok := d.GetOk("stop_action"); ok {
		for _, element := range v.(*schema.Set).List() {
			elem := element.(map[string]interface{})

			stopAction := &ses.StopAction{
				Scope: aws.String(elem["scope"].(string)),
			}

			if elem["topic_arn"] != "" {
				stopAction.TopicArn = aws.String(elem["topic_arn"].(string))
			}

			actions[elem["position"].(int)] = &ses.ReceiptAction{
				StopAction: stopAction,
			}
		}
	}

	if v, ok := d.GetOk("workmail_action"); ok {
		for _, element := range v.(*schema.Set).List() {
			elem := element.(map[string]interface{})

			workmailAction := &ses.WorkmailAction{
				OrganizationArn: aws.String(elem["organization_arn"].(string)),
			}

			if elem["topic_arn"] != "" {
				workmailAction.TopicArn = aws.String(elem["topic_arn"].(string))
			}

			actions[elem["position"].(int)] = &ses.ReceiptAction{
				WorkmailAction: workmailAction,
			}
		}
	}

	var keys []int
	for k := range actions {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	sortedActions := []*ses.ReceiptAction{}
	for _, k := range keys {
		sortedActions = append(sortedActions, actions[k])
	}

	receiptRule.Actions = sortedActions

	return receiptRule
}
