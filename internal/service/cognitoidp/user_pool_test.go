package cognitoidp_test

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcognitoidp "github.com/hashicorp/terraform-provider-aws/internal/service/cognitoidp"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(cognitoidentityprovider.EndpointsID, testAccErrorCheckSkip)

}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"not supported in this region",
	)
}

func TestAccCognitoIDPUserPool_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(resourceName, nil),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "cognito-idp", regexp.MustCompile(`userpool/.+`)),
					resource.TestMatchResourceAttr(resourceName, "endpoint", regexp.MustCompile(`^cognito-idp\.[^.]+\.amazonaws.com/[\w-]+_[0-9a-zA-Z]+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_date"),
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "account_recovery_setting.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.allow_admin_create_user_only", "false"),
					resource.TestCheckResourceAttr(resourceName, "email_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "email_configuration.0.email_sending_account", "COGNITO_DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.minimum_length", "8"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.require_lowercase", "true"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.require_numbers", "true"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.require_symbols", "true"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.require_uppercase", "true"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.temporary_password_validity_days", "7"),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.default_email_option", "CONFIRM_WITH_CODE"),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "estimated_number_of_users", "0"),
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

func TestAccCognitoIDPUserPool_recovery(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_accountRecoverySingle(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(resourceName, nil),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "account_recovery_setting.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "account_recovery_setting.0.recovery_mechanism.#", "1"),
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
					testAccCheckUserPoolExists(resourceName, nil),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "account_recovery_setting.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "account_recovery_setting.0.recovery_mechanism.#", "2"),
				),
			},
			{
				Config: testAccUserPoolConfig_accountRecoveryUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(resourceName, nil),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "account_recovery_setting.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "account_recovery_setting.0.recovery_mechanism.#", "1"),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_withAdminCreateUser(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_adminCreateConfiguration(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(resourceName, nil),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.allow_admin_create_user_only", "true"),
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
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.allow_admin_create_user_only", "false"),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_adminCreateAndPasswordPolicy(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(resourceName, nil),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.allow_admin_create_user_only", "true"),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_advancedSecurityMode(rName, "OFF"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(resourceName, nil),
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
					resource.TestCheckResourceAttr(resourceName, "user_pool_add_ons.#", "0"),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_withDevice(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_deviceConfiguration(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(resourceName, nil),
					resource.TestCheckResourceAttr(resourceName, "device_configuration.0.challenge_required_on_new_device", "true"),
					resource.TestCheckResourceAttr(resourceName, "device_configuration.0.device_only_remembered_on_user_prompt", "false"),
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
					resource.TestCheckResourceAttr(resourceName, "device_configuration.0.challenge_required_on_new_device", "false"),
					resource.TestCheckResourceAttr(resourceName, "device_configuration.0.device_only_remembered_on_user_prompt", "true"),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_withEmailVerificationMessage(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	subject := sdkacctest.RandString(10)
	updatedSubject := sdkacctest.RandString(10)
	message := fmt.Sprintf("%s {####}", sdkacctest.RandString(10))
	upatedMessage := fmt.Sprintf("%s {####}", sdkacctest.RandString(10))
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_emailVerificationMessage(rName, subject, message),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(resourceName, nil),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_mfaConfigurationSMSConfiguration(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "ON"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.0.external_id", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "sms_configuration.0.sns_caller_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolConfig_mfaConfiguration(rName, cognitoidentityprovider.UserPoolMfaTypeOff),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", "0"),
				),
			},
			{
				Config: testAccUserPoolConfig_mfaConfigurationSMSConfiguration(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "ON"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.0.external_id", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "sms_configuration.0.sns_caller_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", "0"),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_MFA_smsAndSoftwareTokenMFA(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_mfaConfigurationSMSConfigurationAndSoftwareTokenMFAConfigurationEnabled(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "ON"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.0.external_id", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "sms_configuration.0.sns_caller_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.0.enabled", "false"),
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
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.0.external_id", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "sms_configuration.0.sns_caller_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.0.enabled", "true"),
				),
			},
			{
				Config: testAccUserPoolConfig_mfaConfiguration(rName, cognitoidentityprovider.UserPoolMfaTypeOff),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", "0"),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_MFA_smsToSoftwareTokenMFA(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_mfaConfigurationSMSConfiguration(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "ON"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.0.external_id", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "sms_configuration.0.sns_caller_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", "0"),
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
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.0.enabled", "true"),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_MFA_softwareTokenMFA(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_mfaConfigurationSoftwareTokenMFAConfigurationEnabled(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "ON"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.0.enabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolConfig_mfaConfiguration(rName, cognitoidentityprovider.UserPoolMfaTypeOff),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", "0"),
				),
			},
			{
				Config: testAccUserPoolConfig_mfaConfigurationSoftwareTokenMFAConfigurationEnabled(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "ON"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.0.enabled", "true"),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_MFA_softwareTokenMFAToSMS(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_mfaConfigurationSoftwareTokenMFAConfigurationEnabled(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "ON"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.0.enabled", "true"),
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
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.0.external_id", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "sms_configuration.0.sns_caller_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", "0"),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_smsAuthenticationMessage(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	smsAuthenticationMessage1 := "test authentication message {####}"
	smsAuthenticationMessage2 := "test authentication message updated {####}"
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_smsAuthenticationMessage(rName, smsAuthenticationMessage1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(resourceName, nil),
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
					testAccCheckUserPoolExists(resourceName, nil),
					resource.TestCheckResourceAttr(resourceName, "sms_authentication_message", smsAuthenticationMessage2),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_sms(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_smsConfigurationExternalID(rName, "test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.0.external_id", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "sms_configuration.0.sns_caller_arn", iamRoleResourceName, "arn"),
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
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "1"),
				),
			},
			{
				Config: testAccUserPoolConfig_smsConfigurationExternalID(rName, "test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.0.external_id", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "sms_configuration.0.sns_caller_arn", iamRoleResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_SMS_externalID(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_smsConfigurationExternalID(rName, "test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.0.external_id", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "sms_configuration.0.sns_caller_arn", iamRoleResourceName, "arn"),
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
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.0.external_id", "test2"),
					resource.TestCheckResourceAttrPair(resourceName, "sms_configuration.0.sns_caller_arn", iamRoleResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_SMS_snsCallerARN(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_smsConfigurationExternalID(rName, "test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.0.external_id", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "sms_configuration.0.sns_caller_arn", iamRoleResourceName, "arn"),
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
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.0.external_id", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "sms_configuration.0.sns_caller_arn", iamRoleResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_smsVerificationMessage(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	smsVerificationMessage1 := "test verification message {####}"
	smsVerificationMessage2 := "test verification message updated {####}"
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_smsVerificationMessage(rName, smsVerificationMessage1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(resourceName, nil),
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
					testAccCheckUserPoolExists(resourceName, nil),
					resource.TestCheckResourceAttr(resourceName, "sms_verification_message", smsVerificationMessage2),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_withEmail(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_emailConfiguration(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "email_configuration.#", "1"),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	replyTo := acctest.DefaultEmailAddress
	resourceName := "aws_cognito_user_pool.test"
	resourceName2 := "aws_ses_configuration_set.test"

	sourceARN, ok := os.LookupEnv("TEST_AWS_SES_VERIFIED_EMAIL_ARN")
	if !ok {
		t.Skip("'TEST_AWS_SES_VERIFIED_EMAIL_ARN' not set, skipping test.")
	}
	emailTo := sourceARN[strings.LastIndex(sourceARN, "/")+1:]

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_emailConfigurationSource(rName, replyTo, sourceARN, emailTo, "DEVELOPER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "email_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "email_configuration.0.reply_to_email_address", replyTo),
					resource.TestCheckResourceAttr(resourceName, "email_configuration.0.email_sending_account", "DEVELOPER"),
					resource.TestCheckResourceAttr(resourceName, "email_configuration.0.source_arn", sourceARN),
					resource.TestCheckResourceAttr(resourceName, "email_configuration.0.from_email_address", emailTo),
					resource.TestCheckResourceAttrPair(resourceName, "email_configuration.0.configuration_set", resourceName2, "name"),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_withTags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(resourceName, nil),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccUserPoolConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(resourceName, nil),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_withAliasAttributes(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_aliasAttributes(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(resourceName, nil),
					resource.TestCheckResourceAttr(resourceName, "alias_attributes.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "alias_attributes.*", "preferred_username"),
					resource.TestCheckResourceAttr(resourceName, "auto_verified_attributes.#", "0"),
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
					resource.TestCheckResourceAttr(resourceName, "alias_attributes.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "alias_attributes.*", "email"),
					resource.TestCheckTypeSetElemAttr(resourceName, "alias_attributes.*", "preferred_username"),
					resource.TestCheckResourceAttr(resourceName, "auto_verified_attributes.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "auto_verified_attributes.*", "email"),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_withUsernameAttributes(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_nameAttributes(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(resourceName, nil),
					resource.TestCheckResourceAttr(resourceName, "username_attributes.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "username_attributes.*", "phone_number"),
					resource.TestCheckResourceAttr(resourceName, "auto_verified_attributes.#", "0"),
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
					resource.TestCheckResourceAttr(resourceName, "username_attributes.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "username_attributes.*", "email"),
					resource.TestCheckTypeSetElemAttr(resourceName, "username_attributes.*", "phone_number"),
					resource.TestCheckResourceAttr(resourceName, "auto_verified_attributes.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "auto_verified_attributes.*", "email"),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_withPasswordPolicy(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_passwordPolicy(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(resourceName, nil),
					resource.TestCheckResourceAttr(resourceName, "password_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.minimum_length", "7"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.require_lowercase", "true"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.require_numbers", "false"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.require_symbols", "true"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.require_uppercase", "false"),
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
					resource.TestCheckResourceAttr(resourceName, "password_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.minimum_length", "9"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.require_lowercase", "false"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.require_numbers", "true"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.require_symbols", "false"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.require_uppercase", "true"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.temporary_password_validity_days", "14"),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_withUsername(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_nameConfiguration(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(resourceName, nil),
					resource.TestCheckResourceAttr(resourceName, "username_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "username_configuration.0.case_sensitive", "true"),
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
					testAccCheckUserPoolExists(resourceName, nil),
					resource.TestCheckResourceAttr(resourceName, "username_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "username_configuration.0.case_sensitive", "false"),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_withLambda(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"
	lambdaResourceName := "aws_lambda_function.test"
	lambdaUpdatedResourceName := "aws_lambda_function.second"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_lambda(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(resourceName, nil),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.create_auth_challenge", lambdaResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.custom_message", lambdaResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.define_auth_challenge", lambdaResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.post_authentication", lambdaResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.post_confirmation", lambdaResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.pre_authentication", lambdaResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.pre_sign_up", lambdaResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.pre_token_generation", lambdaResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.user_migration", lambdaResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.verify_auth_challenge_response", lambdaResourceName, "arn"),
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
					resource.TestCheckResourceAttr(resourceName, "lambda_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.create_auth_challenge", lambdaUpdatedResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.custom_message", lambdaUpdatedResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.define_auth_challenge", lambdaUpdatedResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.post_authentication", lambdaUpdatedResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.post_confirmation", lambdaUpdatedResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.pre_authentication", lambdaUpdatedResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.pre_sign_up", lambdaUpdatedResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.pre_token_generation", lambdaUpdatedResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.user_migration", lambdaUpdatedResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.verify_auth_challenge_response", lambdaUpdatedResourceName, "arn"),
				),
			},
			{
				Config: testAccUserPoolConfig_name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(resourceName, nil),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.#", "0"),
				),
			},
		},
	})
}

func testAccCheckUserPoolNotRecreated(pool1, pool2 *cognitoidentityprovider.DescribeUserPoolOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.TimeValue(pool1.UserPool.CreationDate).Equal(aws.TimeValue(pool2.UserPool.CreationDate)) {
			return fmt.Errorf("user pool was recreated. expected: %s, got: %s", pool1.UserPool.CreationDate, pool2.UserPool.CreationDate)
		}
		return nil
	}
}

func TestAccCognitoIDPUserPool_WithLambda_email(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"
	lambdaResourceName := "aws_lambda_function.test"
	lambdaUpdatedResourceName := "aws_lambda_function.second"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_lambdaEmailSender(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(resourceName, nil),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.0.custom_email_sender.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.custom_email_sender.0.lambda_arn", lambdaResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.0.custom_email_sender.0.lambda_version", "V1_0"),
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
					resource.TestCheckResourceAttr(resourceName, "lambda_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.0.custom_email_sender.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.custom_email_sender.0.lambda_arn", lambdaUpdatedResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.0.custom_email_sender.0.lambda_version", "V1_0"),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_WithLambda_sms(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"
	lambdaResourceName := "aws_lambda_function.test"
	lambdaUpdatedResourceName := "aws_lambda_function.second"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_lambdaSMSSender(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(resourceName, nil),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.0.custom_sms_sender.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.custom_sms_sender.0.lambda_arn", lambdaResourceName, "arn"),
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
					resource.TestCheckResourceAttr(resourceName, "lambda_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.0.custom_sms_sender.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.custom_sms_sender.0.lambda_arn", lambdaUpdatedResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.0.custom_sms_sender.0.lambda_version", "V1_0"),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_schemaAttributes(t *testing.T) {
	var pool1, pool2 cognitoidentityprovider.DescribeUserPoolOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_schemaAttributes(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(resourceName, &pool1),
					resource.TestCheckResourceAttr(resourceName, "schema.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema.*", map[string]string{
						"attribute_data_type":            "String",
						"developer_only_attribute":       "false",
						"mutable":                        "false",
						"name":                           "email",
						"number_attribute_constraints.#": "0",
						"required":                       "true",
						"string_attribute_constraints.#": "1",
						"string_attribute_constraints.0.min_length": "5",
						"string_attribute_constraints.0.max_length": "10",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema.*", map[string]string{
						"attribute_data_type":            "Boolean",
						"developer_only_attribute":       "true",
						"mutable":                        "false",
						"name":                           "mybool",
						"number_attribute_constraints.#": "0",
						"required":                       "false",
						"string_attribute_constraints.#": "0",
					}),
				),
			},
			{
				Config: testAccUserPoolConfig_schemaAttributesUpdated(rName, "mybool"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(resourceName, &pool2),
					testAccCheckUserPoolNotRecreated(&pool1, &pool2),
					resource.TestCheckResourceAttr(resourceName, "schema.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema.*", map[string]string{
						"attribute_data_type":            "String",
						"developer_only_attribute":       "false",
						"mutable":                        "false",
						"name":                           "email",
						"number_attribute_constraints.#": "0",
						"required":                       "true",
						"string_attribute_constraints.#": "1",
						"string_attribute_constraints.0.min_length": "5",
						"string_attribute_constraints.0.max_length": "10",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema.*", map[string]string{
						"attribute_data_type":            "Boolean",
						"developer_only_attribute":       "true",
						"mutable":                        "false",
						"name":                           "mybool",
						"number_attribute_constraints.#": "0",
						"required":                       "false",
						"string_attribute_constraints.#": "0",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema.*", map[string]string{
						"attribute_data_type":            "Number",
						"developer_only_attribute":       "false",
						"mutable":                        "true",
						"name":                           "mynondevnumber",
						"number_attribute_constraints.#": "1",
						"number_attribute_constraints.0.min_value": "2",
						"number_attribute_constraints.0.max_value": "6",
						"required":                       "false",
						"string_attribute_constraints.#": "0",
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_schemaAttributesUpdated(rName, "mybool"),
			},
			{
				Config:      testAccUserPoolConfig_schemaAttributes(rName),
				ExpectError: regexp.MustCompile("cannot modify or remove schema items"),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_schemaAttributesModified(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_schemaAttributesUpdated(rName, "mybool"),
			},
			{
				Config:      testAccUserPoolConfig_schemaAttributesUpdated(rName, "mybool2"),
				ExpectError: regexp.MustCompile("cannot modify or remove schema items"),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_withVerificationMessageTemplate(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_verificationMessageTemplate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(resourceName, nil),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.default_email_option", "CONFIRM_WITH_LINK"),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.email_message", "foo {####} bar"),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.email_message_by_link", "{##foobar##}"),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.email_subject", "foobar {####}"),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.email_subject_by_link", "foobar"),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.sms_message", "{####} baz"),

					/* Setting Verification template attributes like EmailMessage, EmailSubject or SmsMessage
					will implicitly set EmailVerificationMessage, EmailVerificationSubject and SmsVerificationMessage attributes.
					*/
					resource.TestCheckResourceAttr(resourceName, "email_verification_message", "foo {####} bar"),
					resource.TestCheckResourceAttr(resourceName, "email_verification_subject", "foobar {####}"),
					resource.TestCheckResourceAttr(resourceName, "sms_verification_message", "{####} baz"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolConfig_verificationMessageTemplateDefaultEmailOption(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.default_email_option", "CONFIRM_WITH_CODE"),
					resource.TestCheckResourceAttr(resourceName, "email_verification_message", "{####} Baz"),
					resource.TestCheckResourceAttr(resourceName, "email_verification_subject", "BazBaz {####}"),
					resource.TestCheckResourceAttr(resourceName, "sms_verification_message", "{####} BazBazBar?"),

					/* Setting EmailVerificationMessage, EmailVerificationSubject and SmsVerificationMessage attributes
					will implicitly set verification template attributes like EmailMessage, EmailSubject or SmsMessage.
					*/
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.email_message", "{####} Baz"),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.email_subject", "BazBaz {####}"),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.sms_message", "{####} BazBazBar?"),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_update(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	optionalMfa := "OPTIONAL"
	offMfa := "OFF"
	authenticationMessage := fmt.Sprintf("%s {####}", sdkacctest.RandString(10))
	updatedAuthenticationMessage := fmt.Sprintf("%s {####}", sdkacctest.RandString(10))
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_update(rName, optionalMfa, authenticationMessage),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(resourceName, nil),
					resource.TestCheckResourceAttr(resourceName, "auto_verified_attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", optionalMfa),
					resource.TestCheckResourceAttr(resourceName, "email_verification_message", "Foo {####} Bar"),
					resource.TestCheckResourceAttr(resourceName, "email_verification_subject", "FooBar {####}"),
					resource.TestCheckResourceAttr(resourceName, "sms_verification_message", "{####} Baz"),
					resource.TestCheckResourceAttr(resourceName, "sms_authentication_message", authenticationMessage),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.allow_admin_create_user_only", "true"),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.email_message", "Your username is {username} and temporary password is {####}. "),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.email_subject", "FooBar {####}"),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.sms_message", "Your username is {username} and temporary password is {####}."),
					resource.TestCheckResourceAttr(resourceName, "device_configuration.0.challenge_required_on_new_device", "true"),
					resource.TestCheckResourceAttr(resourceName, "device_configuration.0.device_only_remembered_on_user_prompt", "false"),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.default_email_option", "CONFIRM_WITH_CODE"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "1"),
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
					testAccCheckUserPoolExists(resourceName, nil),
					resource.TestCheckResourceAttr(resourceName, "auto_verified_attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", optionalMfa),
					resource.TestCheckResourceAttr(resourceName, "email_verification_message", "Foo {####} Bar"),
					resource.TestCheckResourceAttr(resourceName, "email_verification_subject", "FooBar {####}"),
					resource.TestCheckResourceAttr(resourceName, "sms_verification_message", "{####} Baz"),
					resource.TestCheckResourceAttr(resourceName, "sms_authentication_message", updatedAuthenticationMessage),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.allow_admin_create_user_only", "true"),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.email_message", "Your username is {username} and temporary password is {####}. "),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.email_subject", "FooBar {####}"),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.sms_message", "Your username is {username} and temporary password is {####}."),
					resource.TestCheckResourceAttr(resourceName, "device_configuration.0.challenge_required_on_new_device", "true"),
					resource.TestCheckResourceAttr(resourceName, "device_configuration.0.device_only_remembered_on_user_prompt", "false"),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.default_email_option", "CONFIRM_WITH_CODE"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "sms_configuration.0.external_id"),
					resource.TestCheckResourceAttrSet(resourceName, "sms_configuration.0.sns_caller_arn"),
				),
			},
			{
				Config: testAccUserPoolConfig_update(rName, offMfa, updatedAuthenticationMessage),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(resourceName, nil),
					resource.TestCheckResourceAttr(resourceName, "auto_verified_attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", offMfa),
					resource.TestCheckResourceAttr(resourceName, "email_verification_message", "Foo {####} Bar"),
					resource.TestCheckResourceAttr(resourceName, "email_verification_subject", "FooBar {####}"),
					resource.TestCheckResourceAttr(resourceName, "sms_verification_message", "{####} Baz"),
					resource.TestCheckResourceAttr(resourceName, "sms_authentication_message", updatedAuthenticationMessage),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.allow_admin_create_user_only", "true"),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.email_message", "Your username is {username} and temporary password is {####}. "),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.email_subject", "FooBar {####}"),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.sms_message", "Your username is {username} and temporary password is {####}."),
					resource.TestCheckResourceAttr(resourceName, "device_configuration.0.challenge_required_on_new_device", "true"),
					resource.TestCheckResourceAttr(resourceName, "device_configuration.0.device_only_remembered_on_user_prompt", "false"),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.default_email_option", "CONFIRM_WITH_CODE"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "sms_configuration.0.external_id"),
					resource.TestCheckResourceAttrSet(resourceName, "sms_configuration.0.sns_caller_arn"),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPool_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolConfig_name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(resourceName, nil),
					acctest.CheckResourceDisappears(acctest.Provider, tfcognitoidp.ResourceUserPool(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckUserPoolDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cognito_user_pool" {
			continue
		}

		params := &cognitoidentityprovider.DescribeUserPoolInput{
			UserPoolId: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeUserPool(params)

		if err != nil {
			if tfawserr.ErrCodeEquals(err, cognitoidentityprovider.ErrCodeResourceNotFoundException) {
				return nil
			}
			return err
		}
	}

	return nil
}

func testAccCheckUserPoolExists(name string, pool *cognitoidentityprovider.DescribeUserPoolOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Cognito User Pool ID set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPConn

		params := &cognitoidentityprovider.DescribeUserPoolInput{
			UserPoolId: aws.String(rs.Primary.ID),
		}

		poolOut, err := conn.DescribeUserPool(params)
		if err != nil {
			return err
		}

		if pool != nil {
			*pool = *poolOut
		}

		return nil
	}
}

func testAccPreCheckIdentityProvider(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPConn

	input := &cognitoidentityprovider.ListUserPoolsInput{
		MaxResults: aws.Int64(1),
	}

	_, err := conn.ListUserPools(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccUserPoolSMSConfigurationBaseConfig(rName string, externalID string) string {
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
	return testAccUserPoolSMSConfigurationBaseConfig(rName, "test") + fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  mfa_configuration = "ON"
  name              = %[1]q

  sms_configuration {
    external_id    = "test"
    sns_caller_arn = aws_iam_role.test.arn
  }
}
`, rName)
}

