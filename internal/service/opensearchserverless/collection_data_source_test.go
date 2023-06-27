package opensearchserverless_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOpenSearchServerlessCollectionDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var collection types.CollectionDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_opensearchserverless_collection.test"
	resourceName := "aws_opensearchserverless_collection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCollectionDataSourceConfig_basic(rName, "encryption"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionExists(ctx, dataSourceName, &collection),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
				),
			},
		},
	})
}

func testAccCollectionDataSourceConfig_basic(rName, policyType string) string {
	return fmt.Sprintf(`
resource "aws_opensearchserverless_security_policy" "test" {
  name = %[1]q
  type = %[2]q
  policy = jsonencode({
    Rules = [
      {
        Resource = [
          "collection/%[1]s"
        ],
        ResourceType = "collection"
      }
    ],
    AWSOwnedKey = true
  })
}

resource "aws_opensearchserverless_collection" "test" {
  name       = %[1]q
  depends_on = [aws_opensearchserverless_security_policy.test]
}

data "aws_opensearchserverless_collection" "test" {
  id = aws_opensearchserverless_collection.test.id
}
`, rName, policyType)
}
