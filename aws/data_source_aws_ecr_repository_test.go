package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAWSEcrDataSource_ecrRepository(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecr_repository.test"
	dataSourceName := "data.aws_ecr_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsEcrRepositoryDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "registry_id", dataSourceName, "registry_id"),
					resource.TestCheckResourceAttrPair(resourceName, "repository_url", dataSourceName, "repository_url"),
					resource.TestCheckResourceAttrPair(resourceName, "tags", dataSourceName, "tags"),
				),
			},
		},
	})
}

func TestAccAWSEcrDataSource_nonExistent(t *testing.T) {

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckAwsEcrRepositoryDataSourceConfig_NonExistent,
				ExpectError: regexp.MustCompile(`not found`),
			},
		},
	})
}

const testAccCheckAwsEcrRepositoryDataSourceConfig_NonExistent = `
data "aws_ecr_repository" "test" {
  name = "tf-acc-test-non-existent"
}
`

func testAccCheckAwsEcrRepositoryDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name = %q

  tags = {
    Environment = "production"
    Usage       = "original"
  }
}

data "aws_ecr_repository" "test" {
  name = aws_ecr_repository.test.name
}
`, rName)
}
