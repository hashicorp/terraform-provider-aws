package cognitoidp

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceRiskConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceRiskConfigurationPut,
		Read:   resourceRiskConfigurationRead,
		Delete: resourceRiskConfigurationDelete,
		Update: resourceRiskConfigurationPut,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"user_pool_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validUserPoolID,
			},
			"client_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"account_takeover_risk_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"actions": {
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
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(cognitoidentityprovider.AccountTakeoverEventActionType_Values(), false),
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
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(cognitoidentityprovider.AccountTakeoverEventActionType_Values(), false),
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
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(cognitoidentityprovider.AccountTakeoverEventActionType_Values(), false),
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
													ValidateFunc: validation.StringLenBetween(6, 2000),
												},
												"subject": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 140),
												},
												"text_body": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(6, 2000),
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
													ValidateFunc: validation.StringLenBetween(6, 2000),
												},
												"subject": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 140),
												},
												"text_body": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(6, 2000),
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
													ValidateFunc: validation.StringLenBetween(6, 2000),
												},
												"subject": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 140),
												},
												"text_body": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(6, 2000),
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
			"compromised_credentials_risk_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"event_filter": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(cognitoidentityprovider.EventFilterType_Values(), false),
							},
						},
						"actions": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"event_action": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(cognitoidentityprovider.CompromisedCredentialsEventActionType_Values(), false),
									},
								},
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
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateFunc: validation.All(
									validation.StringLenBetween(0, 200),
									validation.IsCIDR,
								),
							},
						},
						"skipped_ip_range_list": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateFunc: validation.All(
									validation.StringLenBetween(0, 200),
									validation.IsCIDR,
								)},
						},
					},
				},
			},
		},
	}
}

func resourceRiskConfigurationPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CognitoIDPConn

	userPoolId := d.Get("user_pool_id").(string)
	id := userPoolId
	input := &cognitoidentityprovider.SetRiskConfigurationInput{
		UserPoolId: aws.String(userPoolId),
	}

	if v, ok := d.GetOk("client_id"); ok {
		input.ClientId = aws.String(v.(string))
		id = fmt.Sprintf("%s:%s", userPoolId, v.(string))
	}

	if v, ok := d.GetOk("risk_exception_configuration"); ok && len(v.([]interface{})) > 0 {
		input.RiskExceptionConfiguration = expandRiskExceptionConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("compromised_credentials_risk_configuration"); ok && len(v.([]interface{})) > 0 {
		input.CompromisedCredentialsRiskConfiguration = expandCompromisedCredentialsRiskConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("account_takeover_risk_configuration"); ok && len(v.([]interface{})) > 0 {
		input.AccountTakeoverRiskConfiguration = expandAccountTakeoverRiskConfiguration(v.([]interface{}))
	}

	_, err := conn.SetRiskConfiguration(input)

	if err != nil {
		return fmt.Errorf("error setting risk configuration: %w", err)
	}

	d.SetId(id)

	return resourceRiskConfigurationRead(d, meta)
}

func resourceRiskConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CognitoIDPConn

	riskConfig, err := FindRiskConfigurationById(conn, d.Id())

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, cognitoidentityprovider.ErrCodeResourceNotFoundException) {
		names.LogNotFoundRemoveState(names.CognitoIDP, names.ErrActionReading, ResRiskConfiguration, d.Id())
		d.SetId("")
		return nil
	}

	userPoolId, clientId, err := RiskConfigurationParseID(d.Id())
	if err != nil {
		return err
	}

	d.Set("user_pool_id", userPoolId)

	if clientId != "" {
		d.Set("client_id", clientId)
	}

	if riskConfig.RiskExceptionConfiguration != nil {
		if err := d.Set("risk_exception_configuration", flattenRiskExceptionConfiguration(riskConfig.RiskExceptionConfiguration)); err != nil {
			return fmt.Errorf("error setting risk_exception_configuration: %w", err)
		}
	}

	if err := d.Set("compromised_credentials_risk_configuration", flattenCompromisedCredentialsRiskConfiguration(riskConfig.CompromisedCredentialsRiskConfiguration)); err != nil {
		return fmt.Errorf("error setting compromised_credentials_risk_configuration: %w", err)
	}

	if err := d.Set("account_takeover_risk_configuration", flattenAccountTakeoverRiskConfiguration(riskConfig.AccountTakeoverRiskConfiguration)); err != nil {
		return fmt.Errorf("error setting account_takeover_risk_configuration: %w", err)
	}

	return nil
}

func resourceRiskConfigurationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CognitoIDPConn

	userPoolId, clientId, err := RiskConfigurationParseID(d.Id())
	if err != nil {
		return err
	}

	input := &cognitoidentityprovider.SetRiskConfigurationInput{
		UserPoolId: aws.String(userPoolId),
	}

	if clientId != "" {
		input.ClientId = aws.String(clientId)
	}

	_, err = conn.SetRiskConfiguration(input)

	if err != nil {
		return fmt.Errorf("error removing risk configuration: %w", err)
	}

	return nil
}

