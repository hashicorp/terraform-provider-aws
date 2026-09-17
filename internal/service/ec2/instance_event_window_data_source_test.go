// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2InstanceEventWindowDataSource_id(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_instance_event_window.test"
	resourceName := "aws_ec2_instance_event_window.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceEventWindowDataSourceConfig_id(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, names.AttrID, regexache.MustCompile(`^iew-`)),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttr(dataSourceName, "time_ranges.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "time_ranges.0.start_week_day", "sunday"),
					resource.TestCheckResourceAttr(dataSourceName, "time_ranges.0.start_hour", "2"),
				),
			},
		},
	})
}

func TestAccEC2InstanceEventWindowDataSource_filter(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_instance_event_window.test"
	resourceName := "aws_ec2_instance_event_window.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceEventWindowDataSourceConfig_filter(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
				),
			},
		},
	})
}

func testAccInstanceEventWindowDataSourceConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_instance_event_window" "test" {
  name = %[1]q

  time_ranges {
    start_week_day = "sunday"
    start_hour     = 2
    end_week_day   = "sunday"
    end_hour       = 6
  }
}
`, rName)
}

func testAccInstanceEventWindowDataSourceConfig_id(rName string) string {
	return acctest.ConfigCompose(testAccInstanceEventWindowDataSourceConfig_base(rName), `
data "aws_ec2_instance_event_window" "test" {
  id = aws_ec2_instance_event_window.test.id
}
`)
}

func testAccInstanceEventWindowDataSourceConfig_filter(rName string) string {
	return acctest.ConfigCompose(testAccInstanceEventWindowDataSourceConfig_base(rName), `
data "aws_ec2_instance_event_window" "test" {
  filter {
    name   = "event-window-name"
    values = [aws_ec2_instance_event_window.test.name]
  }
}
`)
}
