package secretsmanager_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/secretsmanager"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfsecretsmanager "github.com/hashicorp/terraform-provider-aws/internal/service/secretsmanager"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccSecretsManagerSecret_basic(t *testing.T) {
	var secret secretsmanager.DescribeSecretOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, secretsmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecretDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(resourceName, &secret),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "secretsmanager", regexp.MustCompile(fmt.Sprintf("secret:%s-[[:alnum:]]+$", rName))),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "force_overwrite_replica_secret", "false"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "recovery_window_in_days", "30"),
					resource.TestCheckResourceAttr(resourceName, "rotation_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "rotation_lambda_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"recovery_window_in_days", "force_overwrite_replica_secret"},
			},
		},
	})
}

func TestAccSecretsManagerSecret_withNamePrefix(t *testing.T) {
	var secret secretsmanager.DescribeSecretOutput
	resourceName := "aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, secretsmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecretDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretConfig_namePrefix("tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(resourceName, &secret),
					create.TestCheckResourceAttrNameFromPrefix(resourceName, "name", "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "tf-acc-test-prefix-"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"recovery_window_in_days", "force_overwrite_replica_secret"},
			},
		},
	})
}

func TestAccSecretsManagerSecret_description(t *testing.T) {
	var secret secretsmanager.DescribeSecretOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, secretsmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecretDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				Config: testAccSecretConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"recovery_window_in_days", "force_overwrite_replica_secret"},
			},
		},
	})
}

func TestAccSecretsManagerSecret_basicReplica(t *testing.T) {
	var providers []*schema.Provider
	var secret secretsmanager.DescribeSecretOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t); acctest.PreCheckMultipleRegion(t, 2) },
		ErrorCheck:        acctest.ErrorCheck(t, secretsmanager.EndpointsID),
		ProviderFactories: acctest.FactoriesMultipleRegion(&providers, 2),
		CheckDestroy:      testAccCheckSecretDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretConfig_basicReplica(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "force_overwrite_replica_secret", "false"),
					resource.TestCheckResourceAttr(resourceName, "replica.#", "1"),
				),
			},
		},
	})
}

func TestAccSecretsManagerSecret_overwriteReplica(t *testing.T) {
	var providers []*schema.Provider
	var secret secretsmanager.DescribeSecretOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t); acctest.PreCheckMultipleRegion(t, 3) },
		ErrorCheck:        acctest.ErrorCheck(t, secretsmanager.EndpointsID),
		ProviderFactories: acctest.FactoriesMultipleRegion(&providers, 3),
		CheckDestroy:      testAccCheckSecretDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretConfig_overwriteReplica(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "force_overwrite_replica_secret", "true"),
				),
			},
			{
				Config: testAccSecretConfig_overwriteReplicaUpdate(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "force_overwrite_replica_secret", "true"),
				),
			},
			{
				Config: testAccSecretConfig_overwriteReplica(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "force_overwrite_replica_secret", "false"),
				),
			},
		},
	})
}

func TestAccSecretsManagerSecret_kmsKeyID(t *testing.T) {
	var secret secretsmanager.DescribeSecretOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, secretsmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecretDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretConfig_kmsKeyID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(resourceName, &secret),
					resource.TestCheckResourceAttrSet(resourceName, "kms_key_id"),
				),
			},
			{
				Config: testAccSecretConfig_kmsKeyIDUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(resourceName, &secret),
					resource.TestCheckResourceAttrSet(resourceName, "kms_key_id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"recovery_window_in_days", "force_overwrite_replica_secret"},
			},
		},
	})
}

func TestAccSecretsManagerSecret_RecoveryWindowInDays_recreate(t *testing.T) {
	var secret secretsmanager.DescribeSecretOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, secretsmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecretDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretConfig_recoveryWindowInDays(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "recovery_window_in_days", "0"),
				),
			},
			{
				Config: testAccSecretConfig_recoveryWindowInDays(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "recovery_window_in_days", "0"),
				),
				Taint: []string{resourceName},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"recovery_window_in_days", "force_overwrite_replica_secret"},
			},
		},
	})
}

func TestAccSecretsManagerSecret_rotationLambdaARN(t *testing.T) {
	var secret secretsmanager.DescribeSecretOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret.test"
	lambdaFunctionResourceName := "aws_lambda_function.test1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, secretsmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecretDestroy,
		Steps: []resource.TestStep{
			// Test enabling rotation on resource creation
			{
				Config: testAccSecretConfig_rotationLambdaARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "rotation_enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "rotation_lambda_arn", lambdaFunctionResourceName, "arn"),
				),
			},
			// Test updating rotation
			// We need a valid rotation function for this testing
			// InvalidRequestException: A previous rotation isn’t complete. That rotation will be reattempted.
			/*
				{
					Config: testAccSecretConfig_managerRotationLambdaARNUpdated(rName),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckSecretExists(resourceName, &secret),
						resource.TestCheckResourceAttr(resourceName, "rotation_enabled", "true"),
						resource.TestMatchResourceAttr(resourceName, "rotation_lambda_arn", regexp.MustCompile(fmt.Sprintf("^arn:[^:]+:lambda:[^:]+:[^:]+:function:%s-2$", rName))),
					),
				},
			*/
			// Test importing rotation
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"recovery_window_in_days", "force_overwrite_replica_secret"},
			},
			// Test removing rotation on resource update
			{
				Config: testAccSecretConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "rotation_enabled", "true"), // Must be removed with aws_secretsmanager_secret_rotation after version 2.67.0
				),
			},
		},
	})
}

