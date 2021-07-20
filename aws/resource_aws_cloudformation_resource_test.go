package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAwsCloudformationResource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudformation_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudformationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudformationResourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "resource_model", regexp.MustCompile(`^\{.*\}$`)),
					resource.TestMatchResourceAttr(resourceName, "schema", regexp.MustCompile(`^\{.*`)),
				),
			},
		},
	})
}

func TestAccAwsCloudformationResource_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudformation_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudformationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudformationResourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCloudFormationResource(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsCloudformationResource_DesiredState_BooleanValueAdded(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudformation_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudformationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudformationResourceConfigDesiredStateBooleanValueRemoved(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "resource_model", regexp.MustCompile(`"Enabled":false`)),
				),
			},
			{
				Config: testAccAwsCloudformationResourceConfigDesiredStateBooleanValue(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "resource_model", regexp.MustCompile(`"Enabled":true`)),
				),
			},
		},
	})
}

func TestAccAwsCloudformationResource_DesiredState_BooleanValueRemoved(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudformation_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudformationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudformationResourceConfigDesiredStateBooleanValue(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "resource_model", regexp.MustCompile(`"Enabled":true`)),
				),
			},
			{
				Config: testAccAwsCloudformationResourceConfigDesiredStateBooleanValueRemoved(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "resource_model", regexp.MustCompile(`"Enabled":false`)),
				),
			},
		},
	})
}

func TestAccAwsCloudformationResource_DesiredState_BooleanValueUpdate(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudformation_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudformationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudformationResourceConfigDesiredStateBooleanValue(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "resource_model", regexp.MustCompile(`"Enabled":true`)),
				),
			},
			{
				Config: testAccAwsCloudformationResourceConfigDesiredStateBooleanValue(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "resource_model", regexp.MustCompile(`"Enabled":false`)),
				),
			},
		},
	})
}

func TestAccAwsCloudformationResource_DesiredState_CreateOnly(t *testing.T) {
	rName1 := acctest.RandomWithPrefix("tf-acc-test")
	rName2 := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudformation_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudformationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudformationResourceConfigDesiredStateCreateOnly(rName1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "resource_model", regexp.MustCompile(`"LogGroupName":"`+rName1+`"`)),
				),
			},
			{
				Config: testAccAwsCloudformationResourceConfigDesiredStateCreateOnly(rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "resource_model", regexp.MustCompile(`"LogGroupName":"`+rName2+`"`)),
				),
			},
		},
	})
}

func TestAccAwsCloudformationResource_DesiredState_IntegerValueAdded(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudformation_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudformationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudformationResourceConfigDesiredStateIntegerValueRemoved(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "resource_model", regexp.MustCompile(`"LogGroupName":`)),
				),
			},
			{
				Config: testAccAwsCloudformationResourceConfigDesiredStateIntegerValue(rName, 14),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "resource_model", regexp.MustCompile(`"RetentionInDays":14`)),
				),
			},
		},
	})
}

func TestAccAwsCloudformationResource_DesiredState_IntegerValueRemoved(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudformation_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudformationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudformationResourceConfigDesiredStateIntegerValue(rName, 14),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "resource_model", regexp.MustCompile(`"RetentionInDays":14`)),
				),
			},
			{
				Config: testAccAwsCloudformationResourceConfigDesiredStateIntegerValueRemoved(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "resource_model", regexp.MustCompile(`"LogGroupName":`)),
				),
			},
		},
	})
}

func TestAccAwsCloudformationResource_DesiredState_IntegerValueUpdate(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudformation_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudformationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudformationResourceConfigDesiredStateIntegerValue(rName, 7),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "resource_model", regexp.MustCompile(`"RetentionInDays":7`)),
				),
			},
			{
				Config: testAccAwsCloudformationResourceConfigDesiredStateIntegerValue(rName, 14),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "resource_model", regexp.MustCompile(`"RetentionInDays":14`)),
				),
			},
		},
	})
}

func TestAccAwsCloudformationResource_DesiredState_InvalidPropertyName(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudformationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAwsCloudformationResourceConfigDesiredStateInvalidPropertyName(rName),
				ExpectError: regexp.MustCompile(`\(root\): Additional property InvalidName is not allowed`),
			},
		},
	})
}

