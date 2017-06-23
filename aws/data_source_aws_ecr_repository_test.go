package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSEcrDataSource_ecrRepository(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsEcrRepositoryDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("data.aws_ecr_repository.default", "arn", regexp.MustCompile("^arn:aws:ecr:[a-zA-Z]+-[a-zA-Z]+-\\d+:\\d+:repository/foo-repository-terraform-\\d+$")),
					resource.TestCheckResourceAttrSet("data.aws_ecr_repository.default", "registry_id"),
					resource.TestMatchResourceAttr("data.aws_ecr_repository.default", "repository_url", regexp.MustCompile("^\\d+\\.dkr\\.ecr\\.[a-zA-Z]+-[a-zA-Z]+-\\d+\\.amazonaws\\.com/foo-repository-terraform-\\d+$")),
				),
			},
		},
	})
}

var testAccCheckAwsEcrRepositoryDataSourceConfig = fmt.Sprintf(`
resource "aws_ecr_repository" "default" {
  name = "foo-repository-terraform-%d"
}

data "aws_ecr_repository" "default" {
  name = "${aws_ecr_repository.default.name}"
}
`, acctest.RandInt())