func TestAccSecretsManagerSecret_rotationRules(t *testing.T) {
	var secret secretsmanager.DescribeSecretOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, secretsmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecretDestroy,
		Steps: []resource.TestStep{
			// Test creating rotation rules on resource creation
			{
				Config: testAccSecretConfig_rotationRules(rName, 7),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "rotation_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.0.automatically_after_days", "7"),
				),
			},
			// Test updating rotation rules
			// We need a valid rotation function for this testing
			// InvalidRequestException: A previous rotation isn’t complete. That rotation will be reattempted.
			/*
				{
					Config: testAccSecretConfig_rotationRules(rName, 1),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckSecretExists(resourceName, &secret),
						resource.TestCheckResourceAttr(resourceName, "rotation_enabled", "true"),
						resource.TestCheckResourceAttr(resourceName, "rotation_rules.#", "1"),
						resource.TestCheckResourceAttr(resourceName, "rotation_rules.0.automatically_after_days", "1"),
					),
				},
			*/
			// Test importing rotation rules
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"recovery_window_in_days", "force_overwrite_replica_secret"},
			},
			// Test removing rotation rules on resource update
			{
				Config: testAccSecretConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "rotation_enabled", "true"), // Must be removed with aws_secretsmanager_secret_rotation after version 2.67.0
				),
			},
		},
	})
}

func TestAccSecretsManagerSecret_tags(t *testing.T) {
	var secret secretsmanager.DescribeSecretOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, secretsmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecretDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretConfig_tagsSingle(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", "tag1value"),
				),
			},
			{
				Config: testAccSecretConfig_tagsSingleUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", "tag1value-updated"),
				),
			},
			{
				Config: testAccSecretConfig_tagsMultiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", "tag1value"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag2", "tag2value"),
				),
			},
			{
				Config: testAccSecretConfig_tagsSingle(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", "tag1value"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"recovery_window_in_days", "force_overwrite_replica_secret"},
			},
		},
	})
}

func TestAccSecretsManagerSecret_policy(t *testing.T) {
	var secret secretsmanager.DescribeSecretOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, secretsmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecretDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretConfig_policy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "description", "San Holo feat. Duskus"),
					resource.TestMatchResourceAttr(resourceName, "policy",
						regexp.MustCompile(`{"Action":"secretsmanager:GetSecretValue".+`)),
				),
			},
			{
				Config: testAccSecretConfig_policyEmpty(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "description", "Poliça"),
					resource.TestCheckResourceAttr(resourceName, "policy", ""),
				),
			},
			{
				Config: testAccSecretConfig_policy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(resourceName, &secret),
					resource.TestMatchResourceAttr(resourceName, "policy",
						regexp.MustCompile(`{"Action":"secretsmanager:GetSecretValue".+`)),
				),
			},
		},
	})
}

func testAccCheckSecretDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SecretsManagerConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_secretsmanager_secret" {
			continue
		}

		_, err := tfsecretsmanager.FindSecretByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Secrets Manager Secret %s still exists", rs.Primary.ID)
	}

	return nil

}

func testAccCheckSecretExists(n string, v *secretsmanager.DescribeSecretOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Secrets Manager Secret ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecretsManagerConn

		output, err := tfsecretsmanager.FindSecretByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SecretsManagerConn

	input := &secretsmanager.ListSecretsInput{}

	_, err := conn.ListSecrets(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccSecretConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  description = "%s"
  name        = "%s"
}
`, description, rName)
}

func testAccSecretConfig_basicReplica(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		fmt.Sprintf(`
data "aws_region" "alternate" {
  provider = awsalternate
}

resource "aws_secretsmanager_secret" "test" {
  name = %[1]q

  replica {
    region = data.aws_region.alternate.name
  }
}
`, rName))
}

func testAccSecretConfig_overwriteReplica(rName string, force_overwrite_replica_secret bool) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3),
		fmt.Sprintf(`
resource "aws_kms_key" "test" {
  provider                = awsalternate
  deletion_window_in_days = 7
}

resource "aws_kms_key" "test2" {
  provider                = awsthird
  deletion_window_in_days = 7
}

data "aws_region" "alternate" {
  provider = awsalternate
}

resource "aws_secretsmanager_secret" "test" {
  name                           = %[1]q
  force_overwrite_replica_secret = %[2]t

  replica {
    kms_key_id = aws_kms_key.test.key_id
    region     = data.aws_region.alternate.name
  }
}
`, rName, force_overwrite_replica_secret))
}

func testAccSecretConfig_overwriteReplicaUpdate(rName string, force_overwrite_replica_secret bool) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3),
		fmt.Sprintf(`
resource "aws_kms_key" "test" {
  provider                = awsalternate
  deletion_window_in_days = 7
}

resource "aws_kms_key" "test2" {
  provider                = awsthird
  deletion_window_in_days = 7
}

data "aws_region" "third" {
  provider = awsthird
}

resource "aws_secretsmanager_secret" "test" {
  name                           = %[1]q
  force_overwrite_replica_secret = %[2]t

  replica {
    kms_key_id = aws_kms_key.test2.key_id
    region     = data.aws_region.third.name
  }
}
`, rName, force_overwrite_replica_secret))
}

func testAccSecretConfig_name(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = "%s"
}
`, rName)
}

