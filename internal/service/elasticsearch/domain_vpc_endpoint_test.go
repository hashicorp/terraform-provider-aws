package elasticsearch_test

import (
	"context"
	"fmt"
	elasticsearch "github.com/aws/aws-sdk-go/service/elasticsearchservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfelasticsearch "github.com/hashicorp/terraform-provider-aws/internal/service/elasticsearch"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"testing"
)

func TestAccElasticsearchDomainVPCEndpoint_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("vpc-endpoint")

	var domain elasticsearch.ElasticsearchDomainStatus
	resourceName := "aws_elasticsearch_domain_vpc_endpoint.test"
	esDomainResourceName := "aws_elasticsearch_domain.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainVPCEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainVPCEndpointConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(esDomainResourceName, &domain),
					testAccCheckDomainVPCEndpointExists(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccElasticsearchDomainVPCEndpoint_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("vpc-endpoint")

	var domain elasticsearch.ElasticsearchDomainStatus
	resourceName := "aws_elasticsearch_domain_vpc_endpoint.test"
	esDomainResourceName := "aws_elasticsearch_domain.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainVPCEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainVPCEndpointConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(esDomainResourceName, &domain),
					testAccCheckDomainVPCEndpointExists(resourceName),
				),
			},
		},
	})
}

func testAccCheckDomainVPCEndpointDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elasticsearch_domain_vpc_endpoint" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticsearchConn
		_, err := tfelasticsearch.FindVPCEndpointByID(context.Background(), conn, rs.Primary.ID)
		if tfresource.NotFound(err) {
			continue
		}
		if err != nil {
			return err
		}

		return fmt.Errorf("elasticsearch domain vpc endpoint %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccCheckDomainVPCEndpointExists(esResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[esResource]
		if !ok {
			return fmt.Errorf("Not found: %s", esResource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticsearchConn
		_, err := tfelasticsearch.FindVPCEndpointByID(context.Background(), conn, rs.Primary.ID)
		return err
	}
}

func testAccDomainVPCEndpointConfig_basic(domainName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(domainName, 2), fmt.Sprintf(`
resource "aws_elasticsearch_domain" "example" {
  domain_name           = %[1]q
  elasticsearch_version = "7.10"

  cluster_config {
    instance_type = "r5.large.elasticsearch"
  }

  # Advanced security option must be enabled to configure SAML.
  advanced_security_options {
    enabled                        = true
    internal_user_database_enabled = false
    master_user_options {
      master_user_arn = aws_iam_user.es_master_user.arn
    }
  }

  # You must enable node-to-node encryption to use advanced security options.
  encrypt_at_rest {
    enabled = true
  }

  domain_endpoint_options {
    enforce_https       = true
    tls_security_policy = "Policy-Min-TLS-1-2-2019-07"
  }

  node_to_node_encryption {
    enabled = true
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
}

resource "aws_elasticsearch_domain_vpc_endpoint" "test" {
  domain_arn = aws_elasticsearch_domain.example.arn
  vpc_options = {
    security_group_ids = [ aws_security_group.test.id ]
	subnet_ids = aws_subnet.test[*].id
  }
}

resource "aws_security_group" "test" {
  name   = local.random_name
  vpc_id = aws_vpc.test.id
}

resource "aws_security_group_rule" "test" {
  type        = "ingress"
  from_port   = 443
  to_port     = 443
  protocol    = "tcp"
  cidr_blocks = ["0.0.0.0/0"]

  security_group_id = aws_security_group.test.id
}
`, domainName))
}
