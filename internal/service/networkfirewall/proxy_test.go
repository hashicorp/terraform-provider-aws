// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkfirewall_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfnetworkfirewall "github.com/hashicorp/terraform-provider-aws/internal/service/networkfirewall"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccNetworkFirewallProxy_basic(t *testing.T) {
	t.Helper()

	ctx := acctest.Context(t)
	var v networkfirewall.DescribeProxyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_proxy.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProxyExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "network-firewall", regexache.MustCompile(`proxy/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "nat_gateway_id"),
					resource.TestCheckResourceAttr(resourceName, "listener_properties.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "listener_properties.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "listener_properties.0.type", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "listener_properties.1.port", "443"),
					resource.TestCheckResourceAttr(resourceName, "listener_properties.1.type", "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, "tls_intercept_properties.0.tls_intercept_mode", "DISABLED"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreateTime),
					resource.TestCheckResourceAttrSet(resourceName, "update_token"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"update_token"},
			},
		},
	})
}

func testAccNetworkFirewallProxy_disappears(t *testing.T) {
	t.Helper()

	ctx := acctest.Context(t)
	var v networkfirewall.DescribeProxyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_proxy.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfnetworkfirewall.ResourceProxy, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccNetworkFirewallProxy_tlsInterceptEnabled(t *testing.T) {
	t.Helper()

	ctx := acctest.Context(t)
	var v networkfirewall.DescribeProxyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_proxy.test"
	pcaResourceName := "aws_acmpca_certificate_authority.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyConfig_tlsInterceptEnabled(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProxyExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "network-firewall", regexache.MustCompile(`proxy/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "nat_gateway_id"),
					resource.TestCheckResourceAttr(resourceName, "listener_properties.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "listener_properties.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "listener_properties.0.type", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "listener_properties.1.port", "443"),
					resource.TestCheckResourceAttr(resourceName, "listener_properties.1.type", "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, "tls_intercept_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tls_intercept_properties.0.tls_intercept_mode", "ENABLED"),
					resource.TestCheckResourceAttrPair(resourceName, "tls_intercept_properties.0.pca_arn", pcaResourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreateTime),
					resource.TestCheckResourceAttrSet(resourceName, "update_token"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"update_token"},
			},
		},
	})
}

func testAccNetworkFirewallProxy_logging(t *testing.T) {
	t.Helper()

	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyConfig_logging(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// CloudWatch Logs delivery source
					resource.TestCheckResourceAttr("aws_cloudwatch_log_delivery_source.cwl", "log_type", "ALERT_LOGS"),
					resource.TestCheckResourceAttrPair("aws_cloudwatch_log_delivery_source.cwl", names.AttrResourceARN, "aws_networkfirewall_proxy.test", names.AttrARN),
					// CloudWatch Logs delivery destination
					resource.TestCheckResourceAttr("aws_cloudwatch_log_delivery_destination.cwl", "delivery_destination_type", "CWL"),
					// CloudWatch Logs delivery
					resource.TestCheckResourceAttrSet("aws_cloudwatch_log_delivery.cwl", names.AttrID),
					resource.TestCheckResourceAttrPair("aws_cloudwatch_log_delivery.cwl", "delivery_source_name", "aws_cloudwatch_log_delivery_source.cwl", names.AttrName),
					resource.TestCheckResourceAttrPair("aws_cloudwatch_log_delivery.cwl", "delivery_destination_arn", "aws_cloudwatch_log_delivery_destination.cwl", names.AttrARN),
					// S3 delivery source
					resource.TestCheckResourceAttr("aws_cloudwatch_log_delivery_source.s3", "log_type", "ALLOW_LOGS"),
					resource.TestCheckResourceAttrPair("aws_cloudwatch_log_delivery_source.s3", names.AttrResourceARN, "aws_networkfirewall_proxy.test", names.AttrARN),
					// S3 delivery destination
					resource.TestCheckResourceAttr("aws_cloudwatch_log_delivery_destination.s3", "delivery_destination_type", "S3"),
					// S3 delivery
					resource.TestCheckResourceAttrSet("aws_cloudwatch_log_delivery.s3", names.AttrID),
					resource.TestCheckResourceAttrPair("aws_cloudwatch_log_delivery.s3", "delivery_source_name", "aws_cloudwatch_log_delivery_source.s3", names.AttrName),
					resource.TestCheckResourceAttrPair("aws_cloudwatch_log_delivery.s3", "delivery_destination_arn", "aws_cloudwatch_log_delivery_destination.s3", names.AttrARN),
				),
			},
		},
	})
}

func testAccCheckProxyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).NetworkFirewallClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networkfirewall_proxy" {
				continue
			}

			out, err := tfnetworkfirewall.FindProxyByARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if out != nil && out.Proxy != nil && out.Proxy.DeleteTime != nil {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("NetworkFirewall Proxy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckProxyExists(ctx context.Context, t *testing.T, n string, v *networkfirewall.DescribeProxyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).NetworkFirewallClient(ctx)

		output, err := tfnetworkfirewall.FindProxyByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

// testAccProxyConfig_baseVPC creates a reusable VPC configuration for proxy tests.
// It includes:
// - VPC with CIDR 10.0.0.0/16
// - Public subnet (10.0.1.0/24) with Internet Gateway
// - Private subnet (10.0.2.0/24)
// - Internet Gateway
// - NAT Gateway in the public subnet
// - Route tables for public and private subnets
func testAccProxyConfig_baseVPC(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "public" {
  vpc_id                  = aws_vpc.test.id
  cidr_block              = "10.0.1.0/24"
  availability_zone       = data.aws_availability_zones.available.names[0]
  map_public_ip_on_launch = true

  tags = {
    Name = "%[1]s-public"
  }
}

resource "aws_subnet" "private" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "%[1]s-private"
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_eip" "test" {
  domain = "vpc"

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_internet_gateway.test]
}

resource "aws_nat_gateway" "test" {
  allocation_id = aws_eip.test.id
  subnet_id     = aws_subnet.public.id

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_internet_gateway.test]
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = "%[1]s-public"
  }
}

resource "aws_route_table" "private" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.test.id
  }

  tags = {
    Name = "%[1]s-private"
  }
}

resource "aws_route_table_association" "public" {
  subnet_id      = aws_subnet.public.id
  route_table_id = aws_route_table.public.id
}

resource "aws_route_table_association" "private" {
  subnet_id      = aws_subnet.private.id
  route_table_id = aws_route_table.private.id
}
`, rName))
}

// testAccProxyConfig_baseProxyConfiguration creates a basic proxy configuration
// that can be reused across proxy tests.
func testAccProxyConfig_baseProxyConfiguration(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_proxy_configuration" "test" {
  name = %[1]q

  default_rule_phase_actions {
    post_response = "ALLOW"
    pre_dns       = "ALLOW"
    pre_request   = "ALLOW"
  }
}
`, rName)
}

func testAccProxyConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccProxyConfig_baseVPC(rName),
		testAccProxyConfig_baseProxyConfiguration(rName),
		fmt.Sprintf(`
resource "aws_networkfirewall_proxy" "test" {
  name                     = %[1]q
  nat_gateway_id           = aws_nat_gateway.test.id
  proxy_configuration_arn = aws_networkfirewall_proxy_configuration.test.arn

  tls_intercept_properties {
     tls_intercept_mode = "DISABLED"
  }

  listener_properties {
    port = 8080
    type = "HTTP"
  }

  listener_properties {
    port = 443
    type = "HTTPS"
  }
}
`, rName))
}

func testAccProxyConfig_logging(rName string) string {
	return acctest.ConfigCompose(
		testAccProxyConfig_baseVPC(rName),
		testAccProxyConfig_baseProxyConfiguration(rName),
		fmt.Sprintf(`
resource "aws_networkfirewall_proxy" "test" {
  name                    = %[1]q
  nat_gateway_id          = aws_nat_gateway.test.id
  proxy_configuration_arn = aws_networkfirewall_proxy_configuration.test.arn

  tls_intercept_properties {
    tls_intercept_mode = "DISABLED"
  }

  listener_properties {
    port = 8080
    type = "HTTP"
  }

  listener_properties {
    port = 443
    type = "HTTPS"
  }
}

# CloudWatch Logs delivery

resource "aws_cloudwatch_log_group" "test" {
  name              = %[1]q
  retention_in_days = 7
}

resource "aws_cloudwatch_log_delivery_source" "cwl" {
  name         = "%[1]s-cwl"
  log_type     = "ALERT_LOGS"
  resource_arn = aws_networkfirewall_proxy.test.arn
}

resource "aws_cloudwatch_log_delivery_destination" "cwl" {
  name = "%[1]s-cwl"

  delivery_destination_configuration {
    destination_resource_arn = aws_cloudwatch_log_group.test.arn
  }
}

resource "aws_cloudwatch_log_delivery" "cwl" {
  delivery_source_name     = aws_cloudwatch_log_delivery_source.cwl.name
  delivery_destination_arn = aws_cloudwatch_log_delivery_destination.cwl.arn
}

# S3 delivery

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_cloudwatch_log_delivery_source" "s3" {
  name         = "%[1]s-s3"
  log_type     = "ALLOW_LOGS"
  resource_arn = aws_networkfirewall_proxy.test.arn
}

resource "aws_cloudwatch_log_delivery_destination" "s3" {
  name = "%[1]s-s3"

  delivery_destination_configuration {
    destination_resource_arn = aws_s3_bucket.test.arn
  }
}

resource "aws_cloudwatch_log_delivery" "s3" {
  delivery_source_name     = aws_cloudwatch_log_delivery_source.s3.name
  delivery_destination_arn = aws_cloudwatch_log_delivery_destination.s3.arn
}
`, rName))
}

