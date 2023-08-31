// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/appsync"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappsync "github.com/hashicorp/terraform-provider-aws/internal/service/appsync"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func testAccAPICache_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var apiCache appsync.ApiCache
	resourceName := "aws_appsync_api_cache.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appsync.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, appsync.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPICacheDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAPICacheConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPICacheExists(ctx, resourceName, &apiCache),
					resource.TestCheckResourceAttrPair(resourceName, "api_id", "aws_appsync_graphql_api.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "type", "SMALL"),
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
	var apiCache appsync.ApiCache
	resourceName := "aws_appsync_api_cache.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appsync.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, appsync.EndpointsID),
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
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppSyncConn(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appsync_api_cache" {
				continue
			}

			_, err := tfappsync.FindAPICacheByID(ctx, conn, rs.Primary.ID)
			if err == nil {
				if tfresource.NotFound(err) {
					return nil
				}
				return err
			}

			return nil
		}
		return nil
	}
}

func testAccCheckAPICacheExists(ctx context.Context, resourceName string, apiCache *appsync.ApiCache) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Appsync Api Cache Not found in state: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppSyncConn(ctx)
		cache, err := tfappsync.FindAPICacheByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*apiCache = *cache

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
