// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package secretsmanager_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	tfcversion "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsecretsmanager "github.com/hashicorp/terraform-provider-aws/internal/service/secretsmanager"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSecretsManagerSecretVersion_basicString(t *testing.T) {
	ctx := acctest.Context(t)
	var version secretsmanager.GetSecretValueOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret_version.test"
	secretResourceName := "aws_secretsmanager_secret.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretVersionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretVersionConfig_string(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(ctx, t, resourceName, &version),
					resource.TestCheckResourceAttr(resourceName, "secret_string", "test-string"),
					resource.TestCheckResourceAttrSet(resourceName, "version_id"),
					resource.TestCheckResourceAttr(resourceName, "version_stages.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "version_stages.*", "AWSCURRENT"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, secretResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"has_secret_string_wo"},
			},
		},
	})
}

func TestAccSecretsManagerSecretVersion_base64Binary(t *testing.T) {
	ctx := acctest.Context(t)
	var version secretsmanager.GetSecretValueOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret_version.test"
	secretResourceName := "aws_secretsmanager_secret.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretVersionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretVersionConfig_binary(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(ctx, t, resourceName, &version),
					resource.TestCheckResourceAttr(resourceName, "secret_binary", inttypes.Base64EncodeOnce([]byte("test-binary"))),
					resource.TestCheckResourceAttrSet(resourceName, "version_id"),
					resource.TestCheckResourceAttr(resourceName, "version_stages.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "version_stages.*", "AWSCURRENT"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, secretResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"has_secret_string_wo"},
			},
		},
	})
}

func TestAccSecretsManagerSecretVersion_versionStages(t *testing.T) {
	ctx := acctest.Context(t)
	var version secretsmanager.GetSecretValueOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret_version.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretVersionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretVersionConfig_stagesSingle(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(ctx, t, resourceName, &version),
					resource.TestCheckResourceAttr(resourceName, "secret_string", "test-string"),
					resource.TestCheckResourceAttr(resourceName, "version_stages.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "version_stages.*", "AWSCURRENT"),
					resource.TestCheckTypeSetElemAttr(resourceName, "version_stages.*", "one"),
				),
			},
			{
				Config: testAccSecretVersionConfig_stagesSingleUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(ctx, t, resourceName, &version),
					resource.TestCheckResourceAttr(resourceName, "secret_string", "test-string"),
					resource.TestCheckResourceAttr(resourceName, "version_stages.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "version_stages.*", "AWSCURRENT"),
					resource.TestCheckTypeSetElemAttr(resourceName, "version_stages.*", "two"),
				),
			},
			{
				Config: testAccSecretVersionConfig_stagesMultiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(ctx, t, resourceName, &version),
					resource.TestCheckResourceAttr(resourceName, "secret_string", "test-string"),
					resource.TestCheckResourceAttr(resourceName, "version_stages.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "version_stages.*", "AWSCURRENT"),
					resource.TestCheckTypeSetElemAttr(resourceName, "version_stages.*", "one"),
					resource.TestCheckTypeSetElemAttr(resourceName, "version_stages.*", "two"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"has_secret_string_wo"},
			},
		},
	})
}

func TestAccSecretsManagerSecretVersion_versionStagesExternalUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var version secretsmanager.GetSecretValueOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret_version.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretVersionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretVersionConfig_stagesSingle(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(ctx, t, resourceName, &version),
					resource.TestCheckResourceAttr(resourceName, "secret_string", "test-string"),
					resource.TestCheckResourceAttr(resourceName, "version_stages.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "version_stages.*", "AWSCURRENT"),
					resource.TestCheckTypeSetElemAttr(resourceName, "version_stages.*", "one"),
				),
			},
			{
				PreConfig: func() {
					conn := acctest.ProviderMeta(ctx, t).SecretsManagerClient(ctx)

					_, err := conn.PutSecretValue(ctx, &secretsmanager.PutSecretValueInput{
						SecretId:     version.ARN,
						SecretString: aws.String("external_update"),
					})

					if err != nil {
						t.Fatalf("externally updating Secrets Manager Secret Version: %s", err)
					}
				},
				Config: testAccSecretVersionConfig_stagesSingle(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(ctx, t, resourceName, &version),
					resource.TestCheckResourceAttr(resourceName, "secret_string", "test-string"),
					resource.TestCheckResourceAttr(resourceName, "version_stages.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "version_stages.*", "AWSCURRENT"),
					resource.TestCheckTypeSetElemAttr(resourceName, "version_stages.*", "one"),
				),
			},
		},
	})
}

