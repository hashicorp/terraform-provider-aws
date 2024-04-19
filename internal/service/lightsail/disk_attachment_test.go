// Copyright (c) HashiCorp, Inc.
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
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tflightsail "github.com/hashicorp/terraform-provider-aws/internal/service/lightsail"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLightsailDiskAttachment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lightsail_disk_attachment.test"
	dName := sdkacctest.RandomWithPrefix("tf-acc-test")
	liName := sdkacctest.RandomWithPrefix("tf-acc-test")
	diskPath := "/dev/xvdf"
	diskPathBad := "/jenkins-home"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDiskAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDiskAttachmentConfig_basic(dName, liName, diskPath),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiskAttachmentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "disk_name", dName),
					resource.TestCheckResourceAttr(resourceName, "disk_path", diskPath),
					resource.TestCheckResourceAttr(resourceName, "instance_name", liName),
				),
			},
			{
				Config:      testAccDiskAttachmentConfig_basic(dName, liName, diskPathBad),
				ExpectError: regexache.MustCompile(`The disk path is invalid. You must specify a valid disk path.`),
			},
		},
	})
}

func TestAccLightsailDiskAttachment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lightsail_disk_attachment.test"
	dName := sdkacctest.RandomWithPrefix("tf-acc-test")
	liName := sdkacctest.RandomWithPrefix("tf-acc-test")
	diskPath := "/dev/xvdf"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDiskAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDiskAttachmentConfig_basic(dName, liName, diskPath),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiskAttachmentExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflightsail.ResourceDiskAttachment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDiskAttachmentExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No LightsailDiskAttachment ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailClient(ctx)

		out, err := tflightsail.FindDiskAttachmentById(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if out == nil {
			return fmt.Errorf("Disk Attachment %q does not exist", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDiskAttachmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lightsail_disk_attachment" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailClient(ctx)

			_, err := tflightsail.FindDiskAttachmentById(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.Lightsail, create.ErrActionCheckingDestroyed, tflightsail.ResDiskAttachment, rs.Primary.ID, errors.New("still exists"))
		}

		return nil
	}
}

func testAccDiskAttachmentConfig_basic(dName string, liName string, diskPath string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}
resource "aws_lightsail_disk" "test" {
  name              = %[1]q
  size_in_gb        = 8
  availability_zone = data.aws_availability_zones.available.names[0]
}

resource "aws_lightsail_instance" "test" {
  name              = %[2]q
  availability_zone = data.aws_availability_zones.available.names[0]
  blueprint_id      = "amazon_linux_2"
  bundle_id         = "nano_3_0"
}

resource "aws_lightsail_disk_attachment" "test" {
  disk_name     = aws_lightsail_disk.test.name
  instance_name = aws_lightsail_instance.test.name
  disk_path     = %[3]q
}
`, dName, liName, diskPath)
}
