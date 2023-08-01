// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sesv2_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfsesv2 "github.com/hashicorp/terraform-provider-aws/internal/service/sesv2"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESV2EmailIdentity_basic_emailAddress(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomEmailAddress(acctest.RandomDomainName())
	resourceName := "aws_sesv2_email_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailIdentityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "email_identity", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ses", regexp.MustCompile(`identity/.+`)),
					resource.TestCheckResourceAttr(resourceName, "dkim_signing_attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dkim_signing_attributes.0.current_signing_key_length", ""),
					resource.TestCheckResourceAttr(resourceName, "dkim_signing_attributes.0.last_key_generation_timestamp", ""),
					resource.TestCheckResourceAttr(resourceName, "dkim_signing_attributes.0.next_signing_key_length", "RSA_1024_BIT"),
					resource.TestCheckResourceAttr(resourceName, "dkim_signing_attributes.0.signing_attributes_origin", "AWS_SES"),
					resource.TestCheckResourceAttr(resourceName, "dkim_signing_attributes.0.status", "NOT_STARTED"),
					resource.TestCheckResourceAttr(resourceName, "dkim_signing_attributes.0.tokens.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "identity_type", "EMAIL_ADDRESS"),
					resource.TestCheckResourceAttr(resourceName, "verified_for_sending_status", "false"),
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

func TestAccSESV2EmailIdentity_basic_domain(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomDomainName()
	resourceName := "aws_sesv2_email_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailIdentityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "email_identity", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ses", regexp.MustCompile(`identity/.+`)),
					resource.TestCheckResourceAttr(resourceName, "dkim_signing_attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dkim_signing_attributes.0.current_signing_key_length", "RSA_2048_BIT"),
					acctest.CheckResourceAttrRFC3339(resourceName, "dkim_signing_attributes.0.last_key_generation_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "dkim_signing_attributes.0.next_signing_key_length", "RSA_2048_BIT"),
					resource.TestCheckResourceAttr(resourceName, "dkim_signing_attributes.0.signing_attributes_origin", "AWS_SES"),
					resource.TestCheckResourceAttr(resourceName, "dkim_signing_attributes.0.status", "PENDING"),
					resource.TestCheckResourceAttr(resourceName, "dkim_signing_attributes.0.tokens.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "identity_type", "DOMAIN"),
					resource.TestCheckResourceAttr(resourceName, "verified_for_sending_status", "false"),
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

func TestAccSESV2EmailIdentity_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomEmailAddress(acctest.RandomDomainName())
	resourceName := "aws_sesv2_email_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailIdentityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsesv2.ResourceEmailIdentity(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSESV2EmailIdentity_configurationSetName(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomEmailAddress(acctest.RandomDomainName())
	resourceName := "aws_sesv2_email_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailIdentityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityConfig_configurationSetName1(rName, sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "configuration_set_name", "aws_sesv2_configuration_set.test1", "configuration_set_name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEmailIdentityConfig_configurationSetName2(rName, sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "configuration_set_name", "aws_sesv2_configuration_set.test2", "configuration_set_name"),
				),
			},
		},
	})
}

func TestAccSESV2EmailIdentity_nextSigningKeyLength(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomDomainName()
	resourceName := "aws_sesv2_email_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailIdentityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityConfig_nextSigningKeyLength(rName, string(types.DkimSigningKeyLengthRsa2048Bit)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "dkim_signing_attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dkim_signing_attributes.0.next_signing_key_length", "RSA_2048_BIT"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEmailIdentityConfig_nextSigningKeyLength(rName, string(types.DkimSigningKeyLengthRsa1024Bit)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "dkim_signing_attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dkim_signing_attributes.0.next_signing_key_length", "RSA_1024_BIT"),
				),
			},
		},
	})
}

func TestAccSESV2EmailIdentity_domainSigning(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomDomainName()
	resourceName := "aws_sesv2_email_identity.test"

	key1 := verify.Base64Encode([]byte(acctest.TLSRSAPrivateKeyPEM(t, 2048)))
	selector1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	key2 := verify.Base64Encode([]byte(acctest.TLSRSAPrivateKeyPEM(t, 2048)))
	selector2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailIdentityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityConfig_domainSigning(rName, key1, selector1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "dkim_signing_attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dkim_signing_attributes.0.domain_signing_private_key", key1),
					resource.TestCheckResourceAttr(resourceName, "dkim_signing_attributes.0.domain_signing_selector", selector1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"dkim_signing_attributes.0.domain_signing_private_key", "dkim_signing_attributes.0.domain_signing_selector"},
			},
			{
				Config: testAccEmailIdentityConfig_domainSigning(rName, key2, selector2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "dkim_signing_attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dkim_signing_attributes.0.domain_signing_private_key", key2),
					resource.TestCheckResourceAttr(resourceName, "dkim_signing_attributes.0.domain_signing_selector", selector2),
				),
			},
		},
	})
}

func TestAccSESV2EmailIdentity_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomEmailAddress(acctest.RandomDomainName())
	resourceName := "aws_sesv2_email_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailIdentityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityExists(ctx, resourceName),
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
				Config: testAccEmailIdentityConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccEmailIdentityConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckEmailIdentityDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sesv2_email_identity" {
				continue
			}

			_, err := tfsesv2.FindEmailIdentityByID(ctx, conn, rs.Primary.ID)

			if err != nil {
				var nfe *types.NotFoundException
				if errors.As(err, &nfe) {
					return nil
				}
				return err
			}

			return create.Error(names.SESV2, create.ErrActionCheckingDestroyed, tfsesv2.ResNameEmailIdentity, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckEmailIdentityExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SESV2, create.ErrActionCheckingExistence, tfsesv2.ResNameEmailIdentity, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.SESV2, create.ErrActionCheckingExistence, tfsesv2.ResNameEmailIdentity, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SESV2Client(ctx)

		_, err := tfsesv2.FindEmailIdentityByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.SESV2, create.ErrActionCheckingExistence, tfsesv2.ResNameEmailIdentity, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccEmailIdentityConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_email_identity" "test" {
  email_identity = %[1]q
}
`, rName)
}

func testAccEmailIdentityConfig_configurationSetName1(rName, configurationSetName string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_configuration_set" "test1" {
  configuration_set_name = %[2]q
}

resource "aws_sesv2_email_identity" "test" {
  email_identity         = %[1]q
  configuration_set_name = aws_sesv2_configuration_set.test1.configuration_set_name
}
`, rName, configurationSetName)
}

func testAccEmailIdentityConfig_configurationSetName2(rName, configurationSetName string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_configuration_set" "test2" {
  configuration_set_name = %[2]q
}

resource "aws_sesv2_email_identity" "test" {
  email_identity         = %[1]q
  configuration_set_name = aws_sesv2_configuration_set.test2.configuration_set_name
}
`, rName, configurationSetName)
}

func testAccEmailIdentityConfig_nextSigningKeyLength(rName, nextSigningKeyLength string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_email_identity" "test" {
  email_identity = %[1]q

  dkim_signing_attributes {
    next_signing_key_length = %[2]q
  }
}
`, rName, nextSigningKeyLength)
}

func testAccEmailIdentityConfig_domainSigning(rName, domainSigningPrivateKey, domainSigningSelector string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_email_identity" "test" {
  email_identity = %[1]q

  dkim_signing_attributes {
    domain_signing_private_key = %[2]q
    domain_signing_selector    = %[3]q
  }
}
`, rName, domainSigningPrivateKey, domainSigningSelector)
}

func testAccEmailIdentityConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_email_identity" "test" {
  email_identity = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccEmailIdentityConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_email_identity" "test" {
  email_identity = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