func TestAccSecretsManagerSecretVersion_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var version secretsmanager.GetSecretValueOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret_version.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretVersionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretVersionConfig_string(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(ctx, t, resourceName, &version),
					acctest.CheckSDKResourceDisappears(ctx, t, tfsecretsmanager.ResourceSecretVersion(), resourceName),
				),
				// Because resource Delete leaves a secret version with a single stage ("AWSCURRENT"), the resource is still there.
				// ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSecretsManagerSecretVersion_Disappears_secret(t *testing.T) {
	ctx := acctest.Context(t)
	var version secretsmanager.GetSecretValueOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret_version.test"
	secretResourceName := "aws_secretsmanager_secret.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretVersionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretVersionConfig_string(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(ctx, t, resourceName, &version),
					acctest.CheckSDKResourceDisappears(ctx, t, tfsecretsmanager.ResourceSecret(), secretResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSecretsManagerSecretVersion_multipleVersions(t *testing.T) {
	ctx := acctest.Context(t)
	var version1, version2, version3 secretsmanager.GetSecretValueOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resource1Name := "aws_secretsmanager_secret_version.test1"
	resource2Name := "aws_secretsmanager_secret_version.test2"
	resource3Name := "aws_secretsmanager_secret_version.test3"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretVersionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretVersionConfig_multipleVersions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(ctx, t, resource1Name, &version1),
					resource.TestCheckResourceAttr(resource1Name, "version_stages.#", "1"),
					resource.TestCheckTypeSetElemAttr(resource1Name, "version_stages.*", "one"),
					testAccCheckSecretVersionExists(ctx, t, resource2Name, &version2),
					resource.TestCheckResourceAttr(resource2Name, "version_stages.#", "2"),
					resource.TestCheckTypeSetElemAttr(resource2Name, "version_stages.*", "two"),
					resource.TestCheckTypeSetElemAttr(resource2Name, "version_stages.*", "2"),
					testAccCheckSecretVersionExists(ctx, t, resource3Name, &version3),
					resource.TestCheckResourceAttr(resource3Name, "version_stages.#", "2"),
					resource.TestCheckTypeSetElemAttr(resource3Name, "version_stages.*", "three"),
					resource.TestCheckTypeSetElemAttr(resource3Name, "version_stages.*", "AWSCURRENT"),
				),
			},
		},
	})
}

func TestAccSecretsManagerSecretVersion_stringWriteOnly(t *testing.T) {
	ctx := acctest.Context(t)
	var version secretsmanager.GetSecretValueOutput
	var versionWriteOnly secretsmanager.GetSecretValueOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret_version.test"
	secretResourceName := "aws_secretsmanager_secret.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck: acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfcversion.Must(tfcversion.NewVersion("1.11.0"))),
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretVersionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretVersionConfig_stringWriteOnly(rName, "test-secret", 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(ctx, t, resourceName, &version),
					testAccCheckSecretVersionExistsWriteOnly(ctx, t, resourceName, &versionWriteOnly),
					testAccCheckSecretVersionWriteOnlyValueEqual(t, &version, "test-secret"),
					testAccCheckSecretVersionWriteOnlyValueEmpty(t, &versionWriteOnly),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, secretResourceName, names.AttrARN),
				),
			},
			{
				Config: testAccSecretVersionConfig_stringWriteOnly(rName, "test-secret2", 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(ctx, t, resourceName, &version),
					testAccCheckSecretVersionExistsWriteOnly(ctx, t, resourceName, &versionWriteOnly),
					testAccCheckSecretVersionWriteOnlyValueEqual(t, &version, "test-secret2"),
					testAccCheckSecretVersionWriteOnlyValueEmpty(t, &versionWriteOnly),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, secretResourceName, names.AttrARN),
				),
			},
		},
	})
}

func TestAccSecretsManagerSecretVersion_stringWriteOnlyLimitedPermissions(t *testing.T) {
	ctx := acctest.Context(t)
	var version secretsmanager.GetSecretValueOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret_version.test"
	secretResourceName := "aws_secretsmanager_secret.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			acctest.PreCheckAssumeRoleARN(t)
		},
		ErrorCheck: acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfcversion.Must(tfcversion.NewVersion("1.11.0"))),
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretVersionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretVersionConfig_stringWriteOnlyLimitedPermissions(rName, "test-secret", 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(ctx, t, resourceName, &version),
					testAccCheckSecretVersionWriteOnlyValueEqual(t, &version, "test-secret"),
					resource.TestCheckResourceAttr(resourceName, "has_secret_string_wo", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, secretResourceName, names.AttrARN),
				),
			},
			{
				Config: testAccSecretVersionConfig_string(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(ctx, t, resourceName, &version),
					resource.TestCheckResourceAttr(resourceName, "secret_string", "test-string"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, secretResourceName, names.AttrARN),
				),
			},
		},
	})
}

