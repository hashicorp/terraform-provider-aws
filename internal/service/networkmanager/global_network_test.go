package networkmanager

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func init() {
	resource.AddTestSweepers("aws_networkmanager_global_network", &resource.Sweeper{
		Name: "aws_networkmanager_global_network",
		F:    testSweepGlobalNetwork,
	})
}

func testSweepGlobalNetwork(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).networkmanagerconn
	input := &networkmanager.DescribeGlobalNetworksInput{}
	var sweeperErrs *multierror.Error

	err = conn.DescribeGlobalNetworksPages(input,
		func(page *networkmanager.DescribeGlobalNetworksOutput, lastPage bool) bool {
			for _, globalNetwork := range page.GlobalNetworks {
				input := &networkmanager.DeleteGlobalNetworkInput{
					GlobalNetworkId: globalNetwork.GlobalNetworkId,
				}
				id := aws.StringValue(globalNetwork.GlobalNetworkId)

				log.Printf("[INFO] Deleting Network Manager Global Network: %s", id)
				_, err := conn.DeleteGlobalNetwork(input)

				if tfawserr.ErrCodeEquals(err, "InvalidGlobalNetworkID.NotFound", "") {
					continue
				}

				if err != nil {
					sweeperErr := fmt.Errorf("failed to delete Network Manager Global Network %s: %s", id, err)
					log.Printf("[ERROR] %s", sweeperErr)
					sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
					continue
				}

				if err := waitForGlobalNetworkDeletion(conn, id); err != nil {
					sweeperErr := fmt.Errorf("error waiting for Network Manager Global Network (%s) deletion: %s", id, err)
					log.Printf("[ERROR] %s", sweeperErr)
					sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
					continue
				}
			}
			return !lastPage
		})
	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Network Manager Global Network sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error retrieving Network Manager Global Networks: %s", err)
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccGlobalNetwork_basic(t *testing.T) {
	resourceName := "aws_networkmanager_global_network.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsGlobalNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalNetworkConfig("test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGlobalNetworkExists(resourceName),
					testAccMatchResourceAttrGlobalARN(resourceName, "arn", "networkmanager", regexp.MustCompile(`global-network/global-network-.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGlobalNetworkConfig_Update("test updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGlobalNetworkExists(resourceName),
					testAccMatchResourceAttrGlobalARN(resourceName, "arn", "networkmanager", regexp.MustCompile(`global-network/global-network-.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", "test updated"),
				),
			},
		},
	})
}

func TestAccGlobalNetwork_tags(t *testing.T) {
	resourceName := "aws_networkmanager_global_network.test"
	description := "test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsGlobalNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalNetworkConfigTags1(description, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGlobalNetworkExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
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
				Config: testAccGlobalNetworkConfigTags2(description, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGlobalNetworkExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccGlobalNetworkConfigTags1(description, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGlobalNetworkExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAwsGlobalNetworkDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).networkmanagerconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_networkmanager_global_network" {
			continue
		}

		globalNetwork, err := networkmanagerDescribeGlobalNetwork(conn, rs.Primary.ID)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, networkmanager.ErrCodeResourceNotFoundException, "") {
				return nil
			}
			return err
		}

		if globalNetwork == nil {
			continue
		}

		return fmt.Errorf("Expected Global Network to be destroyed, %s found", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAwsGlobalNetworkExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).networkmanagerconn

		globalNetwork, err := networkmanagerDescribeGlobalNetwork(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if globalNetwork == nil {
			return fmt.Errorf("Network Manager Global Network not found")
		}

		if aws.StringValue(globalNetwork.State) != networkmanager.GlobalNetworkStateAvailable && aws.StringValue(globalNetwork.State) != networkmanager.GlobalNetworkStatePending {
			return fmt.Errorf("Network Manager Global Network (%s) exists in (%s) state", rs.Primary.ID, aws.StringValue(globalNetwork.State))
		}

		return err
	}
}

func testAccGlobalNetworkConfig(description string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
  description = %q
}
`, description)
}

func testAccGlobalNetworkConfigTags1(description, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
  description = %q

  tags = {
  	%q = %q
  }
}
`, description, tagKey1, tagValue1)
}

func testAccGlobalNetworkConfigTags2(description, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
 description = %q

  tags = {
  	%q = %q
	%q = %q
  }
}
`, description, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccGlobalNetworkConfig_Update(description string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
  description = %q
}
`, description)
}
