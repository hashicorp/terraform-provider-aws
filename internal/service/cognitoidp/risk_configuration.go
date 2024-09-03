// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
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

// @SDKResource("aws_cognito_risk_configuration", name="Risk Configuration")
func resourceRiskConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRiskConfigurationPut,
		ReadWithoutTimeout:   resourceRiskConfigurationRead,
		UpdateWithoutTimeout: resourceRiskConfigurationPut,
		DeleteWithoutTimeout: resourceRiskConfigurationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"account_takeover_risk_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				AtLeastOneOf: []string{
					"account_takeover_risk_configuration",
					"compromised_credentials_risk_configuration",
					"risk_exception_configuration",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrActions: {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"high_action": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"event_action": {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[awstypes.AccountTakeoverEventActionType](),
												},
												"notify": {
													Type:     schema.TypeBool,
													Required: true,
												},
											},
										},
									},
									"low_action": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"event_action": {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[awstypes.AccountTakeoverEventActionType](),
												},
												"notify": {
													Type:     schema.TypeBool,
													Required: true,
												},
											},
										},
									},
									"medium_action": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"event_action": {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[awstypes.AccountTakeoverEventActionType](),
												},
												"notify": {
													Type:     schema.TypeBool,
													Required: true,
												},
											},
										},
									},
								},
							},
						},
						"notify_configuration": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"block_email": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"html_body": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(6, 20000),
												},
												"subject": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 140),
												},
												"text_body": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(6, 20000),
												},
											},
										},
									},
									"from": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"mfa_email": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"html_body": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(6, 20000),
												},
												"subject": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 140),
												},
												"text_body": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(6, 20000),
												},
											},
										},
									},
									"no_action_email": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"html_body": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(6, 20000),
												},
												"subject": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 140),
												},
												"text_body": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(6, 20000),
												},
											},
										},
									},
									"reply_to": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"source_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
								},
							},
						},
					},
				},
			},
			names.AttrClientID: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"compromised_credentials_risk_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrActions: {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"event_action": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[awstypes.CompromisedCredentialsEventActionType](),
									},
								},
							},
						},
						"event_filter": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem: &schema.Schema{
								Type:             schema.TypeString,
								ValidateDiagFunc: enum.Validate[awstypes.EventFilterType](),
							},
						},
					},
				},
			},
			"risk_exception_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"blocked_ip_range_list": {
							Type:     schema.TypeSet,
							Optional: true,
							MinItems: 1,
							MaxItems: 200,
							AtLeastOneOf: []string{
								"risk_exception_configuration.0.blocked_ip_range_list",
								"risk_exception_configuration.0.skipped_ip_range_list",
							},
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateFunc: validation.All(
									validation.IsCIDR,
								),
							},
						},
						"skipped_ip_range_list": {
							Type:     schema.TypeSet,
							Optional: true,
							MinItems: 1,
							MaxItems: 200,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateFunc: validation.All(
									validation.IsCIDR,
								)},
						},
					},
				},
			},
			names.AttrUserPoolID: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validUserPoolID,
			},
		},
	}
}

func resourceRiskConfigurationPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPClient(ctx)

	userPoolID := d.Get(names.AttrUserPoolID).(string)
	id := userPoolID
	input := &cognitoidentityprovider.SetRiskConfigurationInput{
		UserPoolId: aws.String(userPoolID),
	}

	if v, ok := d.GetOk(names.AttrClientID); ok {
		v := v.(string)
		input.ClientId = aws.String(v)
		id = userPoolID + riskConfigurationResourceIDSeparator + v
	}

	if v, ok := d.GetOk("account_takeover_risk_configuration"); ok && len(v.([]interface{})) > 0 {
		input.AccountTakeoverRiskConfiguration = expandAccountTakeoverRiskConfigurationType(v.([]interface{}))
	}

	if v, ok := d.GetOk("compromised_credentials_risk_configuration"); ok && len(v.([]interface{})) > 0 {
		input.CompromisedCredentialsRiskConfiguration = expandCompromisedCredentialsRiskConfigurationType(v.([]interface{}))
	}

	if v, ok := d.GetOk("risk_exception_configuration"); ok && len(v.([]interface{})) > 0 {
		input.RiskExceptionConfiguration = expandRiskExceptionConfigurationType(v.([]interface{}))
	}

	_, err := conn.SetRiskConfiguration(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Cognito Risk Configuration (%s): %s", id, err)
	}

	if d.IsNewResource() {
		d.SetId(id)
	}

	return append(diags, resourceRiskConfigurationRead(ctx, d, meta)...)
}

func resourceRiskConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPClient(ctx)

	userPoolID, clientID, err := riskConfigurationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	riskConfig, err := findRiskConfigurationByTwoPartKey(ctx, conn, userPoolID, clientID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Cognito Risk Configuration %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Cognito Risk Configuration (%s): %s", d.Id(), err)
	}

	if err := d.Set("account_takeover_risk_configuration", flattenAccountTakeoverRiskConfigurationType(riskConfig.AccountTakeoverRiskConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting account_takeover_risk_configuration: %s", err)
	}
	if clientID != "" {
		d.Set(names.AttrClientID, clientID)
	}
	if err := d.Set("compromised_credentials_risk_configuration", flattenCompromisedCredentialsRiskConfiguration(riskConfig.CompromisedCredentialsRiskConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting compromised_credentials_risk_configuration: %s", err)
	}
	if riskConfig.RiskExceptionConfiguration != nil {
		if err := d.Set("risk_exception_configuration", flattenRiskExceptionConfiguration(riskConfig.RiskExceptionConfiguration)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting risk_exception_configuration: %s", err)
		}
	}
	d.Set(names.AttrUserPoolID, userPoolID)

	return diags
}

func resourceRiskConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPClient(ctx)

	userPoolID, clientID, err := riskConfigurationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &cognitoidentityprovider.SetRiskConfigurationInput{
		UserPoolId: aws.String(userPoolID),
	}
	if clientID != "" {
		input.ClientId = aws.String(clientID)
	}

	log.Printf("[DEBUG] Deleting Cognito Risk Configuration: %s", d.Id())
	_, err = conn.SetRiskConfiguration(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Cognito Risk Configuration (%s): %s", d.Id(), err)
	}

	return diags
}

const riskConfigurationResourceIDSeparator = ":"

func riskConfigurationParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, riskConfigurationResourceIDSeparator)

	if len(parts) == 1 && parts[0] != "" {
		return parts[0], "", nil
	}

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected UserPoolID%[2]sClientID or UserPoolID", id, riskConfigurationResourceIDSeparator)
}

func findRiskConfigurationByTwoPartKey(ctx context.Context, conn *cognitoidentityprovider.Client, userPoolID, clientID string) (*awstypes.RiskConfigurationType, error) {
	input := &cognitoidentityprovider.DescribeRiskConfigurationInput{
		UserPoolId: aws.String(userPoolID),
	}
	if clientID != "" {
		input.ClientId = aws.String(clientID)
	}

	output, err := conn.DescribeRiskConfiguration(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.RiskConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.RiskConfiguration, nil
}

func expandRiskExceptionConfigurationType(tfList []interface{}) *awstypes.RiskExceptionConfigurationType {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &awstypes.RiskExceptionConfigurationType{}

	if v, ok := tfMap["blocked_ip_range_list"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.BlockedIPRangeList = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["skipped_ip_range_list"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SkippedIPRangeList = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func flattenRiskExceptionConfiguration(apiObject *awstypes.RiskExceptionConfigurationType) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.BlockedIPRangeList; v != nil {
		tfMap["blocked_ip_range_list"] = v
	}

	if v := apiObject.SkippedIPRangeList; v != nil {
		tfMap["skipped_ip_range_list"] = v
	}

	return []interface{}{tfMap}
}

func expandCompromisedCredentialsRiskConfigurationType(tfList []interface{}) *awstypes.CompromisedCredentialsRiskConfigurationType {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &awstypes.CompromisedCredentialsRiskConfigurationType{}

	if v, ok := tfMap["event_filter"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.EventFilter = flex.ExpandStringyValueSet[awstypes.EventFilterType](v)
	}

	if v, ok := tfMap[names.AttrActions].([]interface{}); ok && len(v) > 0 {
		apiObject.Actions = expandCompromisedCredentialsActionsType(v)
	}

	return apiObject
}

func flattenCompromisedCredentialsRiskConfiguration(apiObject *awstypes.CompromisedCredentialsRiskConfigurationType) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.EventFilter; v != nil {
		tfMap["event_filter"] = v
	}

	if v := apiObject.Actions; v != nil {
		tfMap[names.AttrActions] = flattenCompromisedCredentialsActions(v)
	}

	return []interface{}{tfMap}
}

func expandCompromisedCredentialsActionsType(tfList []interface{}) *awstypes.CompromisedCredentialsActionsType {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &awstypes.CompromisedCredentialsActionsType{}

	if v, ok := tfMap["event_action"].(string); ok && v != "" {
		apiObject.EventAction = awstypes.CompromisedCredentialsEventActionType(v)
	}

	return apiObject
}

func flattenCompromisedCredentialsActions(apiObject *awstypes.CompromisedCredentialsActionsType) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"event_action": apiObject.EventAction,
	}

	return []interface{}{tfMap}
}

func expandAccountTakeoverRiskConfigurationType(tfList []interface{}) *awstypes.AccountTakeoverRiskConfigurationType {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &awstypes.AccountTakeoverRiskConfigurationType{}

	if v, ok := tfMap[names.AttrActions].([]interface{}); ok && len(v) > 0 {
		apiObject.Actions = expandAccountTakeoverActionsType(v)
	}

	if v, ok := tfMap["notify_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.NotifyConfiguration = expandNotifyConfigurationType(v)
	}

	return apiObject
}

