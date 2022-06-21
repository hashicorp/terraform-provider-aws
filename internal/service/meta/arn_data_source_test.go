package meta_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfmeta "github.com/hashicorp/terraform-provider-aws/internal/service/meta"
)

func TestAccMetaARNDataSource_basic(t *testing.T) {
	resourceName := "data.aws_arn.test"

	testARN := arn.ARN{
		AccountID: "123456789012",
		Partition: endpoints.AwsPartitionID,
		Region:    endpoints.EuWest1RegionID,
		Resource:  "db:mysql-db",
		Service:   "rds",
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccARNDataSourceConfig_basic(testARN.String()),
				Check: resource.ComposeTestCheckFunc(
					testAccARNDataSource(resourceName),
					resource.TestCheckResourceAttr(resourceName, "account", testARN.AccountID),
					resource.TestCheckResourceAttr(resourceName, "partition", testARN.Partition),
					resource.TestCheckResourceAttr(resourceName, "region", testARN.Region),
					resource.TestCheckResourceAttr(resourceName, "resource", testARN.Resource),
					resource.TestCheckResourceAttr(resourceName, "service", testARN.Service),
				),
			},
		},
	})
}

func testAccARNDataSource(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		return nil
	}
}

func testAccARNDataSourceConfig_basic(arn string) string {
	return fmt.Sprintf(`
data "aws_arn" "test" {
  arn = %q
}
`, arn)
}
