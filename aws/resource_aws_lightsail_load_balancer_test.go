package aws

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func init() {
	resource.AddTestSweepers("aws_lightsail_load_balancer", &resource.Sweeper{
		Name: "aws_lightsail_load_balancer",
		F:    testSweepLightsailLoadBalancers,
	})
}

func testSweepLightsailLoadBalancers(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.(*AWSClient).lightsailconn

	input := &lightsail.GetLoadBalancersInput{}
	var sweeperErrs *multierror.Error

	for {
		output, err := conn.GetLoadBalancers(input)

		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Lightsail Load Balancer sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving Lightsail Load Balancers: %s", err)
		}

		for _, loadBalancer := range output.LoadBalancers {
			name := aws.StringValue(loadBalancer.Name)
			input := &lightsail.DeleteLoadBalancerInput{
				LoadBalancerName: loadBalancer.Name,
			}

			log.Printf("[INFO] Deleting Lightsail Load Balancer: %s", name)
			_, err := conn.DeleteLoadBalancer(input)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting Lightsail Load Balancer (%s): %s", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			}
		}

		if aws.StringValue(output.NextPageToken) == "" {
			break
		}

		input.PageToken = output.NextPageToken
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSLightsailLoadBalancer_basic(t *testing.T) {
	var loadBalancer lightsail.LoadBalancer
	lightsailLoadBalancerName := fmt.Sprintf("tf-test-lightsail-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t); testAccPreCheckAWSLightsail(t) },
		IDRefreshName: "aws_lightsail_load_balancer.lightsail_load_balancer_test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLightsailLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailLoadBalancerConfig_basic(lightsailLoadBalancerName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLightsailLoadBalancerExists("aws_lightsail_load_balancer.lightsail_load_balancer_test", &loadBalancer),
					resource.TestCheckResourceAttrSet("aws_lightsail_load_balancer.lightsail_load_balancer_test", "health_check_path"),
					resource.TestCheckResourceAttrSet("aws_lightsail_load_balancer.lightsail_load_balancer_test", "instance_port"),
					resource.TestCheckResourceAttr("aws_lightsail_load_balancer.lightsail_load_balancer_test", "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSLightsailLoadBalancer_Name(t *testing.T) {
	var conf lightsail.LoadBalancer
	lightsailName := fmt.Sprintf("tf-test-lightsail-%d", acctest.RandInt())
	lightsailNameWithSpaces := fmt.Sprint(lightsailName, "string with spaces")
	lightsailNameWithStartingDigit := fmt.Sprintf("01-%s", lightsailName)
	lightsailNameWithUnderscore := fmt.Sprintf("%s_123456", lightsailName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_lightsail_load_balancer.lightsail_load_balancer_test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLightsailLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSLightsailLoadBalancerConfig_basic(lightsailNameWithSpaces),
				ExpectError: regexp.MustCompile(`must contain only alphanumeric characters, underscores, hyphens, and dots`),
			},
			{
				Config:      testAccAWSLightsailLoadBalancerConfig_basic(lightsailNameWithStartingDigit),
				ExpectError: regexp.MustCompile(`must begin with an alphabetic character`),
			},
			{
				Config: testAccAWSLightsailLoadBalancerConfig_basic(lightsailName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLightsailLoadBalancerExists("aws_lightsail_load_balancer.lightsail_load_balancer_test", &conf),
					resource.TestCheckResourceAttrSet("aws_lightsail_load_balancer.lightsail_load_balancer_test", "health_check_path"),
					resource.TestCheckResourceAttrSet("aws_lightsail_load_balancer.lightsail_load_balancer_test", "instance_port"),
				),
			},
			{
				Config: testAccAWSLightsailLoadBalancerConfig_basic(lightsailNameWithUnderscore),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLightsailLoadBalancerExists("aws_lightsail_load_balancer.lightsail_load_balancer_test", &conf),
					resource.TestCheckResourceAttrSet("aws_lightsail_load_balancer.lightsail_load_balancer_test", "health_check_path"),
					resource.TestCheckResourceAttrSet("aws_lightsail_load_balancer.lightsail_load_balancer_test", "instance_port"),
				),
			},
		},
	})
}

