// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appsync/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappsync "github.com/hashicorp/terraform-provider-aws/internal/service/appsync"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccType_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var typ awstypes.Type
	resourceName := "aws_appsync_type.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTypeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTypeConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTypeExists(ctx, resourceName, &typ),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "appsync", regexache.MustCompile("apis/.+/types/.+")),
					resource.TestCheckResourceAttrPair(resourceName, "api_id", "aws_appsync_graphql_api.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrFormat, "SDL"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "Mutation"),
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

func testAccType_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var typ awstypes.Type
	resourceName := "aws_appsync_type.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTypeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTypeConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTypeExists(ctx, resourceName, &typ),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfappsync.ResourceType(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckTypeDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppSyncClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appsync_type" {
				continue
			}

			_, err := tfappsync.FindTypeByThreePartKey(ctx, conn, rs.Primary.Attributes["api_id"], awstypes.TypeDefinitionFormat(rs.Primary.Attributes[names.AttrFormat]), rs.Primary.Attributes[names.AttrName])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Appsync Type %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckTypeExists(ctx context.Context, n string, v *awstypes.Type) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppSyncClient(ctx)

		output, err := tfappsync.FindTypeByThreePartKey(ctx, conn, rs.Primary.Attributes["api_id"], awstypes.TypeDefinitionFormat(rs.Primary.Attributes[names.AttrFormat]), rs.Primary.Attributes[names.AttrName])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccTypeConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %[1]q
}

resource "aws_appsync_type" "test" {
  api_id     = aws_appsync_graphql_api.test.id
  format     = "SDL"
  definition = <<EOF
type Mutation

{
putPost(id: ID!,title: String! ): Post

}
EOF  
}
`, rName)
}
