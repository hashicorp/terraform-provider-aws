// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ram_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ram"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfram "github.com/hashicorp/terraform-provider-aws/internal/service/ram"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRAMPrincipalAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var association ram.ResourceShareAssociation
	resourceName := "aws_ram_principal_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPrincipalAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPrincipalAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrincipalAssociationExists(ctx, resourceName, &association),
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

func TestAccRAMPrincipalAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var association ram.ResourceShareAssociation
	resourceName := "aws_ram_principal_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPrincipalAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPrincipalAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrincipalAssociationExists(ctx, resourceName, &association),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfram.ResourcePrincipalAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckPrincipalAssociationExists(ctx context.Context, n string, v *ram.ResourceShareAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RAMConn(ctx)

		output, err := tfram.FindPrincipalAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["resource_share_arn"], rs.Primary.Attributes["principal"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckPrincipalAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RAMConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ram_principal_association" {
				continue
			}

			_, err := tfram.FindPrincipalAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["resource_share_arn"], rs.Primary.Attributes["principal"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("RAM Resource Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccPrincipalAssociationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_ram_resource_share" "test" {
  allow_external_principals = true
  name                      = %[1]q
}

resource "aws_ram_principal_association" "test" {
  principal          = "111111111111"
  resource_share_arn = aws_ram_resource_share.test.id
}
`, rName)
}
