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
	rName := acctest.ResourcePrefix + "-" + sdkacctest.RandString(28-(len(acctest.ResourcePrefix)+1))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudsearch.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudsearch.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCloudSearchDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudSearchDomainConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCloudSearchDomainExists(resourceName, &v),
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
	rName := acctest.ResourcePrefix + "-" + sdkacctest.RandString(28-(len(acctest.ResourcePrefix)+1))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudsearch.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudsearch.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCloudSearchDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudSearchDomainConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCloudSearchDomainExists(resourceName, &v),
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
	rName := acctest.ResourcePrefix + "-" + sdkacctest.RandString(28-(len(acctest.ResourcePrefix)+1))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudsearch.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudsearch.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCloudSearchDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudSearchDomainIndexFieldsConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCloudSearchDomainExists(resourceName, &v),
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
		},
	})
}

func TestAccCloudSearchDomain_simple(t *testing.T) {
	var v cloudsearch.DomainStatus
	resourceName := "aws_cloudsearch_domain.test"
	rName := acctest.ResourcePrefix + "-" + sdkacctest.RandString(28-(len(acctest.ResourcePrefix)+1))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudsearch.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudsearch.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCloudSearchDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudSearchDomainConfig_simple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCloudSearchDomainExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCloudSearchDomainConfig_basicIndexMix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCloudSearchDomainExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func TestAccCloudSearchDomain_textAnalysisScheme(t *testing.T) {
	var v cloudsearch.DomainStatus
	resourceName := "aws_cloudsearch_domain.test"
	rName := acctest.ResourcePrefix + "-" + sdkacctest.RandString(28-(len(acctest.ResourcePrefix)+1))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudsearch.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudsearch.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCloudSearchDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudSearchDomainConfig_textAnalysisScheme(rName, "_en_default_"),
				Check: resource.ComposeTestCheckFunc(
					testAccCloudSearchDomainExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", acctest.RandomFQDomainName()),
				),
			},
			{
				Config: testAccCloudSearchDomainConfig_textAnalysisScheme(rName, "_fr_default_"),
				Check: resource.ComposeTestCheckFunc(
					testAccCloudSearchDomainExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func testAccCloudSearchDomainExists(n string, v *cloudsearch.DomainStatus) resource.TestCheckFunc {
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

func testAccCloudSearchDomainDestroy(s *terraform.State) error {
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

func testAccCloudSearchDomainConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudsearch_domain" "test" {
  name = %[1]q
}
`, rName)
}

func testAccCloudSearchDomainIndexFieldsConfig(rName string) string {
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

func testAccCloudSearchDomainConfig_simple(name string) string {
	return fmt.Sprintf(`
resource "aws_cloudsearch_domain" "test" {
  name = "%s"

  index {
    name   = "date_test"
    type   = "date"
    facet  = true
    search = true
    return = true
    sort   = true
  }

  index {
    name   = "date_array_test"
    type   = "date-array"
    facet  = true
    search = true
    return = true
  }

  index {
    name   = "double_test"
    type   = "double"
    facet  = true
    search = true
    return = true
    sort   = true
  }

  index {
    name   = "double_array_test"
    type   = "double-array"
    facet  = true
    search = true
    return = true
  }

  index {
    name   = "int_test"
    type   = "int"
    facet  = true
    search = true
    return = true
    sort   = true
  }

  index {
    name   = "int_array_test"
    type   = "int-array"
    facet  = true
    search = true
    return = true
  }

  index {
    name   = "latlon_test"
    type   = "latlon"
    facet  = true
    search = true
    return = true
    sort   = true
  }

  index {
    name   = "literal_test"
    type   = "literal"
    facet  = true
    search = true
    return = true
    sort   = true
  }

  index {
    name   = "literal_array_test"
    type   = "literal-array"
    facet  = true
    search = true
    return = true
  }

  index {
    name            = "text_test"
    type            = "text"
    analysis_scheme = "_en_default_"
    highlight       = true
    return          = true
    sort            = true
  }

  index {
    name            = "text_array_test"
    type            = "text-array"
    analysis_scheme = "_en_default_"
    highlight       = true
    return          = true
  }

  wait_for_endpoints      = false
  service_access_policies = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "AWS": ["*"]
    },
    "Action": ["cloudsearch:*"]
  }]
}
EOF
}
`, name)
}

func testAccCloudSearchDomainConfig_basicIndexMix(name string) string {
	return fmt.Sprintf(`
resource "aws_cloudsearch_domain" "test" {
  name = "%s"

  index {
    name            = "how_about_one_up_here"
    type            = "text"
    analysis_scheme = "_en_default_"
  }

  index {
    name   = "date_test"
    type   = "date"
    facet  = true
    search = true
    return = true
    sort   = true
  }

  index {
    name   = "double_test_2"
    type   = "double"
    facet  = true
    search = true
    return = true
    sort   = true
  }

  index {
    name   = "double_array_test"
    type   = "double-array"
    facet  = true
    search = true
    return = true
  }

  index {
    name   = "just_another_index_name"
    type   = "literal-array"
    facet  = true
    search = true
    return = true
  }

  index {
    name            = "text_test"
    type            = "text"
    analysis_scheme = "_en_default_"
    highlight       = true
    return          = true
    sort            = true
  }

  index {
    name = "captain_janeway_is_pretty_cool"
    type = "double"
  }

  wait_for_endpoints      = false
  service_access_policies = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "AWS": ["*"]
    },
    "Action": ["cloudsearch:*"]
  }]
}
EOF
}
`, name)
}

// NOTE: I'd like to get text and text arrays field to work properly without having to explicitly set the
// `analysis_scheme` field, but I cannot find a way to suppress the diff Terraform ends up generating as a result.
func testAccCloudSearchDomainConfig_textAnalysisScheme(name string, scheme string) string {
	return fmt.Sprintf(`
resource "aws_cloudsearch_domain" "test" {
  name = "%s"

  #index {
  #  name = "use_default_scheme"
  #  type = "text"
  #}

  index {
    name            = "specify_scheme"
    type            = "text"
    analysis_scheme = "%s"
  }

  wait_for_endpoints      = false
  service_access_policies = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "AWS": ["*"]
    },
    "Action": ["cloudsearch:*"]
  }]
}
EOF
}
`, name, scheme)
}

func testAccCloudSearchDomainConfig_withInstanceType(name string, instance_type string) string {
	return fmt.Sprintf(`
resource "aws_cloudsearch_domain" "test" {
  name = "%s"

  instance_type = "%s"

  wait_for_endpoints      = false
  service_access_policies = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "AWS": ["*"]
    },
    "Action": ["cloudsearch:*"]
  }]
}
EOF
}
`, name, instance_type)
}

func testAccCloudSearchDomainConfig_withIndex(name string, index_name string, index_type string) string {
	return fmt.Sprintf(`
resource "aws_cloudsearch_domain" "test" {
  name = "%s"

  index {
    name            = "%s"
    type            = "%s"
    facet           = false
    search          = false
    return          = true
    sort            = true
    highlight       = false
    analysis_scheme = "_en_default_"
  }

  wait_for_endpoints      = false
  service_access_policies = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "AWS": ["*"]
    },
    "Action": ["cloudsearch:*"]
  }]
}
EOF
}
`, name, index_name, index_type)
}
