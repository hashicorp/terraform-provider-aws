package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	schemas "github.com/aws/aws-sdk-go/service/schemas"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var schemaType = schemas.TypeOpenApi3
var content = `{
	"openapi": "3.0.0",
	"info": {
	  "version": "1.0.0",
	  "title": "Event"
	},
	"paths": {},
	"components": {
	  "schemas": {
		"Event": {
		  "type": "object",
		  "properties": {
			"name": {
			  "type": "string"
			}
		  }
		}
	  }
	}
  }
`
var contentModified = `{
	"openapi": "3.0.0",
	"info": {
	  "version": "2.0.0",
	  "title": "Event"
	},
	"paths": {},
	"components": {
	  "schemas": {
		"Event": {
		  "type": "object",
		  "properties": {
			"name": {
			  "type": "string"
			},
			"created_at": {
			  "type": "string",
			  "format": "date-time"
			}
		  }
		}
	  }
	}
  }
`

func init() {
	resource.AddTestSweepers("aws_schemas_schema", &resource.Sweeper{
		Name: "aws_schemas_schema",
		F:    testSweepSchemasSchema,
		Dependencies: []string{
			"aws_schemas_registry",
		},
	})
}

func testSweepSchemasSchema(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.(*AWSClient).schemasconn

	var sweeperErrs *multierror.Error
	var deletedSchemas int

	input := &schemas.ListRegistriesInput{
		Limit: aws.Int64(100),
	}
	var registries []*schemas.RegistrySummary
	for {
		output, err := conn.ListRegistries(input)
		if err != nil {
			return err
		}
		registries = append(registries, output.Registries...)

		if aws.StringValue(output.NextToken) == "" {
			break
		}
		input.NextToken = output.NextToken
	}

	for _, registry := range registries {
		input := &schemas.ListSchemasInput{
			Limit:        aws.Int64(100),
			RegistryName: registry.RegistryName,
		}
		var existingSchemas []*schemas.SchemaSummary
		for {
			output, err := conn.ListSchemas(input)
			if err != nil {
				return err
			}
			existingSchemas = append(existingSchemas, output.Schemas...)

			if aws.StringValue(output.NextToken) == "" {
				break
			}
			input.NextToken = output.NextToken
		}

		for _, existingSchema := range existingSchemas {

			input := &schemas.DeleteSchemaInput{
				SchemaName:   existingSchema.SchemaName,
				RegistryName: registry.RegistryName,
			}
			_, err := conn.DeleteSchema(input)
			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("Error deleting Schemas Schema (%s): %w", *existingSchema.SchemaName, err))
				continue
			}
			deletedSchemas += 1
		}
	}

	log.Printf("[INFO] Deleted %d Schemas Schemas", deletedSchemas)

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSSchemasSchema_basic(t *testing.T) {
	var v1, v2, v3 schemas.DescribeSchemaOutput
	name := acctest.RandomWithPrefix("tf-acc-test")
	nameModified := acctest.RandomWithPrefix("tf-acc-test")

	registry := acctest.RandomWithPrefix("tf-acc-test")

	description := acctest.RandomWithPrefix("tf-acc-test")
	descriptionModified := acctest.RandomWithPrefix("tf-acc-test")

	resourceName := "aws_schemas_schema.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, schemas.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSchemasSchemaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSchemasSchemaConfig(
					name,
					registry,
					schemaType,
					content,
					description,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasSchemaExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "registry", registry),
					resource.TestCheckResourceAttr(resourceName, "type", schemaType),
					resource.TestCheckResourceAttr(resourceName, "content", content),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "schemas", fmt.Sprintf("schema/%s/%s", registry, name)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSchemasSchemaConfig(
					nameModified,
					registry,
					schemaType,
					contentModified,
					descriptionModified,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasSchemaExists(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "name", nameModified),
					resource.TestCheckResourceAttr(resourceName, "registry", registry),
					resource.TestCheckResourceAttr(resourceName, "type", schemaType),
					resource.TestCheckResourceAttr(resourceName, "content", contentModified),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionModified),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "schemas", fmt.Sprintf("schema/%s/%s", registry, nameModified)),
					testAccCheckSchemasSchemaRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccAWSSchemasSchemaConfig_Tags1(
					nameModified,
					registry,
					schemaType,
					contentModified,
					"key",
					"value",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasSchemaExists(resourceName, &v3),
					testAccCheckSchemasSchemaNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key", "value"),
				),
			},
		},
	})
}

