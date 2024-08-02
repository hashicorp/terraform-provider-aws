// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcognitoidp "github.com/hashicorp/terraform-provider-aws/internal/service/cognitoidp"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(names.CognitoIDPServiceID, testAccErrorCheckSkip)
}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"not supported in this region",
	)
}

func TestAccCognitoIDPUserPool_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.UserPoolType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "cognito-idp", regexache.MustCompile(`userpool/.+`)),
					resource.TestMatchResourceAttr(resourceName, names.AttrEndpoint, regexache.MustCompile(`^cognito-idp\.[^.]+\.amazonaws.com/[\w-]+_[0-9A-Za-z]+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationDate),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_date"),
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "account_recovery_setting.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.allow_admin_create_user_only", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "email_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "email_configuration.0.email_sending_account", "COGNITO_DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.minimum_length", "8"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.require_lowercase", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.require_numbers", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.require_symbols", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.require_uppercase", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.temporary_password_validity_days", "7"),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.default_email_option", "CONFIRM_WITH_CODE"),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "estimated_number_of_users", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, "INACTIVE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCognitoIDPUserPool_deletionProtection(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.UserPoolType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_deletionProtection(rName, "ACTIVE"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, "ACTIVE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolConfig_deletionProtection(rName, "INACTIVE"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, "INACTIVE"),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_recovery(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.UserPoolType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_accountRecoverySingle(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "account_recovery_setting.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "account_recovery_setting.0.recovery_mechanism.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "account_recovery_setting.0.recovery_mechanism.*", map[string]string{
						names.AttrName:     "verified_email",
						names.AttrPriority: acctest.Ct1,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolConfig_accountRecoveryMulti(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "account_recovery_setting.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "account_recovery_setting.0.recovery_mechanism.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "account_recovery_setting.0.recovery_mechanism.*", map[string]string{
						names.AttrName:     "verified_email",
						names.AttrPriority: acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "account_recovery_setting.0.recovery_mechanism.*", map[string]string{
						names.AttrName:     "verified_phone_number",
						names.AttrPriority: acctest.Ct2,
					}),
				),
			},
			{
				Config: testAccUserPoolConfig_accountRecoveryUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "account_recovery_setting.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "account_recovery_setting.0.recovery_mechanism.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "account_recovery_setting.0.recovery_mechanism.*", map[string]string{
						names.AttrName:     "verified_phone_number",
						names.AttrPriority: acctest.Ct1,
					}),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_withAdminCreateUser(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.UserPoolType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_adminCreateConfiguration(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.allow_admin_create_user_only", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.email_message", "Your username is {username} and temporary password is {####}. "),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.email_subject", "FooBar {####}"),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.sms_message", "Your username is {username} and temporary password is {####}."),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolConfig_adminCreateConfigurationUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.allow_admin_create_user_only", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.email_message", "Your username is {username} and constant password is {####}. "),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.email_subject", "Foo{####}BaBaz"),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.sms_message", "Your username is {username} and constant password is {####}."),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/11858
func TestAccCognitoIDPUserPool_withAdminCreateUserAndPasswordPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.UserPoolType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_adminCreateAndPasswordPolicy(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.allow_admin_create_user_only", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.temporary_password_validity_days", "7"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCognitoIDPUserPool_withAdvancedSecurityMode(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.UserPoolType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_advancedSecurityMode(rName, "OFF"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "user_pool_add_ons.0.advanced_security_mode", "OFF"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolConfig_advancedSecurityMode(rName, "ENFORCED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "user_pool_add_ons.0.advanced_security_mode", "ENFORCED"),
				),
			},
			{
				Config: testAccUserPoolConfig_name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "user_pool_add_ons.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_withDevice(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.UserPoolType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_deviceConfiguration(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "device_configuration.0.challenge_required_on_new_device", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "device_configuration.0.device_only_remembered_on_user_prompt", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolConfig_deviceConfigurationUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "device_configuration.0.challenge_required_on_new_device", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "device_configuration.0.device_only_remembered_on_user_prompt", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_withEmailVerificationMessage(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.UserPoolType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	subject := sdkacctest.RandString(10)
	updatedSubject := sdkacctest.RandString(10)
	message := fmt.Sprintf("%s {####}", sdkacctest.RandString(10))
	upatedMessage := fmt.Sprintf("%s {####}", sdkacctest.RandString(10))
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_emailVerificationMessage(rName, subject, message),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "email_verification_subject", subject),
					resource.TestCheckResourceAttr(resourceName, "email_verification_message", message),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolConfig_emailVerificationMessage(rName, updatedSubject, upatedMessage),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "email_verification_subject", updatedSubject),
					resource.TestCheckResourceAttr(resourceName, "email_verification_message", upatedMessage),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_MFA_sms(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.UserPoolType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_mfaConfigurationSMSConfiguration(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "ON"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.0.external_id", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "sms_configuration.0.sns_caller_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolConfig_mfaConfiguration(rName, string(awstypes.UserPoolMfaTypeOff)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", acctest.Ct0),
				),
			},
			{
				Config: testAccUserPoolConfig_mfaConfigurationSMSConfiguration(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "ON"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.0.external_id", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "sms_configuration.0.sns_caller_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_MFA_smsAndSoftwareTokenMFA(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.UserPoolType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_mfaConfigurationSMSConfigurationAndSoftwareTokenMFAConfigurationEnabled(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "ON"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.0.external_id", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "sms_configuration.0.sns_caller_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.0.enabled", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolConfig_mfaConfigurationSMSConfigurationAndSoftwareTokenMFAConfigurationEnabled(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "ON"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.0.external_id", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "sms_configuration.0.sns_caller_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.0.enabled", acctest.CtTrue),
				),
			},
			{
				Config: testAccUserPoolConfig_mfaConfiguration(rName, string(awstypes.UserPoolMfaTypeOff)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_MFA_smsToSoftwareTokenMFA(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.UserPoolType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_mfaConfigurationSMSConfiguration(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "ON"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.0.external_id", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "sms_configuration.0.sns_caller_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolConfig_mfaConfigurationSoftwareTokenMFAConfigurationEnabled(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "ON"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.0.enabled", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_MFA_softwareTokenMFA(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.UserPoolType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_mfaConfigurationSoftwareTokenMFAConfigurationEnabled(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "ON"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.0.enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolConfig_mfaConfiguration(rName, string(awstypes.UserPoolMfaTypeOff)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", acctest.Ct0),
				),
			},
			{
				Config: testAccUserPoolConfig_mfaConfigurationSoftwareTokenMFAConfigurationEnabled(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "ON"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.0.enabled", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_MFA_softwareTokenMFAToSMS(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.UserPoolType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_mfaConfigurationSoftwareTokenMFAConfigurationEnabled(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "ON"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.0.enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolConfig_mfaConfigurationSMSConfiguration(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "ON"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.0.external_id", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "sms_configuration.0.sns_caller_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_smsAuthenticationMessage(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.UserPoolType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	smsAuthenticationMessage1 := "test authentication message {####}"
	smsAuthenticationMessage2 := "test authentication message updated {####}"
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_smsAuthenticationMessage(rName, smsAuthenticationMessage1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "sms_authentication_message", smsAuthenticationMessage1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolConfig_smsAuthenticationMessage(rName, smsAuthenticationMessage2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "sms_authentication_message", smsAuthenticationMessage2),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_sms(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.UserPoolType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_smsConfigurationExternalID(rName, "test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.0.external_id", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "sms_configuration.0.sns_caller_arn", iamRoleResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolConfig_name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", acctest.Ct1),
				),
			},
			{
				Config: testAccUserPoolConfig_smsConfigurationExternalID(rName, "test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.0.external_id", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "sms_configuration.0.sns_caller_arn", iamRoleResourceName, names.AttrARN),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_SMS_snsRegion(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.UserPoolType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_smsConfigurationSNSRegion(rName, acctest.Region()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.0.sns_region", acctest.Region()),
					resource.TestCheckResourceAttrPair(resourceName, "sms_configuration.0.sns_caller_arn", iamRoleResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCognitoIDPUserPool_SMS_externalID(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.UserPoolType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_smsConfigurationExternalID(rName, "test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.0.external_id", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "sms_configuration.0.sns_caller_arn", iamRoleResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolConfig_smsConfigurationExternalID(rName, "test2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.0.external_id", "test2"),
					resource.TestCheckResourceAttrPair(resourceName, "sms_configuration.0.sns_caller_arn", iamRoleResourceName, names.AttrARN),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_SMS_snsCallerARN(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.UserPoolType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_smsConfigurationExternalID(rName, "test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.0.external_id", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "sms_configuration.0.sns_caller_arn", iamRoleResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolConfig_smsConfigurationSNSCallerARN2(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.0.external_id", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "sms_configuration.0.sns_caller_arn", iamRoleResourceName, names.AttrARN),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_smsVerificationMessage(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.UserPoolType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	smsVerificationMessage1 := "test verification message {####}"
	smsVerificationMessage2 := "test verification message updated {####}"
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_smsVerificationMessage(rName, smsVerificationMessage1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "sms_verification_message", smsVerificationMessage1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolConfig_smsVerificationMessage(rName, smsVerificationMessage2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "sms_verification_message", smsVerificationMessage2),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_withEmail(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.UserPoolType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_emailConfiguration(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "email_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "email_configuration.0.reply_to_email_address", ""),
					resource.TestCheckResourceAttr(resourceName, "email_configuration.0.email_sending_account", "COGNITO_DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "email_configuration.0.from_email_address", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCognitoIDPUserPool_withEmailSource(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.UserPoolType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	replyTo := acctest.DefaultEmailAddress
	resourceName := "aws_cognito_user_pool.test"
	resourceName2 := "aws_ses_configuration_set.test"

	sourceARN := acctest.SkipIfEnvVarNotSet(t, "TEST_AWS_SES_VERIFIED_EMAIL_ARN")
	emailTo := sourceARN[strings.LastIndex(sourceARN, "/")+1:]

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_emailConfigurationSource(rName, replyTo, sourceARN, emailTo, "DEVELOPER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "email_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "email_configuration.0.reply_to_email_address", replyTo),
					resource.TestCheckResourceAttr(resourceName, "email_configuration.0.email_sending_account", "DEVELOPER"),
					resource.TestCheckResourceAttr(resourceName, "email_configuration.0.source_arn", sourceARN),
					resource.TestCheckResourceAttr(resourceName, "email_configuration.0.from_email_address", emailTo),
					resource.TestCheckResourceAttrPair(resourceName, "email_configuration.0.configuration_set", resourceName2, names.AttrName),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.UserPoolType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccUserPoolConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_withAliasAttributes(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.UserPoolType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_aliasAttributes(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "alias_attributes.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "alias_attributes.*", "preferred_username"),
					resource.TestCheckResourceAttr(resourceName, "auto_verified_attributes.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolConfig_aliasAttributesUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "alias_attributes.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "alias_attributes.*", names.AttrEmail),
					resource.TestCheckTypeSetElemAttr(resourceName, "alias_attributes.*", "preferred_username"),
					resource.TestCheckResourceAttr(resourceName, "auto_verified_attributes.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "auto_verified_attributes.*", names.AttrEmail),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_withUsernameAttributes(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.UserPoolType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_nameAttributes(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "username_attributes.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "username_attributes.*", "phone_number"),
					resource.TestCheckResourceAttr(resourceName, "auto_verified_attributes.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolConfig_nameAttributesUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "username_attributes.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "username_attributes.*", names.AttrEmail),
					resource.TestCheckTypeSetElemAttr(resourceName, "username_attributes.*", "phone_number"),
					resource.TestCheckResourceAttr(resourceName, "auto_verified_attributes.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "auto_verified_attributes.*", names.AttrEmail),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_withPasswordPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.UserPoolType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_passwordPolicy(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "password_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.minimum_length", "7"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.require_lowercase", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.require_numbers", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.require_symbols", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.require_uppercase", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.temporary_password_validity_days", "7"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolConfig_passwordPolicyUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "password_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.minimum_length", "9"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.require_lowercase", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.require_numbers", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.require_symbols", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.require_uppercase", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.temporary_password_validity_days", "14"),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_withUsername(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.UserPoolType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_nameConfiguration(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "username_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "username_configuration.0.case_sensitive", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolConfig_nameConfigurationUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "username_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "username_configuration.0.case_sensitive", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_withLambda(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.UserPoolType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"
	lambdaResourceName := "aws_lambda_function.test"
	lambdaUpdatedResourceName := "aws_lambda_function.second"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_lambda(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.create_auth_challenge", lambdaResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.custom_message", lambdaResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.define_auth_challenge", lambdaResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.post_authentication", lambdaResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.post_confirmation", lambdaResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.pre_authentication", lambdaResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.pre_sign_up", lambdaResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.pre_token_generation", lambdaResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.user_migration", lambdaResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.verify_auth_challenge_response", lambdaResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolConfig_lambdaUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "lambda_config.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.create_auth_challenge", lambdaUpdatedResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.custom_message", lambdaUpdatedResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.define_auth_challenge", lambdaUpdatedResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.post_authentication", lambdaUpdatedResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.post_confirmation", lambdaUpdatedResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.pre_authentication", lambdaUpdatedResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.pre_sign_up", lambdaUpdatedResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.pre_token_generation", lambdaUpdatedResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.user_migration", lambdaUpdatedResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.verify_auth_challenge_response", lambdaUpdatedResourceName, names.AttrARN),
				),
			},
			{
				Config: testAccUserPoolConfig_name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_WithLambda_email(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.UserPoolType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"
	lambdaResourceName := "aws_lambda_function.test"
	lambdaUpdatedResourceName := "aws_lambda_function.second"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_lambdaEmailSender(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.0.custom_email_sender.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.custom_email_sender.0.lambda_arn", lambdaResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.0.custom_email_sender.0.lambda_version", "V1_0"),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.0.custom_sms_sender.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolConfig_lambdaEmailSenderUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "lambda_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.0.custom_email_sender.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.custom_email_sender.0.lambda_arn", lambdaUpdatedResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.0.custom_email_sender.0.lambda_version", "V1_0"),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.0.custom_sms_sender.#", acctest.Ct0),
				),
			},
			{
				Config: testAccUserPoolConfig_lambdaEmailSenderUpdatedRemove(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "lambda_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.0.custom_email_sender.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.0.custom_sms_sender.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_WithLambda_sms(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.UserPoolType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"
	lambdaResourceName := "aws_lambda_function.test"
	lambdaUpdatedResourceName := "aws_lambda_function.second"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_lambdaSMSSender(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.0.custom_email_sender.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.0.custom_sms_sender.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.custom_sms_sender.0.lambda_arn", lambdaResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.0.custom_sms_sender.0.lambda_version", "V1_0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolConfig_lambdaSMSSenderUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "lambda_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.0.custom_email_sender.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.0.custom_sms_sender.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.custom_sms_sender.0.lambda_arn", lambdaUpdatedResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.0.custom_sms_sender.0.lambda_version", "V1_0"),
				),
			},
			{
				Config: testAccUserPoolConfig_lambdaSMSSenderUpdatedRemove(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "lambda_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.0.custom_email_sender.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.0.custom_sms_sender.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_WithLambda_preGenerationTokenConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.UserPoolType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"
	lambdaResourceName := "aws_lambda_function.test"
	lambdaUpdatedResourceName := "aws_lambda_function.second"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_lambdaPreTokenGenerationConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "user_pool_add_ons.0.advanced_security_mode", "ENFORCED"),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.0.pre_token_generation_config.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.pre_token_generation_config.0.lambda_arn", lambdaResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.0.pre_token_generation_config.0.lambda_version", "V2_0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolConfig_lambdaPreTokenGenerationConfigUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "lambda_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "user_pool_add_ons.0.advanced_security_mode", "ENFORCED"),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.0.pre_token_generation_config.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.pre_token_generation_config.0.lambda_arn", lambdaUpdatedResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.0.pre_token_generation_config.0.lambda_version", "V2_0"),
				),
			},
			{
				Config: testAccUserPoolConfig_lambdaPreTokenGenerationConfigRemove(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "lambda_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "user_pool_add_ons.0.advanced_security_mode", "ENFORCED"),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.0.pre_token_generation_config.#", acctest.Ct0),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/38164.
func TestAccCognitoIDPUserPool_addLambda(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.UserPoolType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"
	lambdaResourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.#", acctest.Ct0),
				),
			},
			{
				Config: testAccUserPoolConfig_lambda(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.create_auth_challenge", lambdaResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.custom_message", lambdaResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.define_auth_challenge", lambdaResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.post_authentication", lambdaResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.post_confirmation", lambdaResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.pre_authentication", lambdaResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.pre_sign_up", lambdaResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.pre_token_generation", lambdaResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.user_migration", lambdaResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.verify_auth_challenge_response", lambdaResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCognitoIDPUserPool_schemaAttributes(t *testing.T) {
	ctx := acctest.Context(t)
	var pool1, pool2 awstypes.UserPoolType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_schemaAttributes(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool1),
					resource.TestCheckResourceAttr(resourceName, "schema.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema.*", map[string]string{
						"attribute_data_type":                       "String",
						"developer_only_attribute":                  acctest.CtFalse,
						"mutable":                                   acctest.CtFalse,
						names.AttrName:                              names.AttrEmail,
						"number_attribute_constraints.#":            acctest.Ct0,
						"required":                                  acctest.CtTrue,
						"string_attribute_constraints.#":            acctest.Ct1,
						"string_attribute_constraints.0.min_length": "5",
						"string_attribute_constraints.0.max_length": acctest.Ct10,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema.*", map[string]string{
						"attribute_data_type":            "Boolean",
						"developer_only_attribute":       acctest.CtTrue,
						"mutable":                        acctest.CtFalse,
						names.AttrName:                   "mybool",
						"number_attribute_constraints.#": acctest.Ct0,
						"required":                       acctest.CtFalse,
						"string_attribute_constraints.#": acctest.Ct0,
					}),
				),
			},
			{
				Config: testAccUserPoolConfig_schemaAttributesUpdated(rName, "mybool"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool2),
					testAccCheckUserPoolNotRecreated(&pool1, &pool2),
					resource.TestCheckResourceAttr(resourceName, "schema.#", acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema.*", map[string]string{
						"attribute_data_type":                       "String",
						"developer_only_attribute":                  acctest.CtFalse,
						"mutable":                                   acctest.CtFalse,
						names.AttrName:                              names.AttrEmail,
						"number_attribute_constraints.#":            acctest.Ct0,
						"required":                                  acctest.CtTrue,
						"string_attribute_constraints.#":            acctest.Ct1,
						"string_attribute_constraints.0.min_length": "5",
						"string_attribute_constraints.0.max_length": acctest.Ct10,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema.*", map[string]string{
						"attribute_data_type":            "Boolean",
						"developer_only_attribute":       acctest.CtTrue,
						"mutable":                        acctest.CtFalse,
						names.AttrName:                   "mybool",
						"number_attribute_constraints.#": acctest.Ct0,
						"required":                       acctest.CtFalse,
						"string_attribute_constraints.#": acctest.Ct0,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema.*", map[string]string{
						"attribute_data_type":                      "Number",
						"developer_only_attribute":                 acctest.CtFalse,
						"mutable":                                  acctest.CtTrue,
						names.AttrName:                             "mynondevnumber",
						"number_attribute_constraints.#":           acctest.Ct1,
						"number_attribute_constraints.0.min_value": acctest.Ct2,
						"number_attribute_constraints.0.max_value": "6",
						"required":                                 acctest.CtFalse,
						"string_attribute_constraints.#":           acctest.Ct0,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCognitoIDPUserPool_schemaAttributesRemoved(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_schemaAttributesUpdated(rName, "mybool"),
			},
			{
				Config:      testAccUserPoolConfig_schemaAttributes(rName),
				ExpectError: regexache.MustCompile("cannot modify or remove schema items"),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_schemaAttributesModified(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_schemaAttributesUpdated(rName, "mybool"),
			},
			{
				Config:      testAccUserPoolConfig_schemaAttributesUpdated(rName, "mybool2"),
				ExpectError: regexache.MustCompile("cannot modify or remove schema items"),
			},
		},
	})
}

// Ref: https://github.com/hashicorp/terraform-provider-aws/issues/21654
func TestAccCognitoIDPUserPool_schemaAttributesStringAttributeConstraints(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.UserPoolType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				// Omit optional "string_attribute_constraints" schema argument to verify a persistent
				// diff is not present when AWS returns default values in the nested object.
				Config: testAccUserPoolConfig_schemaAttributesStringAttributeConstraints(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
				),
			},
			{
				// Attempting to explicitly set constraints to non-default values after creation
				// should trigger an error
				Config:      testAccUserPoolConfig_schemaAttributes(rName),
				ExpectError: regexache.MustCompile("cannot modify or remove schema items"),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_withVerificationMessageTemplate(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.UserPoolType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"

	emailMessage := "foo {####} bar"
	emailMessageByLink := "{##foobar##}"
	emailSubject := "foobar {####}"
	emailSubjectByLink := "foobar"
	smsMessage := "{####} baz"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_verificationMessageTemplate(rName, emailMessage, emailMessageByLink, emailSubject, emailSubjectByLink, smsMessage),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.default_email_option", "CONFIRM_WITH_LINK"),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.email_message", emailMessage),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.email_message_by_link", emailMessageByLink),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.email_subject", emailSubject),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.email_subject_by_link", emailSubjectByLink),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.sms_message", smsMessage),

					/* Setting Verification template attributes like EmailMessage, EmailSubject or SmsMessage
					will implicitly set EmailVerificationMessage, EmailVerificationSubject and SmsVerificationMessage attributes.
					*/
					resource.TestCheckResourceAttr(resourceName, "email_verification_message", emailMessage),
					resource.TestCheckResourceAttr(resourceName, "email_verification_subject", emailSubject),
					resource.TestCheckResourceAttr(resourceName, "sms_verification_message", smsMessage),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolConfig_verificationMessageTemplateDefaultEmailOption(rName, emailMessage, emailSubject, smsMessage),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.default_email_option", "CONFIRM_WITH_CODE"),
					resource.TestCheckResourceAttr(resourceName, "email_verification_message", emailMessage),
					resource.TestCheckResourceAttr(resourceName, "email_verification_subject", emailSubject),
					resource.TestCheckResourceAttr(resourceName, "sms_verification_message", smsMessage),

					/* Setting EmailVerificationMessage, EmailVerificationSubject and SmsVerificationMessage attributes
					will implicitly set verification template attributes like EmailMessage, EmailSubject or SmsMessage.
					*/
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.email_message", emailMessage),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.email_subject", emailSubject),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.sms_message", smsMessage),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_withVerificationMessageTemplateUTF8(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.UserPoolType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"

	emailMessage := "{####}" + strings.Repeat("あ", 994)             // = 1000
	emailMessageByLink := "{##foobar##}" + strings.Repeat("い", 988) // = 1000
	emailSubject := strings.Repeat("う", 140)
	emailSubjectByLink := strings.Repeat("え", 140)
	smsMessage := "{####}" + strings.Repeat("お", 134) // = 140

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_verificationMessageTemplate(rName, emailMessage, emailMessageByLink, emailSubject, emailSubjectByLink, smsMessage),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.default_email_option", "CONFIRM_WITH_LINK"),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.email_message", emailMessage),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.email_message_by_link", emailMessageByLink),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.email_subject", emailSubject),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.email_subject_by_link", emailSubjectByLink),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.sms_message", smsMessage),

					/* Setting Verification template attributes like EmailMessage, EmailSubject or SmsMessage
					will implicitly set EmailVerificationMessage, EmailVerificationSubject and SmsVerificationMessage attributes.
					*/
					resource.TestCheckResourceAttr(resourceName, "email_verification_message", emailMessage),
					resource.TestCheckResourceAttr(resourceName, "email_verification_subject", emailSubject),
					resource.TestCheckResourceAttr(resourceName, "sms_verification_message", smsMessage),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolConfig_verificationMessageTemplateDefaultEmailOption(rName, emailMessage, emailSubject, smsMessage),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.default_email_option", "CONFIRM_WITH_CODE"),
					resource.TestCheckResourceAttr(resourceName, "email_verification_message", emailMessage),
					resource.TestCheckResourceAttr(resourceName, "email_verification_subject", emailSubject),
					resource.TestCheckResourceAttr(resourceName, "sms_verification_message", smsMessage),

					/* Setting EmailVerificationMessage, EmailVerificationSubject and SmsVerificationMessage attributes
					will implicitly set verification template attributes like EmailMessage, EmailSubject or SmsMessage.
					*/
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.email_message", emailMessage),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.email_subject", emailSubject),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.sms_message", smsMessage),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_update(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.UserPoolType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	optionalMfa := "OPTIONAL"
	offMfa := "OFF"
	authenticationMessage := fmt.Sprintf("%s {####}", sdkacctest.RandString(10))
	updatedAuthenticationMessage := fmt.Sprintf("%s {####}", sdkacctest.RandString(10))
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_update(rName, optionalMfa, authenticationMessage),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "auto_verified_attributes.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", optionalMfa),
					resource.TestCheckResourceAttr(resourceName, "email_verification_message", "Foo {####} Bar"),
					resource.TestCheckResourceAttr(resourceName, "email_verification_subject", "FooBar {####}"),
					resource.TestCheckResourceAttr(resourceName, "sms_verification_message", "{####} Baz"),
					resource.TestCheckResourceAttr(resourceName, "sms_authentication_message", authenticationMessage),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.allow_admin_create_user_only", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.email_message", "Your username is {username} and temporary password is {####}. "),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.email_subject", "FooBar {####}"),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.sms_message", "Your username is {username} and temporary password is {####}."),
					resource.TestCheckResourceAttr(resourceName, "device_configuration.0.challenge_required_on_new_device", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "device_configuration.0.device_only_remembered_on_user_prompt", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.default_email_option", "CONFIRM_WITH_CODE"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "sms_configuration.0.external_id"),
					resource.TestCheckResourceAttrSet(resourceName, "sms_configuration.0.sns_caller_arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolConfig_update(rName, optionalMfa, updatedAuthenticationMessage),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "auto_verified_attributes.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", optionalMfa),
					resource.TestCheckResourceAttr(resourceName, "email_verification_message", "Foo {####} Bar"),
					resource.TestCheckResourceAttr(resourceName, "email_verification_subject", "FooBar {####}"),
					resource.TestCheckResourceAttr(resourceName, "sms_verification_message", "{####} Baz"),
					resource.TestCheckResourceAttr(resourceName, "sms_authentication_message", updatedAuthenticationMessage),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.allow_admin_create_user_only", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.email_message", "Your username is {username} and temporary password is {####}. "),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.email_subject", "FooBar {####}"),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.sms_message", "Your username is {username} and temporary password is {####}."),
					resource.TestCheckResourceAttr(resourceName, "device_configuration.0.challenge_required_on_new_device", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "device_configuration.0.device_only_remembered_on_user_prompt", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.default_email_option", "CONFIRM_WITH_CODE"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "sms_configuration.0.external_id"),
					resource.TestCheckResourceAttrSet(resourceName, "sms_configuration.0.sns_caller_arn"),
				),
			},
			{
				Config: testAccUserPoolConfig_update(rName, offMfa, updatedAuthenticationMessage),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "auto_verified_attributes.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", offMfa),
					resource.TestCheckResourceAttr(resourceName, "email_verification_message", "Foo {####} Bar"),
					resource.TestCheckResourceAttr(resourceName, "email_verification_subject", "FooBar {####}"),
					resource.TestCheckResourceAttr(resourceName, "sms_verification_message", "{####} Baz"),
					resource.TestCheckResourceAttr(resourceName, "sms_authentication_message", updatedAuthenticationMessage),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.allow_admin_create_user_only", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.email_message", "Your username is {username} and temporary password is {####}. "),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.email_subject", "FooBar {####}"),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.sms_message", "Your username is {username} and temporary password is {####}."),
					resource.TestCheckResourceAttr(resourceName, "device_configuration.0.challenge_required_on_new_device", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "device_configuration.0.device_only_remembered_on_user_prompt", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.default_email_option", "CONFIRM_WITH_CODE"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "sms_configuration.0.external_id"),
					resource.TestCheckResourceAttrSet(resourceName, "sms_configuration.0.sns_caller_arn"),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.UserPoolType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcognitoidp.ResourceUserPool(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCognitoIDPUserPool_withUserAttributeUpdateSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.UserPoolType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_userAttributeUpdateSettings(rName, string(awstypes.VerifiedAttributeTypeEmail)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "auto_verified_attributes.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "auto_verified_attributes.0", string(awstypes.VerifiedAttributeTypeEmail)),
					resource.TestCheckResourceAttr(resourceName, "user_attribute_update_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "user_attribute_update_settings.0.attributes_require_verification_before_update.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "user_attribute_update_settings.0.attributes_require_verification_before_update.0", string(awstypes.VerifiedAttributeTypeEmail)),
				),
			},
			{
				Config: testAccUserPoolConfig_name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "auto_verified_attributes.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "user_attribute_update_settings.#", acctest.Ct0),
				),
			},
		},
	})
}

func testAccCheckUserPoolDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cognito_user_pool" {
				continue
			}

			_, err := tfcognitoidp.FindUserPoolByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Cognito User Pool %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckUserPoolExists(ctx context.Context, n string, v *awstypes.UserPoolType) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPClient(ctx)

		output, err := tfcognitoidp.FindUserPoolByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if v != nil {
			*v = *output
		}

		return nil
	}
}

func testAccCheckUserPoolNotRecreated(pool1, pool2 *awstypes.UserPoolType) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.ToTime(pool1.CreationDate).Equal(aws.ToTime(pool2.CreationDate)) {
			return fmt.Errorf("user pool was recreated. expected: %s, got: %s", pool1.CreationDate, pool2.CreationDate)
		}
		return nil
	}
}

func testAccPreCheckIdentityProvider(ctx context.Context, t *testing.T) {
	t.Helper()
	acctest.PreCheckCognitoIdentityProvider(ctx, t)
}

func testAccUserPoolSMSConfigurationConfig_base(rName string, externalID string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Condition = {
        "StringEquals" = {
          "sts:ExternalId" = %[2]q
        }
      }

      Effect = "Allow"
      Principal = {
        Service = "cognito-idp.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.id

  policy = jsonencode({
    Statement = [{
      Action   = "sns:publish"
      Effect   = "Allow"
      Resource = "*"
    }]
    Version = "2012-10-17"
  })
}
`, rName, externalID)
}

