// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfredshiftserverless "github.com/hashicorp/terraform-provider-aws/internal/service/redshiftserverless"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRedshiftServerlessWorkgroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshiftserverless_workgroup.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkgroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkgroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkgroupExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "redshift-serverless", regexache.MustCompile("workgroup/.+$")),
					resource.TestCheckResourceAttr(resourceName, "namespace_name", rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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

// Tests the complex logic involved in updating 'base_capacity' and 'max_capacity'.
// The order of updates is crucial and is determined by their current state values.
// Note, if the default VPC does not have an internet gateway, enabling public accessibility will silently fail,
// causing the test to fail.
// TODO: break up this test into specific tests on a specific transition
func TestAccRedshiftServerlessWorkgroup_baseAndMaxCapacityAndPubliclyAccessible(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshiftserverless_workgroup.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkgroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkgroupConfig_baseCapacity(rName, 128),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkgroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "base_capacity", "128"),
					resource.TestCheckResourceAttr(resourceName, names.AttrMaxCapacity, "0"),
				),
			},
			{
				Config: testAccWorkgroupConfig_baseCapacity(rName, 256),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "base_capacity", "256"),
					resource.TestCheckResourceAttr(resourceName, names.AttrMaxCapacity, "0"),
				),
			},
			{
				Config: testAccWorkgroupConfig_baseAndMaxCapacityAndPubliclyAccessible(rName, 64, 128, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "base_capacity", "64"),
					resource.TestCheckResourceAttr(resourceName, names.AttrMaxCapacity, "128"),
				),
			},
			{
				Config: testAccWorkgroupConfig_baseAndMaxCapacityAndPubliclyAccessible(rName, 128, 256, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "base_capacity", "128"),
					resource.TestCheckResourceAttr(resourceName, names.AttrMaxCapacity, "256"),
				),
			},
			{
				Config: testAccWorkgroupConfig_baseAndMaxCapacityAndPubliclyAccessible(rName, 512, 5632, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "base_capacity", "512"),
					resource.TestCheckResourceAttr(resourceName, names.AttrMaxCapacity, "5632"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPubliclyAccessible, acctest.CtFalse),
				),
			},
			{
				Config: testAccWorkgroupConfig_baseAndMaxCapacityAndPubliclyAccessible(rName, 128, 256, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "base_capacity", "128"),
					resource.TestCheckResourceAttr(resourceName, names.AttrMaxCapacity, "256"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPubliclyAccessible, acctest.CtTrue),
				),
			},
			{
				Config: testAccWorkgroupConfig_baseCapacity(rName, 128),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "base_capacity", "128"),
					resource.TestCheckResourceAttr(resourceName, names.AttrMaxCapacity, "0"),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkgroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkgroupConfig_configParameters(rName, "10", "20", "30", "40", "50", "100", "60", "70", "80"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkgroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "config_parameter.#", "17"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "datestyle",
						"parameter_value": "ISO, MDY",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "enable_user_activity_logging",
						"parameter_value": acctest.CtTrue,
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
						"parameter_key":   "max_query_cpu_time",
						"parameter_value": "10",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "max_query_blocks_read",
						"parameter_value": "20",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "max_scan_row_count",
						"parameter_value": "30",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "max_query_execution_time",
						"parameter_value": "40",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "max_query_queue_time",
						"parameter_value": "50",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "max_query_cpu_usage_percent",
						"parameter_value": "100",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "max_query_temp_blocks_to_disk",
						"parameter_value": "60",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "max_join_row_count",
						"parameter_value": "70",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "max_nested_loop_join_row_count",
						"parameter_value": "80",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "auto_mv",
						"parameter_value": acctest.CtTrue,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "enable_case_sensitive_identifier",
						"parameter_value": acctest.CtFalse,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "require_ssl",
						"parameter_value": acctest.CtTrue,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "use_fips_ssl",
						"parameter_value": acctest.CtFalse,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWorkgroupConfig_configParameters(rName, "100", "200", "300", "400", "500", "50", "600", "700", "800"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkgroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "config_parameter.#", "17"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "datestyle",
						"parameter_value": "ISO, MDY",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "enable_user_activity_logging",
						"parameter_value": acctest.CtTrue,
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
						"parameter_key":   "max_query_cpu_time",
						"parameter_value": "100",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "max_query_blocks_read",
						"parameter_value": "200",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "max_scan_row_count",
						"parameter_value": "300",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "max_query_execution_time",
						"parameter_value": "400",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "max_query_queue_time",
						"parameter_value": "500",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "max_query_cpu_usage_percent",
						"parameter_value": "50",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "max_query_temp_blocks_to_disk",
						"parameter_value": "600",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "max_join_row_count",
						"parameter_value": "700",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "max_nested_loop_join_row_count",
						"parameter_value": "800",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "auto_mv",
						"parameter_value": acctest.CtTrue,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "enable_case_sensitive_identifier",
						"parameter_value": acctest.CtFalse,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "require_ssl",
						"parameter_value": acctest.CtTrue,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "use_fips_ssl",
						"parameter_value": acctest.CtFalse,
					}),
				),
			},
			{
				Config: testAccWorkgroupConfig_removeLimits(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkgroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "config_parameter.#", "8"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "datestyle",
						"parameter_value": "ISO, MDY",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "enable_user_activity_logging",
						"parameter_value": acctest.CtTrue,
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
						"parameter_key":   "auto_mv",
						"parameter_value": acctest.CtTrue,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "enable_case_sensitive_identifier",
						"parameter_value": acctest.CtFalse,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "require_ssl",
						"parameter_value": acctest.CtFalse,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config_parameter.*", map[string]string{
						"parameter_key":   "use_fips_ssl",
						"parameter_value": acctest.CtFalse,
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
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkgroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkgroupConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkgroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWorkgroupConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccWorkgroupConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkgroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServerlessServiceID),
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

func TestAccRedshiftServerlessWorkgroup_port(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshiftserverless_workgroup.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkgroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkgroupConfig_port(rName, 8191),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkgroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "8191"),
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

func testAccCheckWorkgroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftServerlessClient(ctx)

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

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftServerlessClient(ctx)

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

func testAccWorkgroupConfig_baseAndMaxCapacityAndPubliclyAccessible(rName string, baseCapacity int, maxCapacity int, publiclyAccessible bool) string {
	return fmt.Sprintf(`
resource "aws_redshiftserverless_namespace" "test" {
  namespace_name = %[1]q
}

resource "aws_redshiftserverless_workgroup" "test" {
  namespace_name      = aws_redshiftserverless_namespace.test.namespace_name
  workgroup_name      = %[1]q
  base_capacity       = %[2]d
  max_capacity        = %[3]d
  publicly_accessible = %[4]t
}
`, rName, baseCapacity, maxCapacity, publiclyAccessible)
}

func testAccWorkgroupConfig_baseCapacity(rName string, baseCapacity int) string {
	return fmt.Sprintf(`
resource "aws_redshiftserverless_namespace" "test" {
  namespace_name = %[1]q
}

resource "aws_redshiftserverless_workgroup" "test" {
  namespace_name = aws_redshiftserverless_namespace.test.namespace_name
  workgroup_name = %[1]q
  base_capacity  = %[2]d
}

`, rName, baseCapacity)
}

func testAccWorkgroupConfig_configParameters(
	rName,
	maxQueryCpuTime,
	maxQueryBlocksRead,
	maxScanRowCount,
	maxQueryExecutionTime,
	maxQueryQueueTime,
	maxQueryCpuUsagePercent,
	maxQueryTempBlocksToDisk,
	maxJoinRowCount,
	maxNestedLoopJoinRowCount string,
) string {
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
    parameter_key   = "max_query_cpu_time"
    parameter_value = %[2]q
  }
  config_parameter {
    parameter_key   = "max_query_blocks_read"
    parameter_value = %[3]q
  }
  config_parameter {
    parameter_key   = "max_scan_row_count"
    parameter_value = %[4]q
  }
  config_parameter {
    parameter_key   = "max_query_execution_time"
    parameter_value = %[5]q
  }
  config_parameter {
    parameter_key   = "max_query_queue_time"
    parameter_value = %[6]q
  }
  config_parameter {
    parameter_key   = "max_query_cpu_usage_percent"
    parameter_value = %[7]q
  }
  config_parameter {
    parameter_key   = "max_query_temp_blocks_to_disk"
    parameter_value = %[8]q
  }
  config_parameter {
    parameter_key   = "max_join_row_count"
    parameter_value = %[9]q
  }
  config_parameter {
    parameter_key   = "max_nested_loop_join_row_count"
    parameter_value = %[10]q
  }
  config_parameter {
    parameter_key   = "auto_mv"
    parameter_value = "true"
  }
  config_parameter {
    parameter_key   = "enable_case_sensitive_identifier"
    parameter_value = "false"
  }
  config_parameter {
    parameter_key   = "require_ssl"
    parameter_value = "true"
  }
  config_parameter {
    parameter_key   = "use_fips_ssl"
    parameter_value = "false"
  }
}
`, rName,
		maxQueryCpuTime,
		maxQueryBlocksRead,
		maxScanRowCount,
		maxQueryExecutionTime,
		maxQueryQueueTime,
		maxQueryCpuUsagePercent,
		maxQueryTempBlocksToDisk,
		maxJoinRowCount,
		maxNestedLoopJoinRowCount,
	)
}

func testAccWorkgroupConfig_removeLimits(rName string) string {
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
    parameter_key   = "auto_mv"
    parameter_value = "true"
  }
  config_parameter {
    parameter_key   = "enable_case_sensitive_identifier"
    parameter_value = "false"
  }
  config_parameter {
    parameter_key   = "require_ssl"
    parameter_value = "false"
  }
  config_parameter {
    parameter_key   = "use_fips_ssl"
    parameter_value = "false"
  }
}
`, rName)
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

func testAccWorkgroupConfig_port(rName string, port int) string {
	return fmt.Sprintf(`
resource "aws_redshiftserverless_namespace" "test" {
  namespace_name = %[1]q
}

resource "aws_redshiftserverless_workgroup" "test" {
  namespace_name = aws_redshiftserverless_namespace.test.namespace_name
  workgroup_name = %[1]q
  port           = %[2]d
}
`, rName, port)
}