func TestAccSecretsManagerSecretVersion_stringWriteOnly_stages(t *testing.T) {
	ctx := acctest.Context(t)
	var version secretsmanager.GetSecretValueOutput
	var versionWriteOnly secretsmanager.GetSecretValueOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret_version.test"
	secretResourceName := "aws_secretsmanager_secret.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck: acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfcversion.Must(tfcversion.NewVersion("1.11.0"))),
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretVersionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretVersionConfig_stringWriteOnly_stagesSingle(rName, "test-secret", 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(ctx, t, resourceName, &version),
					testAccCheckSecretVersionExistsWriteOnly(ctx, t, resourceName, &versionWriteOnly),
					testAccCheckSecretVersionWriteOnlyValueEmpty(t, &versionWriteOnly),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, secretResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "version_stages.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "version_stages.*", "AWSCURRENT"),
					resource.TestCheckTypeSetElemAttr(resourceName, "version_stages.*", "one"),
					testAccCheckSecretVersionWriteOnlyStagesEqual(t, &versionWriteOnly, []string{"one", "AWSCURRENT"}),
				),
			},
			{
				Config: testAccSecretVersionConfig_stringWriteOnly_stagesSingleUpdated(rName, "test-secret", 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(ctx, t, resourceName, &version),
					testAccCheckSecretVersionExistsWriteOnly(ctx, t, resourceName, &versionWriteOnly),
					testAccCheckSecretVersionWriteOnlyValueEmpty(t, &versionWriteOnly),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, secretResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "version_stages.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "version_stages.*", "AWSCURRENT"),
					resource.TestCheckTypeSetElemAttr(resourceName, "version_stages.*", "two"),
					testAccCheckSecretVersionWriteOnlyStagesEqual(t, &versionWriteOnly, []string{"AWSCURRENT", "two"}),
				),
			},
			{
				Config: testAccSecretVersionConfig_stringWriteOnly_stagesMultiple(rName, "test-secret", 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(ctx, t, resourceName, &version),
					testAccCheckSecretVersionExistsWriteOnly(ctx, t, resourceName, &versionWriteOnly),
					testAccCheckSecretVersionWriteOnlyValueEmpty(t, &versionWriteOnly),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, secretResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "version_stages.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "version_stages.*", "AWSCURRENT"),
					resource.TestCheckTypeSetElemAttr(resourceName, "version_stages.*", "two"),
					resource.TestCheckTypeSetElemAttr(resourceName, "version_stages.*", "one"),
					testAccCheckSecretVersionWriteOnlyStagesEqual(t, &versionWriteOnly, []string{"one", "AWSCURRENT", "two"}),
				),
			},
		},
	})
}

func testAccCheckSecretVersionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SecretsManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_secretsmanager_secret_version" {
				continue
			}

			output, err := tfsecretsmanager.FindSecretVersionByTwoPartKey(ctx, conn, rs.Primary.Attributes["secret_id"], rs.Primary.Attributes["version_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			if len(output.VersionStages) == 0 || (len(output.VersionStages) == 1 && (output.VersionStages[0] == "AWSCURRENT" || output.VersionStages[0] == "AWSPREVIOUS")) {
				continue
			}

			return fmt.Errorf("Secrets Manager Secret Version %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckSecretVersionExists(ctx context.Context, t *testing.T, n string, v *secretsmanager.GetSecretValueOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SecretsManagerClient(ctx)

		output, err := tfsecretsmanager.FindSecretVersionByTwoPartKey(ctx, conn, rs.Primary.Attributes["secret_id"], rs.Primary.Attributes["version_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccSecretVersionImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s|%s", rs.Primary.Attributes["secret_id"], rs.Primary.Attributes["version_id"]), nil
	}
}

func testAccCheckSecretVersionWriteOnlyValueEqual(t *testing.T, param *secretsmanager.GetSecretValueOutput, writeOnlyValue string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(param.SecretString) != writeOnlyValue {
			t.Fatalf("Expected SecretsManger SecretString to be %v, but got %v", writeOnlyValue, aws.ToString(param.SecretString))
		}
		return nil
	}
}