func testAccUserPoolConfig_name(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}
`, rName)
}

func testAccUserPoolConfig_deletionProtection(rName, active string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name                = %[1]q
  deletion_protection = %[2]q
}
`, rName, active)
}

func testAccUserPoolConfig_accountRecoverySingle(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  account_recovery_setting {
    recovery_mechanism {
      name     = "verified_email"
      priority = 1
    }
  }
}
`, rName)
}

func testAccUserPoolConfig_accountRecoveryMulti(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  account_recovery_setting {
    recovery_mechanism {
      name     = "verified_email"
      priority = 1
    }

    recovery_mechanism {
      name     = "verified_phone_number"
      priority = 2
    }
  }
}
`, rName)
}

func testAccUserPoolConfig_accountRecoveryUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  account_recovery_setting {
    recovery_mechanism {
      name     = "verified_phone_number"
      priority = 1
    }
  }
}
`, rName)
}

func testAccUserPoolConfig_adminCreateConfiguration(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  admin_create_user_config {
    allow_admin_create_user_only = true

    invite_message_template {
      email_message = "Your username is {username} and temporary password is {####}. "
      email_subject = "FooBar {####}"
      sms_message   = "Your username is {username} and temporary password is {####}."
    }
  }
}
`, rName)
}

func testAccUserPoolConfig_adminCreateConfigurationUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  admin_create_user_config {
    allow_admin_create_user_only = false

    invite_message_template {
      email_message = "Your username is {username} and constant password is {####}. "
      email_subject = "Foo{####}BaBaz"
      sms_message   = "Your username is {username} and constant password is {####}."
    }
  }
}
`, rName)
}

func testAccUserPoolConfig_advancedSecurityMode(rName string, advancedSecurityMode string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  user_pool_add_ons {
    advanced_security_mode = %[2]q
  }
}
`, rName, advancedSecurityMode)
}

