package rds_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/rds"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccRDSSecurityGroup_basic(t *testing.T) {
	var v rds.DBSecurityGroup
	resourceName := "aws_db_security_group.test"
	rName := fmt.Sprintf("tf-acc-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckEC2Classic(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(resourceName, &v),
					testAccCheckSecurityGroupAttributes(&v),
					acctest.CheckResourceAttrRegionalARNEC2Classic(resourceName, "arn", "rds", fmt.Sprintf("secgrp:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "Managed by Terraform"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"cidr": "10.0.0.1/24",
					}),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
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

func testAccCheckSecurityGroupDestroy(s *terraform.State) error {
	conn := acctest.ProviderEC2Classic.Meta().(*conns.AWSClient).RDSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_db_security_group" {
			continue
		}

		// Try to find the Group
		resp, err := conn.DescribeDBSecurityGroups(
			&rds.DescribeDBSecurityGroupsInput{
				DBSecurityGroupName: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if len(resp.DBSecurityGroups) != 0 &&
				*resp.DBSecurityGroups[0].DBSecurityGroupName == rs.Primary.ID {
				return fmt.Errorf("DB Security Group still exists")
			}
		}

		// Verify the error
		newerr, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if newerr.Code() != "DBSecurityGroupNotFound" {
			return err
		}
	}

	return nil
}

func testAccCheckSecurityGroupAttributes(group *rds.DBSecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(group.IPRanges) == 0 {
			return fmt.Errorf("no cidr: %#v", group.IPRanges)
		}

		if *group.IPRanges[0].CIDRIP != "10.0.0.1/24" {
			return fmt.Errorf("bad cidr: %#v", group.IPRanges)
		}

		statuses := make([]string, 0, len(group.IPRanges))
		for _, ips := range group.IPRanges {
			statuses = append(statuses, *ips.Status)
		}

		if statuses[0] != "authorized" {
			return fmt.Errorf("bad status: %#v", statuses)
		}

		return nil
	}
}

func testAccCheckSecurityGroupExists(n string, v *rds.DBSecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DB Security Group ID is set")
		}

		conn := acctest.ProviderEC2Classic.Meta().(*conns.AWSClient).RDSConn

		opts := rds.DescribeDBSecurityGroupsInput{
			DBSecurityGroupName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeDBSecurityGroups(&opts)

		if err != nil {
			return err
		}

		if len(resp.DBSecurityGroups) != 1 ||
			*resp.DBSecurityGroups[0].DBSecurityGroupName != rs.Primary.ID {
			return fmt.Errorf("DB Security Group not found")
		}

		*v = *resp.DBSecurityGroups[0]

		return nil
	}
}

func testAccSecurityGroupConfig_basic(name string) string {
	return acctest.ConfigCompose(
		acctest.ConfigEC2ClassicRegionProvider(),
		fmt.Sprintf(`
resource "aws_db_security_group" "test" {
  name = "%s"

  ingress {
    cidr = "10.0.0.1/24"
  }

  tags = {
    foo = "test"
  }
}
`, name))
}
