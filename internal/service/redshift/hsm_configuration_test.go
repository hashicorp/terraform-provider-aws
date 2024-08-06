// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfredshift "github.com/hashicorp/terraform-provider-aws/internal/service/redshift"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRedshiftHSMConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshift_hsm_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHSMConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccHSMConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHSMConfigurationExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "redshift", regexache.MustCompile(`hsmconfiguration:.+`)),
					resource.TestCheckResourceAttr(resourceName, "hsm_configuration_identifier", rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"hsm_partition_password", "hsm_server_public_certificate"},
			},
		},
	})
}

func TestAccRedshiftHSMConfiguration_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshift_hsm_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHSMConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccHSMConfigurationConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHSMConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"hsm_partition_password", "hsm_server_public_certificate"},
			},
			{
				Config: testAccHSMConfigurationConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHSMConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			}, {
				Config: testAccHSMConfigurationConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHSMConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccRedshiftHSMConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshift_hsm_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHSMConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccHSMConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHSMConfigurationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfredshift.ResourceHSMConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckHSMConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_redshift_hsm_configuration" {
				continue
			}

			_, err := tfredshift.FindHSMConfigurationByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Redshift Hsm Configuration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckHSMConfigurationExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Redshift Hsm Configuration is not set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn(ctx)

		_, err := tfredshift.FindHSMConfigurationByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccHSMConfigurationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_redshift_hsm_configuration" "test" {
  description                   = %[1]q
  hsm_configuration_identifier  = %[1]q
  hsm_ip_address                = "10.0.0.1"
  hsm_partition_name            = "aws"
  hsm_partition_password        = %[1]q
  hsm_server_public_certificate = %[1]q
}
`, rName)
}

func testAccHSMConfigurationConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_redshift_hsm_configuration" "test" {
  description                   = %[1]q
  hsm_configuration_identifier  = %[1]q
  hsm_ip_address                = "10.0.0.1"
  hsm_partition_name            = "aws"
  hsm_partition_password        = %[1]q
  hsm_server_public_certificate = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccHSMConfigurationConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_redshift_hsm_configuration" "test" {
  description                   = %[1]q
  hsm_configuration_identifier  = %[1]q
  hsm_ip_address                = "10.0.0.1"
  hsm_partition_name            = "aws"
  hsm_partition_password        = %[1]q
  hsm_server_public_certificate = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
