package aws

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_vpc_peering_connection", &resource.Sweeper{
		Name: "aws_vpc_peering_connection",
		F:    testSweepEc2VpcPeeringConnections,
	})
}

func testSweepEc2VpcPeeringConnections(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*AWSClient).ec2conn
	input := &ec2.DescribeVpcPeeringConnectionsInput{}

	err = conn.DescribeVpcPeeringConnectionsPages(input, func(page *ec2.DescribeVpcPeeringConnectionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, vpcPeeringConnection := range page.VpcPeeringConnections {
			deletedStatuses := map[string]bool{
				ec2.VpcPeeringConnectionStateReasonCodeDeleted:  true,
				ec2.VpcPeeringConnectionStateReasonCodeExpired:  true,
				ec2.VpcPeeringConnectionStateReasonCodeFailed:   true,
				ec2.VpcPeeringConnectionStateReasonCodeRejected: true,
			}

			if _, ok := deletedStatuses[aws.StringValue(vpcPeeringConnection.Status.Code)]; ok {
				continue
			}

			id := aws.StringValue(vpcPeeringConnection.VpcPeeringConnectionId)
			input := &ec2.DeleteVpcPeeringConnectionInput{
				VpcPeeringConnectionId: vpcPeeringConnection.VpcPeeringConnectionId,
			}

			log.Printf("[INFO] Deleting EC2 VPC Peering Connection: %s", id)

			_, err := conn.DeleteVpcPeeringConnection(input)

			if isAWSErr(err, "InvalidVpcPeeringConnectionID.NotFound", "") {
				continue
			}

			if err != nil {
				log.Printf("[ERROR] Error deleting EC2 VPC Peering Connection (%s): %s", id, err)
				continue
			}

			if err := waitForEc2VpcPeeringConnectionDeletion(conn, id, 5*time.Minute); err != nil {
				log.Printf("[ERROR] Error waiting for EC2 VPC Peering Connection (%s) to be deleted: %s", id, err)
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 VPC Peering Connection sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error describing EC2 VPC Peering Connections: %s", err)
	}

	return nil
}

func TestAccAWSVPCPeeringConnection_basic(t *testing.T) {
	var connection ec2.VpcPeeringConnection
	resourceName := "aws_vpc_peering_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshName:   resourceName,
		IDRefreshIgnore: []string{"auto_accept"},

		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSVpcPeeringConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcPeeringConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSVpcPeeringConnectionExists(
						resourceName,
						&connection),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"auto_accept"},
			},
		},
	})
}

func TestAccAWSVPCPeeringConnection_plan(t *testing.T) {
	var connection ec2.VpcPeeringConnection
	resourceName := "aws_vpc_peering_connection.test"

	// reach out and DELETE the VPC Peering connection outside of Terraform
	testDestroy := func(*terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		log.Printf("[DEBUG] Test deleting the VPC Peering Connection.")
		_, err := conn.DeleteVpcPeeringConnection(
			&ec2.DeleteVpcPeeringConnectionInput{
				VpcPeeringConnectionId: connection.VpcPeeringConnectionId,
			})

		return err
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshIgnore: []string{"auto_accept"},
		Providers:       testAccProviders,
		CheckDestroy:    testAccCheckAWSVpcPeeringConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcPeeringConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSVpcPeeringConnectionExists(
						resourceName,
						&connection),
					testDestroy,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSVPCPeeringConnection_tags(t *testing.T) {
	var connection ec2.VpcPeeringConnection
	resourceName := "aws_vpc_peering_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshName:   resourceName,
		IDRefreshIgnore: []string{"auto_accept"},

		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcPeeringConfigTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSVpcPeeringConnectionExists(
						resourceName,
						&connection),
					testAccCheckTags(&connection.Tags, "test", "bar"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"auto_accept"},
			},
		},
	})
}

