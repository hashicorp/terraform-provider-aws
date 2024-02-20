// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIoTEndpointDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_iot_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "endpoint_address", regexache.MustCompile(fmt.Sprintf("^[0-9a-z]+(-ats)?.iot.%s.amazonaws.com$", acctest.Region()))),
				),
			},
		},
	})
}

func TestAccIoTEndpointDataSource_EndpointType_iotCredentialProvider(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_iot_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointDataSourceConfig_type("iot:CredentialProvider"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "endpoint_address", regexache.MustCompile(fmt.Sprintf("^[0-9a-z]+.credentials.iot.%s.amazonaws.com$", acctest.Region()))),
				),
			},
		},
	})
}

func TestAccIoTEndpointDataSource_EndpointType_iotData(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_iot_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointDataSourceConfig_type("iot:Data"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "endpoint_address", regexache.MustCompile(fmt.Sprintf("^[0-9a-z]+.iot.%s.amazonaws.com$", acctest.Region()))),
				),
			},
		},
	})
}

func TestAccIoTEndpointDataSource_EndpointType_iotDataATS(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_iot_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointDataSourceConfig_type("iot:Data-ATS"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "endpoint_address", regexache.MustCompile(fmt.Sprintf("^[0-9a-z]+-ats.iot.%s.amazonaws.com$", acctest.Region()))),
				),
			},
		},
	})
}

func TestAccIoTEndpointDataSource_EndpointType_iotJobs(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_iot_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointDataSourceConfig_type("iot:Jobs"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "endpoint_address", regexache.MustCompile(fmt.Sprintf("^[0-9a-z]+.jobs.iot.%s.amazonaws.com$", acctest.Region()))),
				),
			},
		},
	})
}

const testAccEndpointDataSourceConfig_basic = `
data "aws_iot_endpoint" "test" {}
`

func testAccEndpointDataSourceConfig_type(endpointType string) string {
	return fmt.Sprintf(`
data "aws_iot_endpoint" "test" {
  endpoint_type = %q
}
`, endpointType)
}
