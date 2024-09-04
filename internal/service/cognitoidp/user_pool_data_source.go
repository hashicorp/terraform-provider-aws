// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_cognito_user_group", name="User Pool")
func newUserPoolDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &userPoolDataSource{}, nil
}

type userPoolDataSource struct {
	framework.DataSourceWithConfigure
}

func (*userPoolDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	response.TypeName = "aws_cognito_user_pool"
}

func (d *userPoolDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"account_recovery_setting": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[accountRecoverySettingTypeModel](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: fwtypes.AttributeTypesMust[accountRecoverySettingTypeModel](ctx),
				},
			},
			"admin_create_user_config": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[adminCreateUserConfigTypeModel](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: fwtypes.AttributeTypesMust[adminCreateUserConfigTypeModel](ctx),
				},
			},
			names.AttrARN: schema.StringAttribute{
				Computed: true,
			},
			"auto_verified_attributes": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
			names.AttrCreationDate: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"custom_domain": schema.StringAttribute{
				Computed: true,
			},
			names.AttrDeletionProtection: schema.StringAttribute{
				Computed: true,
			},
			"device_configuration": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[deviceConfigurationTypeModel](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: fwtypes.AttributeTypesMust[deviceConfigurationTypeModel](ctx),
				},
			},
			names.AttrDomain: schema.StringAttribute{
				Computed: true,
			},
			"email_configuration": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[emailConfigurationTypeModel](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: fwtypes.AttributeTypesMust[emailConfigurationTypeModel](ctx),
				},
			},
			"estimated_number_of_users": schema.Int64Attribute{
				Computed: true,
			},
			names.AttrID: framework.IDAttribute(),
			"lambda_config": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[lambdaConfigTypeModel](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: fwtypes.AttributeTypesMust[lambdaConfigTypeModel](ctx),
				},
			},
			"last_modified_date": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"mfa_configuration": schema.StringAttribute{
				Computed: true,
			},
			names.AttrName: schema.StringAttribute{
				Computed: true,
			},
			"schema_attributes": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[schemaAttributeTypeModel](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: fwtypes.AttributeTypesMust[schemaAttributeTypeModel](ctx),
				},
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
			names.AttrUserPoolID: schema.StringAttribute{
				Required: true,
			},
			"user_pool_tags": tftags.TagsAttributeComputedOnly(),
			"username_attributes": schema.ListAttribute{
				Computed:    true,
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *userPoolDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data userPoolDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().CognitoIDPClient(ctx)

	userPoolID := data.UserPoolID.ValueString()
	output, err := findUserPoolByID(ctx, conn, userPoolID)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Cognito User Pool (%s)", userPoolID), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.ID = fwflex.StringValueToFramework(ctx, userPoolID)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type userPoolDataSourceModel struct {
	AccountRecoverySetting   fwtypes.ListNestedObjectValueOf[accountRecoverySettingTypeModel] `tfsdk:"account_recovery_setting"`
	AdminCreateUserConfig    fwtypes.ListNestedObjectValueOf[adminCreateUserConfigTypeModel]  `tfsdk:"admin_create_user_config"`
	ARN                      types.String                                                     `tfsdk:"arn"`
	AutoVerifiedAttributes   fwtypes.ListValueOf[types.String]                                `tfsdk:"auto_verified_attributes"`
	CreationDate             timetypes.RFC3339                                                `tfsdk:"creation_date"`
	CustomDomain             types.String                                                     `tfsdk:"custom_domain"`
	DeletionProtection       types.String                                                     `tfsdk:"deletion_protection"`
	DeviceConfiguration      fwtypes.ListNestedObjectValueOf[deviceConfigurationTypeModel]    `tfsdk:"device_configuration"`
	Domain                   types.String                                                     `tfsdk:"domain"`
	EmailConfiguration       fwtypes.ListNestedObjectValueOf[emailConfigurationTypeModel]     `tfsdk:"email_configuration"`
	EstimatedNumberOfUsers   types.Int64                                                      `tfsdk:"estimated_number_of_users"`
	ID                       types.String                                                     `tfsdk:"id"`
	LambdaConfig             fwtypes.ListNestedObjectValueOf[lambdaConfigTypeModel]           `tfsdk:"lambda_config"`
	LastModifiedDate         timetypes.RFC3339                                                `tfsdk:"last_modified_date"`
	MFAConfiguration         types.String                                                     `tfsdk:"mfa_configuration"`
	Name                     types.String                                                     `tfsdk:"name"`
	SchemaAttributes         fwtypes.ListNestedObjectValueOf[schemaAttributeTypeModel]        `tfsdk:"schema_attributes"`
	SMSAuthenticationMessage types.String                                                     `tfsdk:"sms_authentication_message"`
	SMSConfigurationFailure  types.String                                                     `tfsdk:"sms_configuration_failure"`
	SMSVerificationMessage   types.String                                                     `tfsdk:"sms_verification_message"`
	UserPoolID               types.String                                                     `tfsdk:"user_pool_id"`
	UserPoolTags             types.Map                                                        `tfsdk:"user_pool_tags"`
	UsernameAttributes       fwtypes.ListValueOf[types.String]                                `tfsdk:"username_attributes"`
}

