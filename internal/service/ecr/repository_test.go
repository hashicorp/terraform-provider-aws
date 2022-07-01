package ecr_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func TestAccECRRepository_basic(t *testing.T) {
	var v ecr.Repository
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(resourceName, &v),
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

func TestAccECRRepository_tags(t *testing.T) {
	var v1, v2 ecr.Repository
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Usage", "original"),
				),
			},
			{
				Config: testAccRepositoryConfig_tagsChanged(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Usage", "changed"),
				),
			},
		},
	})
}

func TestAccECRRepository_immutability(t *testing.T) {
	var v ecr.Repository
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_immutability(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(resourceName, &v),
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
	var v1, v2 ecr.Repository
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_imageScanningConfiguration(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(resourceName, &v1),
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
					testAccCheckRepositoryExists(resourceName, &v2),
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
	var v1, v2 ecr.Repository
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository.test"
	kmsKeyDataSourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_encryptionKMSDefaultkey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(resourceName, &v1),
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
					testAccCheckRepositoryExists(resourceName, &v2),
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
	var v1, v2 ecr.Repository
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				// Test that the addition of the default encryption_configuration doesn't recreation in the next step
				Config: testAccRepositoryConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(resourceName, &v1),
				),
			},
			{
				Config: testAccRepositoryConfig_encryptionAES256(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(resourceName, &v2),
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

func testAccCheckRepositoryDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ECRConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ecr_repository" {
			continue
		}

		input := ecr.DescribeRepositoriesInput{
			RepositoryNames: []*string{aws.String(rs.Primary.Attributes["name"])},
		}

		out, err := conn.DescribeRepositories(&input)

		if tfawserr.ErrCodeEquals(err, ecr.ErrCodeRepositoryNotFoundException) {
			return nil
		}

		if err != nil {
			return err
		}

		for _, repository := range out.Repositories {
			if aws.StringValue(repository.RepositoryName) == rs.Primary.Attributes["name"] {
				return fmt.Errorf("ECR repository still exists: %s", rs.Primary.Attributes["name"])
			}
		}
	}

	return nil
}

func testAccCheckRepositoryExists(name string, res *ecr.Repository) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ECR repository ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ECRConn

		output, err := conn.DescribeRepositories(&ecr.DescribeRepositoriesInput{
			RepositoryNames: aws.StringSlice([]string{rs.Primary.ID}),
		})
		if err != nil {
			return err
		}
		if len(output.Repositories) == 0 {
			return fmt.Errorf("ECR repository %s not found", rs.Primary.ID)
		}

		*res = *output.Repositories[0]

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

func testAccCheckRepositoryRecreated(i, j *ecr.Repository) resource.TestCheckFunc { // nosemgrep:ecr-in-func-name
	return func(s *terraform.State) error {
		if aws.TimeValue(i.CreatedAt).Equal(aws.TimeValue(j.CreatedAt)) {
			return fmt.Errorf("ECR repository was not recreated")
		}

		return nil
	}
}

func testAccCheckRepositoryNotRecreated(i, j *ecr.Repository) resource.TestCheckFunc { // nosemgrep:ecr-in-func-name
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
  name = %q
}
`, rName)
}

func testAccRepositoryConfig_tags(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name = %q

  tags = {
    Environment = "production"
    Usage       = "original"
  }
}
`, rName)
}

func testAccRepositoryConfig_tagsChanged(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name = %q

  tags = {
    Usage = "changed"
  }
}
`, rName)
}

func testAccRepositoryConfig_immutability(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name                 = %q
  image_tag_mutability = "IMMUTABLE"
}
`, rName)
}

func testAccRepositoryConfig_imageScanningConfiguration(rName string, scanOnPush bool) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name = %q

  image_scanning_configuration {
    scan_on_push = %t
  }
}
`, rName, scanOnPush)
}

func testAccRepositoryConfig_encryptionKMSDefaultkey(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name = %q

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
  name = %q

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
  name = %q

  encryption_configuration {
    encryption_type = "AES256"
  }
}
`, rName)
}
