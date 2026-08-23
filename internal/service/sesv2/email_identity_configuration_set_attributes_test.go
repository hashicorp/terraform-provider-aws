// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sesv2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsesv2 "github.com/hashicorp/terraform-provider-aws/internal/service/sesv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESV2EmailIdentityConfigurationSetAttributes_basic(t *testing.T) {
	ctx := acctest.Context(t)
	domain := acctest.RandomDomainName()
	configSetName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sesv2_email_identity_configuration_set_attributes.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailIdentityConfigurationSetAttributesDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityConfigurationSetAttributesConfig_basic(domain, configSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEmailIdentityConfigurationSetAttributesExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "email_identity", "aws_sesv2_email_identity.test", "email_identity"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration_set_name", "aws_sesv2_configuration_set.test", "configuration_set_name"),
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

func TestAccSESV2EmailIdentityConfigurationSetAttributes_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	domain := acctest.RandomDomainName()
	configSetName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sesv2_email_identity_configuration_set_attributes.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailIdentityConfigurationSetAttributesDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityConfigurationSetAttributesConfig_basic(domain, configSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEmailIdentityConfigurationSetAttributesExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfsesv2.ResourceEmailIdentityConfigurationSetAttributes, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSESV2EmailIdentityConfigurationSetAttributes_disappearsEmailIdentity(t *testing.T) {
	ctx := acctest.Context(t)
	domain := acctest.RandomDomainName()
	configSetName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sesv2_email_identity_configuration_set_attributes.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailIdentityConfigurationSetAttributesDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityConfigurationSetAttributesConfig_basic(domain, configSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEmailIdentityConfigurationSetAttributesExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfsesv2.ResourceEmailIdentity(), "aws_sesv2_email_identity.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSESV2EmailIdentityConfigurationSetAttributes_update(t *testing.T) {
	ctx := acctest.Context(t)
	domain := acctest.RandomDomainName()
	configSetName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	configSetName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sesv2_email_identity_configuration_set_attributes.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailIdentityConfigurationSetAttributesDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityConfigurationSetAttributesConfig_basic(domain, configSetName1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEmailIdentityConfigurationSetAttributesExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration_set_name", configSetName1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEmailIdentityConfigurationSetAttributesConfig_updated(domain, configSetName1, configSetName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEmailIdentityConfigurationSetAttributesExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration_set_name", configSetName2),
				),
			},
		},
	})
}

func testAccCheckEmailIdentityConfigurationSetAttributesDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SESV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sesv2_email_identity_configuration_set_attributes" {
				continue
			}

			out, err := tfsesv2.FindEmailIdentityByID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return err
			}
			if out.ConfigurationSetName == nil || *out.ConfigurationSetName == "" {
				return nil
			}
			return fmt.Errorf("SESv2 Email Identity Configuration Set Attributes %s still configured with %s", rs.Primary.ID, *out.ConfigurationSetName)
		}

		return nil
	}
}

func testAccCheckEmailIdentityConfigurationSetAttributesExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SESV2Client(ctx)

		out, err := tfsesv2.FindEmailIdentityByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}
		if out.ConfigurationSetName == nil || *out.ConfigurationSetName == "" {
			return fmt.Errorf("SESv2 Email Identity Configuration Set Attributes %s has no configuration set", rs.Primary.ID)
		}

		return nil
	}
}

func testAccEmailIdentityConfigurationSetAttributesConfig_basic(domain, configSetName string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_email_identity" "test" {
  email_identity = %[1]q
}

resource "aws_sesv2_configuration_set" "test" {
  configuration_set_name = %[2]q
}

resource "aws_sesv2_email_identity_configuration_set_attributes" "test" {
  email_identity         = aws_sesv2_email_identity.test.email_identity
  configuration_set_name = aws_sesv2_configuration_set.test.configuration_set_name
}
`, domain, configSetName)
}

func testAccEmailIdentityConfigurationSetAttributesConfig_updated(domain, configSetName1, configSetName2 string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_email_identity" "test" {
  email_identity = %[1]q
}

resource "aws_sesv2_configuration_set" "test" {
  configuration_set_name = %[2]q
}

resource "aws_sesv2_configuration_set" "test2" {
  configuration_set_name = %[3]q
}

resource "aws_sesv2_email_identity_configuration_set_attributes" "test" {
  email_identity         = aws_sesv2_email_identity.test.email_identity
  configuration_set_name = aws_sesv2_configuration_set.test2.configuration_set_name
}
`, domain, configSetName1, configSetName2)
}
