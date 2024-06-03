// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secretsmanager_test

import (
	"fmt"
	"testing"
	"time"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSecretsManagerSecretsDataSource_filter(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret.test"
	dataSourceName := "data.aws_secretsmanager_secrets.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretsDataSourceConfig_base(rName),
				// Sleep to allow secrets become visible in the list.
				Check: acctest.CheckSleep(t, 30*time.Second),
			},
			{
				Config: testAccSecretsDataSourceConfig_filter(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "names.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(dataSourceName, "arns.0", resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "names.0", resourceName, names.AttrName),
				),
			},
		},
	})
}

func testAccSecretsDataSourceConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}
`, rName)
}

func testAccSecretsDataSourceConfig_filter(rName string) string {
	return acctest.ConfigCompose(testAccSecretsDataSourceConfig_base(rName), `
data "aws_secretsmanager_secrets" "test" {
  filter {
    name   = "name"
    values = [aws_secretsmanager_secret.test.name]
  }
}
`)
}
