// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/redshiftserverless"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfredshiftserverless "github.com/hashicorp/terraform-provider-aws/internal/service/redshiftserverless"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccRedshiftServerlessWorkgroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshiftserverless_workgroup.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, redshiftserverless.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkgroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkgroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkgroupExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "redshift-serverless", regexp.MustCompile("workgroup/.+$")),
					resource.TestCheckResourceAttr(resourceName, "namespace_name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "workgroup_id"),
					resource.TestCheckResourceAttr(resourceName, "workgroup_name", rName),
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

func TestAccRedshiftServerlessWorkgroup_baseCapacityAndPubliclyAccessible(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshiftserverless_workgroup.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, redshiftserverless.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkgroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkgroupConfig_baseCapacityAndPubliclyAccessible(rName, 64, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkgroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "base_capacity", "64"),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWorkgroupConfig_baseCapacityAndPubliclyAccessible(rName, 128, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkgroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "base_capacity", "128"),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "false"),
				),
			},
		},
	})
}

func TestAccRedshiftServerlessWorkgroup_configParameters(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshiftserverless_workgroup.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, redshiftserverless.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkgroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkgroupConfig_configParameters(rName, "14400"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkgroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "config_parameter.#", "7"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "datestyle",
						"parameter_value": "ISO, MDY",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "enable_user_activity_logging",
						"parameter_value": "true",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "query_group",
						"parameter_value": "default",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "search_path",
						"parameter_value": "$user, public",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "max_query_execution_time",
						"parameter_value": "14400",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "auto_mv",
						"parameter_value": "true",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "enable_case_sensitive_identifier",
						"parameter_value": "false",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWorkgroupConfig_configParameters(rName, "28800"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkgroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "config_parameter.#", "7"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "datestyle",
						"parameter_value": "ISO, MDY",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "enable_user_activity_logging",
						"parameter_value": "true",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "query_group",
						"parameter_value": "default",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "search_path",
						"parameter_value": "$user, public",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "max_query_execution_time",
						"parameter_value": "28800",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "auto_mv",
						"parameter_value": "true",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "enable_case_sensitive_identifier",
						"parameter_value": "false",
					}),
				),
			},
		},
	})
}

func TestAccRedshiftServerlessWorkgroup_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshiftserverless_workgroup.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, redshiftserverless.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkgroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkgroupConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkgroupExists(ctx, resourceName),
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
				Config: testAccWorkgroupConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccWorkgroupConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkgroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccRedshiftServerlessWorkgroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshiftserverless_workgroup.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, redshiftserverless.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkgroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkgroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkgroupExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfredshiftserverless.ResourceWorkgroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckWorkgroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftServerlessConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_redshiftserverless_workgroup" {
				continue
			}
			_, err := tfredshiftserverless.FindWorkgroupByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Redshift Serverless Workgroup %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckWorkgroupExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Redshift Serverless Workgroup ID is not set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftServerlessConn(ctx)

		_, err := tfredshiftserverless.FindWorkgroupByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccWorkgroupConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_redshiftserverless_namespace" "test" {
  namespace_name = %[1]q
}

resource "aws_redshiftserverless_workgroup" "test" {
  namespace_name = aws_redshiftserverless_namespace.test.namespace_name
  workgroup_name = %[1]q
}
`, rName)
}

func testAccWorkgroupConfig_baseCapacityAndPubliclyAccessible(rName string, baseCapacity int, publiclyAccessible bool) string {
	return fmt.Sprintf(`
resource "aws_redshiftserverless_namespace" "test" {
  namespace_name = %[1]q
}

resource "aws_redshiftserverless_workgroup" "test" {
  namespace_name      = aws_redshiftserverless_namespace.test.namespace_name
  workgroup_name      = %[1]q
  base_capacity       = %[2]d
  publicly_accessible = %[3]t
}
`, rName, baseCapacity, publiclyAccessible)
}

func testAccWorkgroupConfig_configParameters(rName, maxQueryExecutionTime string) string {
	return fmt.Sprintf(`
resource "aws_redshiftserverless_namespace" "test" {
  namespace_name = %[1]q
}

resource "aws_redshiftserverless_workgroup" "test" {
  namespace_name = aws_redshiftserverless_namespace.test.namespace_name
  workgroup_name = %[1]q

  config_parameter {
    parameter_key   = "datestyle"
    parameter_value = "ISO, MDY"
  }
  config_parameter {
    parameter_key   = "enable_user_activity_logging"
    parameter_value = "true"
  }
  config_parameter {
    parameter_key   = "query_group"
    parameter_value = "default"
  }
  config_parameter {
    parameter_key   = "search_path"
    parameter_value = "$user, public"
  }
  config_parameter {
    parameter_key   = "max_query_execution_time"
    parameter_value = %[2]q
  }
  config_parameter {
    parameter_key   = "auto_mv"
    parameter_value = "true"
  }
  config_parameter {
    parameter_key   = "enable_case_sensitive_identifier"
    parameter_value = "false"
  }
}
`, rName, maxQueryExecutionTime)
}

func testAccWorkgroupConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_redshiftserverless_namespace" "test" {
  namespace_name = %[1]q
}

resource "aws_redshiftserverless_workgroup" "test" {
  namespace_name = aws_redshiftserverless_namespace.test.namespace_name
  workgroup_name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccWorkgroupConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_redshiftserverless_namespace" "test" {
  namespace_name = %[1]q
}

resource "aws_redshiftserverless_workgroup" "test" {
  namespace_name = aws_redshiftserverless_namespace.test.namespace_name
  workgroup_name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
