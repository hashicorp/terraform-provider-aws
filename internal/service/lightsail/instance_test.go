package lightsail_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/lightsail"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tflightsail "github.com/hashicorp/terraform-provider-aws/internal/service/lightsail"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLightsailInstance_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone"),
					resource.TestCheckResourceAttrSet(resourceName, "blueprint_id"),
					resource.TestCheckResourceAttrSet(resourceName, "bundle_id"),
					resource.TestMatchResourceAttr(resourceName, "ipv6_address", regexp.MustCompile(`([a-f0-9]{1,4}:){7}[a-f0-9]{1,4}`)),
					resource.TestCheckResourceAttr(resourceName, "ipv6_addresses.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "key_pair_name"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestMatchResourceAttr(resourceName, "ram_size", regexp.MustCompile(`\d+(.\d+)?`)),
				),
			},
		},
	})
}

func TestAccLightsailInstance_name(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_instance.test"
	rNameWithSpaces := fmt.Sprint(rName, "string with spaces")
	rNameWithStartingDigit := fmt.Sprintf("01-%s", rName)
	rNameWithUnderscore := fmt.Sprintf("%s_123456", rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccInstanceConfig_basic(rNameWithSpaces),
				ExpectError: regexp.MustCompile(`must contain only alphanumeric characters, underscores, hyphens, and dots`),
			},
			{
				Config:      testAccInstanceConfig_basic(rNameWithStartingDigit),
				ExpectError: regexp.MustCompile(`must begin with an alphabetic character`),
			},
			{
				Config: testAccInstanceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone"),
					resource.TestCheckResourceAttrSet(resourceName, "blueprint_id"),
					resource.TestCheckResourceAttrSet(resourceName, "bundle_id"),
					resource.TestCheckResourceAttrSet(resourceName, "key_pair_name"),
				),
			},
			{
				Config: testAccInstanceConfig_basic(rNameWithUnderscore),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone"),
					resource.TestCheckResourceAttrSet(resourceName, "blueprint_id"),
					resource.TestCheckResourceAttrSet(resourceName, "bundle_id"),
					resource.TestCheckResourceAttrSet(resourceName, "key_pair_name"),
				),
			},
		},
	})
}

func TestAccLightsailInstance_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_tags1(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone"),
					resource.TestCheckResourceAttrSet(resourceName, "blueprint_id"),
					resource.TestCheckResourceAttrSet(resourceName, "bundle_id"),
					resource.TestCheckResourceAttrSet(resourceName, "key_pair_name"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
				),
			},
			{
				Config: testAccInstanceConfig_tags2(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone"),
					resource.TestCheckResourceAttrSet(resourceName, "blueprint_id"),
					resource.TestCheckResourceAttrSet(resourceName, "bundle_id"),
					resource.TestCheckResourceAttrSet(resourceName, "key_pair_name"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
				),
			},
		},
	})
}

func TestAccLightsailInstance_IPAddressType(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_IPAddressType(rName, "ipv4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "ip_address_type", "ipv4"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInstanceConfig_IPAddressType(rName, "dualstack"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "ip_address_type", "dualstack"),
				),
			},
		},
	})
}

func TestAccLightsailInstance_addOn(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_instance.test"
	statusEnabled := "Enabled"
	statusDisabled := "Disabled"
	snapshotTime1 := "06:00"
	snapshotTime2 := "10:00"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_addOn(rName, snapshotTime1, statusEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "add_on.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "add_on.*", map[string]string{
						"type": "AutoSnapshot",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "add_on.*", map[string]string{
						"snapshot_time": snapshotTime1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "add_on.*", map[string]string{
						"status": statusEnabled,
					}),
				),
			},
			{
				Config: testAccInstanceConfig_addOn(rName, snapshotTime2, statusEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "add_on.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "add_on.*", map[string]string{
						"type": "AutoSnapshot",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "add_on.*", map[string]string{
						"snapshot_time": snapshotTime2,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "add_on.*", map[string]string{
						"status": statusEnabled,
					}),
				),
			},
			{
				Config: testAccInstanceConfig_addOn(rName, snapshotTime2, statusDisabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "add_on.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "add_on.*", map[string]string{
						"type": "AutoSnapshot",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "add_on.*", map[string]string{
						"snapshot_time": snapshotTime2,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "add_on.*", map[string]string{
						"status": statusDisabled,
					}),
				),
			},
		},
	})
}

func TestAccLightsailInstance_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflightsail.ResourceInstance(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckInstanceExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No LightsailInstance ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn()

		out, err := tflightsail.FindInstanceById(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if out == nil {
			return fmt.Errorf("Instance (%s) not found", rs.Primary.Attributes["name"])
		}

		return nil
	}
}

func testAccCheckInstanceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lightsail_instance" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn()

			_, err := tflightsail.FindInstanceById(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.Lightsail, create.ErrActionCheckingDestroyed, tflightsail.ResInstance, rs.Primary.ID, errors.New("still exists"))
		}

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn()

	input := &lightsail.GetInstancesInput{}

	_, err := conn.GetInstancesWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccInstanceConfigBase() string {
	return `
data "aws_availability_zones" "available" {
  state = "available"
  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}
`
}

func testAccInstanceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfigBase(),
		fmt.Sprintf(`	
resource "aws_lightsail_instance" "test" {
  name              = "%s"
  availability_zone = data.aws_availability_zones.available.names[0]
  blueprint_id      = "amazon_linux"
  bundle_id         = "nano_1_0"
}
`, rName))
}

func testAccInstanceConfig_tags1(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfigBase(),
		fmt.Sprintf(`
resource "aws_lightsail_instance" "test" {
  name              = "%s"
  availability_zone = data.aws_availability_zones.available.names[0]
  blueprint_id      = "amazon_linux"
  bundle_id         = "nano_1_0"

  tags = {
    Name       = "tf-test"
    KeyOnlyTag = ""
  }
}
`, rName))
}

func testAccInstanceConfig_tags2(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfigBase(),
		fmt.Sprintf(`
resource "aws_lightsail_instance" "test" {
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
`, rName))
}

func testAccInstanceConfig_IPAddressType(rName string, rIPAddressType string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfigBase(),
		fmt.Sprintf(`
resource "aws_lightsail_instance" "test" {
  name              = %[1]q
  availability_zone = data.aws_availability_zones.available.names[0]
  blueprint_id      = "amazon_linux"
  bundle_id         = "nano_1_0"
  ip_address_type   = %[2]q
}
`, rName, rIPAddressType))
}

func testAccInstanceConfig_addOn(rName string, snapshotTime string, status string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfigBase(),
		fmt.Sprintf(`
resource "aws_lightsail_instance" "test" {
  name              = %[1]q
  availability_zone = data.aws_availability_zones.available.names[0]
  blueprint_id      = "amazon_linux"
  bundle_id         = "nano_1_0"
  add_on {
    type          = "AutoSnapshot"
    snapshot_time = %[2]q
    status        = %[3]q
  }
}
`, rName, snapshotTime, status))
}
