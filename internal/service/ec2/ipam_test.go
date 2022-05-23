package ec2_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func testAccIPAMPreCheck(t *testing.T) {
	acctest.PreCheckIAMServiceLinkedRole(t, "/aws-service-role/ipam.amazonaws.com")
}

func TestAccIPAM_basic(t *testing.T) {
	resourceName := "aws_vpc_ipam.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccIPAMPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckIPAMDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMConfig_base,
				Check: resource.ComposeTestCheckFunc(
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "ec2", regexp.MustCompile(`ipam/ipam-[\da-f]+$`)),
					resource.TestCheckResourceAttr(resourceName, "operating_regions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scope_count", "2"),
					resource.TestMatchResourceAttr(resourceName, "private_default_scope_id", regexp.MustCompile(`^ipam-scope-[\da-f]+`)),
					resource.TestMatchResourceAttr(resourceName, "public_default_scope_id", regexp.MustCompile(`^ipam-scope-[\da-f]+`)),
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

func TestAccIPAM_modify(t *testing.T) {
	var providers []*schema.Provider
	resourceName := "aws_vpc_ipam.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccIPAMPreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.FactoriesMultipleRegion(&providers, 2),
		CheckDestroy:      testAccCheckIPAMDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMConfig_base,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccIPAMConfig_operatingRegion(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
				),
			},
			{
				Config: testAccIPAMConfig_base,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
				),
			},
			{
				Config: testAccIPAMConfig_baseAlternateDescription,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "test ipam"),
				),
			},
		},
	})
}

func TestAccIPAM_cascade(t *testing.T) {
	resourceName := "aws_vpc_ipam.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccIPAMPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckIPAMDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMConfig_cascade(),
				Check: resource.ComposeTestCheckFunc(
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "ec2", regexp.MustCompile(`ipam/ipam-[\da-f]+$`)),
					resource.TestCheckResourceAttr(resourceName, "operating_regions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scope_count", "2"),
					resource.TestMatchResourceAttr(resourceName, "private_default_scope_id", regexp.MustCompile(`^ipam-scope-[\da-f]+`)),
					resource.TestMatchResourceAttr(resourceName, "public_default_scope_id", regexp.MustCompile(`^ipam-scope-[\da-f]+`)),
					testAccCheckIPAMScopeCreate(resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"cascade"},
			},
		},
	})
}

func TestAccIPAM_tags(t *testing.T) {
	resourceName := "aws_vpc_ipam.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccIPAMPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckIPAMDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMConfig_tags("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
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
				Config: testAccIPAMConfig_tags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccIPAMConfig_tags("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckIPAMDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc_ipam" {
			continue
		}

		id := aws.String(rs.Primary.ID)

		if _, err := tfec2.WaiterIPAMDeleted(conn, *id, tfec2.IPAMDeleteTimeout); err != nil {
			if tfresource.NotFound(err) {
				return nil
			}
			return fmt.Errorf("error waiting for IPAM to be deleted: %w", err)
		}
	}

	return nil
}

func testAccCheckIPAMScopeCreate(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpc_ipam" {
				continue
			}
			conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
			input := &ec2.CreateIpamScopeInput{
				ClientToken: aws.String(resource.UniqueId()),
				IpamId:      aws.String(rs.Primary.ID),
			}
			_, err := conn.CreateIpamScope(input)
			return err
		}
		return fmt.Errorf("could not create VPC IPAM Scope")
	}
}

const testAccIPAMConfig_base = `
data "aws_region" "current" {}

resource "aws_vpc_ipam" "test" {
  description = "test"
  operating_regions {
    region_name = data.aws_region.current.name
  }
}
`

func testAccIPAMConfig_cascade() string {
	return `
data "aws_region" "current" {}

resource "aws_vpc_ipam" "test" {
  description = "test"
  operating_regions {
    region_name = data.aws_region.current.name
  }
  cascade = true
}
	`
}

const testAccIPAMConfig_baseAlternateDescription = `
data "aws_region" "current" {}

resource "aws_vpc_ipam" "test" {
  description = "test ipam"
  operating_regions {
    region_name = data.aws_region.current.name
  }
}
`

func testAccIPAMConfig_operatingRegion() string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2), `
data "aws_region" "current" {}

data "aws_region" "alternate" {
  provider = awsalternate
}


resource "aws_vpc_ipam" "test" {
  description = "test"
  operating_regions {
    region_name = data.aws_region.current.name
  }
  operating_regions {
    region_name = data.aws_region.alternate.name
  }
}
`)
}

func testAccIPAMConfig_tags(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc_ipam" "test" {
  description = "test"
  operating_regions {
    region_name = data.aws_region.current.name
  }
  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccIPAMConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc_ipam" "test" {
  description = "test"
  operating_regions {
    region_name = data.aws_region.current.name
  }
  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
	`, tagKey1, tagValue1, tagKey2, tagValue2)
}
