package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sagemaker/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func TestAccAWSSagemakerModelPackageGroupPolicy_basic(t *testing.T) {
	var mpg sagemaker.GetModelPackageGroupPolicyOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_model_package_group_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sagemaker.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerModelPackageGroupPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerModelPackageGroupPolicyBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerModelPackageGroupPolicyExists(resourceName, &mpg),
					resource.TestCheckResourceAttr(resourceName, "model_package_group_name", rName),
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

func TestAccAWSSagemakerModelPackageGroupPolicy_disappears(t *testing.T) {
	var mpg sagemaker.GetModelPackageGroupPolicyOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_model_package_group_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sagemaker.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerModelPackageGroupPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerModelPackageGroupPolicyBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerModelPackageGroupPolicyExists(resourceName, &mpg),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSagemakerModelPackageGroupPolicy(), resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSagemakerModelPackageGroupPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSSagemakerModelPackageGroupPolicy_disappears_modelPackageGroup(t *testing.T) {
	var mpg sagemaker.GetModelPackageGroupPolicyOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_model_package_group_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sagemaker.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerModelPackageGroupPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerModelPackageGroupPolicyBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerModelPackageGroupPolicyExists(resourceName, &mpg),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSagemakerModelPackageGroup(), "aws_sagemaker_model_package_group.test"),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSagemakerModelPackageGroupPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSSagemakerModelPackageGroupPolicyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sagemakerconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_model_package_group_policy" {
			continue
		}

		_, err := finder.ModelPackageGroupPolicyByName(conn, rs.Primary.ID)
		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading Sagemaker Model Package Group Policy (%s): %w", rs.Primary.ID, err)
		}
	}

	return nil
}

func testAccCheckAWSSagemakerModelPackageGroupPolicyExists(n string, mpg *sagemaker.GetModelPackageGroupPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker Model Package Group ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sagemakerconn
		resp, err := finder.ModelPackageGroupPolicyByName(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*mpg = *resp

		return nil
	}
}

func testAccAWSSagemakerModelPackageGroupPolicyBasicConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    sid       = "AddPermModelPackageGroup"
    actions   = ["sagemaker:DescribeModelPackage", "sagemaker:ListModelPackages"]
    resources = [aws_sagemaker_model_package_group.test.arn]
    principals {
      identifiers = [data.aws_caller_identity.current.account_id]
      type        = "AWS"
    }
  }
}

resource "aws_sagemaker_model_package_group" "test" {
  model_package_group_name = %[1]q
}

resource "aws_sagemaker_model_package_group_policy" "test" {
  model_package_group_name = aws_sagemaker_model_package_group.test.model_package_group_name
  resource_policy          = jsonencode(jsondecode(data.aws_iam_policy_document.test.json))
}
`, rName)
}
