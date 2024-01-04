// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package location_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/location"
	"github.com/aws/aws-sdk-go-v2/service/location/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tflocation "github.com/hashicorp/terraform-provider-aws/internal/service/location"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLocationKey_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var key location.DescribeKeyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_location_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LocationEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &key),
					acctest.CheckResourceAttrRFC3339(resourceName, "create_time"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "key_arn", "geo", fmt.Sprintf("api-key/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "key_name", rName),
					resource.TestCheckResourceAttr(resourceName, "no_expiry", "true"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					acctest.CheckResourceAttrRFC3339(resourceName, "update_time"),
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

func TestAccLocationKey_expirationDate(t *testing.T) {
	ctx := acctest.Context(t)
	var key location.DescribeKeyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	expireTime := time.Now().Add(48 * time.Hour).UTC().Format(time.RFC3339)
	resourceName := "aws_location_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LocationEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig_expireTime(rName, expireTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "expire_time", expireTime),
					resource.TestCheckNoResourceAttr(resourceName, "no_expiry"),
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

func TestAccLocationKey_update(t *testing.T) {
	ctx := acctest.Context(t)
	var key location.DescribeKeyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	expireTime := time.Now().Add(48 * time.Hour).UTC().Format(time.RFC3339)
	resourceName := "aws_location_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LocationEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &key),
					resource.TestCheckNoResourceAttr(resourceName, "expire_time"),
					resource.TestCheckResourceAttr(resourceName, "no_expiry", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccKeyConfig_basicUpdated(rName, expireTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "expire_time", expireTime),
					resource.TestCheckNoResourceAttr(resourceName, "no_expiry"),
				),
			},
		},
	})
}

func TestAccLocationKey_forceUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var key location.DescribeKeyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_location_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LocationEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "restrictions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "restrictions.0.allow_actions.#", "4"),
					resource.TestCheckResourceAttr(resourceName, "restrictions.0.allow_actions.0", "geo:SearchPlaceIndexForText"),
					resource.TestCheckResourceAttr(resourceName, "restrictions.0.allow_actions.1", "geo:SearchPlaceIndexForPosition"),
					resource.TestCheckResourceAttr(resourceName, "restrictions.0.allow_actions.2", "geo:SearchPlaceIndexForSuggestions"),
					resource.TestCheckResourceAttr(resourceName, "restrictions.0.allow_actions.3", "geo:GetPlace"),
					resource.TestCheckResourceAttr(resourceName, "restrictions.0.allow_resources.#", "1"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "restrictions.0.allow_resources.0", "geo", fmt.Sprintf("place-index/%s", rName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccKeyConfig_forceUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "restrictions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "restrictions.0.allow_actions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "restrictions.0.allow_actions.0", "geo:GetMap*"),
					resource.TestCheckResourceAttr(resourceName, "restrictions.0.allow_referers.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "restrictions.0.allow_referers.0", "example.test"),
					resource.TestCheckResourceAttr(resourceName, "restrictions.0.allow_referers.1", "test.test"),
					resource.TestCheckResourceAttr(resourceName, "restrictions.0.allow_resources.#", "1"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "restrictions.0.allow_resources.0", "geo", fmt.Sprintf("map/%s", rName)),
				),
			},
		},
	})
}

