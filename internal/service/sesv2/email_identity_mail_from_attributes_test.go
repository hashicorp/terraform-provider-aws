// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sesv2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfsesv2 "github.com/hashicorp/terraform-provider-aws/internal/service/sesv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESV2EmailIdentityMailFromAttributes_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomDomainName()
	resourceName := "aws_sesv2_email_identity_mail_from_attributes.test"
	emailIdentityName := "aws_sesv2_email_identity.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailIdentityDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityMailFromAttributesConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityMailFromAttributesExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "email_identity", emailIdentityName, "email_identity"),
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

func TestAccSESV2EmailIdentityMailFromAttributes_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	domain := acctest.RandomDomain()
	mailFromDomain := domain.Subdomain("test")

	rName := domain.String()
	resourceName := "aws_sesv2_email_identity_mail_from_attributes.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailIdentityDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityMailFromAttributesConfig_behaviorOnMXFailureAndMailFromDomain(rName, string(types.BehaviorOnMxFailureUseDefaultValue), mailFromDomain.String()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityMailFromAttributesExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfsesv2.ResourceEmailIdentityMailFromAttributes(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSESV2EmailIdentityMailFromAttributes_disappearsEmailIdentity(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomDomainName()
	resourceName := "aws_sesv2_email_identity_mail_from_attributes.test"
	emailIdentityName := "aws_sesv2_email_identity.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailIdentityDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityMailFromAttributesConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityMailFromAttributesExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfsesv2.ResourceEmailIdentity(), emailIdentityName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSESV2EmailIdentityMailFromAttributes_behaviorOnMXFailure(t *testing.T) {
	ctx := acctest.Context(t)
	domain := acctest.RandomDomain()
	mailFromDomain := domain.Subdomain("test")

	rName := domain.String()
	resourceName := "aws_sesv2_email_identity_mail_from_attributes.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailIdentityDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityMailFromAttributesConfig_behaviorOnMXFailureAndMailFromDomain(rName, string(types.BehaviorOnMxFailureUseDefaultValue), mailFromDomain.String()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityMailFromAttributesExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "behavior_on_mx_failure", string(types.BehaviorOnMxFailureUseDefaultValue)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEmailIdentityMailFromAttributesConfig_behaviorOnMXFailureAndMailFromDomain(rName, string(types.BehaviorOnMxFailureRejectMessage), mailFromDomain.String()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityMailFromAttributesExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "behavior_on_mx_failure", string(types.BehaviorOnMxFailureRejectMessage)),
				),
			},
		},
	})
}

func TestAccSESV2EmailIdentityMailFromAttributes_mailFromDomain(t *testing.T) {
	ctx := acctest.Context(t)
	domain := acctest.RandomDomain()
	mailFromDomain1 := domain.Subdomain("test1")
	mailFromDomain2 := domain.Subdomain("test2")

	rName := domain.String()
	resourceName := "aws_sesv2_email_identity_mail_from_attributes.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailIdentityDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityMailFromAttributesConfig_behaviorOnMXFailureAndMailFromDomain(rName, string(types.BehaviorOnMxFailureUseDefaultValue), mailFromDomain1.String()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityMailFromAttributesExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "mail_from_domain", mailFromDomain1.String()),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEmailIdentityMailFromAttributesConfig_behaviorOnMXFailureAndMailFromDomain(rName, string(types.BehaviorOnMxFailureUseDefaultValue), mailFromDomain2.String()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityMailFromAttributesExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "mail_from_domain", mailFromDomain2.String()),
				),
			},
		},
	})
}

func testAccCheckEmailIdentityMailFromAttributesExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SESV2Client(ctx)

		_, err := tfsesv2.FindEmailIdentityByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccEmailIdentityMailFromAttributesConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_email_identity" "test" {
  email_identity = %[1]q
}

resource "aws_sesv2_email_identity_mail_from_attributes" "test" {
  email_identity = aws_sesv2_email_identity.test.email_identity
}
`, rName)
}

func testAccEmailIdentityMailFromAttributesConfig_behaviorOnMXFailureAndMailFromDomain(rName, behaviorOnMXFailure, mailFromDomain string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_email_identity" "test" {
  email_identity = %[1]q
}

resource "aws_sesv2_email_identity_mail_from_attributes" "test" {
  email_identity         = aws_sesv2_email_identity.test.email_identity
  behavior_on_mx_failure = %[2]q
  mail_from_domain       = %[3]q
}
`, rName, behaviorOnMXFailure, mailFromDomain)
}
