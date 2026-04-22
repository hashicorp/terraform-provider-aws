// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package paymentcryptography_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/paymentcryptography"
	"github.com/aws/aws-sdk-go-v2/service/paymentcryptography/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfpaymentcryptography "github.com/hashicorp/terraform-provider-aws/internal/service/paymentcryptography"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccPaymentCryptographyKey_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var key paymentcryptography.GetKeyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_paymentcryptography_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PaymentCryptographyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, t, resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "exportable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "key_attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "key_attributes.0.key_algorithm", "TDES_3KEY"),
					resource.TestCheckResourceAttr(resourceName, "key_attributes.0.key_class", "SYMMETRIC_KEY"),
					resource.TestCheckResourceAttr(resourceName, "key_attributes.0.key_usage", "TR31_P0_PIN_ENCRYPTION_KEY"),
					resource.TestCheckResourceAttr(resourceName, "key_attributes.0.key_modes_of_use.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "key_attributes.0.key_modes_of_use.0.decrypt", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "key_attributes.0.key_modes_of_use.0.encrypt", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "key_attributes.0.key_modes_of_use.0.wrap", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "key_attributes.0.key_modes_of_use.0.unwrap", acctest.CtTrue),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "payment-cryptography", regexache.MustCompile(`key/[0-9a-z]{16}$`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"deletion_window_in_days"},
			},
		},
	})
}

func TestAccPaymentCryptographyKey_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var key1, key2 paymentcryptography.GetKeyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_paymentcryptography_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PaymentCryptographyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, t, resourceName, &key1),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Other", "Value"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "payment-cryptography", regexache.MustCompile(`key/.+`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"deletion_window_in_days"},
			},
			{
				Config: testAccKeyConfig_tags2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, t, resourceName, &key2),
					testAccCheckKeyNotRecreated(&key1, &key2),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name2", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Other", "Value2"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "payment-cryptography", regexache.MustCompile(`key/.+`)),
				),
			},
		},
	})
}

func TestAccPaymentCryptographyKey_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var key1, key2, key3 paymentcryptography.GetKeyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_paymentcryptography_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PaymentCryptographyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, t, resourceName, &key1),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "payment-cryptography", regexache.MustCompile(`key/.+`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"deletion_window_in_days"},
			},
			{
				Config: testAccKeyConfig_disable(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, t, resourceName, &key2),
					testAccCheckKeyNotRecreated(&key1, &key2),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "payment-cryptography", regexache.MustCompile(`key/.+`)),
				),
			},
			{
				Config: testAccKeyConfig_enable(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, t, resourceName, &key3),
					testAccCheckKeyNotRecreated(&key2, &key3),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "payment-cryptography", regexache.MustCompile(`key/.+`)),
				),
			},
		},
	})
}

