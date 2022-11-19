package iam_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/iam"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccIAMAttachedGroupPoliciesDataSource_basic(t *testing.T) {
	resourceName := "aws_iam_group_policy_attachment.test"
	dataSourceName := "data.aws_iam_attached_group_policies.test"

	group := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, iam.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAttachedGroupPoliciesDataSourceConfig_basic(group, policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "group", resourceName, "group"),
					resource.TestCheckResourceAttr(dataSourceName, "path_prefix", "/"),
					resource.TestCheckResourceAttr(dataSourceName, "names.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "names.0", policyName),
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "arns.0", resourceName, "policy_arn"),
				),
			},
		},
	})
}

func TestAccIAMAttachedGroupPoliciesDataSource_withPathPrefixMatching(t *testing.T) {
	resourceName := "aws_iam_group_policy_attachment.test"
	dataSourceName := "data.aws_iam_attached_group_policies.test"

	group := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyPath := "/test/"
	pathPrefix := policyPath

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, iam.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAttachedGroupPoliciesDataSourceConfig_withPathPrefix(group, policyName, policyPath, pathPrefix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "group", resourceName, "group"),
					resource.TestCheckResourceAttr(dataSourceName, "path_prefix", pathPrefix),
					resource.TestCheckResourceAttr(dataSourceName, "names.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "names.0", policyName),
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "arns.0", resourceName, "policy_arn"),
				),
			},
		},
	})
}

func TestAccIAMAttachedGroupPoliciesDataSource_withPathPrefixNotMatching(t *testing.T) {
	resourceName := "aws_iam_group_policy_attachment.test"
	dataSourceName := "data.aws_iam_attached_group_policies.test"

	group := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyPath := "/test/"
	pathPrefix := "/different/"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, iam.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAttachedGroupPoliciesDataSourceConfig_withPathPrefix(group, policyName, policyPath, pathPrefix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "group", resourceName, "group"),
					resource.TestCheckResourceAttr(dataSourceName, "path_prefix", pathPrefix),
					resource.TestCheckResourceAttr(dataSourceName, "names.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", "0"),
				),
			},
		},
	})
}

func testAccAttachedGroupPoliciesDataSourceConfig_basic(name, policyName string) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "test" {
  name = "%s"
}

resource "aws_iam_policy" "test" {
  name = "%s"

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

resource "aws_iam_group_policy_attachment" "test" {
  group      = aws_iam_group.test.name
  policy_arn = aws_iam_policy.test.arn
}

data "aws_iam_attached_group_policies" "test" {
  depends_on = [aws_iam_group_policy_attachment.test]

  group = aws_iam_group.test.name
}
`, name, policyName)
}

func testAccAttachedGroupPoliciesDataSourceConfig_withPathPrefix(name, policyName, policyPath, pathPrefix string) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "test" {
  name = "%s"
}

resource "aws_iam_policy" "test" {
  name = "%s"
  path = "%s"

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

resource "aws_iam_group_policy_attachment" "test" {
  group      = aws_iam_group.test.name
  policy_arn = aws_iam_policy.test.arn
}

data "aws_iam_attached_group_policies" "test" {
  depends_on = [aws_iam_group_policy_attachment.test]

  group       = aws_iam_group.test.name
  path_prefix = "%s"
}
`, name, policyName, policyPath, pathPrefix)
}
