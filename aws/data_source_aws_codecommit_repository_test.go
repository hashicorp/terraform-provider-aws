package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSCodeCommitRepositoryDataSource_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-acctest-%d", acctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsCodeCommitRepositoryDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_codecommit_repository.default", "repository_name", rName),
					resource.TestMatchResourceAttr("data.aws_codecommit_repository.default", "arn",
						regexp.MustCompile(fmt.Sprintf("^arn:aws:codecommit:[^:]+:\\d{12}:%s", rName))),
					resource.TestMatchResourceAttr("data.aws_codecommit_repository.default", "clone_url_http",
						regexp.MustCompile(fmt.Sprintf("^https://git-codecommit.[^:]+.amazonaws.com/v1/repos/%s", rName))),
					resource.TestMatchResourceAttr("data.aws_codecommit_repository.default", "clone_url_ssh",
						regexp.MustCompile(fmt.Sprintf("^ssh://git-codecommit.[^:]+.amazonaws.com/v1/repos/%s", rName))),
				),
			},
		},
	})
}

func testAccCheckAwsCodeCommitRepositoryDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_codecommit_repository" "default" {
  repository_name = "%s"
}

data "aws_codecommit_repository" "default" {
  repository_name = "${aws_codecommit_repository.default.repository_name}"
}
`, rName)
}
