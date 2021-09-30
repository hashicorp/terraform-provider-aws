package aws

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudcontrolapi"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudcontrol/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func init() {
	RegisterServiceErrorCheckFunc(cloudcontrolapi.EndpointsID, testAccErrorCheckSkipCloudControlAPI)
}

func testAccErrorCheckSkipCloudControlAPI(t *testing.T) resource.ErrorCheckFunc {
	return testAccErrorCheckSkipMessagesContaining(t,
		"UnsupportedActionException",
	)
}

func TestAccAwsCloudControlApiResource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudcontrolapi_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudcontrolapi.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudControlApiResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudControlApiResourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`^\{.*\}$`)),
					resource.TestMatchResourceAttr(resourceName, "schema", regexp.MustCompile(`^\{.*`)),
				),
			},
		},
	})
}

func TestAccAwsCloudControlApiResource_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudcontrolapi_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudcontrolapi.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudControlApiResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudControlApiResourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCloudControlApiResource(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsCloudControlApiResource_DesiredState_BooleanValueAdded(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudcontrolapi_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudcontrolapi.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudControlApiResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudControlApiResourceConfigDesiredStateBooleanValueRemoved(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"Enabled":false`)),
				),
			},
			{
				Config: testAccAwsCloudControlApiResourceConfigDesiredStateBooleanValue(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"Enabled":true`)),
				),
			},
		},
	})
}

func TestAccAwsCloudControlApiResource_DesiredState_BooleanValueRemoved(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudcontrolapi_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudcontrolapi.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudControlApiResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudControlApiResourceConfigDesiredStateBooleanValue(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"Enabled":true`)),
				),
			},
			{
				Config: testAccAwsCloudControlApiResourceConfigDesiredStateBooleanValueRemoved(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"Enabled":false`)),
				),
			},
		},
	})
}

func TestAccAwsCloudControlApiResource_DesiredState_BooleanValueUpdate(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudcontrolapi_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudcontrolapi.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudControlApiResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudControlApiResourceConfigDesiredStateBooleanValue(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"Enabled":true`)),
				),
			},
			{
				Config: testAccAwsCloudControlApiResourceConfigDesiredStateBooleanValue(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"Enabled":false`)),
				),
			},
		},
	})
}

func TestAccAwsCloudControlApiResource_DesiredState_CreateOnly(t *testing.T) {
	rName1 := acctest.RandomWithPrefix("tf-acc-test")
	rName2 := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudcontrolapi_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudcontrolapi.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudControlApiResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudControlApiResourceConfigDesiredStateCreateOnly(rName1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"LogGroupName":"`+rName1+`"`)),
				),
			},
			{
				Config: testAccAwsCloudControlApiResourceConfigDesiredStateCreateOnly(rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"LogGroupName":"`+rName2+`"`)),
				),
			},
		},
	})
}

func TestAccAwsCloudControlApiResource_DesiredState_IntegerValueAdded(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudcontrolapi_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudcontrolapi.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudControlApiResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudControlApiResourceConfigDesiredStateIntegerValueRemoved(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"LogGroupName":`)),
				),
			},
			{
				Config: testAccAwsCloudControlApiResourceConfigDesiredStateIntegerValue(rName, 14),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"RetentionInDays":14`)),
				),
			},
		},
	})
}

func TestAccAwsCloudControlApiResource_DesiredState_IntegerValueRemoved(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudcontrolapi_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudcontrolapi.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudControlApiResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudControlApiResourceConfigDesiredStateIntegerValue(rName, 14),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"RetentionInDays":14`)),
				),
			},
			{
				Config: testAccAwsCloudControlApiResourceConfigDesiredStateIntegerValueRemoved(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"LogGroupName":`)),
				),
			},
		},
	})
}

func TestAccAwsCloudControlApiResource_DesiredState_IntegerValueUpdate(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudcontrolapi_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudcontrolapi.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudControlApiResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudControlApiResourceConfigDesiredStateIntegerValue(rName, 7),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"RetentionInDays":7`)),
				),
			},
			{
				Config: testAccAwsCloudControlApiResourceConfigDesiredStateIntegerValue(rName, 14),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"RetentionInDays":14`)),
				),
			},
		},
	})
}

