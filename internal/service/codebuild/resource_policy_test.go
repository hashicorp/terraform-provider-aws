package codebuild_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/codebuild"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcodebuild "github.com/hashicorp/terraform-provider-aws/internal/service/codebuild"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccCodeBuildResourcePolicy_basic(t *testing.T) {
	var reportGroup codebuild.GetResourcePolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_resource_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourcePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourcePolicyExists(resourceName, &reportGroup),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", "aws_codebuild_report_group.test", "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "policy"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCodeBuildResourcePolicy_disappears(t *testing.T) {
	var reportGroup codebuild.GetResourcePolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_resource_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourcePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourcePolicyExists(resourceName, &reportGroup),
					acctest.CheckResourceDisappears(acctest.Provider, tfcodebuild.ResourceResourcePolicy(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfcodebuild.ResourceResourcePolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCodeBuildResourcePolicy_disappears_resource(t *testing.T) {
	var reportGroup codebuild.GetResourcePolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_resource_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourcePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourcePolicyExists(resourceName, &reportGroup),
					acctest.CheckResourceDisappears(acctest.Provider, tfcodebuild.ResourceReportGroup(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfcodebuild.ResourceResourcePolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckResourcePolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CodeBuildConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_codebuild_resource_policy" {
			continue
		}

		resp, err := tfcodebuild.FindResourcePolicyByARN(conn, rs.Primary.ID)
		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		if resp != nil {
			return fmt.Errorf("Found Resource Policy %s", rs.Primary.ID)
		}

	}
	return nil
}

func testAccCheckResourcePolicyExists(name string, policy *codebuild.GetResourcePolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeBuildConn

		resp, err := tfcodebuild.FindResourcePolicyByARN(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if resp == nil {
			return fmt.Errorf("Resource Policy %s not found", rs.Primary.ID)
		}

		*policy = *resp

		return nil
	}
}

func testAccResourcePolicyBasicConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_codebuild_report_group" "test" {
  name = %[1]q
  type = "TEST"

  export_config {
    type = "NO_EXPORT"
  }
}

data "aws_partition" "current" {}

data "aws_caller_identity" "current" {}

resource "aws_codebuild_resource_policy" "test" {
  resource_arn = aws_codebuild_report_group.test.arn
  policy = jsonencode({
    Version = "2012-10-17"
    Id      = "default"
    Statement = [{
      Sid    = "default"
      Effect = "Allow"
      Principal = {
        AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      }
      Action = [
        "codebuild:BatchGetReportGroups",
        "codebuild:BatchGetReports",
        "codebuild:ListReportsForReportGroup",
        "codebuild:DescribeTestCases",
      ]
      Resource = aws_codebuild_report_group.test.arn
    }]
  })
}
`, rName)
}
