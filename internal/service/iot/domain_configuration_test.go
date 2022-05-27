package iot_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccIotDomainConfiguration_basic(t *testing.T) {
	resourceName := "aws_iot_domain_configuration.test"
	domain := acctest.RandomDomainName()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfigurationConfig(domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "status", "ENABLED"),
				),
			},
		},
	})
}

func testAccDomainConfigurationConfig(domain string) string {
	return fmt.Sprintf(`
resource "aws_iot_domain_configuration" "test" {
  name         = "test"
  domain_name  = %[1]q
  service_type = "DATA"
}
`, domain)
}
