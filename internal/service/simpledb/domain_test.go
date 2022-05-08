package simpledb_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/simpledb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccSimpleDBDomain_basic(t *testing.T) {
	resourceName := "aws_simpledb_domain.test_domain"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(simpledb.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, simpledb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName),
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

func testAccCheckDomainDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SimpleDBConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_simpledb_domain" {
			continue
		}

		input := &simpledb.DomainMetadataInput{
			DomainName: aws.String(rs.Primary.ID),
		}
		_, err := conn.DomainMetadata(input)
		if err == nil {
			return fmt.Errorf("Domain exists when it should be destroyed!")
		}

		// Verify the error is an API error, not something else
		_, ok := err.(awserr.Error)
		if !ok {
			return err
		}
	}

	return nil
}

func testAccCheckDomainExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SimpleDB domain with that name exists")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SimpleDBConn
		input := &simpledb.DomainMetadataInput{
			DomainName: aws.String(rs.Primary.ID),
		}
		_, err := conn.DomainMetadata(input)
		return err
	}
}

var testAccDomainConfig = `
resource "aws_simpledb_domain" "test_domain" {
  name = "terraform-test-domain"
}
`
