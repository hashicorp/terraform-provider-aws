package iam_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/iam"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
)

func TestPolicySearchDetails(t *testing.T) {
	testCases := []struct {
		Name       string
		PathPrefix string
		Expected   string
	}{
		{
			Name:       "",
			PathPrefix: "",
			Expected:   "",
		},
		{
			Name:       "tf-acc-test-policy",
			PathPrefix: "",
			Expected:   "Name: tf-acc-test-policy",
		},
		{
			Name:       "",
			PathPrefix: "/test-prefix/",
			Expected:   "PathPrefix: /test-prefix/",
		},
		{
			Name:       "tf-acc-test-policy",
			PathPrefix: "/test-prefix/",
			Expected:   "Name: tf-acc-test-policy, PathPrefix: /test-prefix/",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			got := tfiam.PolicySearchDetails(testCase.Name, testCase.PathPrefix)

			if got != testCase.Expected {
				t.Errorf("got %s, expected %s", got, testCase.Expected)
			}
		})
	}
}

func TestAccIAMPolicyDataSource_arn(t *testing.T) {
	datasourceName := "data.aws_iam_policy.test"
	resourceName := "aws_iam_policy.test"
	policyName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDataSourceConfig_arn(policyName, "/"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasourceName, "path", resourceName, "path"),
					resource.TestCheckResourceAttrPair(datasourceName, "policy", resourceName, "policy"),
					resource.TestCheckResourceAttrPair(datasourceName, "policy_id", resourceName, "policy_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
				),
			},
		},
	})
}

func TestAccIAMPolicyDataSource_arnTags(t *testing.T) {
	datasourceName := "data.aws_iam_policy.test"
	resourceName := "aws_iam_policy.test"
	policyName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDataSourceConfig_arnTags(policyName, "/"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasourceName, "path", resourceName, "path"),
					resource.TestCheckResourceAttrPair(datasourceName, "policy", resourceName, "policy"),
					resource.TestCheckResourceAttrPair(datasourceName, "policy_id", resourceName, "policy_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.key", resourceName, "tags.key"),
				),
			},
		},
	})
}

func TestAccIAMPolicyDataSource_name(t *testing.T) {
	datasourceName := "data.aws_iam_policy.test"
	resourceName := "aws_iam_policy.test"
	policyName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDataSourceConfig_name(policyName, "/"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasourceName, "path", resourceName, "path"),
					resource.TestCheckResourceAttrPair(datasourceName, "policy", resourceName, "policy"),
					resource.TestCheckResourceAttrPair(datasourceName, "policy_id", resourceName, "policy_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
				),
			},
		},
	})
}

func TestAccIAMPolicyDataSource_nameTags(t *testing.T) {
	datasourceName := "data.aws_iam_policy.test"
	resourceName := "aws_iam_policy.test"
	policyName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDataSourceConfig_nameTags(policyName, "/"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasourceName, "path", resourceName, "path"),
					resource.TestCheckResourceAttrPair(datasourceName, "policy", resourceName, "policy"),
					resource.TestCheckResourceAttrPair(datasourceName, "policy_id", resourceName, "policy_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.key", resourceName, "tags.key"),
				),
			},
		},
	})
}

func TestAccIAMPolicyDataSource_nameAndPathPrefix(t *testing.T) {
	datasourceName := "data.aws_iam_policy.test"
	resourceName := "aws_iam_policy.test"

	policyName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyPath := "/test-path/"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDataSourceConfig_pathPrefix(policyName, policyPath),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasourceName, "path", resourceName, "path"),
					resource.TestCheckResourceAttrPair(datasourceName, "policy", resourceName, "policy"),
					resource.TestCheckResourceAttrPair(datasourceName, "policy_id", resourceName, "policy_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
				),
			},
		},
	})
}

func TestAccIAMPolicyDataSource_nameAndPathPrefixTags(t *testing.T) {
	datasourceName := "data.aws_iam_policy.test"
	resourceName := "aws_iam_policy.test"

	policyName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyPath := "/test-path/"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDataSourceConfig_pathPrefixTags(policyName, policyPath),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasourceName, "path", resourceName, "path"),
					resource.TestCheckResourceAttrPair(datasourceName, "policy", resourceName, "policy"),
					resource.TestCheckResourceAttrPair(datasourceName, "policy_id", resourceName, "policy_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.key", resourceName, "tags.key"),
				),
			},
		},
	})
}

func TestAccIAMPolicyDataSource_nonExistent(t *testing.T) {
	policyName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyPath := "/test-path/"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccPolicyDataSourceConfig_nonExistent(policyName, policyPath),
				ExpectError: regexp.MustCompile(`no IAM policy found matching criteria`),
			},
		},
	})
}

func testAccPolicyBaseDataSourceConfig(policyName, policyPath string) string {
	return fmt.Sprintf(`
resource "aws_iam_policy" "test" {
  name        = %q
  path        = %q
  description = "My test policy"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}`, policyName, policyPath)
}

func testAccPolicyBaseDataSourceTagsConfig(policyName, policyPath string) string {
	return fmt.Sprintf(`
resource "aws_iam_policy" "test" {
  name        = %[1]q
  path        = %[2]q
  description = "My test policy"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF

  tags = {
    "key" = "value"
  }
}`, policyName, policyPath)
}

func testAccPolicyDataSourceConfig_arn(policyName, policyPath string) string {
	return acctest.ConfigCompose(
		testAccPolicyBaseDataSourceConfig(policyName, policyPath), `
data "aws_iam_policy" "test" {
  arn = aws_iam_policy.test.arn
}
`)
}

func testAccPolicyDataSourceConfig_arnTags(policyName, policyPath string) string {
	return acctest.ConfigCompose(
		testAccPolicyBaseDataSourceTagsConfig(policyName, policyPath), `
data "aws_iam_policy" "test" {
  arn = aws_iam_policy.test.arn
}
`)
}

func testAccPolicyDataSourceConfig_name(policyName, policyPath string) string {
	return acctest.ConfigCompose(
		testAccPolicyBaseDataSourceConfig(policyName, policyPath), `
data "aws_iam_policy" "test" {
  name = aws_iam_policy.test.name
}
`)
}

func testAccPolicyDataSourceConfig_nameTags(policyName, policyPath string) string {
	return acctest.ConfigCompose(
		testAccPolicyBaseDataSourceTagsConfig(policyName, policyPath), `
data "aws_iam_policy" "test" {
  name = aws_iam_policy.test.name
}
`)
}

func testAccPolicyDataSourceConfig_pathPrefix(policyName, policyPath string) string {
	return acctest.ConfigCompose(
		testAccPolicyBaseDataSourceConfig(policyName, policyPath),
		fmt.Sprintf(`
data "aws_iam_policy" "test" {
  name        = aws_iam_policy.test.name
  path_prefix = %q
}
`, policyPath))
}

func testAccPolicyDataSourceConfig_pathPrefixTags(policyName, policyPath string) string {
	return acctest.ConfigCompose(
		testAccPolicyBaseDataSourceTagsConfig(policyName, policyPath),
		fmt.Sprintf(`
data "aws_iam_policy" "test" {
  name        = aws_iam_policy.test.name
  path_prefix = %q
}
`, policyPath))
}

func testAccPolicyDataSourceConfig_nonExistent(policyName, policyPath string) string {
	return acctest.ConfigCompose(
		testAccPolicyBaseDataSourceConfig(policyName, policyPath),
		fmt.Sprintf(`
data "aws_iam_policy" "test" {
  name        = "non-existent"
  path_prefix = %q
}
`, policyPath))
}
