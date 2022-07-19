package iam_test


import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccS3BucketsDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_iam_buckets.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccS3BucketsDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "ids.#", regexp.MustCompile("[^0].*$")),
				),
			},
		},
	})
}

func TestAccS3BucketsDataSource_nameRegex(t *testing.T) {
	rCount := strconv.Itoa(sdkacctest.RandIntRange(1, 4))
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_iam_buckets.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccS3BucketsDataSourceConfig_nameRegex(rCount, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", rCount),
				),
			},
		},
	})
}

const testAccS3BucketsDataSourceConfig_basic = `
data "aws_s3_buckets" "test" {}
`

func testAccS3BucketsDataSourceConfig_nameRegex(rCount, rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  count   = %[1]s
  bucket  = "%[2]s-${count.index}-bucket"

  tags = {
    Seed = %[2]q
  }
}

data "aws_s3_buckets" "test" {
  name_regex = "${aws_s3_bucket.test[0].tags["Seed"]}-.*-bucket"
}
`, rCount, rName)
}
