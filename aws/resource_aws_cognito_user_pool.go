package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"
)

func resourceAwsCognitoUserPool() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCognitoUserPoolCreate,
		Read:   resourceAwsCognitoUserPoolRead,
		Update: resourceAwsCognitoUserPoolUpdate,
		Delete: resourceAwsCognitoUserPoolDelete,

		Schema: map[string]*schema.Schema{
			"alias_attributes": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateCognitoUserPoolAliasAttribute,
				},
			},

			"auto_verified_attributes": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateCognitoUserPoolAutoVerifiedAttribute,
				},
			},

			"email_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 0,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"reply_to_email_address": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateCognitoUserPoolReplyEmailAddress,
						},
						"source_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateArn,
						},
					},
				},
			},

			"email_verification_subject": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateCognitoUserPoolEmailVerificationSubject,
			},

			"email_verification_message": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateCognitoUserPoolEmailVerificationMessage,
			},

			"mfa_configuration": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      cognitoidentityprovider.UserPoolMfaTypeOff,
				ValidateFunc: validateCognitoUserPoolMfaConfiguration,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"password_policy": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 0,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"minimum_length": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validateIntegerInRange(6, 99),
						},
						"require_lowercase": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"require_numbers": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"require_symbols": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"require_uppercase": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},

			"schema": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				MaxItems: 50,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attribute_data_type": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								cognitoidentityprovider.AttributeDataTypeString,
								cognitoidentityprovider.AttributeDataTypeNumber,
								cognitoidentityprovider.AttributeDataTypeDateTime,
								cognitoidentityprovider.AttributeDataTypeBoolean,
							}, false),
						},
						"developer_only_attribute": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"mutable": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"name": {
							Type:         schema.TypeBool,
							Optional:     true,
							ValidateFunc: validateCognitoUserPoolSchemaName,
						},
						"number_attribute_constraints": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 0,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"min_value": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"max_value": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"required": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"string_attribute_constraints": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 0,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"min_length": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"max_length": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},

			"sms_authentication_message": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateCognitoUserPoolSmsAuthenticationMessage,
			},

			"sms_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 0,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"external_id": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateArn,
						},
						"sns_caller_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},
					},
				},
			},

			"sms_verification_message": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateCognitoUserPoolSmsVerificationMessage,
			},

			"tags": tagsSchema(),

			"verification_message_template": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"default_email_option": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      cognitoidentityprovider.DefaultEmailOptionTypeConfirmWithCode,
							ValidateFunc: validateCognitoUserPoolTemplateDefaultEmailOption,
						},
						"email_message": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validateCognitoUserPoolTemplateEmailMessage,
						},
						"email_message_by_link": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validateCognitoUserPoolTemplateEmailMessageByLink,
						},
						"email_subject": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validateCognitoUserPoolTemplateEmailSubject,
						},
						"email_subject_by_link": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validateCognitoUserPoolTemplateEmailSubjectByLink,
						},
						"sms_message": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validateCognitoUserPoolTemplateSmsMessage,
						},
					},
				},
			},
		},
	}
}

func resourceAwsCognitoUserPoolCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	params := &cognitoidentityprovider.CreateUserPoolInput{
		PoolName: aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("alias_attributes"); ok {
		params.AliasAttributes = expandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("auto_verified_attributes"); ok {
		params.AutoVerifiedAttributes = expandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("email_configuration"); ok {
		configs := v.([]interface{})
		config, ok := configs[0].(map[string]interface{})

		if !ok {
			return errors.New("email_configuration is <nil>")
		}

		if config != nil {
			emailConfigurationType := &cognitoidentityprovider.EmailConfigurationType{}

			if v, ok := config["reply_to_email_address"]; ok && v.(string) != "" {
				emailConfigurationType.ReplyToEmailAddress = aws.String(v.(string))
			}

			if v, ok := config["source_arn"]; ok && v.(string) != "" {
				emailConfigurationType.SourceArn = aws.String(v.(string))
			}

			params.EmailConfiguration = emailConfigurationType
		}
	}

	if v, ok := d.GetOk("email_verification_subject"); ok {
		params.EmailVerificationSubject = aws.String(v.(string))
	}

	if v, ok := d.GetOk("email_verification_message"); ok {
		params.EmailVerificationMessage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("mfa_configuration"); ok {
		params.MfaConfiguration = aws.String(v.(string))
	}

	policies := &cognitoidentityprovider.UserPoolPolicyType{}

	if v, ok := d.GetOk("password_policy"); ok {
		configs := v.([]interface{})
		config, ok := configs[0].(map[string]interface{})

		if !ok {
			return errors.New("password_policy is <nil>")
		}

		if config != nil {
			policies.PasswordPolicy = expandCognitoUserPoolPasswordPolicy(config)
		}
	}

	params.Policies = policies

	if v, ok := d.GetOk("schema"); ok {
		configs := v.([]interface{})
		config, ok := configs[0].(map[string]interface{})

		if !ok {
			return errors.New("schema is <nil>")
		}

		if config != nil {
			params.Schema = expandCognitoUserPoolSchema(config)
		}
	}

	if v, ok := d.GetOk("sms_authentication_message"); ok {
		params.SmsAuthenticationMessage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("sms_configuration"); ok {
		configs := v.([]interface{})
		config, ok := configs[0].(map[string]interface{})

		if !ok {
			return errors.New("sms_configuration is <nil>")
		}

		if config != nil {
			smsConfigurationType := &cognitoidentityprovider.SmsConfigurationType{
				SnsCallerArn: aws.String(config["sns_caller_arn"].(string)),
			}

			if v, ok := config["external_id"]; ok && v.(string) != "" {
				smsConfigurationType.ExternalId = aws.String(v.(string))
			}

			params.SmsConfiguration = smsConfigurationType
		}
	}

	if v, ok := d.GetOk("verification_message_template"); ok {
		configs := v.([]interface{})
		config, ok := configs[0].(map[string]interface{})

		if !ok {
			return errors.New("verification_message_template is <nil>")
		}

		if config != nil {
			params.VerificationMessageTemplate = expandCognitoUserPoolVerificationMessageTemplate(config)
		}
	}

	if v, ok := d.GetOk("sms_verification_message"); ok {
		params.SmsVerificationMessage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tags"); ok {
		params.UserPoolTags = tagsFromMapGeneric(v.(map[string]interface{}))
	}
	log.Printf("[DEBUG] Creating Cognito User Pool: %s", params)

	resp, err := conn.CreateUserPool(params)

	if err != nil {
		return errwrap.Wrapf("Error creating Cognito User Pool: {{err}}", err)
	}

	d.SetId(*resp.UserPool.Id)

	return resourceAwsCognitoUserPoolRead(d, meta)
}

func resourceAwsCognitoUserPoolRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	params := &cognitoidentityprovider.DescribeUserPoolInput{
		UserPoolId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading Cognito User Pool: %s", params)

	resp, err := conn.DescribeUserPool(params)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "ResourceNotFoundException" {
			log.Printf("[WARN] Cognito User Pool %s is already gone", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	if resp.UserPool.AliasAttributes != nil {
		d.Set("alias_attributes", flattenStringList(resp.UserPool.AliasAttributes))
	}
	if resp.UserPool.AutoVerifiedAttributes != nil {
		d.Set("auto_verified_attributes", flattenStringList(resp.UserPool.AutoVerifiedAttributes))
	}
	if resp.UserPool.EmailVerificationSubject != nil {
		d.Set("email_verification_subject", *resp.UserPool.EmailVerificationSubject)
	}
	if resp.UserPool.EmailVerificationMessage != nil {
		d.Set("email_verification_message", *resp.UserPool.EmailVerificationMessage)
	}
	if resp.UserPool.MfaConfiguration != nil {
		d.Set("mfa_configuration", *resp.UserPool.MfaConfiguration)
	}
	if resp.UserPool.SmsVerificationMessage != nil {
		d.Set("sms_verification_message", *resp.UserPool.SmsVerificationMessage)
	}
	if resp.UserPool.SmsAuthenticationMessage != nil {
		d.Set("sms_authentication_message", *resp.UserPool.SmsAuthenticationMessage)
	}

	if err := d.Set("email_configuration", flattenCognitoUserPoolEmailConfiguration(resp.UserPool.EmailConfiguration)); err != nil {
		return errwrap.Wrapf("Failed setting email_configuration: {{err}}", err)
	}

	if resp.UserPool.Policies != nil && resp.UserPool.Policies.PasswordPolicy != nil {
		if err := d.Set("password_policy", flattenCognitoUserPoolPasswordPolicy(resp.UserPool.Policies.PasswordPolicy)); err != nil {
			return errwrap.Wrapf("Failed setting password_policy: {{err}}", err)
		}
	}

	if err := d.Set("sms_configuration", flattenCognitoUserPoolSmsConfiguration(resp.UserPool.SmsConfiguration)); err != nil {
		return errwrap.Wrapf("Failed setting sms_configuration: {{err}}", err)
	}

	if err := d.Set("verification_message_template", flattenCognitoUserPoolVerificationMessageTemplate(resp.UserPool.VerificationMessageTemplate)); err != nil {
		return errwrap.Wrapf("Failed setting verification_message_template: {{err}}", err)
	}

	d.Set("tags", tagsToMapGeneric(resp.UserPool.UserPoolTags))

	return nil
}

func resourceAwsCognitoUserPoolUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	params := &cognitoidentityprovider.UpdateUserPoolInput{
		UserPoolId: aws.String(d.Id()),
	}

	// TODO - Handle update of AliasAttributes

	if d.HasChange("auto_verified_attributes") {
		params.AutoVerifiedAttributes = expandStringList(d.Get("auto_verified_attributes").([]interface{}))
	}

	if d.HasChange("email_configuration") {
		configs := d.Get("email_configuration").([]interface{})
		config, ok := configs[0].(map[string]interface{})

		if !ok {
			return errors.New("email_configuration is <nil>")
		}

		if config != nil {
			emailConfigurationType := &cognitoidentityprovider.EmailConfigurationType{}

			if v, ok := config["reply_to_email_address"]; ok && v.(string) != "" {
				emailConfigurationType.ReplyToEmailAddress = aws.String(v.(string))
			}

			if v, ok := config["source_arn"]; ok && v.(string) != "" {
				emailConfigurationType.SourceArn = aws.String(v.(string))
			}

			params.EmailConfiguration = emailConfigurationType
		}
	}

	if d.HasChange("email_verification_subject") {
		v := d.Get("email_verification_subject").(string)
		if v == "" {
			return errors.New("email_verification_subject cannot be set to nil")
		}
		params.EmailVerificationSubject = aws.String(v)
	}

	if d.HasChange("email_verification_message") {
		v := d.Get("email_verification_message").(string)
		if v == "" {
			return errors.New("email_verification_message cannot be set to nil")
		}
		params.EmailVerificationMessage = aws.String(v)
	}

	if d.HasChange("mfa_configuration") {
		params.MfaConfiguration = aws.String(d.Get("mfa_configuration").(string))
	}

	if d.HasChange("sms_authentication_message") {
		params.SmsAuthenticationMessage = aws.String(d.Get("sms_authentication_message").(string))
	}

	policies := &cognitoidentityprovider.UserPoolPolicyType{}
	if d.HasChange("password_policy") {
		configs := d.Get("password_policy").([]interface{})
		config, ok := configs[0].(map[string]interface{})

		if !ok {
			return errors.New("password_policy is <nil>")
		}

		if config != nil {
			policies.PasswordPolicy = expandCognitoUserPoolPasswordPolicy(config)
		}
	}
	params.Policies = policies

	if d.HasChange("sms_configuration") {
		configs := d.Get("sms_configuration").([]interface{})
		config, ok := configs[0].(map[string]interface{})

		if !ok {
			return errors.New("sms_configuration is <nil>")
		}

		if config != nil {
			smsConfigurationType := &cognitoidentityprovider.SmsConfigurationType{
				SnsCallerArn: aws.String(config["sns_caller_arn"].(string)),
			}

			if v, ok := config["external_id"]; ok && v.(string) != "" {
				smsConfigurationType.ExternalId = aws.String(v.(string))
			}

			params.SmsConfiguration = smsConfigurationType
		}
	}

	if d.HasChange("verification_message_template") {
		configs := d.Get("verification_message_template").([]interface{})
		config, ok := configs[0].(map[string]interface{})

		if !ok {
			return errors.New("verification_message_template is <nil>")
		}

		if config != nil {
			params.VerificationMessageTemplate = expandCognitoUserPoolVerificationMessageTemplate(config)
		}
	}

	if d.HasChange("sms_verification_message") {
		v := d.Get("sms_verification_message").(string)
		if v == "" {
			return errors.New("sms_verification_message cannot be set to nil")
		}
		params.SmsVerificationMessage = aws.String(v)
	}

	if d.HasChange("tags") {
		params.UserPoolTags = tagsFromMapGeneric(d.Get("tags").(map[string]interface{}))
	}

	log.Printf("[DEBUG] Updating Cognito User Pool: %s", params)

	_, err := conn.UpdateUserPool(params)
	if err != nil {
		return errwrap.Wrapf("Error updating Cognito User pool: {{err}}", err)
	}

	return resourceAwsCognitoUserPoolRead(d, meta)
}

func resourceAwsCognitoUserPoolDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	params := &cognitoidentityprovider.DeleteUserPoolInput{
		UserPoolId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Cognito User Pool: %s", params)

	_, err := conn.DeleteUserPool(params)

	if err != nil {
		return errwrap.Wrapf("Error deleting user pool: {{err}}", err)
	}

	return nil
}
