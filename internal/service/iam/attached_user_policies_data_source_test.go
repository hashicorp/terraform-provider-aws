package iam_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/iam"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccIAMAttachedUserPoliciesDataSource_basic(t *testing.T) {
	resourceName := "aws_iam_user_policy_attachment.test"
	dataSourceName := "data.aws_iam_attached_user_policies.test"

	userName := fmt.Sprintf("test-datasource-user-%d", sdkacctest.RandInt())
	policyName := "IAMReadOnlyAccess"
	policyArn := fmt.Sprintf("arn:aws:iam::aws:policy/%s", policyName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, iam.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMAttachedUserPoliciesDataSourceConfig_basic(userName, policyArn),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "user", resourceName, "user"),
					resource.TestCheckResourceAttr(dataSourceName, "path_prefix", "/"),
					resource.TestCheckResourceAttr(dataSourceName, "names.*", policyName),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "arns.*", resourceName, "policy_arn"),
				),
			},
		},
	})
}

func TestAccIAMAttachedUserPoliciesDataSource_withPathPrefixWithResults(t *testing.T) {
	resourceName := "aws_iam_user_policy_attachment.test"
	dataSourceName := "data.aws_iam_attached_user_policies.test"

	userName := fmt.Sprintf("test-datasource-user-%d", sdkacctest.RandInt())
	policyName := "DenyAll"
	policyPath := "/test"
	pathPrefix := policyPath

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, iam.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMAttachedUserPoliciesDataSourceConfig_withPathPrefix(userName, policyName, policyPath, pathPrefix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "user", resourceName, "user"),
					resource.TestCheckResourceAttr(dataSourceName, "path_prefix", pathPrefix),
					resource.TestCheckResourceAttr(dataSourceName, "policy_names.*", policyName),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "policy_arns.*", resourceName, "policy_arn"),
				),
			},
		},
	})
}

func TestAccIAMAttachedUserPoliciesDataSource_withPathPrefixWithoutResults(t *testing.T) {
	resourceName := "aws_iam_user_policy_attachment.test"
	dataSourceName := "data.aws_iam_attached_user_policies.test"

	userName := fmt.Sprintf("test-datasource-user-%d", sdkacctest.RandInt())
	policyName := "DenyAll"
	policyPath := "/test"
	pathPrefix := "/different"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, iam.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMAttachedUserPoliciesDataSourceConfig_withPathPrefix(userName, policyName, policyPath, pathPrefix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "user", resourceName, "user"),
					resource.TestCheckResourceAttr(dataSourceName, "path_prefix", pathPrefix),
					resource.TestCheckNoResourceAttr(dataSourceName, "names.{0}"),
					resource.TestCheckNoResourceAttr(dataSourceName, "arns.{0}"),
				),
			},
		},
	})
}

func testAccIAMAttachedUserPoliciesDataSourceConfig_basic(name, policyArn string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = "%s"
}

resource "aws_iam_user_policy_attachment" "test" {
  user       = aws_iam_user.test.name
  policy_arn = "%s"
}

data "aws_iam_attached_user_policies" "test" {
  user = aws_iam_user.test.name
}
`, name, policyArn)
}

func testAccIAMAttachedUserPoliciesDataSourceConfig_withPathPrefix(name, policyName, policyPath, pathPrefix string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = "%s"
}

resource "aws_iam_policy" "test" {
  name        = "%s"
  path        = "%s"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Deny"
        Action = "*"
        Resource = "*"
      },
    ],
  })
}

resource "aws_iam_user_policy_attachment" "test" {
  user       = aws_iam_user.test.name
  policy_arn = aws_iam_policy.test.arn
}

data "aws_iam_attached_user_policies" "test" {
  user = aws_iam_user.test.name
  path_prefix = "%s"
}
`, name, policyName, policyPath, pathPrefix)
}
