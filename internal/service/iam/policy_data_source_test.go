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
		Arn        string
		Name       string
		PathPrefix string
		Expected   string
	}{
		{
			Arn:        "",
			Name:       "",
			PathPrefix: "",
			Expected:   "",
		},
		{
			Arn:        "arn:aws:iam::aws:policy/TestPolicy", //lintignore:AWSAT005
			Name:       "",
			PathPrefix: "",
			Expected:   "ARN: arn:aws:iam::aws:policy/TestPolicy", //lintignore:AWSAT005
		},
		{
			Arn:        "",
			Name:       "tf-acc-test-policy",
			PathPrefix: "",
			Expected:   "Name: tf-acc-test-policy",
		},
		{
			Arn:        "",
			Name:       "",
			PathPrefix: "/test-prefix/",
			Expected:   "PathPrefix: /test-prefix/",
		},
		{
			Arn:        "arn:aws:iam::aws:policy/TestPolicy", //lintignore:AWSAT005
			Name:       "tf-acc-test-policy",
			PathPrefix: "",
			Expected:   "ARN: arn:aws:iam::aws:policy/TestPolicy, Name: tf-acc-test-policy", //lintignore:AWSAT005
		},
		{
			Arn:        "arn:aws:iam::aws:policy/TestPolicy", //lintignore:AWSAT005
			Name:       "",
			PathPrefix: "/test-prefix/",
			Expected:   "ARN: arn:aws:iam::aws:policy/TestPolicy, PathPrefix: /test-prefix/", //lintignore:AWSAT005
		},
		{
			Arn:        "",
			Name:       "tf-acc-test-policy",
			PathPrefix: "/test-prefix/",
			Expected:   "Name: tf-acc-test-policy, PathPrefix: /test-prefix/",
		},
		{
			Arn:        "arn:aws:iam::aws:policy/TestPolicy", //lintignore:AWSAT005
			Name:       "tf-acc-test-policy",
			PathPrefix: "/test-prefix/",
			Expected:   "ARN: arn:aws:iam::aws:policy/TestPolicy, Name: tf-acc-test-policy, PathPrefix: /test-prefix/", //lintignore:AWSAT005
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			got := tfiam.PolicySearchDetails(testCase.Arn, testCase.Name, testCase.PathPrefix)

			if got != testCase.Expected {
				t.Errorf("got %s, expected %s", got, testCase.Expected)
			}
		})
	}
}

func TestAccIAMPolicyDataSource_arn(t *testing.T) {
	datasourceName := "data.aws_iam_policy.test"
	resourceName := "aws_iam_policy.test"
	policyName := fmt.Sprintf("test-policy-%s", sdkacctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDataSourceConfig_ARN(policyName, "/"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasourceName, "path", resourceName, "path"),
					resource.TestCheckResourceAttrPair(datasourceName, "policy", resourceName, "policy"),
					resource.TestCheckResourceAttrPair(datasourceName, "policy_id", resourceName, "policy_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags", resourceName, "tags"),
				),
			},
		},
	})
}

func TestAccIAMPolicyDataSource_name(t *testing.T) {
	datasourceName := "data.aws_iam_policy.test"
	resourceName := "aws_iam_policy.test"
	policyName := fmt.Sprintf("test-policy-%s", sdkacctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDataSourceConfig_Name(policyName, "/"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasourceName, "path", resourceName, "path"),
					resource.TestCheckResourceAttrPair(datasourceName, "policy", resourceName, "policy"),
					resource.TestCheckResourceAttrPair(datasourceName, "policy_id", resourceName, "policy_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags", resourceName, "tags"),
				),
			},
		},
	})
}

func TestAccIAMPolicyDataSource_nameAndPathPrefix(t *testing.T) {
	datasourceName := "data.aws_iam_policy.test"
	resourceName := "aws_iam_policy.test"

	policyName := fmt.Sprintf("test-policy-%s", sdkacctest.RandString(10))
	policyPath := "/test-path/"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDataSourceConfig_PathPrefix(policyName, policyPath),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasourceName, "path", resourceName, "path"),
					resource.TestCheckResourceAttrPair(datasourceName, "policy", resourceName, "policy"),
					resource.TestCheckResourceAttrPair(datasourceName, "policy_id", resourceName, "policy_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags", resourceName, "tags"),
				),
			},
		},
	})
}

func TestAccIAMPolicyDataSource_nonExistent(t *testing.T) {
	policyName := fmt.Sprintf("test-policy-%s", sdkacctest.RandString(10))
	policyPath := "/test-path/"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccPolicyDataSourceConfig_NonExistent(policyName, policyPath),
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

func testAccPolicyDataSourceConfig_ARN(policyName, policyPath string) string {
	return acctest.ConfigCompose(
		testAccPolicyBaseDataSourceConfig(policyName, policyPath),
		`
data "aws_iam_policy" "test" {
  arn = aws_iam_policy.test.arn
}
`)
}

func testAccPolicyDataSourceConfig_Name(policyName, policyPath string) string {
	return acctest.ConfigCompose(
		testAccPolicyBaseDataSourceConfig(policyName, policyPath),
		`
data "aws_iam_policy" "test" {
  name = aws_iam_policy.test.name
}
`)
}

func testAccPolicyDataSourceConfig_PathPrefix(policyName, policyPath string) string {
	return acctest.ConfigCompose(
		testAccPolicyBaseDataSourceConfig(policyName, policyPath),
		fmt.Sprintf(`
data "aws_iam_policy" "test" {
  name        = aws_iam_policy.test.name
  path_prefix = %q
}
`, policyPath))
}

func testAccPolicyDataSourceConfig_NonExistent(policyName, policyPath string) string {
	return acctest.ConfigCompose(
		testAccPolicyBaseDataSourceConfig(policyName, policyPath),
		fmt.Sprintf(`
data "aws_iam_policy" "test" {
  name        = "non-existent"
  path_prefix = %q
}
`, policyPath))
}
