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
