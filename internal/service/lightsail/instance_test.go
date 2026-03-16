// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lightsail_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/envvar"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tflightsail "github.com/hashicorp/terraform-provider-aws/internal/service/lightsail"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	availabilityZoneKey = "TF_AWS_LIGHTSAIL_AVAILABILITY_ZONE"
)

const (
	envVarAvailabilityZoneKeyError = "The availability zone that is outside the providers current region."
)

func TestAccLightsailInstance_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lightsail_instance.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttrSet(resourceName, "blueprint_id"),
					resource.TestCheckResourceAttrSet(resourceName, "bundle_id"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_addresses.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "ipv6_addresses.0", regexache.MustCompile(`([0-9a-f]{1,4}:){7}[0-9a-f]{1,4}`)),
					resource.TestCheckResourceAttrSet(resourceName, "key_pair_name"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestMatchResourceAttr(resourceName, "ram_size", regexache.MustCompile(`\d+(.\d+)?`)),
				),
			},
		},
	})
}

func TestAccLightsailInstance_name(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lightsail_instance.test"
	rNameWithSpaces := fmt.Sprint(rName, "string with spaces")
	rNameWithStartingDigit := fmt.Sprintf("01-%s", rName)
	rNameWithStartingHyphen := fmt.Sprintf("-%s", rName)
	rNameWithUnderscore := fmt.Sprintf("%s_123456", rName)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccInstanceConfig_basic(rNameWithSpaces),
				ExpectError: regexache.MustCompile(`must contain only alphanumeric characters, underscores, hyphens, and dots`),
			},
			{
				Config:      testAccInstanceConfig_basic(rNameWithStartingHyphen),
				ExpectError: regexache.MustCompile(`must begin with an alphanumeric character`),
			},
			{
				Config: testAccInstanceConfig_basic(rNameWithStartingDigit),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttrSet(resourceName, "blueprint_id"),
					resource.TestCheckResourceAttrSet(resourceName, "bundle_id"),
					resource.TestCheckResourceAttrSet(resourceName, "key_pair_name"),
				),
			},
			{
				Config: testAccInstanceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttrSet(resourceName, "blueprint_id"),
					resource.TestCheckResourceAttrSet(resourceName, "bundle_id"),
					resource.TestCheckResourceAttrSet(resourceName, "key_pair_name"),
				),
			},
			{
				Config: testAccInstanceConfig_basic(rNameWithUnderscore),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAvailabilityZone),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lightsail_instance.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttrSet(resourceName, "blueprint_id"),
					resource.TestCheckResourceAttrSet(resourceName, "bundle_id"),
					resource.TestCheckResourceAttrSet(resourceName, "key_pair_name"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInstanceConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttrSet(resourceName, "blueprint_id"),
					resource.TestCheckResourceAttrSet(resourceName, "bundle_id"),
					resource.TestCheckResourceAttrSet(resourceName, "key_pair_name"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccInstanceConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttrSet(resourceName, "blueprint_id"),
					resource.TestCheckResourceAttrSet(resourceName, "bundle_id"),
					resource.TestCheckResourceAttrSet(resourceName, "key_pair_name"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccLightsailInstance_keyOnlyTags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lightsail_instance.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_tags1(rName, acctest.CtKey1, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttrSet(resourceName, "blueprint_id"),
					resource.TestCheckResourceAttrSet(resourceName, "bundle_id"),
					resource.TestCheckResourceAttrSet(resourceName, "key_pair_name"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInstanceConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttrSet(resourceName, "blueprint_id"),
					resource.TestCheckResourceAttrSet(resourceName, "bundle_id"),
					resource.TestCheckResourceAttrSet(resourceName, "key_pair_name"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, ""),
				),
			},
			{
				Config: testAccInstanceConfig_tags1(rName, acctest.CtKey2, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttrSet(resourceName, "blueprint_id"),
					resource.TestCheckResourceAttrSet(resourceName, "bundle_id"),
					resource.TestCheckResourceAttrSet(resourceName, "key_pair_name"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, ""),
				),
			},
		},
	})
}

func TestAccLightsailInstance_IPAddressType(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lightsail_instance.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_IPAddressType(rName, "ipv4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddressType, "ipv4"),
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
					testAccCheckInstanceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddressType, "dualstack"),
				),
			},
		},
	})
}

