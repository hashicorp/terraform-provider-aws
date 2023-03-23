package iam_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/service/iam"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccIAMUsersDataSource_nameRegex(t *testing.T) {
	dataSourceName := "data.aws_iam_users.test"
	rCount := strconv.Itoa(sdkacctest.RandIntRange(1, 4))
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUsersDataSourceConfig_nameRegex(rCount, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "names.#", rCount),
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", rCount),
				),
			},
		},
	})
}

func TestAccIAMUsersDataSource_pathPrefix(t *testing.T) {
	dataSourceName := "data.aws_iam_users.test"
	rCount := strconv.Itoa(sdkacctest.RandIntRange(1, 4))
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rPathPrefix := sdkacctest.RandomWithPrefix("tf-acc-path")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUsersDataSourceConfig_pathPrefix(rCount, rName, rPathPrefix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "names.#", rCount),
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", rCount),
				),
			},
		},
	})
}

func TestAccIAMUsersDataSource_nonExistentNameRegex(t *testing.T) {
	dataSourceName := "data.aws_iam_users.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUsersDataSourceConfig_nonExistentNameRegex,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "names.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", "0"),
				),
			},
		},
	})
}

func TestAccIAMUsersDataSource_nonExistentPathPrefix(t *testing.T) {
	dataSourceName := "data.aws_iam_users.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUsersDataSourceConfig_nonExistentPathPrefix,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "names.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", "0"),
				),
			},
		},
	})
}

func testAccUsersDataSourceConfig_nameRegex(rCount, rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  count = %[1]s
  name  = "%[2]s-${count.index}-user"

  tags = {
    Seed = %[2]q
  }
}

data "aws_iam_users" "test" {
  name_regex = "${aws_iam_user.test[0].tags["Seed"]}-.*-user"
}
`, rCount, rName)
}

func testAccUsersDataSourceConfig_pathPrefix(rCount, rName, rPathPrefix string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  count = %[1]s
  name  = "%[2]s-${count.index}-user"
  path  = "/%[3]s/"
}

data "aws_iam_users" "test" {
  path_prefix = aws_iam_user.test[0].path
}
`, rCount, rName, rPathPrefix)
}

const testAccUsersDataSourceConfig_nonExistentNameRegex = `
data "aws_iam_users" "test" {
  name_regex = "dne-regex"
}
`

const testAccUsersDataSourceConfig_nonExistentPathPrefix = `
data "aws_iam_users" "test" {
  path_prefix = "/dne/path"
}
`