func testAccUserPoolConfig_deviceConfiguration(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  device_configuration {
    challenge_required_on_new_device      = true
    device_only_remembered_on_user_prompt = false
  }
}
`, rName)
}

func testAccUserPoolConfig_deviceConfigurationUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  device_configuration {
    challenge_required_on_new_device      = false
    device_only_remembered_on_user_prompt = true
  }
}
`, rName)
}

func testAccUserPoolConfig_emailVerificationMessage(rName, subject, message string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name                       = %[1]q
  email_verification_subject = "%[2]s"
  email_verification_message = "%[3]s"

  verification_message_template {
    default_email_option = "CONFIRM_WITH_CODE"
  }
}
`, rName, subject, message)
}

func testAccUserPoolConfig_mfaConfiguration(rName string, mfaConfiguration string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  mfa_configuration = %[2]q
  name              = %[1]q
}
`, rName, mfaConfiguration)
}

func testAccUserPoolConfig_mfaConfigurationSMSConfiguration(rName string) string {
	return acctest.ConfigCompose(testAccUserPoolSMSConfigurationConfig_base(rName, "test"), fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  mfa_configuration = "ON"
  name              = %[1]q

  account_recovery_setting {
    recovery_mechanism {
      name     = "verified_email"
      priority = 1
    }
  }

  sms_configuration {
    external_id    = "test"
    sns_caller_arn = aws_iam_role.test.arn
  }
}
`, rName))
}

func testAccUserPoolConfig_mfaConfigurationSMSConfigurationAndSoftwareTokenMFAConfigurationEnabled(rName string, enabled bool) string {
	return acctest.ConfigCompose(testAccUserPoolSMSConfigurationConfig_base(rName, "test"), fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  mfa_configuration = "ON"
  name              = %[1]q

  account_recovery_setting {
    recovery_mechanism {
      name     = "verified_email"
      priority = 1
    }
  }

  sms_configuration {
    external_id    = "test"
    sns_caller_arn = aws_iam_role.test.arn
  }

  software_token_mfa_configuration {
    enabled = %[2]t
  }
}
`, rName, enabled))
}

