// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccGlueSecurityConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var securityConfiguration awstypes.SecurityConfiguration

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_security_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityConfigurationExists(ctx, resourceName, &securityConfiguration),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.cloudwatch_encryption.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.cloudwatch_encryption.0.cloudwatch_encryption_mode", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.cloudwatch_encryption.0.kms_key_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.job_bookmarks_encryption.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.job_bookmarks_encryption.0.job_bookmarks_encryption_mode", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.job_bookmarks_encryption.0.kms_key_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.s3_encryption.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.s3_encryption.0.kms_key_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.s3_encryption.0.s3_encryption_mode", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
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

func TestAccGlueSecurityConfiguration_CloudWatchEncryptionCloudWatchEncryptionMode_sseKMS(t *testing.T) {
	ctx := acctest.Context(t)
	var securityConfiguration awstypes.SecurityConfiguration

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_glue_security_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityConfigurationConfig_cloudWatchEncryptionModeSSEKMS(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityConfigurationExists(ctx, resourceName, &securityConfiguration),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.cloudwatch_encryption.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.cloudwatch_encryption.0.cloudwatch_encryption_mode", "SSE-KMS"),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_configuration.0.cloudwatch_encryption.0.kms_key_arn", kmsKeyResourceName, names.AttrARN),
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

func TestAccGlueSecurityConfiguration_JobBookmarksEncryptionJobBookmarksEncryptionMode_cseKMS(t *testing.T) {
	ctx := acctest.Context(t)
	var securityConfiguration awstypes.SecurityConfiguration

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_glue_security_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityConfigurationConfig_jobBookmarksEncryptionModeCSEKMS(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityConfigurationExists(ctx, resourceName, &securityConfiguration),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.job_bookmarks_encryption.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.job_bookmarks_encryption.0.job_bookmarks_encryption_mode", "CSE-KMS"),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_configuration.0.job_bookmarks_encryption.0.kms_key_arn", kmsKeyResourceName, names.AttrARN),
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

func TestAccGlueSecurityConfiguration_S3EncryptionS3EncryptionMode_sseKMS(t *testing.T) {
	ctx := acctest.Context(t)
	var securityConfiguration awstypes.SecurityConfiguration

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_glue_security_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityConfigurationConfig_s3EncryptionModeSSEKMS(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityConfigurationExists(ctx, resourceName, &securityConfiguration),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.s3_encryption.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.s3_encryption.0.s3_encryption_mode", "SSE-KMS"),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_configuration.0.s3_encryption.0.kms_key_arn", kmsKeyResourceName, names.AttrARN),
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

func TestAccGlueSecurityConfiguration_S3EncryptionS3EncryptionMode_sseS3(t *testing.T) {
	ctx := acctest.Context(t)
	var securityConfiguration awstypes.SecurityConfiguration

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_security_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityConfigurationConfig_s3EncryptionModeSSES3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityConfigurationExists(ctx, resourceName, &securityConfiguration),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.s3_encryption.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.s3_encryption.0.s3_encryption_mode", "SSE-S3"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.s3_encryption.0.kms_key_arn", ""),
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

func testAccCheckSecurityConfigurationExists(ctx context.Context, resourceName string, securityConfiguration *awstypes.SecurityConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Glue Security Configuration ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueClient(ctx)

		output, err := conn.GetSecurityConfiguration(ctx, &glue.GetSecurityConfigurationInput{
			Name: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		if output.SecurityConfiguration == nil {
			return fmt.Errorf("Glue Security Configuration (%s) not found", rs.Primary.ID)
		}

		if aws.ToString(output.SecurityConfiguration.Name) == rs.Primary.ID {
			*securityConfiguration = *output.SecurityConfiguration
			return nil
		}

		return fmt.Errorf("Glue Security Configuration (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckSecurityConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_glue_security_configuration" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).GlueClient(ctx)

			output, err := conn.GetSecurityConfiguration(ctx, &glue.GetSecurityConfigurationInput{
				Name: aws.String(rs.Primary.ID),
			})

			if errs.IsA[*awstypes.EntityNotFoundException](err) {
				continue
			}

			if err != nil {
				return err
			}

			securityConfiguration := output.SecurityConfiguration
			if securityConfiguration != nil && aws.ToString(securityConfiguration.Name) == rs.Primary.ID {
				return fmt.Errorf("Glue Security Configuration %s still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccSecurityConfigurationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_security_configuration" "test" {
  name = %q

  encryption_configuration {
    cloudwatch_encryption {
      cloudwatch_encryption_mode = "DISABLED"
    }

    job_bookmarks_encryption {
      job_bookmarks_encryption_mode = "DISABLED"
    }

    s3_encryption {
      s3_encryption_mode = "DISABLED"
    }
  }
}
`, rName)
}

func testAccSecurityConfigurationConfig_cloudWatchEncryptionModeSSEKMS(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_glue_security_configuration" "test" {
  name = %q

  encryption_configuration {
    cloudwatch_encryption {
      cloudwatch_encryption_mode = "SSE-KMS"
      kms_key_arn                = aws_kms_key.test.arn
    }

    job_bookmarks_encryption {
      job_bookmarks_encryption_mode = "DISABLED"
    }

    s3_encryption {
      s3_encryption_mode = "DISABLED"
    }
  }
}
`, rName)
}

func testAccSecurityConfigurationConfig_jobBookmarksEncryptionModeCSEKMS(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_glue_security_configuration" "test" {
  name = %q

  encryption_configuration {
    cloudwatch_encryption {
      cloudwatch_encryption_mode = "DISABLED"
    }

    job_bookmarks_encryption {
      job_bookmarks_encryption_mode = "CSE-KMS"
      kms_key_arn                   = aws_kms_key.test.arn
    }

    s3_encryption {
      s3_encryption_mode = "DISABLED"
    }
  }
}
`, rName)
}

func testAccSecurityConfigurationConfig_s3EncryptionModeSSEKMS(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_glue_security_configuration" "test" {
  name = %q

  encryption_configuration {
    cloudwatch_encryption {
      cloudwatch_encryption_mode = "DISABLED"
    }

    job_bookmarks_encryption {
      job_bookmarks_encryption_mode = "DISABLED"
    }

    s3_encryption {
      kms_key_arn        = aws_kms_key.test.arn
      s3_encryption_mode = "SSE-KMS"
    }
  }
}
`, rName)
}

func testAccSecurityConfigurationConfig_s3EncryptionModeSSES3(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_security_configuration" "test" {
  name = %q

  encryption_configuration {
    cloudwatch_encryption {
      cloudwatch_encryption_mode = "DISABLED"
    }

    job_bookmarks_encryption {
      job_bookmarks_encryption_mode = "DISABLED"
    }

    s3_encryption {
      s3_encryption_mode = "SSE-S3"
    }
  }
}
`, rName)
}
