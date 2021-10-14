package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/ec2/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
)

func init() {
	resource.AddTestSweepers("aws_ec2_host", &resource.Sweeper{
		Name: "aws_ec2_host",
		F:    testSweepEc2Hosts,
		Dependencies: []string{
			"aws_instance",
		},
	})
}

func testSweepEc2Hosts(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).ec2conn
	input := &ec2.DescribeHostsInput{}
	sweepResources := make([]*testSweepResource, 0)

	err = conn.DescribeHostsPages(input, func(page *ec2.DescribeHostsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, host := range page.Hosts {
			r := resourceAwsEc2Host()
			d := r.Data(nil)
			d.SetId(aws.StringValue(host.HostId))

			sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 Host sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing EC2 Hosts (%s): %w", region, err)
	}

	err = testSweepResourceOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 Hosts (%s): %w", region, err)
	}

	return nil
}

func TestAccAWSEc2Host_basic(t *testing.T) {
	var host ec2.Host
	resourceName := "aws_ec2_host.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEc2HostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2HostConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2HostExists(resourceName, &host),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`dedicated-host/.+`)),
					resource.TestCheckResourceAttr(resourceName, "auto_placement", "on"),
					resource.TestCheckResourceAttr(resourceName, "host_recovery", "off"),
					resource.TestCheckResourceAttr(resourceName, "instance_family", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "a1.large"),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccAWSEc2Host_disappears(t *testing.T) {
	var host ec2.Host
	resourceName := "aws_ec2_host.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEc2HostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2HostConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2HostExists(resourceName, &host),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsEc2Host(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSEc2Host_InstanceFamily(t *testing.T) {
	var host ec2.Host
	resourceName := "aws_ec2_host.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEc2HostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2HostConfigInstanceFamily(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2HostExists(resourceName, &host),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`dedicated-host/.+`)),
					resource.TestCheckResourceAttr(resourceName, "auto_placement", "off"),
					resource.TestCheckResourceAttr(resourceName, "host_recovery", "on"),
					resource.TestCheckResourceAttr(resourceName, "instance_family", "c5"),
					resource.TestCheckResourceAttr(resourceName, "instance_type", ""),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEc2HostConfigInstanceType(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2HostExists(resourceName, &host),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`dedicated-host/.+`)),
					resource.TestCheckResourceAttr(resourceName, "auto_placement", "on"),
					resource.TestCheckResourceAttr(resourceName, "host_recovery", "off"),
					resource.TestCheckResourceAttr(resourceName, "instance_family", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "c5.xlarge"),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
		},
	})
}

func TestAccAWSEc2Host_Tags(t *testing.T) {
	var host ec2.Host
	resourceName := "aws_ec2_host.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEc2HostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2HostConfigTags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2HostExists(resourceName, &host),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEc2HostConfigTags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2HostExists(resourceName, &host),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSEc2HostConfigTags1("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2HostExists(resourceName, &host),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckEc2HostExists(n string, v *ec2.Host) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Host ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		output, err := finder.HostByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckEc2HostDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_host" {
			continue
		}

		_, err := finder.HostByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EC2 Host %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccAWSEc2HostConfig() string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), `
resource "aws_ec2_host" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  instance_type     = "a1.large"
}
`)
}

func testAccAWSEc2HostConfigInstanceFamily(rName string) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_ec2_host" "test" {
  auto_placement    = "off"
  availability_zone = data.aws_availability_zones.available.names[0]
  host_recovery     = "on"
  instance_family   = "c5"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccAWSEc2HostConfigInstanceType(rName string) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_ec2_host" "test" {
  auto_placement    = "on"
  availability_zone = data.aws_availability_zones.available.names[0]
  host_recovery     = "off"
  instance_type     = "c5.xlarge"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccAWSEc2HostConfigTags1(tagKey1, tagValue1 string) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_ec2_host" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  instance_type     = "a1.large"

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccAWSEc2HostConfigTags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_ec2_host" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  instance_type     = "a1.large"

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}