func testAccUserPoolConfig_mfaConfigurationSoftwareTokenMFAConfigurationEnabled(rName string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  mfa_configuration = "ON"
  name              = %[1]q

  account_recovery_setting {
    recovery_mechanism {
      name     = "verified_email"
      priority = 1
    }
  }

  software_token_mfa_configuration {
    enabled = %[2]t
  }
}
`, rName, enabled)
}

func testAccUserPoolConfig_smsAuthenticationMessage(rName, smsAuthenticationMessage string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name                       = %[1]q
  sms_authentication_message = %[2]q
}
`, rName, smsAuthenticationMessage)
}

func testAccUserPoolConfig_smsConfigurationExternalID(rName string, externalID string) string {
	return acctest.ConfigCompose(testAccUserPoolSMSConfigurationConfig_base(rName, externalID), fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  sms_configuration {
    external_id    = %[2]q
    sns_caller_arn = aws_iam_role.test.arn
  }
}
`, rName, externalID))
}

func testAccUserPoolConfig_smsConfigurationSNSRegion(rName string, snsRegion string) string {
	return acctest.ConfigCompose(testAccUserPoolSMSConfigurationConfig_base(rName, "test"), fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  sms_configuration {
    external_id    = "test"
    sns_caller_arn = aws_iam_role.test.arn
    sns_region     = %[2]q
  }
}
`, rName, snsRegion))
}

