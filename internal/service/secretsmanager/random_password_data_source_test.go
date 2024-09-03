// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secretsmanager_test

import (
	"fmt"
	"testing"
	"unicode"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSecretsManagerRandomPasswordDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_secretsmanager_random_password.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRandomPasswordDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccRandomPasswordDataSource(datasourceName, 40),
				),
			},
		},
	})
}

func TestAccSecretsManagerRandomPasswordDataSource_exclude(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_secretsmanager_random_password.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRandomPasswordDataSourceConfig_exclude(),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						dataSource, ok := s.RootModule().Resources[datasourceName]
						if !ok {
							return fmt.Errorf("root module has no resource called %s", datasourceName)
						}

						if len(dataSource.Primary.Attributes["random_password"]) != 40 {
							return fmt.Errorf(
								"len(%s) != %d",
								dataSource.Primary.Attributes["random_password"],
								40,
							)
						}

						for _, r := range dataSource.Primary.Attributes["random_password"] {
							if !(unicode.IsLower(r) && unicode.IsLetter(r)) {
								return fmt.Errorf("expected only lowercase letters")
							}
						}

						return nil
					},
				),
			},
		},
	})
}

func testAccRandomPasswordDataSource(datasourceName string, expectedLength int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		dataSource, ok := s.RootModule().Resources[datasourceName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", datasourceName)
		}

		if len(dataSource.Primary.Attributes["random_password"]) != expectedLength {
			return fmt.Errorf(
				"len(%s) != %d",
				dataSource.Primary.Attributes["random_password"],
				expectedLength,
			)
		}

		return nil
	}
}

func testAccRandomPasswordDataSourceConfig_basic() string {
	return `
data "aws_secretsmanager_random_password" "test" {
  password_length = 40
}
`
}

func testAccRandomPasswordDataSourceConfig_exclude() string {
	return `
data "aws_secretsmanager_random_password" "test" {
  password_length     = 40
  exclude_numbers     = true
  exclude_uppercase   = true
  exclude_punctuation = true
}
`
}
