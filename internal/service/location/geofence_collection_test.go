// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package location_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/location"
	awstypes "github.com/aws/aws-sdk-go-v2/service/location/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tflocation "github.com/hashicorp/terraform-provider-aws/internal/service/location"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLocationGeofenceCollection_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_location_geofence_collection.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGeofenceCollectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGeofenceCollectionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGeofenceCollectionExists(ctx, t, resourceName),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, "collection_arn", "geo", fmt.Sprintf("geofence-collection/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "collection_name", rName),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreateTime),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
					acctest.CheckResourceAttrRFC3339(resourceName, "update_time"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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

func TestAccLocationGeofenceCollection_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_location_geofence_collection.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGeofenceCollectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGeofenceCollectionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGeofenceCollectionExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tflocation.ResourceGeofenceCollection(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLocationGeofenceCollection_description(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_location_geofence_collection.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGeofenceCollectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGeofenceCollectionConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGeofenceCollectionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGeofenceCollectionConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGeofenceCollectionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
				),
			},
		},
	})
}

func TestAccLocationGeofenceCollection_kmsKeyID(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_location_geofence_collection.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGeofenceCollectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGeofenceCollectionConfig_kmsKeyID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGeofenceCollectionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, "aws_kms_key.test", names.AttrARN),
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

func TestAccLocationGeofenceCollection_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_location_geofence_collection.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGeofenceCollectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGeofenceCollectionConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGeofenceCollectionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGeofenceCollectionConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGeofenceCollectionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccGeofenceCollectionConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGeofenceCollectionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckGeofenceCollectionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LocationClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_location_geofence_collection" {
				continue
			}

			input := &location.DescribeGeofenceCollectionInput{
				CollectionName: aws.String(rs.Primary.ID),
			}

			_, err := conn.DescribeGeofenceCollection(ctx, input)
			if err != nil {
				if errs.IsA[*awstypes.ResourceNotFoundException](err) {
					return nil
				}
				return err
			}

			return create.Error(names.Location, create.ErrActionCheckingDestroyed, tflocation.ResNameGeofenceCollection, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckGeofenceCollectionExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Location, create.ErrActionCheckingExistence, tflocation.ResNameGeofenceCollection, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Location, create.ErrActionCheckingExistence, tflocation.ResNameGeofenceCollection, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).LocationClient(ctx)
		_, err := conn.DescribeGeofenceCollection(ctx, &location.DescribeGeofenceCollectionInput{
			CollectionName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.Location, create.ErrActionCheckingExistence, tflocation.ResNameGeofenceCollection, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccGeofenceCollectionConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_location_geofence_collection" "test" {
  collection_name = %[1]q
}
`, rName)
}

func testAccGeofenceCollectionConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_location_geofence_collection" "test" {
  collection_name = %[1]q
  description     = %[2]q
}
`, rName, description)
}

func testAccGeofenceCollectionConfig_kmsKeyID(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_location_geofence_collection" "test" {
  collection_name = %[1]q
  kms_key_id      = aws_kms_key.test.arn
}
`, rName)
}

func testAccGeofenceCollectionConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_location_geofence_collection" "test" {
  collection_name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccGeofenceCollectionConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_location_geofence_collection" "test" {
  collection_name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