func TestAccLocationKey_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var key location.DescribeKeyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_location_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LocationEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "key_name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccKeyConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "key_name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccKeyConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "key_name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccLocationKey_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var key location.DescribeKeyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_location_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LocationEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &key),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tflocation.ResourceKey, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckKeyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LocationClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_location_key" {
				continue
			}

			_, err := conn.DescribeKey(ctx, &location.DescribeKeyInput{
				KeyName: aws.String(rs.Primary.ID),
			})
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.Location, create.ErrActionCheckingDestroyed, tflocation.ResNameKey, rs.Primary.ID, err)
			}

			return create.Error(names.Location, create.ErrActionCheckingDestroyed, tflocation.ResNameKey, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckKeyExists(ctx context.Context, name string, key *location.DescribeKeyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Location, create.ErrActionCheckingExistence, tflocation.ResNameKey, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Location, create.ErrActionCheckingExistence, tflocation.ResNameKey, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LocationClient(ctx)
		resp, err := conn.DescribeKey(ctx, &location.DescribeKeyInput{
			KeyName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.Location, create.ErrActionCheckingExistence, tflocation.ResNameKey, rs.Primary.ID, err)
		}

		*key = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LocationClient(ctx)

	input := &location.ListKeysInput{}
	_, err := conn.ListKeys(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccKeyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_location_place_index" "test" {
  data_source = "Here"
  index_name  = %[1]q
}

resource "aws_location_key" "test" {
  key_name = %[1]q
  no_expiry = true

  restrictions {
    allow_actions = [
      "geo:SearchPlaceIndexForText",
      "geo:SearchPlaceIndexForPosition",
      "geo:SearchPlaceIndexForSuggestions",
      "geo:GetPlace",
    ]
    allow_resources = [
      aws_location_place_index.test.index_arn,
    ]
  }
}
`, rName)
}

func testAccKeyConfig_expireTime(rName string, expireTime string) string {
	return fmt.Sprintf(`
resource "aws_location_place_index" "test" {
	data_source = "Esri"
	index_name  = %[1]q
}

resource "aws_location_key" "test" {
	key_name = %[1]q
	expire_time = "%[2]s"

	restrictions {
		allow_actions = [
			"geo:SearchPlaceIndexForText",
		]
		allow_resources = [
			aws_location_place_index.test.index_arn,
		]
	}
}
`, rName, expireTime)
}

func testAccKeyConfig_basicUpdated(rName string, expireTime string) string {
	return fmt.Sprintf(`
resource "aws_location_place_index" "test" {
  data_source = "Here"
  index_name  = %[1]q
}

resource "aws_location_key" "test" {
  key_name = %[1]q
  expire_time = "%[2]s"

  restrictions {
    allow_actions = [
      "geo:SearchPlaceIndexForText",
      "geo:SearchPlaceIndexForPosition",
      "geo:SearchPlaceIndexForSuggestions",
      "geo:GetPlace",
    ]
    allow_resources = [
      aws_location_place_index.test.index_arn,
    ]
  }
}
`, rName, expireTime)
}

func testAccKeyConfig_forceUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_location_map" "test" {
  configuration {
    style = "VectorHereBerlin"
  }

  map_name = %[1]q
}

resource "aws_location_key" "test" {
  key_name = %[1]q
  no_expiry = true
  force_update = true

  restrictions {
    allow_actions = [
      "geo:GetMap*",
    ]
    allow_referers = [
      "example.test",
      "test.test",
	]
    allow_resources = [
      aws_location_map.test.map_arn,
    ]
  }
}
`, rName)
}

func testAccKeyConfig_tags1(rName, key1, value1 string) string {
	return fmt.Sprintf(`
resource "aws_location_place_index" "test" {
  data_source = "Here"
  index_name  = %[1]q
}

resource "aws_location_key" "test" {
  key_name = %[1]q
  no_expiry = true

  restrictions {
    allow_actions = [
      "geo:SearchPlaceIndexForText",
    ]
    allow_resources = [
      aws_location_place_index.test.index_arn,
    ]
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, key1, value1)
}

func testAccKeyConfig_tags2(rName, key1, value1, key2, value2 string) string {
	return fmt.Sprintf(`
resource "aws_location_place_index" "test" {
  data_source = "Here"
  index_name  = %[1]q
}

resource "aws_location_key" "test" {
  key_name = %[1]q
  no_expiry = true

  restrictions {
    allow_actions = [
      "geo:SearchPlaceIndexForText",
    ]
    allow_resources = [
      aws_location_place_index.test.index_arn,
    ]
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, key1, value1, key2, value2)
}
