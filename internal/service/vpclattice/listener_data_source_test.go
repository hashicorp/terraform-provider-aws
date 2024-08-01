// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCLatticeListenerDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_vpclattice_listener.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccListenerDataSourceConfig_fixedResponseHTTP(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrProtocol, "HTTP"),
					resource.TestCheckResourceAttr(dataSourceName, "default_action.0.fixed_response.0.status_code", "404"),
					acctest.MatchResourceAttrRegionalARN(dataSourceName, names.AttrARN, "vpc-lattice", regexache.MustCompile(`service/svc-.*/listener/listener-.+`)),
				),
			},
		},
	})
}

func TestAccVPCLatticeListenerDataSource_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_vpclattice_listener.test_tags"
	tag_name := "tag0"
	tag_value := "value0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccListenerDataSourceConfig_one_tag(rName, tag_name, tag_value),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "tags.tag0", "value0"),
					acctest.MatchResourceAttrRegionalARN(dataSourceName, names.AttrARN, "vpc-lattice", regexache.MustCompile(`service/svc-.*/listener/listener-.+`)),
				),
			},
		},
	})
}

func TestAccVPCLatticeListenerDataSource_forwardMultiTargetGroupHTTP(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	targetGroupName1 := fmt.Sprintf("testtargetgroup-%s", sdkacctest.RandString(10))

	targetGroupResourceName := "aws_vpclattice_target_group.test"
	targetGroup1ResourceName := "aws_vpclattice_target_group.test1"
	dataSourceName := "data.aws_vpclattice_listener.test_multi_target"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccListenerDataSourceConfig_forwardMultiTargetGroupHTTP(rName, targetGroupName1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "default_action.0.forward.0.target_groups.0.target_group_identifier", targetGroupResourceName, names.AttrID),
					resource.TestCheckResourceAttr(dataSourceName, "default_action.0.forward.0.target_groups.0.weight", "80"),
					resource.TestCheckResourceAttrPair(dataSourceName, "default_action.0.forward.0.target_groups.1.target_group_identifier", targetGroup1ResourceName, names.AttrID),
					resource.TestCheckResourceAttr(dataSourceName, "default_action.0.forward.0.target_groups.1.weight", "20"),
					acctest.MatchResourceAttrRegionalARN(dataSourceName, names.AttrARN, "vpc-lattice", regexache.MustCompile(`service/svc-.*/listener/listener-.+`)),
				),
			},
		},
	})
}

func testAccListenerDataSourceConfig_one_tag(rName, tag_key, tag_value string) string {
	return acctest.ConfigCompose(testAccListenerDataSourceConfig_basic(rName), fmt.Sprintf(`
resource "aws_vpclattice_listener" "test_tags" {
  name               = %[1]q
  protocol           = "HTTP"
  service_identifier = aws_vpclattice_service.test.id

  default_action {
    forward {
      target_groups {
        target_group_identifier = aws_vpclattice_target_group.test.id
        weight                  = 100
      }
    }
  }

  tags = {
    %[2]q = %[3]q
  }
}

data "aws_vpclattice_listener" "test_tags" {
  service_identifier  = aws_vpclattice_service.test.id
  listener_identifier = aws_vpclattice_listener.test_tags.arn
}
`, rName, tag_key, tag_value))
}

func testAccListenerDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 0), fmt.Sprintf(`
resource "aws_vpclattice_service" "test" {
  name = %[1]q
}

resource "aws_vpclattice_target_group" "test" {
  name = %[1]q
  type = "INSTANCE"

  config {
    port           = 80
    protocol       = "HTTP"
    vpc_identifier = aws_vpc.test.id
  }
}
`, rName))
}

func testAccListenerDataSourceConfig_fixedResponseHTTP(rName string) string {
	return acctest.ConfigCompose(testAccListenerDataSourceConfig_basic(rName), fmt.Sprintf(`
resource "aws_vpclattice_listener" "test" {
  name               = %[1]q
  protocol           = "HTTP"
  service_identifier = aws_vpclattice_service.test.id
  default_action {
    fixed_response {
      status_code = 404
    }
  }
}

data "aws_vpclattice_listener" "test" {
  service_identifier  = aws_vpclattice_service.test.arn
  listener_identifier = aws_vpclattice_listener.test.arn
}
`, rName))
}

func testAccListenerDataSourceConfig_forwardMultiTargetGroupHTTP(rName string, targetGroupName1 string) string {
	return acctest.ConfigCompose(testAccListenerConfig_basic(rName), fmt.Sprintf(`
resource "aws_vpclattice_target_group" "test1" {
  name = %[2]q
  type = "INSTANCE"

  config {
    port           = 8080
    protocol       = "HTTP"
    vpc_identifier = aws_vpc.test.id
  }
}

resource "aws_vpclattice_listener" "test" {
  name               = %[1]q
  protocol           = "HTTP"
  service_identifier = aws_vpclattice_service.test.id
  default_action {
    forward {
      target_groups {
        target_group_identifier = aws_vpclattice_target_group.test.id
        weight                  = 80
      }
      target_groups {
        target_group_identifier = aws_vpclattice_target_group.test1.id
        weight                  = 20
      }
    }
  }
}

data "aws_vpclattice_listener" "test_multi_target" {
  service_identifier  = aws_vpclattice_service.test.id
  listener_identifier = aws_vpclattice_listener.test.arn
}
`, rName, targetGroupName1))
}