func testAccUserPoolConfig_smsConfigurationSNSCallerARN2(rName string) string {
	return acctest.ConfigCompose(testAccUserPoolSMSConfigurationConfig_base(rName+"-2", "test"), fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  sms_configuration {
    external_id    = "test"
    sns_caller_arn = aws_iam_role.test.arn
  }
}
`, rName))
}

func testAccUserPoolConfig_smsVerificationMessage(rName, smsVerificationMessage string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name                     = %[1]q
  sms_verification_message = %[2]q
}
`, rName, smsVerificationMessage)
}

func testAccUserPoolConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccUserPoolConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccUserPoolConfig_emailConfiguration(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  email_configuration {
    email_sending_account = "COGNITO_DEFAULT"
  }
}
`, rName)
}

func testAccUserPoolConfig_emailConfigurationSource(rName, email, arn, from, account string) string {
	return fmt.Sprintf(`
resource "aws_ses_configuration_set" "test" {
  name = %[1]q

  delivery_options {
    tls_policy = "Optional"
  }
}

resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  email_configuration {
    reply_to_email_address = %[2]q
    source_arn             = %[3]q
    from_email_address     = %[4]q
    email_sending_account  = %[5]q
    configuration_set      = aws_ses_configuration_set.test.name
  }
}
`, rName, email, arn, from, account)
}

func testAccUserPoolConfig_aliasAttributes(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  alias_attributes = ["preferred_username"]
}
`, rName)
}

