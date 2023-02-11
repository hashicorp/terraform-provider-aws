package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccIPAMPool_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var pool ec2.IpamPool
	resourceName := "aws_vpc_ipam_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMPoolConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPAMPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckNoResourceAttr(resourceName, "allocation_default_netmask_length"),
					resource.TestCheckNoResourceAttr(resourceName, "allocation_max_netmask_length"),
					resource.TestCheckNoResourceAttr(resourceName, "allocation_min_netmask_length"),
					resource.TestCheckResourceAttr(resourceName, "allocation_resource_tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "auto_import", "false"),
					resource.TestCheckResourceAttr(resourceName, "aws_service", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "ipam_scope_type"),
					resource.TestCheckResourceAttr(resourceName, "locale", "None"),
					resource.TestCheckResourceAttrSet(resourceName, "pool_depth"),
					resource.TestCheckResourceAttr(resourceName, "state", "create-complete"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccIPAMPoolConfig_updated,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPAMPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "allocation_default_netmask_length", "32"),
					resource.TestCheckResourceAttr(resourceName, "allocation_max_netmask_length", "32"),
					resource.TestCheckResourceAttr(resourceName, "allocation_min_netmask_length", "32"),
					resource.TestCheckResourceAttr(resourceName, "allocation_resource_tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "allocation_resource_tags.test", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "auto_import", "true"),
					resource.TestCheckResourceAttr(resourceName, "aws_service", ""),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttrSet(resourceName, "ipam_scope_type"),
					resource.TestCheckResourceAttr(resourceName, "locale", "None"),
					resource.TestCheckResourceAttrSet(resourceName, "pool_depth"),
					resource.TestCheckResourceAttr(resourceName, "state", "modify-complete"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccIPAMPool_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var pool ec2.IpamPool
	resourceName := "aws_vpc_ipam_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMPoolConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMPoolExists(ctx, resourceName, &pool),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceIPAMPool(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIPAMPool_ipv6Basic(t *testing.T) {
	ctx := acctest.Context(t)
	var pool ec2.IpamPool
	resourceName := "aws_vpc_ipam_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMPoolConfig_ipv6,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv6"),
					resource.TestCheckResourceAttr(resourceName, "publicly_advertisable", "false"),
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

func TestAccIPAMPool_ipv6Contiguous(t *testing.T) {
	ctx := acctest.Context(t)
	var pool ec2.IpamPool
	resourceName := "aws_vpc_ipam_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMPoolConfig_ipv6Contiguous,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv6"),
					resource.TestCheckResourceAttr(resourceName, "public_ip_source", "byoip"),
					resource.TestCheckResourceAttr(resourceName, "publicly_advertisable", "false"),
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

func TestAccIPAMPool_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var pool ec2.IpamPool
	resourceName := "aws_vpc_ipam_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMPoolConfig_tags("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMPoolExists(ctx, resourceName, &pool),
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
				Config: testAccIPAMPoolConfig_tags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccIPAMPoolConfig_tags("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckIPAMPoolExists(ctx context.Context, n string, v *ec2.IpamPool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IPAM Pool ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn()

		output, err := tfec2.FindIPAMPoolByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckIPAMPoolDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpc_ipam_pool" {
				continue
			}

			_, err := tfec2.FindIPAMPoolByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IPAM Pool still exists: %s", rs.Primary.ID)
		}

		return nil
	}
}

const testAccIPAMPoolConfig_base = `
data "aws_region" "current" {}

resource "aws_vpc_ipam" "test" {
  operating_regions {
    region_name = data.aws_region.current.name
  }
}
`

var testAccIPAMPoolConfig_basic = acctest.ConfigCompose(testAccIPAMPoolConfig_base, `
resource "aws_vpc_ipam_pool" "test" {
  address_family = "ipv4"
  ipam_scope_id  = aws_vpc_ipam.test.private_default_scope_id
}
`)

var testAccIPAMPoolConfig_updated = acctest.ConfigCompose(testAccIPAMPoolConfig_base, `
resource "aws_vpc_ipam_pool" "test" {
  address_family                    = "ipv4"
  ipam_scope_id                     = aws_vpc_ipam.test.private_default_scope_id
  auto_import                       = true
  allocation_default_netmask_length = 32
  allocation_max_netmask_length     = 32
  allocation_min_netmask_length     = 32
  allocation_resource_tags = {
    test = "1"
  }
  description = "test"
}
`)

var testAccIPAMPoolConfig_ipv6 = acctest.ConfigCompose(testAccIPAMPoolConfig_base, `
resource "aws_vpc_ipam_pool" "test" {
  address_family        = "ipv6"
  ipam_scope_id         = aws_vpc_ipam.test.public_default_scope_id
  locale                = data.aws_region.current.name
  publicly_advertisable = false
}
`)

var testAccIPAMPoolConfig_ipv6Contiguous = acctest.ConfigCompose(testAccIPAMPoolConfig_base, `
resource "aws_vpc_ipam_pool" "test" {
  address_family        = "ipv6"
  ipam_scope_id         = aws_vpc_ipam.test.public_default_scope_id
  locale                = data.aws_region.current.name
  public_ip_source      = "byoip"
  aws_service           = "ec2"
  publicly_advertisable = false
}
`)

func testAccIPAMPoolConfig_tags(tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccIPAMPoolConfig_base, fmt.Sprintf(`
resource "aws_vpc_ipam_pool" "test" {
  address_family = "ipv4"
  ipam_scope_id  = aws_vpc_ipam.test.private_default_scope_id

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccIPAMPoolConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccIPAMPoolConfig_base, fmt.Sprintf(`
resource "aws_vpc_ipam_pool" "test" {
  address_family = "ipv4"
  ipam_scope_id  = aws_vpc_ipam.test.private_default_scope_id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}
