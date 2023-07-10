// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecr_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfecr "github.com/hashicorp/terraform-provider-aws/internal/service/ecr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func TestAccECRRepository_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v ecr.Repository
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecr.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "ecr", fmt.Sprintf("repository/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					testAccCheckRepositoryRegistryID(resourceName),
					testAccCheckRepositoryRepositoryURL(resourceName, rName),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.encryption_type", ecr.EncryptionTypeAes256),
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
	var v ecr.Repository
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecr.EndpointsID),
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
	var v1, v2 ecr.Repository
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecr.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v1),
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
				Config: testAccRepositoryConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccRepositoryConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccECRRepository_immutability(t *testing.T) {
	ctx := acctest.Context(t)
	var v ecr.Repository
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecr.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_immutability(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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
	var v1, v2 ecr.Repository
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecr.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_imageScanningConfiguration(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "image_scanning_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "image_scanning_configuration.0.scan_on_push", "true"),
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
					resource.TestCheckResourceAttr(resourceName, "image_scanning_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "image_scanning_configuration.0.scan_on_push", "false"),
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
	var v1, v2 ecr.Repository
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository.test"
	kmsKeyDataSourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecr.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_encryptionKMSDefaultkey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.encryption_type", ecr.EncryptionTypeKms),
					// This will be the default ECR service KMS key. We don't currently have a way to look this up.
					acctest.MatchResourceAttrRegionalARN(resourceName, "encryption_configuration.0.kms_key", "kms", regexp.MustCompile(fmt.Sprintf("key/%s$", verify.UUIDRegexPattern))),
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
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.encryption_type", ecr.EncryptionTypeKms),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_configuration.0.kms_key", kmsKeyDataSourceName, "arn"),
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
	var v1, v2 ecr.Repository
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecr.EndpointsID),
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
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.encryption_type", ecr.EncryptionTypeAes256),
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
		conn := acctest.Provider.Meta().(*conns.AWSClient).ECRConn(ctx)

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

func testAccCheckRepositoryExists(ctx context.Context, n string, v *ecr.Repository) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ECR Repository ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ECRConn(ctx)

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

func testAccCheckRepositoryRecreated(i, j *ecr.Repository) resource.TestCheckFunc { // nosemgrep:ci.ecr-in-func-name
	return func(s *terraform.State) error {
		if aws.TimeValue(i.CreatedAt).Equal(aws.TimeValue(j.CreatedAt)) {
			return fmt.Errorf("ECR repository was not recreated")
		}

		return nil
	}
}

func testAccCheckRepositoryNotRecreated(i, j *ecr.Repository) resource.TestCheckFunc { // nosemgrep:ci.ecr-in-func-name
	return func(s *terraform.State) error {
		if !aws.TimeValue(i.CreatedAt).Equal(aws.TimeValue(j.CreatedAt)) {
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
