// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync_test

import (
	"context"
	"fmt"
	"testing"
	"time"

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

func testAccAPIKey_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var apiKey awstypes.ApiKey
	dateAfterSevenDays := time.Now().UTC().Add(time.Hour * 24 * time.Duration(7)).Truncate(time.Hour)
	resourceName := "aws_appsync_api_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyConfig_required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists(ctx, resourceName, &apiKey),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Managed by Terraform"),
					testAccCheckAPIKeyExpiresDate(&apiKey, dateAfterSevenDays),
					resource.TestMatchResourceAttr(resourceName, names.AttrKey, regexache.MustCompile(`.+`)),
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

func testAccAPIKey_description(t *testing.T) {
	ctx := acctest.Context(t)
	var apiKey awstypes.ApiKey
	resourceName := "aws_appsync_api_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists(ctx, resourceName, &apiKey),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
				),
			},
			{
				Config: testAccAPIKeyConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists(ctx, resourceName, &apiKey),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
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

func testAccAPIKey_expires(t *testing.T) {
	ctx := acctest.Context(t)
	var apiKey awstypes.ApiKey
	dateAfterTenDays := time.Now().UTC().Add(time.Hour * 24 * time.Duration(10)).Truncate(time.Hour)
	dateAfterTwentyDays := time.Now().UTC().Add(time.Hour * 24 * time.Duration(20)).Truncate(time.Hour)
	resourceName := "aws_appsync_api_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyConfig_expires(rName, dateAfterTenDays.Format(time.RFC3339)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists(ctx, resourceName, &apiKey),
					testAccCheckAPIKeyExpiresDate(&apiKey, dateAfterTenDays),
				),
			},
			{
				Config: testAccAPIKeyConfig_expires(rName, dateAfterTwentyDays.Format(time.RFC3339)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists(ctx, resourceName, &apiKey),
					testAccCheckAPIKeyExpiresDate(&apiKey, dateAfterTwentyDays),
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

func testAccCheckAPIKeyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppSyncClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appsync_api_key" {
				continue
			}

			_, err := tfappsync.FindAPIKeyByTwoPartKey(ctx, conn, rs.Primary.Attributes["api_id"], rs.Primary.Attributes["api_key_id"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Appsync API Key %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAPIKeyExists(ctx context.Context, n string, v *awstypes.ApiKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppSyncClient(ctx)

		output, err := tfappsync.FindAPIKeyByTwoPartKey(ctx, conn, rs.Primary.Attributes["api_id"], rs.Primary.Attributes["api_key_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckAPIKeyExpiresDate(apiKey *awstypes.ApiKey, expectedTime time.Time) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiKeyExpiresTime := time.Unix(apiKey.Expires, 0)
		if !apiKeyExpiresTime.Equal(expectedTime) {
			return fmt.Errorf("Appsync API Key expires difference: got %s and expected %s", apiKeyExpiresTime.Format(time.RFC3339), expectedTime.Format(time.RFC3339))
		}

		return nil
	}
}

func testAccAPIKeyConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q
}

resource "aws_appsync_api_key" "test" {
  api_id      = aws_appsync_graphql_api.test.id
  description = %q
}
`, rName, description)
}

func testAccAPIKeyConfig_expires(rName, expires string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q
}

resource "aws_appsync_api_key" "test" {
  api_id  = aws_appsync_graphql_api.test.id
  expires = %q
}
`, rName, expires)
}

func testAccAPIKeyConfig_required(rName string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q
}

resource "aws_appsync_api_key" "test" {
  api_id = aws_appsync_graphql_api.test.id
}
`, rName)
}
