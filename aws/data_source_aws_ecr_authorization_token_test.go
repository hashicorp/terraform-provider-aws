package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAWSEcrAuthorizationTokenDataSource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "data.aws_ecr_authorization_token.repo"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsEcrAuthorizationTokenDataSourceBasicConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "authorization_token"),
					resource.TestCheckResourceAttrSet(resourceName, "proxy_endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, "expires_at"),
					resource.TestCheckResourceAttrSet(resourceName, "user_name"),
					resource.TestMatchResourceAttr(resourceName, "user_name", regexp.MustCompile(`AWS`)),
					resource.TestCheckResourceAttrSet(resourceName, "password"),
				),
			},
			{
				Config: testAccCheckAwsEcrAuthorizationTokenDataSourceRepositoryConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "registry_id", "aws_ecr_repository.repo", "registry_id"),
					resource.TestCheckResourceAttrSet(resourceName, "authorization_token"),
					resource.TestCheckResourceAttrSet(resourceName, "proxy_endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, "expires_at"),
					resource.TestCheckResourceAttrSet(resourceName, "user_name"),
					resource.TestMatchResourceAttr(resourceName, "user_name", regexp.MustCompile(`AWS`)),
					resource.TestCheckResourceAttrSet(resourceName, "password"),
				),
			},
		},
	})
}

var testAccCheckAwsEcrAuthorizationTokenDataSourceBasicConfig = `
data "aws_ecr_authorization_token" "repo" {}
`

func testAccCheckAwsEcrAuthorizationTokenDataSourceRepositoryConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "repo" {
  name = %q
}
data "aws_ecr_authorization_token" "repo" {
	registry_id = "${aws_ecr_repository.repo.registry_id}"
}
`, rName)
}
