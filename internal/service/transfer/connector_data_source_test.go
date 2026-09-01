// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package transfer_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccTransferConnectorDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_transfer_connector.test"
	resourceName := "aws_transfer_connector.test"
	url := "http://www.example.com"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TransferEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
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
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrURL, resourceName, names.AttrURL),
				),
			},
		},
	})
}

func TestAccTransferConnectorDataSource_egressConfig(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_transfer_connector.test"
	resourceName := "aws_transfer_connector.test"
	publicKey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDNt3kA/dBkS6ZyU/sVDiGMuWJQaRPmLNbs/25K/e/fIl07ZWUgqqsFkcycLLMNFGD30Cmgp6XCXfNlIjzFWhNam+4cBb4DPpvieUw44VgsHK5JQy3JKlUfglmH5rs4G5pLiVfZpFU6jqvTsu4mE1CHCP0sXJlJhGxMG3QbsqYWNKiqGFEhuzGMs6fQlMkNiXsFoDmh33HAcXCbaFSC7V7xIqT1hlKu0iOL+GNjMj4R3xy0o3jafhO4MG2s3TwCQQCyaa5oyjL8iP8p3L9yp6cbIcXaS72SIgbCSGCyrcQPIKP2lJJHvE1oVWzLVBhR4eSzrlFDv7K4IErzaJmHqdiz" // nosemgrep:ci.ssh-key

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TransferEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
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
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func testAccConnectorDataSourceConfig_basic(rName, url string) string {
	return acctest.ConfigCompose(testAccConnectorConfig_basic(rName, url), `
data "aws_transfer_connector" "test" {
  id = aws_transfer_connector.test.id
}
`)
}

func testAccConnectorDataSourceConfig_egressConfig(rName, publickey string) string {
	return acctest.ConfigCompose(testAccConnectorConfig_egressConfig(rName, publickey), `
data "aws_transfer_connector" "test" {
  id = aws_transfer_connector.test.id
}
`)
}
