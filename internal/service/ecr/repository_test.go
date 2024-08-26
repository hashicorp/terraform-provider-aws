// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecr_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfecr "github.com/hashicorp/terraform-provider-aws/internal/service/ecr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECRRepository_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.Repository
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "ecr", fmt.Sprintf("repository/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					testAccCheckRepositoryRegistryID(resourceName),
					testAccCheckRepositoryRepositoryURL(resourceName, rName),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.encryption_type", string(types.EncryptionTypeAes256)),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.kms_key", ""),
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

func TestAccECRRepository_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.Repository
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfecr.ResourceRepository(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccECRRepository_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 types.Repository
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v1),
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
				Config: testAccRepositoryConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccRepositoryConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccECRRepository_immutability(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.Repository
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_immutability(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "image_tag_mutability", "IMMUTABLE"),
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

func TestAccECRRepository_Image_scanning(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 types.Repository
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_imageScanningConfiguration(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "image_scanning_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "image_scanning_configuration.0.scan_on_push", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Test that the removal of the non-default image_scanning_configuration causes plan changes
				Config:             testAccRepositoryConfig_basic(rName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				// Test attribute update
				Config: testAccRepositoryConfig_imageScanningConfiguration(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "image_scanning_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "image_scanning_configuration.0.scan_on_push", acctest.CtFalse),
				),
			},
			{
				// Test that the removal of the default image_scanning_configuration doesn't cause any plan changes
				Config:             testAccRepositoryConfig_basic(rName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccECRRepository_Encryption_kms(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 types.Repository
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository.test"
	kmsKeyDataSourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_encryptionKMSDefaultkey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.encryption_type", string(types.EncryptionTypeKms)),
					// This will be the default ECR service KMS key. We don't currently have a way to look this up.
					acctest.MatchResourceAttrRegionalARN(resourceName, "encryption_configuration.0.kms_key", "kms", regexache.MustCompile(fmt.Sprintf("key/%s$", verify.UUIDRegexPattern))),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRepositoryConfig_encryptionKMSCustomkey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v2),
					testAccCheckRepositoryRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.encryption_type", string(types.EncryptionTypeKms)),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_configuration.0.kms_key", kmsKeyDataSourceName, names.AttrARN),
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

func TestAccECRRepository_Encryption_aes256(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 types.Repository
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				// Test that the addition of the default encryption_configuration doesn't recreation in the next step
				Config: testAccRepositoryConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v1),
				),
			},
			{
				Config: testAccRepositoryConfig_encryptionAES256(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v2),
					testAccCheckRepositoryNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.encryption_type", string(types.EncryptionTypeAes256)),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.kms_key", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Test that the removal of the default encryption_configuration doesn't cause any plan changes
				Config:   testAccRepositoryConfig_basic(rName),
				PlanOnly: true,
			},
		},
	})
}

func testAccCheckRepositoryDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ECRClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ecr_repository" {
				continue
			}

			_, err := tfecr.FindRepositoryByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ECR Repository %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckRepositoryExists(ctx context.Context, n string, v *types.Repository) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ECR Repository ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ECRClient(ctx)

		output, err := tfecr.FindRepositoryByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckRepositoryRegistryID(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attributeValue := acctest.AccountID()
		return resource.TestCheckResourceAttr(resourceName, "registry_id", attributeValue)(s)
	}
}

func testAccCheckRepositoryRepositoryURL(resourceName, repositoryName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attributeValue := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/%s", acctest.AccountID(), acctest.Region(), repositoryName)
		return resource.TestCheckResourceAttr(resourceName, "repository_url", attributeValue)(s)
	}
}

func testAccCheckRepositoryRecreated(i, j *types.Repository) resource.TestCheckFunc { // nosemgrep:ci.ecr-in-func-name
	return func(s *terraform.State) error {
		if aws.ToTime(i.CreatedAt).Equal(aws.ToTime(j.CreatedAt)) {
			return fmt.Errorf("ECR repository was not recreated")
		}

		return nil
	}
}

func testAccCheckRepositoryNotRecreated(i, j *types.Repository) resource.TestCheckFunc { // nosemgrep:ci.ecr-in-func-name
	return func(s *terraform.State) error {
		if !aws.ToTime(i.CreatedAt).Equal(aws.ToTime(j.CreatedAt)) {
			return fmt.Errorf("ECR repository was recreated")
		}

		return nil
	}
}

func testAccRepositoryConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name = %[1]q
}
`, rName)
}

func testAccRepositoryConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccRepositoryConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccRepositoryConfig_immutability(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name                 = %[1]q
  image_tag_mutability = "IMMUTABLE"
}
`, rName)
}

func testAccRepositoryConfig_imageScanningConfiguration(rName string, scanOnPush bool) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name = %[1]q

  image_scanning_configuration {
    scan_on_push = %[2]t
  }
}
`, rName, scanOnPush)
}

func testAccRepositoryConfig_encryptionKMSDefaultkey(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name = %[1]q

  encryption_configuration {
    encryption_type = "KMS"
  }
}
`, rName)
}

func testAccRepositoryConfig_encryptionKMSCustomkey(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {}

resource "aws_ecr_repository" "test" {
  name = %[1]q

  encryption_configuration {
    encryption_type = "KMS"
    kms_key         = aws_kms_key.test.arn
  }
}
`, rName)
}

func testAccRepositoryConfig_encryptionAES256(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name = %[1]q

  encryption_configuration {
    encryption_type = "AES256"
  }
}
`, rName)
}
