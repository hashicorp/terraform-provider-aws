package cloudcontrol_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudcontrolapi"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudcontrol "github.com/hashicorp/terraform-provider-aws/internal/service/cloudcontrol"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(cloudcontrolapi.EndpointsID, testAccErrorCheckSkip)
}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"UnsupportedActionException",
	)
}

func TestAccCloudControlResource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudcontrolapi_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudcontrolapi.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`^\{.*\}$`)),
					resource.TestMatchResourceAttr(resourceName, "schema", regexp.MustCompile(`^\{.*`)),
				),
			},
		},
	})
}

func TestAccCloudControlResource_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudcontrolapi_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudcontrolapi.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceDisappears(acctest.Provider, tfcloudcontrol.ResourceResource(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudControlResource_DesiredState_booleanValueAdded(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudcontrolapi_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudcontrolapi.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDesiredStateBooleanValueRemovedConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"Enabled":false`)),
				),
			},
			{
				Config: testAccResourceDesiredStateBooleanValueConfig(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"Enabled":true`)),
				),
			},
		},
	})
}

func TestAccCloudControlResource_DesiredState_booleanValueRemoved(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudcontrolapi_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudcontrolapi.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDesiredStateBooleanValueConfig(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"Enabled":true`)),
				),
			},
			{
				Config: testAccResourceDesiredStateBooleanValueRemovedConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"Enabled":false`)),
				),
			},
		},
	})
}

func TestAccCloudControlResource_DesiredState_booleanValueUpdate(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudcontrolapi_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudcontrolapi.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDesiredStateBooleanValueConfig(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"Enabled":true`)),
				),
			},
			{
				Config: testAccResourceDesiredStateBooleanValueConfig(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"Enabled":false`)),
				),
			},
		},
	})
}

func TestAccCloudControlResource_DesiredState_createOnly(t *testing.T) {
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudcontrolapi_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudcontrolapi.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDesiredStateCreateOnlyConfig(rName1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"LogGroupName":"`+rName1+`"`)),
				),
			},
			{
				Config: testAccResourceDesiredStateCreateOnlyConfig(rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"LogGroupName":"`+rName2+`"`)),
				),
			},
		},
	})
}

func TestAccCloudControlResource_DesiredState_integerValueAdded(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudcontrolapi_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudcontrolapi.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDesiredStateIntegerValueRemovedConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"LogGroupName":`)),
				),
			},
			{
				Config: testAccResourceDesiredStateIntegerValueConfig(rName, 14),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"RetentionInDays":14`)),
				),
			},
		},
	})
}

func TestAccCloudControlResource_DesiredState_integerValueRemoved(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudcontrolapi_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudcontrolapi.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDesiredStateIntegerValueConfig(rName, 14),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"RetentionInDays":14`)),
				),
			},
			{
				Config: testAccResourceDesiredStateIntegerValueRemovedConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"LogGroupName":`)),
				),
			},
		},
	})
}

func TestAccCloudControlResource_DesiredState_integerValueUpdate(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudcontrolapi_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudcontrolapi.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDesiredStateIntegerValueConfig(rName, 7),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"RetentionInDays":7`)),
				),
			},
			{
				Config: testAccResourceDesiredStateIntegerValueConfig(rName, 14),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"RetentionInDays":14`)),
				),
			},
		},
	})
}

func TestAccCloudControlResource_DesiredState_invalidPropertyName(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudcontrolapi.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceDesiredStateInvalidPropertyNameConfig(rName),
				ExpectError: regexp.MustCompile(`\(root\): Additional property InvalidName is not allowed`),
			},
		},
	})
}

func TestAccCloudControlResource_DesiredState_invalidPropertyValue(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudcontrolapi.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceDesiredStateInvalidPropertyValueConfig(rName),
				ExpectError: regexp.MustCompile(`Model validation failed`),
			},
		},
	})
}

func TestAccCloudControlResource_DesiredState_objectValueAdded(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudcontrolapi_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudcontrolapi.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDesiredStateObjectValueRemovedConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"Name":`)),
				),
			},
			{
				Config: testAccResourceDesiredStateObjectValue1Config(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"Value":"value1"`)),
				),
			},
		},
	})
}

func TestAccCloudControlResource_DesiredState_objectValueRemoved(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudcontrolapi_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudcontrolapi.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDesiredStateObjectValue1Config(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"Value":"value1"`)),
				),
			},
			{
				Config: testAccResourceDesiredStateObjectValueRemovedConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"Name":`)),
				),
			},
		},
	})
}

