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

func init() {
	resource.AddTestSweepers("aws_schemas_registry", &resource.Sweeper{
		Name: "aws_schemas_registry",
		F:    testSweepSchemasRegistry,
	})
}

func testSweepSchemasRegistry(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.(*AWSClient).schemasconn

	var sweeperErrs *multierror.Error

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

		input := &schemas.DeleteRegistryInput{
			RegistryName: registry.RegistryName,
		}
		_, err := conn.DeleteRegistry(input)
		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("Error deleting Schemas Registry (%s): %w", *registry.RegistryName, err))
			continue
		}
	}

	log.Printf("[INFO] Deleted %d Schemas Registries", len(registries))

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSSchemasRegistry_basic(t *testing.T) {
	var v1, v2, v3 schemas.DescribeRegistryOutput
	registryName := acctest.RandomWithPrefix("tf-acc-test")
	registryNameModified := acctest.RandomWithPrefix("tf-acc-test")

	description := acctest.RandomWithPrefix("tf-acc-test")
	descriptionModified := acctest.RandomWithPrefix("tf-acc-test")

	resourceName := "aws_schemas_registry.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, schemas.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSchemasRegistryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSchemasRegistryConfig(registryName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasRegistryExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "name", registryName),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "schemas", fmt.Sprintf("registry/%s", registryName)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSchemasRegistryConfig(registryNameModified, descriptionModified),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasRegistryExists(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "name", registryNameModified),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionModified),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "schemas", fmt.Sprintf("registry/%s", registryNameModified)),
					testAccCheckSchemasRegistryRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccAWSSchemasRegistryConfig_Tags1(registryNameModified, "key", "value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasRegistryExists(resourceName, &v3),
					testAccCheckSchemasRegistryNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key", "value"),
				),
			},
		},
	})
}

func TestAccAWSSchemasRegistry_tags(t *testing.T) {
	var v1, v2, v3, v4 schemas.DescribeRegistryOutput
	registryName := acctest.RandomWithPrefix("tf-acc-test")
	description := acctest.RandomWithPrefix("tf-acc-test")

	resourceName := "aws_schemas_registry.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, schemas.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSchemasRegistryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSchemasRegistryConfig_Tags1(registryName, "key1", "value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasRegistryExists(resourceName, &v1),
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
				Config: testAccAWSSchemasRegistryConfig_Tags2(registryName, "key1", "updated", "key2", "added"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasRegistryExists(resourceName, &v2),
					testAccCheckSchemasRegistryNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "added"),
				),
			},
			{
				Config: testAccAWSSchemasRegistryConfig_Tags1(registryName, "key2", "added"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasRegistryExists(resourceName, &v3),
					testAccCheckSchemasRegistryNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "added"),
				),
			},
			{
				Config: testAccAWSSchemasRegistryConfig(registryName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasRegistryExists(resourceName, &v4),
					testAccCheckSchemasRegistryNotRecreated(&v3, &v4),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSSchemasRegistry_disappears(t *testing.T) {
	var v schemas.DescribeRegistryOutput
	registryName := acctest.RandomWithPrefix("tf-acc-test")
	description := acctest.RandomWithPrefix("tf-acc-test")

	resourceName := "aws_schemas_registry.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, schemas.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSchemasRegistryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSchemasRegistryConfig(registryName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasRegistryExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSchemasRegistry(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSSchemasRegistryDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).schemasconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_schemas_registry" {
			continue
		}

		params := schemas.DescribeRegistryInput{
			RegistryName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeRegistry(&params)

		if err == nil {
			return fmt.Errorf("Schemas Registry (%s) still exists: %s", rs.Primary.ID, resp)
		}
	}

	return nil
}

func testAccCheckSchemasRegistryExists(n string, v *schemas.DescribeRegistryOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).schemasconn
		params := schemas.DescribeRegistryInput{
			RegistryName: aws.String(rs.Primary.ID),
		}
		resp, err := conn.DescribeRegistry(&params)
		if err != nil {
			return err
		}
		if resp == nil {
			return fmt.Errorf("Schemas Registry (%s) not found", n)
		}

		*v = *resp

		return nil
	}
}

func testAccCheckSchemasRegistryRecreated(i, j *schemas.DescribeRegistryOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.RegistryArn) == aws.StringValue(j.RegistryArn) {
			return fmt.Errorf("Schemas Registry not recreated")
		}
		return nil
	}
}

func testAccCheckSchemasRegistryNotRecreated(i, j *schemas.DescribeRegistryOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.RegistryArn) != aws.StringValue(j.RegistryArn) {
			return fmt.Errorf("Schemas Registry was recreated")
		}
		return nil
	}
}

func testAccAWSSchemasRegistryConfig(name, description string) string {
	return fmt.Sprintf(`
resource "aws_schemas_registry" "test" {
  name = %[1]q
  description = %[2]q
}
`, name, description)
}

func testAccAWSSchemasRegistryConfig_Tags1(name, key, value string) string {
	return fmt.Sprintf(`
resource "aws_schemas_registry" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, name, key, value)
}

func testAccAWSSchemasRegistryConfig_Tags2(name, key1, value1, key2, value2 string) string {
	return fmt.Sprintf(`
resource "aws_schemas_registry" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, name, key1, value1, key2, value2)
}
