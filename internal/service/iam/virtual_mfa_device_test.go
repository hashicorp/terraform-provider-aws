// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMVirtualMFADevice_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.VirtualMFADevice
	resourceName := "aws_iam_virtual_mfa_device.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVirtualMFADeviceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualMFADeviceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVirtualMFADeviceExists(ctx, resourceName, &conf),
					acctest.CheckResourceAttrGlobalARN(resourceName, names.AttrARN, "iam", fmt.Sprintf("mfa/%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "base_32_string_seed"),
					resource.TestCheckNoResourceAttr(resourceName, "enable_date"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPath, "/"),
					resource.TestCheckResourceAttrSet(resourceName, "qr_code_png"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrUserName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"base_32_string_seed",
					"qr_code_png",
				},
			},
		},
	})
}

func TestAccIAMVirtualMFADevice_path(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.VirtualMFADevice
	resourceName := "aws_iam_virtual_mfa_device.test"

	path := "/path/"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVirtualMFADeviceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualMFADeviceConfig_path(rName, path),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVirtualMFADeviceExists(ctx, resourceName, &conf),
					acctest.CheckResourceAttrGlobalARN(resourceName, names.AttrARN, "iam", fmt.Sprintf("mfa%s%s", path, rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrPath, path),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"base_32_string_seed",
					"qr_code_png",
				},
			},
		},
	})
}

func TestAccIAMVirtualMFADevice_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.VirtualMFADevice
	resourceName := "aws_iam_virtual_mfa_device.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVirtualMFADeviceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualMFADeviceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVirtualMFADeviceExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfiam.ResourceVirtualMFADevice(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVirtualMFADeviceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iam_virtual_mfa_device" {
				continue
			}

			output, err := tfiam.FindVirtualMFADeviceBySerialNumber(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if output != nil {
				return fmt.Errorf("IAM Virtual MFA Device (%s) still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckVirtualMFADeviceExists(ctx context.Context, n string, v *awstypes.VirtualMFADevice) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Virtual MFA Device ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		output, err := tfiam.FindVirtualMFADeviceBySerialNumber(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccVirtualMFADeviceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_virtual_mfa_device" "test" {
  virtual_mfa_device_name = %[1]q
}
`, rName)
}

func testAccVirtualMFADeviceConfig_path(rName, path string) string {
	return fmt.Sprintf(`
resource "aws_iam_virtual_mfa_device" "test" {
  virtual_mfa_device_name = %[1]q

  path = %[2]q
}
`, rName, path)
}
