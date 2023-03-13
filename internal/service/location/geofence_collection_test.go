package location_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/locationservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tflocation "github.com/hashicorp/terraform-provider-aws/internal/service/location"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLocationGeofenceCollection_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_location_geofence_collection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, locationservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGeofenceCollectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGeofenceCollectionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGeofenceCollectionExists(ctx, resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "collection_arn", "geo", fmt.Sprintf("geofence-collection/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "collection_name", rName),
					acctest.CheckResourceAttrRFC3339(resourceName, "create_time"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
					acctest.CheckResourceAttrRFC3339(resourceName, "update_time"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_location_geofence_collection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, locationservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGeofenceCollectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGeofenceCollectionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGeofenceCollectionExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflocation.ResourceGeofenceCollection(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLocationGeofenceCollection_description(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_location_geofence_collection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, locationservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGeofenceCollectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGeofenceCollectionConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGeofenceCollectionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
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
					testAccCheckGeofenceCollectionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func TestAccLocationGeofenceCollection_kmsKeyID(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_location_geofence_collection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, locationservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGeofenceCollectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGeofenceCollectionConfig_kmsKeyID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGeofenceCollectionExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", "aws_kms_key.test", "arn"),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_location_geofence_collection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, locationservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGeofenceCollectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGeofenceCollectionConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGeofenceCollectionExists(ctx, resourceName),
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
				Config: testAccGeofenceCollectionConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGeofenceCollectionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccGeofenceCollectionConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGeofenceCollectionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckGeofenceCollectionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LocationConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_location_geofence_collection" {
				continue
			}

			input := &locationservice.DescribeGeofenceCollectionInput{
				CollectionName: aws.String(rs.Primary.ID),
			}

			_, err := conn.DescribeGeofenceCollectionWithContext(ctx, input)
			if err != nil {
				if tfawserr.ErrCodeEquals(err, locationservice.ErrCodeResourceNotFoundException) {
					return nil
				}
				return err
			}

			return create.Error(names.Location, create.ErrActionCheckingDestroyed, tflocation.ResNameGeofenceCollection, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckGeofenceCollectionExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Location, create.ErrActionCheckingExistence, tflocation.ResNameGeofenceCollection, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Location, create.ErrActionCheckingExistence, tflocation.ResNameGeofenceCollection, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LocationConn()
		_, err := conn.DescribeGeofenceCollectionWithContext(ctx, &locationservice.DescribeGeofenceCollectionInput{
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
