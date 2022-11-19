package iam_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/iam"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccIAMUserPolicyAttachmentsDataSource_basic(t *testing.T) {
	resourceName := "aws_iam_policy.test"
	dataSourceName := "data.aws_iam_user_policy_attachments.test"

	user := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, iam.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPolicyAttachmentsDataSourceConfig_basic(user),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "user", user),
					resource.TestCheckResourceAttr(dataSourceName, "path_prefix", "/"),
					resource.TestCheckResourceAttr(dataSourceName, "names.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "names.0", resourceName, "name"),
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "arns.0", resourceName, "arn"),
				),
			},
		},
	})
}

func TestAccIAMUserPolicyAttachmentsDataSource_multiple(t *testing.T) {
	rPolicy0 := "aws_iam_policy.test.0"
	rPolicy1 := "aws_iam_policy.test.1"
	dataSourceName := "data.aws_iam_user_policy_attachments.test"

	user := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	count := "2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, iam.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPolicyAttachmentsDataSourceConfig_multiple(user, count),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "user", user),
					resource.TestCheckResourceAttr(dataSourceName, "path_prefix", "/"),
					resource.TestCheckResourceAttr(dataSourceName, "names.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "names.*", rPolicy0, "name"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "names.*", rPolicy1, "name"),
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "arns.*", rPolicy0, "arn"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "arns.*", rPolicy1, "arn"),
				),
			},
		},
	})
}

func TestAccIAMUserPolicyAttachmentsDataSource_withPathPrefixMatching(t *testing.T) {
	resourceName := "aws_iam_policy.test"
	dataSourceName := "data.aws_iam_user_policy_attachments.test"

	user := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyPath := "/test/"
	pathPrefix := policyPath

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, iam.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPolicyAttachmentsDataSourceConfig_withPathPrefix(user, policyPath, pathPrefix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "user", user),
					resource.TestCheckResourceAttr(dataSourceName, "path_prefix", pathPrefix),
					resource.TestCheckResourceAttr(dataSourceName, "names.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "names.0", resourceName, "name"),
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "arns.0", resourceName, "arn"),
				),
			},
		},
	})
}

func TestAccIAMUserPolicyAttachmentsDataSource_withPathPrefixNotMatching(t *testing.T) {
	dataSourceName := "data.aws_iam_user_policy_attachments.test"

	user := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyPath := "/test/"
	pathPrefix := "/different/"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, iam.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPolicyAttachmentsDataSourceConfig_withPathPrefix(user, policyPath, pathPrefix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "user", user),
					resource.TestCheckResourceAttr(dataSourceName, "path_prefix", pathPrefix),
					resource.TestCheckResourceAttr(dataSourceName, "names.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", "0"),
				),
			},
		},
	})
}

func testAccUserPolicyAttachmentsDataSourceConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = "%s"
}

resource "aws_iam_policy" "test" {
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

resource "aws_iam_user_policy_attachment" "test" {
  user       = aws_iam_user.test.name
  policy_arn = aws_iam_policy.test.arn
}

data "aws_iam_user_policy_attachments" "test" {
  depends_on = [aws_iam_user_policy_attachment.test]

  user = aws_iam_user.test.name
}
`, name)
}

func testAccUserPolicyAttachmentsDataSourceConfig_multiple(name, count string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = "%s"
}

resource "aws_iam_policy" "test" {
  count = "%s"

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

resource "aws_iam_user_policy_attachment" "test" {
  count = length(aws_iam_policy.test)

  user       = aws_iam_user.test.name
  policy_arn = aws_iam_policy.test[count.index].arn
}

data "aws_iam_user_policy_attachments" "test" {
  depends_on = [aws_iam_user_policy_attachment.test]

  user = aws_iam_user.test.name
}
`, name, count)
}

func testAccUserPolicyAttachmentsDataSourceConfig_withPathPrefix(name, policyPath, pathPrefix string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = "%s"
}

resource "aws_iam_policy" "test" {
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

resource "aws_iam_user_policy_attachment" "test" {
  user       = aws_iam_user.test.name
  policy_arn = aws_iam_policy.test.arn
}

data "aws_iam_user_policy_attachments" "test" {
  depends_on = [aws_iam_user_policy_attachment.test]

  user        = aws_iam_user.test.name
  path_prefix = "%s"
}
`, name, policyPath, pathPrefix)
}
