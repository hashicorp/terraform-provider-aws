// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package drs_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/drs/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdrs "github.com/hashicorp/terraform-provider-aws/internal/service/drs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDRSReplicationConfigurationTemplate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	//if testing.Short() {
	//	t.Skip("skipping long-running test in short mode")
	//}

	resourceName := "aws_drs_replication_configuration_template.test"
	var rct awstypes.ReplicationConfigurationTemplate

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DRSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckReplicationConfigurationTemplateDestroy(ctx),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationConfigurationTemplateConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationConfigurationTemplateExists(ctx, resourceName, &rct),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "cache_usage_limits.#"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreateTime),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint.#"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngine),
					resource.TestCheckResourceAttrSet(resourceName, "full_engine_version"),
					resource.TestCheckResourceAttrSet(resourceName, "reader_endpoint.#"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, "subnet_ids.#"),
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

func testAccCheckReplicationConfigurationTemplateExists(ctx context.Context, n string, v *awstypes.ReplicationConfigurationTemplate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DRSClient(ctx)

		output, err := tfdrs.FindReplicationConfigurationTemplateByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckReplicationConfigurationTemplateDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DRSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_drs_replication_configuration_template" {
				continue
			}

			_, err := tfdrs.FindReplicationConfigurationTemplateByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("DRS Replication Configuration Template (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccReplicationConfigurationTemplateConfig_basic() string {
	return `
resource "aws_drs_replication_configuration_template" "test" {
  associate_default_security_group        = false
  bandwidth_throttling                    = 1
  create_public_ip                        = false
  data_plane_routing                      = "PRIVATE_IP"
  default_large_staging_disk_type         = "GP2"
  ebs_encryption                          = "NONE"
  use_dedicated_replication_server        = false
  replication_server_instance_type        = "t2.micro"
  replication_servers_security_groups_ids = []
  staging_area_subnet_id                  = ""

  pit_policy {
    interval           = 1
    retention_duration = 1
    units              = "DAY"
  }

  staging_area_tags = {
    Name = "staging-area"
  }
}
`
}
