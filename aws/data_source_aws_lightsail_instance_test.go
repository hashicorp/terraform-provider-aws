package aws

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceAWSLightsailInstance_basic(t *testing.T) {
	var conf lightsail.Instance
	lightsailName := fmt.Sprintf("tf-data-test-lightsail-%d", acctest.RandInt())
	resourceName := "aws_lightsail_instance.lightsail_instance_test"
	dataSourceName := "data.aws_lightsail_instance.lightsail_instance_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(lightsail.EndpointsID, t)
			testAccPreCheckAWSLightsail(t)
		},
		IDRefreshName: "aws_lightsail_instance.lightsail_instance_test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccDataSourceCheckAWSLightsailInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSLightsailInstanceConfig_basic(lightsailName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "availability_zone", dataSourceName, "availability_zone"),
					resource.TestCheckResourceAttrPair(resourceName, "blueprint_id", dataSourceName, "blueprint_id"),
					resource.TestCheckResourceAttrPair(resourceName, "bundle_id", dataSourceName, "bundle_id"),
					resource.TestCheckResourceAttrPair(resourceName, "tags", dataSourceName, "tags"),
					testAccDataSourceCheckAWSLightsailInstanceExists("aws_lightsail_instance.lightsail_instance_test", &conf),
				),
			},
		},
	})
}

func TestAccDataSourceAWSLightsailInstance_disapear(t *testing.T) {
	var conf lightsail.Instance
	lightsailName := fmt.Sprintf("tf-data-test-lightsail-%d", acctest.RandInt())

	testDestroy := func(*terraform.State) error {
		// reach out and DELETE the Instance
		conn := testAccProvider.Meta().(*AWSClient).lightsailconn
		_, err := conn.DeleteInstance(&lightsail.DeleteInstanceInput{
			InstanceName: aws.String(lightsailName),
		})

		if err != nil {
			return fmt.Errorf("error deleting Lightsail Instance in disappear test")
		}

		// sleep 7 seconds to give it time, so we don't have to poll
		time.Sleep(7 * time.Second)

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(lightsail.EndpointsID, t)
			testAccPreCheckAWSLightsail(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccDataSourceCheckAWSLightsailInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSLightsailInstanceConfig_basic(lightsailName),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceCheckAWSLightsailInstanceExists("aws_lightsail_instance.lightsail_instance_test", &conf),
					testDestroy,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccDataSourceCheckAWSLightsailInstanceExists(n string, res *lightsail.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No LightsailInstance ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).lightsailconn

		respInstance, err := conn.GetInstance(&lightsail.GetInstanceInput{
			InstanceName: aws.String(rs.Primary.Attributes["name"]),
		})

		if err != nil {
			return err
		}

		if respInstance == nil || respInstance.Instance == nil {
			return fmt.Errorf("Instance (%s) not found", rs.Primary.Attributes["name"])
		}
		*res = *respInstance.Instance
		return nil
	}
}

func testAccDataSourceCheckAWSLightsailInstanceDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lightsail_instance" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).lightsailconn

		respInstance, err := conn.GetInstance(&lightsail.GetInstanceInput{
			InstanceName: aws.String(rs.Primary.Attributes["name"]),
		})

		if err == nil {
			if respInstance.Instance != nil {
				return fmt.Errorf("LightsailInstance %q still exists", rs.Primary.ID)
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

func testAccDataSourceAWSLightsailInstanceConfig_basic(lightsailName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
	state = "available"

	filter {
		name   = "opt-in-status"
		values = ["opt-in-not-required"]
	}
}

resource "aws_lightsail_instance" "lightsail_instance_test" {
	name              = "%s"
	availability_zone = data.aws_availability_zones.available.names[0]
	blueprint_id      = "amazon_linux"
	bundle_id         = "nano_1_0"

	tags = {
		Name       = "tf-test"
		KeyOnlyTag = ""
		ExtraName  = "tf-test"
	}
}

data "aws_lightsail_instance" "lightsail_instance_test" {
	name = aws_lightsail_instance.lightsail_instance_test.id
}
`, lightsailName)
}
