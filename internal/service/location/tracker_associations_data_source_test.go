package location_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/locationservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccLocationTrackerAssociationsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_location_tracker_associations.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, locationservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrackerAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrackerAssociationsDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "consumer_arns.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "consumer_arns.0", "aws_location_tracker_association.test", "consumer_arn"),
				),
			},
		},
	})
}

func testAccTrackerAssociationsDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_location_geofence_collection" "test" {
  collection_name = %[1]q
}

resource "aws_location_tracker" "test" {
  tracker_name = %[1]q
}

resource "aws_location_tracker_association" "test" {
  consumer_arn = aws_location_geofence_collection.test.collection_arn
  tracker_name = aws_location_tracker.test.tracker_name
}

data "aws_location_tracker_associations" "test" {
  tracker_name = aws_location_tracker_association.test.tracker_name
}
`, rName)
}
