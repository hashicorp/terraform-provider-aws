package inspector_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/inspector"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccInspectorAssessmentTemplate_basic(t *testing.T) {
	var v inspector.AssessmentTemplate
	resourceName := "aws_inspector_assessment_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, inspector.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateAssessmentBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateExists(resourceName, &v),
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

func TestAccInspectorAssessmentTemplate_disappears(t *testing.T) {
	var v inspector.AssessmentTemplate
	resourceName := "aws_inspector_assessment_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, inspector.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateAssessmentBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateExists(resourceName, &v),
					testAccCheckTemplateDisappears(&v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccInspectorAssessmentTemplate_tags(t *testing.T) {
	var v inspector.AssessmentTemplate
	resourceName := "aws_inspector_assessment_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, inspector.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateAssessmentTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateExists(resourceName, &v),
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
				Config: testAccTemplateAssessmentTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccTemplateAssessmentTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccTemplateAssessmentBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccCheckTemplateDestroy(s *terraform.State) error {
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

		if tfawserr.ErrCodeEquals(err, inspector.ErrCodeInvalidInputException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("Error finding Inspector Assessment Template: %s", err)
		}

		if len(resp.AssessmentTemplates) > 0 {
			return fmt.Errorf("Found Template, expected none: %s", resp)
		}
	}

	return nil
}

func testAccCheckTemplateDisappears(v *inspector.AssessmentTemplate) resource.TestCheckFunc {
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

func testAccCheckTemplateExists(name string, v *inspector.AssessmentTemplate) resource.TestCheckFunc {
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

func testAccTemplateAssessmentBase(rName string) string {
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

func testAccTemplateAssessmentBasic(rName string) string {
	return testAccTemplateAssessmentBase(rName) + fmt.Sprintf(`
resource "aws_inspector_assessment_template" "test" {
  name       = %[1]q
  target_arn = aws_inspector_assessment_target.test.arn
  duration   = 3600

  rules_package_arns = data.aws_inspector_rules_packages.available.arns
}
`, rName)
}

func testAccTemplateAssessmentTags1(rName, tagKey1, tagValue1 string) string {
	return testAccTemplateAssessmentBase(rName) + fmt.Sprintf(`
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

func testAccTemplateAssessmentTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccTemplateAssessmentBase(rName) + fmt.Sprintf(`
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
