package glue_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccGlueSecurityConfiguration_basic(t *testing.T) {
	var securityConfiguration glue.SecurityConfiguration

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_security_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityConfigurationConfig_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityConfigurationExists(resourceName, &securityConfiguration),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.cloudwatch_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.cloudwatch_encryption.0.cloudwatch_encryption_mode", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.cloudwatch_encryption.0.kms_key_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.job_bookmarks_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.job_bookmarks_encryption.0.job_bookmarks_encryption_mode", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.job_bookmarks_encryption.0.kms_key_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.s3_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.s3_encryption.0.kms_key_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.s3_encryption.0.s3_encryption_mode", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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
	var securityConfiguration glue.SecurityConfiguration

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_glue_security_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityConfigurationConfig_CloudWatchEncryption_CloudWatchEncryptionMode_SSEKMS(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityConfigurationExists(resourceName, &securityConfiguration),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.cloudwatch_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.cloudwatch_encryption.0.cloudwatch_encryption_mode", "SSE-KMS"),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_configuration.0.cloudwatch_encryption.0.kms_key_arn", kmsKeyResourceName, "arn"),
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
	var securityConfiguration glue.SecurityConfiguration

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_glue_security_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityConfigurationConfig_JobBookmarksEncryption_JobBookmarksEncryptionMode_CSEKMS(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityConfigurationExists(resourceName, &securityConfiguration),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.job_bookmarks_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.job_bookmarks_encryption.0.job_bookmarks_encryption_mode", "CSE-KMS"),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_configuration.0.job_bookmarks_encryption.0.kms_key_arn", kmsKeyResourceName, "arn"),
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
	var securityConfiguration glue.SecurityConfiguration

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_glue_security_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityConfigurationConfig_S3Encryption_S3EncryptionMode_SSEKMS(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityConfigurationExists(resourceName, &securityConfiguration),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.s3_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.s3_encryption.0.s3_encryption_mode", "SSE-KMS"),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_configuration.0.s3_encryption.0.kms_key_arn", kmsKeyResourceName, "arn"),
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
	var securityConfiguration glue.SecurityConfiguration

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_security_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityConfigurationConfig_S3Encryption_S3EncryptionMode_SSES3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityConfigurationExists(resourceName, &securityConfiguration),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.s3_encryption.#", "1"),
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

func testAccCheckSecurityConfigurationExists(resourceName string, securityConfiguration *glue.SecurityConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Glue Security Configuration ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueConn

		output, err := conn.GetSecurityConfiguration(&glue.GetSecurityConfigurationInput{
			Name: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		if output.SecurityConfiguration == nil {
			return fmt.Errorf("Glue Security Configuration (%s) not found", rs.Primary.ID)
		}

		if aws.StringValue(output.SecurityConfiguration.Name) == rs.Primary.ID {
			*securityConfiguration = *output.SecurityConfiguration
			return nil
		}

		return fmt.Errorf("Glue Security Configuration (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckSecurityConfigurationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_glue_security_configuration" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueConn

		output, err := conn.GetSecurityConfiguration(&glue.GetSecurityConfigurationInput{
			Name: aws.String(rs.Primary.ID),
		})

		if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		securityConfiguration := output.SecurityConfiguration
		if securityConfiguration != nil && aws.StringValue(securityConfiguration.Name) == rs.Primary.ID {
			return fmt.Errorf("Glue Security Configuration %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccSecurityConfigurationConfig_Basic(rName string) string {
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

func testAccSecurityConfigurationConfig_CloudWatchEncryption_CloudWatchEncryptionMode_SSEKMS(rName string) string {
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

func testAccSecurityConfigurationConfig_JobBookmarksEncryption_JobBookmarksEncryptionMode_CSEKMS(rName string) string {
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

func testAccSecurityConfigurationConfig_S3Encryption_S3EncryptionMode_SSEKMS(rName string) string {
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

func testAccSecurityConfigurationConfig_S3Encryption_S3EncryptionMode_SSES3(rName string) string {
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