func TestAccAWSVPCPeeringConnection_options(t *testing.T) {
	var connection ec2.VpcPeeringConnection
	resourceName := "aws_vpc_peering_connection.test"

	testAccepterChange := func(*terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		log.Printf("[DEBUG] Test change to the VPC Peering Connection Options.")

		_, err := conn.ModifyVpcPeeringConnectionOptions(
			&ec2.ModifyVpcPeeringConnectionOptionsInput{
				VpcPeeringConnectionId: connection.VpcPeeringConnectionId,
				AccepterPeeringConnectionOptions: &ec2.PeeringConnectionOptionsRequest{
					AllowDnsResolutionFromRemoteVpc: aws.Bool(false),
				},
			})

		return err
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshName:   resourceName,
		IDRefreshIgnore: []string{"auto_accept"},

		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSVpcPeeringConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcPeeringConfigOptions,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSVpcPeeringConnectionExists(
						resourceName,
						&connection),
					resource.TestCheckResourceAttr(
						resourceName,
						"accepter.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName,
						"accepter.1102046665.allow_remote_vpc_dns_resolution", "true"),
					testAccCheckAWSVpcPeeringConnectionOptions(
						resourceName, "accepter",
						&ec2.VpcPeeringConnectionOptionsDescription{
							AllowDnsResolutionFromRemoteVpc:            aws.Bool(true),
							AllowEgressFromLocalClassicLinkToRemoteVpc: aws.Bool(false),
							AllowEgressFromLocalVpcToRemoteClassicLink: aws.Bool(false),
						}),
					resource.TestCheckResourceAttr(
						resourceName,
						"requester.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName,
						"requester.41753983.allow_classic_link_to_remote_vpc", "true"),
					resource.TestCheckResourceAttr(
						resourceName,
						"requester.41753983.allow_vpc_to_remote_classic_link", "true"),
					testAccCheckAWSVpcPeeringConnectionOptions(
						resourceName, "requester",
						&ec2.VpcPeeringConnectionOptionsDescription{
							AllowDnsResolutionFromRemoteVpc:            aws.Bool(false),
							AllowEgressFromLocalClassicLinkToRemoteVpc: aws.Bool(true),
							AllowEgressFromLocalVpcToRemoteClassicLink: aws.Bool(true),
						},
					),
					testAccepterChange,
				),
				ExpectNonEmptyPlan: true,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"auto_accept"},
			},
			{
				Config: testAccVpcPeeringConfigOptions,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSVpcPeeringConnectionExists(
						resourceName,
						&connection),
					resource.TestCheckResourceAttr(
						resourceName,
						"accepter.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName,
						"accepter.1102046665.allow_remote_vpc_dns_resolution", "true"),
					testAccCheckAWSVpcPeeringConnectionOptions(
						resourceName, "accepter",
						&ec2.VpcPeeringConnectionOptionsDescription{
							AllowDnsResolutionFromRemoteVpc:            aws.Bool(true),
							AllowEgressFromLocalClassicLinkToRemoteVpc: aws.Bool(false),
							AllowEgressFromLocalVpcToRemoteClassicLink: aws.Bool(false),
						},
					),
				),
			},
		},
	})
}

func TestAccAWSVPCPeeringConnection_failedState(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshIgnore: []string{"auto_accept"},
		Providers:       testAccProviders,
		CheckDestroy:    testAccCheckAWSVpcPeeringConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccVpcPeeringConfigFailedState,
				ExpectError: regexp.MustCompile(`.*Error waiting.*\(pcx-\w+\).*incorrect.*VPC-ID.*`),
			},
		},
	})
}

func testAccCheckAWSVpcPeeringConnectionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc_peering_connection" {
			continue
		}

		describe, err := conn.DescribeVpcPeeringConnections(
			&ec2.DescribeVpcPeeringConnectionsInput{
				VpcPeeringConnectionIds: []*string{aws.String(rs.Primary.ID)},
			})

		if err != nil {
			return err
		}

		var pc *ec2.VpcPeeringConnection
		for _, c := range describe.VpcPeeringConnections {
			if rs.Primary.ID == *c.VpcPeeringConnectionId {
				pc = c
			}
		}

		if pc == nil {
			// not found
			return nil
		}

		if pc.Status != nil {
			if *pc.Status.Code == "deleted" || *pc.Status.Code == "rejected" {
				return nil
			}
			return fmt.Errorf("Found the VPC Peering Connection in an unexpected state: %s", pc)
		}

		// return error here; we've found the vpc_peering object we want, however
		// it's not in an expected state
		return fmt.Errorf("Fall through error for testAccCheckAWSVpcPeeringConnectionDestroy.")
	}

	return nil
}

func testAccCheckAWSVpcPeeringConnectionExists(n string, connection *ec2.VpcPeeringConnection) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VPC Peering Connection ID is set.")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		resp, err := conn.DescribeVpcPeeringConnections(
			&ec2.DescribeVpcPeeringConnectionsInput{
				VpcPeeringConnectionIds: []*string{aws.String(rs.Primary.ID)},
			})
		if err != nil {
			return err
		}
		if len(resp.VpcPeeringConnections) == 0 {
			return fmt.Errorf("VPC Peering Connection could not be found")
		}

		*connection = *resp.VpcPeeringConnections[0]

		return nil
	}
}

func testAccCheckAWSVpcPeeringConnectionOptions(n, block string, options *ec2.VpcPeeringConnectionOptionsDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VPC Peering Connection ID is set.")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		resp, err := conn.DescribeVpcPeeringConnections(
			&ec2.DescribeVpcPeeringConnectionsInput{
				VpcPeeringConnectionIds: []*string{aws.String(rs.Primary.ID)},
			})
		if err != nil {
			return err
		}

		pc := resp.VpcPeeringConnections[0]

		o := pc.AccepterVpcInfo
		if block == "requester" {
			o = pc.RequesterVpcInfo
		}

		if !reflect.DeepEqual(o.PeeringOptions, options) {
			return fmt.Errorf("Expected the VPC Peering Connection Options to be %#v, got %#v",
				options, o.PeeringOptions)
		}

		return nil
	}
}

