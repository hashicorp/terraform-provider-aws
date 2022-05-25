package secretsmanager_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsecretsmanager "github.com/hashicorp/terraform-provider-aws/internal/service/secretsmanager"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func TestAccSecretsManagerSecretVersion_basicString(t *testing.T) {
	var version secretsmanager.GetSecretValueOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret_version.test"
	secretResourceName := "aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, secretsmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecretVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretVersionConfig_string(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(resourceName, &version),
					resource.TestCheckResourceAttr(resourceName, "secret_string", "test-string"),
					resource.TestCheckResourceAttrSet(resourceName, "version_id"),
					resource.TestCheckResourceAttr(resourceName, "version_stages.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "version_stages.*", "AWSCURRENT"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", secretResourceName, "arn"),
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

func TestAccSecretsManagerSecretVersion_base64Binary(t *testing.T) {
	var version secretsmanager.GetSecretValueOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret_version.test"
	secretResourceName := "aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, secretsmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecretVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretVersionConfig_binary(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(resourceName, &version),
					resource.TestCheckResourceAttr(resourceName, "secret_binary", verify.Base64Encode([]byte("test-binary"))),
					resource.TestCheckResourceAttrSet(resourceName, "version_id"),
					resource.TestCheckResourceAttr(resourceName, "version_stages.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "version_stages.*", "AWSCURRENT"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", secretResourceName, "arn"),
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

func TestAccSecretsManagerSecretVersion_versionStages(t *testing.T) {
	var version secretsmanager.GetSecretValueOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, secretsmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecretVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretVersionConfig_stagesSingle(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(resourceName, &version),
					resource.TestCheckResourceAttr(resourceName, "secret_string", "test-string"),
					resource.TestCheckResourceAttr(resourceName, "version_stages.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "version_stages.*", "AWSCURRENT"),
					resource.TestCheckTypeSetElemAttr(resourceName, "version_stages.*", "one"),
				),
			},
			{
				Config: testAccSecretVersionConfig_stagesSingleUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(resourceName, &version),
					resource.TestCheckResourceAttr(resourceName, "secret_string", "test-string"),
					resource.TestCheckResourceAttr(resourceName, "version_stages.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "version_stages.*", "AWSCURRENT"),
					resource.TestCheckTypeSetElemAttr(resourceName, "version_stages.*", "two"),
				),
			},
			{
				Config: testAccSecretVersionConfig_stagesMultiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(resourceName, &version),
					resource.TestCheckResourceAttr(resourceName, "secret_string", "test-string"),
					resource.TestCheckResourceAttr(resourceName, "version_stages.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "version_stages.*", "AWSCURRENT"),
					resource.TestCheckTypeSetElemAttr(resourceName, "version_stages.*", "one"),
					resource.TestCheckTypeSetElemAttr(resourceName, "version_stages.*", "two"),
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

func testAccCheckSecretVersionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SecretsManagerConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_secretsmanager_secret_version" {
			continue
		}

		secretID, versionID, err := tfsecretsmanager.DecodeSecretVersionID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &secretsmanager.GetSecretValueInput{
			SecretId:  aws.String(secretID),
			VersionId: aws.String(versionID),
		}

		output, err := conn.GetSecretValue(input)

		if err != nil {
			if tfawserr.ErrCodeEquals(err, secretsmanager.ErrCodeResourceNotFoundException) {
				return nil
			}
			if tfawserr.ErrMessageContains(err, secretsmanager.ErrCodeInvalidRequestException, "was deleted") || tfawserr.ErrMessageContains(err, secretsmanager.ErrCodeInvalidRequestException, "was marked for deletion") {
				return nil
			}
			return err
		}

		if output == nil {
			return nil
		}

		if len(output.VersionStages) == 0 {
			return nil
		}

		if len(output.VersionStages) == 1 && aws.StringValue(output.VersionStages[0]) == "AWSCURRENT" {
			return nil
		}

		return fmt.Errorf("Secret Version %q still exists", rs.Primary.ID)
	}

	return nil

}

func testAccCheckSecretVersionExists(resourceName string, version *secretsmanager.GetSecretValueOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		secretID, versionID, err := tfsecretsmanager.DecodeSecretVersionID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecretsManagerConn

		input := &secretsmanager.GetSecretValueInput{
			SecretId:  aws.String(secretID),
			VersionId: aws.String(versionID),
		}

		output, err := conn.GetSecretValue(input)

		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("Secret Version %q does not exist", rs.Primary.ID)
		}

		*version = *output

		return nil
	}
}

func testAccSecretVersionConfig_string(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = "%s"
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = "test-string"
}
`, rName)
}

func testAccSecretVersionConfig_binary(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = "%s"
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_binary = base64encode("test-binary")
}
`, rName)
}

func testAccSecretVersionConfig_stagesSingle(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = "%s"
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = "test-string"

  version_stages = ["one", "AWSCURRENT"]
}
`, rName)
}

func testAccSecretVersionConfig_stagesSingleUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = "%s"
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = "test-string"

  version_stages = ["two", "AWSCURRENT"]
}
`, rName)
}

func testAccSecretVersionConfig_stagesMultiple(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = "%s"
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = "test-string"

  version_stages = ["one", "two", "AWSCURRENT"]
}
`, rName)
}
