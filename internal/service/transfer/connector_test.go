// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transfer_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/transfer"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftransfer "github.com/hashicorp/terraform-provider-aws/internal/service/transfer"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccTransferConnector_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedConnector
	resourceName := "aws_transfer_connector.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectorConfig_basic(rName, "http://www.example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectorExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "url", "http://www.example.com"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConnectorConfig_basic(rName, "http://www.example.net"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectorExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "url", "http://www.example.net"),
				),
			},
		},
	})
}

func TestAccTransferConnector_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedConnector
	resourceName := "aws_transfer_connector.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectorConfig_basic(rName, "http://www.example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectorExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tftransfer.ResourceConnector(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccTransferConnector_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedConnector
	resourceName := "aws_transfer_connector.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectorConfig_tags1(rName, "http://www.example.com", "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectorExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConnectorConfig_tags2(rName, "http://www.example.com", "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectorExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccConnectorConfig_tags1(rName, "http://www.example.com", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectorExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckConnectorExists(ctx context.Context, n string, v *transfer.DescribedConnector) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Transfer Connector ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).TransferConn(ctx)

		output, err := tftransfer.FindConnectorByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckConnectorDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).TransferConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_transfer_connector" {
				continue
			}

			_, err := tftransfer.FindConnectorByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Transfer Connector %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccConnectorConfig_base(rName string) string {
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
	 "Statement":[
		{
		   "Sid":"AllowFullAccesstoS3",
		   "Effect":"Allow",
		   "Action":[
			  "s3:*"
		   ],
		   "Resource":"*"
		}
	 ]
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
`, rName)
}

func testAccConnectorConfig_basic(rName, url string) string {
	return acctest.ConfigCompose(testAccConnectorConfig_base(rName), fmt.Sprintf(`
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
`, rName, url))
}

func testAccConnectorConfig_tags1(rName, url, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccConnectorConfig_base(rName), fmt.Sprintf(`
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

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, url, tagKey1, tagValue1))
}

func testAccConnectorConfig_tags2(rName, url, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccConnectorConfig_base(rName), fmt.Sprintf(`
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

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, url, tagKey1, tagValue1, tagKey2, tagValue2))
}
