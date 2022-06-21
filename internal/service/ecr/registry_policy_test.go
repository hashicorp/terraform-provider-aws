package ecr_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfecr "github.com/hashicorp/terraform-provider-aws/internal/service/ecr"
)

func TestAccECRRegistryPolicy_serial(t *testing.T) {
	testFuncs := map[string]func(t *testing.T){
		"basic":      testAccRegistryPolicy_basic,
		"disappears": testAccRegistryPolicy_disappears,
	}

	for name, testFunc := range testFuncs {
		testFunc := testFunc

		t.Run(name, func(t *testing.T) {
			testFunc(t)
		})
	}
}

func testAccRegistryPolicy_basic(t *testing.T) {
	var v ecr.GetRegistryPolicyOutput
	resourceName := "aws_ecr_registry_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRegistryPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryPolicyConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegistryPolicyExists(resourceName, &v),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(`"ecr:ReplicateImage".+`)),
					acctest.CheckResourceAttrAccountID(resourceName, "registry_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRegistryPolicyConfig_updated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegistryPolicyExists(resourceName, &v),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(`"ecr:ReplicateImage".+`)),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(`"ecr:CreateRepository".+`)),
					acctest.CheckResourceAttrAccountID(resourceName, "registry_id"),
				),
			},
		},
	})
}

func testAccRegistryPolicy_disappears(t *testing.T) {
	var v ecr.GetRegistryPolicyOutput
	resourceName := "aws_ecr_registry_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRegistryPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryPolicyConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegistryPolicyExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfecr.ResourceRegistryPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRegistryPolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ECRConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ecr_registry_policy" {
			continue
		}

		_, err := conn.GetRegistryPolicy(&ecr.GetRegistryPolicyInput{})
		if err != nil {
			if tfawserr.ErrCodeEquals(err, ecr.ErrCodeRegistryPolicyNotFoundException) {
				return nil
			}
			return err
		}
	}

	return nil
}

func testAccCheckRegistryPolicyExists(name string, res *ecr.GetRegistryPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ECR registry policy ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ECRConn

		output, err := conn.GetRegistryPolicy(&ecr.GetRegistryPolicyInput{})
		if err != nil {
			if tfawserr.ErrCodeEquals(err, ecr.ErrCodeRegistryPolicyNotFoundException) {
				return fmt.Errorf("ECR repository %s not found", rs.Primary.ID)
			}
			return err
		}

		*res = *output

		return nil
	}
}

func testAccRegistryPolicyConfig_basic() string {
	return `
data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

data "aws_partition" "current" {}

resource "aws_ecr_registry_policy" "test" {
  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Sid" : "testpolicy",
        "Effect" : "Allow",
        "Principal" : {
          "AWS" : "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
        },
        "Action" : "ecr:ReplicateImage",
        "Resource" : "arn:${data.aws_partition.current.partition}:ecr:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:repository/*",
      }
    ]
  })
}
`
}

func testAccRegistryPolicyConfig_updated() string {
	return `
data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

data "aws_partition" "current" {}

resource "aws_ecr_registry_policy" "test" {
  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Sid" : "testpolicy",
        "Effect" : "Allow",
        "Principal" : {
          "AWS" : "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
        },
        "Action" : [
          "ecr:ReplicateImage",
          "ecr:CreateRepository"
        ],
        "Resource" : [
          "arn:${data.aws_partition.current.partition}:ecr:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:repository/*"
        ]
      }
    ]
  })
}
`
}
