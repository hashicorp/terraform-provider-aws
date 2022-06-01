package cloudhsmv2_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudhsmv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudhsmv2 "github.com/hashicorp/terraform-provider-aws/internal/service/cloudhsmv2"
)

func testAccHSM_basic(t *testing.T) {
	resourceName := "aws_cloudhsm_v2_hsm.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudhsmv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckHSMDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHSMSubnetIDConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHSMExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "availability_zone", "aws_subnet.test.0", "availability_zone"),
					resource.TestCheckResourceAttrPair(resourceName, "cluster_id", "aws_cloudhsm_v2_cluster.test", "id"),
					resource.TestMatchResourceAttr(resourceName, "hsm_eni_id", regexp.MustCompile(`^eni-.+`)),
					resource.TestMatchResourceAttr(resourceName, "hsm_id", regexp.MustCompile(`^hsm-.+`)),
					resource.TestCheckResourceAttr(resourceName, "hsm_state", cloudhsmv2.HsmStateActive),
					resource.TestCheckResourceAttrSet(resourceName, "ip_address"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_id", "aws_subnet.test.0", "id"),
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

func testAccHSM_disappears(t *testing.T) {
	resourceName := "aws_cloudhsm_v2_hsm.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudhsmv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHSMSubnetIDConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloudhsmv2.ResourceHSM(), resourceName),
					// Verify Delete error handling
					acctest.CheckResourceDisappears(acctest.Provider, tfcloudhsmv2.ResourceHSM(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccHSM_disappears_Cluster(t *testing.T) {
	clusterResourceName := "aws_cloudhsm_v2_cluster.test"
	resourceName := "aws_cloudhsm_v2_hsm.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudhsmv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHSMSubnetIDConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloudhsmv2.ResourceHSM(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloudhsmv2.ResourceCluster(), clusterResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccHSM_AvailabilityZone(t *testing.T) {
	resourceName := "aws_cloudhsm_v2_hsm.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudhsmv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckHSMDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHSMAvailabilityZoneConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHSMExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "availability_zone", "aws_subnet.test.0", "availability_zone"),
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

func testAccHSM_IPAddress(t *testing.T) {
	resourceName := "aws_cloudhsm_v2_hsm.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudhsmv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckHSMDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHSMIPAddressConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHSMExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "ip_address", "10.0.0.5"),
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

func testAccHSMBaseConfig() string {
	return `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = element(data.aws_availability_zones.available.names, count.index)
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  vpc_id            = aws_vpc.test.id
}

resource "aws_cloudhsm_v2_cluster" "test" {
  hsm_type   = "hsm1.medium"
  subnet_ids = aws_subnet.test[*].id
}
`
}

func testAccHSMAvailabilityZoneConfig() string {
	return acctest.ConfigCompose(
		testAccHSMBaseConfig(),
		`
resource "aws_cloudhsm_v2_hsm" "test" {
  availability_zone = aws_subnet.test[0].availability_zone
  cluster_id        = aws_cloudhsm_v2_cluster.test.cluster_id
}
`)
}

func testAccHSMIPAddressConfig() string {
	return acctest.ConfigCompose(
		testAccHSMBaseConfig(),
		`
resource "aws_cloudhsm_v2_hsm" "test" {
  cluster_id = aws_cloudhsm_v2_cluster.test.cluster_id
  ip_address = cidrhost(aws_subnet.test[0].cidr_block, 5)
  subnet_id  = aws_subnet.test[0].id
}
`)
}

func testAccHSMSubnetIDConfig() string {
	return acctest.ConfigCompose(
		testAccHSMBaseConfig(),
		`
resource "aws_cloudhsm_v2_hsm" "test" {
  cluster_id = aws_cloudhsm_v2_cluster.test.cluster_id
  subnet_id  = aws_subnet.test[0].id
}
`)
}

func testAccCheckHSMDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CloudHSMV2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudhsm_v2_hsm" {
			continue
		}

		hsm, err := tfcloudhsmv2.FindHSM(conn, rs.Primary.ID, rs.Primary.Attributes["hsm_eni_id"])

		if err != nil {
			return err
		}

		if hsm != nil && aws.StringValue(hsm.State) != "DELETED" {
			return fmt.Errorf("HSM still exists:\n%s", hsm)
		}
	}

	return nil
}

func testAccCheckHSMExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudHSMV2Conn

		it, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		_, err := tfcloudhsmv2.FindHSM(conn, it.Primary.ID, it.Primary.Attributes["hsm_eni_id"])
		if err != nil {
			return fmt.Errorf("CloudHSM cluster not found: %s", err)
		}

		return nil
	}
}