func TestAccAwsCloudControlApiResource_DesiredState_InvalidPropertyName(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudcontrolapi.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudControlApiResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAwsCloudControlApiResourceConfigDesiredStateInvalidPropertyName(rName),
				ExpectError: regexp.MustCompile(`\(root\): Additional property InvalidName is not allowed`),
			},
		},
	})
}

func TestAccAwsCloudControlApiResource_DesiredState_InvalidPropertyValue(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudcontrolapi.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudControlApiResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAwsCloudControlApiResourceConfigDesiredStateInvalidPropertyValue(rName),
				ExpectError: regexp.MustCompile(`Model validation failed`),
			},
		},
	})
}

func TestAccAwsCloudControlApiResource_DesiredState_ObjectValueAdded(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudcontrolapi_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudcontrolapi.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudControlApiResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudControlApiResourceConfigDesiredStateObjectValueRemoved(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"Name":`)),
				),
			},
			{
				Config: testAccAwsCloudControlApiResourceConfigDesiredStateObjectValue1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"Value":"value1"`)),
				),
			},
		},
	})
}

func TestAccAwsCloudControlApiResource_DesiredState_ObjectValueRemoved(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudcontrolapi_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudcontrolapi.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudControlApiResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudControlApiResourceConfigDesiredStateObjectValue1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"Value":"value1"`)),
				),
			},
			{
				Config: testAccAwsCloudControlApiResourceConfigDesiredStateObjectValueRemoved(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"Name":`)),
				),
			},
		},
	})
}

func TestAccAwsCloudControlApiResource_DesiredState_ObjectValueUpdate(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudcontrolapi_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudcontrolapi.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudControlApiResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudControlApiResourceConfigDesiredStateObjectValue1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"Value":"value1"`)),
				),
			},
			{
				Config: testAccAwsCloudControlApiResourceConfigDesiredStateObjectValue1(rName, "key1", "value1updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"Value":"value1updated"`)),
				),
			},
			{
				Config: testAccAwsCloudControlApiResourceConfigDesiredStateObjectValue1(rName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"Value":"value2"`)),
				),
			},
		},
	})
}

func TestAccAwsCloudControlApiResource_DesiredState_StringValueAdded(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudcontrolapi_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudcontrolapi.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudControlApiResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudControlApiResourceConfigDesiredStateStringValueRemoved(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"Name":`)),
				),
			},
			{
				Config: testAccAwsCloudControlApiResourceConfigDesiredStateStringValue(rName, "description1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"Description":"description1"`)),
				),
			},
		},
	})
}

func TestAccAwsCloudControlApiResource_DesiredState_StringValueRemoved(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudcontrolapi_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudcontrolapi.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudControlApiResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudControlApiResourceConfigDesiredStateStringValue(rName, "description1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"Description":"description1"`)),
				),
			},
			{
				Config: testAccAwsCloudControlApiResourceConfigDesiredStateStringValueRemoved(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"Name":`)),
				),
			},
		},
	})
}

func TestAccAwsCloudControlApiResource_DesiredState_StringValueUpdate(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudcontrolapi_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudcontrolapi.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudControlApiResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudControlApiResourceConfigDesiredStateStringValue(rName, "description1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"Description":"description1"`)),
				),
			},
			{
				Config: testAccAwsCloudControlApiResourceConfigDesiredStateStringValue(rName, "description2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"Description":"description2"`)),
				),
			},
		},
	})
}

func TestAccAwsCloudControlApiResource_ResourceSchema(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudcontrolapi_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudcontrolapi.EndpointsID, cloudformation.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudControlApiResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudControlApiResourceConfigResourceSchema(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "schema", "data.aws_cloudformation_type.test", "schema"),
				),
			},
		},
	})
}

func testAccCheckAwsCloudControlApiResourceDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudcontrolapiconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudcontrolapi_resource" {
			continue
		}

		_, err := finder.ResourceByID(context.TODO(), conn, rs.Primary.ID, rs.Primary.Attributes["type_name"], "", "")

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Cloud Control API Resource %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccAwsCloudControlApiResourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudcontrolapi_resource" "test" {
  type_name = "AWS::Logs::LogGroup"

  desired_state = jsonencode({
    LogGroupName = %[1]q
  })
}
`, rName)
}

func testAccAwsCloudControlApiResourceConfigDesiredStateBooleanValue(rName string, booleanValue bool) string {
	return fmt.Sprintf(`
