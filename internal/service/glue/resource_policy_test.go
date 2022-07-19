package glue_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfglue "github.com/hashicorp/terraform-provider-aws/internal/service/glue"
)

func testAccResourcePolicy_basic(t *testing.T) {
	resourceName := "aws_glue_resource_policy.test"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourcePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_required("glue:CreateTable"),
				Check: resource.ComposeTestCheckFunc(
					testAccResourcePolicy(resourceName, "glue:CreateTable"),
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

func testAccResourcePolicy_hybrid(t *testing.T) {
	resourceName := "aws_glue_resource_policy.test"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourcePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_hybrid("glue:CreateTable", "TRUE"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enable_hybrid", "TRUE"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"enable_hybrid"},
			},
			{
				Config: testAccResourcePolicyConfig_hybrid("glue:CreateTable", "FALSE"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enable_hybrid", "FALSE"),
				),
			},
			{
				Config: testAccResourcePolicyConfig_hybrid("glue:CreateTable", "TRUE"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enable_hybrid", "TRUE"),
				),
			},
		},
	})
}
func testAccResourcePolicy_disappears(t *testing.T) {
	resourceName := "aws_glue_resource_policy.test"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourcePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_required("glue:CreateTable"),
				Check: resource.ComposeTestCheckFunc(
					testAccResourcePolicy(resourceName, "glue:CreateTable"),
					acctest.CheckResourceDisappears(acctest.Provider, tfglue.ResourceResourcePolicy(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfglue.ResourceResourcePolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccResourcePolicy_update(t *testing.T) {
	resourceName := "aws_glue_resource_policy.test"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourcePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_required("glue:CreateTable"),
				Check: resource.ComposeTestCheckFunc(
					testAccResourcePolicy(resourceName, "glue:CreateTable"),
				),
			},
			{
				Config: testAccResourcePolicyConfig_required("glue:DeleteTable"),
				Check: resource.ComposeTestCheckFunc(
					testAccResourcePolicy(resourceName, "glue:DeleteTable"),
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

func testAccResourcePolicy_ignoreEquivalent(t *testing.T) {
	resourceName := "aws_glue_resource_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourcePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_equivalent(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourcePolicy(resourceName, "glue:CreateTable"),
				),
			},
			{
				Config:   testAccResourcePolicyConfig_equivalent2(),
				PlanOnly: true,
			},
		},
	})
}

func testAccResourcePolicy(n string, action string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No policy id set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueConn

		policy, err := conn.GetResourcePolicy(&glue.GetResourcePolicyInput{})
		if err != nil {
			return fmt.Errorf("Get resource policy error: %v", err)
		}

		actualPolicyText := aws.StringValue(policy.PolicyInJson)

		expectedPolicy := CreateTablePolicy(action)
		equivalent, err := awspolicy.PoliciesAreEquivalent(actualPolicyText, expectedPolicy)
		if err != nil {
			return fmt.Errorf("Error testing policy equivalence: %s", err)
		}
		if !equivalent {
			return fmt.Errorf("Non-equivalent policy error:\n\nexpected: %s\n\n     got: %s\n",
				expectedPolicy, actualPolicyText)
		}

		return nil
	}
}

func testAccCheckResourcePolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).GlueConn

	policy, err := conn.GetResourcePolicy(&glue.GetResourcePolicyInput{})

	if err != nil {
		if tfawserr.ErrMessageContains(err, glue.ErrCodeEntityNotFoundException, "Policy not found") {
			return nil
		}
		return err
	}

	if *policy.PolicyInJson != "" {
		return fmt.Errorf("Aws glue resource policy still exists: %s", *policy.PolicyInJson)
	}
	return nil
}

func CreateTablePolicy(action string) string {
	return fmt.Sprintf(`{
  "Version" : "2012-10-17",
  "Statement" : [
    {
      "Effect" : "Allow",
      "Action" : [
        "%s"
      ],
      "Principal" : {
         "AWS": "*"
       },
      "Resource" : "arn:%s:glue:%s:%s:*"
    }
  ]
}`, action, acctest.Partition(), acctest.Region(), acctest.AccountID())
}

func testAccResourcePolicyConfig_required(action string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {}

data "aws_iam_policy_document" "glue-example-policy" {
  statement {
    actions   = [%[1]q]
    resources = ["arn:${data.aws_partition.current.partition}:glue:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:*"]
    principals {
      identifiers = ["*"]
      type        = "AWS"
    }
  }
}

resource "aws_glue_resource_policy" "test" {
  policy = data.aws_iam_policy_document.glue-example-policy.json
}
`, action)
}

func testAccResourcePolicyConfig_hybrid(action, hybrid string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {}

data "aws_iam_policy_document" "glue-example-policy" {
  statement {
    actions   = [%[1]q]
    resources = ["arn:${data.aws_partition.current.partition}:glue:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:*"]
    principals {
      identifiers = ["*"]
      type        = "AWS"
    }
  }
}

resource "aws_glue_resource_policy" "test" {
  policy        = data.aws_iam_policy_document.glue-example-policy.json
  enable_hybrid = %[2]q
}
`, action, hybrid)
}

func testAccResourcePolicyConfig_equivalent() string {
	return `
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_glue_resource_policy" "test" {
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = {
      Action = "glue:CreateTable"
      Effect = "Allow"
      Resource = [
        "arn:${data.aws_partition.current.partition}:glue:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:*"
      ]
      Principal = {
        AWS = "*"
      }
    }
  })
}
`
}

func testAccResourcePolicyConfig_equivalent2() string {
	return `
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_glue_resource_policy" "test" {
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = {
      Effect = "Allow"
      Action = [
        "glue:CreateTable",
      ]
      Resource = "arn:${data.aws_partition.current.partition}:glue:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:*"
      Principal = {
        AWS = ["*"]
      }
    }
  })
}
`
}
