// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package transfer_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/transfer/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftransfer "github.com/hashicorp/terraform-provider-aws/internal/service/transfer"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccTransferConnector_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.DescribedConnector
	resourceName := "aws_transfer_connector.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TransferEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectorConfig_basic(rName, "http://www.example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectorExists(ctx, t, resourceName, &conf),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "transfer", "connector/{id}"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, "connector_id"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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
					testAccCheckConnectorExists(ctx, t, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	publicKey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDNt3kA/dBkS6ZyU/sVDiGMuWJQaRPmLNbs/25K/e/fIl07ZWUgqqsFkcycLLMNFGD30Cmgp6XCXfNlIjzFWhNam+4cBb4DPpvieUw44VgsHK5JQy3JKlUfglmH5rs4G5pLiVfZpFU6jqvTsu4mE1CHCP0sXJlJhGxMG3QbsqYWNKiqGFEhuzGMs6fQlMkNiXsFoDmh33HAcXCbaFSC7V7xIqT1hlKu0iOL+GNjMj4R3xy0o3jafhO4MG2s3TwCQQCyaa5oyjL8iP8p3L9yp6cbIcXaS72SIgbCSGCyrcQPIKP2lJJHvE1oVWzLVBhR4eSzrlFDv7K4IErzaJmHqdiz" // nosemgrep:ci.ssh-key

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TransferEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectorConfig_sftpConfig(rName, "sftp://s-fakeserver.server.transfer.test.amazonaws.com", publicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectorExists(ctx, t, resourceName, &conf),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "transfer", "connector/{id}"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	publicKey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDNt3kA/dBkS6ZyU/sVDiGMuWJQaRPmLNbs/25K/e/fIl07ZWUgqqsFkcycLLMNFGD30Cmgp6XCXfNlIjzFWhNam+4cBb4DPpvieUw44VgsHK5JQy3JKlUfglmH5rs4G5pLiVfZpFU6jqvTsu4mE1CHCP0sXJlJhGxMG3QbsqYWNKiqGFEhuzGMs6fQlMkNiXsFoDmh33HAcXCbaFSC7V7xIqT1hlKu0iOL+GNjMj4R3xy0o3jafhO4MG2s3TwCQQCyaa5oyjL8iP8p3L9yp6cbIcXaS72SIgbCSGCyrcQPIKP2lJJHvE1oVWzLVBhR4eSzrlFDv7K4IErzaJmHqdiz" // nosemgrep:ci.ssh-key
	url := "sftp://s-fakeserver.server.transfer.test.amazonaws.com"
	securityPolicyName := "TransferSFTPConnectorSecurityPolicy-2024-03"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TransferEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectorConfig_securityPolicyName(rName, url, publicKey, securityPolicyName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectorExists(ctx, t, resourceName, &conf),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TransferEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectorConfig_basic(rName, "http://www.example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectorExists(ctx, t, resourceName, &conf),
					acctest.CheckSDKResourceDisappears(ctx, t, tftransfer.ResourceConnector(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccTransferConnector_egressConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.DescribedConnector
	resourceName := "aws_transfer_connector.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	publicKey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDNt3kA/dBkS6ZyU/sVDiGMuWJQaRPmLNbs/25K/e/fIl07ZWUgqqsFkcycLLMNFGD30Cmgp6XCXfNlIjzFWhNam+4cBb4DPpvieUw44VgsHK5JQy3JKlUfglmH5rs4G5pLiVfZpFU6jqvTsu4mE1CHCP0sXJlJhGxMG3QbsqYWNKiqGFEhuzGMs6fQlMkNiXsFoDmh33HAcXCbaFSC7V7xIqT1hlKu0iOL+GNjMj4R3xy0o3jafhO4MG2s3TwCQQCyaa5oyjL8iP8p3L9yp6cbIcXaS72SIgbCSGCyrcQPIKP2lJJHvE1oVWzLVBhR4eSzrlFDv7K4IErzaJmHqdiz" // nosemgrep:ci.ssh-key

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TransferEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectorConfig_egressConfig(rName, publicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectorExists(ctx, t, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "egress_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "egress_config.0.vpc_lattice.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "egress_config.0.vpc_lattice.0.resource_configuration_arn"),
					resource.TestCheckResourceAttr(resourceName, "egress_config.0.vpc_lattice.0.port_number", "22"),
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

func TestAccTransferConnector_egressConfigUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.DescribedConnector
	resourceName := "aws_transfer_connector.test"
	resourceConfigName := "aws_vpclattice_resource_configuration.test"
	resourceConfigName2 := "aws_vpclattice_resource_configuration.test2"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	publicKey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDNt3kA/dBkS6ZyU/sVDiGMuWJQaRPmLNbs/25K/e/fIl07ZWUgqqsFkcycLLMNFGD30Cmgp6XCXfNlIjzFWhNam+4cBb4DPpvieUw44VgsHK5JQy3JKlUfglmH5rs4G5pLiVfZpFU6jqvTsu4mE1CHCP0sXJlJhGxMG3QbsqYWNKiqGFEhuzGMs6fQlMkNiXsFoDmh33HAcXCbaFSC7V7xIqT1hlKu0iOL+GNjMj4R3xy0o3jafhO4MG2s3TwCQQCyaa5oyjL8iP8p3L9yp6cbIcXaS72SIgbCSGCyrcQPIKP2lJJHvE1oVWzLVBhR4eSzrlFDv7K4IErzaJmHqdiz" // nosemgrep:ci.ssh-key

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TransferEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectorConfig_egressConfig(rName, publicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectorExists(ctx, t, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "egress_config.0.vpc_lattice.0.resource_configuration_arn", resourceConfigName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "egress_config.0.vpc_lattice.0.port_number", "22"),
				),
			},
			{
				Config: testAccConnectorConfig_egressConfigUpdated(rName, publicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectorExists(ctx, t, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "egress_config.0.vpc_lattice.0.resource_configuration_arn", resourceConfigName2, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "egress_config.0.vpc_lattice.0.port_number", "2222"),
				),
			},
		},
	})
}

func TestAccTransferConnector_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.DescribedConnector
	resourceName := "aws_transfer_connector.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TransferEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectorConfig_tags1(rName, "http://www.example.com", acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectorExists(ctx, t, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
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
					testAccCheckConnectorExists(ctx, t, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccConnectorConfig_tags1(rName, "http://www.example.com", acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectorExists(ctx, t, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckConnectorExists(ctx context.Context, t *testing.T, n string, v *awstypes.DescribedConnector) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).TransferClient(ctx)

		output, err := tftransfer.FindConnectorByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckConnectorDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).TransferClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_transfer_connector" {
				continue
			}

			_, err := tftransfer.FindConnectorByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
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

func testAccConnectorConfig_egressConfigBase(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
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

resource "aws_vpclattice_resource_gateway" "test" {
  name       = %[1]q
  vpc_id     = aws_vpc.test.id
  subnet_ids = aws_subnet.test[*].id
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
`, rName))
}

func testAccConnectorConfig_egressConfig(rName, publickey string) string {
	return acctest.ConfigCompose(testAccConnectorConfig_egressConfigBase(rName), fmt.Sprintf(`
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

  tags = {
    Name = %[1]q
  }
}
`, rName, publickey))
}

func testAccConnectorConfig_egressConfigUpdated(rName, publickey string) string {
	return acctest.ConfigCompose(testAccConnectorConfig_egressConfigBase(rName), fmt.Sprintf(`
resource "aws_vpclattice_resource_configuration" "test2" {
  name = "%[1]s-updated"

  resource_gateway_identifier = aws_vpclattice_resource_gateway.test.id

  port_ranges = ["2222"]
  protocol    = "TCP"

  resource_configuration_definition {
    dns_resource {
      domain_name     = "sftp-updated.example.com"
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
      resource_configuration_arn = aws_vpclattice_resource_configuration.test2.arn
      port_number                = 2222
    }
  }
}
`, rName, publickey))
}
