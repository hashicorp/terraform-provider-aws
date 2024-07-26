// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transfer_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccServerDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_transfer_server.test"
	datasourceName := "data.aws_transfer_server.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrDomain, resourceName, names.AttrDomain),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrEndpoint, resourceName, names.AttrEndpoint),
					resource.TestCheckResourceAttrPair(datasourceName, "identity_provider_type", resourceName, "identity_provider_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "logging_role", resourceName, "logging_role"),
					resource.TestCheckResourceAttrPair(datasourceName, "structured_log_destinations.#", resourceName, "structured_log_destinations.#"),
				),
			},
		},
	})
}

func testAccServerDataSource_Service_managed(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandString(5)
	resourceName := "aws_transfer_server.test"
	datasourceName := "data.aws_transfer_server.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerDataSourceConfig_serviceManaged(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrCertificate, resourceName, names.AttrCertificate),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrEndpoint, resourceName, names.AttrEndpoint),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrEndpointType, resourceName, names.AttrEndpointType),
					resource.TestCheckResourceAttrPair(datasourceName, "identity_provider_type", resourceName, "identity_provider_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "invocation_role", resourceName, "invocation_role"),
					resource.TestCheckResourceAttrPair(datasourceName, "logging_role", resourceName, "logging_role"),
					resource.TestCheckResourceAttrPair(datasourceName, "protocols.#", resourceName, "protocols.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "security_policy_name", resourceName, "security_policy_name"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrURL, resourceName, names.AttrURL),
				),
			},
		},
	})
}

func testAccServerDataSource_apigateway(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandString(5)
	resourceName := "aws_transfer_server.test"
	datasourceName := "data.aws_transfer_server.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerDataSourceConfig_apigateway(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrEndpoint, resourceName, names.AttrEndpoint),
					resource.TestCheckResourceAttrPair(datasourceName, "identity_provider_type", resourceName, "identity_provider_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "invocation_role", resourceName, "invocation_role"),
					resource.TestCheckResourceAttrPair(datasourceName, "logging_role", resourceName, "logging_role"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrURL, resourceName, names.AttrURL),
				),
			},
		},
	})
}

const testAccServerDataSourceConfig_basic = `
resource "aws_transfer_server" "test" {}

data "aws_transfer_server" "test" {
  server_id = aws_transfer_server.test.id
}
`

func testAccServerDataSourceConfig_serviceManaged(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = "tf-test-transfer-server-iam-role-%[1]s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "transfer.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = "tf-test-transfer-server-iam-policy-%[1]s"
  role = aws_iam_role.test.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowFullAccesstoCloudWatchLogs",
      "Effect": "Allow",
      "Action": [
        "logs:*"
      ],
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_transfer_server" "test" {
  identity_provider_type = "SERVICE_MANAGED"
  logging_role           = aws_iam_role.test.arn
}

data "aws_transfer_server" "test" {
  server_id = aws_transfer_server.test.id
}
`, rName)
}

func testAccServerDataSourceConfig_apigateway(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = "test"
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "test"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_resource.test.id
  http_method   = "GET"
  authorization = "NONE"
}

resource "aws_api_gateway_method_response" "error" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method
  status_code = "400"
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method

  type                    = "HTTP"
  uri                     = "https://www.google.de"
  integration_http_method = "GET"
}

resource "aws_api_gateway_integration_response" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_integration.test.http_method
  status_code = aws_api_gateway_method_response.error.status_code
}

resource "aws_api_gateway_deployment" "test" {
  depends_on = [aws_api_gateway_integration.test]

  rest_api_id       = aws_api_gateway_rest_api.test.id
  stage_name        = "test"
  description       = "%[1]s"
  stage_description = "%[1]s"

  variables = {
    "a" = "2"
  }
}

resource "aws_iam_role" "test" {
  name = "tf-test-transfer-server-iam-role-for-apigateway-%[1]s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "transfer.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = "tf-test-transfer-server-iam-policy-%[1]s"
  role = aws_iam_role.test.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowFullAccesstoCloudWatchLogs",
      "Effect": "Allow",
      "Action": [
        "logs:*"
      ],
      "Resource": "*"
    }
  ]
}
POLICY
}

data "aws_region" "current" {}

resource "aws_transfer_server" "test" {
  identity_provider_type = "API_GATEWAY"
  url                    = "https://${aws_api_gateway_rest_api.test.id}.execute-api.${data.aws_region.current.name}.amazonaws.com${aws_api_gateway_resource.test.path}"
  invocation_role        = aws_iam_role.test.arn
  logging_role           = aws_iam_role.test.arn
}

data "aws_transfer_server" "test" {
  server_id = aws_transfer_server.test.id
}
`, rName)
}
