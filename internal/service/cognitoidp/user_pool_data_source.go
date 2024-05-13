// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_cognito_user_group", name="User Pool")
func newDataSourceUserPool(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceUserPool{}, nil
}

const (
	DSNameUserPool = "User Pool Data Source"
)

type dataSourceUserPool struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceUserPool) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_cognito_user_pool"
}

func (d *dataSourceUserPool) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				Computed: true,
			},
			"auto_verified_attributes": schema.ListAttribute{
				Computed:    true,
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
			},
			names.AttrCreationDate: schema.StringAttribute{
				Computed: true,
			},
			"custom_domain": schema.StringAttribute{
				Computed: true,
			},
			"deletion_protection": schema.StringAttribute{
				Computed: true,
			},
			names.AttrDomain: schema.StringAttribute{
				Computed: true,
			},
			"estimated_number_of_users": schema.Int64Attribute{
				Computed: true,
			},
			names.AttrID: schema.StringAttribute{
				Required: true,
			},
			"last_modified_date": schema.StringAttribute{
				Computed: true,
			},
			"mfa_configuration": schema.StringAttribute{
				Computed: true,
			},
			names.AttrName: schema.StringAttribute{
				Computed: true,
			},
			"sms_authentication_message": schema.StringAttribute{
				Computed: true,
			},
			"sms_configuration_failure": schema.StringAttribute{
				Computed: true,
			},
			"sms_verification_message": schema.StringAttribute{
				Computed: true,
			},
			"user_pool_tags": tftags.TagsAttributeComputedOnly(),
			"username_attributes": schema.ListAttribute{
				Computed:    true,
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
			},
		},
		Blocks: map[string]schema.Block{
			"account_recovery_setting": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[accountRecoverySettingType](ctx),
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"recovery_mechanism": schema.ListNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrName: schema.StringAttribute{
										Computed: true,
									},
									names.AttrPriority: schema.Int64Attribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"admin_create_user_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[adminCreateUserConfigType](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"allow_admin_create_user_only": schema.BoolAttribute{
							Computed: true,
						},
						"unused_account_validity_days": schema.Int64Attribute{
							Computed: true,
						},
					},
					Blocks: map[string]schema.Block{
						"invite_message_template": schema.ListNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"email_message": schema.StringAttribute{
										Computed: true,
									},
									"email_subject": schema.StringAttribute{
										Computed: true,
									},
									"sms_message": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"device_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[deviceConfigurationType](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"challenge_required_on_new_device": schema.BoolAttribute{
							Computed: true,
						},
						"device_only_remembered_on_user_prompt": schema.BoolAttribute{
							Computed: true,
						},
					},
				},
			},
			"email_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[emailConfigurationType](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"configuration_set": schema.StringAttribute{
							Computed: true,
						},
						"email_sending_account": schema.StringAttribute{
							Computed: true,
						},
						"from": schema.StringAttribute{
							Computed: true,
						},
						"reply_to_email_address": schema.StringAttribute{
							Computed: true,
						},
						"source_arn": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
			"lambda_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[lambdaConfigType](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"create_auth_challenge": schema.StringAttribute{
							Computed: true,
						},
						"custom_message": schema.StringAttribute{
							Computed: true,
						},
						"define_auth_challenge": schema.StringAttribute{
							Computed: true,
						},
						names.AttrKMSKeyID: schema.StringAttribute{
							Computed: true,
						},
						"post_authentication": schema.StringAttribute{
							Computed: true,
						},
						"post_confirmation": schema.StringAttribute{
							Computed: true,
						},
						"pre_authentication": schema.StringAttribute{
							Computed: true,
						},
						"pre_sign_up": schema.StringAttribute{
							Computed: true,
						},
						"pre_token_generation": schema.StringAttribute{
							Computed: true,
						},
						"user_migration": schema.StringAttribute{
							Computed: true,
						},
						"verify_auth_challenge_response": schema.StringAttribute{
							Computed: true,
						},
					},
					Blocks: map[string]schema.Block{
						"custom_email_sender": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[customEmailSenderType](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"lambda_arn": schema.StringAttribute{
										Computed: true,
									},
									"lambda_version": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
						"custom_sms_sender": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[customSMSSenderType](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"lambda_arn": schema.StringAttribute{
										Computed: true,
									},
									"lambda_version": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
						"pre_token_generation_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[preTokenGenerationConfigType](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"lambda_arn": schema.StringAttribute{
										Computed: true,
									},
									"lambda_version": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"schema_attributes": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[schemaAttributeType](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"attribute_data_type": schema.StringAttribute{
							Computed: true,
						},
						"developer_only_attribute": schema.BoolAttribute{
							Computed: true,
						},
						"mutable": schema.BoolAttribute{
							Computed: true,
						},
						names.AttrName: schema.StringAttribute{
							Computed: true,
						},
						"required": schema.BoolAttribute{
							Computed: true,
						},
					},
					Blocks: map[string]schema.Block{
						"number_attribute_constraints": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[numberAttributeConstraintsType](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"max_value": schema.StringAttribute{
										Computed: true,
									},
									"min_value": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
						"string_attribute_constraints": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[stringAttributeConstraintsType](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"max_length": schema.StringAttribute{
										Computed: true,
									},
									"min_length": schema.StringAttribute{
										Computed: true,
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

func (d *dataSourceUserPool) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().CognitoIDPConn(ctx)

	var data dataSourceUserPoolData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findUserPoolByID(ctx, conn, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CognitoIDP, create.ErrActionReading, DSNameUserPool, data.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceUserPoolData struct {
	AccountRecoverySetting   fwtypes.ListNestedObjectValueOf[accountRecoverySettingType] `tfsdk:"account_recovery_setting"`
	AdminCreateUserConfig    fwtypes.ListNestedObjectValueOf[adminCreateUserConfigType]  `tfsdk:"admin_create_user_config"`
	Arn                      types.String                                                `tfsdk:"arn"`
	AutoVerifiedAttributes   fwtypes.ListValueOf[types.String]                           `tfsdk:"auto_verified_attributes"`
	CreationDate             types.String                                                `tfsdk:"creation_date"`
	CustomDomain             types.String                                                `tfsdk:"custom_domain"`
	DeletionProtection       types.String                                                `tfsdk:"deletion_protection"`
	DeviceConfiguration      fwtypes.ListNestedObjectValueOf[deviceConfigurationType]    `tfsdk:"device_configuration"`
	Domain                   types.String                                                `tfsdk:"domain"`
	EmailConfiguration       fwtypes.ListNestedObjectValueOf[emailConfigurationType]     `tfsdk:"email_configuration"`
	EstimatedNumberOfUsers   types.Int64                                                 `tfsdk:"estimated_number_of_users"`
	ID                       types.String                                                `tfsdk:"id"`
	LambdaConfig             fwtypes.ListNestedObjectValueOf[lambdaConfigType]           `tfsdk:"lambda_config"`
	LastModifiedDate         types.String                                                `tfsdk:"last_modified_date"`
	MfaConfiguration         types.String                                                `tfsdk:"mfa_configuration"`
	SchemaAttributes         fwtypes.ListNestedObjectValueOf[schemaAttributeType]        `tfsdk:"schema_attributes"`
	Name                     types.String                                                `tfsdk:"name"`
	SmsAuthenticationMessage types.String                                                `tfsdk:"sms_authentication_message"`
	SmsConfigurationFailure  types.String                                                `tfsdk:"sms_configuration_failure"`
	SmsVerificationMessage   types.String                                                `tfsdk:"sms_verification_message"`
	UserPoolTags             types.Map                                                   `tfsdk:"user_pool_tags"`
	UsernameAttributes       fwtypes.ListValueOf[types.String]                           `tfsdk:"username_attributes"`
}

type accountRecoverySettingType struct {
	RecoveryMechanism fwtypes.ListNestedObjectValueOf[recoveryMechanismType] `tfsdk:"recovery_mechanism"`
}

type adminCreateUserConfigType struct {
	AllowAdminCreateUserOnly  types.Bool                                                 `tfsdk:"allow_admin_create_user_only"`
	InviteMessageTemplate     fwtypes.ListNestedObjectValueOf[inviteMessageTemplateType] `tfsdk:"invite_message_template"`
	UnusedAccountValidityDays types.Int64                                                `tfsdk:"unused_account_validity_days"`
}

type inviteMessageTemplateType struct {
	EmailMessage types.String `tfsdk:"email_message"`
	EmailSubject types.String `tfsdk:"email_subject"`
	SmsMessage   types.String `tfsdk:"sms_message"`
}

type deviceConfigurationType struct {
	ChallengeRequiredOnNewDevice     types.Bool `tfsdk:"challenge_required_on_new_device"`
	DeviceOnlyRememberedOnUserPrompt types.Bool `tfsdk:"device_only_remembered_on_user_prompt"`
}

type emailConfigurationType struct {
	ConfigurationSet    types.String `tfsdk:"configuration_set"`
	EmailSendingAccount types.String `tfsdk:"email_sending_account"`
	From                types.String `tfsdk:"from"`
	ReplyToEmailAddress types.String `tfsdk:"reply_to_email_address"`
	SourceArn           types.String `tfsdk:"source_arn"`
}

type lambdaConfigType struct {
	CreateAuthChallenge         types.String                                                  `tfsdk:"create_auth_challenge"`
	CustomEmailSender           fwtypes.ListNestedObjectValueOf[customEmailSenderType]        `tfsdk:"custom_email_sender"`
	CustomMessage               types.String                                                  `tfsdk:"custom_message"`
	CustomSMSSender             fwtypes.ListNestedObjectValueOf[customSMSSenderType]          `tfsdk:"custom_sms_sender"`
	DefineAuthChallenge         types.String                                                  `tfsdk:"define_auth_challenge"`
	KmsKeyId                    types.String                                                  `tfsdk:"kms_key_id"`
	PostAuthentication          types.String                                                  `tfsdk:"post_authentication"`
	PostConfirmation            types.String                                                  `tfsdk:"post_confirmation"`
	PreAuthentication           types.String                                                  `tfsdk:"pre_authentication"`
	PreSignUp                   types.String                                                  `tfsdk:"pre_sign_up"`
	PreTokenGeneration          types.String                                                  `tfsdk:"pre_token_generation"`
	PreTokenGenerationConfig    fwtypes.ListNestedObjectValueOf[preTokenGenerationConfigType] `tfsdk:"pre_token_generation_config"`
	UserMigration               types.String                                                  `tfsdk:"user_migration"`
	VerifyAuthChallengeResponse types.String                                                  `tfsdk:"verify_auth_challenge_response"`
}

type customEmailSenderType struct {
	LambdaArn     types.String `tfsdk:"lambda_arn"`
	LambdaVersion types.String `tfsdk:"lambda_version"`
}

type customSMSSenderType struct {
	LambdaArn     types.String `tfsdk:"lambda_arn"`
	LambdaVersion types.String `tfsdk:"lambda_version"`
}

type preTokenGenerationConfigType struct {
	LambdaArn     types.String `tfsdk:"lambda_arn"`
	LambdaVersion types.String `tfsdk:"lambda_version"`
}

type recoveryMechanismType struct {
	Name     types.String `tfsdk:"name"`
	Priority types.Int64  `tfsdk:"priority"`
}

type schemaAttributeType struct {
	AttributeDataType          types.String                                                    `tfsdk:"attribute_data_type"`
	DeveloperOnlyAttribute     types.Bool                                                      `tfsdk:"developer_only_attribute"`
	Mutable                    types.Bool                                                      `tfsdk:"mutable"`
	Name                       types.String                                                    `tfsdk:"name"`
	NumberAttributeConstraints fwtypes.ListNestedObjectValueOf[numberAttributeConstraintsType] `tfsdk:"number_attribute_constraints"`
	Required                   types.Bool                                                      `tfsdk:"required"`
	StringAttributeConstraints fwtypes.ListNestedObjectValueOf[stringAttributeConstraintsType] `tfsdk:"string_attribute_constraints"`
}

type numberAttributeConstraintsType struct {
	MaxValue types.String `tfsdk:"max_value"`
	MinValue types.String `tfsdk:"min_value"`
}

type stringAttributeConstraintsType struct {
	MaxLength types.String `tfsdk:"max_length"`
	MinLength types.String `tfsdk:"min_length"`
}
