package inspector2_test

import (
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	//
	// The provider linter wants your imports to be in two groups: first,
	// standard library (i.e., "fmt" or "strings"), second, everything else.
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/inspector2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfinspector2 "github.com/hashicorp/terraform-provider-aws/internal/service/inspector2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccInspector2Filter_basic(t *testing.T) {
	var filter inspector2.ListFiltersOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_inspector2_filter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(inspector2.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, inspector2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFilterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFilterExists(resourceName, &filter),
					resource.TestCheckResourceAttr(resourceName, "action", "SUPPRESS"),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "reason", "reason"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.*", map[string]string{
						"aws_account_id.0.comparison":                          "EQUALS",
						"aws_account_id.0.value":                               "aws_account_id",
						"component_id.0.comparison":                            "EQUALS",
						"component_id.0.value":                                 "component_id",
						"component_type.0.comparison":                          "EQUALS",
						"component_type.0.value":                               "component_type",
						"ec2_instance_image_id.0.comparison":                   "EQUALS",
						"ec2_instance_image_id.0.value":                        "ec2_instance_image_id",
						"ec2_instance_subnet_id.0.comparison":                  "EQUALS",
						"ec2_instance_subnet_id.0.value":                       "ec2_instance_subnet_id",
						"ec2_instance_vpc_id.0.comparison":                     "EQUALS",
						"ec2_instance_vpc_id.0.value":                          "ec2_instance_vpc_id",
						"ecr_image_architecture.0.comparison":                  "EQUALS",
						"ecr_image_architecture.0.value":                       "ecr_image_architecture",
						"ecr_image_hash.0.comparison":                          "EQUALS",
						"ecr_image_hash.0.value":                               "ecr_image_hash",
						"ecr_image_pushed_at.0.end_inclusive":                  "2021-01-31T00:00:00Z",
						"ecr_image_pushed_at.0.start_inclusive":                "2021-01-01T00:00:00Z",
						"ecr_image_registry.0.comparison":                      "EQUALS",
						"ecr_image_registry.0.value":                           "ecr_image_registry",
						"ecr_image_repository_name.0.comparison":               "EQUALS",
						"ecr_image_repository_name.0.value":                    "ecr_image_repository_name",
						"ecr_image_tags.0.comparison":                          "EQUALS",
						"ecr_image_tags.0.value":                               "ecr_image_tags",
						"finding_arn.0.comparison":                             "EQUALS",
						"finding_arn.0.value":                                  "finding_arn",
						"finding_status.0.comparison":                          "EQUALS",
						"finding_status.0.value":                               "SUPPRESSED",
						"finding_type.0.comparison":                            "EQUALS",
						"finding_type.0.value":                                 "finding_type",
						"first_observed_at.0.end_inclusive":                    "2021-01-31T00:00:00Z",
						"first_observed_at.0.start_inclusive":                  "2021-01-01T00:00:00Z",
						"inspector_score.0.lower_inclusive":                    "1",
						"inspector_score.0.upper_inclusive":                    "10",
						"last_observed_at.0.end_inclusive":                     "2021-01-31T00:00:00Z",
						"last_observed_at.0.start_inclusive":                   "2021-01-01T00:00:00Z",
						"network_protocol.0.comparison":                        "EQUALS",
						"network_protocol.0.value":                             "network_protocol",
						"port_range.0.begin_inclusive":                         "80",
						"port_range.0.end_inclusive":                           "443",
						"related_vulnerabilities.0.comparison":                 "EQUALS",
						"related_vulnerabilities.0.value":                      "related_vulnerabilities",
						"resource_id.0.comparison":                             "EQUALS",
						"resource_id.0.value":                                  "resource_id",
						"resource_tags.0.comparison":                           "EQUALS",
						"resource_tags.0.value":                                "resource_tags",
						"resource_tags.0.key":                                  "resource_tags",
						"resource_type.0.comparison":                           "EQUALS",
						"resource_type.0.value":                                "AWS_EC2_INSTANCE",
						"severity.0.comparison":                                "EQUALS",
						"severity.0.value":                                     "INFORMATIONAL",
						"updated_at.0.end_inclusive":                           "2021-01-31T00:00:00Z",
						"updated_at.0.start_inclusive":                         "2021-01-01T00:00:00Z",
						"vendor_severity.0.comparison":                         "EQUALS",
						"vendor_severity.0.value":                              "vendor_severity",
						"vulnerability_id.0.comparison":                        "EQUALS",
						"vulnerability_id.0.value":                             "vulnerability_id",
						"vulnerability_source.0.comparison":                    "EQUALS",
						"vulnerability_source.0.value":                         "vulnerability_source",
						"vulnerable_packages.0.architecture.0.comparison":      "EQUALS",
						"vulnerable_packages.0.epoch.0.lower_inclusive":        "2",
						"vulnerable_packages.0.epoch.0.upper_inclusive":        "1",
						"vulnerable_packages.0.name.0.comparison":              "EQUALS",
						"vulnerable_packages.0.name.0.value":                   "name",
						"vulnerable_packages.0.release.0.comparison":           "EQUALS",
						"vulnerable_packages.0.release.0.value":                "release",
						"vulnerable_packages.0.source_layer_hash.0.comparison": "EQUALS",
						"vulnerable_packages.0.source_layer_hash.0.value":      "source_layer_hash",
						"vulnerable_packages.0.version.0.comparison":           "EQUALS",
						"vulnerable_packages.0.version.0.value":                "version",
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "inspector2", regexp.MustCompile(`owner/.+$`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFilterConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFilterExists(resourceName, &filter),
					resource.TestCheckResourceAttr(resourceName, "action", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "description", "update_description"),
					resource.TestCheckResourceAttr(resourceName, "reason", "update_reason"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.*", map[string]string{
						"aws_account_id.0.comparison":                          "NOT_EQUALS",
						"aws_account_id.0.value":                               "not_aws_account_id",
						"component_id.0.comparison":                            "NOT_EQUALS",
						"component_id.0.value":                                 "not_component_id",
						"component_type.0.comparison":                          "NOT_EQUALS",
						"component_type.0.value":                               "not_component_type",
						"ec2_instance_image_id.0.comparison":                   "NOT_EQUALS",
						"ec2_instance_image_id.0.value":                        "not_ec2_instance_image_id",
						"ec2_instance_subnet_id.0.comparison":                  "NOT_EQUALS",
						"ec2_instance_subnet_id.0.value":                       "not_ec2_instance_subnet_id",
						"ec2_instance_vpc_id.0.comparison":                     "NOT_EQUALS",
						"ec2_instance_vpc_id.0.value":                          "not_ec2_instance_vpc_id",
						"ecr_image_architecture.0.comparison":                  "NOT_EQUALS",
						"ecr_image_architecture.0.value":                       "not_ecr_image_architecture",
						"ecr_image_hash.0.comparison":                          "NOT_EQUALS",
						"ecr_image_hash.0.value":                               "not_ecr_image_hash",
						"ecr_image_pushed_at.0.end_inclusive":                  "2021-03-31T00:00:00Z",
						"ecr_image_pushed_at.0.start_inclusive":                "2021-03-01T00:00:00Z",
						"ecr_image_registry.0.comparison":                      "NOT_EQUALS",
						"ecr_image_registry.0.value":                           "not_ecr_image_registry",
						"ecr_image_repository_name.0.comparison":               "NOT_EQUALS",
						"ecr_image_repository_name.0.value":                    "not_ecr_image_repository_name",
						"ecr_image_tags.0.comparison":                          "NOT_EQUALS",
						"ecr_image_tags.0.value":                               "not_ecr_image_tags",
						"finding_arn.0.comparison":                             "NOT_EQUALS",
						"finding_arn.0.value":                                  "not_finding_arn",
						"finding_status.0.comparison":                          "NOT_EQUALS",
						"finding_status.0.value":                               "CLOSED",
						"finding_type.0.comparison":                            "NOT_EQUALS",
						"finding_type.0.value":                                 "not_finding_type",
						"first_observed_at.0.end_inclusive":                    "2021-03-31T00:00:00Z",
						"first_observed_at.0.start_inclusive":                  "2021-03-01T00:00:00Z",
						"inspector_score.0.lower_inclusive":                    "10",
						"inspector_score.0.upper_inclusive":                    "100",
						"last_observed_at.0.end_inclusive":                     "2021-03-31T00:00:00Z",
						"last_observed_at.0.start_inclusive":                   "2021-03-01T00:00:00Z",
						"network_protocol.0.comparison":                        "NOT_EQUALS",
						"network_protocol.0.value":                             "not_network_protocol",
						"port_range.0.begin_inclusive":                         "8080",
						"port_range.0.end_inclusive":                           "8443",
						"related_vulnerabilities.0.comparison":                 "NOT_EQUALS",
						"related_vulnerabilities.0.value":                      "not_related_vulnerabilities",
						"resource_id.0.comparison":                             "NOT_EQUALS",
						"resource_id.0.value":                                  "not_resource_id",
						"resource_tags.0.comparison":                           "EQUALS",
						"resource_tags.0.value":                                "not_resource_tags",
						"resource_tags.0.key":                                  "not_resource_tags",
						"resource_type.0.comparison":                           "NOT_EQUALS",
						"resource_type.0.value":                                "AWS_ECR_CONTAINER_IMAGE",
						"severity.0.comparison":                                "NOT_EQUALS",
						"severity.0.value":                                     "LOW",
						"updated_at.0.end_inclusive":                           "2021-03-31T00:00:00Z",
						"updated_at.0.start_inclusive":                         "2021-03-01T00:00:00Z",
						"vendor_severity.0.comparison":                         "NOT_EQUALS",
						"vendor_severity.0.value":                              "not_vendor_severity",
						"vulnerability_id.0.comparison":                        "NOT_EQUALS",
						"vulnerability_id.0.value":                             "not_vulnerability_id",
						"vulnerability_source.0.comparison":                    "NOT_EQUALS",
						"vulnerability_source.0.value":                         "not_vulnerability_source",
						"vulnerable_packages.0.architecture.0.comparison":      "NOT_EQUALS",
						"vulnerable_packages.0.architecture.0.value":           "not_architecture",
						"vulnerable_packages.0.epoch.0.lower_inclusive":        "3",
						"vulnerable_packages.0.epoch.0.upper_inclusive":        "2",
						"vulnerable_packages.0.name.0.comparison":              "NOT_EQUALS",
						"vulnerable_packages.0.name.0.value":                   "not_name",
						"vulnerable_packages.0.release.0.comparison":           "NOT_EQUALS",
						"vulnerable_packages.0.release.0.value":                "not_release",
						"vulnerable_packages.0.source_layer_hash.0.comparison": "NOT_EQUALS",
						"vulnerable_packages.0.source_layer_hash.0.value":      "not_source_layer_hash",
						"vulnerable_packages.0.version.0.comparison":           "NOT_EQUALS",
						"vulnerable_packages.0.version.0.value":                "not_version",
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "inspector2", regexp.MustCompile(`owner/.+$`)),
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

func TestAccInspector2Filter_nameGenerated(t *testing.T) {
	var filter inspector2.ListFiltersOutput
	resourceName := "aws_inspector2_filter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, inspector2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFilterConfig_nameGenerated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFilterExists(resourceName, &filter),
					acctest.CheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", resource.UniqueIdPrefix),
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

func TestAccInspector2Filter_namePrefix(t *testing.T) {
	var filter inspector2.ListFiltersOutput
	resourceName := "aws_inspector2_filter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, inspector2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFilterConfig_namePrefix("tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFilterExists(resourceName, &filter),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, "name", "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "tf-acc-test-prefix-"),
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

func TestAccInspector2Filter_tags(t *testing.T) {
	var filter1, filter2, filter3 inspector2.ListFiltersOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_inspector2_filter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, inspector2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFilterConfig_Tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFilterExists(resourceName, &filter1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFilterConfig_Tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFilterExists(resourceName, &filter2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccFilterConfig_Tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFilterExists(resourceName, &filter3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccInspector2Filter_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var filter inspector2.ListFiltersOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_inspector2_filter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(inspector2.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, inspector2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFilterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFilterExists(resourceName, &filter),
					acctest.CheckResourceDisappears(acctest.Provider, tfinspector2.ResourceFilter(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFilterDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).Inspector2Conn
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_inspector2_filter" {
			continue
		}

		out, err := conn.ListFiltersWithContext(ctx, &inspector2.ListFiltersInput{
			Arns: []*string{aws.String(rs.Primary.ID)},
		})

		if err != nil {
			if tfawserr.ErrCodeEquals(err, inspector2.ErrCodeResourceNotFoundException) {
				return nil
			}
			return err
		}

		if len(out.Filters) == 0 {
			return nil
		}

		return create.Error(names.Inspector2, create.ErrActionCheckingDestroyed, tfinspector2.ResNameFilter, rs.Primary.ID, errors.New("not destroyed"))
	}

	return nil
}

func testAccCheckFilterExists(name string, filter *inspector2.ListFiltersOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Inspector2, create.ErrActionCheckingExistence, tfinspector2.ResNameFilter, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Inspector2, create.ErrActionCheckingExistence, tfinspector2.ResNameFilter, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Inspector2Conn
		ctx := context.Background()
		resp, err := conn.ListFiltersWithContext(ctx, &inspector2.ListFiltersInput{
			Arns: []*string{aws.String(rs.Primary.ID)},
		})

		if err != nil {
			return create.Error(names.Inspector2, create.ErrActionCheckingExistence, tfinspector2.ResNameFilter, rs.Primary.ID, err)
		}

		*filter = *resp

		return nil
	}
}

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).Inspector2Conn
	ctx := context.Background()

	input := &inspector2.ListFiltersInput{}
	_, err := conn.ListFiltersWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccFilterConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_inspector2_filter" "test" {
  name   = %[1]q
  action = "SUPPRESS"

  description = "description"
  reason = "reason"

  filter_criteria {
    aws_account_id {
      comparison = "EQUALS"
      value      = "aws_account_id"
    }

    component_id {
      comparison = "EQUALS"
      value      = "component_id"
    }

    component_type {
      comparison = "EQUALS"
      value      = "component_type"
    }

    ec2_instance_image_id {
      comparison = "EQUALS"
      value      = "ec2_instance_image_id"
    }

    ec2_instance_subnet_id {
      comparison = "EQUALS"
      value      = "ec2_instance_subnet_id"
    }

    ec2_instance_vpc_id {
      comparison = "EQUALS"
      value      = "ec2_instance_vpc_id"
    }

    ecr_image_architecture {
      comparison = "EQUALS"
      value      = "ecr_image_architecture"
    }

    ecr_image_hash {
      comparison = "EQUALS"
      value      = "ecr_image_hash"
    }

    ecr_image_pushed_at {
      end_inclusive   = "2021-01-31T00:00:00Z"
      start_inclusive = "2021-01-01T00:00:00Z"
    }

    ecr_image_registry {
      comparison = "EQUALS"
      value      = "ecr_image_registry"
    }

    ecr_image_repository_name {
      comparison = "EQUALS"
      value      = "ecr_image_repository_name"
    }

    ecr_image_tags {
      comparison = "EQUALS"
      value      = "ecr_image_tags"
    }

    finding_arn {
      comparison = "EQUALS"
      value      = "finding_arn"
    }

    finding_status {
      comparison = "EQUALS"
      value      = "SUPPRESSED"
    }

    finding_type {
      comparison = "EQUALS"
      value      = "finding_type"
    }

    first_observed_at {
      end_inclusive   = "2021-01-31T00:00:00Z"
      start_inclusive = "2021-01-01T00:00:00Z"
    }

    inspector_score {
      lower_inclusive = "1"
      upper_inclusive = "10"
    }

    last_observed_at {
      end_inclusive   = "2021-01-31T00:00:00Z"
      start_inclusive = "2021-01-01T00:00:00Z"
    }

    network_protocol {
      comparison = "EQUALS"
      value      = "network_protocol"
    }

    port_range {
      begin_inclusive = "80"
      end_inclusive   = "443"
    }

    related_vulnerabilities {
      comparison = "EQUALS"
      value      = "related_vulnerabilities"
    }

    resource_id {
      comparison = "EQUALS"
      value      = "resource_id"
    }

    resource_tags {
      comparison = "EQUALS"
      key        = "resource_tags"
      value      = "resource_tags"
    }

    resource_type {
      comparison = "EQUALS"
      value      = "AWS_EC2_INSTANCE"
    }

    severity {
      comparison = "EQUALS"
      value      = "INFORMATIONAL"
    }

    updated_at {
      end_inclusive   = "2021-01-31T00:00:00Z"
      start_inclusive = "2021-01-01T00:00:00Z"
    }

    vendor_severity {
      comparison = "EQUALS"
      value      = "vendor_severity"
    }

    vulnerability_id {
      comparison = "EQUALS"
      value      = "vulnerability_id"
    }

    vulnerability_source {
      comparison = "EQUALS"
      value      = "vulnerability_source"
    }

    vulnerable_packages {
      architecture {
        comparison = "EQUALS"
        value      = "architecture"
      }
      epoch {
        lower_inclusive = "2"
        upper_inclusive = "1"
      }
      name {
        comparison = "EQUALS"
        value      = "name"
      }
      release {
        comparison = "EQUALS"
        value      = "release"
      }
      source_layer_hash {
        comparison = "EQUALS"
        value      = "source_layer_hash"
      }
      version {
        comparison = "EQUALS"
        value      = "version"
      }
    }
  }
}
`, rName)
}

func testAccFilterConfig_update(rName string) string {
	return fmt.Sprintf(`
resource "aws_inspector2_filter" "test" {
  name   = %[1]q
  action = "NONE"

  description = "update_description"
  reason = "update_reason"

  filter_criteria {
    aws_account_id {
      comparison = "NOT_EQUALS"
      value      = "not_aws_account_id"
    }

    component_id {
      comparison = "NOT_EQUALS"
      value      = "not_component_id"
    }

    component_type {
      comparison = "NOT_EQUALS"
      value      = "not_component_type"
    }

    ec2_instance_image_id {
      comparison = "NOT_EQUALS"
      value      = "not_ec2_instance_image_id"
    }

    ec2_instance_subnet_id {
      comparison = "NOT_EQUALS"
      value      = "not_ec2_instance_subnet_id"
    }

    ec2_instance_vpc_id {
      comparison = "NOT_EQUALS"
      value      = "not_ec2_instance_vpc_id"
    }

    ecr_image_architecture {
      comparison = "NOT_EQUALS"
      value      = "not_ecr_image_architecture"
    }

    ecr_image_hash {
      comparison = "NOT_EQUALS"
      value      = "not_ecr_image_hash"
    }

    ecr_image_pushed_at {
      end_inclusive   = "2021-03-31T00:00:00Z"
      start_inclusive = "2021-03-01T00:00:00Z"
    }

    ecr_image_registry {
      comparison = "NOT_EQUALS"
      value      = "not_ecr_image_registry"
    }

    ecr_image_repository_name {
      comparison = "NOT_EQUALS"
      value      = "not_ecr_image_repository_name"
    }

    ecr_image_tags {
      comparison = "NOT_EQUALS"
      value      = "not_ecr_image_tags"
    }

    finding_arn {
      comparison = "NOT_EQUALS"
      value      = "not_finding_arn"
    }

    finding_status {
      comparison = "NOT_EQUALS"
      value      = "CLOSED"
    }

    finding_type {
      comparison = "NOT_EQUALS"
      value      = "not_finding_type"
    }

    first_observed_at {
      end_inclusive   = "2021-03-31T00:00:00Z"
      start_inclusive = "2021-03-01T00:00:00Z"
    }

    inspector_score {
      lower_inclusive = "10"
      upper_inclusive = "100"
    }

    last_observed_at {
      end_inclusive   = "2021-03-31T00:00:00Z"
      start_inclusive = "2021-03-01T00:00:00Z"
    }

    network_protocol {
      comparison = "NOT_EQUALS"
      value      = "not_network_protocol"
    }

    port_range {
      begin_inclusive = "8080"
      end_inclusive   = "8443"
    }

    related_vulnerabilities {
      comparison = "NOT_EQUALS"
      value      = "not_related_vulnerabilities"
    }

    resource_id {
      comparison = "NOT_EQUALS"
      value      = "not_resource_id"
    }

    resource_tags {
      comparison = "EQUALS"
      key        = "not_resource_tags"
      value      = "not_resource_tags"
    }

    resource_type {
      comparison = "NOT_EQUALS"
      value      = "AWS_ECR_CONTAINER_IMAGE"
    }

    severity {
      comparison = "NOT_EQUALS"
      value      = "LOW"
    }

    updated_at {
      end_inclusive   = "2021-03-31T00:00:00Z"
      start_inclusive = "2021-03-01T00:00:00Z"
    }

    vendor_severity {
      comparison = "NOT_EQUALS"
      value      = "not_vendor_severity"
    }

    vulnerability_id {
      comparison = "NOT_EQUALS"
      value      = "not_vulnerability_id"
    }

    vulnerability_source {
      comparison = "NOT_EQUALS"
      value      = "not_vulnerability_source"
    }

    vulnerable_packages {
      architecture {
        comparison = "NOT_EQUALS"
        value      = "not_architecture"
      }
      epoch {
        lower_inclusive = "3"
        upper_inclusive = "2"
      }
      name {
        comparison = "NOT_EQUALS"
        value      = "not_name"
      }
      release {
        comparison = "NOT_EQUALS"
        value      = "not_release"
      }
      source_layer_hash {
        comparison = "NOT_EQUALS"
        value      = "not_source_layer_hash"
      }
      version {
        comparison = "NOT_EQUALS"
        value      = "not_version"
      }
    }
  }
}
`, rName)
}

func testAccFilterConfig_nameGenerated() string {
	return `
resource "aws_inspector2_filter" "test" {
  action = "SUPPRESS"

  filter_criteria {
    aws_account_id {
      comparison = "EQUALS"
      value      = "aws_account_id"
    }
  }
}
`
}

func testAccFilterConfig_namePrefix(namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_inspector2_filter" "test" {
  name_prefix = %[1]q

  action = "SUPPRESS"

  filter_criteria {
    aws_account_id {
      comparison = "EQUALS"
      value      = "aws_account_id"
    }
  }
}
`, namePrefix)
}

func testAccFilterConfig_Tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_inspector2_filter" "test" {
  name = %[1]q

  action = "SUPPRESS"

  filter_criteria {
    aws_account_id {
      comparison = "EQUALS"
      value      = "aws_account_id"
    }
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccFilterConfig_Tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_inspector2_filter" "test" {
  name     = %[1]q

  action = "SUPPRESS"

  filter_criteria {
    aws_account_id {
      comparison = "EQUALS"
      value      = "aws_account_id"
    }
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
