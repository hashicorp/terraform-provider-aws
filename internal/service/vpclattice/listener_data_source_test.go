package vpclattice_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCLatticeListenerDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var listener vpclattice.GetListenerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_vpclattice_listener.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerDataSourceConfig_fixedResponseHTTP(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerExists(ctx, dataSourceName, &listener),
					resource.TestCheckResourceAttr(dataSourceName, "name", rName),
					resource.TestCheckResourceAttr(dataSourceName, "protocol", "HTTP"),
					resource.TestCheckResourceAttr(dataSourceName, "default_action.0.fixed_response.0.status_code", "404"),
					acctest.MatchResourceAttrRegionalARN(dataSourceName, "arn", "vpc-lattice", regexp.MustCompile(`service/svc-.*/listener/listener-.+`)),
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

	var listener vpclattice.GetListenerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_vpclattice_listener.test_tags"
	tag_name := sdkacctest.RandomWithPrefix("tag_name")
	tag_value := sdkacctest.RandomWithPrefix("tag_value")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerDataSourceConfig_one_tag(rName, tag_name, tag_value),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerExists(ctx, dataSourceName, &listener),
					resource.TestCheckResourceAttr(dataSourceName, fmt.Sprintf("tags.%s", tag_name), tag_value),
					acctest.MatchResourceAttrRegionalARN(dataSourceName, "arn", "vpc-lattice", regexp.MustCompile(`service/svc-.*/listener/listener-.+`)),
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
	service_identifier = aws_vpclattice_service.test.id
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
	service_identifier = aws_vpclattice_service.test.arn
	listener_identifier = aws_vpclattice_listener.test.arn
}
`, rName))
}
