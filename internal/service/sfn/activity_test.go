// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sfn_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	awstypes "github.com/aws/aws-sdk-go-v2/service/sfn/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsfn "github.com/hashicorp/terraform-provider-aws/internal/service/sfn"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSFNActivity_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sfn_activity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SFNServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckActivityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccActivityConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckActivityExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationDate),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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

func TestAccSFNActivity_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sfn_activity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SFNServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckActivityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccActivityConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckActivityExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsfn.ResourceActivity(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSFNActivity_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sfn_activity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SFNServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckActivityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccActivityConfig_basicTags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckActivityExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccActivityConfig_basicTags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckActivityExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccActivityConfig_basicTags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckActivityExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccSFNActivity_encryptionConfigurationCustomerManagedKMSKey(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sfn_activity.test"
	reusePeriodSeconds := 900
	kmsKeyResource := "aws_kms_key.kms_key_for_sfn"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SFNServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckActivityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccActivityConfig_encryptionConfigurationCustomerManagedKMSKey(rName, string(awstypes.EncryptionTypeCustomerManagedKmsKey), reusePeriodSeconds),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckActivityExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.type", string(awstypes.EncryptionTypeCustomerManagedKmsKey)),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.kms_data_key_reuse_period_seconds", strconv.Itoa(reusePeriodSeconds)),
					resource.TestCheckResourceAttrSet(resourceName, "encryption_configuration.0.kms_key_id"),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_configuration.0.kms_key_id", kmsKeyResource, names.AttrARN),
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

func TestAccSFNActivity_encryptionConfigurationServiceOwnedKey(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sfn_activity.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SFNServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckActivityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccActivityConfig_encryptionConfigurationServiceOwnedKey(rName, string(awstypes.EncryptionTypeAwsOwnedKey)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckActivityExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.type", string(awstypes.EncryptionTypeAwsOwnedKey)),
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

func testAccCheckActivityExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SFNClient(ctx)

		_, err := tfsfn.FindActivityByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckActivityDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SFNClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sfn_activity" {
				continue
			}

			// Retrying as Read after Delete is not always consistent.
			err := retry.RetryContext(ctx, 1*time.Minute, func() *retry.RetryError {
				_, err := tfsfn.FindActivityByARN(ctx, conn, rs.Primary.ID)

				if tfresource.NotFound(err) {
					return nil
				}

				if err != nil {
					return retry.NonRetryableError(err)
				}

				return retry.RetryableError(fmt.Errorf("Step Functions Activity still exists: %s", rs.Primary.ID))
			})

			return err
		}

		return nil
	}
}

func testAccActivityConfig_kmsBase() string {
	return `
resource "aws_kms_key" "kms_key_for_sfn" {}
`
}

func testAccActivityConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sfn_activity" "test" {
  name = %[1]q
}
`, rName)
}

func testAccActivityConfig_basicTags1(rName, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_sfn_activity" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tag1Key, tag1Value)
}

func testAccActivityConfig_basicTags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_sfn_activity" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value)
}

func testAccActivityConfig_encryptionConfigurationCustomerManagedKMSKey(rName string, rType string, reusePeriodSeconds int) string {
	return acctest.ConfigCompose(testAccActivityConfig_kmsBase(), fmt.Sprintf(`
resource "aws_sfn_activity" "test" {
  name = %[1]q
  encryption_configuration {
    kms_key_id                        = aws_kms_key.kms_key_for_sfn.arn
    type                              = %[2]q
    kms_data_key_reuse_period_seconds = %[3]d
  }
}
`, rName, rType, reusePeriodSeconds))
}

func testAccActivityConfig_encryptionConfigurationServiceOwnedKey(rName string, rType string) string {
	return acctest.ConfigCompose(testAccActivityConfig_kmsBase(), fmt.Sprintf(`
resource "aws_sfn_activity" "test" {
  name = %[1]q
  encryption_configuration {
    type = %[2]q
  }
}
`, rName, rType))
}