func TestAccCloudControlResource_DesiredState_objectValueUpdate(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudcontrolapi_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudcontrolapi.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDesiredStateObjectValue1Config(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"Value":"value1"`)),
				),
			},
			{
				Config: testAccResourceDesiredStateObjectValue1Config(rName, "key1", "value1updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"Value":"value1updated"`)),
				),
			},
			{
				Config: testAccResourceDesiredStateObjectValue1Config(rName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"Value":"value2"`)),
				),
			},
		},
	})
}

func TestAccCloudControlResource_DesiredState_stringValueAdded(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudcontrolapi_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudcontrolapi.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDesiredStateStringValueRemovedConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"Name":`)),
				),
			},
			{
				Config: testAccResourceDesiredStateStringValueConfig(rName, "description1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"Description":"description1"`)),
				),
			},
		},
	})
}

func TestAccCloudControlResource_DesiredState_stringValueRemoved(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudcontrolapi_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudcontrolapi.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDesiredStateStringValueConfig(rName, "description1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"Description":"description1"`)),
				),
			},
			{
				Config: testAccResourceDesiredStateStringValueRemovedConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"Name":`)),
				),
			},
		},
	})
}

func TestAccCloudControlResource_DesiredState_stringValueUpdate(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudcontrolapi_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudcontrolapi.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDesiredStateStringValueConfig(rName, "description1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"Description":"description1"`)),
				),
			},
			{
				Config: testAccResourceDesiredStateStringValueConfig(rName, "description2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "properties", regexp.MustCompile(`"Description":"description2"`)),
				),
			},
		},
	})
}

func TestAccCloudControlResource_resourceSchema(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudcontrolapi_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudcontrolapi.EndpointsID, cloudformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceResourceSchemaConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "schema", "data.aws_cloudformation_type.test", "schema"),
				),
			},
		},
	})
}

func testAccCheckResourceDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CloudControlConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudcontrolapi_resource" {
			continue
		}

		_, err := tfcloudcontrol.FindResourceByID(context.TODO(), conn, rs.Primary.ID, rs.Primary.Attributes["type_name"], "", "")

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

func testAccResourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudcontrolapi_resource" "test" {
  type_name = "AWS::Logs::LogGroup"

  desired_state = jsonencode({
    LogGroupName = %[1]q
  })
}
`, rName)
}

func testAccResourceDesiredStateBooleanValueConfig(rName string, booleanValue bool) string {
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

func testAccResourceDesiredStateBooleanValueRemovedConfig(rName string) string {
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

func testAccResourceDesiredStateCreateOnlyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudcontrolapi_resource" "test" {
  type_name = "AWS::Logs::LogGroup"

  desired_state = jsonencode({
    LogGroupName = %[1]q
  })
}
`, rName)
}

func testAccResourceDesiredStateIntegerValueConfig(rName string, integerValue int) string {
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

func testAccResourceDesiredStateIntegerValueRemovedConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudcontrolapi_resource" "test" {
  type_name = "AWS::Logs::LogGroup"

  desired_state = jsonencode({
    LogGroupName = %[1]q
  })
}
`, rName)
}

func testAccResourceDesiredStateInvalidPropertyNameConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudcontrolapi_resource" "test" {
  type_name = "AWS::Logs::LogGroup"

  desired_state = jsonencode({
    InvalidName = %[1]q
  })
}
`, rName)
}

func testAccResourceDesiredStateInvalidPropertyValueConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudcontrolapi_resource" "test" {
  type_name = "AWS::Logs::LogGroup"

  desired_state = jsonencode({
    LogGroupName = "%[1]s!exclamation-not-valid"
  })
}
`, rName)
}

func testAccResourceDesiredStateObjectValue1Config(rName string, key1 string, value1 string) string {
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

func testAccResourceDesiredStateObjectValueRemovedConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudcontrolapi_resource" "test" {
  type_name = "AWS::ECS::Cluster"

  desired_state = jsonencode({
    ClusterName = %[1]q
  })
}
`, rName)
}

func testAccResourceDesiredStateStringValueConfig(rName string, stringValue string) string {
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

func testAccResourceDesiredStateStringValueRemovedConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudcontrolapi_resource" "test" {
  type_name = "AWS::Athena::WorkGroup"

  desired_state = jsonencode({
    Name = %[1]q
  })
}
`, rName)
}

func testAccResourceResourceSchemaConfig(rName string) string {
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
