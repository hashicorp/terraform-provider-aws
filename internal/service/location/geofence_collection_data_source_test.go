package location_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/locationservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccLocationGeofenceCollectionDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_location_geofence_collection.test"
	resourceName := "aws_location_geofence_collection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, locationservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGeofenceCollectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGeofenceCollectionDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGeofenceCollectionExists(ctx, dataSourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, "collection_arn", resourceName, "collection_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "collection_name", resourceName, "collection_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "create_time", resourceName, "create_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "kms_key_id", resourceName, "kms_key_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "update_time", resourceName, "update_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
				),
			},
		},
	})
}

func testAccGeofenceCollectionDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_location_geofence_collection" "test" {
  collection_name = %[1]q
}

data "aws_location_geofence_collection" "test" {
  collection_name = aws_location_geofence_collection.test.collection_name
}
`, rName)
}