func testAccUserPoolConfig_mfaConfigurationSMSConfigurationAndSoftwareTokenMFAConfigurationEnabled(rName string, enabled bool) string {
	return testAccUserPoolSMSConfigurationBaseConfig(rName, "test") + fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  mfa_configuration = "ON"
  name              = %[1]q

  sms_configuration {
    external_id    = "test"
    sns_caller_arn = aws_iam_role.test.arn
  }

  software_token_mfa_configuration {
    enabled = %[2]t
  }
}
`, rName, enabled)
}

func testAccUserPoolConfig_mfaConfigurationSoftwareTokenMFAConfigurationEnabled(rName string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  mfa_configuration = "ON"
  name              = %[1]q

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
	return testAccUserPoolSMSConfigurationBaseConfig(rName, externalID) + fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  sms_configuration {
    external_id    = %[2]q
    sns_caller_arn = aws_iam_role.test.arn
  }
}
`, rName, externalID)
}

func testAccUserPoolConfig_smsConfigurationSNSCallerARN2(rName string) string {
	return testAccUserPoolSMSConfigurationBaseConfig(rName+"-2", "test") + fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  sms_configuration {
    external_id    = "test"
    sns_caller_arn = aws_iam_role.test.arn
  }
}
`, rName)
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

func testAccUserPoolLambdaBaseConfig(name string) string {
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
  runtime       = "nodejs12.x"
}

resource "aws_kms_key" "test" {
  description             = "Terraform acc test %[1]s"
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
	return testAccUserPoolLambdaBaseConfig(name) + fmt.Sprintf(`
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
`, name)
}

func testAccUserPoolConfig_lambdaUpdated(name string) string {
	return testAccUserPoolLambdaBaseConfig(name) + fmt.Sprintf(`
