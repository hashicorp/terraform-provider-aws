package glue_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfglue "github.com/hashicorp/terraform-provider-aws/internal/service/glue"
)

func TestAccGlueSchema_basic(t *testing.T) {
	var schema glue.GetSchemaOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_schema.test"
	registryResourceName := "aws_glue_registry.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckSchema(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSchemaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSchemaConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemaExists(resourceName, &schema),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("schema/%s/%s", rName, rName)),
					resource.TestCheckResourceAttr(resourceName, "schema_name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "compatibility", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "data_format", "AVRO"),
					resource.TestCheckResourceAttr(resourceName, "schema_checkpoint", "1"),
					resource.TestCheckResourceAttr(resourceName, "latest_schema_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "next_schema_version", "2"),
					resource.TestCheckResourceAttr(resourceName, "schema_definition", "{\"type\": \"record\", \"name\": \"r1\", \"fields\": [ {\"name\": \"f1\", \"type\": \"int\"}, {\"name\": \"f2\", \"type\": \"string\"} ]}"),
					resource.TestCheckResourceAttrPair(resourceName, "registry_name", registryResourceName, "registry_name"),
					resource.TestCheckResourceAttrPair(resourceName, "registry_arn", registryResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccGlueSchema_json(t *testing.T) {
	var schema glue.GetSchemaOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_schema.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckSchema(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSchemaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSchemaConfig_json(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemaExists(resourceName, &schema),
					resource.TestCheckResourceAttr(resourceName, "data_format", "JSON"),
					resource.TestCheckResourceAttr(resourceName, "schema_definition", "{\"$id\":\"https://example.com/person.schema.json\",\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"title\":\"Person\",\"type\":\"object\",\"properties\":{\"firstName\":{\"type\":\"string\",\"description\":\"The person's first name.\"},\"lastName\":{\"type\":\"string\",\"description\":\"The person's last name.\"},\"age\":{\"description\":\"Age in years which must be equal to or greater than zero.\",\"type\":\"integer\",\"minimum\":0}}}"),
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

func TestAccGlueSchema_protobuf(t *testing.T) {
	var schema glue.GetSchemaOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_schema.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckSchema(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSchemaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSchemaConfig_protobuf(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemaExists(resourceName, &schema),
					resource.TestCheckResourceAttr(resourceName, "data_format", "PROTOBUF"),
					resource.TestCheckResourceAttr(resourceName, "schema_definition", "syntax = \"proto2\";\n\npackage tutorial;\n\noption java_multiple_files = true;\noption java_package = \"com.example.tutorial.protos\";\noption java_outer_classname = \"AddressBookProtos\";\n\nmessage Person {\n  optional string name = 1;\n  optional int32 id = 2;\n  optional string email = 3;\n\n  enum PhoneType {\n    MOBILE = 0;\n    HOME = 1;\n    WORK = 2;\n  }\n\n  message PhoneNumber {\n    optional string number = 1;\n    optional PhoneType type = 2 [default = HOME];\n  }\n\n  repeated PhoneNumber phones = 4;\n}\n\nmessage AddressBook {\n  repeated Person people = 1;\n}"),
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

func TestAccGlueSchema_description(t *testing.T) {
	var schema glue.GetSchemaOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_schema.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckSchema(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSchemaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSchemaConfig_description(rName, "First Description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemaExists(resourceName, &schema),
					resource.TestCheckResourceAttr(resourceName, "description", "First Description"),
				),
			},
			{
				Config: testAccSchemaConfig_description(rName, "Second Description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemaExists(resourceName, &schema),
					resource.TestCheckResourceAttr(resourceName, "description", "Second Description"),
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

func TestAccGlueSchema_compatibility(t *testing.T) {
	var schema glue.GetSchemaOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_schema.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckSchema(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSchemaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSchemaConfig_compatibility(rName, "DISABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemaExists(resourceName, &schema),
					resource.TestCheckResourceAttr(resourceName, "compatibility", "DISABLED"),
				),
			},
			{
				Config: testAccSchemaConfig_compatibility(rName, "FULL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemaExists(resourceName, &schema),
					resource.TestCheckResourceAttr(resourceName, "compatibility", "FULL"),
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

func TestAccGlueSchema_tags(t *testing.T) {
	var schema glue.GetSchemaOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_schema.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckSchema(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSchemaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSchemaConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemaExists(resourceName, &schema),
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
				Config: testAccSchemaConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemaExists(resourceName, &schema),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccSchemaConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemaExists(resourceName, &schema),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccGlueSchema_schemaDefUpdated(t *testing.T) {
	var schema glue.GetSchemaOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_schema.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckSchema(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSchemaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSchemaConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemaExists(resourceName, &schema),
					resource.TestCheckResourceAttr(resourceName, "schema_definition", "{\"type\": \"record\", \"name\": \"r1\", \"fields\": [ {\"name\": \"f1\", \"type\": \"int\"}, {\"name\": \"f2\", \"type\": \"string\"} ]}"),
					resource.TestCheckResourceAttr(resourceName, "latest_schema_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "next_schema_version", "2"),
				),
			},
			{
				Config: testAccSchemaConfig_definitionUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemaExists(resourceName, &schema),
					resource.TestCheckResourceAttr(resourceName, "schema_definition", "{\"type\": \"record\", \"name\": \"r1\", \"fields\": [ {\"name\": \"f1\", \"type\": \"string\"}, {\"name\": \"f2\", \"type\": \"int\"} ]}"),
					resource.TestCheckResourceAttr(resourceName, "latest_schema_version", "2"),
					resource.TestCheckResourceAttr(resourceName, "next_schema_version", "3"),
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

func TestAccGlueSchema_disappears(t *testing.T) {
	var schema glue.GetSchemaOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_schema.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckSchema(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSchemaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSchemaConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemaExists(resourceName, &schema),
					acctest.CheckResourceDisappears(acctest.Provider, tfglue.ResourceSchema(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccGlueSchema_Disappears_registry(t *testing.T) {
	var schema glue.GetSchemaOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_schema.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckSchema(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSchemaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSchemaConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemaExists(resourceName, &schema),
					acctest.CheckResourceDisappears(acctest.Provider, tfglue.ResourceRegistry(), "aws_glue_registry.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccPreCheckSchema(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).GlueConn

	_, err := conn.ListRegistries(&glue.ListRegistriesInput{})

	// Some endpoints that do not support Glue Schemas return InternalFailure
	if acctest.PreCheckSkipError(err) || tfawserr.ErrCodeEquals(err, "InternalFailure") {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckSchemaExists(resourceName string, schema *glue.GetSchemaOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Glue Schema ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueConn
		output, err := tfglue.FindSchemaByID(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("Glue Schema (%s) not found", rs.Primary.ID)
		}

		if aws.StringValue(output.SchemaArn) == rs.Primary.ID {
			*schema = *output
			return nil
		}

		return fmt.Errorf("Glue Schema (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckSchemaDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_glue_schema" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueConn
		output, err := tfglue.FindSchemaByID(conn, rs.Primary.ID)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
				return nil
			}

		}

		if output != nil && aws.StringValue(output.SchemaArn) == rs.Primary.ID {
			return fmt.Errorf("Glue Schema %s still exists", rs.Primary.ID)
		}

		return err
	}

	return nil
}

func testAccSchemaBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_registry" "test" {
  registry_name = %[1]q
}
`, rName)
}

func testAccSchemaConfig_description(rName, description string) string {
	return testAccSchemaBase(rName) + fmt.Sprintf(`
resource "aws_glue_schema" "test" {
  schema_name       = %[1]q
  registry_arn      = aws_glue_registry.test.arn
  data_format       = "AVRO"
  compatibility     = "NONE"
  description       = %[2]q
  schema_definition = "{\"type\": \"record\", \"name\": \"r1\", \"fields\": [ {\"name\": \"f1\", \"type\": \"int\"}, {\"name\": \"f2\", \"type\": \"string\"} ]}"
}
`, rName, description)
}

func testAccSchemaConfig_compatibility(rName, compat string) string {
	return testAccSchemaBase(rName) + fmt.Sprintf(`
resource "aws_glue_schema" "test" {
  schema_name       = %[1]q
  registry_arn      = aws_glue_registry.test.arn
  data_format       = "AVRO"
  compatibility     = %[2]q
  schema_definition = "{\"type\": \"record\", \"name\": \"r1\", \"fields\": [ {\"name\": \"f1\", \"type\": \"int\"}, {\"name\": \"f2\", \"type\": \"string\"} ]}"
}
`, rName, compat)
}

func testAccSchemaConfig_basic(rName string) string {
	return testAccSchemaBase(rName) + fmt.Sprintf(`
resource "aws_glue_schema" "test" {
  schema_name       = %[1]q
  registry_arn      = aws_glue_registry.test.arn
  data_format       = "AVRO"
  compatibility     = "NONE"
  schema_definition = "{\"type\": \"record\", \"name\": \"r1\", \"fields\": [ {\"name\": \"f1\", \"type\": \"int\"}, {\"name\": \"f2\", \"type\": \"string\"} ]}"
}
`, rName)
}

func testAccSchemaConfig_json(rName string) string {
	return testAccSchemaBase(rName) + fmt.Sprintf(`
resource "aws_glue_schema" "test" {
  schema_name       = %[1]q
  registry_arn      = aws_glue_registry.test.arn
  data_format       = "JSON"
  compatibility     = "NONE"
  schema_definition = "{\"$id\":\"https://example.com/person.schema.json\",\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"title\":\"Person\",\"type\":\"object\",\"properties\":{\"firstName\":{\"type\":\"string\",\"description\":\"The person's first name.\"},\"lastName\":{\"type\":\"string\",\"description\":\"The person's last name.\"},\"age\":{\"description\":\"Age in years which must be equal to or greater than zero.\",\"type\":\"integer\",\"minimum\":0}}}"
}
`, rName)
}

func testAccSchemaConfig_protobuf(rName string) string {
	return testAccSchemaBase(rName) + fmt.Sprintf(`
resource "aws_glue_schema" "test" {
  schema_name       = %[1]q
  registry_arn      = aws_glue_registry.test.arn
  data_format       = "PROTOBUF"
  compatibility     = "NONE"
  schema_definition = "syntax = \"proto2\";\n\npackage tutorial;\n\noption java_multiple_files = true;\noption java_package = \"com.example.tutorial.protos\";\noption java_outer_classname = \"AddressBookProtos\";\n\nmessage Person {\n  optional string name = 1;\n  optional int32 id = 2;\n  optional string email = 3;\n\n  enum PhoneType {\n    MOBILE = 0;\n    HOME = 1;\n    WORK = 2;\n  }\n\n  message PhoneNumber {\n    optional string number = 1;\n    optional PhoneType type = 2 [default = HOME];\n  }\n\n  repeated PhoneNumber phones = 4;\n}\n\nmessage AddressBook {\n  repeated Person people = 1;\n}"
}
`, rName)
}

func testAccSchemaConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return testAccSchemaBase(rName) + fmt.Sprintf(`
resource "aws_glue_schema" "test" {
  schema_name       = %[1]q
  registry_arn      = aws_glue_registry.test.arn
  data_format       = "AVRO"
  compatibility     = "NONE"
  schema_definition = "{\"type\": \"record\", \"name\": \"r1\", \"fields\": [ {\"name\": \"f1\", \"type\": \"int\"}, {\"name\": \"f2\", \"type\": \"string\"} ]}"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccSchemaConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccSchemaBase(rName) + fmt.Sprintf(`
resource "aws_glue_schema" "test" {
  schema_name       = %[1]q
  registry_arn      = aws_glue_registry.test.arn
  data_format       = "AVRO"
  compatibility     = "NONE"
  schema_definition = "{\"type\": \"record\", \"name\": \"r1\", \"fields\": [ {\"name\": \"f1\", \"type\": \"int\"}, {\"name\": \"f2\", \"type\": \"string\"} ]}"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccSchemaConfig_definitionUpdated(rName string) string {
	return testAccSchemaBase(rName) + fmt.Sprintf(`
resource "aws_glue_schema" "test" {
  schema_name       = %[1]q
  registry_arn      = aws_glue_registry.test.arn
  data_format       = "AVRO"
  compatibility     = "NONE"
  schema_definition = "{\"type\": \"record\", \"name\": \"r1\", \"fields\": [ {\"name\": \"f1\", \"type\": \"string\"}, {\"name\": \"f2\", \"type\": \"int\"} ]}"
}
`, rName)
}
