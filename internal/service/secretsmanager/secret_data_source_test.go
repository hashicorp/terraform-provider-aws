// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secretsmanager_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSecretsManagerSecretDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSecretDataSourceConfig_missingRequired,
				ExpectError: regexache.MustCompile("one of `arn,name` must be specified"),
			},
			{
				Config:      testAccSecretDataSourceConfig_multipleSpecified,
				ExpectError: regexache.MustCompile("only one of `arn,name` can be specified"),
			},
			{
				Config:      testAccSecretDataSourceConfig_nonExistent,
				ExpectError: regexache.MustCompile(`couldn't find resource`),
			},
		},
	})
}

func TestAccSecretsManagerSecretDataSource_arn(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret.test"
	datasourceName := "data.aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretDataSourceConfig_arn(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccSecretCheckDataSource(datasourceName, resourceName),
				),
			},
		},
	})
}

func TestAccSecretsManagerSecretDataSource_name(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret.test"
	datasourceName := "data.aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretDataSourceConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccSecretCheckDataSource(datasourceName, resourceName),
				),
			},
		},
	})
}

func TestAccSecretsManagerSecretDataSource_policy(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret.test"
	datasourceName := "data.aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretDataSourceConfig_policy(rName),
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
			names.AttrARN,
			names.AttrDescription,
			names.AttrKMSKeyID,
			names.AttrName,
			names.AttrPolicy,
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

func testAccSecretDataSourceConfig_arn(rName string) string {
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

const testAccSecretDataSourceConfig_missingRequired = `
data "aws_secretsmanager_secret" "test" {}
`

// lintignore:AWSAT003,AWSAT005
const testAccSecretDataSourceConfig_multipleSpecified = `
data "aws_secretsmanager_secret" "test" {
  arn  = "arn:aws:secretsmanager:us-east-1:123456789012:secret:tf-acc-test-does-not-exist"
  name = "tf-acc-test-does-not-exist"
}
`

func testAccSecretDataSourceConfig_name(rName string) string {
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

func testAccSecretDataSourceConfig_policy(rName string) string {
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

const testAccSecretDataSourceConfig_nonExistent = `
data "aws_secretsmanager_secret" "test" {
  name = "tf-acc-test-does-not-exist"
}
`
