// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless_test

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
	tfredshiftserverless "github.com/hashicorp/terraform-provider-aws/internal/service/redshiftserverless"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRedshiftServerlessEndpointAccess_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshiftserverless_endpoint_access.test"
	rName := sdkacctest.RandStringFromCharSet(30, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointAccessDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointAccessConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointAccessExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "redshift-serverless", regexache.MustCompile("managedvpcendpoint/.+$")),
					resource.TestCheckResourceAttr(resourceName, "endpoint_name", rName),
					resource.TestCheckResourceAttr(resourceName, "owner_account", ""),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPort),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", "aws_subnet.test.0", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "workgroup_name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEndpointAccessConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointAccessExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "redshift-serverless", regexache.MustCompile("managedvpcendpoint/.+$")),
					resource.TestCheckResourceAttr(resourceName, "endpoint_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPort),
					resource.TestCheckResourceAttr(resourceName, "owner_account", ""),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", "aws_subnet.test.0", names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_security_group_ids.*", "aws_security_group.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "workgroup_name", rName),
				),
			},
		},
	})
}

func TestAccRedshiftServerlessEndpointAccess_Disappears_workgroup(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshiftserverless_endpoint_access.test"
	rName := sdkacctest.RandStringFromCharSet(30, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointAccessDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointAccessConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointAccessExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfredshiftserverless.ResourceWorkgroup(), "aws_redshiftserverless_workgroup.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRedshiftServerlessEndpointAccess_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshiftserverless_endpoint_access.test"
	rName := sdkacctest.RandStringFromCharSet(30, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointAccessDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointAccessConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointAccessExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfredshiftserverless.ResourceEndpointAccess(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckEndpointAccessDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftServerlessConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_redshiftserverless_endpoint_access" {
				continue
			}
			_, err := tfredshiftserverless.FindEndpointAccessByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Redshift Serverless Endpoint Access %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckEndpointAccessExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftServerlessConn(ctx)

		_, err := tfredshiftserverless.FindEndpointAccessByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccEndpointAccessConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_redshiftserverless_namespace" "test" {
  namespace_name = %[1]q
}

resource "aws_redshiftserverless_workgroup" "test" {
  namespace_name = aws_redshiftserverless_namespace.test.namespace_name
  workgroup_name = %[1]q
}
`, rName))
}

func testAccEndpointAccessConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccEndpointAccessConfig_base(rName), fmt.Sprintf(`
resource "aws_redshiftserverless_endpoint_access" "test" {
  workgroup_name = aws_redshiftserverless_workgroup.test.workgroup_name
  endpoint_name  = %[1]q
  subnet_ids     = [aws_subnet.test[0].id]
}
`, rName))
}

func testAccEndpointAccessConfig_updated(rName string) string {
	return acctest.ConfigCompose(testAccEndpointAccessConfig_base(rName), fmt.Sprintf(`
resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_redshiftserverless_endpoint_access" "test" {
  workgroup_name         = aws_redshiftserverless_workgroup.test.workgroup_name
  endpoint_name          = %[1]q
  subnet_ids             = [aws_subnet.test[0].id]
  vpc_security_group_ids = [aws_security_group.test.id]
}
`, rName))
}