resource "aws_cloudcontrolapi_resource" "test" {
  type_name = "AWS::ApiGateway::ApiKey"

  desired_state = jsonencode({
    Enabled = %[2]t
    Name    = %[1]q
    Value   = %[1]q
  })
}
`, rName, booleanValue)
}

func testAccAwsCloudControlApiResourceConfigDesiredStateBooleanValueRemoved(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudcontrolapi_resource" "test" {
  type_name = "AWS::ApiGateway::ApiKey"

  desired_state = jsonencode({
    Name  = %[1]q
    Value = %[1]q
  })
}
`, rName)
}

func testAccAwsCloudControlApiResourceConfigDesiredStateCreateOnly(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudcontrolapi_resource" "test" {
  type_name = "AWS::Logs::LogGroup"

  desired_state = jsonencode({
    LogGroupName = %[1]q
  })
}
`, rName)
}

func testAccAwsCloudControlApiResourceConfigDesiredStateIntegerValue(rName string, integerValue int) string {
	return fmt.Sprintf(`
resource "aws_cloudcontrolapi_resource" "test" {
  type_name = "AWS::Logs::LogGroup"

  desired_state = jsonencode({
    LogGroupName    = %[1]q
    RetentionInDays = %[2]d
  })
}
`, rName, integerValue)
}

func testAccAwsCloudControlApiResourceConfigDesiredStateIntegerValueRemoved(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudcontrolapi_resource" "test" {
  type_name = "AWS::Logs::LogGroup"

  desired_state = jsonencode({
    LogGroupName = %[1]q
  })
}
`, rName)
}

func testAccAwsCloudControlApiResourceConfigDesiredStateInvalidPropertyName(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudcontrolapi_resource" "test" {
  type_name = "AWS::Logs::LogGroup"

  desired_state = jsonencode({
    InvalidName = %[1]q
  })
}
`, rName)
}

func testAccAwsCloudControlApiResourceConfigDesiredStateInvalidPropertyValue(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudcontrolapi_resource" "test" {
  type_name = "AWS::Logs::LogGroup"

  desired_state = jsonencode({
    LogGroupName = "%[1]s!exclamation-not-valid"
  })
}
`, rName)
}

func testAccAwsCloudControlApiResourceConfigDesiredStateObjectValue1(rName string, key1 string, value1 string) string {
	return fmt.Sprintf(`
resource "aws_cloudcontrolapi_resource" "test" {
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

func testAccAwsCloudControlApiResourceConfigDesiredStateObjectValueRemoved(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudcontrolapi_resource" "test" {
  type_name = "AWS::ECS::Cluster"

  desired_state = jsonencode({
    ClusterName = %[1]q
  })
}
`, rName)
}

func testAccAwsCloudControlApiResourceConfigDesiredStateStringValue(rName string, stringValue string) string {
	return fmt.Sprintf(`
resource "aws_cloudcontrolapi_resource" "test" {
  type_name = "AWS::Athena::WorkGroup"

  desired_state = jsonencode({
    Description = %[2]q
    Name        = %[1]q
  })
}
`, rName, stringValue)
}

func testAccAwsCloudControlApiResourceConfigDesiredStateStringValueRemoved(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudcontrolapi_resource" "test" {
  type_name = "AWS::Athena::WorkGroup"

  desired_state = jsonencode({
    Name = %[1]q
  })
}
`, rName)
}

func testAccAwsCloudControlApiResourceConfigResourceSchema(rName string) string {
	return fmt.Sprintf(`
data "aws_cloudformation_type" "test" {
  type      = "RESOURCE"
  type_name = "AWS::Logs::LogGroup"
}

resource "aws_cloudcontrolapi_resource" "test" {
  schema    = data.aws_cloudformation_type.test.schema
  type_name = data.aws_cloudformation_type.test.type_name

  desired_state = jsonencode({
    LogGroupName = %[1]q
  })
}
`, rName)
}
