package ec2_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccEC2DefaultVPC_basic(t *testing.T) {
	var vpc ec2.Vpc

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDefaultVPCDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig: func() { destroyDefaultVPCBeforeTest(t) },
				Config:    testAccDefaultVPCBasicConfig,
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists("aws_default_vpc.foo", &vpc),
					resource.TestCheckResourceAttr("aws_default_vpc.foo", "cidr_block", "172.31.0.0/16"),
					resource.TestCheckResourceAttr(
						"aws_default_vpc.foo", "cidr_block", "172.31.0.0/16"),
					resource.TestCheckResourceAttr(
						"aws_default_vpc.foo", "tags.%", "1"),
					resource.TestCheckResourceAttr(
						"aws_default_vpc.foo", "tags.Name", "Default VPC"),
					resource.TestCheckResourceAttrSet(
						"aws_default_vpc.foo", "arn"),
					resource.TestCheckResourceAttr(
						"aws_default_vpc.foo", "assign_generated_ipv6_cidr_block", "false"),
					resource.TestCheckResourceAttr(
						"aws_default_vpc.foo", "ipv6_association_id", ""),
					resource.TestCheckResourceAttr(
						"aws_default_vpc.foo", "ipv6_cidr_block", ""),
					acctest.CheckResourceAttrAccountID("aws_default_vpc.foo", "owner_id"),
				),
			},
		},
	})
}

func destroyDefaultVPCBeforeTest(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
	req := &ec2.DescribeVpcsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("isDefault"),
				Values: aws.StringSlice([]string{"true"}),
			},
		},
	}

	resp, err := conn.DescribeVpcs(req)
	if err != nil {
		t.Fatalf("Error describing default vpcs: %s", err)
	}

	if resp.Vpcs == nil || len(resp.Vpcs) == 0 {
		return // No vpc, all good
	}

	//delete all subnets
	reqDS := &ec2.DescribeSubnetsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: aws.StringSlice([]string{aws.StringValue(resp.Vpcs[0].VpcId)}),
			},
		},
	}

	respDS, err := conn.DescribeSubnets(reqDS)
	if err != nil {
		t.Fatalf("Error describing default subnets: %s", err)
	}
	for _, subnet := range respDS.Subnets {
		_, err := conn.DeleteSubnet(&ec2.DeleteSubnetInput{
			SubnetId: subnet.SubnetId,
		})
		if err != nil {
			t.Fatalf("Error deleting default subnet: %s", err)
		}
	}

	//delete internet gateway
	reqIG := &ec2.DescribeInternetGatewaysInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("attachment.vpc-id"),
				Values: aws.StringSlice([]string{aws.StringValue(resp.Vpcs[0].VpcId)}),
			},
		},
	}

	respIG, err := conn.DescribeInternetGateways(reqIG)
	if err != nil {
		t.Fatalf("Error describing default internet gateways: %s", err)
	}
	for _, ig := range respIG.InternetGateways {
		_, err := conn.DetachInternetGateway(&ec2.DetachInternetGatewayInput{
			InternetGatewayId: ig.InternetGatewayId,
			VpcId:             ig.Attachments[0].VpcId,
		})
		if err != nil {
			t.Fatalf("Error detaching default internet gateway: %s", err)
		}
		_, err = conn.DeleteInternetGateway(&ec2.DeleteInternetGatewayInput{
			InternetGatewayId: ig.InternetGatewayId,
		})
		if err != nil {
			t.Fatalf("Error deleting default internet gateway: %s", err)
		}
	}

	delReq := &ec2.DeleteVpcInput{
		VpcId: resp.Vpcs[0].VpcId,
	}

	_, err = conn.DeleteVpc(delReq)
	if err != nil {
		t.Fatalf("Error deleting default VPC: %s", err)
	}
}

func testAccCheckDefaultVPCDestroy(s *terraform.State) error {
	// We expect VPC to still exist
	return nil
}

const testAccDefaultVPCBasicConfig = `
resource "aws_default_vpc" "foo" {
  tags = {
    Name = "Default VPC"
  }
}
`
