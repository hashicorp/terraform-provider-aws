package aws

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsCognitoUserPool() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCognitoUserPoolCreate,
		Read:   resourceAwsCognitoUserPoolRead,
		Update: resourceAwsCognitoUserPoolUpdate,
		Delete: resourceAwsCognitoUserPoolDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		// https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_CreateUserPool.html
		Schema: map[string]*schema.Schema{
			"admin_create_user_config": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allow_admin_create_user_only": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"invite_message_template": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"email_message": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validateCognitoUserPoolInviteTemplateEmailMessage,
									},
									"email_subject": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validateCognitoUserPoolTemplateEmailSubject,
									},
									"sms_message": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validateCognitoUserPoolInviteTemplateSmsMessage,
									},
								},
							},
						},
						"unused_account_validity_days": {
							Type:          schema.TypeInt,
							Optional:      true,
							Computed:      true,
							Deprecated:    "Use password_policy.temporary_password_validity_days instead",
							ValidateFunc:  validation.IntBetween(0, 90),
							ConflictsWith: []string{"password_policy.0.temporary_password_validity_days"},
						},
					},
				},
			},

			"alias_attributes": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						cognitoidentityprovider.AliasAttributeTypeEmail,
						cognitoidentityprovider.AliasAttributeTypePhoneNumber,
						cognitoidentityprovider.AliasAttributeTypePreferredUsername,
					}, false),
				},
				ConflictsWith: []string{"username_attributes"},
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"auto_verified_attributes": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						cognitoidentityprovider.VerifiedAttributeTypePhoneNumber,
						cognitoidentityprovider.VerifiedAttributeTypeEmail,
					}, false),
				},
			},

			"creation_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"device_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"challenge_required_on_new_device": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"device_only_remembered_on_user_prompt": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},

			"email_configuration": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"reply_to_email_address": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.Any(
								validation.StringInSlice([]string{""}, false),
								validation.StringMatch(regexp.MustCompile(`[\p{L}\p{M}\p{S}\p{N}\p{P}]+@[\p{L}\p{M}\p{S}\p{N}\p{P}]+`),
									`must satisfy regular expression pattern: [\p{L}\p{M}\p{S}\p{N}\p{P}]+@[\p{L}\p{M}\p{S}\p{N}\p{P}]+`),
							),
						},
						"source_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateArn,
						},
						"email_sending_account": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  cognitoidentityprovider.EmailSendingAccountTypeCognitoDefault,
							ValidateFunc: validation.StringInSlice([]string{
								cognitoidentityprovider.EmailSendingAccountTypeCognitoDefault,
								cognitoidentityprovider.EmailSendingAccountTypeDeveloper,
							}, false),
						},
					},
				},
			},

			"email_verification_subject": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ValidateFunc:  validateCognitoUserPoolEmailVerificationSubject,
				ConflictsWith: []string{"verification_message_template.0.email_subject"},
			},

			"email_verification_message": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ValidateFunc:  validateCognitoUserPoolEmailVerificationMessage,
				ConflictsWith: []string{"verification_message_template.0.email_message"},
			},

			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"lambda_config": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"create_auth_challenge": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateArn,
						},
						"custom_message": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateArn,
						},
						"define_auth_challenge": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateArn,
						},
						"post_authentication": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateArn,
						},
						"post_confirmation": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateArn,
						},
						"pre_authentication": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateArn,
						},
						"pre_sign_up": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateArn,
						},
						"pre_token_generation": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateArn,
						},
						"user_migration": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateArn,
						},
						"verify_auth_challenge_response": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateArn,
						},
					},
				},
			},

			"last_modified_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"mfa_configuration": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  cognitoidentityprovider.UserPoolMfaTypeOff,
				ValidateFunc: validation.StringInSlice([]string{
					cognitoidentityprovider.UserPoolMfaTypeOff,
					cognitoidentityprovider.UserPoolMfaTypeOn,
					cognitoidentityprovider.UserPoolMfaTypeOptional,
				}, false),
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"password_policy": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"minimum_length": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(6, 99),
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
						"temporary_password_validity_days": {
							Type:          schema.TypeInt,
							Optional:      true,
							ValidateFunc:  validation.IntBetween(0, 365),
							ConflictsWith: []string{"admin_create_user_config.0.unused_account_validity_days"},
						},
					},
				},
			},

			"schema": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				MinItems: 1,
				MaxItems: 50,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attribute_data_type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
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
							ForceNew: true,
						},
						"mutable": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
						},
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validateCognitoUserPoolSchemaName,
						},
						"number_attribute_constraints": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"min_value": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
									"max_value": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
								},
							},
						},
						"required": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
						},
						"string_attribute_constraints": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"min_length": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
									"max_length": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
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
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"external_id": {
							Type:     schema.TypeString,
							Required: true,
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
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ValidateFunc:  validateCognitoUserPoolSmsVerificationMessage,
				ConflictsWith: []string{"verification_message_template.0.sms_message"},
			},

			"tags": tagsSchema(),

			"username_attributes": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						cognitoidentityprovider.UsernameAttributeTypeEmail,
						cognitoidentityprovider.UsernameAttributeTypePhoneNumber,
					}, false),
				},
				ConflictsWith: []string{"alias_attributes"},
			},

			"user_pool_add_ons": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"advanced_security_mode": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								cognitoidentityprovider.AdvancedSecurityModeTypeAudit,
								cognitoidentityprovider.AdvancedSecurityModeTypeEnforced,
								cognitoidentityprovider.AdvancedSecurityModeTypeOff,
							}, false),
						},
					},
				},
			},

			"verification_message_template": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"default_email_option": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  cognitoidentityprovider.DefaultEmailOptionTypeConfirmWithCode,
							ValidateFunc: validation.StringInSlice([]string{
								cognitoidentityprovider.DefaultEmailOptionTypeConfirmWithLink,
								cognitoidentityprovider.DefaultEmailOptionTypeConfirmWithCode,
							}, false),
						},
						"email_message": {
							Type:          schema.TypeString,
							Optional:      true,
							Computed:      true,
							ValidateFunc:  validateCognitoUserPoolTemplateEmailMessage,
							ConflictsWith: []string{"email_verification_message"},
						},
						"email_message_by_link": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validateCognitoUserPoolTemplateEmailMessageByLink,
						},
						"email_subject": {
							Type:          schema.TypeString,
							Optional:      true,
							Computed:      true,
							ValidateFunc:  validateCognitoUserPoolTemplateEmailSubject,
							ConflictsWith: []string{"email_verification_subject"},
						},
						"email_subject_by_link": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validateCognitoUserPoolTemplateEmailSubjectByLink,
						},
						"sms_message": {
							Type:          schema.TypeString,
							Optional:      true,
							Computed:      true,
							ValidateFunc:  validateCognitoUserPoolTemplateSmsMessage,
							ConflictsWith: []string{"sms_verification_message"},
						},
					},
				},
			},
			"risk_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account_takeover_risk_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"actions": {
										Type:     schema.TypeList,
										Optional: true,
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
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.StringInSlice([]string{
																	cognitoidentityprovider.AccountTakeoverEventActionTypeBlock,
																	cognitoidentityprovider.AccountTakeoverEventActionTypeMfaIfConfigured,
																	cognitoidentityprovider.AccountTakeoverEventActionTypeMfaRequired,
																	cognitoidentityprovider.AccountTakeoverEventActionTypeNoAction,
																}, false),
															},
															"notify": {
																Type:     schema.TypeBool,
																Optional: true,
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
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.StringInSlice([]string{
																	cognitoidentityprovider.AccountTakeoverEventActionTypeBlock,
																	cognitoidentityprovider.AccountTakeoverEventActionTypeMfaIfConfigured,
																	cognitoidentityprovider.AccountTakeoverEventActionTypeMfaRequired,
																	cognitoidentityprovider.AccountTakeoverEventActionTypeNoAction,
																}, false),
															},
															"notify": {
																Type:     schema.TypeBool,
																Optional: true,
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
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.StringInSlice([]string{
																	cognitoidentityprovider.AccountTakeoverEventActionTypeBlock,
																	cognitoidentityprovider.AccountTakeoverEventActionTypeMfaIfConfigured,
																	cognitoidentityprovider.AccountTakeoverEventActionTypeMfaRequired,
																	cognitoidentityprovider.AccountTakeoverEventActionTypeNoAction,
																}, false),
															},
															"notify": {
																Type:     schema.TypeBool,
																Optional: true,
															},
														},
													},
												},
											},
										},
									},
									"notify_configuration": {
										Type:     schema.TypeList,
										Optional: true,
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
																Type:     schema.TypeString,
																Optional: true,
															},
															"subject": {
																Type:     schema.TypeString,
																Optional: true,
															},
															"text_body": {
																Type:     schema.TypeString,
																Optional: true,
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
																Type:     schema.TypeString,
																Optional: true,
															},
															"subject": {
																Type:     schema.TypeString,
																Optional: true,
															},
															"text_body": {
																Type:     schema.TypeString,
																Optional: true,
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
																Type:     schema.TypeString,
																Optional: true,
															},
															"subject": {
																Type:     schema.TypeString,
																Optional: true,
															},
															"text_body": {
																Type:     schema.TypeString,
																Optional: true,
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
													Optional:     true,
													ValidateFunc: validateArn,
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
									"actions": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"event_action": {
													Type: schema.TypeString,
													ValidateFunc: validation.StringInSlice([]string{
														cognitoidentityprovider.CompromisedCredentialsEventActionTypeBlock,
														cognitoidentityprovider.CompromisedCredentialsEventActionTypeNoAction,
													}, false),
												},
											},
										},
									},
									"event_filter": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
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
									"block_ip_range_list": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.IsCIDR,
										},
									},
									"skip_ip_range_list": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.IsCIDR,
										},
									},
								},
							},
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

	if v, ok := d.GetOk("admin_create_user_config"); ok {
		configs := v.([]interface{})
		config, ok := configs[0].(map[string]interface{})

		if ok && config != nil {
			params.AdminCreateUserConfig = expandCognitoUserPoolAdminCreateUserConfig(config)
		}
	}

	if v, ok := d.GetOk("alias_attributes"); ok {
		params.AliasAttributes = expandStringList(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("auto_verified_attributes"); ok {
		params.AutoVerifiedAttributes = expandStringList(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("email_configuration"); ok {
		configs := v.([]interface{})
		config, ok := configs[0].(map[string]interface{})

		if ok && config != nil {
			emailConfigurationType := &cognitoidentityprovider.EmailConfigurationType{}

			if v, ok := config["reply_to_email_address"]; ok && v.(string) != "" {
				emailConfigurationType.ReplyToEmailAddress = aws.String(v.(string))
			}

			if v, ok := config["source_arn"]; ok && v.(string) != "" {
				emailConfigurationType.SourceArn = aws.String(v.(string))
			}

			if v, ok := config["email_sending_account"]; ok && v.(string) != "" {
				emailConfigurationType.EmailSendingAccount = aws.String(v.(string))
			}

			params.EmailConfiguration = emailConfigurationType
		}
	}

	if v, ok := d.GetOk("admin_create_user_config"); ok {
		configs := v.([]interface{})
		config, ok := configs[0].(map[string]interface{})

		if ok && config != nil {
			params.AdminCreateUserConfig = expandCognitoUserPoolAdminCreateUserConfig(config)
		}
	}

	if v, ok := d.GetOk("device_configuration"); ok {
		configs := v.([]interface{})
		config, ok := configs[0].(map[string]interface{})

		if ok && config != nil {
			params.DeviceConfiguration = expandCognitoUserPoolDeviceConfiguration(config)
		}
	}

	if v, ok := d.GetOk("email_verification_subject"); ok {
		params.EmailVerificationSubject = aws.String(v.(string))
	}

	if v, ok := d.GetOk("email_verification_message"); ok {
		params.EmailVerificationMessage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("lambda_config"); ok {
		configs := v.([]interface{})
		config, ok := configs[0].(map[string]interface{})

		if ok && config != nil {
			params.LambdaConfig = expandCognitoUserPoolLambdaConfig(config)
		}
	}

	if v, ok := d.GetOk("mfa_configuration"); ok {
		params.MfaConfiguration = aws.String(v.(string))
	}

	if v, ok := d.GetOk("password_policy"); ok {
		configs := v.([]interface{})
		config, ok := configs[0].(map[string]interface{})

		if ok && config != nil {
			policies := &cognitoidentityprovider.UserPoolPolicyType{}
			policies.PasswordPolicy = expandCognitoUserPoolPasswordPolicy(config)
			params.Policies = policies
		}
	}

	if v, ok := d.GetOk("schema"); ok {
		configs := v.(*schema.Set).List()
		params.Schema = expandCognitoUserPoolSchema(configs)
	}

	if v, ok := d.GetOk("sms_authentication_message"); ok {
		params.SmsAuthenticationMessage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("sms_configuration"); ok {
		configs := v.([]interface{})
		config, ok := configs[0].(map[string]interface{})

		if ok && config != nil {
			params.SmsConfiguration = expandCognitoUserPoolSmsConfiguration(config)
		}
	}

	if v, ok := d.GetOk("username_attributes"); ok {
		params.UsernameAttributes = expandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("user_pool_add_ons"); ok {
		configs := v.([]interface{})
		config, ok := configs[0].(map[string]interface{})

		if ok {
			userPoolAddons := &cognitoidentityprovider.UserPoolAddOnsType{}

			if v, ok := config["advanced_security_mode"]; ok && v.(string) != "" {
				userPoolAddons.AdvancedSecurityMode = aws.String(v.(string))
			}
			params.UserPoolAddOns = userPoolAddons
		}
	}

	if v, ok := d.GetOk("verification_message_template"); ok {
		configs := v.([]interface{})
		config, ok := configs[0].(map[string]interface{})

		if ok && config != nil {
			params.VerificationMessageTemplate = expandCognitoUserPoolVerificationMessageTemplate(config)
		}
	}

	if v, ok := d.GetOk("sms_verification_message"); ok {
		params.SmsVerificationMessage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tags"); ok {
		params.UserPoolTags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().CognitoidentityproviderTags()
	}
	log.Printf("[DEBUG] Creating Cognito User Pool: %s", params)

	// IAM roles & policies can take some time to propagate and be attached
	// to the User Pool
	var resp *cognitoidentityprovider.CreateUserPoolOutput
	err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		var err error
		resp, err = conn.CreateUserPool(params)
		if isAWSErr(err, cognitoidentityprovider.ErrCodeInvalidSmsRoleTrustRelationshipException, "Role does not have a trust relationship allowing Cognito to assume the role") {
			log.Printf("[DEBUG] Received %s, retrying CreateUserPool", err)
			return resource.RetryableError(err)
		}
		if isAWSErr(err, cognitoidentityprovider.ErrCodeInvalidSmsRoleAccessPolicyException, "Role does not have permission to publish with SNS") {
			log.Printf("[DEBUG] Received %s, retrying CreateUserPool", err)
			return resource.RetryableError(err)
		}

		return resource.NonRetryableError(err)
	})
	if isResourceTimeoutError(err) {
		resp, err = conn.CreateUserPool(params)
	}
	if err != nil {
		return fmt.Errorf("Error creating Cognito User Pool: %s", err)
	}

	d.SetId(*resp.UserPool.Id)

	if v, ok := d.GetOk("risk_configuration"); ok {

		input := expandAwsCognitoUserPoolRiskConfiguration(v.([]interface{}), d.Id())
		_, err = conn.SetRiskConfiguration(input)
		if err != nil {
			return fmt.Errorf("Error setting Cognito User Pool Risk Configuration: %s", err)
		}
	}

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
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == cognitoidentityprovider.ErrCodeResourceNotFoundException {
			log.Printf("[WARN] Cognito User Pool %s is already gone", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	if err := d.Set("admin_create_user_config", flattenCognitoUserPoolAdminCreateUserConfig(resp.UserPool.AdminCreateUserConfig)); err != nil {
		return fmt.Errorf("Failed setting admin_create_user_config: %s", err)
	}
	if resp.UserPool.AliasAttributes != nil {
		d.Set("alias_attributes", flattenStringList(resp.UserPool.AliasAttributes))
	}
	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Service:   "cognito-idp",
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("userpool/%s", d.Id()),
	}
	d.Set("arn", arn.String())
	d.Set("endpoint", fmt.Sprintf("cognito-idp.%s.amazonaws.com/%s", meta.(*AWSClient).region, d.Id()))
	d.Set("auto_verified_attributes", flattenStringList(resp.UserPool.AutoVerifiedAttributes))

	if resp.UserPool.EmailVerificationSubject != nil {
		d.Set("email_verification_subject", resp.UserPool.EmailVerificationSubject)
	}
	if resp.UserPool.EmailVerificationMessage != nil {
		d.Set("email_verification_message", resp.UserPool.EmailVerificationMessage)
	}
	if err := d.Set("lambda_config", flattenCognitoUserPoolLambdaConfig(resp.UserPool.LambdaConfig)); err != nil {
		return fmt.Errorf("Failed setting lambda_config: %s", err)
	}
	if resp.UserPool.MfaConfiguration != nil {
		d.Set("mfa_configuration", resp.UserPool.MfaConfiguration)
	}
	if resp.UserPool.SmsVerificationMessage != nil {
		d.Set("sms_verification_message", resp.UserPool.SmsVerificationMessage)
	}
	if resp.UserPool.SmsAuthenticationMessage != nil {
		d.Set("sms_authentication_message", resp.UserPool.SmsAuthenticationMessage)
	}

	if err := d.Set("device_configuration", flattenCognitoUserPoolDeviceConfiguration(resp.UserPool.DeviceConfiguration)); err != nil {
		return fmt.Errorf("Failed setting device_configuration: %s", err)
	}

	if resp.UserPool.EmailConfiguration != nil {
		if err := d.Set("email_configuration", flattenCognitoUserPoolEmailConfiguration(resp.UserPool.EmailConfiguration)); err != nil {
			return fmt.Errorf("Failed setting email_configuration: %s", err)
		}
	}

	if resp.UserPool.Policies != nil && resp.UserPool.Policies.PasswordPolicy != nil {
		if err := d.Set("password_policy", flattenCognitoUserPoolPasswordPolicy(resp.UserPool.Policies.PasswordPolicy)); err != nil {
			return fmt.Errorf("Failed setting password_policy: %s", err)
		}
	}

	var configuredSchema []interface{}
	if v, ok := d.GetOk("schema"); ok {
		configuredSchema = v.(*schema.Set).List()
	}
	if err := d.Set("schema", flattenCognitoUserPoolSchema(expandCognitoUserPoolSchema(configuredSchema), resp.UserPool.SchemaAttributes)); err != nil {
		return fmt.Errorf("Failed setting schema: %s", err)
	}

	if err := d.Set("sms_configuration", flattenCognitoUserPoolSmsConfiguration(resp.UserPool.SmsConfiguration)); err != nil {
		return fmt.Errorf("Failed setting sms_configuration: %s", err)
	}

	if resp.UserPool.UsernameAttributes != nil {
		d.Set("username_attributes", flattenStringList(resp.UserPool.UsernameAttributes))
	}

	if err := d.Set("user_pool_add_ons", flattenCognitoUserPoolUserPoolAddOns(resp.UserPool.UserPoolAddOns)); err != nil {
		return fmt.Errorf("Failed setting user_pool_add_ons: %s", err)
	}

	if err := d.Set("verification_message_template", flattenCognitoUserPoolVerificationMessageTemplate(resp.UserPool.VerificationMessageTemplate)); err != nil {
		return fmt.Errorf("Failed setting verification_message_template: %s", err)
	}

	d.Set("creation_date", resp.UserPool.CreationDate.Format(time.RFC3339))
	d.Set("last_modified_date", resp.UserPool.LastModifiedDate.Format(time.RFC3339))
	d.Set("name", resp.UserPool.Name)
	if err := d.Set("tags", keyvaluetags.CognitoidentityKeyValueTags(resp.UserPool.UserPoolTags).IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	if err := d.Set("risk_configuration", flattenAwsCognitoUserPoolRiskConfiguration(conn, d.Id())); err != nil {
		return fmt.Errorf("error setting User Pool Risk Config: %s", err)
	}

	return nil
}

func resourceAwsCognitoUserPoolUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	params := &cognitoidentityprovider.UpdateUserPoolInput{
		UserPoolId: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("admin_create_user_config"); ok {
		configs := v.([]interface{})
		config, ok := configs[0].(map[string]interface{})

		if ok && config != nil {
			params.AdminCreateUserConfig = expandCognitoUserPoolAdminCreateUserConfig(config)
		}
	}

	if v, ok := d.GetOk("auto_verified_attributes"); ok {
		params.AutoVerifiedAttributes = expandStringList(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("device_configuration"); ok {
		configs := v.([]interface{})
		config, ok := configs[0].(map[string]interface{})

		if ok && config != nil {
			params.DeviceConfiguration = expandCognitoUserPoolDeviceConfiguration(config)
		}
	}

	if v, ok := d.GetOk("email_configuration"); ok {

		configs := v.([]interface{})
		config, ok := configs[0].(map[string]interface{})

		if ok && config != nil {
			log.Printf("[DEBUG] Set Values to update from configs")
			emailConfigurationType := &cognitoidentityprovider.EmailConfigurationType{}

			if v, ok := config["reply_to_email_address"]; ok && v.(string) != "" {
				emailConfigurationType.ReplyToEmailAddress = aws.String(v.(string))
			}

			if v, ok := config["source_arn"]; ok && v.(string) != "" {
				emailConfigurationType.SourceArn = aws.String(v.(string))
			}

			if v, ok := config["email_sending_account"]; ok && v.(string) != "" {
				emailConfigurationType.EmailSendingAccount = aws.String(v.(string))
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

	if v, ok := d.GetOk("lambda_config"); ok {
		configs := v.([]interface{})
		config, ok := configs[0].(map[string]interface{})

		if ok && config != nil {
			params.LambdaConfig = expandCognitoUserPoolLambdaConfig(config)
		}
	}

	if v, ok := d.GetOk("mfa_configuration"); ok {
		params.MfaConfiguration = aws.String(v.(string))
	}

	if v, ok := d.GetOk("password_policy"); ok {
		configs := v.([]interface{})
		config, ok := configs[0].(map[string]interface{})

		if ok && config != nil {
			policies := &cognitoidentityprovider.UserPoolPolicyType{}
			policies.PasswordPolicy = expandCognitoUserPoolPasswordPolicy(config)
			params.Policies = policies
		}
	}

	if v, ok := d.GetOk("sms_authentication_message"); ok {
		params.SmsAuthenticationMessage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("sms_configuration"); ok {
		configs := v.([]interface{})
		config, ok := configs[0].(map[string]interface{})

		if ok && config != nil {
			params.SmsConfiguration = expandCognitoUserPoolSmsConfiguration(config)
		}
	}

	if v, ok := d.GetOk("user_pool_add_ons"); ok {
		configs := v.([]interface{})
		config, ok := configs[0].(map[string]interface{})

		if ok && config != nil {
			userPoolAddons := &cognitoidentityprovider.UserPoolAddOnsType{}

			if v, ok := config["advanced_security_mode"]; ok && v.(string) != "" {
				userPoolAddons.AdvancedSecurityMode = aws.String(v.(string))
			}
			params.UserPoolAddOns = userPoolAddons
		}
	}

	if v, ok := d.GetOk("verification_message_template"); ok {
		configs := v.([]interface{})
		config, ok := configs[0].(map[string]interface{})

		if d.HasChange("email_verification_message") {
			config["email_message"] = d.Get("email_verification_message")
		}
		if d.HasChange("email_verification_subject") {
			config["email_subject"] = d.Get("email_verification_subject")
		}
		if d.HasChange("sms_verification_message") {
			config["sms_message"] = d.Get("sms_verification_message")
		}

		if ok && config != nil {
			params.VerificationMessageTemplate = expandCognitoUserPoolVerificationMessageTemplate(config)
		}
	}

	if v, ok := d.GetOk("sms_verification_message"); ok {
		params.SmsVerificationMessage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tags"); ok {
		params.UserPoolTags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().CognitoidentityproviderTags()
	}

	log.Printf("[DEBUG] Updating Cognito User Pool: %s", params)

	// IAM roles & policies can take some time to propagate and be attached
	// to the User Pool.
	err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		var err error
		_, err = conn.UpdateUserPool(params)
		if isAWSErr(err, cognitoidentityprovider.ErrCodeInvalidSmsRoleTrustRelationshipException, "Role does not have a trust relationship allowing Cognito to assume the role") {
			log.Printf("[DEBUG] Received %s, retrying UpdateUserPool", err)
			return resource.RetryableError(err)
		}
		if isAWSErr(err, cognitoidentityprovider.ErrCodeInvalidSmsRoleAccessPolicyException, "Role does not have permission to publish with SNS") {
			log.Printf("[DEBUG] Received %s, retrying UpdateUserPool", err)
			return resource.RetryableError(err)
		}
		if isAWSErr(err, cognitoidentityprovider.ErrCodeInvalidParameterException, "Please use TemporaryPasswordValidityDays in PasswordPolicy instead of UnusedAccountValidityDays") {
			log.Printf("[DEBUG] Received %s, retrying UpdateUserPool without UnusedAccountValidityDays", err)
			params.AdminCreateUserConfig.UnusedAccountValidityDays = nil
			return resource.RetryableError(err)
		}

		return resource.NonRetryableError(err)
	})
	if isResourceTimeoutError(err) {
		_, err = conn.UpdateUserPool(params)
	}
	if err != nil {
		return fmt.Errorf("Error updating Cognito User pool: %s", err)
	}

	if v, ok := d.GetOk("risk_configuration"); ok {

		input := expandAwsCognitoUserPoolRiskConfiguration(v.([]interface{}), d.Id())
		_, err = conn.SetRiskConfiguration(input)
		if err != nil {
			return fmt.Errorf("Error setting Cognito User Pool Risk Configuration: %s", err)
		}
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
		return fmt.Errorf("Error deleting user pool: %s", err)
	}

	return nil
}

func expandAwsCognitoUserPoolRiskConfiguration(v []interface{}, userPoolId string) *cognitoidentityprovider.SetRiskConfigurationInput {
	config := &cognitoidentityprovider.SetRiskConfigurationInput{
		UserPoolId: aws.String(userPoolId),
	}

	if len(v) == 0 || v[0] == nil {
		return config
	}
	mConfig := v[0].(map[string]interface{})

	if v, ok := mConfig["account_takeover_risk_configuration"]; ok {
		config.AccountTakeoverRiskConfiguration = expandAwsCognitoUserPoolAccountTakeoverConfiguration(v.([]interface{}))
	}

	if v, ok := mConfig["risk_exception_configuration"]; ok {
		config.RiskExceptionConfiguration = expandAwsCognitoUserPoolRiskExceptionConfiguration(v.([]interface{}))
	}

	if v, ok := mConfig["compromised_credentials_risk_configuration"]; ok {
		config.CompromisedCredentialsRiskConfiguration = expandAwsCognitoUserPoolCompromisedCredentialsRiskConfiguration(v.([]interface{}))
	}

	return config
}

func flattenAwsCognitoUserPoolRiskConfiguration(conn *cognitoidentityprovider.CognitoIdentityProvider, userPoolId string) []map[string]interface{} {
	input := &cognitoidentityprovider.DescribeRiskConfigurationInput{
		UserPoolId: aws.String(userPoolId),
	}

	out, err := conn.DescribeRiskConfiguration(input)
	if err == nil {
		return nil
	}
	result := make([]map[string]interface{}, 0)
	if out == nil {
		return result
	}

	riskConfig := out.RiskConfiguration
	item := make(map[string]interface{})
	item["account_takeover_risk_configuration"] = flattenAwsCognitoUserPoolAccountTakeoverConfiguration(riskConfig.AccountTakeoverRiskConfiguration)
	item["risk_exception_configuration"] = flattenAwsCognitoUserPoolRiskExceptionConfiguration(riskConfig.RiskExceptionConfiguration)
	item["compromised_credentials_risk_configuration"] = flattenAwsCognitoUserPoolCompromisedCredentialsRiskConfiguration(riskConfig.CompromisedCredentialsRiskConfiguration)

	return append(result, item)
}

func expandAwsCognitoUserPoolRiskExceptionConfiguration(v []interface{}) *cognitoidentityprovider.RiskExceptionConfigurationType {
	config := &cognitoidentityprovider.RiskExceptionConfigurationType{}

	if len(v) == 0 || v[0] == nil {
		return config
	}
	mConfig := v[0].(map[string]interface{})

	if v, ok := mConfig["block_ip_range_list"]; ok && len(v.(*schema.Set).List()) > 0 {
		config.BlockedIPRangeList = expandStringSet(v.(*schema.Set))
	}

	if v, ok := mConfig["skipped_ip_range_list"]; ok && len(v.(*schema.Set).List()) > 0 {
		config.SkippedIPRangeList = expandStringSet(v.(*schema.Set))
	}

	return config
}

func flattenAwsCognitoUserPoolRiskExceptionConfiguration(riskConfig *cognitoidentityprovider.RiskExceptionConfigurationType) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	if riskConfig == nil {
		return result
	}

	item := make(map[string]interface{})
	item["blocked_ip_range_list"] = flattenStringSet(riskConfig.BlockedIPRangeList)
	item["skipped_ip_range_list"] = flattenStringSet(riskConfig.SkippedIPRangeList)

	return append(result, item)
}

func expandAwsCognitoUserPoolAccountTakeoverConfiguration(v []interface{}) *cognitoidentityprovider.AccountTakeoverRiskConfigurationType {
	config := &cognitoidentityprovider.AccountTakeoverRiskConfigurationType{}

	if len(v) == 0 || v[0] == nil {
		return config
	}
	mConfig := v[0].(map[string]interface{})

	if v, ok := mConfig["actions"]; ok {
		config.Actions = expandAwsCognitoUserPoolAccountTakeoverConfigurationActions(v.([]interface{}))
	}

	if v, ok := mConfig["notify_configuration"]; ok {
		config.NotifyConfiguration = expandAwsCognitoUserPoolAccountTakeoverNotificationConfiguration(v.([]interface{}))
	}

	return config
}

func flattenAwsCognitoUserPoolAccountTakeoverConfiguration(accountTakeoverConf *cognitoidentityprovider.AccountTakeoverRiskConfigurationType) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	if accountTakeoverConf == nil {
		return result
	}

	item := make(map[string]interface{})
	item["actions"] = flattenAwsCognitoUserPoolAccountTakeoverConfigurationActions(accountTakeoverConf.Actions)
	item["notify_configuration"] = flattendAwsCognitoUserPoolAccountTakeoverNotificationConfiguration(accountTakeoverConf.NotifyConfiguration)

	return append(result, item)
}

func expandAwsCognitoUserPoolAccountTakeoverConfigurationActions(v []interface{}) *cognitoidentityprovider.AccountTakeoverActionsType {
	config := &cognitoidentityprovider.AccountTakeoverActionsType{}

	if len(v) == 0 || v[0] == nil {
		return config
	}
	mConfig := v[0].(map[string]interface{})

	if v, ok := mConfig["high_action"]; ok {
		config.HighAction = expandAwsCognitoUserPoolAccountTakeoverConfigurationAction(v.([]interface{}))
	}

	if v, ok := mConfig["low_action"]; ok {
		config.LowAction = expandAwsCognitoUserPoolAccountTakeoverConfigurationAction(v.([]interface{}))
	}

	if v, ok := mConfig["medium_action"]; ok {
		config.MediumAction = expandAwsCognitoUserPoolAccountTakeoverConfigurationAction(v.([]interface{}))
	}

	return config
}

func flattenAwsCognitoUserPoolAccountTakeoverConfigurationActions(actions *cognitoidentityprovider.AccountTakeoverActionsType) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	if actions == nil {
		return result
	}

	item := make(map[string]interface{})
	item["high_action"] = flattenAwsCognitoUserPoolAccountTakeoverConfigurationAction(actions.HighAction)
	item["low_action"] = flattenAwsCognitoUserPoolAccountTakeoverConfigurationAction(actions.LowAction)
	item["medium_action"] = flattenAwsCognitoUserPoolAccountTakeoverConfigurationAction(actions.MediumAction)

	return append(result, item)
}

func expandAwsCognitoUserPoolAccountTakeoverConfigurationAction(v []interface{}) *cognitoidentityprovider.AccountTakeoverActionType {
	config := &cognitoidentityprovider.AccountTakeoverActionType{}

	if len(v) == 0 || v[0] == nil {
		return config
	}
	mConfig := v[0].(map[string]interface{})

	if v, ok := mConfig["event_action"]; ok {
		config.EventAction = aws.String(v.(string))
	}

	if v, ok := mConfig["notify"]; ok {
		config.Notify = aws.Bool(v.(bool))
	}

	return config
}

func expandAwsCognitoUserPoolAccountTakeoverNotificationConfiguration(v []interface{}) *cognitoidentityprovider.NotifyConfigurationType {
	config := &cognitoidentityprovider.NotifyConfigurationType{}

	if len(v) == 0 || v[0] == nil {
		return config
	}
	mConfig := v[0].(map[string]interface{})

	if v, ok := mConfig["block_email"]; ok {
		config.BlockEmail = expandAwsCognitoUserPoolAccountTakeoverNotificationEmailConfiguration(v.([]interface{}))
	}

	if v, ok := mConfig["from"]; ok {
		config.From = aws.String(v.(string))
	}

	if v, ok := mConfig["event_action"]; ok {
		config.MfaEmail = expandAwsCognitoUserPoolAccountTakeoverNotificationEmailConfiguration(v.([]interface{}))
	}

	if v, ok := mConfig["event_action"]; ok {
		config.NoActionEmail = expandAwsCognitoUserPoolAccountTakeoverNotificationEmailConfiguration(v.([]interface{}))
	}

	if v, ok := mConfig["reply_to"]; ok {
		config.ReplyTo = aws.String(v.(string))
	}

	if v, ok := mConfig["source_arn"]; ok {
		config.SourceArn = aws.String(v.(string))
	}

	return config
}

func flattenAwsCognitoUserPoolAccountTakeoverConfigurationAction(action *cognitoidentityprovider.AccountTakeoverActionType) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	if action == nil {
		return result
	}

	item := make(map[string]interface{})
	item["event_action"] = aws.StringValue(action.EventAction)
	item["notify"] = aws.BoolValue(action.Notify)

	return append(result, item)
}

func flattendAwsCognitoUserPoolAccountTakeoverNotificationConfiguration(notifConfig *cognitoidentityprovider.NotifyConfigurationType) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	if notifConfig == nil {
		return result
	}

	item := make(map[string]interface{})
	item["block_email"] = flattenAwsCognitoUserPoolAccountTakeoverNotificationEmailConfiguration(notifConfig.BlockEmail)
	item["from"] = aws.StringValue(notifConfig.From)
	item["mfa_email"] = flattenAwsCognitoUserPoolAccountTakeoverNotificationEmailConfiguration(notifConfig.MfaEmail)
	item["no_action_email"] = flattenAwsCognitoUserPoolAccountTakeoverNotificationEmailConfiguration(notifConfig.NoActionEmail)
	item["reply_to"] = aws.StringValue(notifConfig.ReplyTo)
	item["source_arn"] = aws.StringValue(notifConfig.SourceArn)

	return append(result, item)
}

func expandAwsCognitoUserPoolAccountTakeoverNotificationEmailConfiguration(v []interface{}) *cognitoidentityprovider.NotifyEmailType {
	config := &cognitoidentityprovider.NotifyEmailType{}

	if len(v) == 0 || v[0] == nil {

		return config
	}
	mConfig := v[0].(map[string]interface{})

	if v, ok := mConfig["html_body"]; ok {
		config.HtmlBody = aws.String(v.(string))
	}

	if v, ok := mConfig["subject"]; ok {
		config.Subject = aws.String(v.(string))
	}

	if v, ok := mConfig["text_body"]; ok {
		config.TextBody = aws.String(v.(string))
	}

	return config
}

func flattenAwsCognitoUserPoolAccountTakeoverNotificationEmailConfiguration(notifyConfig *cognitoidentityprovider.NotifyEmailType) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	if notifyConfig == nil {
		return result
	}

	item := make(map[string]interface{})
	item["html_body"] = aws.StringValue(notifyConfig.HtmlBody)
	item["subject"] = aws.StringValue(notifyConfig.Subject)
	item["text_body"] = aws.StringValue(notifyConfig.TextBody)

	return append(result, item)
}

func expandAwsCognitoUserPoolCompromisedCredentialsRiskConfiguration(v []interface{}) *cognitoidentityprovider.CompromisedCredentialsRiskConfigurationType {
	config := &cognitoidentityprovider.CompromisedCredentialsRiskConfigurationType{}

	if len(v) == 0 || v[0] == nil {
		return config
	}
	mConfig := v[0].(map[string]interface{})

	if v, ok := mConfig["event_filter"]; ok && len(v.(*schema.Set).List()) > 0 {
		config.EventFilter = expandStringSet(v.(*schema.Set))
	}

	if v, ok := mConfig["actions"]; ok {
		mConfig := v.([]interface{})
		if len(mConfig) == 0 || mConfig[0] == nil {
			return nil
		}

		actions := mConfig[0].(map[string]interface{})
		compromisedActions := &cognitoidentityprovider.CompromisedCredentialsActionsType{
			EventAction: aws.String(actions["event_action"].(string)),
		}
		config.Actions = compromisedActions
	}

	return config
}

func flattenAwsCognitoUserPoolCompromisedCredentialsRiskConfiguration(comp *cognitoidentityprovider.CompromisedCredentialsRiskConfigurationType) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	if comp == nil {
		return result
	}

	item := make(map[string]interface{})
	item["event_filter"] = flattenStringSet(comp.EventFilter)

	actions := make(map[string]interface{})
	actions["event_action"] = aws.StringValue(comp.Actions.EventAction)

	item["actions"] = actions

	return append(result, item)
}
