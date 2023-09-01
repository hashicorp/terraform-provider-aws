// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/connect"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccInstanceDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("datasource-test-terraform")
	dataSourceName := "data.aws_connect_instance.test"
	resourceName := "aws_connect_instance.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, connect.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
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
					resource.TestCheckResourceAttrPair(resourceName, "multi_party_conference_enabled", dataSourceName, "multi_party_conference_enabled"),
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
					resource.TestCheckResourceAttrPair(resourceName, "multi_party_conference_enabled", dataSourceName, "multi_party_conference_enabled"),
					resource.TestCheckResourceAttrPair(resourceName, "status", dataSourceName, "status"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", dataSourceName, "service_role"),
				),
			},
		},
	})
}

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
