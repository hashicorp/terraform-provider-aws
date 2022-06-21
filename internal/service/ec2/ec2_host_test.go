package ec2_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccEC2Host_basic(t *testing.T) {
	var host ec2.Host
	resourceName := "aws_ec2_host.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHostConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostExists(resourceName, &host),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`dedicated-host/.+`)),
					resource.TestCheckResourceAttr(resourceName, "auto_placement", "on"),
					resource.TestCheckResourceAttr(resourceName, "host_recovery", "off"),
					resource.TestCheckResourceAttr(resourceName, "instance_family", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "a1.large"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
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

func TestAccEC2Host_disappears(t *testing.T) {
	var host ec2.Host
	resourceName := "aws_ec2_host.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHostConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostExists(resourceName, &host),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceHost(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2Host_instanceFamily(t *testing.T) {
	var host ec2.Host
	resourceName := "aws_ec2_host.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHostConfig_instanceFamily(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostExists(resourceName, &host),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`dedicated-host/.+`)),
					resource.TestCheckResourceAttr(resourceName, "auto_placement", "off"),
					resource.TestCheckResourceAttr(resourceName, "host_recovery", "on"),
					resource.TestCheckResourceAttr(resourceName, "instance_family", "c5"),
					resource.TestCheckResourceAttr(resourceName, "instance_type", ""),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
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
				Config: testAccHostConfig_instanceType(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostExists(resourceName, &host),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`dedicated-host/.+`)),
					resource.TestCheckResourceAttr(resourceName, "auto_placement", "on"),
					resource.TestCheckResourceAttr(resourceName, "host_recovery", "off"),
					resource.TestCheckResourceAttr(resourceName, "instance_family", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "c5.xlarge"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
		},
	})
}

func TestAccEC2Host_tags(t *testing.T) {
	var host ec2.Host
	resourceName := "aws_ec2_host.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHostConfig_tags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostExists(resourceName, &host),
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
				Config: testAccHostConfig_tags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostExists(resourceName, &host),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccHostConfig_tags1("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostExists(resourceName, &host),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccEC2Host_outpost(t *testing.T) {
	var host ec2.Host
	resourceName := "aws_ec2_host.test"
	outpostDataSourceName := "data.aws_outposts_outpost.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHostConfig_outpost(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostExists(resourceName, &host),
					resource.TestCheckResourceAttrPair(resourceName, "outpost_arn", outpostDataSourceName, "arn"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckHostExists(n string, v *ec2.Host) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Host ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		output, err := tfec2.FindHostByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckHostDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_host" {
			continue
		}

		_, err := tfec2.FindHostByID(conn, rs.Primary.ID)

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

func testAccHostConfig_basic() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), `
resource "aws_ec2_host" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  instance_type     = "a1.large"
}
`)
}

func testAccHostConfig_instanceFamily(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
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

func testAccHostConfig_instanceType(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
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

func testAccHostConfig_tags1(tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_ec2_host" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  instance_type     = "a1.large"

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccHostConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
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

func testAccHostConfig_outpost(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

resource "aws_ec2_host" "test" {
  instance_family   = "r5d"
  availability_zone = data.aws_availability_zones.available.names[1]
  outpost_arn       = data.aws_outposts_outpost.test.arn

  tags = {
    Name = %[1]q
  }
}
`, rName))
}