func TestAccAWSSchemasSchema_tags(t *testing.T) {
	var v1, v2, v3, v4 schemas.DescribeSchemaOutput
	name := acctest.RandomWithPrefix("tf-acc-test")
	registry := acctest.RandomWithPrefix("tf-acc-test")
	description := acctest.RandomWithPrefix("tf-acc-test")

	resourceName := "aws_schemas_schema.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, schemas.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSchemasSchemaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSchemasSchemaConfig_Tags1(
					name,
					registry,
					schemaType,
					content,
					"key1",
					"value",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasSchemaExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSchemasSchemaConfig_Tags2(
					name,
					registry,
					schemaType,
					content,
					"key1",
					"updated",
					"key2",
					"added",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasSchemaExists(resourceName, &v2),
					testAccCheckSchemasSchemaNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "added"),
				),
			},
			{
				Config: testAccAWSSchemasSchemaConfig_Tags1(
					name,
					registry,
					schemaType,
					content,
					"key2",
					"added",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasSchemaExists(resourceName, &v3),
					testAccCheckSchemasSchemaNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "added"),
				),
			},
			{
				Config: testAccAWSSchemasSchemaConfig(
					name,
					registry,
					schemaType,
					content,
					description,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasSchemaExists(resourceName, &v4),
					testAccCheckSchemasSchemaNotRecreated(&v3, &v4),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSSchemasSchema_disappears(t *testing.T) {
	var v schemas.DescribeSchemaOutput
	name := acctest.RandomWithPrefix("tf-acc-test")
	registry := acctest.RandomWithPrefix("tf-acc-test")
	description := acctest.RandomWithPrefix("tf-acc-test")

	resourceName := "aws_schemas_schema.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, schemas.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSchemasSchemaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSchemasSchemaConfig(
					name,
					registry,
					schemaType,
					content,
					description,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasSchemaExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSchemasSchema(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSSchemasSchemaDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).schemasconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_schemas_schema" {
			continue
		}

		schemaName, registryName, err := parseSchemaID(rs.Primary.ID)
		if err != nil {
			return err
		}

		params := schemas.DescribeSchemaInput{
			SchemaName:   aws.String(schemaName),
			RegistryName: aws.String(registryName),
		}

		resp, err := conn.DescribeSchema(&params)

		if err == nil {
			return fmt.Errorf("Schemas Schema (%s) still exists: %s", rs.Primary.ID, resp)
		}
	}

	return nil
}

func testAccCheckSchemasSchemaExists(n string, v *schemas.DescribeSchemaOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		schemaName, registryName, err := parseSchemaID(rs.Primary.ID)
		if err != nil {
			return err
		}

		params := schemas.DescribeSchemaInput{
			SchemaName:   aws.String(schemaName),
			RegistryName: aws.String(registryName),
		}
		conn := testAccProvider.Meta().(*AWSClient).schemasconn

		resp, err := conn.DescribeSchema(&params)
		if err != nil {
			return err
		}
		if resp == nil {
			return fmt.Errorf("Schemas Schema (%s) not found", n)
		}

		*v = *resp

		return nil
	}
}

func testAccCheckSchemasSchemaRecreated(i, j *schemas.DescribeSchemaOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.SchemaArn) == aws.StringValue(j.SchemaArn) {
			return fmt.Errorf("Schemas Schema not recreated")
		}
		return nil
	}
}

func testAccCheckSchemasSchemaNotRecreated(i, j *schemas.DescribeSchemaOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.SchemaArn) != aws.StringValue(j.SchemaArn) {
			return fmt.Errorf("Schemas Schema was recreated")
		}
		return nil
	}
}

func testAccAWSSchemasSchemaConfig(name, registry, schemaType, content, description string) string {
	return fmt.Sprintf(`
resource "aws_schemas_registry" "test" {
	name = %[2]q
}

resource "aws_schemas_schema" "test" {
  name = %[1]q
  registry = aws_schemas_registry.test.name
  type = %[3]q
  content = %[4]q
  description = %[5]q
}
`, name, registry, schemaType, content, description)
}

func testAccAWSSchemasSchemaConfig_Tags1(name, registry, schemaType, content, key, value string) string {
	return fmt.Sprintf(`
resource "aws_schemas_registry" "test" {
	name = %[2]q
}

resource "aws_schemas_schema" "test" {
	name = %[1]q
	registry = aws_schemas_registry.test.name
	type = %[3]q
	content = %[4]q

  tags = {
    %[5]q = %[6]q
  }
}
`, name, registry, schemaType, content, key, value)
}

func testAccAWSSchemasSchemaConfig_Tags2(name, registry, schemaType, content, key1, value1, key2, value2 string) string {
	return fmt.Sprintf(`	
resource "aws_schemas_registry" "test" {
	name = %[2]q
}

resource "aws_schemas_schema" "test" {
	name = %[1]q
	registry = aws_schemas_registry.test.name
	type = %[3]q
	content = %[4]q

  tags = {
    %[5]q = %[6]q
    %[7]q = %[8]q
  }
}
`, name, registry, schemaType, content, key1, value1, key2, value2)
}
