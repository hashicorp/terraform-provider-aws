// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package opensearch_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearch/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfopensearch "github.com/hashicorp/terraform-provider-aws/internal/service/opensearch"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOpenSearchDomainSAMLOptions_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var domain awstypes.DomainStatus

	rName := acctest.RandomWithPrefix(t, "acc-test")
	rUserName := acctest.RandomWithPrefix(t, "opensearch-master-user")
	idpEntityId := fmt.Sprintf("https://%s", acctest.RandomDomainName())

	resourceName := "aws_opensearch_domain_saml_options.test"
	esDomainResourceName := "aws_opensearch_domain.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckESDomainSAMLOptionsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainSAMLOptionsConfig_basic(rUserName, rName, idpEntityId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, t, esDomainResourceName, &domain),
					testAccCheckESDomainSAMLOptions(ctx, t, esDomainResourceName, resourceName),
					resource.TestCheckResourceAttr(resourceName, "saml_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "saml_options.0.enabled", acctest.CtTrue),
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

func TestAccOpenSearchDomainSAMLOptions_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, "acc-test")
	rUserName := acctest.RandomWithPrefix(t, "opensearch-master-user")
	idpEntityId := fmt.Sprintf("https://%s", acctest.RandomDomainName())

	resourceName := "aws_opensearch_domain_saml_options.test"
	esDomainResourceName := "aws_opensearch_domain.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckESDomainSAMLOptionsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainSAMLOptionsConfig_basic(rUserName, rName, idpEntityId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainSAMLOptions(ctx, t, esDomainResourceName, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfopensearch.ResourceDomainSAMLOptions(), resourceName),
				),
			},
		},
	})
}

func TestAccOpenSearchDomainSAMLOptions_disappears_Domain(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, "acc-test")
	rUserName := acctest.RandomWithPrefix(t, "opensearch-master-user")
	idpEntityId := fmt.Sprintf("https://%s", acctest.RandomDomainName())

	resourceName := "aws_opensearch_domain_saml_options.test"
	esDomainResourceName := "aws_opensearch_domain.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckESDomainSAMLOptionsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainSAMLOptionsConfig_basic(rUserName, rName, idpEntityId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainSAMLOptions(ctx, t, esDomainResourceName, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfopensearch.ResourceDomain(), esDomainResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccOpenSearchDomainSAMLOptions_Update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, "acc-test")
	rUserName := acctest.RandomWithPrefix(t, "opensearch-master-user")
	idpEntityId := fmt.Sprintf("https://%s", acctest.RandomDomainName())

	resourceName := "aws_opensearch_domain_saml_options.test"
	esDomainResourceName := "aws_opensearch_domain.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckESDomainSAMLOptionsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainSAMLOptionsConfig_basic(rUserName, rName, idpEntityId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "saml_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "saml_options.0.session_timeout_minutes", "60"),
					testAccCheckESDomainSAMLOptions(ctx, t, esDomainResourceName, resourceName),
				),
			},
			{
				Config: testAccDomainSAMLOptionsConfig_update(rUserName, rName, idpEntityId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "saml_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "saml_options.0.session_timeout_minutes", "180"),
					testAccCheckESDomainSAMLOptions(ctx, t, esDomainResourceName, resourceName),
				),
			},
		},
	})
}

func TestAccOpenSearchDomainSAMLOptions_Disabled(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, "acc-test")
	rUserName := acctest.RandomWithPrefix(t, "opensearch-master-user")
	idpEntityId := fmt.Sprintf("https://%s", acctest.RandomDomainName())

	resourceName := "aws_opensearch_domain_saml_options.test"
	esDomainResourceName := "aws_opensearch_domain.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckESDomainSAMLOptionsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainSAMLOptionsConfig_basic(rUserName, rName, idpEntityId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "saml_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "saml_options.0.session_timeout_minutes", "60"),
					testAccCheckESDomainSAMLOptions(ctx, t, esDomainResourceName, resourceName),
				),
			},
			{
				Config: testAccDomainSAMLOptionsConfig_disabled(rUserName, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "saml_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "saml_options.0.session_timeout_minutes", "0"),
					testAccCheckESDomainSAMLOptions(ctx, t, esDomainResourceName, resourceName),
				),
			},
		},
	})
}

func testAccCheckESDomainSAMLOptionsDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_opensearch_domain_saml_options" {
				continue
			}

			conn := acctest.ProviderMeta(ctx, t).OpenSearchClient(ctx)
			_, err := tfopensearch.FindDomainByName(ctx, conn, rs.Primary.Attributes[names.AttrDomainName])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("OpenSearch domain saml options %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckESDomainSAMLOptions(ctx context.Context, t *testing.T, esResource string, samlOptionsResource string) resource.TestCheckFunc {
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

		conn := acctest.ProviderMeta(ctx, t).OpenSearchClient(ctx)
		_, err := tfopensearch.FindDomainByName(ctx, conn, options.Primary.Attributes[names.AttrDomainName])

		return err
	}
}

func testAccDomainSAMLOptionsConfig_basic(userName, domainName, idpEntityId string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
}

resource "aws_opensearch_domain" "test" {
  domain_name    = %[2]q
  engine_version = "Elasticsearch_7.10"

  cluster_config {
    instance_type = "r5.large.search"
  }

  # Advanced security option must be enabled to configure SAML.
  advanced_security_options {
    enabled                        = true
    internal_user_database_enabled = false
    master_user_options {
      master_user_arn = aws_iam_user.test.arn
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

resource "aws_opensearch_domain_saml_options" "test" {
  domain_name = aws_opensearch_domain.test.domain_name

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

func testAccDomainSAMLOptionsConfig_update(userName, domainName, idpEntityId string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
}

resource "aws_opensearch_domain" "test" {
  domain_name    = %[2]q
  engine_version = "Elasticsearch_7.10"

  cluster_config {
    instance_type = "r5.large.search"
  }

  # Advanced security option must be enabled to configure SAML.
  advanced_security_options {
    enabled                        = true
    internal_user_database_enabled = false
    master_user_options {
      master_user_arn = aws_iam_user.test.arn
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

resource "aws_opensearch_domain_saml_options" "test" {
  domain_name = aws_opensearch_domain.test.domain_name

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

func testAccDomainSAMLOptionsConfig_disabled(userName string, domainName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
}

resource "aws_opensearch_domain" "test" {
  domain_name    = %[2]q
  engine_version = "Elasticsearch_7.10"

  cluster_config {
    instance_type = "r5.large.search"
  }

  # Advanced security option must be enabled to configure SAML.
  advanced_security_options {
    enabled                        = true
    internal_user_database_enabled = false
    master_user_options {
      master_user_arn = aws_iam_user.test.arn
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

resource "aws_opensearch_domain_saml_options" "test" {
  domain_name = aws_opensearch_domain.test.domain_name

  saml_options {
    enabled = false
  }
}
`, userName, domainName)
}
