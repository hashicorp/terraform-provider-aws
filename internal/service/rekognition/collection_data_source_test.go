package rekognition_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/rekognition"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccRekognitionCollectionDataSource_basic(t *testing.T) {
	resourceName := "aws_rekognition_collection.test"
	datasourceName := "data.aws_rekognition_collection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, rekognition.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCollectionDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "collection_id", resourceName, "collection_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "collection_arn", resourceName, "collection_arn"),
				),
			},
		},
	})
}

func testAccCollectionDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_rekognition_collection" "test" {
  collection_id = %[1]q
}

data "aws_rekognition_collection" "test" {
  collection_id = aws_rekognition_collection.test.collection_id
}
`, rName)
}