func TestAccAWSLightsailLoadBalancer_Tags(t *testing.T) {
	var conf lightsail.LoadBalancer
	lightsailLoadBalancerName := fmt.Sprintf("tf-test-lightsail-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t); testAccPreCheckAWSLightsail(t) },
		IDRefreshName: "aws_lightsail_load_balancer.lightsail_load_balancer_test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLightsailLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailLoadBalancerConfig_tags1(lightsailLoadBalancerName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLightsailLoadBalancerExists("aws_lightsail_load_balancer.lightsail_load_balancer_test", &conf),
					resource.TestCheckResourceAttrSet("aws_lightsail_load_balancer.lightsail_load_balancer_test", "health_check_path"),
					resource.TestCheckResourceAttrSet("aws_lightsail_load_balancer.lightsail_load_balancer_test", "instance_port"),
					resource.TestCheckResourceAttr("aws_lightsail_load_balancer.lightsail_load_balancer_test", "tags.%", "1"),
				),
			},
			{
				Config: testAccAWSLightsailLoadBalancerConfig_tags2(lightsailLoadBalancerName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLightsailLoadBalancerExists("aws_lightsail_load_balancer.lightsail_load_balancer_test", &conf),
					resource.TestCheckResourceAttrSet("aws_lightsail_load_balancer.lightsail_load_balancer_test", "health_check_path"),
					resource.TestCheckResourceAttrSet("aws_lightsail_load_balancer.lightsail_load_balancer_test", "instance_port"),
					resource.TestCheckResourceAttr("aws_lightsail_load_balancer.lightsail_load_balancer_test", "tags.%", "2"),
				),
			},
		},
	})
}

func testAccCheckAWSLightsailLoadBalancerExists(n string, res *lightsail.LoadBalancer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No LightsailLoadBalancer ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).lightsailconn

		respLoadBalancer, err := conn.GetLoadBalancer(&lightsail.GetLoadBalancerInput{
			LoadBalancerName: aws.String(rs.Primary.Attributes["name"]),
		})

		if err != nil {
			return err
		}

		if respLoadBalancer == nil || respLoadBalancer.LoadBalancer == nil {
			return fmt.Errorf("Load Balancer (%s) not found", rs.Primary.Attributes["name"])
		}
		*res = *respLoadBalancer.LoadBalancer
		return nil
	}
}

func TestAccAWSLightsailLoadBalancer_disappear(t *testing.T) {
	var conf lightsail.LoadBalancer
	lightsailLoadBalancerName := fmt.Sprintf("tf-test-lightsail-%d", acctest.RandInt())

	testDestroy := func(*terraform.State) error {
		// reach out and DELETE the Load Balancer
		conn := testAccProvider.Meta().(*AWSClient).lightsailconn
		_, err := conn.DeleteLoadBalancer(&lightsail.DeleteLoadBalancerInput{
			LoadBalancerName: aws.String(lightsailLoadBalancerName),
		})

		if err != nil {
			return fmt.Errorf("error deleting Lightsail Load Balancer in disappear test")
		}

		// sleep 7 seconds to give it time, so we don't have to poll
		time.Sleep(7 * time.Second)

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSLightsail(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLightsailLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailLoadBalancerConfig_basic(lightsailLoadBalancerName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailLoadBalancerExists("aws_lightsail_load_balancer.lightsail_load_balancer_test", &conf),
					testDestroy,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSLightsailLoadBalancerDestroy(s *terraform.State) error {

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lightsail_load_balancer" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).lightsailconn

		resp, err := conn.GetLoadBalancer(&lightsail.GetLoadBalancerInput{
			LoadBalancerName: aws.String(rs.Primary.ID),
		})

		if err == nil {
			if resp.LoadBalancer != nil {
				return fmt.Errorf("Lightsail Load Balancer %q still exists", rs.Primary.ID)
			}
		}

		// Verify the error
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "NotFoundException" {
				return nil
			}
		}
		return err
	}

	return nil
}

func testAccAWSLightsailLoadBalancerConfig_basic(lightsailLoadBalancerName string) string {
	return fmt.Sprintf(`
provider "aws" {
  region = "us-east-1"
}

resource "aws_lightsail_load_balancer" "lightsail_load_balancer_test" {
  name               = "%s"
  health_check_path  = "/"
  instance_port      = "80"
}
`, lightsailLoadBalancerName)
}

func testAccAWSLightsailLoadBalancerConfig_tags1(lightsailLoadBalancerName string) string {
	return fmt.Sprintf(`
provider "aws" {
  region = "us-east-1"
}

resource "aws_lightsail_load_balancer" "lightsail_load_balancer_test" {
  name               = "%s"
  health_check_path  = "/"
  instance_port      = "80"
  tags = {
    Name = "tf-test"
  }
}
`, lightsailLoadBalancerName)
}

func testAccAWSLightsailLoadBalancerConfig_tags2(lightsailLoadBalancerName string) string {
	return fmt.Sprintf(`
provider "aws" {
  region = "us-east-1"
}

resource "aws_lightsail_load_balancer" "lightsail_load_balancer_test" {
  name               = "%s"
  health_check_path  = "/"
  instance_port      = "80"
  tags = {
    Name      = "tf-test",
    ExtraName = "tf-test"
  }
}
`, lightsailLoadBalancerName)
}
