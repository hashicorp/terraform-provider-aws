// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearch_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/opensearchservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfopensearch "github.com/hashicorp/terraform-provider-aws/internal/service/opensearch"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOpenSearchOutboundConnection_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var domain opensearchservice.DomainStatus
	ri := sdkacctest.RandString(10)
	name := fmt.Sprintf("tf-test-%s", ri)
	resourceName := "aws_opensearch_outbound_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOutboundConnectionConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, "aws_opensearch_domain.domain_1", &domain),
					testAccCheckDomainExists(ctx, "aws_opensearch_domain.domain_2", &domain),
					resource.TestCheckResourceAttr(resourceName, "connection_status", "ACTIVE"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"accept_connection"},
			},
		},
	})
}

func TestAccOpenSearchOutboundConnection_vpc(t *testing.T) {
	ctx := acctest.Context(t)
	var domain opensearchservice.DomainStatus
	ri := sdkacctest.RandString(10)
	name := fmt.Sprintf("tf-test-%s", ri)
	resourceName := "aws_opensearch_outbound_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOutboundConnectionConfig_vpcEndpoint(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, "aws_opensearch_domain.domain_1", &domain),
					testAccCheckDomainExists(ctx, "aws_opensearch_domain.domain_2", &domain),
					resource.TestCheckResourceAttrSet(resourceName, "connection_properties.0.endpoint"),
					resource.TestCheckResourceAttr(resourceName, "connection_status", "ACTIVE"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"accept_connection"},
			},
		},
	})
}

func TestAccOpenSearchOutboundConnection_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var domain opensearchservice.DomainStatus
	ri := sdkacctest.RandString(10)
	name := fmt.Sprintf("tf-test-%s", ri)
	resourceName := "aws_opensearch_outbound_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOutboundConnectionConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, "aws_opensearch_domain.domain_1", &domain),
					testAccCheckDomainExists(ctx, "aws_opensearch_domain.domain_2", &domain),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfopensearch.ResourceOutboundConnection(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccOutboundConnectionConfig(name string) string {
	// Satisfy the pw requirements
	pw := fmt.Sprintf("Aa1-%s", sdkacctest.RandString(10))
	return fmt.Sprintf(`
resource "aws_opensearch_domain" "domain_1" {
  domain_name = "%s-1"

  cluster_config {
    instance_type = "t3.small.search" # supported in both aws and aws-us-gov
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  node_to_node_encryption {
    enabled = true
  }

  advanced_security_options {
    enabled                        = true
    internal_user_database_enabled = true

    master_user_options {
      master_user_name     = "test"
      master_user_password = "%s"
    }
  }

  encrypt_at_rest {
    enabled = true
  }

  domain_endpoint_options {
    enforce_https       = true
    tls_security_policy = "Policy-Min-TLS-1-2-2019-07"
  }
}

resource "aws_opensearch_domain" "domain_2" {
  domain_name = "%s-2"

  cluster_config {
    instance_type = "t3.small.search" # supported in both aws and aws-us-gov
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  node_to_node_encryption {
    enabled = true
  }

  advanced_security_options {
    enabled                        = true
    internal_user_database_enabled = true

    master_user_options {
      master_user_name     = "test"
      master_user_password = "%s"
    }
  }

  encrypt_at_rest {
    enabled = true
  }

  domain_endpoint_options {
    enforce_https       = true
    tls_security_policy = "Policy-Min-TLS-1-2-2019-07"
  }
}

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

resource "aws_opensearch_outbound_connection" "test" {
  connection_alias  = "%s"
  connection_mode   = "DIRECT"
  accept_connection = true

  connection_properties {
    cross_cluster_search {
      skip_unavailable = "ENABLED"
    }
  }

  local_domain_info {
    owner_id    = data.aws_caller_identity.current.account_id
    region      = data.aws_region.current.name
    domain_name = aws_opensearch_domain.domain_1.domain_name
  }

  remote_domain_info {
    owner_id    = data.aws_caller_identity.current.account_id
    region      = data.aws_region.current.name
    domain_name = aws_opensearch_domain.domain_2.domain_name
  }
}
`, name, pw, name, pw, name)
}

func testAccOutboundConnectionConfig_vpcEndpoint(name string) string {
	// Satisfy the pw requirements
	pw := fmt.Sprintf("Aa1-%s", sdkacctest.RandString(10))

	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/22"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "192.168.0.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "192.168.1.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_security_group" "test2" {
  vpc_id = aws_vpc.test.id
}

resource "aws_opensearch_domain" "domain_1" {
  domain_name = "%[1]s-1"

  cluster_config {
    instance_type = "t3.small.search" # supported in both aws and aws-us-gov
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  node_to_node_encryption {
    enabled = true
  }

  advanced_security_options {
    enabled                        = true
    internal_user_database_enabled = true

    master_user_options {
      master_user_name     = "test"
      master_user_password = %[2]q
    }
  }

  encrypt_at_rest {
    enabled = true
  }

  domain_endpoint_options {
    enforce_https       = true
    tls_security_policy = "Policy-Min-TLS-1-2-2019-07"
  }
}

resource "aws_opensearch_domain" "domain_2" {
  domain_name = "%[1]s-2"

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  cluster_config {
    instance_count         = 2
    zone_awareness_enabled = true
    instance_type          = "t3.small.search"
  }

  vpc_options {
    security_group_ids = [aws_security_group.test.id, aws_security_group.test2.id]
    subnet_ids         = [aws_subnet.test.id, aws_subnet.test2.id]
  }
}

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

resource "aws_opensearch_outbound_connection" "test" {
  connection_alias  = %[1]q
  connection_mode   = "VPC_ENDPOINT"
  accept_connection = true

  local_domain_info {
    owner_id    = data.aws_caller_identity.current.account_id
    region      = data.aws_region.current.name
    domain_name = aws_opensearch_domain.domain_1.domain_name
  }

  remote_domain_info {
    owner_id    = data.aws_caller_identity.current.account_id
    region      = data.aws_region.current.name
    domain_name = aws_opensearch_domain.domain_2.domain_name
  }
}

`, name, pw))
}