resource "aws_lambda_function" "second" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s_second"
  role          = aws_iam_role.test.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
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
`, name)
}

func testAccUserPoolConfig_lambdaEmailSender(name string) string {
	return testAccUserPoolLambdaBaseConfig(name) + fmt.Sprintf(`
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
`, name)
}

func testAccUserPoolConfig_lambdaEmailSenderUpdated(name string) string {
	return testAccUserPoolLambdaBaseConfig(name) + fmt.Sprintf(`
resource "aws_lambda_function" "second" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s_second"
  role          = aws_iam_role.test.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
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
`, name)
}

func testAccUserPoolConfig_lambdaSMSSender(name string) string {
	return testAccUserPoolLambdaBaseConfig(name) + fmt.Sprintf(`
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
`, name)
}

func testAccUserPoolConfig_lambdaSMSSenderUpdated(name string) string {
	return testAccUserPoolLambdaBaseConfig(name) + fmt.Sprintf(`
resource "aws_lambda_function" "second" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s_second"
  role          = aws_iam_role.test.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
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
`, name)
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

func testAccUserPoolConfig_verificationMessageTemplate(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  # Setting Verification template attributes like EmailMessage, EmailSubject or SmsMessage
  # will implicitly set EmailVerificationMessage, EmailVerificationSubject and SmsVerificationMessage
  # attributes.

  verification_message_template {
    default_email_option  = "CONFIRM_WITH_LINK"
    email_message         = "foo {####} bar"
    email_message_by_link = "{##foobar##}"
    email_subject         = "foobar {####}"
    email_subject_by_link = "foobar"
    sms_message           = "{####} baz"
  }
}
`, name)
}

func testAccUserPoolConfig_verificationMessageTemplateDefaultEmailOption(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  email_verification_message = "{####} Baz"
  email_verification_subject = "BazBaz {####}"
  sms_verification_message   = "{####} BazBazBar?"

  verification_message_template {
    default_email_option = "CONFIRM_WITH_CODE"
  }
}
`, name)
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
