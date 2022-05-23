package appintegrations_test

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appintegrationsservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func testAccCheckDataIntegrationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AppIntegrationsConn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appintegrations_data_integration" {
			continue
		}

		input := &appintegrationsservice.GetDataIntegrationInput{
			Identifier: aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetDataIntegration(input)

		if err == nil {
			if aws.StringValue(resp.Id) == rs.Primary.ID {
				return fmt.Errorf("Data Integration '%s' was not deleted properly", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckDataIntegrationExists(name string, dataIntegration *appintegrationsservice.GetDataIntegrationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppIntegrationsConn
		input := &appintegrationsservice.GetDataIntegrationInput{
			Identifier: aws.String(rs.Primary.ID),
		}
		resp, err := conn.GetDataIntegration(input)

		if err != nil {
			return err
		}

		*dataIntegration = *resp

		return nil
	}
}