func expandRiskExceptionConfiguration(riskConfig []interface{}) *cognitoidentityprovider.RiskExceptionConfigurationType {
	config := riskConfig[0].(map[string]interface{})

	riskExceptionConfigurationType := &cognitoidentityprovider.RiskExceptionConfigurationType{}

	if v, ok := config["blocked_ip_range_list"].(*schema.Set); ok && v.Len() > 0 {
		riskExceptionConfigurationType.BlockedIPRangeList = flex.ExpandStringSet(v)
	}

	if v, ok := config["skipped_ip_range_list"].(*schema.Set); ok && v.Len() > 0 {
		riskExceptionConfigurationType.SkippedIPRangeList = flex.ExpandStringSet(v)
	}

	return riskExceptionConfigurationType
}

func flattenRiskExceptionConfiguration(apiObject *cognitoidentityprovider.RiskExceptionConfigurationType) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.BlockedIPRangeList; v != nil {
		tfMap["blocked_ip_range_list"] = flex.FlattenStringSet(v)
	}

	if v := apiObject.SkippedIPRangeList; v != nil {
		tfMap["skipped_ip_range_list"] = flex.FlattenStringSet(v)
	}

	return []interface{}{tfMap}
}

func expandCompromisedCredentialsRiskConfiguration(riskConfig []interface{}) *cognitoidentityprovider.CompromisedCredentialsRiskConfigurationType {
	config := riskConfig[0].(map[string]interface{})

	riskExceptionConfigurationType := &cognitoidentityprovider.CompromisedCredentialsRiskConfigurationType{}

	if v, ok := config["event_filter"].(*schema.Set); ok && v.Len() > 0 {
		riskExceptionConfigurationType.EventFilter = flex.ExpandStringSet(v)
	}

	if v, ok := config["actions"].([]interface{}); ok && len(v) > 0 {
		riskExceptionConfigurationType.Actions = expandCompromisedCredentialsActions(v)
	}

	return riskExceptionConfigurationType
}

func flattenCompromisedCredentialsRiskConfiguration(apiObject *cognitoidentityprovider.CompromisedCredentialsRiskConfigurationType) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.EventFilter; v != nil {
		tfMap["event_filter"] = flex.FlattenStringSet(v)
	}

	if v := apiObject.Actions; v != nil {
		tfMap["actions"] = flattenCompromisedCredentialsActions(v)
	}

	return []interface{}{tfMap}
}

func expandCompromisedCredentialsActions(riskConfig []interface{}) *cognitoidentityprovider.CompromisedCredentialsActionsType {
	config := riskConfig[0].(map[string]interface{})

	compromisedCredentialsAction := &cognitoidentityprovider.CompromisedCredentialsActionsType{}

	if v, ok := config["event_action"].(string); ok && v != "" {
		compromisedCredentialsAction.EventAction = aws.String(v)
	}

	return compromisedCredentialsAction
}

func flattenCompromisedCredentialsActions(apiObject *cognitoidentityprovider.CompromisedCredentialsActionsType) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.EventAction; v != nil {
		tfMap["event_action"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

func expandAccountTakeoverRiskConfiguration(riskConfig []interface{}) *cognitoidentityprovider.AccountTakeoverRiskConfigurationType {
	config := riskConfig[0].(map[string]interface{})

	accountTakeoverRiskConfiguration := &cognitoidentityprovider.AccountTakeoverRiskConfigurationType{}

	if v, ok := config["notify_configuration"].([]interface{}); ok && len(v) > 0 {
		accountTakeoverRiskConfiguration.NotifyConfiguration = expandNotifyConfiguration(v)
	}

	if v, ok := config["actions"].([]interface{}); ok && len(v) > 0 {
		accountTakeoverRiskConfiguration.Actions = expandAccountTakeoverActions(v)
	}

	return accountTakeoverRiskConfiguration
}

func flattenAccountTakeoverRiskConfiguration(apiObject *cognitoidentityprovider.AccountTakeoverRiskConfigurationType) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Actions; v != nil {
		tfMap["actions"] = flattenAccountTakeoverActions(v)
	}

	if v := apiObject.NotifyConfiguration; v != nil {
		tfMap["notify_configuration"] = flattenNotifyConfiguration(v)
	}

	return []interface{}{tfMap}
}

func expandAccountTakeoverActions(riskConfig []interface{}) *cognitoidentityprovider.AccountTakeoverActionsType {
	config := riskConfig[0].(map[string]interface{})

	actions := &cognitoidentityprovider.AccountTakeoverActionsType{}

	if v, ok := config["high_action"].([]interface{}); ok && len(v) > 0 {
		actions.HighAction = expandAccountTakeoverAction(v)
	}

	if v, ok := config["low_action"].([]interface{}); ok && len(v) > 0 {
		actions.LowAction = expandAccountTakeoverAction(v)
	}

	if v, ok := config["medium_action"].([]interface{}); ok && len(v) > 0 {
		actions.MediumAction = expandAccountTakeoverAction(v)
	}

	return actions
}

func flattenAccountTakeoverActions(apiObject *cognitoidentityprovider.AccountTakeoverActionsType) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.HighAction; v != nil {
		tfMap["high_action"] = flattenAccountTakeoverAction(v)
	}

	if v := apiObject.LowAction; v != nil {
		tfMap["low_action"] = flattenAccountTakeoverAction(v)
	}

	if v := apiObject.MediumAction; v != nil {
		tfMap["medium_action"] = flattenAccountTakeoverAction(v)
	}

	return []interface{}{tfMap}
}

