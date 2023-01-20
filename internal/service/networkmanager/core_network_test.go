package networkmanager_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfnetworkmanager "github.com/hashicorp/terraform-provider-aws/internal/service/networkmanager"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccNetworkManagerCoreNetwork_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_core_network.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCoreNetworkDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkExists(ctx, resourceName),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "networkmanager", regexp.MustCompile(`core-network/core-network-.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestMatchResourceAttr(resourceName, "id", regexp.MustCompile(`core-network-.+`)),
					resource.TestCheckResourceAttr(resourceName, "state", networkmanager.CoreNetworkStateAvailable),
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

func TestAccNetworkManagerCoreNetwork_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_core_network.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCoreNetworkDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfnetworkmanager.ResourceCoreNetwork(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccNetworkManagerCoreNetwork_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_core_network.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCoreNetworkDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkConfig_tags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkExists(ctx, resourceName),
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
				Config: testAccCoreNetworkConfig_tags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccCoreNetworkConfig_tags1("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccNetworkManagerCoreNetwork_description(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_core_network.test"
	originalDescription := "description1"
	updatedDescription := "description2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCoreNetworkDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkConfig_description(originalDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", originalDescription),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCoreNetworkConfig_description(updatedDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", updatedDescription),
				),
			},
		},
	})
}

func TestAccNetworkManagerCoreNetwork_policyDocument(t *testing.T) {
	ctx := acctest.Context(t)
	client := acctest.Provider.Meta().(*conns.AWSClient)
	resourceName := "aws_networkmanager_core_network.test"
	originalSegmentValue := "segmentValue1"
	updatedSegmentValue := "segmentValue2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCoreNetworkDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkConfig_policyDocument(originalSegmentValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkExists(ctx, resourceName),
					testAccCheckPolicyDocument(resourceName, originalSegmentValue),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "edges.*", map[string]string{
						"asn":                  "65022",
						"edge_location":        client.Region,
						"inside_cidr_blocks.#": "0",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "segments.*", map[string]string{
						"edge_locations.#":  "1",
						"edge_locations.0":  client.Region,
						"name":              originalSegmentValue,
						"shared_segments.#": "0",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCoreNetworkConfig_policyDocument(updatedSegmentValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkExists(ctx, resourceName),
					testAccCheckPolicyDocument(resourceName, updatedSegmentValue),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "edges.*", map[string]string{
						"asn":                  "65022",
						"edge_location":        client.Region,
						"inside_cidr_blocks.#": "0",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "segments.*", map[string]string{
						"edge_locations.#":  "1",
						"edge_locations.0":  client.Region,
						"name":              updatedSegmentValue,
						"shared_segments.#": "0",
					}),
				),
			},
		},
	})
}

func testAccCheckPolicyDocument(n, segmentValue string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		client := acctest.Provider.Meta().(*conns.AWSClient)

		policyDocumentTarget := fmt.Sprintf("{\"core-network-configuration\":{\"asn-ranges\":[\"65022-65534\"],\"edge-locations\":[{\"location\":\"%s\"}],\"vpn-ecmp-support\":true},\"segments\":[{\"isolate-attachments\":false,\"name\":\"%s\",\"require-attachment-acceptance\":true}],\"version\":\"2021.12\"}", client.Region, segmentValue)

		if rs.Primary.Attributes["policy_document"] != policyDocumentTarget {
			return fmt.Errorf("Expected policy_document: %s, given %s", policyDocumentTarget, rs.Primary.Attributes["policy_document"])
		}

		return nil
	}
}

func testAccCheckCoreNetworkDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkManagerConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networkmanager_core_network" {
				continue
			}

			_, err := tfnetworkmanager.FindCoreNetworkByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Network Manager Core Network %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckCoreNetworkExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Network Manager Core Network ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkManagerConn()

		_, err := tfnetworkmanager.FindCoreNetworkByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCoreNetworkConfig_basic() string {
	return `
resource "aws_networkmanager_global_network" "test" {}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
}`
}

func testAccCoreNetworkConfig_tags1(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_networkmanager_global_network" "test" {}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccCoreNetworkConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_networkmanager_global_network" "test" {}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccCoreNetworkConfig_description(description string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_networkmanager_global_network" "test" {}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
  description       = %[1]q
}
`, description)
}

func testAccCoreNetworkConfig_policyDocument(segmentValue string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_networkmanager_global_network" "test" {}

data "aws_networkmanager_core_network_policy_document" "test" {
  core_network_configuration {
    asn_ranges = ["65022-65534"]

    edge_locations {
      location = data.aws_region.current.name
    }
  }

  segments {
    name = %[1]q
  }
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
  policy_document   = data.aws_networkmanager_core_network_policy_document.test.json
}
`, segmentValue)
}
