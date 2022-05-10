package lightsail_test

import (
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccLightsailInstance_basic(t *testing.T) {
	var conf lightsail.Instance
	lightsailName := fmt.Sprintf("tf-test-lightsail-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_basic(lightsailName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists("aws_lightsail_instance.lightsail_instance_test", &conf),
					resource.TestCheckResourceAttrSet("aws_lightsail_instance.lightsail_instance_test", "availability_zone"),
					resource.TestCheckResourceAttrSet("aws_lightsail_instance.lightsail_instance_test", "blueprint_id"),
					resource.TestCheckResourceAttrSet("aws_lightsail_instance.lightsail_instance_test", "bundle_id"),
					resource.TestMatchResourceAttr("aws_lightsail_instance.lightsail_instance_test", "ipv6_address", regexp.MustCompile(`([a-f0-9]{1,4}:){7}[a-f0-9]{1,4}`)),
					resource.TestCheckResourceAttr("aws_lightsail_instance.lightsail_instance_test", "ipv6_addresses.#", "1"),
					resource.TestCheckResourceAttrSet("aws_lightsail_instance.lightsail_instance_test", "key_pair_name"),
					resource.TestCheckResourceAttr("aws_lightsail_instance.lightsail_instance_test", "tags.%", "0"),
					resource.TestMatchResourceAttr("aws_lightsail_instance.lightsail_instance_test", "ram_size", regexp.MustCompile(`\d+(.\d+)?`)),
				),
			},
		},
	})
}

func TestAccLightsailInstance_name(t *testing.T) {
	var conf lightsail.Instance
	lightsailName := fmt.Sprintf("tf-test-lightsail-%d", sdkacctest.RandInt())
	lightsailNameWithSpaces := fmt.Sprint(lightsailName, "string with spaces")
	lightsailNameWithStartingDigit := fmt.Sprintf("01-%s", lightsailName)
	lightsailNameWithUnderscore := fmt.Sprintf("%s_123456", lightsailName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccInstanceConfig_basic(lightsailNameWithSpaces),
				ExpectError: regexp.MustCompile(`must contain only alphanumeric characters, underscores, hyphens, and dots`),
			},
			{
				Config:      testAccInstanceConfig_basic(lightsailNameWithStartingDigit),
				ExpectError: regexp.MustCompile(`must begin with an alphabetic character`),
			},
			{
				Config: testAccInstanceConfig_basic(lightsailName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists("aws_lightsail_instance.lightsail_instance_test", &conf),
					resource.TestCheckResourceAttrSet("aws_lightsail_instance.lightsail_instance_test", "availability_zone"),
					resource.TestCheckResourceAttrSet("aws_lightsail_instance.lightsail_instance_test", "blueprint_id"),
					resource.TestCheckResourceAttrSet("aws_lightsail_instance.lightsail_instance_test", "bundle_id"),
					resource.TestCheckResourceAttrSet("aws_lightsail_instance.lightsail_instance_test", "key_pair_name"),
				),
			},
			{
				Config: testAccInstanceConfig_basic(lightsailNameWithUnderscore),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists("aws_lightsail_instance.lightsail_instance_test", &conf),
					resource.TestCheckResourceAttrSet("aws_lightsail_instance.lightsail_instance_test", "availability_zone"),
					resource.TestCheckResourceAttrSet("aws_lightsail_instance.lightsail_instance_test", "blueprint_id"),
					resource.TestCheckResourceAttrSet("aws_lightsail_instance.lightsail_instance_test", "bundle_id"),
					resource.TestCheckResourceAttrSet("aws_lightsail_instance.lightsail_instance_test", "key_pair_name"),
				),
			},
		},
	})
}

func TestAccLightsailInstance_tags(t *testing.T) {
	var conf lightsail.Instance
	lightsailName := fmt.Sprintf("tf-test-lightsail-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_tags1(lightsailName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists("aws_lightsail_instance.lightsail_instance_test", &conf),
					resource.TestCheckResourceAttrSet("aws_lightsail_instance.lightsail_instance_test", "availability_zone"),
					resource.TestCheckResourceAttrSet("aws_lightsail_instance.lightsail_instance_test", "blueprint_id"),
					resource.TestCheckResourceAttrSet("aws_lightsail_instance.lightsail_instance_test", "bundle_id"),
					resource.TestCheckResourceAttrSet("aws_lightsail_instance.lightsail_instance_test", "key_pair_name"),
					resource.TestCheckResourceAttr("aws_lightsail_instance.lightsail_instance_test", "tags.%", "2"),
				),
			},
			{
				Config: testAccInstanceConfig_tags2(lightsailName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists("aws_lightsail_instance.lightsail_instance_test", &conf),
					resource.TestCheckResourceAttrSet("aws_lightsail_instance.lightsail_instance_test", "availability_zone"),
					resource.TestCheckResourceAttrSet("aws_lightsail_instance.lightsail_instance_test", "blueprint_id"),
					resource.TestCheckResourceAttrSet("aws_lightsail_instance.lightsail_instance_test", "bundle_id"),
					resource.TestCheckResourceAttrSet("aws_lightsail_instance.lightsail_instance_test", "key_pair_name"),
					resource.TestCheckResourceAttr("aws_lightsail_instance.lightsail_instance_test", "tags.%", "3"),
				),
			},
		},
	})
}

func TestAccLightsailInstance_disappears(t *testing.T) {
	var conf lightsail.Instance
	lightsailName := fmt.Sprintf("tf-test-lightsail-%d", sdkacctest.RandInt())

	testDestroy := func(*terraform.State) error {
		// reach out and DELETE the Instance
		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn
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
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_basic(lightsailName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("aws_lightsail_instance.lightsail_instance_test", &conf),
					testDestroy,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckInstanceExists(n string, res *lightsail.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No LightsailInstance ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn

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

func testAccCheckInstanceDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lightsail_instance" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn

		respInstance, err := conn.GetInstance(&lightsail.GetInstanceInput{
			InstanceName: aws.String(rs.Primary.Attributes["name"]),
		})

		if tfawserr.ErrCodeEquals(err, lightsail.ErrCodeNotFoundException) {
			continue
		}

		if err == nil {
			if respInstance.Instance != nil {
				return fmt.Errorf("LightsailInstance %q still exists", rs.Primary.ID)
			}
		}

		return err
	}

	return nil
}

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn

	input := &lightsail.GetInstancesInput{}

	_, err := conn.GetInstances(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccInstanceConfig_basic(lightsailName string) string {
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
}
`, lightsailName)
}

func testAccInstanceConfig_tags1(lightsailName string) string {
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
  }
}
`, lightsailName)
}

func testAccInstanceConfig_tags2(lightsailName string) string {
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
    Name       = "tf-test",
    KeyOnlyTag = ""
    ExtraName  = "tf-test"
  }
}
`, lightsailName)
}
