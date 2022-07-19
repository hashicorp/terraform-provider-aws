package apprunner_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/apprunner"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapprunner "github.com/hashicorp/terraform-provider-aws/internal/service/apprunner"
)

func TestAccAppRunnerConnection_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apprunner.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "apprunner", regexp.MustCompile(fmt.Sprintf(`connection/%s/.+`, rName))),
					resource.TestCheckResourceAttr(resourceName, "connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "provider_type", apprunner.ProviderTypeGithub),
					resource.TestCheckResourceAttr(resourceName, "status", apprunner.ConnectionStatusPendingHandshake),
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

func TestAccAppRunnerConnection_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apprunner.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfapprunner.ResourceConnection(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAppRunnerConnection_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apprunner.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConnectionConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key2", "value2"),
				),
			},
			{
				Config: testAccConnectionConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckConnectionDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_apprunner_connection" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppRunnerConn

		connection, err := tfapprunner.FindConnectionSummaryByName(context.Background(), conn, rs.Primary.ID)

		if tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		if connection != nil {
			return fmt.Errorf("App Runner Connection (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckConnectionExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No App Runner Connection ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppRunnerConn

		connection, err := tfapprunner.FindConnectionSummaryByName(context.Background(), conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if connection == nil {
			return fmt.Errorf("App Runner Connection (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccConnectionConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_connection" "test" {
  connection_name = %q
  provider_type   = "GITHUB"
}
`, rName)
}

func testAccConnectionConfig_tags1(rName string, tagKey1 string, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_connection" "test" {
  connection_name = %[1]q
  provider_type   = "GITHUB"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccConnectionConfig_tags2(rName string, tagKey1 string, tagValue1 string, tagKey2 string, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_connection" "test" {
  connection_name = %[1]q
  provider_type   = "GITHUB"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
