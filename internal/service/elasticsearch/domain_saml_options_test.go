package elasticsearch_test

import (
	"fmt"
	"testing"

	elasticsearch "github.com/aws/aws-sdk-go/service/elasticsearchservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfelasticsearch "github.com/hashicorp/terraform-provider-aws/internal/service/elasticsearch"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccElasticsearchDomainSAMLOptions_basic(t *testing.T) {
	var domain elasticsearch.ElasticsearchDomainStatus

	rName := sdkacctest.RandomWithPrefix("acc-test")
	rUserName := sdkacctest.RandomWithPrefix("es-master-user")
	idpEntityId := fmt.Sprintf("https://%s", acctest.RandomDomainName())

	resourceName := "aws_elasticsearch_domain_saml_options.main"
	esDomainResourceName := "aws_elasticsearch_domain.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckESDomainSAMLOptionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainSAMLOptionsConfig(rUserName, rName, idpEntityId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(esDomainResourceName, &domain),
					testAccCheckESDomainSAMLOptions(esDomainResourceName, resourceName),
					resource.TestCheckResourceAttr(resourceName, "saml_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "saml_options.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "saml_options.0.idp.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "saml_options.0.idp.0.entity_id", idpEntityId),
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

func TestAccElasticsearchDomainSAMLOptions_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("acc-test")
	rUserName := sdkacctest.RandomWithPrefix("es-master-user")
	idpEntityId := fmt.Sprintf("https://%s", acctest.RandomDomainName())

	resourceName := "aws_elasticsearch_domain_saml_options.main"
	esDomainResourceName := "aws_elasticsearch_domain.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckESDomainSAMLOptionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainSAMLOptionsConfig(rUserName, rName, idpEntityId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainSAMLOptions(esDomainResourceName, resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfelasticsearch.ResourceDomainSAMLOptions(), resourceName),
				),
			},
		},
	})
}

func TestAccElasticsearchDomainSAMLOptions_disappears_Domain(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("acc-test")
	rUserName := sdkacctest.RandomWithPrefix("es-master-user")
	idpEntityId := fmt.Sprintf("https://%s", acctest.RandomDomainName())

	resourceName := "aws_elasticsearch_domain_saml_options.main"
	esDomainResourceName := "aws_elasticsearch_domain.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckESDomainSAMLOptionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainSAMLOptionsConfig(rUserName, rName, idpEntityId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainSAMLOptions(esDomainResourceName, resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfelasticsearch.ResourceDomain(), esDomainResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccElasticsearchDomainSAMLOptions_Update(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("acc-test")
	rUserName := sdkacctest.RandomWithPrefix("es-master-user")
	idpEntityId := fmt.Sprintf("https://%s", acctest.RandomDomainName())

	resourceName := "aws_elasticsearch_domain_saml_options.main"
	esDomainResourceName := "aws_elasticsearch_domain.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckESDomainSAMLOptionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainSAMLOptionsConfig(rUserName, rName, idpEntityId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "saml_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "saml_options.0.session_timeout_minutes", "60"),
					testAccCheckESDomainSAMLOptions(esDomainResourceName, resourceName),
				),
			},
			{
				Config: testAccDomainSAMLOptionsConfigUpdate(rUserName, rName, idpEntityId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "saml_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "saml_options.0.session_timeout_minutes", "180"),
					testAccCheckESDomainSAMLOptions(esDomainResourceName, resourceName),
				),
			},
		},
	})
}

func TestAccElasticsearchDomainSAMLOptions_Disabled(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("acc-test")
	rUserName := sdkacctest.RandomWithPrefix("es-master-user")
	idpEntityId := fmt.Sprintf("https://%s", acctest.RandomDomainName())

	resourceName := "aws_elasticsearch_domain_saml_options.main"
	esDomainResourceName := "aws_elasticsearch_domain.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckESDomainSAMLOptionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainSAMLOptionsConfig(rUserName, rName, idpEntityId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "saml_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "saml_options.0.session_timeout_minutes", "60"),
					testAccCheckESDomainSAMLOptions(esDomainResourceName, resourceName),
				),
			},
			{
				Config: testAccDomainSAMLOptionsConfigDisabled(rUserName, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "saml_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "saml_options.0.session_timeout_minutes", "0"),
					testAccCheckESDomainSAMLOptions(esDomainResourceName, resourceName),
				),
			},
		},
	})
}

func testAccCheckESDomainSAMLOptionsDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elasticsearch_domain_saml_options" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticsearchConn
		_, err := tfelasticsearch.FindDomainByName(conn, rs.Primary.Attributes["domain_name"])

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Elasticsearch domain saml options %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckESDomainSAMLOptions(esResource string, samlOptionsResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[esResource]
		if !ok {
			return fmt.Errorf("Not found: %s", esResource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		options, ok := s.RootModule().Resources[samlOptionsResource]
		if !ok {
			return fmt.Errorf("Not found: %s", samlOptionsResource)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticsearchConn
		_, err := tfelasticsearch.FindDomainByName(conn, options.Primary.Attributes["domain_name"])

		return err
	}
}

func testAccDomainSAMLOptionsConfig(userName, domainName, idpEntityId string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "es_master_user" {
  name = %[1]q
}

resource "aws_elasticsearch_domain" "example" {
  domain_name           = %[2]q
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

resource "aws_elasticsearch_domain_saml_options" "main" {
  domain_name = aws_elasticsearch_domain.example.domain_name

  saml_options {
    enabled = true
    idp {
      entity_id        = %[3]q
      metadata_content = templatefile("./test-fixtures/saml-metadata.xml.tpl", { entity_id = %[3]q })
    }
  }
}
`, userName, domainName, idpEntityId)
}

func testAccDomainSAMLOptionsConfigUpdate(userName, domainName, idpEntityId string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "es_master_user" {
  name = %[1]q
}

resource "aws_elasticsearch_domain" "example" {
  domain_name           = %[2]q
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

resource "aws_elasticsearch_domain_saml_options" "main" {
  domain_name = aws_elasticsearch_domain.example.domain_name

  saml_options {
    enabled = true
    idp {
      entity_id        = %[3]q
      metadata_content = templatefile("./test-fixtures/saml-metadata.xml.tpl", { entity_id = %[3]q })
    }
    session_timeout_minutes = 180
  }
}
`, userName, domainName, idpEntityId)
}

func testAccDomainSAMLOptionsConfigDisabled(userName string, domainName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "es_master_user" {
  name = "%s"
}

resource "aws_elasticsearch_domain" "example" {
  domain_name           = "%s"
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

resource "aws_elasticsearch_domain_saml_options" "main" {
  domain_name = aws_elasticsearch_domain.example.domain_name

  saml_options {
    enabled = false
  }
}
`, userName, domainName)
}
