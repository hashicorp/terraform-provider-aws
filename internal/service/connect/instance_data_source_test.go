package connect_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/connect"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccConnectInstanceDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("datasource-test-terraform")
	dataSourceName := "data.aws_connect_instance.test"
	resourceName := "aws_connect_instance.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, connect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccInstanceDataSourceConfig_nonExistentID,
				ExpectError: regexp.MustCompile(`error getting Connect Instance by instance_id`),
			},
			{
				Config:      testAccInstanceDataSourceConfig_nonExistentAlias,
				ExpectError: regexp.MustCompile(`error finding Connect Instance Summary by instance_alias`),
			},
			{
				Config: testAccInstanceDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "created_time", dataSourceName, "created_time"),
					resource.TestCheckResourceAttrPair(resourceName, "identity_management_type", dataSourceName, "identity_management_type"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_alias", dataSourceName, "instance_alias"),
					resource.TestCheckResourceAttrPair(resourceName, "inbound_calls_enabled", dataSourceName, "inbound_calls_enabled"),
					resource.TestCheckResourceAttrPair(resourceName, "outbound_calls_enabled", dataSourceName, "outbound_calls_enabled"),
					resource.TestCheckResourceAttrPair(resourceName, "contact_flow_logs_enabled", dataSourceName, "contact_flow_logs_enabled"),
					resource.TestCheckResourceAttrPair(resourceName, "contact_lens_enabled", dataSourceName, "contact_lens_enabled"),
					resource.TestCheckResourceAttrPair(resourceName, "auto_resolve_best_voices_enabled", dataSourceName, "auto_resolve_best_voices_enabled"),
					resource.TestCheckResourceAttrPair(resourceName, "early_media_enabled", dataSourceName, "early_media_enabled"),
					resource.TestCheckResourceAttrPair(resourceName, "status", dataSourceName, "status"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", dataSourceName, "service_role"),
				),
			},
			{
				Config: testAccInstanceDataSourceConfig_alias(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "created_time", dataSourceName, "created_time"),
					resource.TestCheckResourceAttrPair(resourceName, "identity_management_type", dataSourceName, "identity_management_type"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_alias", dataSourceName, "instance_alias"),
					resource.TestCheckResourceAttrPair(resourceName, "inbound_calls_enabled", dataSourceName, "inbound_calls_enabled"),
					resource.TestCheckResourceAttrPair(resourceName, "outbound_calls_enabled", dataSourceName, "outbound_calls_enabled"),
					resource.TestCheckResourceAttrPair(resourceName, "contact_flow_logs_enabled", dataSourceName, "contact_flow_logs_enabled"),
					resource.TestCheckResourceAttrPair(resourceName, "contact_lens_enabled", dataSourceName, "contact_lens_enabled"),
					resource.TestCheckResourceAttrPair(resourceName, "auto_resolve_best_voices_enabled", dataSourceName, "auto_resolve_best_voices_enabled"),
					resource.TestCheckResourceAttrPair(resourceName, "early_media_enabled", dataSourceName, "early_media_enabled"),
					resource.TestCheckResourceAttrPair(resourceName, "status", dataSourceName, "status"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", dataSourceName, "service_role"),
				),
			},
		},
	})
}

const testAccInstanceDataSourceConfig_nonExistentID = `
data "aws_connect_instance" "test" {
  instance_id = "97afc98d-101a-ba98-ab97-ae114fc115ec"
}
`

const testAccInstanceDataSourceConfig_nonExistentAlias = `
data "aws_connect_instance" "test" {
  instance_alias = "tf-acc-test-does-not-exist"
}
`

func testAccInstanceDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  instance_alias           = %[1]q
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  outbound_calls_enabled   = true
}

data "aws_connect_instance" "test" {
  instance_id = aws_connect_instance.test.id
}
`, rName)
}

func testAccInstanceDataSourceConfig_alias(rName string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  instance_alias           = %[1]q
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  outbound_calls_enabled   = true
}

data "aws_connect_instance" "test" {
  instance_alias = aws_connect_instance.test.instance_alias
}
`, rName)
}
