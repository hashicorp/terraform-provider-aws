package servicequotas_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/servicequotas"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccServiceQuotasServiceDataSource_serviceName(t *testing.T) {
	dataSourceName := "data.aws_servicequotas_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(servicequotas.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, servicequotas.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDataSourceConfig_name("Amazon Virtual Private Cloud (Amazon VPC)"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "service_code", "vpc"),
				),
			},
		},
	})
}

func testAccServiceDataSourceConfig_name(serviceName string) string {
	return fmt.Sprintf(`
data "aws_servicequotas_service" "test" {
  service_name = %[1]q
}
`, serviceName)
}
