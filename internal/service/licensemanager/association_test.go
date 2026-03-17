// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package licensemanager_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tflicensemanager "github.com/hashicorp/terraform-provider-aws/internal/service/licensemanager"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLicenseManagerAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_licensemanager_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LicenseManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssociationExists(ctx, t, resourceName),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_licensemanager_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LicenseManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssociationExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tflicensemanager.ResourceAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAssociationExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).LicenseManagerClient(ctx)

		return tflicensemanager.FindAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrResourceARN], rs.Primary.Attributes["license_configuration_arn"])
	}
}

func testAccCheckAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LicenseManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_licensemanager_association" {
				continue
			}

			err := tflicensemanager.FindAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrResourceARN], rs.Primary.Attributes["license_configuration_arn"])

			if retry.NotFound(err) {
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
