// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSOptionGroupDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	desc := "terraform test"
	engineName := "sqlserver-ee"
	majorEngineVersion := "11.00"

	dataSourceName := "data.aws_db_option_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOptionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOptionGroupDataSourceConfig_basic(rName, desc, engineName, majorEngineVersion),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "option_group_name", rName),
					resource.TestCheckResourceAttr(dataSourceName, "option_group_description", desc),
					resource.TestCheckResourceAttr(dataSourceName, "engine_name", engineName),
					resource.TestCheckResourceAttr(dataSourceName, "major_engineVersion", majorEngineVersion),
					resource.TestCheckResourceAttr(dataSourceName, "options.#", "3"),
					resource.TestCheckResourceAttr(dataSourceName, "options.0.option_name", "Timezone"),
					resource.TestCheckResourceAttr(dataSourceName, "options.0.option_settings.0.name", "TIME_ZONE"),
					resource.TestCheckResourceAttr(dataSourceName, "options.0.option_settings.0.value", "UTC"),
				),
			},
		},
	})
}

func testAccOptionGroupDataSourceConfig_basic(rName, desc, engineName, majorEngineVersion string) string {
	return fmt.Sprintf(`
resource "aws_db_option_group" "test" {
  name = %[1]q
	option_group_description = %[2]q
  engine_name              = %[3]q
  major_engine_version     = %[4]q

  option {
    option_name = "Timezone"

    option_settings {
      name  = "TIME_ZONE"
      value = "UTC"
    }
  }

  option {
    option_name = "SQLSERVER_BACKUP_RESTORE"

    option_settings {
      name  = "IAM_ROLE_ARN"
      value = aws_iam_role.example.arn
    }
  }

  option {
    option_name = "TDE"
  }
}

data "aws_db_option_group" "test" {
  option_group_name             = aws_db_option_group.test.name
}
`, rName, desc, engineName, majorEngineVersion)
}
