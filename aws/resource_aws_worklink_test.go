package aws

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/worklink"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSWorkLink_Basic(t *testing.T) {
	suffix := randomString(10)
	resourceName := "aws_worklink.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWorkLinkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWorkLinkConfig(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWorkLinkExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "optimize_for_end_user_location", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "company_code"),
					resource.TestCheckResourceAttrSet(resourceName, "created_time"),
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

func TestAccAWSWorkLink_DisplayName(t *testing.T) {
	suffix := randomString(10)
	resourceName := "aws_worklink.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWorkLinkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWorkLinkConfigDisplayName(suffix, "display1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWorkLinkExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "display_name", "display1"),
				),
			},
			{
				Config: testAccAWSWorkLinkConfigDisplayName(suffix, "display2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWorkLinkExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "display_name", "display2"),
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

func TestAccAWSWorkLink_OptimizeForEndUserLocation(t *testing.T) {
	suffix := randomString(10)
	resourceName := "aws_worklink.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWorkLinkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWorkLinkConfigOptimizeForEndUserLocation(suffix, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWorkLinkExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "optimize_for_end_user_location", "false"),
				),
			},
			{
				Config: testAccAWSWorkLinkConfigOptimizeForEndUserLocation(suffix, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWorkLinkExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "optimize_for_end_user_location", "true"),
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

func TestAccAWSWorkLink_AuditStreamArn(t *testing.T) {
	rName := randomString(10)
	resourceName := "aws_worklink.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWorkLinkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWorkLinkConfigAuditStreamArn(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWorkLinkExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "audit_stream_arn", "aws_kinesis_stream.test_stream", "arn"),
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

func TestAccAWSWorkLink_Network(t *testing.T) {
	rName := randomString(10)
	resourceName := "aws_worklink.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWorkLinkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWorkLinkConfigNetwork(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWorkLinkExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "network.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "network.0.vpc_id", "aws_vpc.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "network.0.subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "network.0.security_group_ids.#", "1"),
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

func TestAccAWSWorkLink_DeviceCaCertificate(t *testing.T) {
	rName := randomString(10)
	resourceName := "aws_worklink.test"
	fName := "test-fixtures/worklink-device-ca-certificate.pem"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWorkLinkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWorkLinkConfigDeviceCaCertificate(rName, fName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWorkLinkExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "device_ca_certificate", regexp.MustCompile("^-----BEGIN CERTIFICATE-----")),
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

func TestAccAWSWorkLink_IdentityProvider(t *testing.T) {
	rName := randomString(10)
	resourceName := "aws_worklink.test"
	fName := "test-fixtures/saml-metadata.xml"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWorkLinkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWorkLinkConfigIdentityProvider(rName, fName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWorkLinkExists(resourceName),
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

func TestAccAWSWorkLink_Disappears(t *testing.T) {
	rName := randomString(10)
	resourceName := "aws_worklink.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWorkLinkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWorkLinkConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWorkLinkExists(resourceName),
					testAccCheckAWSWorkLinkDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSWorkLinkDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No resource ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).worklinkconn

		input := &worklink.DeleteFleetInput{
			FleetArn: aws.String(rs.Primary.ID),
		}

		if _, err := conn.DeleteFleet(input); err != nil {
			return err
		}

		stateConf := &resource.StateChangeConf{
			Pending:    []string{"DELETING"},
			Target:     []string{"DELETED"},
			Refresh:    worklinkFleetStateRefresh(conn, rs.Primary.ID),
			Timeout:    15 * time.Minute,
			Delay:      10 * time.Second,
			MinTimeout: 3 * time.Second,
		}

		_, err := stateConf.WaitForState()

		return err
	}
}

func testAccCheckAWSWorkLinkDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).worklinkconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_worklink" {
			continue
		}

		_, err := conn.DescribeFleetMetadata(
			&worklink.DescribeFleetMetadataInput{
				FleetArn: aws.String(rs.Primary.ID),
			})

		if err != nil {
			// Return nil if the Worklink Fleet is already destroyed
			if isAWSErr(err, worklink.ErrCodeResourceNotFoundException, "") {
				return nil
			}

			return err
		}
		return fmt.Errorf("Worklink Fleet %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAWSWorkLinkExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Worklink Fleet ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).worklinkconn
		_, err := conn.DescribeFleetMetadata(&worklink.DescribeFleetMetadataInput{
			FleetArn: aws.String(rs.Primary.ID),
		})

		return err
	}
}

func testAccAWSWorkLinkConfig(r string) string {
	return fmt.Sprintf(`

resource "aws_worklink" "test" {
	name = "tf-worklink-fleet-%s"
}

`, r)
}

func testAccAWSWorkLinkConfigDisplayName(r, displayName string) string {
	return fmt.Sprintf(`

resource "aws_worklink" "test" {
	name = "tf-worklink-fleet-%s"
	display_name = "%s"
}

`, r, displayName)
}

func testAccAWSWorkLinkConfigOptimizeForEndUserLocation(r string, b bool) string {
	return fmt.Sprintf(`

resource "aws_worklink" "test" {
	name = "tf-worklink-fleet-%s"
	optimize_for_end_user_location = %t
}

`, r, b)
}

func testAccAWSWorkLinkConfigNetwork_Base(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {}

resource "aws_vpc" "test" {
	cidr_block = "192.168.0.0/16"

	tags = {
		Name = %q
	}
}

resource "aws_security_group" "test" {
	name = "tf_test_foo"
	description = "foo"
	vpc_id="${aws_vpc.test.id}"
  
	ingress {
	  protocol = "icmp"
	  from_port = -1
	  to_port = -1
	  self = true
	}
}

resource "aws_subnet" "test" {
	count = 2

	availability_zone = "${data.aws_availability_zones.available.names[count.index]}"
	cidr_block        = "192.168.${count.index}.0/24"
	vpc_id            = "${aws_vpc.test.id}"

	tags = {
		Name = %q
	}
}
`, rName, rName)
}

func testAccAWSWorkLinkConfigNetwork(r string) string {
	return fmt.Sprintf(`
%s

resource "aws_worklink" "test" {
	name = "tf-worklink-fleet-%s"

	network {
		vpc_id = "${aws_vpc.test.id}"
		subnet_ids = ["${aws_subnet.test.*.id}"]
		security_group_ids = ["${aws_security_group.test.id}"]
	}
}

`, testAccAWSWorkLinkConfigNetwork_Base(r), r)
}

func testAccAWSWorkLinkConfigAuditStreamArn(r string) string {
	return fmt.Sprintf(`

resource "aws_kinesis_stream" "test_stream" {
	name = "%s_kinesis_test"
	shard_count = 1
}

resource "aws_worklink" "test" {
	name = "tf-worklink-fleet-%s"

	audit_stream_arn = "${aws_kinesis_stream.test_stream.arn}"
}

`, r, r)
}

func testAccAWSWorkLinkConfigDeviceCaCertificate(r string, fName string) string {
	return fmt.Sprintf(`

resource "aws_worklink" "test" {
	name = "tf-worklink-fleet-%s"

	device_ca_certificate = "${file("%s")}"
}

`, r, fName)
}

func testAccAWSWorkLinkConfigIdentityProvider(r string, fName string) string {
	return fmt.Sprintf(`

resource "aws_worklink" "test" {
	name = "tf-worklink-fleet-%s"

	identity_provider {
		type = "SAML"
		saml_metadata = "${file("%s")}"
	}
}

`, r, fName)
}
