// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transfer_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/transfer/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftransfer "github.com/hashicorp/terraform-provider-aws/internal/service/transfer"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccTransferConnector_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.DescribedConnector
	resourceName := "aws_transfer_connector.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
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
				Config: testAccConnectorConfig_basic(rName, "http://www.example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectorExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrURL, "http://www.example.com"),
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
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrURL, "http://www.example.net"),
				),
			},
		},
	})
}

func TestAccTransferConnector_sftpConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.DescribedConnector
	resourceName := "aws_transfer_connector.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	publicKey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDNt3kA/dBkS6ZyU/sVDiGMuWJQaRPmLNbs/25K/e/fIl07ZWUgqqsFkcycLLMNFGD30Cmgp6XCXfNlIjzFWhNam+4cBb4DPpvieUw44VgsHK5JQy3JKlUfglmH5rs4G5pLiVfZpFU6jqvTsu4mE1CHCP0sXJlJhGxMG3QbsqYWNKiqGFEhuzGMs6fQlMkNiXsFoDmh33HAcXCbaFSC7V7xIqT1hlKu0iOL+GNjMj4R3xy0o3jafhO4MG2s3TwCQQCyaa5oyjL8iP8p3L9yp6cbIcXaS72SIgbCSGCyrcQPIKP2lJJHvE1oVWzLVBhR4eSzrlFDv7K4IErzaJmHqdiz" // nosemgrep:ci.ssh-key

	resource.Test(t, resource.TestCase{
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
				Config: testAccConnectorConfig_sftpConfig(rName, "sftp://s-fakeserver.server.transfer.test.amazonaws.com", publicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectorExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrURL, "sftp://s-fakeserver.server.transfer.test.amazonaws.com"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTransferConnector_securityPolicyName(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.DescribedConnector
	resourceName := "aws_transfer_connector.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	publicKey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDNt3kA/dBkS6ZyU/sVDiGMuWJQaRPmLNbs/25K/e/fIl07ZWUgqqsFkcycLLMNFGD30Cmgp6XCXfNlIjzFWhNam+4cBb4DPpvieUw44VgsHK5JQy3JKlUfglmH5rs4G5pLiVfZpFU6jqvTsu4mE1CHCP0sXJlJhGxMG3QbsqYWNKiqGFEhuzGMs6fQlMkNiXsFoDmh33HAcXCbaFSC7V7xIqT1hlKu0iOL+GNjMj4R3xy0o3jafhO4MG2s3TwCQQCyaa5oyjL8iP8p3L9yp6cbIcXaS72SIgbCSGCyrcQPIKP2lJJHvE1oVWzLVBhR4eSzrlFDv7K4IErzaJmHqdiz" // nosemgrep:ci.ssh-key
	url := "sftp://s-fakeserver.server.transfer.test.amazonaws.com"
	securityPolicyName := "TransferSFTPConnectorSecurityPolicy-2024-03"

	resource.Test(t, resource.TestCase{
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
				Config: testAccConnectorConfig_securityPolicyName(rName, url, publicKey, securityPolicyName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectorExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "security_policy_name", securityPolicyName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTransferConnector_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.DescribedConnector
	resourceName := "aws_transfer_connector.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
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
	var conf awstypes.DescribedConnector
	resourceName := "aws_transfer_connector.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
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
				Config: testAccConnectorConfig_tags1(rName, "http://www.example.com", acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectorExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConnectorConfig_tags2(rName, "http://www.example.com", acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectorExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccConnectorConfig_tags1(rName, "http://www.example.com", acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectorExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckConnectorExists(ctx context.Context, n string, v *awstypes.DescribedConnector) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Transfer Connector ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).TransferClient(ctx)

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
		conn := acctest.Provider.Meta().(*conns.AWSClient).TransferClient(ctx)

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

func testAccConnectorConfig_securityPolicyName(rName, url, publickey, securityPolicyName string) string {
	return acctest.ConfigCompose(testAccConnectorConfig_base(rName), fmt.Sprintf(`
resource "aws_transfer_connector" "test" {
  access_role = aws_iam_role.test.arn

  sftp_config {
    trusted_host_keys = [%[3]q]
    user_secret_id    = aws_secretsmanager_secret.test.id
  }

  url                  = %[2]q
  security_policy_name = %[4]q
}

resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}
`, rName, url, publickey, securityPolicyName))
}

func testAccConnectorConfig_sftpConfig(rName, url, publickey string) string {
	return acctest.ConfigCompose(testAccConnectorConfig_base(rName), fmt.Sprintf(`
resource "aws_transfer_connector" "test" {
  access_role = aws_iam_role.test.arn

  sftp_config {
    trusted_host_keys = [%[3]q]
    user_secret_id    = aws_secretsmanager_secret.test.id
  }

  url = %[2]q
}

resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}
`, rName, url, publickey))
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