type accountRecoverySettingTypeModel struct {
	RecoveryMechanism fwtypes.ListNestedObjectValueOf[recoveryOptionTypeModel] `tfsdk:"recovery_mechanism"`
}

type recoveryOptionTypeModel struct {
	Name     types.String `tfsdk:"name"`
	Priority types.Int64  `tfsdk:"priority"`
}

type adminCreateUserConfigTypeModel struct {
	AllowAdminCreateUserOnly  types.Bool                                                `tfsdk:"allow_admin_create_user_only"`
	InviteMessageTemplate     fwtypes.ListNestedObjectValueOf[messageTemplateTypeModel] `tfsdk:"invite_message_template"`
	UnusedAccountValidityDays types.Int64                                               `tfsdk:"unused_account_validity_days"`
}

type messageTemplateTypeModel struct {
	EmailMessage types.String `tfsdk:"email_message"`
	EmailSubject types.String `tfsdk:"email_subject"`
	SMSMessage   types.String `tfsdk:"sms_message"`
}

type deviceConfigurationTypeModel struct {
	ChallengeRequiredOnNewDevice     types.Bool `tfsdk:"challenge_required_on_new_device"`
	DeviceOnlyRememberedOnUserPrompt types.Bool `tfsdk:"device_only_remembered_on_user_prompt"`
}

type emailConfigurationTypeModel struct {
	ConfigurationSet    types.String `tfsdk:"configuration_set"`
	EmailSendingAccount types.String `tfsdk:"email_sending_account"`
	From                types.String `tfsdk:"from"`
	ReplyToEmailAddress types.String `tfsdk:"reply_to_email_address"`
	SourceARN           types.String `tfsdk:"source_arn"`
}

type lambdaConfigTypeModel struct {
	CreateAuthChallenge         types.String                                                              `tfsdk:"create_auth_challenge"`
	CustomEmailSender           fwtypes.ListNestedObjectValueOf[customEmailLambdaVersionConfigTypeModel]  `tfsdk:"custom_email_sender"`
	CustomMessage               types.String                                                              `tfsdk:"custom_message"`
	CustomSMSSender             fwtypes.ListNestedObjectValueOf[customSMSLambdaVersionConfigTypeModel]    `tfsdk:"custom_sms_sender"`
	DefineAuthChallenge         types.String                                                              `tfsdk:"define_auth_challenge"`
	KMSKeyID                    types.String                                                              `tfsdk:"kms_key_id"`
	PostAuthentication          types.String                                                              `tfsdk:"post_authentication"`
	PostConfirmation            types.String                                                              `tfsdk:"post_confirmation"`
	PreAuthentication           types.String                                                              `tfsdk:"pre_authentication"`
	PreSignUp                   types.String                                                              `tfsdk:"pre_sign_up"`
	PreTokenGeneration          types.String                                                              `tfsdk:"pre_token_generation"`
	PreTokenGenerationConfig    fwtypes.ListNestedObjectValueOf[preTokenGenerationVersionConfigTypeModel] `tfsdk:"pre_token_generation_config"`
	UserMigration               types.String                                                              `tfsdk:"user_migration"`
	VerifyAuthChallengeResponse types.String                                                              `tfsdk:"verify_auth_challenge_response"`
}

type customEmailLambdaVersionConfigTypeModel struct {
	LambdaARN     types.String `tfsdk:"lambda_arn"`
	LambdaVersion types.String `tfsdk:"lambda_version"`
}

type customSMSLambdaVersionConfigTypeModel struct {
	LambdaARN     types.String `tfsdk:"lambda_arn"`
	LambdaVersion types.String `tfsdk:"lambda_version"`
}

type preTokenGenerationVersionConfigTypeModel struct {
	LambdaARN     types.String `tfsdk:"lambda_arn"`
	LambdaVersion types.String `tfsdk:"lambda_version"`
}

type schemaAttributeTypeModel struct {
	AttributeDataType          types.String                                                         `tfsdk:"attribute_data_type"`
	DeveloperOnlyAttribute     types.Bool                                                           `tfsdk:"developer_only_attribute"`
	Mutable                    types.Bool                                                           `tfsdk:"mutable"`
	Name                       types.String                                                         `tfsdk:"name"`
	NumberAttributeConstraints fwtypes.ListNestedObjectValueOf[numberAttributeConstraintsTypeModel] `tfsdk:"number_attribute_constraints"`
	Required                   types.Bool                                                           `tfsdk:"required"`
	StringAttributeConstraints fwtypes.ListNestedObjectValueOf[stringAttributeConstraintsTypeModel] `tfsdk:"string_attribute_constraints"`
}

type numberAttributeConstraintsTypeModel struct {
	MaxValue types.String `tfsdk:"max_value"`
	MinValue types.String `tfsdk:"min_value"`
}

type stringAttributeConstraintsTypeModel struct {
	MaxLength types.String `tfsdk:"max_length"`
	MinLength types.String `tfsdk:"min_length"`
}
