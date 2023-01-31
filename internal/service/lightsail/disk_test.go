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

func TestAccLightsailDisk_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var disk lightsail.Disk
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_disk.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDiskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDiskConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiskExists(ctx, resourceName, &disk),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "lightsail", regexp.MustCompile(`Disk/.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "size_in_gb", "8"),
					resource.TestCheckResourceAttrSet(resourceName, "support_code"),
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

func TestAccLightsailDisk_Tags(t *testing.T) {
	ctx := acctest.Context(t)
	var disk1, disk2, disk3 lightsail.Disk
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_disk.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDiskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDiskConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiskExists(ctx, resourceName, &disk1),
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
				Config: testAccDiskConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiskExists(ctx, resourceName, &disk2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccDiskConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiskExists(ctx, resourceName, &disk3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckDiskExists(ctx context.Context, n string, disk *lightsail.Disk) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No LightsailDisk ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn()

		resp, err := tflightsail.FindDiskById(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if resp == nil {
			return fmt.Errorf("Disk %q does not exist", rs.Primary.ID)
		}

		*disk = *resp

		return nil
	}
}

func TestAccLightsailDisk_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var disk lightsail.Disk
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_disk.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDiskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDiskConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiskExists(ctx, resourceName, &disk),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflightsail.ResourceDisk(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDiskDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lightsail_disk" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn()

			_, err := tflightsail.FindDiskById(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.Lightsail, create.ErrActionCheckingDestroyed, tflightsail.ResDisk, rs.Primary.ID, errors.New("still exists"))
		}

		return nil
	}
}

func testAccDiskConfigBase() string {
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

func testAccDiskConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccDiskConfigBase(),
		fmt.Sprintf(`
resource "aws_lightsail_disk" "test" {
  name              = %[1]q
  size_in_gb        = 8
  availability_zone = data.aws_availability_zones.available.names[0]
}
`, rName))
}

func testAccDiskConfig_tags1(rName string, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccDiskConfigBase(),
		fmt.Sprintf(`
resource "aws_lightsail_disk" "test" {
  name              = %[1]q
  size_in_gb        = 8
  availability_zone = data.aws_availability_zones.available.names[0]
  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccDiskConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccDiskConfigBase(),
		fmt.Sprintf(`
resource "aws_lightsail_disk" "test" {
  name              = %[1]q
  size_in_gb        = 8
  availability_zone = data.aws_availability_zones.available.names[0]
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