func expandAccountTakeoverAction(riskConfig []interface{}) *cognitoidentityprovider.AccountTakeoverActionType {
	config := riskConfig[0].(map[string]interface{})

	action := &cognitoidentityprovider.AccountTakeoverActionType{}

	if v, ok := config["event_action"].(string); ok && v != "" {
		action.EventAction = aws.String(v)
	}

	if v, ok := config["notify"].(bool); ok {
		action.Notify = aws.Bool(v)
	}

	return action
}

func flattenAccountTakeoverAction(apiObject *cognitoidentityprovider.AccountTakeoverActionType) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.EventAction; v != nil {
		tfMap["event_action"] = aws.StringValue(v)
	}

	if v := apiObject.Notify; v != nil {
		tfMap["notify"] = aws.BoolValue(v)
	}

	return []interface{}{tfMap}
}

func expandNotifyConfiguration(riskConfig []interface{}) *cognitoidentityprovider.NotifyConfigurationType {
	config := riskConfig[0].(map[string]interface{})

	notifConfig := &cognitoidentityprovider.NotifyConfigurationType{}

	if v, ok := config["from"].(string); ok && v != "" {
		notifConfig.From = aws.String(v)
	}

	if v, ok := config["reply_to"].(string); ok && v != "" {
		notifConfig.ReplyTo = aws.String(v)
	}

	if v, ok := config["source_arn"].(string); ok && v != "" {
		notifConfig.SourceArn = aws.String(v)
	}

	if v, ok := config["block_email"].([]interface{}); ok && len(v) > 0 {
		notifConfig.BlockEmail = expandNotifyEmail(v)
	}

	if v, ok := config["mfa_email"].([]interface{}); ok && len(v) > 0 {
		notifConfig.MfaEmail = expandNotifyEmail(v)
	}

	if v, ok := config["no_action_email"].([]interface{}); ok && len(v) > 0 {
		notifConfig.NoActionEmail = expandNotifyEmail(v)
	}

	return notifConfig
}

func flattenNotifyConfiguration(apiObject *cognitoidentityprovider.NotifyConfigurationType) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.From; v != nil {
		tfMap["from"] = aws.StringValue(v)
	}

	if v := apiObject.ReplyTo; v != nil {
		tfMap["reply_to"] = aws.StringValue(v)
	}

	if v := apiObject.SourceArn; v != nil {
		tfMap["source_arn"] = aws.StringValue(v)
	}

	if v := apiObject.BlockEmail; v != nil {
		tfMap["block_email"] = flattenNotifyEmail(v)
	}

	if v := apiObject.MfaEmail; v != nil {
		tfMap["mfa_email"] = flattenNotifyEmail(v)
	}

	if v := apiObject.NoActionEmail; v != nil {
		tfMap["no_action_email"] = flattenNotifyEmail(v)
	}

	return []interface{}{tfMap}
}

func expandNotifyEmail(riskConfig []interface{}) *cognitoidentityprovider.NotifyEmailType {
	config := riskConfig[0].(map[string]interface{})

	notifyEmail := &cognitoidentityprovider.NotifyEmailType{}

	if v, ok := config["html_body"].(string); ok && v != "" {
		notifyEmail.HtmlBody = aws.String(v)
	}

	if v, ok := config["subject"].(string); ok && v != "" {
		notifyEmail.Subject = aws.String(v)
	}

	if v, ok := config["text_body"].(string); ok && v != "" {
		notifyEmail.TextBody = aws.String(v)
	}

	return notifyEmail
}

func flattenNotifyEmail(apiObject *cognitoidentityprovider.NotifyEmailType) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.HtmlBody; v != nil {
		tfMap["html_body"] = aws.StringValue(v)
	}

	if v := apiObject.Subject; v != nil {
		tfMap["subject"] = aws.StringValue(v)
	}

	if v := apiObject.TextBody; v != nil {
		tfMap["text_body"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

func RiskConfigurationParseID(id string) (string, string, error) {
	parts := strings.Split(id, ":")

	if len(parts) > 2 || len(parts) < 1 {
		return "", "", fmt.Errorf("Wrong format of resource: %s. Please follow 'userpool-id/client-id' or 'userpool-id'", id)
	}

	if len(parts) == 2 {
		return parts[0], parts[1], nil
	} else {
		return parts[0], "", nil

	}
}
