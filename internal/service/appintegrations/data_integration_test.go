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
