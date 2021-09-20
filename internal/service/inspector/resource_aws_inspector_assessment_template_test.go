package inspector_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/inspector"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func TestAccAWSInspectorTemplate_basic(t *testing.T) {
	var v inspector.AssessmentTemplate
	resourceName := "aws_inspector_assessment_template.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, inspector.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSInspectorTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSInspectorTemplateAssessmentBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSInspectorTemplateExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "inspector", regexp.MustCompile(`target/.+/template/.+`)),
					resource.TestCheckResourceAttr(resourceName, "duration", "3600"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "rules_package_arns.#", "data.aws_inspector_rules_packages.available", "arns.#"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "target_arn", "aws_inspector_assessment_target.test", "arn"),
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

func TestAccAWSInspectorTemplate_disappears(t *testing.T) {
	var v inspector.AssessmentTemplate
	resourceName := "aws_inspector_assessment_template.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, inspector.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSInspectorTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSInspectorTemplateAssessmentBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSInspectorTemplateExists(resourceName, &v),
					testAccCheckAWSInspectorTemplateDisappears(&v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSInspectorTemplate_tags(t *testing.T) {
	var v inspector.AssessmentTemplate
	resourceName := "aws_inspector_assessment_template.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, inspector.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSInspectorTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSInspectorTemplateAssessmentTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSInspectorTemplateExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSInspectorTemplateAssessmentTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSInspectorTemplateExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSInspectorTemplateAssessmentTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSInspectorTemplateExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSInspectorTemplateAssessmentBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSInspectorTemplateExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccCheckAWSInspectorTemplateDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).InspectorConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_inspector_assessment_template" {
			continue
		}

		resp, err := conn.DescribeAssessmentTemplates(&inspector.DescribeAssessmentTemplatesInput{
			AssessmentTemplateArns: []*string{
				aws.String(rs.Primary.ID),
			},
		})

		if err != nil {
			if inspectorerr, ok := err.(awserr.Error); ok && inspectorerr.Code() == "InvalidInputException" {
				return nil
			} else {
				return fmt.Errorf("Error finding Inspector Assessment Template: %s", err)
			}
		}

		if len(resp.AssessmentTemplates) > 0 {
			return fmt.Errorf("Found Template, expected none: %s", resp)
		}
	}

	return nil
}

func testAccCheckAWSInspectorTemplateDisappears(v *inspector.AssessmentTemplate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).InspectorConn

		_, err := conn.DeleteAssessmentTemplate(&inspector.DeleteAssessmentTemplateInput{
			AssessmentTemplateArn: v.Arn,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckAWSInspectorTemplateExists(name string, v *inspector.AssessmentTemplate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Inspector assessment template ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).InspectorConn

		resp, err := conn.DescribeAssessmentTemplates(&inspector.DescribeAssessmentTemplatesInput{
			AssessmentTemplateArns: aws.StringSlice([]string{rs.Primary.ID}),
		})
		if err != nil {
			return err
		}

		if resp.AssessmentTemplates == nil || len(resp.AssessmentTemplates) == 0 {
			return fmt.Errorf("Inspector assessment template not found")
		}

		*v = *resp.AssessmentTemplates[0]

		return nil
	}
}

func testAccAWSInspectorTemplateAssessmentBase(rName string) string {
	return fmt.Sprintf(`
data "aws_inspector_rules_packages" "available" {}

resource "aws_inspector_resource_group" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_inspector_assessment_target" "test" {
  name               = %[1]q
  resource_group_arn = aws_inspector_resource_group.test.arn
}
`, rName)
}

func testAccAWSInspectorTemplateAssessmentBasic(rName string) string {
	return testAccAWSInspectorTemplateAssessmentBase(rName) + fmt.Sprintf(`
resource "aws_inspector_assessment_template" "test" {
  name       = %[1]q
  target_arn = aws_inspector_assessment_target.test.arn
  duration   = 3600

  rules_package_arns = data.aws_inspector_rules_packages.available.arns
}
`, rName)
}

func testAccAWSInspectorTemplateAssessmentTags1(rName, tagKey1, tagValue1 string) string {
	return testAccAWSInspectorTemplateAssessmentBase(rName) + fmt.Sprintf(`
resource "aws_inspector_assessment_template" "test" {
  name       = %[1]q
  target_arn = aws_inspector_assessment_target.test.arn
  duration   = 3600

  rules_package_arns = data.aws_inspector_rules_packages.available.arns

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSInspectorTemplateAssessmentTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAWSInspectorTemplateAssessmentBase(rName) + fmt.Sprintf(`
resource "aws_inspector_assessment_template" "test" {
  name       = %[1]q
  target_arn = aws_inspector_assessment_target.test.arn
  duration   = 3600

  rules_package_arns = data.aws_inspector_rules_packages.available.arns

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
