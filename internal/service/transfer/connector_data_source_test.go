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

func TestAccTransferConnectorDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_transfer_connector.test"
	resourceName := "aws_transfer_connector.test"
	url := "http://www.example.com"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TransferEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectorDataSourceConfig_basic(rName, url),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "access_role", resourceName, "access_role"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "as2_config.#", resourceName, "as2_config.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(dataSourceName, "service_managed_egress_ip_addresses.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "sftp_config.#", resourceName, "sftp_config.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrTags, resourceName, names.AttrTags),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrURL, resourceName, names.AttrURL),
				),
			},
		},
	})
}

func TestAccTransferConnectorDataSource_egressConfig(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_transfer_connector.test"
	resourceName := "aws_transfer_connector.test"
	publicKey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDNt3kA/dBkS6ZyU/sVDiGMuWJQaRPmLNbs/25K/e/fIl07ZWUgqqsFkcycLLMNFGD30Cmgp6XCXfNlIjzFWhNam+4cBb4DPpvieUw44VgsHK5JQy3JKlUfglmH5rs4G5pLiVfZpFU6jqvTsu4mE1CHCP0sXJlJhGxMG3QbsqYWNKiqGFEhuzGMs6fQlMkNiXsFoDmh33HAcXCbaFSC7V7xIqT1hlKu0iOL+GNjMj4R3xy0o3jafhO4MG2s3TwCQQCyaa5oyjL8iP8p3L9yp6cbIcXaS72SIgbCSGCyrcQPIKP2lJJHvE1oVWzLVBhR4eSzrlFDv7K4IErzaJmHqdiz" // nosemgrep:ci.ssh-key

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TransferEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectorDataSourceConfig_egressConfig(rName, publicKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "access_role", resourceName, "access_role"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "egress_config.#", resourceName, "egress_config.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "egress_config.0.vpc_lattice.#", resourceName, "egress_config.0.vpc_lattice.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "egress_config.0.vpc_lattice.0.resource_configuration_arn", resourceName, "egress_config.0.vpc_lattice.0.resource_configuration_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "egress_config.0.vpc_lattice.0.port_number", resourceName, "egress_config.0.vpc_lattice.0.port_number"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, "sftp_config.#", resourceName, "sftp_config.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrTags, resourceName, names.AttrTags),
				),
			},
		},
	})
}

func testAccConnectorDataSourceConfig_basic(rName, url string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "Service": "transfer.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }]
 }
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<POLICY
{
  "Version":"2012-10-17",
  "Statement":[{
    "Sid":"AllowFullAccesstoS3",
    "Effect":"Allow",
    "Action":[
      "s3:*"
    ],
    "Resource":"*"
  }]
}
POLICY
}
resource "aws_transfer_profile" "local" {
  as2_id       = %[1]q
  profile_type = "LOCAL"
}

resource "aws_transfer_profile" "partner" {
  as2_id       = %[1]q
  profile_type = "PARTNER"
}

resource "aws_transfer_connector" "test" {
  access_role = aws_iam_role.test.arn

  as2_config {
    compression           = "DISABLED"
    encryption_algorithm  = "AES128_CBC"
    message_subject       = %[1]q
    local_profile_id      = aws_transfer_profile.local.profile_id
    mdn_response          = "NONE"
    mdn_signing_algorithm = "NONE"
    partner_profile_id    = aws_transfer_profile.partner.profile_id
    signing_algorithm     = "NONE"
  }

  url = %[2]q
}
data "aws_transfer_connector" "test" {
  id = aws_transfer_connector.test.id
}


`, rName, url)
}

func testAccConnectorDataSourceConfig_egressConfig(rName, publickey string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "Service": "transfer.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }]
 }
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<POLICY
{
  "Version":"2012-10-17",
  "Statement":[{
    "Sid":"AllowFullAccesstoS3",
    "Effect":"Allow",
    "Action":[
      "s3:*"
    ],
    "Resource":"*"
  }]
}
POLICY
}

resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpclattice_resource_gateway" "test" {
  name       = %[1]q
  vpc_id     = aws_vpc.test.id
  subnet_ids = [aws_subnet.test.id]
}

resource "aws_vpclattice_resource_configuration" "test" {
  name = %[1]q

  resource_gateway_identifier = aws_vpclattice_resource_gateway.test.id

  port_ranges = ["22"]
  protocol    = "TCP"

  resource_configuration_definition {
    dns_resource {
      domain_name     = "sftp.example.com"
      ip_address_type = "IPV4"
    }
  }
}

resource "aws_transfer_connector" "test" {
  access_role = aws_iam_role.test.arn

  sftp_config {
    trusted_host_keys = [%[2]q]
    user_secret_id    = aws_secretsmanager_secret.test.id
  }

  egress_config {
    vpc_lattice {
      resource_configuration_arn = aws_vpclattice_resource_configuration.test.arn
      port_number                = 22
    }
  }
}

data "aws_transfer_connector" "test" {
  id = aws_transfer_connector.test.id
}
`, rName, publickey))
}
