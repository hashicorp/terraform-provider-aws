// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ses_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ses"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESSendEmailAction_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	testEmail := acctest.SkipIfEnvVarNotSet(t, "SES_VERIFIED_EMAIL")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SESEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		CheckDestroy: acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccSendEmailActionConfig_basic(rName, testEmail),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendEmailAction(ctx, t, testEmail),
				),
			},
		},
	})
}

func TestAccSESSendEmailAction_htmlBody(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	testEmail := acctest.SkipIfEnvVarNotSet(t, "SES_VERIFIED_EMAIL")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SESEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		CheckDestroy: acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccSendEmailActionConfig_htmlBody(rName, testEmail),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendEmailAction(ctx, t, testEmail),
				),
			},
		},
	})
}

func TestAccSESSendEmailAction_multipleRecipients(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	testEmail := acctest.SkipIfEnvVarNotSet(t, "SES_VERIFIED_EMAIL")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SESEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		CheckDestroy: acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccSendEmailActionConfig_multipleRecipients(rName, testEmail),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendEmailAction(ctx, t, testEmail),
				),
			},
		},
	})
}

// testAccCheckSendEmailAction verifies the action can send emails
func testAccCheckSendEmailAction(ctx context.Context, t *testing.T, sourceEmail string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SESClient(ctx)

		// Verify the source email is verified in SES
		input := &ses.GetIdentityVerificationAttributesInput{
			Identities: []string{sourceEmail},
		}

		output, err := conn.GetIdentityVerificationAttributes(ctx, input)
		if err != nil {
			return fmt.Errorf("Failed to get identity verification attributes: %w", err)
		}

		if attrs, ok := output.VerificationAttributes[sourceEmail]; ok {
			if attrs.VerificationStatus != awstypes.VerificationStatusSuccess {
				return fmt.Errorf("Email %s is not verified in SES (status: %s)", sourceEmail, string(attrs.VerificationStatus))
			}
		} else {
			return fmt.Errorf("Email %s not found in SES identities", sourceEmail)
		}

		return nil
	}
}

// Configuration functions

func testAccSendEmailActionConfig_basic(rName, testEmail string) string {
	return fmt.Sprintf(`
action "aws_ses_send_email" "test" {
  config {
    source       = %[2]q
    subject      = "Test Email from %[1]s"
    text_body    = "This is a test email sent from Terraform action test."
    to_addresses = [%[2]q]
  }
}

resource "terraform_data" "trigger" {
  input = "trigger"
  lifecycle {
    action_trigger {
      events  = [before_create]
      actions = [action.aws_ses_send_email.test]
    }
  }
}
`, rName, testEmail)
}

func testAccSendEmailActionConfig_htmlBody(rName, testEmail string) string {
	return fmt.Sprintf(`
action "aws_ses_send_email" "test" {
  config {
    source       = %[2]q
    subject      = "HTML Test Email from %[1]s"
    html_body    = "<h1>Test Email</h1><p>This is a <strong>test email</strong> sent from Terraform action test.</p>"
    to_addresses = [%[2]q]
  }
}

resource "terraform_data" "trigger" {
  input = "trigger"
  lifecycle {
    action_trigger {
      events  = [before_create]
      actions = [action.aws_ses_send_email.test]
    }
  }
}
`, rName, testEmail)
}

func testAccSendEmailActionConfig_multipleRecipients(rName, testEmail string) string {
	return fmt.Sprintf(`
action "aws_ses_send_email" "test" {
  config {
    source             = %[2]q
    subject            = "Multi-recipient Test Email from %[1]s"
    text_body          = "This is a test email sent to multiple recipients."
    to_addresses       = [%[2]q]
    cc_addresses       = [%[2]q]
    reply_to_addresses = [%[2]q]
  }
}

resource "terraform_data" "trigger" {
  input = "trigger"
  lifecycle {
    action_trigger {
      events  = [before_create]
      actions = [action.aws_ses_send_email.test]
    }
  }
}
`, rName, testEmail)
}