func TestAccAWSVPCPeeringConnection_peerRegionAndAutoAccept(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshIgnore: []string{"auto_accept"},
		Providers:       testAccProviders,
		CheckDestroy:    testAccCheckAWSVpcPeeringConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccVpcPeeringConfigRegionAutoAccept,
				ExpectError: regexp.MustCompile(`.*peer_region cannot be set whilst auto_accept is true when creating a vpc peering connection.*`),
			},
		},
	})
}

func TestAccAWSVPCPeeringConnection_region(t *testing.T) {
	var connection ec2.VpcPeeringConnection
	var providers []*schema.Provider
	resourceName := "aws_vpc_peering_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshName:   resourceName,
		IDRefreshIgnore: []string{"auto_accept"},

		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckAWSVpcPeeringConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcPeeringConfigRegion,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSVpcPeeringConnectionExists(
						resourceName,
						&connection),
				),
			},
		},
	})
}

const testAccVpcPeeringConfig = `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "terraform-testacc-vpc-peering-conn-test"
  }
}

resource "aws_vpc" "bar" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = "terraform-testacc-vpc-peering-conn-bar"
  }
}

resource "aws_vpc_peering_connection" "test" {
  vpc_id = "${aws_vpc.test.id}"
  peer_vpc_id = "${aws_vpc.bar.id}"
  auto_accept = true
}
`

const testAccVpcPeeringConfigTags = `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "terraform-testacc-vpc-peering-conn-tags-test"
  }
}

resource "aws_vpc" "bar" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = "terraform-testacc-vpc-peering-conn-tags-bar"
  }
}

resource "aws_vpc_peering_connection" "test" {
  vpc_id = "${aws_vpc.test.id}"
  peer_vpc_id = "${aws_vpc.bar.id}"
  auto_accept = true
  tags = {
    test = "bar"
  }
}
`

const testAccVpcPeeringConfigOptions = `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "terraform-testacc-vpc-peering-conn-options-test"
  }
}

resource "aws_vpc" "bar" {
  cidr_block = "10.1.0.0/16"
  enable_dns_hostnames = true
  tags = {
    Name = "terraform-testacc-vpc-peering-conn-options-bar"
  }
}

resource "aws_vpc_peering_connection" "test" {
  vpc_id = "${aws_vpc.test.id}"
  peer_vpc_id = "${aws_vpc.bar.id}"
  auto_accept = true

  accepter {
    allow_remote_vpc_dns_resolution = true
  }

  requester {
    allow_vpc_to_remote_classic_link = true
    allow_classic_link_to_remote_vpc = true
  }
}
`

const testAccVpcPeeringConfigFailedState = `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "terraform-testacc-vpc-peering-conn-failed-state-test"
  }
}

resource "aws_vpc" "bar" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "terraform-testacc-vpc-peering-conn-failed-state-bar"
  }
}

resource "aws_vpc_peering_connection" "test" {
  vpc_id = "${aws_vpc.test.id}"
  peer_vpc_id = "${aws_vpc.bar.id}"
}
`

const testAccVpcPeeringConfigRegionAutoAccept = `
provider "aws" {
  alias = "main"
  region = "us-west-2"
}

provider "aws" {
  alias = "peer"
  region = "us-east-1"
}

resource "aws_vpc" "test" {
  provider = "aws.main"
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "terraform-testacc-vpc-peering-conn-region-auto-accept-test"
  }
}

resource "aws_vpc" "bar" {
  provider = "aws.peer"
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = "terraform-testacc-vpc-peering-conn-region-auto-accept-bar"
  }
}

resource "aws_vpc_peering_connection" "test" {
  provider = "aws.main"
  vpc_id = "${aws_vpc.test.id}"
  peer_vpc_id = "${aws_vpc.bar.id}"
  peer_region = "us-east-1"
  auto_accept = true
}
`

const testAccVpcPeeringConfigRegion = `
provider "aws" {
  alias = "main"
  region = "us-west-2"
}

provider "aws" {
  alias = "peer"
  region = "us-east-1"
}

resource "aws_vpc" "test" {
  provider = "aws.main"
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "terraform-testacc-vpc-peering-conn-region-test"
  }
}

resource "aws_vpc" "bar" {
  provider = "aws.peer"
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = "terraform-testacc-vpc-peering-conn-region-bar"
  }
}

resource "aws_vpc_peering_connection" "test" {
  provider = "aws.main"
  vpc_id = "${aws_vpc.test.id}"
  peer_vpc_id = "${aws_vpc.bar.id}"
  peer_region = "us-east-1"
}
`
