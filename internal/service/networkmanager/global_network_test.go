package networkmanager_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	nm "github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/service/networkmanager"
)

func TestAccNetworkManagerGlobalNetwork_basic(t *testing.T) {
	var globalNetwork nm.GlobalNetwork
	resourceName := "aws_networkmanager_global_network.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckResourceGlobalNetworkDestroy,
		ErrorCheck:   acctest.ErrorCheck(t, nm.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccresourceGlobalNetworkConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceGlobalNetworkExists(resourceName, &globalNetwork),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "networkmanager", regexp.MustCompile(`global-network/global-network-.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
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

func TestAccNetworkManagerGlobalNetwork_tags(t *testing.T) {
	var globalNetwork nm.GlobalNetwork
	resourceName := "aws_networkmanager_global_network.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckResourceGlobalNetworkDestroy,
		ErrorCheck:   acctest.ErrorCheck(t, nm.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccresourceGlobalNetworkTagsConfig(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceGlobalNetworkExists(resourceName, &globalNetwork),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				Config:            testAccresourceGlobalNetworkTagsConfig(rName, "key1", "value1"),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccresourceGlobalNetworkTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceGlobalNetworkExists(resourceName, &globalNetwork),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				Config: testAccresourceGlobalNetworkConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceGlobalNetworkExists(resourceName, &globalNetwork),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccNetworkManagerGlobalNetwork_description(t *testing.T) {
	var globalNetwork nm.GlobalNetwork
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkmanager_global_network.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, nm.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckResourceGlobalNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccresourceGlobalNetworkDescriptionConfig(rName, "description1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceGlobalNetworkExists(resourceName, &globalNetwork),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccresourceGlobalNetworkDescriptionConfig(rName, "description2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceGlobalNetworkExists(resourceName, &globalNetwork),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func testAccCheckResourceGlobalNetworkDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkManagerConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_networkmanager_global_network" {
			continue
		}

		output, err := networkmanager.DescribeGlobalNetwork(conn, rs.Primary.ID)

		if tfawserr.ErrMessageContains(err, nm.ErrCodeResourceNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}

		if output == nil {
			continue
		}

		if aws.StringValue(output.State) != nm.GlobalNetworkStateDeleting {
			return fmt.Errorf("Network Manager Global Network (%s) still exists in non-deleted (%s) state", rs.Primary.ID, aws.StringValue(output.State))
		}
	}

	return nil
}

func testAccCheckResourceGlobalNetworkExists(resourceName string, globalNetwork *nm.GlobalNetwork) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Network Manager Global Network ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkManagerConn

		var err error
		globalNetwork, err = networkmanager.DescribeGlobalNetwork(conn, rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("Error describing Global Network: %s", err)
		}

		if globalNetwork == nil {
			return fmt.Errorf("Network Manager Global Network not found")
		}

		if aws.StringValue(globalNetwork.State) != nm.GlobalNetworkStateAvailable {
			return fmt.Errorf("Network Manager Global Network (%s) exists in non-available (%s) state", rs.Primary.ID, aws.StringValue(globalNetwork.State))
		}

		return nil

	}
}

func testAccresourceGlobalNetworkConfig() string {
	return `
resource "aws_networkmanager_global_network" "test" { }
`
}

func testAccresourceGlobalNetworkDescriptionConfig(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
	description = %[2]q

	tags = {
		Name = %[1]q
	}
}
`, rName, description)
}

func testAccresourceGlobalNetworkTagsConfig(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
	tags = {
		Name = %[1]q
		%[2]q = %[3]q
	}
}
`, rName, tagKey1, tagValue1)
}

func testAccresourceGlobalNetworkTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
	tags = {
		Name = %[1]q
		%[2]q = %[3]q
		%[4]q = %[5]q
	}
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
