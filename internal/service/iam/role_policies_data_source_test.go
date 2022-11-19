package iam_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/iam"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccIAMRolePoliciesDataSource_basic(t *testing.T) {
	resourceName := "aws_iam_role_policy.test"
	dataSourceName := "data.aws_iam_role_policies.test"

	role := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, iam.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRolePoliciesDataSourceConfig_basic(role),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "role", resourceName, "role"),
					resource.TestCheckResourceAttr(dataSourceName, "names.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "names.0", resourceName, "name"),
				),
			},
		},
	})
}

func TestAccIAMRolePoliciesDataSource_multiple(t *testing.T) {
	resourceName0 := "aws_iam_role_policy.test.0"
	resourceName1 := "aws_iam_role_policy.test.1"
	dataSourceName := "data.aws_iam_role_policies.test"

	role := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	count := "2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, iam.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRolePoliciesDataSourceConfig_multiple(role, count),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "role", resourceName0, "role"),
					resource.TestCheckResourceAttrPair(dataSourceName, "role", resourceName1, "role"),
					resource.TestCheckResourceAttr(dataSourceName, "names.#", count),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "names.*", resourceName0, "name"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "names.*", resourceName1, "name"),
				),
			},
		},
	})
}

func testAccRolePoliciesDataSourceConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = "%s"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.name
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

data "aws_iam_role_policies" "test" {
  depends_on = [aws_iam_role_policy.test]

  role = aws_iam_role.test.name
}
`, name)
}

func testAccRolePoliciesDataSourceConfig_multiple(name, count string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = "%s"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_iam_role_policy" "test" {
  count = "%s"

  role = aws_iam_role.test.name
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

data "aws_iam_role_policies" "test" {
  depends_on = [aws_iam_role_policy.test]

  role = aws_iam_role.test.name
}
`, name, count)
}