func testAccCheckSecretVersionExistsWriteOnly(ctx context.Context, t *testing.T, n string, v *secretsmanager.GetSecretValueOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SecretsManagerClient(ctx)

		arn, versionEntry, err := tfsecretsmanager.FindSecretVersionEntryByTwoPartKey(ctx, conn, rs.Primary.Attributes["secret_id"], rs.Primary.Attributes["version_id"])

		if err != nil {
			return err
		}

		// Construct a GetSecretValueOutput-like structure from ListSecretVersionIds result
		result := &secretsmanager.GetSecretValueOutput{
			ARN:           arn,
			VersionId:     versionEntry.VersionId,
			VersionStages: versionEntry.VersionStages,
		}

		*v = *result

		return nil
	}
}

func testAccCheckSecretVersionWriteOnlyValueEmpty(t *testing.T, param *secretsmanager.GetSecretValueOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(param.SecretString) != "" {
			t.Fatalf("Expected SecretsManger SecretString to be an empty string, but got %v", aws.ToString(param.SecretString))
		}
		return nil
	}
}

func testAccCheckSecretVersionWriteOnlyStagesEqual(t *testing.T, param *secretsmanager.GetSecretValueOutput, stages []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !reflect.DeepEqual(param.VersionStages, stages) {
			t.Fatalf("Expected SecretsManger VersionStages to be %v, but got %v", stages, param.VersionStages)
		}
		return nil
	}
}

func testAccSecretVersionConfig_string(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = "test-string"
}
`, rName)
}

func testAccSecretVersionConfig_stringWriteOnly(rName, secret string, version int) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id                = aws_secretsmanager_secret.test.id
  secret_string_wo         = %[2]q
  secret_string_wo_version = %[3]d
}
`, rName, secret, version)
}

func testAccSecretVersionConfig_stringWriteOnlyLimitedPermissions(rName, secret string, version int) string {
	policy := `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "secretsmanager:DeleteSecret",
        "secretsmanager:DescribeSecret",
        "secretsmanager:GetResourcePolicy",
        "secretsmanager:ListSecretVersionIds",
        "secretsmanager:PutSecretValue"
      ],
      "Resource": "*"
    }
  ]
}`

	return acctest.ConfigCompose(
		acctest.ConfigAssumeRolePolicy(policy),
		fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id                = aws_secretsmanager_secret.test.id
  secret_string_wo         = %[2]q
  secret_string_wo_version = %[3]d
}
`, rName, secret, version),
	)
}

func testAccSecretVersionConfig_binary(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
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
  name = %[1]q
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
  name = %[1]q
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
  name = %[1]q
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = "test-string"

  version_stages = ["one", "two", "AWSCURRENT"]
}
`, rName)
}

func testAccSecretVersionConfig_multipleVersions(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_version" "test1" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = "test1"

  version_stages = ["one"]

  lifecycle {
    ignore_changes = [version_stages] # "AWSPREVIOUS"
  }
}

resource "aws_secretsmanager_secret_version" "test2" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = "test2"

  version_stages = ["two", "2"]

  depends_on = [aws_secretsmanager_secret_version.test1]

  lifecycle {
    ignore_changes = [version_stages] # "AWSPREVIOUS"
  }
}

resource "aws_secretsmanager_secret_version" "test3" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = "test3"

  version_stages = ["three", "AWSCURRENT"]

  depends_on = [aws_secretsmanager_secret_version.test2]
}
`, rName)
}

func testAccSecretVersionConfig_stringWriteOnly_stagesSingle(rName, secret string, version int) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id                = aws_secretsmanager_secret.test.id
  secret_string_wo         = %[2]q
  secret_string_wo_version = %[3]d

  version_stages = ["one", "AWSCURRENT"]
}
`, rName, secret, version)
}

func testAccSecretVersionConfig_stringWriteOnly_stagesSingleUpdated(rName, secret string, version int) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id                = aws_secretsmanager_secret.test.id
  secret_string_wo         = %[2]q
  secret_string_wo_version = %[3]d

  version_stages = ["two", "AWSCURRENT"]
}
`, rName, secret, version)
}

func testAccSecretVersionConfig_stringWriteOnly_stagesMultiple(rName, secret string, version int) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id                = aws_secretsmanager_secret.test.id
  secret_string_wo         = %[2]q
  secret_string_wo_version = %[3]d

  version_stages = ["one", "two", "AWSCURRENT"]
}
`, rName, secret, version)
}
