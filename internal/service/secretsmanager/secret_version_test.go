// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secretsmanager_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsecretsmanager "github.com/hashicorp/terraform-provider-aws/internal/service/secretsmanager"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSecretsManagerSecretVersion_basicString(t *testing.T) {
	ctx := acctest.Context(t)
	var version secretsmanager.GetSecretValueOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret_version.test"
	secretResourceName := "aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretVersionConfig_string(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(ctx, resourceName, &version),
					resource.TestCheckResourceAttr(resourceName, "secret_string", "test-string"),
					resource.TestCheckResourceAttrSet(resourceName, "version_id"),
					resource.TestCheckResourceAttr(resourceName, "version_stages.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "version_stages.*", "AWSCURRENT"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, secretResourceName, names.AttrARN),
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
	ctx := acctest.Context(t)
	var version secretsmanager.GetSecretValueOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret_version.test"
	secretResourceName := "aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretVersionConfig_binary(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(ctx, resourceName, &version),
					resource.TestCheckResourceAttr(resourceName, "secret_binary", itypes.Base64EncodeOnce([]byte("test-binary"))),
					resource.TestCheckResourceAttrSet(resourceName, "version_id"),
					resource.TestCheckResourceAttr(resourceName, "version_stages.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "version_stages.*", "AWSCURRENT"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, secretResourceName, names.AttrARN),
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
	ctx := acctest.Context(t)
	var version secretsmanager.GetSecretValueOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretVersionConfig_stagesSingle(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(ctx, resourceName, &version),
					resource.TestCheckResourceAttr(resourceName, "secret_string", "test-string"),
					resource.TestCheckResourceAttr(resourceName, "version_stages.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "version_stages.*", "AWSCURRENT"),
					resource.TestCheckTypeSetElemAttr(resourceName, "version_stages.*", "one"),
				),
			},
			{
				Config: testAccSecretVersionConfig_stagesSingleUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(ctx, resourceName, &version),
					resource.TestCheckResourceAttr(resourceName, "secret_string", "test-string"),
					resource.TestCheckResourceAttr(resourceName, "version_stages.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "version_stages.*", "AWSCURRENT"),
					resource.TestCheckTypeSetElemAttr(resourceName, "version_stages.*", "two"),
				),
			},
			{
				Config: testAccSecretVersionConfig_stagesMultiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(ctx, resourceName, &version),
					resource.TestCheckResourceAttr(resourceName, "secret_string", "test-string"),
					resource.TestCheckResourceAttr(resourceName, "version_stages.#", acctest.Ct3),
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

func TestAccSecretsManagerSecretVersion_versionStagesExternalUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var version secretsmanager.GetSecretValueOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretVersionConfig_stagesSingle(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(ctx, resourceName, &version),
					resource.TestCheckResourceAttr(resourceName, "secret_string", "test-string"),
					resource.TestCheckResourceAttr(resourceName, "version_stages.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "version_stages.*", "AWSCURRENT"),
					resource.TestCheckTypeSetElemAttr(resourceName, "version_stages.*", "one"),
				),
			},
			{
				PreConfig: func() {
					conn := acctest.Provider.Meta().(*conns.AWSClient).SecretsManagerClient(ctx)

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
					testAccCheckSecretVersionExists(ctx, resourceName, &version),
					resource.TestCheckResourceAttr(resourceName, "secret_string", "test-string"),
					resource.TestCheckResourceAttr(resourceName, "version_stages.#", acctest.Ct2),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretVersionConfig_string(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(ctx, resourceName, &version),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsecretsmanager.ResourceSecretVersion(), resourceName),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret_version.test"
	secretResourceName := "aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretVersionConfig_string(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(ctx, resourceName, &version),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsecretsmanager.ResourceSecret(), secretResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSecretsManagerSecretVersion_multipleVersions(t *testing.T) {
	ctx := acctest.Context(t)
	var version1, version2, version3 secretsmanager.GetSecretValueOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource1Name := "aws_secretsmanager_secret_version.test1"
	resource2Name := "aws_secretsmanager_secret_version.test2"
	resource3Name := "aws_secretsmanager_secret_version.test3"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretVersionConfig_multipleVersions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(ctx, resource1Name, &version1),
					resource.TestCheckResourceAttr(resource1Name, "version_stages.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resource1Name, "version_stages.*", "one"),
					testAccCheckSecretVersionExists(ctx, resource2Name, &version2),
					resource.TestCheckResourceAttr(resource2Name, "version_stages.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resource2Name, "version_stages.*", "two"),
					resource.TestCheckTypeSetElemAttr(resource2Name, "version_stages.*", acctest.Ct2),
					testAccCheckSecretVersionExists(ctx, resource3Name, &version3),
					resource.TestCheckResourceAttr(resource3Name, "version_stages.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resource3Name, "version_stages.*", "three"),
					resource.TestCheckTypeSetElemAttr(resource3Name, "version_stages.*", "AWSCURRENT"),
				),
			},
		},
	})
}

func testAccCheckSecretVersionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SecretsManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_secretsmanager_secret_version" {
				continue
			}

			output, err := tfsecretsmanager.FindSecretVersionByTwoPartKey(ctx, conn, rs.Primary.Attributes["secret_id"], rs.Primary.Attributes["version_id"])

			if tfresource.NotFound(err) {
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

func testAccCheckSecretVersionExists(ctx context.Context, n string, v *secretsmanager.GetSecretValueOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecretsManagerClient(ctx)

		output, err := tfsecretsmanager.FindSecretVersionByTwoPartKey(ctx, conn, rs.Primary.Attributes["secret_id"], rs.Primary.Attributes["version_id"])

		if err != nil {
			return err
		}

		*v = *output

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