func testAccUserPoolConfig_aliasAttributesUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  alias_attributes         = ["email", "preferred_username"]
  auto_verified_attributes = ["email"]
}
`, rName)
}

func testAccUserPoolConfig_nameAttributes(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  username_attributes = ["phone_number"]
}
`, rName)
}

func testAccUserPoolConfig_nameAttributesUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  username_attributes      = ["email", "phone_number"]
  auto_verified_attributes = ["email"]
}
`, rName)
}

func testAccUserPoolConfig_adminCreateAndPasswordPolicy(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  admin_create_user_config {
    allow_admin_create_user_only = true
  }

  password_policy {
    minimum_length                   = 7
    require_lowercase                = true
    require_numbers                  = false
    require_symbols                  = true
    require_uppercase                = false
    temporary_password_validity_days = 7
  }
}
`, rName)
}

func testAccUserPoolConfig_passwordPolicy(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  password_policy {
    minimum_length                   = 7
    require_lowercase                = true
    require_numbers                  = false
    require_symbols                  = true
    require_uppercase                = false
    temporary_password_validity_days = 7
  }
}
`, name)
}

func testAccUserPoolConfig_passwordPolicyUpdated(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  password_policy {
    minimum_length                   = 9
    require_lowercase                = false
    require_numbers                  = true
    require_symbols                  = false
    require_uppercase                = true
    temporary_password_validity_days = 14
  }
}
`, name)
}