func flattenAccountTakeoverRiskConfigurationType(apiObject *awstypes.AccountTakeoverRiskConfigurationType) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Actions; v != nil {
		tfMap[names.AttrActions] = flattenAccountTakeoverActionsType(v)
	}

	if v := apiObject.NotifyConfiguration; v != nil {
		tfMap["notify_configuration"] = flattemNotifyConfigurationType(v)
	}

	return []interface{}{tfMap}
}

func expandAccountTakeoverActionsType(tfList []interface{}) *awstypes.AccountTakeoverActionsType {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &awstypes.AccountTakeoverActionsType{}

	if v, ok := tfMap["high_action"].([]interface{}); ok && len(v) > 0 {
		apiObject.HighAction = expandAccountTakeoverActionType(v)
	}

	if v, ok := tfMap["low_action"].([]interface{}); ok && len(v) > 0 {
		apiObject.LowAction = expandAccountTakeoverActionType(v)
	}

	if v, ok := tfMap["medium_action"].([]interface{}); ok && len(v) > 0 {
		apiObject.MediumAction = expandAccountTakeoverActionType(v)
	}

	return apiObject
}

func flattenAccountTakeoverActionsType(apiObject *awstypes.AccountTakeoverActionsType) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.HighAction; v != nil {
		tfMap["high_action"] = flattenAccountTakeoverActionType(v)
	}

	if v := apiObject.LowAction; v != nil {
		tfMap["low_action"] = flattenAccountTakeoverActionType(v)
	}

	if v := apiObject.MediumAction; v != nil {
		tfMap["medium_action"] = flattenAccountTakeoverActionType(v)
	}

	return []interface{}{tfMap}
}

func expandAccountTakeoverActionType(tfList []interface{}) *awstypes.AccountTakeoverActionType {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &awstypes.AccountTakeoverActionType{}

	if v, ok := tfMap["event_action"].(string); ok && v != "" {
		apiObject.EventAction = awstypes.AccountTakeoverEventActionType(v)
	}

	if v, ok := tfMap["notify"].(bool); ok {
		apiObject.Notify = v
	}

	return apiObject
}

func flattenAccountTakeoverActionType(apiObject *awstypes.AccountTakeoverActionType) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"event_action": apiObject.EventAction,
		"notify":       apiObject.Notify,
	}

	return []interface{}{tfMap}
}

func expandNotifyConfigurationType(tfList []interface{}) *awstypes.NotifyConfigurationType {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &awstypes.NotifyConfigurationType{}

	if v, ok := tfMap["block_email"].([]interface{}); ok && len(v) > 0 {
		apiObject.BlockEmail = expandNotifyEmailType(v)
	}

	if v, ok := tfMap["from"].(string); ok && v != "" {
		apiObject.From = aws.String(v)
	}

	if v, ok := tfMap["mfa_email"].([]interface{}); ok && len(v) > 0 {
		apiObject.MfaEmail = expandNotifyEmailType(v)
	}

	if v, ok := tfMap["no_action_email"].([]interface{}); ok && len(v) > 0 {
		apiObject.NoActionEmail = expandNotifyEmailType(v)
	}

	if v, ok := tfMap["reply_to"].(string); ok && v != "" {
		apiObject.ReplyTo = aws.String(v)
	}

	if v, ok := tfMap["source_arn"].(string); ok && v != "" {
		apiObject.SourceArn = aws.String(v)
	}

	return apiObject
}

func flattemNotifyConfigurationType(apiObject *awstypes.NotifyConfigurationType) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.BlockEmail; v != nil {
		tfMap["block_email"] = flattenNotifyEmailType(v)
	}

	if v := apiObject.From; v != nil {
		tfMap["from"] = aws.ToString(v)
	}

	if v := apiObject.MfaEmail; v != nil {
		tfMap["mfa_email"] = flattenNotifyEmailType(v)
	}

	if v := apiObject.NoActionEmail; v != nil {
		tfMap["no_action_email"] = flattenNotifyEmailType(v)
	}

	if v := apiObject.ReplyTo; v != nil {
		tfMap["reply_to"] = aws.ToString(v)
	}

	if v := apiObject.SourceArn; v != nil {
		tfMap["source_arn"] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

func expandNotifyEmailType(tfList []interface{}) *awstypes.NotifyEmailType {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &awstypes.NotifyEmailType{}

	if v, ok := tfMap["html_body"].(string); ok && v != "" {
		apiObject.HtmlBody = aws.String(v)
	}

	if v, ok := tfMap["subject"].(string); ok && v != "" {
		apiObject.Subject = aws.String(v)
	}

	if v, ok := tfMap["text_body"].(string); ok && v != "" {
		apiObject.TextBody = aws.String(v)
	}

	return apiObject
}

func flattenNotifyEmailType(apiObject *awstypes.NotifyEmailType) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.HtmlBody; v != nil {
		tfMap["html_body"] = aws.ToString(v)
	}

	if v := apiObject.Subject; v != nil {
		tfMap["subject"] = aws.ToString(v)
	}

	if v := apiObject.TextBody; v != nil {
		tfMap["text_body"] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}
