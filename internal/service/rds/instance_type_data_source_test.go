// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSInstanceTypeDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_rds_instance_type.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceTypeDataSourceConfig_basic("db.t3.medium"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "instance_class", "db.t3.medium"),
					resource.TestCheckResourceAttr(dataSourceName, "ec2_instance_type", "t3.medium"),
					resource.TestCheckResourceAttr(dataSourceName, "memory_size", "4096"),
					resource.TestCheckResourceAttr(dataSourceName, "default_vcpus", "2"),
					resource.TestCheckResourceAttrSet(dataSourceName, "burstable_performance_supported"),
					resource.TestCheckResourceAttrSet(dataSourceName, "current_generation"),
				),
			},
		},
	})
}

func TestAccRDSInstanceTypeDataSource_invalidClass(t *testing.T) {
	ctx := acctest.Context(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccInstanceTypeDataSourceConfig_basic("db.serverless"),
				ExpectError: regexp.MustCompile(`No EC2 instance type found|corresponding EC2 instance type`),
			},
		},
	})
}

func testAccInstanceTypeDataSourceConfig_basic(instanceClass string) string {
	return fmt.Sprintf(`
data "aws_rds_instance_type" "test" {
  instance_class = %[1]q
}
`, instanceClass)
}