func testAccUserPoolConfig_nameConfiguration(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  username_configuration {
    case_sensitive = true
  }
}
`, name)
}

func testAccUserPoolConfig_nameConfigurationUpdated(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  username_configuration {
    case_sensitive = false
  }
}
`, name)
}

func testAccUserPoolLambdaConfig_base(name string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.test.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"
}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}
`, name)
}

func testAccUserPoolConfig_lambda(name string) string {
	return acctest.ConfigCompose(testAccUserPoolLambdaConfig_base(name), fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  lambda_config {
    create_auth_challenge          = aws_lambda_function.test.arn
    custom_message                 = aws_lambda_function.test.arn
    define_auth_challenge          = aws_lambda_function.test.arn
    post_authentication            = aws_lambda_function.test.arn
    post_confirmation              = aws_lambda_function.test.arn
    pre_authentication             = aws_lambda_function.test.arn
    pre_sign_up                    = aws_lambda_function.test.arn
    pre_token_generation           = aws_lambda_function.test.arn
    user_migration                 = aws_lambda_function.test.arn
    verify_auth_challenge_response = aws_lambda_function.test.arn
  }
}
`, name))
}

func testAccUserPoolConfig_lambdaUpdated(name string) string {
	return acctest.ConfigCompose(testAccUserPoolLambdaConfig_base(name), fmt.Sprintf(`
resource "aws_lambda_function" "second" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s_second"
  role          = aws_iam_role.test.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"
}

resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  lambda_config {
    create_auth_challenge          = aws_lambda_function.second.arn
    custom_message                 = aws_lambda_function.second.arn
    define_auth_challenge          = aws_lambda_function.second.arn
    post_authentication            = aws_lambda_function.second.arn
    post_confirmation              = aws_lambda_function.second.arn
    pre_authentication             = aws_lambda_function.second.arn
    pre_sign_up                    = aws_lambda_function.second.arn
    pre_token_generation           = aws_lambda_function.second.arn
    user_migration                 = aws_lambda_function.second.arn
    verify_auth_challenge_response = aws_lambda_function.second.arn
  }
}
`, name))
}

func testAccUserPoolConfig_lambdaEmailSender(name string) string {
	return acctest.ConfigCompose(testAccUserPoolLambdaConfig_base(name), fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  lambda_config {
    kms_key_id = aws_kms_key.test.arn

    custom_email_sender {
      lambda_arn     = aws_lambda_function.test.arn
      lambda_version = "V1_0"
    }
  }
}
`, name))
}

func testAccUserPoolConfig_lambdaEmailSenderUpdated(name string) string {
	return acctest.ConfigCompose(testAccUserPoolLambdaConfig_base(name), fmt.Sprintf(`
resource "aws_lambda_function" "second" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s_second"
  role          = aws_iam_role.test.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"
}

resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  lambda_config {
    kms_key_id = aws_kms_key.test.arn

    custom_email_sender {
      lambda_arn     = aws_lambda_function.second.arn
      lambda_version = "V1_0"
    }
  }
}
`, name))
}

func testAccUserPoolConfig_lambdaEmailSenderUpdatedRemove(name string) string {
	return acctest.ConfigCompose(testAccUserPoolLambdaConfig_base(name), fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  lambda_config {
    kms_key_id = aws_kms_key.test.arn
  }
}
`, name))
}

func testAccUserPoolConfig_lambdaSMSSender(name string) string {
	return acctest.ConfigCompose(testAccUserPoolLambdaConfig_base(name), fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  lambda_config {
    kms_key_id = aws_kms_key.test.arn

    custom_sms_sender {
      lambda_arn     = aws_lambda_function.test.arn
      lambda_version = "V1_0"
    }
  }
}
`, name))
}

func testAccUserPoolConfig_lambdaSMSSenderUpdated(name string) string {
	return acctest.ConfigCompose(testAccUserPoolLambdaConfig_base(name), fmt.Sprintf(`
resource "aws_lambda_function" "second" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s_second"
  role          = aws_iam_role.test.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"
}

resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  lambda_config {
    kms_key_id = aws_kms_key.test.arn

    custom_sms_sender {
      lambda_arn     = aws_lambda_function.second.arn
      lambda_version = "V1_0"
    }
  }
}
`, name))
}

func testAccUserPoolConfig_lambdaSMSSenderUpdatedRemove(name string) string {
	return acctest.ConfigCompose(testAccUserPoolLambdaConfig_base(name), fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  lambda_config {
    kms_key_id = aws_kms_key.test.arn
  }
}
`, name))
}

func testAccUserPoolConfig_lambdaPreTokenGenerationConfig(name string) string {
	return acctest.ConfigCompose(testAccUserPoolLambdaConfig_base(name), fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  user_pool_add_ons {
    advanced_security_mode = "ENFORCED"
  }

  lambda_config {
    pre_token_generation_config {
      lambda_arn     = aws_lambda_function.test.arn
      lambda_version = "V2_0"
    }
  }
}
`, name))
}

func testAccUserPoolConfig_lambdaPreTokenGenerationConfigUpdated(name string) string {
	return acctest.ConfigCompose(testAccUserPoolLambdaConfig_base(name), fmt.Sprintf(`
resource "aws_lambda_function" "second" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s_second"
  role          = aws_iam_role.test.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"
}

resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  user_pool_add_ons {
    advanced_security_mode = "ENFORCED"
  }

  lambda_config {
    pre_token_generation_config {
      lambda_arn     = aws_lambda_function.second.arn
      lambda_version = "V2_0"
    }
  }
}
`, name))
}

func testAccUserPoolConfig_lambdaPreTokenGenerationConfigRemove(name string) string {
	return acctest.ConfigCompose(testAccUserPoolLambdaConfig_base(name), fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  user_pool_add_ons {
    advanced_security_mode = "ENFORCED"
  }

}
`, name))
}

func testAccUserPoolConfig_schemaAttributes(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = "%[1]s"

  schema {
    attribute_data_type      = "String"
    developer_only_attribute = false
    mutable                  = false
    name                     = "email"
    required                 = true

    string_attribute_constraints {
      min_length = 5
      max_length = 10
    }
  }

  schema {
    attribute_data_type      = "Boolean"
    developer_only_attribute = true
    mutable                  = false
    name                     = "mybool"
    required                 = false
  }
}
`, name)
}

func testAccUserPoolConfig_schemaAttributesUpdated(name string, boolname string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = "%[1]s"

  schema {
    attribute_data_type      = "String"
    developer_only_attribute = false
    mutable                  = false
    name                     = "email"
    required                 = true

    string_attribute_constraints {
      min_length = 5
      max_length = 10
    }
  }

  schema {
    attribute_data_type      = "Boolean"
    developer_only_attribute = true
    mutable                  = false
    name                     = %[2]q
    required                 = false
  }

  schema {
    attribute_data_type      = "Number"
    developer_only_attribute = false
    mutable                  = true
    name                     = "mynondevnumber"
    required                 = false

    number_attribute_constraints {
      min_value = 2
      max_value = 6
    }
  }
}
`, name, boolname)
}

func testAccUserPoolConfig_schemaAttributesStringAttributeConstraints(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = "%[1]s"

  schema {
    attribute_data_type      = "String"
    developer_only_attribute = false
    mutable                  = false
    name                     = "email"
    required                 = true
  }

  schema {
    attribute_data_type      = "Boolean"
    developer_only_attribute = true
    mutable                  = false
    name                     = "mybool"
    required                 = false
  }
}
`, name)
}

func testAccUserPoolConfig_verificationMessageTemplate(name, emailMessage, emailMessageByLink, emailSubject, emailSubjectByLink, smsMessage string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  # Setting Verification template attributes like EmailMessage, EmailSubject or SmsMessage
  # will implicitly set EmailVerificationMessage, EmailVerificationSubject and SmsVerificationMessage
  # attributes.

  verification_message_template {
    default_email_option  = "CONFIRM_WITH_LINK"
    email_message         = %[2]q
    email_message_by_link = %[3]q
    email_subject         = %[4]q
    email_subject_by_link = %[5]q
    sms_message           = %[6]q
  }
}
`, name, emailMessage, emailMessageByLink, emailSubject, emailSubjectByLink, smsMessage)
}

func testAccUserPoolConfig_verificationMessageTemplateDefaultEmailOption(name, emailVerificationMessage, emailVerificationSubject, smsVerificationMessage string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  email_verification_message = %[2]q
  email_verification_subject = %[3]q
  sms_verification_message   = %[4]q

  verification_message_template {
    default_email_option = "CONFIRM_WITH_CODE"
  }
}
`, name, emailVerificationMessage, emailVerificationSubject, smsVerificationMessage)
}

func testAccUserPoolConfig_update(name string, mfaconfig, smsAuthMsg string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {
}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/service-role/"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "cognito-idp.amazonaws.com"
      },
      "Action": "sts:AssumeRole",
      "Condition": {
        "StringEquals": {
          "sts:ExternalId": "${data.aws_caller_identity.current.account_id}"
        }
      }
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "sns:publish"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}

resource "aws_cognito_user_pool" "test" {
  name                     = %[1]q
  auto_verified_attributes = ["email"]
  mfa_configuration        = "%[2]s"

  email_verification_message = "Foo {####} Bar"
  email_verification_subject = "FooBar {####}"
  sms_verification_message   = "{####} Baz"
  sms_authentication_message = "%[3]s"

  admin_create_user_config {
    allow_admin_create_user_only = true

    invite_message_template {
      email_message = "Your username is {username} and temporary password is {####}. "
      email_subject = "FooBar {####}"
      sms_message   = "Your username is {username} and temporary password is {####}."
    }
  }

  device_configuration {
    challenge_required_on_new_device      = true
    device_only_remembered_on_user_prompt = false
  }

  verification_message_template {
    default_email_option = "CONFIRM_WITH_CODE"
  }

  sms_configuration {
    external_id    = data.aws_caller_identity.current.account_id
    sns_caller_arn = aws_iam_role.test.arn
  }
}
`, name, mfaconfig, smsAuthMsg)
}

func testAccUserPoolConfig_userAttributeUpdateSettings(name, attr string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  auto_verified_attributes = ["%[2]s"]

  user_attribute_update_settings {
    attributes_require_verification_before_update = ["%[2]s"]
  }
}
`, name, attr)
}
