package redshift_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/redshift"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccRedshiftSecurityGroup_basic(t *testing.T) {
	var v redshift.ClusterSecurityGroup
	rInt := sdkacctest.RandInt()
	resourceName := "aws_redshift_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckEC2Classic(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupConfig_ingressCIDR(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(resourceName, &v),
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

func TestAccRedshiftSecurityGroup_ingressCIDR(t *testing.T) {
	var v redshift.ClusterSecurityGroup
	rInt := sdkacctest.RandInt()
	resourceName := "aws_redshift_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckEC2Classic(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupConfig_ingressCIDR(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "name", fmt.Sprintf("redshift-sg-terraform-%d", rInt)),
					resource.TestCheckResourceAttr(
						resourceName, "description", "Managed by Terraform"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"cidr": "10.0.0.1/24",
					}),
					resource.TestCheckResourceAttr(
						resourceName, "ingress.#", "1"),
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

func TestAccRedshiftSecurityGroup_updateIngressCIDR(t *testing.T) {
	var v redshift.ClusterSecurityGroup
	rInt := sdkacctest.RandInt()
	resourceName := "aws_redshift_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckEC2Classic(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupConfig_ingressCIDR(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "ingress.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSecurityGroupConfig_ingressCIDRAdd(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "ingress.#", "3"),
				),
			},
			{
				Config: testAccSecurityGroupConfig_ingressCIDRReduce(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "ingress.#", "2"),
				),
			},
		},
	})
}

func TestAccRedshiftSecurityGroup_ingressSecurityGroup(t *testing.T) {
	var v redshift.ClusterSecurityGroup
	rInt := sdkacctest.RandInt()
	resourceName := "aws_redshift_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckEC2Classic(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupConfig_ingressSgID(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "name", fmt.Sprintf("redshift-sg-terraform-%d", rInt)),
					resource.TestCheckResourceAttr(
						resourceName, "description", "this is a description"),
					resource.TestCheckResourceAttr(
						resourceName, "ingress.#", "1"),
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

func TestAccRedshiftSecurityGroup_updateIngressSecurityGroup(t *testing.T) {
	var v redshift.ClusterSecurityGroup
	rInt := sdkacctest.RandInt()
	resourceName := "aws_redshift_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckEC2Classic(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupConfig_ingressSgID(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "ingress.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSecurityGroupConfig_ingressSgIdAdd(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "ingress.#", "3"),
				),
			},
			{
				Config: testAccSecurityGroupConfig_ingressSgIdReduce(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "ingress.#", "2"),
				),
			},
		},
	})
}

func testAccCheckSecurityGroupExists(n string, v *redshift.ClusterSecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Redshift Security Group ID is set")
		}

		conn := acctest.ProviderEC2Classic.Meta().(*conns.AWSClient).RedshiftConn

		opts := redshift.DescribeClusterSecurityGroupsInput{
			ClusterSecurityGroupName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeClusterSecurityGroups(&opts)

		if err != nil {
			return err
		}

		if len(resp.ClusterSecurityGroups) != 1 ||
			*resp.ClusterSecurityGroups[0].ClusterSecurityGroupName != rs.Primary.ID {
			return fmt.Errorf("Redshift Security Group not found")
		}

		*v = *resp.ClusterSecurityGroups[0]

		return nil
	}
}

func testAccCheckSecurityGroupDestroy(s *terraform.State) error {
	conn := acctest.ProviderEC2Classic.Meta().(*conns.AWSClient).RedshiftConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_redshift_security_group" {
			continue
		}

		// Try to find the Group
		resp, err := conn.DescribeClusterSecurityGroups(
			&redshift.DescribeClusterSecurityGroupsInput{
				ClusterSecurityGroupName: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if len(resp.ClusterSecurityGroups) != 0 &&
				*resp.ClusterSecurityGroups[0].ClusterSecurityGroupName == rs.Primary.ID {
				return fmt.Errorf("Redshift Security Group still exists")
			}
		}

		// Verify the error
		newerr, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if newerr.Code() != "ClusterSecurityGroupNotFound" {
			return err
		}
	}

	return nil
}

func testAccSecurityGroupConfig_ingressCIDR(rInt int) string {
	return acctest.ConfigCompose(
		acctest.ConfigEC2ClassicRegionProvider(),
		fmt.Sprintf(`
resource "aws_redshift_security_group" "test" {
  name = "redshift-sg-terraform-%d"

  ingress {
    cidr = "10.0.0.1/24"
  }
}
`, rInt))
}

