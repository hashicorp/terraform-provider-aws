package route53recoverycontrolconfig_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	r53rcc "github.com/aws/aws-sdk-go/service/route53recoverycontrolconfig"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfroute53recoverycontrolconfig "github.com/hashicorp/terraform-provider-aws/internal/service/route53recoverycontrolconfig"
)

func testAccCluster_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53recoverycontrolconfig_cluster.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(r53rcc.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, r53rcc.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "status", "DEPLOYED"),
					resource.TestCheckResourceAttr(resourceName, "cluster_endpoints.#", "5"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"cluster_endpoints"},
			},
		},
	})
}

func testAccCluster_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53recoverycontrolconfig_cluster.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(r53rcc.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, r53rcc.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfroute53recoverycontrolconfig.ResourceCluster(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckClusterDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).Route53RecoveryControlConfigConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53recoverycontrolconfig_cluster" {
			continue
		}

		input := &r53rcc.DescribeClusterInput{
			ClusterArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeCluster(input)

		if err == nil {
			return fmt.Errorf("Route53RecoveryControlConfig cluster (%s) not deleted", rs.Primary.ID)
		}
	}

	return nil
}

func testAccClusterConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53recoverycontrolconfig_cluster" "test" {
  name = %[1]q
}
`, rName)
}

func testAccCheckClusterExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53RecoveryControlConfigConn

		input := &r53rcc.DescribeClusterInput{
			ClusterArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeCluster(input)

		if err != nil {
			return err
		}
		return nil
	}
}