func TestAccPaymentCryptographyKey_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var key paymentcryptography.GetKeyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_paymentcryptography_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PaymentCryptographyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, t, resourceName, &key),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfpaymentcryptography.ResourceKey, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccPaymentCryptographyKey_upgradeV6_0_0(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var key paymentcryptography.GetKeyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_paymentcryptography_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.PaymentCryptographyServiceID),
		CheckDestroy: testAccCheckKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.94.1",
					},
				},
				Config: testAccKeyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, t, resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "exportable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "key_attributes.key_algorithm", "TDES_3KEY"),
					resource.TestCheckResourceAttr(resourceName, "key_attributes.key_class", "SYMMETRIC_KEY"),
					resource.TestCheckResourceAttr(resourceName, "key_attributes.key_usage", "TR31_P0_PIN_ENCRYPTION_KEY"),
					resource.TestCheckResourceAttr(resourceName, "key_attributes.key_modes_of_use.decrypt", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "key_attributes.key_modes_of_use.encrypt", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "key_attributes.key_modes_of_use.wrap", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "key_attributes.key_modes_of_use.unwrap", acctest.CtTrue),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "payment-cryptography", regexache.MustCompile(`key/.+`)),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccKeyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, t, resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "exportable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "key_attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "key_attributes.0.key_algorithm", "TDES_3KEY"),
					resource.TestCheckResourceAttr(resourceName, "key_attributes.0.key_class", "SYMMETRIC_KEY"),
					resource.TestCheckResourceAttr(resourceName, "key_attributes.0.key_usage", "TR31_P0_PIN_ENCRYPTION_KEY"),
					resource.TestCheckResourceAttr(resourceName, "key_attributes.0.key_modes_of_use.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "key_attributes.0.key_modes_of_use.0.decrypt", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "key_attributes.0.key_modes_of_use.0.encrypt", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "key_attributes.0.key_modes_of_use.0.wrap", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "key_attributes.0.key_modes_of_use.0.unwrap", acctest.CtTrue),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "payment-cryptography", regexache.MustCompile(`key/.+`)),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccCheckKeyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).PaymentCryptographyClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_paymentcryptography_key" {
				continue
			}

			out, err := conn.GetKey(ctx, &paymentcryptography.GetKeyInput{
				KeyIdentifier: aws.String(rs.Primary.ID),
			})
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.PaymentCryptography, create.ErrActionCheckingDestroyed, tfpaymentcryptography.ResNameKey, rs.Primary.ID, err)
			}

			if state := out.Key.KeyState; state == types.KeyStateDeletePending || state == types.KeyStateDeleteComplete {
				return nil // Key is logically deleted
			}

			return create.Error(names.PaymentCryptography, create.ErrActionCheckingDestroyed, tfpaymentcryptography.ResNameKey, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckKeyExists(ctx context.Context, t *testing.T, name string, key *paymentcryptography.GetKeyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.PaymentCryptography, create.ErrActionCheckingExistence, tfpaymentcryptography.ResNameKey, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.PaymentCryptography, create.ErrActionCheckingExistence, tfpaymentcryptography.ResNameKey, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).PaymentCryptographyClient(ctx)
		resp, err := conn.GetKey(ctx, &paymentcryptography.GetKeyInput{
			KeyIdentifier: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.PaymentCryptography, create.ErrActionCheckingExistence, tfpaymentcryptography.ResNameKey, rs.Primary.ID, err)
		}

		*key = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).PaymentCryptographyClient(ctx)

	input := &paymentcryptography.ListKeysInput{}
	_, err := conn.ListKeys(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckKeyNotRecreated(before, after *paymentcryptography.GetKeyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.Key.KeyArn), aws.ToString(after.Key.KeyArn); before != after {
			return create.Error(names.PaymentCryptography, create.ErrActionCheckingNotRecreated, tfpaymentcryptography.ResNameKey, before, errors.New("recreated"))
		}

		return nil
	}
}

func testAccKeyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_paymentcryptography_key" "test" {
  exportable = true
  key_attributes {
    key_algorithm = "TDES_3KEY"
    key_class     = "SYMMETRIC_KEY"
    key_usage     = "TR31_P0_PIN_ENCRYPTION_KEY"
    key_modes_of_use {
      decrypt = true
      encrypt = true
      wrap    = true
      unwrap  = true
    }
  }
  tags = {
    Name  = %[1]q
    Other = "Value"
  }
}
`, rName)
}

func testAccKeyConfig_tags2(rName string) string {
	return fmt.Sprintf(`
resource "aws_paymentcryptography_key" "test" {
  exportable = true
  key_attributes {
    key_algorithm = "TDES_3KEY"
    key_class     = "SYMMETRIC_KEY"
    key_usage     = "TR31_P0_PIN_ENCRYPTION_KEY"
    key_modes_of_use {
      decrypt = true
      encrypt = true
      wrap    = true
      unwrap  = true
    }
  }
  tags = {
    Name2 = %[1]q
    Other = "Value2"
  }
}
`, rName)
}

func testAccKeyConfig_disable(rName string) string {
	return fmt.Sprintf(`
resource "aws_paymentcryptography_key" "test" {
  enabled    = false
  exportable = true
  key_attributes {
    key_algorithm = "TDES_3KEY"
    key_class     = "SYMMETRIC_KEY"
    key_usage     = "TR31_P0_PIN_ENCRYPTION_KEY"
    key_modes_of_use {
      decrypt = true
      encrypt = true
      wrap    = true
      unwrap  = true
    }
  }
  tags = {
    Name  = %[1]q
    Other = "Value"
  }
}
`, rName)
}

func testAccKeyConfig_enable(rName string) string {
	return fmt.Sprintf(`
resource "aws_paymentcryptography_key" "test" {
  enabled    = true
  exportable = true
  key_attributes {
    key_algorithm = "TDES_3KEY"
    key_class     = "SYMMETRIC_KEY"
    key_usage     = "TR31_P0_PIN_ENCRYPTION_KEY"
    key_modes_of_use {
      decrypt = true
      encrypt = true
      wrap    = true
      unwrap  = true
    }
  }
  tags = {
    Name  = %[1]q
    Other = "Value"
  }
}
`, rName)
}
