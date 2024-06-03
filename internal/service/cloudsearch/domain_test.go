// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudsearch_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudsearch/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudsearch "github.com/hashicorp/terraform-provider-aws/internal/service/cloudsearch"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudSearchDomain_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.DomainStatus
	resourceName := "aws_cloudsearch_domain.test"
	rName := testAccDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudSearchEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccDomainExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "cloudsearch", fmt.Sprintf("domain/%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "domain_id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "endpoint_options.0.enforce_https", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_options.0.tls_security_policy"),
					resource.TestCheckResourceAttr(resourceName, "index_field.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "multi_az", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "scaling_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scaling_parameters.0.desired_instance_type", ""),
					resource.TestCheckResourceAttr(resourceName, "scaling_parameters.0.desired_partition_count", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "scaling_parameters.0.desired_replication_count", acctest.Ct0),
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

func TestAccCloudSearchDomain_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.DomainStatus
	resourceName := "aws_cloudsearch_domain.test"
	rName := testAccDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudSearchEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccDomainExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloudsearch.ResourceDomain(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudSearchDomain_indexFields(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.DomainStatus
	resourceName := "aws_cloudsearch_domain.test"
	rName := testAccDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudSearchEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_indexFields(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccDomainExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "index_field.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "index_field.*", map[string]string{
						names.AttrName:         "int_test",
						names.AttrType:         "int",
						names.AttrDefaultValue: acctest.Ct2,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "index_field.*", map[string]string{
						names.AttrName: "literal_test",
						names.AttrType: "literal",
						"facet":        acctest.CtTrue,
						"search":       acctest.CtTrue,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_indexFieldsUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccDomainExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "index_field.#", acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "index_field.*", map[string]string{
						names.AttrName:         "literal_test",
						names.AttrType:         "literal",
						names.AttrDefaultValue: "literally testing",
						"return":               acctest.CtTrue,
						"sort":                 acctest.CtTrue,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "index_field.*", map[string]string{
						names.AttrName:         "double_array_test",
						names.AttrType:         "double-array",
						names.AttrDefaultValue: "-12.34",
						"search":               acctest.CtTrue,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "index_field.*", map[string]string{
						names.AttrName:    "text_test",
						names.AttrType:    "text",
						"analysis_scheme": "_en_default_",
						"highlight":       acctest.CtTrue,
					}),
				),
			},
		},
	})
}

func TestAccCloudSearchDomain_sourceFields(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.DomainStatus
	resourceName := "aws_cloudsearch_domain.test"
	rName := testAccDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudSearchEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_sourceFields(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccDomainExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "index_field.#", acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "index_field.*", map[string]string{
						names.AttrName:         "int_test",
						names.AttrType:         "int",
						names.AttrDefaultValue: acctest.Ct2,
						"source_fields":        "",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "index_field.*", map[string]string{
						names.AttrName:         "int_test_2",
						names.AttrType:         "int",
						names.AttrDefaultValue: acctest.Ct4,
						"source_fields":        "",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "index_field.*", map[string]string{
						names.AttrName:  "int_test_source",
						names.AttrType:  "int-array",
						"source_fields": "int_test,int_test_2",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_sourceFieldsUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccDomainExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "index_field.#", acctest.Ct4),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "index_field.*", map[string]string{
						names.AttrName:         "int_test",
						names.AttrType:         "int",
						names.AttrDefaultValue: acctest.Ct2,
						"source_fields":        "",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "index_field.*", map[string]string{
						names.AttrName:         "int_test_2",
						names.AttrType:         "int",
						names.AttrDefaultValue: acctest.Ct4,
						"source_fields":        "",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "index_field.*", map[string]string{
						names.AttrName:         "int_test_3",
						names.AttrType:         "int",
						names.AttrDefaultValue: "8",
						"source_fields":        "",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "index_field.*", map[string]string{
						names.AttrName:  "int_test_source",
						names.AttrType:  "int-array",
						"source_fields": "int_test_3",
					}),
				),
			},
		},
	})
}

func TestAccCloudSearchDomain_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.DomainStatus
	resourceName := "aws_cloudsearch_domain.test"
	rName := testAccDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudSearchEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_allOptions(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccDomainExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "cloudsearch", fmt.Sprintf("domain/%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "domain_id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "endpoint_options.0.enforce_https", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "endpoint_options.0.tls_security_policy", "Policy-Min-TLS-1-0-2019-07"),
					resource.TestCheckResourceAttr(resourceName, "index_field.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "index_field.*", map[string]string{
						names.AttrName: "latlon_test",
						names.AttrType: "latlon",
					}),
					resource.TestCheckResourceAttr(resourceName, "multi_az", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "scaling_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scaling_parameters.0.desired_instance_type", "search.small"),
					resource.TestCheckResourceAttr(resourceName, "scaling_parameters.0.desired_partition_count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scaling_parameters.0.desired_replication_count", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_allOptionsUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccDomainExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "cloudsearch", fmt.Sprintf("domain/%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "domain_id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "endpoint_options.0.enforce_https", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "endpoint_options.0.tls_security_policy", "Policy-Min-TLS-1-2-2019-07"),
					resource.TestCheckResourceAttr(resourceName, "index_field.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "index_field.*", map[string]string{
						names.AttrName:    "text_array_test",
						names.AttrType:    "text-array",
						"return":          acctest.CtTrue,
						"analysis_scheme": "_fr_default_",
					}),
					resource.TestCheckResourceAttr(resourceName, "multi_az", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "scaling_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scaling_parameters.0.desired_instance_type", "search.medium"),
					resource.TestCheckResourceAttr(resourceName, "scaling_parameters.0.desired_partition_count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scaling_parameters.0.desired_replication_count", acctest.Ct2),
				),
			},
		},
	})
}

