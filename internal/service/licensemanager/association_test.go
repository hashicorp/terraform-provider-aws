// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package licensemanager_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflicensemanager "github.com/hashicorp/terraform-provider-aws/internal/service/licensemanager"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLicenseManagerAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_licensemanager_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LicenseManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "license_configuration_arn", "aws_licensemanager_license_configuration.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, "aws_instance.test", names.AttrARN),
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

func TestAccLicenseManagerAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_licensemanager_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LicenseManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssociationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflicensemanager.ResourceAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAssociationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No License Manager Association ID is set")
		}

		resourceARN, licenseConfigurationARN, err := tflicensemanager.AssociationParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LicenseManagerConn(ctx)

		return tflicensemanager.FindAssociation(ctx, conn, resourceARN, licenseConfigurationARN)
	}
}

func testAccCheckAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LicenseManagerConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_licensemanager_association" {
				continue
			}

			resourceARN, licenseConfigurationARN, err := tflicensemanager.AssociationParseResourceID(rs.Primary.ID)

			if err != nil {
				return err
			}

			err = tflicensemanager.FindAssociation(ctx, conn, resourceARN, licenseConfigurationARN)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("License Manager Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAssociationConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"

  tags = {
    Name = %[1]q
  }
}

resource "aws_licensemanager_license_configuration" "test" {
  name                  = %[1]q
  license_counting_type = "vCPU"
}

resource "aws_licensemanager_association" "test" {
  license_configuration_arn = aws_licensemanager_license_configuration.test.id
  resource_arn              = aws_instance.test.arn
}
`, rName))
}
