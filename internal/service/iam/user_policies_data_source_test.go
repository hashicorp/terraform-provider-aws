package iam_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/iam"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccIAMUserPoliciesDataSource_basic(t *testing.T) {
	resourceName := "aws_iam_user_policy.test"
	dataSourceName := "data.aws_iam_user_policies.test"

	user := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, iam.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoliciesDataSourceConfig_basic(user),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "user", resourceName, "user"),
					resource.TestCheckResourceAttr(dataSourceName, "names.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "names.0", resourceName, "name"),
				),
			},
		},
	})
}

func TestAccIAMUserPoliciesDataSource_multiple(t *testing.T) {
	resourceName0 := "aws_iam_user_policy.test.0"
	resourceName1 := "aws_iam_user_policy.test.1"
	dataSourceName := "data.aws_iam_user_policies.test"

	user := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	count := "2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, iam.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoliciesDataSourceConfig_multiple(user, count),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "user", resourceName0, "user"),
					resource.TestCheckResourceAttrPair(dataSourceName, "user", resourceName1, "user"),
					resource.TestCheckResourceAttr(dataSourceName, "names.#", count),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "names.*", resourceName0, "name"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "names.*", resourceName1, "name"),
				),
			},
		},
	})
}

func testAccUserPoliciesDataSourceConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = "%s"
}

resource "aws_iam_user_policy" "test" {
  user = aws_iam_user.test.name
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Deny"
        Action   = "*"
        Resource = "*"
      },
    ],
  })
}

data "aws_iam_user_policies" "test" {
  depends_on = [aws_iam_user_policy.test]

  user = aws_iam_user.test.name
}
`, name)
}

func testAccUserPoliciesDataSourceConfig_multiple(name, count string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = "%s"
}

resource "aws_iam_user_policy" "test" {
  count = "%s"

  user = aws_iam_user.test.name
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Deny"
        Action   = "*"
        Resource = "*"
      },
    ],
  })
}

data "aws_iam_user_policies" "test" {
  depends_on = [aws_iam_user_policy.test]

  user = aws_iam_user.test.name
}
`, name, count)
}
