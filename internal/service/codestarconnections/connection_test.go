package codestarconnections_test

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codestarconnections"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcodestarconnections "github.com/hashicorp/terraform-provider-aws/internal/service/codestarconnections"
)

func TestAccCodeStarConnectionsConnection_basic(t *testing.T) {
	var v codestarconnections.Connection
	resourceName := "aws_codestarconnections_connection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(codestarconnections.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, codestarconnections.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionBasicConfig(rName),
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
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(codestarconnections.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, codestarconnections.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionHostARNConfig(rName),
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
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(codestarconnections.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, codestarconnections.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionBasicConfig(rName),
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
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(codestarconnections.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, codestarconnections.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionTags1Config(rName, "key1", "value1"),
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
				Config: testAccConnectionTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccConnectionTags1Config(rName, "key2", "value2"),
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
			return errors.New("No CodeStar connection ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeStarConnectionsConn

		resp, err := conn.GetConnection(&codestarconnections.GetConnectionInput{
			ConnectionArn: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*v = *resp.Connection

		return nil
	}
}

func testAccCheckConnectionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CodeStarConnectionsConn

	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "aws_codestarconnections_connection":
			_, err := conn.GetConnection(&codestarconnections.GetConnectionInput{
				ConnectionArn: aws.String(rs.Primary.ID),
			})

			if err != nil && !tfawserr.ErrMessageContains(err, codestarconnections.ErrCodeResourceNotFoundException, "") {
				return err
			}
		}
	}

	return nil
}

func testAccConnectionBasicConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_codestarconnections_connection" "test" {
  name          = %[1]q
  provider_type = "Bitbucket"
}
`, rName)
}

func testAccConnectionHostARNConfig(rName string) string {
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

func testAccConnectionTags1Config(rName string, tagKey1 string, tagValue1 string) string {
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

func testAccConnectionTags2Config(rName string, tagKey1 string, tagValue1 string, tagKey2 string, tagValue2 string) string {
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