func TestAccLightsailInstance_addOn(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lightsail_instance.test"
	statusEnabled := "Enabled"
	statusDisabled := "Disabled"
	snapshotTime1 := "06:00"
	snapshotTime2 := "10:00"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_addOn(rName, snapshotTime1, statusEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "add_on.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "add_on.*", map[string]string{
						names.AttrType: "AutoSnapshot",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "add_on.*", map[string]string{
						"snapshot_time": snapshotTime1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "add_on.*", map[string]string{
						names.AttrStatus: statusEnabled,
					}),
				),
			},
			{
				Config: testAccInstanceConfig_addOn(rName, snapshotTime2, statusEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "add_on.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "add_on.*", map[string]string{
						names.AttrType: "AutoSnapshot",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "add_on.*", map[string]string{
						"snapshot_time": snapshotTime2,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "add_on.*", map[string]string{
						names.AttrStatus: statusEnabled,
					}),
				),
			},
			{
				Config: testAccInstanceConfig_addOn(rName, snapshotTime2, statusDisabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "add_on.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "add_on.*", map[string]string{
						names.AttrType: "AutoSnapshot",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "add_on.*", map[string]string{
						"snapshot_time": snapshotTime2,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "add_on.*", map[string]string{
						names.AttrStatus: statusDisabled,
					}),
				),
			},
		},
	})
}

func TestAccLightsailInstance_availabilityZone(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	// This test is expecting a region to be set in an environment variable that it outside the current provider region
	availabilityZone := envvar.SkipIfEmpty(t, availabilityZoneKey, envVarAvailabilityZoneKeyError)
	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccInstanceConfig_availabilityZone(rName, availabilityZone),
				ExpectError: regexache.MustCompile(`availability_zone must be within the same region as provider region.`),
			},
		},
	})
}

func TestAccLightsailInstance_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lightsail_instance.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tflightsail.ResourceInstance(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckInstanceExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No LightsailInstance ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).LightsailClient(ctx)

		out, err := tflightsail.FindInstanceById(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if out == nil {
			return fmt.Errorf("Instance (%s) not found", rs.Primary.Attributes[names.AttrName])
		}

		return nil
	}
}

func testAccCheckInstanceDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lightsail_instance" {
				continue
			}

			conn := acctest.ProviderMeta(ctx, t).LightsailClient(ctx)

			_, err := tflightsail.FindInstanceById(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
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
	conn := acctest.ProviderMeta(ctx, t).LightsailClient(ctx)

	input := &lightsail.GetInstancesInput{}

	_, err := conn.GetInstances(ctx, input)

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
  name              = %[1]q
  availability_zone = data.aws_availability_zones.available.names[0]
  blueprint_id      = "amazon_linux_2"
  bundle_id         = "nano_3_0"
}
`, rName))
}

func testAccInstanceConfig_availabilityZone(rName, availabilityZone string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfigBase(),
		fmt.Sprintf(`	
resource "aws_lightsail_instance" "test" {
  name              = %[1]q
  availability_zone = %[2]q
  blueprint_id      = "amazon_linux_2"
  bundle_id         = "nano_3_0"
}
`, rName, availabilityZone))
}

func testAccInstanceConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfigBase(),
		fmt.Sprintf(`
resource "aws_lightsail_instance" "test" {
  name              = %[1]q
  availability_zone = data.aws_availability_zones.available.names[0]
  blueprint_id      = "amazon_linux_2"
  bundle_id         = "nano_3_0"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccInstanceConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfigBase(),
		fmt.Sprintf(`
resource "aws_lightsail_instance" "test" {
  name              = %[1]q
  availability_zone = data.aws_availability_zones.available.names[0]
  blueprint_id      = "amazon_linux_2"
  bundle_id         = "nano_3_0"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccInstanceConfig_IPAddressType(rName, rIPAddressType string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfigBase(),
		fmt.Sprintf(`
resource "aws_lightsail_instance" "test" {
  name              = %[1]q
  availability_zone = data.aws_availability_zones.available.names[0]
  blueprint_id      = "amazon_linux_2"
  bundle_id         = "nano_3_0"
  ip_address_type   = %[2]q
}
`, rName, rIPAddressType))
}

func testAccInstanceConfig_addOn(rName, snapshotTime, status string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfigBase(),
		fmt.Sprintf(`
resource "aws_lightsail_instance" "test" {
  name              = %[1]q
  availability_zone = data.aws_availability_zones.available.names[0]
  blueprint_id      = "amazon_linux_2"
  bundle_id         = "nano_3_0"
  add_on {
    type          = "AutoSnapshot"
    snapshot_time = %[2]q
    status        = %[3]q
  }
}
`, rName, snapshotTime, status))
}
