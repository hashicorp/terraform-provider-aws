package cloudsearch_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudsearch"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudsearch "github.com/hashicorp/terraform-provider-aws/internal/service/cloudsearch"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccCloudSearchDomain_basic(t *testing.T) {
	var v cloudsearch.DomainStatus
	resourceName := "aws_cloudsearch_domain.test"
	rName := testAccDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudsearch.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccDomainExists(resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "cloudsearch", fmt.Sprintf("domain/%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "domain_id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_options.0.enforce_https", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_options.0.tls_security_policy"),
					resource.TestCheckResourceAttr(resourceName, "index_field.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "multi_az", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "scaling_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scaling_parameters.0.desired_instance_type", ""),
					resource.TestCheckResourceAttr(resourceName, "scaling_parameters.0.desired_partition_count", "0"),
					resource.TestCheckResourceAttr(resourceName, "scaling_parameters.0.desired_replication_count", "0"),
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
	var v cloudsearch.DomainStatus
	resourceName := "aws_cloudsearch_domain.test"
	rName := testAccDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudsearch.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccDomainExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloudsearch.ResourceDomain(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudSearchDomain_indexFields(t *testing.T) {
	var v cloudsearch.DomainStatus
	resourceName := "aws_cloudsearch_domain.test"
	rName := testAccDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudsearch.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_indexFields(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccDomainExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "index_field.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "index_field.*", map[string]string{
						"name":          "int_test",
						"type":          "int",
						"default_value": "2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "index_field.*", map[string]string{
						"name":   "literal_test",
						"type":   "literal",
						"facet":  "true",
						"search": "true",
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
					testAccDomainExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "index_field.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "index_field.*", map[string]string{
						"name":          "literal_test",
						"type":          "literal",
						"default_value": "literally testing",
						"return":        "true",
						"sort":          "true",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "index_field.*", map[string]string{
						"name":          "double_array_test",
						"type":          "double-array",
						"default_value": "-12.34",
						"search":        "true",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "index_field.*", map[string]string{
						"name":            "text_test",
						"type":            "text",
						"analysis_scheme": "_en_default_",
						"highlight":       "true",
					}),
				),
			},
		},
	})
}

func TestAccCloudSearchDomain_sourceFields(t *testing.T) {
	var v cloudsearch.DomainStatus
	resourceName := "aws_cloudsearch_domain.test"
	rName := testAccDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudsearch.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_sourceFields(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccDomainExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "index_field.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "index_field.*", map[string]string{
						"name":          "int_test",
						"type":          "int",
						"default_value": "2",
						"source_fields": "",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "index_field.*", map[string]string{
						"name":          "int_test_2",
						"type":          "int",
						"default_value": "4",
						"source_fields": "",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "index_field.*", map[string]string{
						"name":          "int_test_source",
						"type":          "int-array",
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
					testAccDomainExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "index_field.#", "4"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "index_field.*", map[string]string{
						"name":          "int_test",
						"type":          "int",
						"default_value": "2",
						"source_fields": "",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "index_field.*", map[string]string{
						"name":          "int_test_2",
						"type":          "int",
						"default_value": "4",
						"source_fields": "",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "index_field.*", map[string]string{
						"name":          "int_test_3",
						"type":          "int",
						"default_value": "8",
						"source_fields": "",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "index_field.*", map[string]string{
						"name":          "int_test_source",
						"type":          "int-array",
						"source_fields": "int_test_3",
					}),
				),
			},
		},
	})
}

func TestAccCloudSearchDomain_update(t *testing.T) {
	var v cloudsearch.DomainStatus
	resourceName := "aws_cloudsearch_domain.test"
	rName := testAccDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudsearch.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_allOptions(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccDomainExists(resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "cloudsearch", fmt.Sprintf("domain/%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "domain_id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_options.0.enforce_https", "true"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_options.0.tls_security_policy", "Policy-Min-TLS-1-0-2019-07"),
					resource.TestCheckResourceAttr(resourceName, "index_field.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "index_field.*", map[string]string{
						"name": "latlon_test",
						"type": "latlon",
					}),
					resource.TestCheckResourceAttr(resourceName, "multi_az", "true"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "scaling_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scaling_parameters.0.desired_instance_type", "search.small"),
					resource.TestCheckResourceAttr(resourceName, "scaling_parameters.0.desired_partition_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "scaling_parameters.0.desired_replication_count", "1"),
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
					testAccDomainExists(resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "cloudsearch", fmt.Sprintf("domain/%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "domain_id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_options.0.enforce_https", "true"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_options.0.tls_security_policy", "Policy-Min-TLS-1-2-2019-07"),
					resource.TestCheckResourceAttr(resourceName, "index_field.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "index_field.*", map[string]string{
						"name":            "text_array_test",
						"type":            "text-array",
						"return":          "true",
						"analysis_scheme": "_fr_default_",
					}),
					resource.TestCheckResourceAttr(resourceName, "multi_az", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "scaling_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scaling_parameters.0.desired_instance_type", "search.medium"),
					resource.TestCheckResourceAttr(resourceName, "scaling_parameters.0.desired_partition_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "scaling_parameters.0.desired_replication_count", "2"),
				),
			},
		},
	})
}

func testAccDomainName() string {
	return acctest.ResourcePrefix + "-" + sdkacctest.RandString(28-(len(acctest.ResourcePrefix)+1))
}

func testAccDomainExists(n string, v *cloudsearch.DomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No CloudSearch Domain ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudSearchConn

		output, err := tfcloudsearch.FindDomainStatusByName(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccDomainDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudsearch_domain" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudSearchConn

		_, err := tfcloudsearch.FindDomainStatusByName(conn, rs.Primary.ID)

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
