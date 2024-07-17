// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync_test

import (
	"context"
	"fmt"
	"testing"

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

func testAccAPICache_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var apiCache awstypes.ApiCache
	resourceName := "aws_appsync_api_cache.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPICacheDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAPICacheConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPICacheExists(ctx, resourceName, &apiCache),
					resource.TestCheckResourceAttrPair(resourceName, "api_id", "aws_appsync_graphql_api.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "SMALL"),
					resource.TestCheckResourceAttr(resourceName, "api_caching_behavior", "FULL_REQUEST_CACHING"),
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

func testAccAPICache_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var apiCache awstypes.ApiCache
	resourceName := "aws_appsync_api_cache.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPICacheDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAPICacheConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPICacheExists(ctx, resourceName, &apiCache),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfappsync.ResourceAPICache(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAPICacheDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppSyncClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appsync_api_cache" {
				continue
			}

			_, err := tfappsync.FindAPICacheByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Appsync API Cache %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAPICacheExists(ctx context.Context, n string, v *awstypes.ApiCache) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppSyncClient(ctx)

		output, err := tfappsync.FindAPICacheByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAPICacheConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %[1]q
}

resource "aws_appsync_api_cache" "test" {
  api_id               = aws_appsync_graphql_api.test.id
  type                 = "SMALL"
  api_caching_behavior = "FULL_REQUEST_CACHING"
  ttl                  = 60
}
`, rName)
}
