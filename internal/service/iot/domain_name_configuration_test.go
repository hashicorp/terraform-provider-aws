package iot_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccIotDomainNameConfiguration_main(t *testing.T) {
	resourceName := "aws_iot_domain_name_configuration.test"
	domain := acctest.RandomDomainName()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccIotDomainNameConfig(domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "status", "ENABLED"),
				),
			},
		},
	})
}

func testAccIotDomainNameConfig(domain string) string {
	return acctest.ConfigCompose(fmt.Sprintf(`
resource "aws_iot_domain_name_configuration" "test" {
  name         = "test"
  domain_name  = "%[1]s"
  service_type = "DATA"
}
`, domain))
}
