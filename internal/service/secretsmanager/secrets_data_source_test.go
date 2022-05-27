package secretsmanager_test

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/secretsmanager"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccSecretsManagerSecretsDataSource_filter(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret.test"
	dataSourceName := "data.aws_secretsmanager_secrets.test"

	propagationSleep := func() resource.TestCheckFunc {
		return func(s *terraform.State) error {
			log.Print("[DEBUG] Test: Sleep to allow secrets become visible in the list.")
			time.Sleep(30 * time.Second)
			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, secretsmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecretDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretsDataSourceConfig_filter2(rName),
				Check:  propagationSleep(),
			},
			{
				Config: testAccSecretsDataSourceConfig_filter(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "names.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "arns.0", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "names.0", resourceName, "name"),
				),
			},
		},
	})
}

func testAccSecretsDataSourceConfig_filter2(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}
`, rName)
}

func testAccSecretsDataSourceConfig_filter(rName string) string {
	return acctest.ConfigCompose(
		testAccSecretsDataSourceConfig_filter2(rName),
		`
data "aws_secretsmanager_secrets" "test" {
  filter {
    name   = "name"
    values = [aws_secretsmanager_secret.test.name]
  }
}
`)
}
