package mwaa_test

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/mwaa"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestMWAATokenDataSource_noMatchReturnsError(t *testing.T) {

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceResourceName := "data.aws_mwaa_cli_token.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, mwaa.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCliTokenDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceResourceName, "environment", rName),
					resource.TestCheckResourceAttrSet(dataSourceResourceName, "cli_token"),
					testAccCheckCliToken(dataSourceResourceName),
				),
			},
		},
	})
}

func testAccCheckCliToken(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		environment := rs.Primary.Attributes["environment"]

		verifier := &mwaa.CreateCliTokenInput{
			Name: aws.String(environment),
		}
		err := verifier.Validate()

		if err != nil {
			return fmt.Errorf("error verifying token for environment %q: %v", environment, err)
		}

		return nil
	}
	return nil
}

func testAccCliTokenDataSourceConfig_basic(environment string) string {
	return fmt.Sprintf(`
data "aws_mwaa_cli_token" "test" {
environment = "%s"
}
`, environment)
}