func testAccSecurityGroupConfig_ingressCIDRAdd(rInt int) string {
	return acctest.ConfigCompose(
		acctest.ConfigEC2ClassicRegionProvider(),
		fmt.Sprintf(`
resource "aws_redshift_security_group" "test" {
  name        = "redshift-sg-terraform-%d"
  description = "this is a description"

  ingress {
    cidr = "10.0.0.1/24"
  }

  ingress {
    cidr = "10.0.10.1/24"
  }

  ingress {
    cidr = "10.0.20.1/24"
  }
}
`, rInt))
}

func testAccSecurityGroupConfig_ingressCIDRReduce(rInt int) string {
	return acctest.ConfigCompose(
		acctest.ConfigEC2ClassicRegionProvider(),
		fmt.Sprintf(`
resource "aws_redshift_security_group" "test" {
  name        = "redshift-sg-terraform-%d"
  description = "this is a description"

  ingress {
    cidr = "10.0.0.1/24"
  }

  ingress {
    cidr = "10.0.10.1/24"
  }
}
`, rInt))
}

func testAccSecurityGroupConfig_ingressSgID(rInt int) string {
	return acctest.ConfigCompose(
		acctest.ConfigEC2ClassicRegionProvider(),
		fmt.Sprintf(`
resource "aws_security_group" "redshift" {
  name        = "terraform_redshift_test_%d"
  description = "Used in the redshift acceptance tests"

  ingress {
    protocol    = "tcp"
    from_port   = 22
    to_port     = 22
    cidr_blocks = ["10.0.0.0/8"]
  }
}

resource "aws_redshift_security_group" "test" {
  name        = "redshift-sg-terraform-%d"
  description = "this is a description"

  ingress {
    security_group_name     = aws_security_group.redshift.name
    security_group_owner_id = aws_security_group.redshift.owner_id
  }
}
`, rInt, rInt))
}

func testAccSecurityGroupConfig_ingressSgIdAdd(rInt int) string {
	return acctest.ConfigCompose(
		acctest.ConfigEC2ClassicRegionProvider(),
		fmt.Sprintf(`
resource "aws_security_group" "redshift" {
  name        = "terraform_redshift_test_%d"
  description = "Used in the redshift acceptance tests"

  ingress {
    protocol    = "tcp"
    from_port   = 22
    to_port     = 22
    cidr_blocks = ["10.0.0.0/16"]
  }
}

resource "aws_security_group" "redshift2" {
  name        = "terraform_redshift_test_2_%d"
  description = "Used in the redshift acceptance tests #2"

  ingress {
    protocol    = "tcp"
    from_port   = 22
    to_port     = 22
    cidr_blocks = ["10.1.0.0/16"]
  }
}

resource "aws_security_group" "redshift3" {
  name        = "terraform_redshift_test_3_%d"
  description = "Used in the redshift acceptance tests #3"

  ingress {
    protocol    = "tcp"
    from_port   = 22
    to_port     = 22
    cidr_blocks = ["10.2.0.0/16"]
  }
}

resource "aws_redshift_security_group" "test" {
  name        = "redshift-sg-terraform-%d"
  description = "this is a description"

  ingress {
    security_group_name     = aws_security_group.redshift.name
    security_group_owner_id = aws_security_group.redshift.owner_id
  }

  ingress {
    security_group_name     = aws_security_group.redshift2.name
    security_group_owner_id = aws_security_group.redshift.owner_id
  }

  ingress {
    security_group_name     = aws_security_group.redshift3.name
    security_group_owner_id = aws_security_group.redshift.owner_id
  }
}
`, rInt, rInt, rInt, rInt))
}

func testAccSecurityGroupConfig_ingressSgIdReduce(rInt int) string {
	return acctest.ConfigCompose(
		acctest.ConfigEC2ClassicRegionProvider(),
		fmt.Sprintf(`
resource "aws_security_group" "redshift" {
  name        = "terraform_redshift_test_%d"
  description = "Used in the redshift acceptance tests"

  ingress {
    protocol    = "tcp"
    from_port   = 22
    to_port     = 22
    cidr_blocks = ["10.0.0.0/16"]
  }
}

resource "aws_security_group" "redshift2" {
  name        = "terraform_redshift_test_2_%d"
  description = "Used in the redshift acceptance tests #2"

  ingress {
    protocol    = "tcp"
    from_port   = 22
    to_port     = 22
    cidr_blocks = ["10.1.0.0/16"]
  }
}

resource "aws_redshift_security_group" "test" {
  name        = "redshift-sg-terraform-%d"
  description = "this is a description"

  ingress {
    security_group_name     = aws_security_group.redshift.name
    security_group_owner_id = aws_security_group.redshift.owner_id
  }

  ingress {
    security_group_name     = aws_security_group.redshift2.name
    security_group_owner_id = aws_security_group.redshift.owner_id
  }
}
`, rInt, rInt, rInt))
}