func testAccSecretConfig_namePrefix(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name_prefix = %[1]q
}
`, rName)
}

func testAccSecretConfig_kmsKeyID(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test1" {
  deletion_window_in_days = 7
}

resource "aws_kms_key" "test2" {
  deletion_window_in_days = 7
}

resource "aws_secretsmanager_secret" "test" {
  kms_key_id = aws_kms_key.test1.id
  name       = "%s"
}
`, rName)
}

func testAccSecretConfig_kmsKeyIDUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test1" {
  deletion_window_in_days = 7
}

resource "aws_kms_key" "test2" {
  deletion_window_in_days = 7
}

resource "aws_secretsmanager_secret" "test" {
  kms_key_id = aws_kms_key.test2.id
  name       = "%s"
}
`, rName)
}

func testAccSecretConfig_recoveryWindowInDays(rName string, recoveryWindowInDays int) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name                    = %q
  recovery_window_in_days = %d
}
`, rName, recoveryWindowInDays)
}

func testAccSecretConfig_rotationLambdaARN(rName string) string {
	return acctest.ConfigLambdaBase(rName, rName, rName) + fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name                = "%[1]s"
  rotation_lambda_arn = aws_lambda_function.test1.arn

  depends_on = [aws_lambda_permission.test1]
}

# Not a real rotation function
resource "aws_lambda_function" "test1" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s-1"
  handler       = "exports.example"
  role          = aws_iam_role.iam_for_lambda.arn
  runtime       = "nodejs12.x"
}

resource "aws_lambda_permission" "test1" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test1.function_name
  principal     = "secretsmanager.amazonaws.com"
  statement_id  = "AllowExecutionFromSecretsManager1"
}

# Not a real rotation function
resource "aws_lambda_function" "test2" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s-2"
  handler       = "exports.example"
  role          = aws_iam_role.iam_for_lambda.arn
  runtime       = "nodejs12.x"
}

resource "aws_lambda_permission" "test2" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test2.function_name
  principal     = "secretsmanager.amazonaws.com"
  statement_id  = "AllowExecutionFromSecretsManager2"
}
`, rName)
}

func testAccSecretConfig_rotationRules(rName string, automaticallyAfterDays int) string {
	return acctest.ConfigLambdaBase(rName, rName, rName) + fmt.Sprintf(`
# Not a real rotation function
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s"
  handler       = "exports.example"
  role          = aws_iam_role.iam_for_lambda.arn
  runtime       = "nodejs12.x"
}

resource "aws_lambda_permission" "test" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.function_name
  principal     = "secretsmanager.amazonaws.com"
  statement_id  = "AllowExecutionFromSecretsManager1"
}

resource "aws_secretsmanager_secret" "test" {
  name                = "%[1]s"
  rotation_lambda_arn = aws_lambda_function.test.arn

  rotation_rules {
    automatically_after_days = %[2]d
  }

  depends_on = [aws_lambda_permission.test]
}
`, rName, automaticallyAfterDays)
}

func testAccSecretConfig_tagsSingle(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = "%s"

  tags = {
    tag1 = "tag1value"
  }
}
`, rName)
}

func testAccSecretConfig_tagsSingleUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = "%s"

  tags = {
    tag1 = "tag1value-updated"
  }
}
`, rName)
}

func testAccSecretConfig_tagsMultiple(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = "%s"

  tags = {
    tag1 = "tag1value"
    tag2 = "tag2value"
  }
}
`, rName)
}

func testAccSecretConfig_policy(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Principal = {
        Service = "ec2.amazonaws.com"
      },
      Effect = "Allow"
      Sid    = ""
    }]
  })
}

resource "aws_secretsmanager_secret" "test" {
  name        = %[1]q
  description = "San Holo feat. Duskus"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid    = "EnableAllPermissions"
      Effect = "Allow"
      Principal = {
        AWS = aws_iam_role.test.arn
      }
      Action   = "secretsmanager:GetSecretValue"
      Resource = "*"
    }]
  })
}
`, rName)
}

func testAccSecretConfig_policyEmpty(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Principal = {
        Service = "ec2.amazonaws.com"
      },
      Effect = "Allow"
      Sid    = ""
    }]
  })
}

resource "aws_secretsmanager_secret" "test" {
  name        = %[1]q
  description = "Poliça"

  policy = "{}"
}
`, rName)
}
