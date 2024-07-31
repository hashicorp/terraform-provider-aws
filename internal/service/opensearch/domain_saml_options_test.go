// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearch_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/opensearchservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfopensearch "github.com/hashicorp/terraform-provider-aws/internal/service/opensearch"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOpenSearchDomainSAMLOptions_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var domain opensearchservice.DomainStatus

	rName := sdkacctest.RandomWithPrefix("acc-test")
	rUserName := sdkacctest.RandomWithPrefix("opensearch-master-user")
	idpEntityId := fmt.Sprintf("https://%s", acctest.RandomDomainName())

	resourceName := "aws_opensearch_domain_saml_options.test"
	esDomainResourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckESDomainSAMLOptionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainSAMLOptionsConfig_basic(rUserName, rName, idpEntityId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, esDomainResourceName, &domain),
					testAccCheckESDomainSAMLOptions(ctx, esDomainResourceName, resourceName),
					resource.TestCheckResourceAttr(resourceName, "saml_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "saml_options.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "saml_options.0.idp.#", acctest.Ct1),
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
	rName := sdkacctest.RandomWithPrefix("acc-test")
	rUserName := sdkacctest.RandomWithPrefix("opensearch-master-user")
	idpEntityId := fmt.Sprintf("https://%s", acctest.RandomDomainName())

	resourceName := "aws_opensearch_domain_saml_options.test"
	esDomainResourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckESDomainSAMLOptionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainSAMLOptionsConfig_basic(rUserName, rName, idpEntityId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainSAMLOptions(ctx, esDomainResourceName, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfopensearch.ResourceDomainSAMLOptions(), resourceName),
				),
			},
		},
	})
}

func TestAccOpenSearchDomainSAMLOptions_disappears_Domain(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("acc-test")
	rUserName := sdkacctest.RandomWithPrefix("opensearch-master-user")
	idpEntityId := fmt.Sprintf("https://%s", acctest.RandomDomainName())

	resourceName := "aws_opensearch_domain_saml_options.test"
	esDomainResourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckESDomainSAMLOptionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainSAMLOptionsConfig_basic(rUserName, rName, idpEntityId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainSAMLOptions(ctx, esDomainResourceName, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfopensearch.ResourceDomain(), esDomainResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccOpenSearchDomainSAMLOptions_Update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("acc-test")
	rUserName := sdkacctest.RandomWithPrefix("opensearch-master-user")
	idpEntityId := fmt.Sprintf("https://%s", acctest.RandomDomainName())

	resourceName := "aws_opensearch_domain_saml_options.test"
	esDomainResourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckESDomainSAMLOptionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainSAMLOptionsConfig_basic(rUserName, rName, idpEntityId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "saml_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "saml_options.0.session_timeout_minutes", "60"),
					testAccCheckESDomainSAMLOptions(ctx, esDomainResourceName, resourceName),
				),
			},
			{
				Config: testAccDomainSAMLOptionsConfig_update(rUserName, rName, idpEntityId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "saml_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "saml_options.0.session_timeout_minutes", "180"),
					testAccCheckESDomainSAMLOptions(ctx, esDomainResourceName, resourceName),
				),
			},
		},
	})
}

func TestAccOpenSearchDomainSAMLOptions_Disabled(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("acc-test")
	rUserName := sdkacctest.RandomWithPrefix("opensearch-master-user")
	idpEntityId := fmt.Sprintf("https://%s", acctest.RandomDomainName())

	resourceName := "aws_opensearch_domain_saml_options.test"
	esDomainResourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckESDomainSAMLOptionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainSAMLOptionsConfig_basic(rUserName, rName, idpEntityId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "saml_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "saml_options.0.session_timeout_minutes", "60"),
					testAccCheckESDomainSAMLOptions(ctx, esDomainResourceName, resourceName),
				),
			},
			{
				Config: testAccDomainSAMLOptionsConfig_disabled(rUserName, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "saml_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "saml_options.0.session_timeout_minutes", acctest.Ct0),
					testAccCheckESDomainSAMLOptions(ctx, esDomainResourceName, resourceName),
				),
			},
		},
	})
}

func testAccCheckESDomainSAMLOptionsDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_opensearch_domain_saml_options" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).OpenSearchConn(ctx)
			_, err := tfopensearch.FindDomainByName(ctx, conn, rs.Primary.Attributes[names.AttrDomainName])

			if tfresource.NotFound(err) {
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

func testAccCheckESDomainSAMLOptions(ctx context.Context, esResource string, samlOptionsResource string) resource.TestCheckFunc {
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

		conn := acctest.Provider.Meta().(*conns.AWSClient).OpenSearchConn(ctx)
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
