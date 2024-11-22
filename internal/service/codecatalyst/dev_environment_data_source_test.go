// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codecatalyst_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCodeCatalystDevEnvironmentDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_codecatalyst_dev_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CodeCatalyst)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeCatalyst),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDevEnvironmentDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "space_name", "tf-cc-aws-provider"),
					resource.TestCheckResourceAttr(dataSourceName, "project_name", "tf-cc"),
				),
			},
		},
	})
}

func testAccDevEnvironmentDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_codecatalyst_dev_environment" "test" {
  alias         = %[1]q
  space_name    = "tf-cc-aws-provider"
  project_name  = "tf-cc"
  instance_type = "dev.standard1.small"
  persistent_storage {
    size = 16
  }
  ides {
    name = "VSCode"
  }
}

data "aws_codecatalyst_dev_environment" "test" {
  space_name   = "tf-cc-aws-provider"
  project_name = "tf-cc"
  env_id       = aws_codecatalyst_dev_environment.test.id
}
`, rName)
}
