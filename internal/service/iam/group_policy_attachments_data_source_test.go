package iam_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/iam"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccIAMGroupPolicyAttachmentsDataSource_basic(t *testing.T) {
	resourceName := "aws_iam_policy.test"
	dataSourceName := "data.aws_iam_group_policy_attachments.test"

	group := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, iam.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupPolicyAttachmentsDataSourceConfig_basic(group),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "group", group),
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

func TestAccIAMGroupPolicyAttachmentsDataSource_multiple(t *testing.T) {
	rPolicy0 := "aws_iam_policy.test.0"
	rPolicy1 := "aws_iam_policy.test.1"
	dataSourceName := "data.aws_iam_group_policy_attachments.test"

	group := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	count := "2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, iam.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupPolicyAttachmentsDataSourceConfig_multiple(group, count),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "group", group),
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

func TestAccIAMGroupPolicyAttachmentsDataSource_withPathPrefixMatching(t *testing.T) {
	resourceName := "aws_iam_policy.test"
	dataSourceName := "data.aws_iam_group_policy_attachments.test"

	group := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyPath := "/test/"
	pathPrefix := policyPath

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, iam.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupPolicyAttachmentsDataSourceConfig_withPathPrefix(group, policyPath, pathPrefix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "group", group),
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

func TestAccIAMGroupPolicyAttachmentsDataSource_withPathPrefixNotMatching(t *testing.T) {
	dataSourceName := "data.aws_iam_group_policy_attachments.test"

	group := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyPath := "/test/"
	pathPrefix := "/different/"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, iam.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupPolicyAttachmentsDataSourceConfig_withPathPrefix(group, policyPath, pathPrefix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "group", group),
					resource.TestCheckResourceAttr(dataSourceName, "path_prefix", pathPrefix),
					resource.TestCheckResourceAttr(dataSourceName, "names.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", "0"),
				),
			},
		},
	})
}

func testAccGroupPolicyAttachmentsDataSourceConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "test" {
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

resource "aws_iam_group_policy_attachment" "test" {
  group      = aws_iam_group.test.name
  policy_arn = aws_iam_policy.test.arn
}

data "aws_iam_group_policy_attachments" "test" {
  depends_on = [aws_iam_group_policy_attachment.test]

  group = aws_iam_group.test.name
}
`, name)
}

func testAccGroupPolicyAttachmentsDataSourceConfig_multiple(name, count string) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "test" {
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

resource "aws_iam_group_policy_attachment" "test" {
  count = length(aws_iam_policy.test)

  group      = aws_iam_group.test.name
  policy_arn = aws_iam_policy.test[count.index].arn
}

data "aws_iam_group_policy_attachments" "test" {
  depends_on = [aws_iam_group_policy_attachment.test]

  group = aws_iam_group.test.name
}
`, name, count)
}

func testAccGroupPolicyAttachmentsDataSourceConfig_withPathPrefix(name, policyPath, pathPrefix string) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "test" {
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

resource "aws_iam_group_policy_attachment" "test" {
  group      = aws_iam_group.test.name
  policy_arn = aws_iam_policy.test.arn
}

data "aws_iam_group_policy_attachments" "test" {
  depends_on = [aws_iam_group_policy_attachment.test]

  group       = aws_iam_group.test.name
  path_prefix = "%s"
}
`, name, policyPath, pathPrefix)
}
