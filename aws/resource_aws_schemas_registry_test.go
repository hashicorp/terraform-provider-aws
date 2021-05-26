package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/schemas"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/schemas/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
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
	var v schemas.DescribeRegistryOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_schemas_registry.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(schemas.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, schemas.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSchemasRegistryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSchemasRegistryConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasRegistryExists(resourceName, &v),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "schemas", fmt.Sprintf("registry/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func TestAccAWSSchemasRegistry_disappears(t *testing.T) {
	var v schemas.DescribeRegistryOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_schemas_registry.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(schemas.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, schemas.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSchemasRegistryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSchemasRegistryConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasRegistryExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSchemasRegistry(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSSchemasRegistry_Description(t *testing.T) {
	var v schemas.DescribeRegistryOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_schemas_registry.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(schemas.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, schemas.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSchemasRegistryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSchemasRegistryConfigDescription(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasRegistryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSchemasRegistryConfigDescription(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasRegistryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
			{
				Config: testAccAWSSchemasRegistryConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasRegistryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
		},
	})
}

func TestAccAWSSchemasRegistry_Tags(t *testing.T) {
	var v schemas.DescribeRegistryOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_schemas_registry.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(schemas.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, schemas.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSchemasRegistryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSchemasRegistryConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasRegistryExists(resourceName, &v),
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
				Config: testAccAWSSchemasRegistryConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasRegistryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSSchemasRegistryConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasRegistryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
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

		_, err := finder.RegistryByName(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EventBridge Schemas Registry %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckSchemasRegistryExists(n string, v *schemas.DescribeRegistryOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EventBridge Schemas Registry ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).schemasconn

		output, err := finder.RegistryByName(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAWSSchemasRegistryConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_schemas_registry" "test" {
  name = %[1]q
}
`, rName)
}

func testAccAWSSchemasRegistryConfigDescription(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_schemas_registry" "test" {
  name        = %[1]q
  description = %[2]q
}
`, rName, description)
}

func testAccAWSSchemasRegistryConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_schemas_registry" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSSchemasRegistryConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_schemas_registry" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
