package iot_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccIoTEndpointDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_iot_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "endpoint_address", regexp.MustCompile(fmt.Sprintf("^[a-z0-9]+(-ats)?.iot.%s.amazonaws.com$", acctest.Region()))),
				),
			},
		},
	})
}

func TestAccIoTEndpointDataSource_EndpointType_iotCredentialProvider(t *testing.T) {
	dataSourceName := "data.aws_iot_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointEndpointTypeConfig("iot:CredentialProvider"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "endpoint_address", regexp.MustCompile(fmt.Sprintf("^[a-z0-9]+.credentials.iot.%s.amazonaws.com$", acctest.Region()))),
				),
			},
		},
	})
}

func TestAccIoTEndpointDataSource_EndpointType_iotData(t *testing.T) {
	dataSourceName := "data.aws_iot_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointEndpointTypeConfig("iot:Data"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "endpoint_address", regexp.MustCompile(fmt.Sprintf("^[a-z0-9]+.iot.%s.amazonaws.com$", acctest.Region()))),
				),
			},
		},
	})
}

func TestAccIoTEndpointDataSource_EndpointType_iotDataATS(t *testing.T) {
	dataSourceName := "data.aws_iot_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointEndpointTypeConfig("iot:Data-ATS"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "endpoint_address", regexp.MustCompile(fmt.Sprintf("^[a-z0-9]+-ats.iot.%s.amazonaws.com$", acctest.Region()))),
				),
			},
		},
	})
}

func TestAccIoTEndpointDataSource_EndpointType_iotJobs(t *testing.T) {
	dataSourceName := "data.aws_iot_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointEndpointTypeConfig("iot:Jobs"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "endpoint_address", regexp.MustCompile(fmt.Sprintf("^[a-z0-9]+.jobs.iot.%s.amazonaws.com$", acctest.Region()))),
				),
			},
		},
	})
}

const testAccEndpointConfig = `
data "aws_iot_endpoint" "test" {}
`

func testAccEndpointEndpointTypeConfig(endpointType string) string {
	return fmt.Sprintf(`
data "aws_iot_endpoint" "test" {
  endpoint_type = %q
}
`, endpointType)
}