func testAccProxyConfig_tlsInterceptEnabled(rName string) string {
	return acctest.ConfigCompose(
		testAccProxyConfig_baseVPC(rName),
		testAccProxyConfig_baseProxyConfiguration(rName),
		fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {}

# Create a root CA for TLS interception
resource "aws_acmpca_certificate_authority" "test" {
  type       = "ROOT"

  certificate_authority_configuration {
    key_algorithm     = "RSA_2048"
    signing_algorithm = "SHA256WITHRSA"

    subject {
      common_name = "%[1]s Terraform Test CA"
    }
  }

  permanent_deletion_time_in_days = 7

  tags = {
    Name = %[1]q
  }
}

resource "aws_acmpca_certificate" "test" {
  certificate_authority_arn   = aws_acmpca_certificate_authority.test.arn
  certificate_signing_request = aws_acmpca_certificate_authority.test.certificate_signing_request
  signing_algorithm           = "SHA512WITHRSA"

  template_arn = "arn:${data.aws_partition.current.partition}:acm-pca:::template/RootCACertificate/V1"

  validity {
    type  = "YEARS"
    value = 1
  }
}

resource "aws_acmpca_certificate_authority_certificate" "test" {
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn

  certificate       = aws_acmpca_certificate.test.certificate
  certificate_chain = aws_acmpca_certificate.test.certificate_chain
}


# Grant Network Firewall proxy permission to use the PCA
resource "aws_acmpca_policy" "test" {
  resource_arn = aws_acmpca_certificate_authority.test.arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "NetworkFirewallProxyReadAccess"
        Effect = "Allow"
        Principal = {
          Service = "proxy.network-firewall.amazonaws.com"
        }
        Action = [
          "acm-pca:GetCertificate",
          "acm-pca:DescribeCertificateAuthority",
          "acm-pca:GetCertificateAuthorityCertificate",
          "acm-pca:ListTags",
          "acm-pca:ListPermissions",
        ]
        Resource = aws_acmpca_certificate_authority.test.arn
        Condition = {
          ArnEquals = {
            "aws:SourceArn" = aws_networkfirewall_proxy.test.arn
          }
        }
      },
      {
        Sid    = "NetworkFirewallProxyIssueCertificate"
        Effect = "Allow"
        Principal = {
          Service = "proxy.network-firewall.amazonaws.com"
        }
        Action = [
          "acm-pca:IssueCertificate",
        ]
        Resource = aws_acmpca_certificate_authority.test.arn
        Condition = {
          StringEquals = {
            "acm-pca:TemplateArn" = "arn:${data.aws_partition.current.partition}:acm-pca:::template/SubordinateCACertificate_PathLen0/V1"
          }
          ArnEquals = {
            "aws:SourceArn" = aws_networkfirewall_proxy.test.arn
          }
        }
      }
    ]
  })
}

# Create RAM resource share for the PCA
resource "aws_ram_resource_share" "test" {
  name                      = %[1]q
  allow_external_principals = true
  permission_arns = ["arn:aws:ram::aws:permission/AWSRAMSubordinateCACertificatePathLen0IssuanceCertificateAuthority"]

  tags = {
    Name = %[1]q
  }
}
  
# Associate the PCA with the RAM share
resource "aws_ram_resource_share_associations_exclusive" "test" {
  principals         = ["proxy.network-firewall.amazonaws.com"]
  resource_arns      = [aws_acmpca_certificate_authority.test.arn]
  resource_share_arn = aws_ram_resource_share.test.arn
  sources            = [data.aws_caller_identity.current.account_id]

  lifecycle {
    ignore_changes = [
       resource_arns
    ]
    replace_triggered_by = [
       aws_acmpca_certificate_authority.test
    ]
  }
}

resource "aws_networkfirewall_proxy" "test" {
  name                     = %[1]q
  nat_gateway_id           = aws_nat_gateway.test.id
  proxy_configuration_arn = aws_networkfirewall_proxy_configuration.test.arn

  tls_intercept_properties {
    tls_intercept_mode = "ENABLED"
    pca_arn            = aws_acmpca_certificate_authority.test.arn
  }

  listener_properties {
    port = 8080
    type = "HTTP"
  }

  listener_properties {
    port = 443
    type = "HTTPS"
  }
}
`, rName))
}
