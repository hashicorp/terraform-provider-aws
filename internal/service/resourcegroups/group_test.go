package resourcegroups_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/resourcegroups"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccResourceGroupsGroup_Resource_basic(t *testing.T) {
	var v resourcegroups.Group
	resourceName := "aws_resourcegroups_group.test"
	n := fmt.Sprintf("test-group-%d", sdkacctest.RandInt())

	desc1 := "Hello World"
	desc2 := "Foo Bar"

	query2 := `{
  "ResourceTypeFilters": [
    "AWS::EC2::Instance"
  ],
  "TagFilters": [
    {
      "Key": "Hello",
      "Values": [
        "World"
      ]
    }
  ]
}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, resourcegroups.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_basic(n, desc1, testAccResourceGroupQueryConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", n),
					resource.TestCheckResourceAttr(resourceName, "description", desc1),
					resource.TestCheckResourceAttr(resourceName, "resource_query.0.query", testAccResourceGroupQueryConfig+"\n"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGroupConfig_basic(n, desc2, query2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", desc2),
					resource.TestCheckResourceAttr(resourceName, "resource_query.0.query", query2+"\n"),
				),
			},
		},
	})
}

func TestAccResourceGroupsGroup_Resource_Configuration(t *testing.T) {
	var v resourcegroups.Group
	resourceName := "aws_resourcegroups_group.test"
	n := fmt.Sprintf("test-group-%d", sdkacctest.RandInt())

	desc1 := "Hello World"
	desc2 := "Foo Bar"
	configType1 := "AWS::EC2::HostManagement"
	configType2 := "AWS::ResourceGroups::Generic"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, resourcegroups.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_config(n, desc1, configType1, configType2, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", n),
					resource.TestCheckResourceAttr(resourceName, "description", desc1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.type", configType1),
					resource.TestCheckResourceAttr(resourceName, "configuration.1.type", configType2),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.parameters.#", "4"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.parameters.0.name", "allowed-host-families"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.parameters.0.values.0", "mac1"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Check that changing the auto-allocate value is represented
			{
				Config: testAccGroupConfig_config(n, desc1, configType1, configType2, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", n),
					resource.TestCheckResourceAttr(resourceName, "description", desc1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.type", configType1),
					resource.TestCheckResourceAttr(resourceName, "configuration.1.type", configType2),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.parameters.#", "4"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.parameters.2.name", "auto-allocate-host"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.parameters.2.values.0", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
				),
			},
			{
				Config: testAccGroupConfig_config(n, desc2, configType1, configType2, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", desc2),
				),
			},
			// Check that trying to change the configuration group to a resource-query group fails
			{
				Config:      testAccGroupConfig_basic(n, desc1, testAccResourceGroupQueryConfig),
				ExpectError: regexp.MustCompile(`conversion between resource-query and configuration group types is not possible`),
			},
		},
	})
}

func TestAccResourceGroupsGroup_Resource_tags(t *testing.T) {
	var v resourcegroups.Group
	resourceName := "aws_resourcegroups_group.test"
	n := fmt.Sprintf("test-group-%d", sdkacctest.RandInt())
	desc1 := "Hello World"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, resourcegroups.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_tags1(n, desc1, testAccResourceGroupQueryConfig, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceGroupExists(resourceName, &v),
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
				Config: testAccGroupConfig_tags2(n, desc1, testAccResourceGroupQueryConfig, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccGroupConfig_tags1(n, desc1, testAccResourceGroupQueryConfig, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckResourceGroupExists(n string, v *resourcegroups.Group) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No resource group name is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ResourceGroupsConn

		resp, err := conn.GetGroup(&resourcegroups.GetGroupInput{
			GroupName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if *resp.Group.Name == rs.Primary.ID {
			*v = *resp.Group
			return nil
		}

		return fmt.Errorf("Resource Group (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckResourceGroupDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_resourcegroups_group" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ResourceGroupsConn
		resp, err := conn.GetGroup(&resourcegroups.GetGroupInput{
			GroupName: aws.String(rs.Primary.ID),
		})

		if err == nil {
			if *resp.Group.Name == rs.Primary.ID {
				return fmt.Errorf("Resource Group %s still exists", rs.Primary.ID)
			}
		}

		if tfawserr.ErrCodeEquals(err, resourcegroups.ErrCodeNotFoundException) {
			return nil
		}

		return err
	}

	return nil
}

const testAccResourceGroupQueryConfig = `{
  "ResourceTypeFilters": [
    "AWS::EC2::Instance"
  ],
  "TagFilters": [
    {
      "Key": "Stage",
      "Values": [
        "Test"
      ]
    }
  ]
}`

func testAccGroupConfig_basic(rName, desc, query string) string {
	return fmt.Sprintf(`
resource "aws_resourcegroups_group" "test" {
  name        = "%s"
  description = "%s"

  resource_query {
    query = <<JSON
%s
JSON

  }
}
`, rName, desc, query)
}

func testAccGroupConfig_tags1(rName, desc, query, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_resourcegroups_group" "test" {
  name        = "%s"
  description = "%s"

  resource_query {
    query = <<JSON
%s
JSON

  }

  tags = {
    %q = %q
  }
}
`, rName, desc, query, tag1Key, tag1Value)
}

func testAccGroupConfig_tags2(rName, desc, query, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_resourcegroups_group" "test" {
  name        = "%s"
  description = "%s"

  resource_query {
    query = <<JSON
%s
JSON

  }

  tags = {
    %q = %q
    %q = %q
  }
}
`, rName, desc, query, tag1Key, tag1Value, tag2Key, tag2Value)
}

func testAccGroupConfig_config(rName, desc, cType1, cType2 string, autoAllocateHost bool) string {
	return fmt.Sprintf(`
resource "aws_resourcegroups_group" "test" {
  name        = "%s"
  description = "%s"

  configuration {
    type = "%s"

    parameters {
      name   = "allowed-host-families"
      values = [
        "mac1",
      ]
    }

    parameters {
      name   = "any-host-based-license-configuration"
      values = [
        "true",
      ]
    }

    parameters {
      name   = "auto-allocate-host"
      values = [
        "%t",
      ]
    }

    parameters {
      name   = "auto-release-host"
      values = [
        "true",
      ]
    }
  }

  configuration {
    type = "%s"

    parameters {
      name = "allowed-resource-types"
      values = [
        "AWS::EC2::Host",
      ]
    }

    parameters {
      name = "deletion-protection"
      values = [
        "UNLESS_EMPTY"
      ]
    }
  }
}
`, rName, desc, cType1, autoAllocateHost, cType2)
}
