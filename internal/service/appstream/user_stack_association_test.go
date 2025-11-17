// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appstream_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/appstream/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappstream "github.com/hashicorp/terraform-provider-aws/internal/service/appstream"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppStreamUserStackAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_appstream_user_stack_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	authType := "USERPOOL"
	domain := acctest.RandomDomainName()
	rEmail := acctest.RandomEmailAddress(domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserStackAssociationDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccUserStackAssociationConfig_basic(rName, authType, rEmail),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserStackAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", authType),
					resource.TestCheckResourceAttr(resourceName, "stack_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrUserName, rEmail),
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

func TestAccAppStreamUserStackAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_appstream_user_stack_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	authType := "USERPOOL"
	domain := acctest.RandomDomainName()
	rEmail := acctest.RandomEmailAddress(domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserStackAssociationDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccUserStackAssociationConfig_basic(rName, authType, rEmail),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserStackAssociationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfappstream.ResourceUserStackAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAppStreamUserStackAssociation_Disappears_user(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_appstream_user_stack_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	authType := "USERPOOL"
	domain := acctest.RandomDomainName()
	rEmail := acctest.RandomEmailAddress(domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserStackAssociationDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccUserStackAssociationConfig_basic(rName, authType, rEmail),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserStackAssociationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfappstream.ResourceUser(), "aws_appstream_user.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAppStreamUserStackAssociation_Disappears_stack(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_appstream_user_stack_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	authType := "USERPOOL"
	domain := acctest.RandomDomainName()
	rEmail := acctest.RandomEmailAddress(domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserStackAssociationDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccUserStackAssociationConfig_basic(rName, authType, rEmail),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserStackAssociationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfappstream.ResourceStack(), "aws_appstream_stack.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAppStreamUserStackAssociation_complete(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_appstream_user_stack_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	authType := "USERPOOL"
	domain := acctest.RandomDomainName()
	rEmail := acctest.RandomEmailAddress(domain)
	rEmailUpdated := acctest.RandomEmailAddress(domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserStackAssociationDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccUserStackAssociationConfig_basic(rName, authType, rEmail),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserStackAssociationExists(ctx, resourceName),
				),
			},
			{
				Config: testAccUserStackAssociationConfig_basic(rName, authType, rEmailUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserStackAssociationExists(ctx, resourceName),
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

func testAccCheckUserStackAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appstream_user_stack_association" {
				continue
			}

			_, err := tfappstream.FindUserStackAssociationByThreePartKey(ctx, conn, rs.Primary.Attributes[names.AttrUserName], awstypes.AuthenticationType(rs.Primary.Attributes["authentication_type"]), rs.Primary.Attributes["stack_name"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("AppStream User Stack Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckUserStackAssociationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamClient(ctx)

		_, err := tfappstream.FindUserStackAssociationByThreePartKey(ctx, conn, rs.Primary.Attributes[names.AttrUserName], awstypes.AuthenticationType(rs.Primary.Attributes["authentication_type"]), rs.Primary.Attributes["stack_name"])

		return err
	}
}

func testAccUserStackAssociationConfig_basic(name, authType, userName string) string {
	return fmt.Sprintf(`
resource "aws_appstream_stack" "test" {
  name = %[1]q
}

resource "aws_appstream_user" "test" {
  authentication_type = %[2]q
  user_name           = %[3]q
}

resource "aws_appstream_user_stack_association" "test" {
  authentication_type = aws_appstream_user.test.authentication_type
  stack_name          = aws_appstream_stack.test.name
  user_name           = aws_appstream_user.test.user_name
}
`, name, authType, userName)
}
