// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package drs_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/drs/types"
	"github.com/aws/aws-sdk-go/aws"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
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
				Config: testAccReplicationConfigurationTemplateConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationConfigurationTemplateExists(ctx, resourceName, &rct),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "associate_default_security_group", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "bandwidth_throttling", "12"),
					resource.TestCheckResourceAttr(resourceName, "create_public_ip", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "data_plane_routing", "PRIVATE_IP"),
					resource.TestCheckResourceAttr(resourceName, "default_large_staging_disk_type", "GP2"),
					resource.TestCheckResourceAttr(resourceName, "ebs_encryption", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "use_dedicated_replication_server", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "replication_server_instance_type", "t3.small"),
					resource.TestCheckResourceAttr(resourceName, "replication_servers_security_groups_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "staging_area_subnet_id", aws.StringValue(rct.StagingAreaSubnetId)),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "pit_policy", map[string]string{
						names.AttrEnabled:    acctest.CtFalse,
						names.AttrInterval:   "14",
						"retention_duration": "21",
						"units":              "DAY",
						"rule_id":            acctest.Ct1,
					}),
					resource.TestCheckResourceAttr(resourceName, "staging_area_tags.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "staging_area_tags.Name", rName),
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

func testAccReplicationConfigurationTemplateConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 1),
		fmt.Sprintf(`
resource "aws_security_group" "test" {
  name        = %[1]q
  description = %[1]q
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_drs_replication_configuration_template" "test" {
  associate_default_security_group        = false
  bandwidth_throttling                    = 12
  create_public_ip                        = false
  data_plane_routing                      = "PRIVATE_IP"
  default_large_staging_disk_type         = "GP2"
  ebs_encryption                          = "NONE"
  use_dedicated_replication_server        = false
  replication_server_instance_type        = "t3.small"
  replication_servers_security_groups_ids = [aws_security_group.test.id]
  staging_area_subnet_id                  = aws_subnet.test[0].id

  pit_policy {
    enabled            = false
    interval           = 14
    retention_duration = 21
    units              = "DAY"
    rule_id            = 1
  }

  staging_area_tags = {
    Name = %[1]q
  }
}
`, rName))
}