func TestAccAwsCloudformationResource_DesiredState_InvalidPropertyValue(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudformationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAwsCloudformationResourceConfigDesiredStateInvalidPropertyValue(rName),
				ExpectError: regexp.MustCompile(`LogGroupName: Does not match pattern`),
			},
		},
	})
}

func TestAccAwsCloudformationResource_DesiredState_ObjectValueAdded(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudformation_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudformationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudformationResourceConfigDesiredStateObjectValueRemoved(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "resource_model", regexp.MustCompile(`"Name":`)),
				),
			},
			{
				Config: testAccAwsCloudformationResourceConfigDesiredStateObjectValue1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "resource_model", regexp.MustCompile(`"Value":"value1"`)),
				),
			},
		},
	})
}

func TestAccAwsCloudformationResource_DesiredState_ObjectValueRemoved(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudformation_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudformationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudformationResourceConfigDesiredStateObjectValue1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "resource_model", regexp.MustCompile(`"Value":"value1"`)),
				),
			},
			{
				Config: testAccAwsCloudformationResourceConfigDesiredStateObjectValueRemoved(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "resource_model", regexp.MustCompile(`"Name":`)),
				),
			},
		},
	})
}

func TestAccAwsCloudformationResource_DesiredState_ObjectValueUpdate(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudformation_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudformationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudformationResourceConfigDesiredStateObjectValue1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "resource_model", regexp.MustCompile(`"Value":"value1"`)),
				),
			},
			{
				Config: testAccAwsCloudformationResourceConfigDesiredStateObjectValue1(rName, "key1", "value1updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "resource_model", regexp.MustCompile(`"Value":"value1updated"`)),
				),
			},
			{
				Config: testAccAwsCloudformationResourceConfigDesiredStateObjectValue1(rName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "resource_model", regexp.MustCompile(`"Value":"value2"`)),
				),
			},
		},
	})
}

func TestAccAwsCloudformationResource_DesiredState_StringValueAdded(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudformation_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudformationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudformationResourceConfigDesiredStateStringValueRemoved(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "resource_model", regexp.MustCompile(`"Name":`)),
				),
			},
			{
				Config: testAccAwsCloudformationResourceConfigDesiredStateStringValue(rName, "description1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "resource_model", regexp.MustCompile(`"Description":"description1"`)),
				),
			},
		},
	})
}

func TestAccAwsCloudformationResource_DesiredState_StringValueRemoved(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudformation_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudformationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudformationResourceConfigDesiredStateStringValue(rName, "description1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "resource_model", regexp.MustCompile(`"Description":"description1"`)),
				),
			},
			{
				Config: testAccAwsCloudformationResourceConfigDesiredStateStringValueRemoved(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "resource_model", regexp.MustCompile(`"Name":`)),
				),
			},
		},
	})
}

func TestAccAwsCloudformationResource_DesiredState_StringValueUpdate(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudformation_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudformationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudformationResourceConfigDesiredStateStringValue(rName, "description1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "resource_model", regexp.MustCompile(`"Description":"description1"`)),
				),
			},
			{
				Config: testAccAwsCloudformationResourceConfigDesiredStateStringValue(rName, "description2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "resource_model", regexp.MustCompile(`"Description":"description2"`)),
				),
			},
		},
	})
}

func TestAccAwsCloudformationResource_ResourceSchema(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudformation_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudformationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudformationResourceConfigResourceSchema(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "schema", "data.aws_cloudformation_type.test", "schema"),
				),
			},
		},
	})
}

func testAccCheckAwsCloudformationResourceDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cfconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudformation_resource" {
			continue
		}

		input := &cloudformation.GetResourceInput{
			Identifier: aws.String(rs.Primary.ID),
			TypeName:   aws.String(rs.Primary.Attributes["type_name"]),
		}

		_, err := conn.GetResource(input)

		if tfawserr.ErrCodeEquals(err, cloudformation.ErrCodeResourceNotFoundException) {
			continue
		}

		// Temporary: Some CloudFormation Resources do not correctly re-map
		// "not found" errors, instead returning a HandlerFailureException.
		// These should be reported and fixed upstream over time, but for now
		// work around the issue only in CheckDestroy.
		if tfawserr.ErrMessageContains(err, cloudformation.ErrCodeHandlerFailureException, "not found") {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading CloudFormation Resource (%s): %w", rs.Primary.ID, err)
		}
	}

	return nil
}

func testAccAwsCloudformationResourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_resource" "test" {
  type_name = "AWS::Logs::LogGroup"

  desired_state = jsonencode({
    LogGroupName = %[1]q
  })
}
`, rName)
}

func testAccAwsCloudformationResourceConfigDesiredStateBooleanValue(rName string, booleanValue bool) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_resource" "test" {
  type_name = "AWS::ApiGateway::ApiKey"

  desired_state = jsonencode({
    Enabled = %[2]t
    Name    = %[1]q
    Value   = %[1]q
  })
}
`, rName, booleanValue)
}

func testAccAwsCloudformationResourceConfigDesiredStateBooleanValueRemoved(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_resource" "test" {
  type_name = "AWS::ApiGateway::ApiKey"

  desired_state = jsonencode({
    Name  = %[1]q
    Value = %[1]q
  })
}
`, rName)
}

func testAccAwsCloudformationResourceConfigDesiredStateCreateOnly(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_resource" "test" {
  type_name = "AWS::Logs::LogGroup"

  desired_state = jsonencode({
    LogGroupName = %[1]q
  })
}
`, rName)
}

func testAccAwsCloudformationResourceConfigDesiredStateIntegerValue(rName string, integerValue int) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_resource" "test" {
  type_name = "AWS::Logs::LogGroup"

  desired_state = jsonencode({
    LogGroupName    = %[1]q
    RetentionInDays = %[2]d
  })
}
`, rName, integerValue)
}

func testAccAwsCloudformationResourceConfigDesiredStateIntegerValueRemoved(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_resource" "test" {
  type_name = "AWS::Logs::LogGroup"

  desired_state = jsonencode({
    LogGroupName = %[1]q
  })
}
`, rName)
}

func testAccAwsCloudformationResourceConfigDesiredStateInvalidPropertyName(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_resource" "test" {
  type_name = "AWS::Logs::LogGroup"

  desired_state = jsonencode({
    InvalidName = %[1]q
  })
}
`, rName)
}

func testAccAwsCloudformationResourceConfigDesiredStateInvalidPropertyValue(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_resource" "test" {
  type_name = "AWS::Logs::LogGroup"

  desired_state = jsonencode({
    LogGroupName = "%[1]s!exclamation-not-valid"
  })
}
`, rName)
}

func testAccAwsCloudformationResourceConfigDesiredStateObjectValue1(rName string, key1 string, value1 string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_resource" "test" {
  type_name = "AWS::ECS::Cluster"

  desired_state = jsonencode({
    ClusterName = %[1]q
    Tags = [
      {
        Key   = %[2]q
        Value = %[3]q
      }
    ]
  })
}
`, rName, key1, value1)
}

func testAccAwsCloudformationResourceConfigDesiredStateObjectValueRemoved(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_resource" "test" {
  type_name = "AWS::ECS::Cluster"

  desired_state = jsonencode({
    ClusterName = %[1]q
  })
}
`, rName)
}

func testAccAwsCloudformationResourceConfigDesiredStateStringValue(rName string, stringValue string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_resource" "test" {
  type_name = "AWS::Athena::WorkGroup"

  desired_state = jsonencode({
    Description = %[2]q
    Name        = %[1]q
  })
}
`, rName, stringValue)
}

func testAccAwsCloudformationResourceConfigDesiredStateStringValueRemoved(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_resource" "test" {
  type_name = "AWS::Athena::WorkGroup"

  desired_state = jsonencode({
    Name = %[1]q
  })
}
`, rName)
}

func testAccAwsCloudformationResourceConfigResourceSchema(rName string) string {
	return fmt.Sprintf(`
data "aws_cloudformation_type" "test" {
  type      = "RESOURCE"
  type_name = "AWS::Logs::LogGroup"
}

resource "aws_cloudformation_resource" "test" {
  schema    = data.aws_cloudformation_type.test.schema
  type_name = data.aws_cloudformation_type.test.type_name

  desired_state = jsonencode({
    LogGroupName = %[1]q
  })
}
`, rName)
}
