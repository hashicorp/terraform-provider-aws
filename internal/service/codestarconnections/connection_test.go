package codestarconnections_test

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/codestarconnections"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcodestarconnections "github.com/hashicorp/terraform-provider-aws/internal/service/codestarconnections"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccCodeStarConnectionsConnection_basic(t *testing.T) {
	var v codestarconnections.Connection
	resourceName := "aws_codestarconnections_connection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(codestarconnections.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, codestarconnections.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "id", "codestar-connections", regexp.MustCompile("connection/.+")),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "codestar-connections", regexp.MustCompile("connection/.+")),
					resource.TestCheckResourceAttr(resourceName, "provider_type", codestarconnections.ProviderTypeBitbucket),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "connection_status", codestarconnections.ConnectionStatusPending),
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

func TestAccCodeStarConnectionsConnection_hostARN(t *testing.T) {
	var v codestarconnections.Connection
	resourceName := "aws_codestarconnections_connection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(codestarconnections.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, codestarconnections.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_hostARN(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "id", "codestar-connections", regexp.MustCompile("connection/.+")),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "codestar-connections", regexp.MustCompile("connection/.+")),
					acctest.MatchResourceAttrRegionalARN(resourceName, "host_arn", "codestar-connections", regexp.MustCompile("host/.+")),
					resource.TestCheckResourceAttr(resourceName, "provider_type", codestarconnections.ProviderTypeGitHubEnterpriseServer),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "connection_status", codestarconnections.ConnectionStatusPending),
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

func TestAccCodeStarConnectionsConnection_disappears(t *testing.T) {
	var v codestarconnections.Connection
	resourceName := "aws_codestarconnections_connection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(codestarconnections.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, codestarconnections.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfcodestarconnections.ResourceConnection(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCodeStarConnectionsConnection_tags(t *testing.T) {
	var v codestarconnections.Connection
	resourceName := "aws_codestarconnections_connection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(codestarconnections.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, codestarconnections.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionExists(resourceName, &v),
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
				Config: testAccConnectionConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccConnectionConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckConnectionExists(n string, v *codestarconnections.Connection) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No CodeStar Connections Connection ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeStarConnectionsConn

		output, err := tfcodestarconnections.FindConnectionByARN(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckConnectionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CodeStarConnectionsConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_codestarconnections_connection" {
			continue
		}

		_, err := tfcodestarconnections.FindConnectionByARN(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("CodeStar Connections Connection %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccConnectionConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_codestarconnections_connection" "test" {
  name          = %[1]q
  provider_type = "Bitbucket"
}
`, rName)
}

func testAccConnectionConfig_hostARN(rName string) string {
	return fmt.Sprintf(`
resource "aws_codestarconnections_host" "test" {
  name              = %[1]q
  provider_endpoint = "https://example.com"
  provider_type     = "GitHubEnterpriseServer"
}

resource "aws_codestarconnections_connection" "test" {
  name     = %[1]q
  host_arn = aws_codestarconnections_host.test.arn
}
`, rName)
}

func testAccConnectionConfig_tags1(rName string, tagKey1 string, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_codestarconnections_connection" "test" {
  name          = %[1]q
  provider_type = "Bitbucket"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccConnectionConfig_tags2(rName string, tagKey1 string, tagValue1 string, tagKey2 string, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_codestarconnections_connection" "test" {
  name          = %[1]q
  provider_type = "Bitbucket"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
