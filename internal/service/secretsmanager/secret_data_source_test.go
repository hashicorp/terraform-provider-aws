package secretsmanager_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/secretsmanager"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccSecretsManagerSecretDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, secretsmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSecretDataSourceConfig_MissingRequired,
				ExpectError: regexp.MustCompile(`must specify either arn or name`),
			},
			{
				Config:      testAccSecretDataSourceConfig_MultipleSpecified,
				ExpectError: regexp.MustCompile(`specify only arn or name`),
			},
			{
				Config:      testAccSecretDataSourceConfig_NonExistent,
				ExpectError: regexp.MustCompile(`not found`),
			},
		},
	})
}

func TestAccSecretsManagerSecretDataSource_arn(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret.test"
	datasourceName := "data.aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, secretsmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretDataSourceConfig_ARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccSecretCheckDataSource(datasourceName, resourceName),
				),
			},
		},
	})
}

func TestAccSecretsManagerSecretDataSource_name(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret.test"
	datasourceName := "data.aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, secretsmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretDataSourceConfig_Name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccSecretCheckDataSource(datasourceName, resourceName),
				),
			},
		},
	})
}

func TestAccSecretsManagerSecretDataSource_policy(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret.test"
	datasourceName := "data.aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, secretsmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretDataSourceConfig_Policy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccSecretCheckDataSource(datasourceName, resourceName),
				),
			},
		},
	})
}

func testAccSecretCheckDataSource(datasourceName, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resource, ok := s.RootModule().Resources[datasourceName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", datasourceName)
		}

		dataSource, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", resourceName)
		}

		attrNames := []string{
			"arn",
			"description",
			"kms_key_id",
			"name",
			"policy",
			"rotation_enabled",
			"rotation_lambda_arn",
			"rotation_rules.#",
			"tags.#",
		}

		for _, attrName := range attrNames {
			if resource.Primary.Attributes[attrName] != dataSource.Primary.Attributes[attrName] {
				return fmt.Errorf(
					"%s is %s; want %s",
					attrName,
					resource.Primary.Attributes[attrName],
					dataSource.Primary.Attributes[attrName],
				)
			}
		}

		return nil
	}
}

func testAccSecretDataSourceConfig_ARN(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "wrong" {
  name = "%[1]s-wrong"
}

resource "aws_secretsmanager_secret" "test" {
  name = "%[1]s"
}

data "aws_secretsmanager_secret" "test" {
  arn = aws_secretsmanager_secret.test.arn
}
`, rName)
}

const testAccSecretDataSourceConfig_MissingRequired = `
data "aws_secretsmanager_secret" "test" {}
`

//lintignore:AWSAT003,AWSAT005
const testAccSecretDataSourceConfig_MultipleSpecified = `
data "aws_secretsmanager_secret" "test" {
  arn  = "arn:aws:secretsmanager:us-east-1:123456789012:secret:tf-acc-test-does-not-exist"
  name = "tf-acc-test-does-not-exist"
}
`

func testAccSecretDataSourceConfig_Name(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "wrong" {
  name = "%[1]s-wrong"
}

resource "aws_secretsmanager_secret" "test" {
  name = "%[1]s"
}

data "aws_secretsmanager_secret" "test" {
  name = aws_secretsmanager_secret.test.name
}
`, rName)
}

func testAccSecretDataSourceConfig_Policy(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = "%[1]s"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "EnableAllPermissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "secretsmanager:GetSecretValue",
      "Resource": "*"
    }
  ]
}
POLICY
}

data "aws_secretsmanager_secret" "test" {
  name = aws_secretsmanager_secret.test.name
}
`, rName)
}

const testAccSecretDataSourceConfig_NonExistent = `
data "aws_secretsmanager_secret" "test" {
  name = "tf-acc-test-does-not-exist"
}
`