func testAccDomainName() string {
	return acctest.ResourcePrefix + "-" + sdkacctest.RandString(28-(len(acctest.ResourcePrefix)+1))
}

func testAccDomainExists(ctx context.Context, n string, v *types.DomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudSearchClient(ctx)

		output, err := tfcloudsearch.FindDomainByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckDomainDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudsearch_domain" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).CloudSearchClient(ctx)

			_, err := tfcloudsearch.FindDomainByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudSearch Domain %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccDomainConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudsearch_domain" "test" {
  name = %[1]q
}
`, rName)
}

func testAccDomainConfig_indexFields(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudsearch_domain" "test" {
  name = %[1]q

  index_field {
    name          = "int_test"
    type          = "int"
    default_value = "2"
  }

  index_field {
    name   = "literal_test"
    type   = "literal"
    facet  = true
    return = false
    search = true
    sort   = false
  }
}
`, rName)
}

func testAccDomainConfig_indexFieldsUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudsearch_domain" "test" {
  name = %[1]q

  index_field {
    name   = "literal_test"
    type   = "literal"
    facet  = false
    return = true
    search = false
    sort   = true

    default_value = "literally testing"
  }

  index_field {
    name   = "double_array_test"
    type   = "double-array"
    search = true

    default_value = "-12.34"
  }

  index_field {
    name            = "text_test"
    type            = "text"
    analysis_scheme = "_en_default_"
    highlight       = true
    search          = true
  }
}
`, rName)
}

func testAccDomainConfig_allOptions(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudsearch_domain" "test" {
  name = %[1]q

  endpoint_options {
    enforce_https       = true
    tls_security_policy = "Policy-Min-TLS-1-0-2019-07"
  }

  multi_az = true

  scaling_parameters {
    desired_instance_type     = "search.small"
    desired_partition_count   = 1
    desired_replication_count = 1
  }

  index_field {
    name = "latlon_test"
    type = "latlon"
  }
}
`, rName)
}

func testAccDomainConfig_sourceFields(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudsearch_domain" "test" {
  name = %[1]q

  index_field {
    name          = "int_test"
    type          = "int"
    default_value = "2"
  }

  index_field {
    name          = "int_test_2"
    type          = "int"
    default_value = "4"
  }

  index_field {
    name = "int_test_source"
    type = "int-array"

    source_fields = "int_test,int_test_2"
  }
}
`, rName)
}

func testAccDomainConfig_sourceFieldsUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudsearch_domain" "test" {
  name = %[1]q

  index_field {
    name          = "int_test"
    type          = "int"
    default_value = "2"
  }

  index_field {
    name          = "int_test_2"
    type          = "int"
    default_value = "4"
  }

  index_field {
    name          = "int_test_3"
    type          = "int"
    default_value = "8"
  }

  index_field {
    name = "int_test_source"
    type = "int-array"

    source_fields = "int_test_3"
  }
}
`, rName)
}

func testAccDomainConfig_allOptionsUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudsearch_domain" "test" {
  name = %[1]q

  endpoint_options {
    enforce_https       = true
    tls_security_policy = "Policy-Min-TLS-1-2-2019-07"
  }

  multi_az = false

  scaling_parameters {
    desired_instance_type     = "search.medium"
    desired_partition_count   = 1
    desired_replication_count = 2
  }

  index_field {
    name            = "text_array_test"
    type            = "text-array"
    return          = true
    search          = true
    analysis_scheme = "_fr_default_"
  }
}
`, rName)
}
