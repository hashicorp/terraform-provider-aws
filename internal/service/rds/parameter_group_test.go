package rds_test

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/rds"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
)

func TestAccRDSParameterGroup_basic(t *testing.T) {
	var v rds.DBParameterGroup
	resourceName := "aws_db_parameter_group.test"
	groupName := fmt.Sprintf("parameter-group-test-terraform-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupConfig(groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &v),
					testAccCheckParameterGroupAttributes(&v, groupName),
					resource.TestCheckResourceAttr(resourceName, "name", groupName),
					resource.TestCheckResourceAttr(resourceName, "family", "mysql5.6"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_results",
						"value": "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_server",
						"value": "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_client",
						"value": "utf8",
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "rds", regexp.MustCompile(fmt.Sprintf("pg:%s$", groupName))),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccParameterGroupAddParametersConfig(groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &v),
					testAccCheckParameterGroupAttributes(&v, groupName),
					resource.TestCheckResourceAttr(resourceName, "name", groupName),
					resource.TestCheckResourceAttr(resourceName, "family", "mysql5.6"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "collation_connection",
						"value": "utf8_unicode_ci",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_results",
						"value": "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_server",
						"value": "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "collation_server",
						"value": "utf8_unicode_ci",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_client",
						"value": "utf8",
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "rds", regexp.MustCompile(fmt.Sprintf("pg:%s$", groupName))),
				),
			},
			{
				Config: testAccParameterGroupConfig(groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &v),
					testAccCheckParameterGroupAttributes(&v, groupName),
					testAccCheckParameterNotUserDefined(resourceName, "collation_connection"),
					testAccCheckParameterNotUserDefined(resourceName, "collation_server"),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_results",
						"value": "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_server",
						"value": "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_client",
						"value": "utf8",
					}),
				),
			},
		},
	})
}

func TestAccRDSParameterGroup_caseWithMixedParameters(t *testing.T) {
	groupName := fmt.Sprintf("parameter-group-test-terraform-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupCaseWithMixedParametersConfig(groupName),
				Check:  resource.ComposeTestCheckFunc(),
			},
		},
	})
}

func TestAccRDSParameterGroup_limit(t *testing.T) {
	var v rds.DBParameterGroup
	resourceName := "aws_db_parameter_group.test"
	groupName := fmt.Sprintf("parameter-group-test-terraform-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: createParameterGroupsExceedDefaultAWSLimit(groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &v),
					testAccCheckParameterGroupAttributes(&v, groupName),
					resource.TestCheckResourceAttr(resourceName, "name", groupName),
					resource.TestCheckResourceAttr(resourceName, "family", "mysql5.6"),
					resource.TestCheckResourceAttr(resourceName, "description", "RDS default parameter group: Exceed default AWS parameter group limit of twenty"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_server",
						"value": "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_client",
						"value": "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "collation_server",
						"value": "utf8_general_ci",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "collation_connection",
						"value": "utf8_general_ci",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "join_buffer_size",
						"value": "16777216",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "key_buffer_size",
						"value": "67108864",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "max_connections",
						"value": "3200",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "max_heap_table_size",
						"value": "67108864",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "performance_schema",
						"value": "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "performance_schema_users_size",
						"value": "1048576",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "query_cache_limit",
						"value": "2097152",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "query_cache_size",
						"value": "67108864",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "sort_buffer_size",
						"value": "16777216",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "table_open_cache",
						"value": "4096",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "tmp_table_size",
						"value": "67108864",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "binlog_cache_size",
						"value": "131072",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "innodb_flush_log_at_trx_commit",
						"value": "0",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "innodb_open_files",
						"value": "4000",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "innodb_read_io_threads",
						"value": "64",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "innodb_thread_concurrency",
						"value": "0",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "innodb_write_io_threads",
						"value": "64",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_connection",
						"value": "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_database",
						"value": "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_filesystem",
						"value": "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_results",
						"value": "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "event_scheduler",
						"value": "on",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "innodb_buffer_pool_dump_at_shutdown",
						"value": "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "innodb_file_format",
						"value": "barracuda",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "innodb_io_capacity",
						"value": "2000",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "innodb_io_capacity_max",
						"value": "3000",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "innodb_lock_wait_timeout",
						"value": "120",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "innodb_max_dirty_pages_pct",
						"value": "90",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "log_bin_trust_function_creators",
						"value": "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "log_warnings",
						"value": "2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "log_output",
						"value": "FILE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "max_allowed_packet",
						"value": "1073741824",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "max_connect_errors",
						"value": "100",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "query_cache_min_res_unit",
						"value": "512",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "slow_query_log",
						"value": "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "sync_binlog",
						"value": "0",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "tx_isolation",
						"value": "repeatable-read",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: updateParameterGroupsExceedDefaultAWSLimit(groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &v),
					testAccCheckParameterGroupAttributes(&v, groupName),
					resource.TestCheckResourceAttr(resourceName, "name", groupName),
					resource.TestCheckResourceAttr(resourceName, "family", "mysql5.6"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated RDS default parameter group: Exceed default AWS parameter group limit of twenty"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_server",
						"value": "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_client",
						"value": "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "collation_server",
						"value": "utf8_general_ci",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "collation_connection",
						"value": "utf8_general_ci",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "join_buffer_size",
						"value": "16777216",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "key_buffer_size",
						"value": "67108864",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "max_connections",
						"value": "3200",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "max_heap_table_size",
						"value": "67108864",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "performance_schema",
						"value": "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "performance_schema_users_size",
						"value": "1048576",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "query_cache_limit",
						"value": "2097152",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "query_cache_size",
						"value": "67108864",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "sort_buffer_size",
						"value": "16777216",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "table_open_cache",
						"value": "4096",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "tmp_table_size",
						"value": "67108864",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "binlog_cache_size",
						"value": "131072",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "innodb_flush_log_at_trx_commit",
						"value": "0",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "innodb_open_files",
						"value": "4000",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "innodb_read_io_threads",
						"value": "64",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "innodb_thread_concurrency",
						"value": "0",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "innodb_write_io_threads",
						"value": "64",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_connection",
						"value": "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_database",
						"value": "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_filesystem",
						"value": "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_results",
						"value": "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "event_scheduler",
						"value": "on",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "innodb_buffer_pool_dump_at_shutdown",
						"value": "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "innodb_file_format",
						"value": "barracuda",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "innodb_io_capacity",
						"value": "2000",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "innodb_io_capacity_max",
						"value": "3000",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "innodb_lock_wait_timeout",
						"value": "120",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "innodb_max_dirty_pages_pct",
						"value": "90",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "log_bin_trust_function_creators",
						"value": "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "log_warnings",
						"value": "2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "log_output",
						"value": "FILE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "max_allowed_packet",
						"value": "1073741824",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "max_connect_errors",
						"value": "100",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "query_cache_min_res_unit",
						"value": "512",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "slow_query_log",
						"value": "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "sync_binlog",
						"value": "0",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "tx_isolation",
						"value": "repeatable-read",
					}),
				),
			},
		},
	})
}

func TestAccRDSParameterGroup_disappears(t *testing.T) {
	var v rds.DBParameterGroup
	resourceName := "aws_db_parameter_group.test"
	groupName := fmt.Sprintf("parameter-group-test-terraform-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupConfig(groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &v),
					testAccCheckParamaterGroupDisappears(&v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRDSParameterGroup_namePrefix(t *testing.T) {
	var v rds.DBParameterGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDBParameterGroupConfig_namePrefix,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists("aws_db_parameter_group.test", &v),
					resource.TestMatchResourceAttr("aws_db_parameter_group.test", "name", regexp.MustCompile("^tf-test-")),
				),
			},
		},
	})
}

func TestAccRDSParameterGroup_generatedName(t *testing.T) {
	var v rds.DBParameterGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDBParameterGroupConfig_generatedName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists("aws_db_parameter_group.test", &v),
				),
			},
		},
	})
}

func TestAccRDSParameterGroup_withApplyMethod(t *testing.T) {
	var v rds.DBParameterGroup
	resourceName := "aws_db_parameter_group.test"
	groupName := fmt.Sprintf("parameter-group-test-terraform-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupWithApplyMethodConfig(groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &v),
					testAccCheckParameterGroupAttributes(&v, groupName),
					resource.TestCheckResourceAttr(resourceName, "name", groupName),
					resource.TestCheckResourceAttr(resourceName, "family", "mysql5.6"),
					resource.TestCheckResourceAttr(resourceName, "description", "Managed by Terraform"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":         "character_set_server",
						"value":        "utf8",
						"apply_method": "immediate",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":         "character_set_client",
						"value":        "utf8",
						"apply_method": "pending-reboot",
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

func TestAccRDSParameterGroup_only(t *testing.T) {
	var v rds.DBParameterGroup
	resourceName := "aws_db_parameter_group.test"
	groupName := fmt.Sprintf("parameter-group-test-terraform-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupOnlyConfig(groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &v),
					testAccCheckParameterGroupAttributes(&v, groupName),
					resource.TestCheckResourceAttr(resourceName, "name", groupName),
					resource.TestCheckResourceAttr(resourceName, "family", "mysql5.6"),
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

func TestAccRDSParameterGroup_matchDefault(t *testing.T) {
	var v rds.DBParameterGroup
	resourceName := "aws_db_parameter_group.test"
	groupName := fmt.Sprintf("parameter-group-test-terraform-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupIncludeDefaultConfig(groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", groupName),
					resource.TestCheckResourceAttr(resourceName, "family", "postgres9.4"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"parameter"},
			},
		},
	})
}

func TestAccRDSParameterGroup_updateParameters(t *testing.T) {
	var v rds.DBParameterGroup
	resourceName := "aws_db_parameter_group.test"
	groupName := fmt.Sprintf("parameter-group-test-terraform-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupUpdateParametersInitialConfig(groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &v),
					testAccCheckParameterGroupAttributes(&v, groupName),
					resource.TestCheckResourceAttr(resourceName, "name", groupName),
					resource.TestCheckResourceAttr(resourceName, "family", "mysql5.6"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_results",
						"value": "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_server",
						"value": "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_client",
						"value": "utf8",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccParameterGroupUpdateParametersUpdatedConfig(groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &v),
					testAccCheckParameterGroupAttributes(&v, groupName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_results",
						"value": "ascii",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_server",
						"value": "ascii",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_client",
						"value": "utf8",
					}),
				),
			},
		},
	})
}

func TestAccRDSParameterGroup_caseParameters(t *testing.T) {
	var v rds.DBParameterGroup
	resourceName := "aws_db_parameter_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupUpperCaseConfig(rName, "Max_connections"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &v),
					testAccCheckParameterGroupAttributes(&v, rName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "family", "mysql5.6"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "Max_connections",
						"value": "LEAST({DBInstanceClassMemory/6000000},10)",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccParameterGroupUpperCaseConfig(rName, "max_connections"),
			},
		},
	})
}

func TestDBParameterModifyChunk(t *testing.T) {
	cases := []struct {
		Name              string
		ChunkSize         int
		Parameters        []*rds.Parameter
		ExpectedModify    []*rds.Parameter
		ExpectedRemainder []*rds.Parameter
	}{
		{
			Name:              "Empty",
			ChunkSize:         20,
			Parameters:        nil,
			ExpectedModify:    nil,
			ExpectedRemainder: nil,
		},
		{
			Name:      "A couple",
			ChunkSize: 20,
			Parameters: []*rds.Parameter{
				{
					ApplyMethod:    aws.String("immediate"),
					ParameterName:  aws.String("tx_isolation"),
					ParameterValue: aws.String("repeatable-read"),
				},
				{
					ApplyMethod:    aws.String("immediate"),
					ParameterName:  aws.String("binlog_cache_size"),
					ParameterValue: aws.String("131072"),
				},
			},
			ExpectedModify: []*rds.Parameter{
				{
					ApplyMethod:    aws.String("immediate"),
					ParameterName:  aws.String("tx_isolation"),
					ParameterValue: aws.String("repeatable-read"),
				},
				{
					ApplyMethod:    aws.String("immediate"),
					ParameterName:  aws.String("binlog_cache_size"),
					ParameterValue: aws.String("131072"),
				},
			},
			ExpectedRemainder: nil,
		},
		{
			Name:      "Over 3 max, 6 in",
			ChunkSize: 3,
			Parameters: []*rds.Parameter{
				{
					ApplyMethod:    aws.String("immediate"),
					ParameterName:  aws.String("tx_isolation"),
					ParameterValue: aws.String("repeatable-read"),
				},
				{
					ApplyMethod:    aws.String("immediate"),
					ParameterName:  aws.String("binlog_cache_size"),
					ParameterValue: aws.String("131072"),
				},
				{
					ApplyMethod:    aws.String("pending-reboot"),
					ParameterName:  aws.String("innodb_read_io_threads"),
					ParameterValue: aws.String("64"),
				},
				{
					ApplyMethod:    aws.String("immediate"),
					ParameterName:  aws.String("character_set_server"),
					ParameterValue: aws.String("utf8"),
				},
				{
					ApplyMethod:    aws.String("immediate"),
					ParameterName:  aws.String("innodb_flush_log_at_trx_commit"),
					ParameterValue: aws.String("0"),
				},
				{
					ApplyMethod:    aws.String("immediate"),
					ParameterName:  aws.String("character_set_filesystem"),
					ParameterValue: aws.String("utf8"),
				},
			},
			ExpectedModify: []*rds.Parameter{
				{
					ApplyMethod:    aws.String("immediate"),
					ParameterName:  aws.String("character_set_server"),
					ParameterValue: aws.String("utf8"),
				},
				{
					ApplyMethod:    aws.String("immediate"),
					ParameterName:  aws.String("character_set_filesystem"),
					ParameterValue: aws.String("utf8"),
				},
				{
					ApplyMethod:    aws.String("immediate"),
					ParameterName:  aws.String("tx_isolation"),
					ParameterValue: aws.String("repeatable-read"),
				},
			},
			ExpectedRemainder: []*rds.Parameter{
				{
					ApplyMethod:    aws.String("immediate"),
					ParameterName:  aws.String("binlog_cache_size"),
					ParameterValue: aws.String("131072"),
				},
				{
					ApplyMethod:    aws.String("pending-reboot"),
					ParameterName:  aws.String("innodb_read_io_threads"),
					ParameterValue: aws.String("64"),
				},
				{
					ApplyMethod:    aws.String("immediate"),
					ParameterName:  aws.String("innodb_flush_log_at_trx_commit"),
					ParameterValue: aws.String("0"),
				},
			},
		},
		{
			Name:      "Over 3 max, 9 in",
			ChunkSize: 3,
			Parameters: []*rds.Parameter{
				{
					ApplyMethod:    aws.String("pending-reboot"),
					ParameterName:  aws.String("tx_isolation"),
					ParameterValue: aws.String("repeatable-read"),
				},
				{
					ApplyMethod:    aws.String("pending-reboot"),
					ParameterName:  aws.String("binlog_cache_size"),
					ParameterValue: aws.String("131072"),
				},
				{
					ApplyMethod:    aws.String("pending-reboot"),
					ParameterName:  aws.String("innodb_read_io_threads"),
					ParameterValue: aws.String("64"),
				},
				{
					ApplyMethod:    aws.String("pending-reboot"),
					ParameterName:  aws.String("character_set_server"),
					ParameterValue: aws.String("utf8"),
				},
				{
					ApplyMethod:    aws.String("immediate"),
					ParameterName:  aws.String("innodb_flush_log_at_trx_commit"),
					ParameterValue: aws.String("0"),
				},
				{
					ApplyMethod:    aws.String("immediate"),
					ParameterName:  aws.String("character_set_filesystem"),
					ParameterValue: aws.String("utf8"),
				},
				{
					ApplyMethod:    aws.String("pending-reboot"),
					ParameterName:  aws.String("innodb_max_dirty_pages_pct"),
					ParameterValue: aws.String("90"),
				},
				{
					ApplyMethod:    aws.String("immediate"),
					ParameterName:  aws.String("character_set_connection"),
					ParameterValue: aws.String("utf8"),
				},
				{
					ApplyMethod:    aws.String("immediate"),
					ParameterName:  aws.String("key_buffer_size"),
					ParameterValue: aws.String("67108864"),
				},
			},
			ExpectedModify: []*rds.Parameter{
				{
					ApplyMethod:    aws.String("immediate"),
					ParameterName:  aws.String("character_set_filesystem"),
					ParameterValue: aws.String("utf8"),
				},
				{
					ApplyMethod:    aws.String("immediate"),
					ParameterName:  aws.String("character_set_connection"),
					ParameterValue: aws.String("utf8"),
				},
				{
					ApplyMethod:    aws.String("immediate"),
					ParameterName:  aws.String("innodb_flush_log_at_trx_commit"),
					ParameterValue: aws.String("0"),
				},
			},
			ExpectedRemainder: []*rds.Parameter{
				{
					ApplyMethod:    aws.String("pending-reboot"),
					ParameterName:  aws.String("tx_isolation"),
					ParameterValue: aws.String("repeatable-read"),
				},
				{
					ApplyMethod:    aws.String("pending-reboot"),
					ParameterName:  aws.String("binlog_cache_size"),
					ParameterValue: aws.String("131072"),
				},
				{
					ApplyMethod:    aws.String("pending-reboot"),
					ParameterName:  aws.String("innodb_read_io_threads"),
					ParameterValue: aws.String("64"),
				},
				{
					ApplyMethod:    aws.String("pending-reboot"),
					ParameterName:  aws.String("character_set_server"),
					ParameterValue: aws.String("utf8"),
				},
				{
					ApplyMethod:    aws.String("pending-reboot"),
					ParameterName:  aws.String("innodb_max_dirty_pages_pct"),
					ParameterValue: aws.String("90"),
				},
				{
					ApplyMethod:    aws.String("immediate"),
					ParameterName:  aws.String("key_buffer_size"),
					ParameterValue: aws.String("67108864"),
				},
			},
		},
	}

	for _, tc := range cases {
		mod, rem := tfrds.ResourceParameterModifyChunk(tc.Parameters, tc.ChunkSize)
		if !reflect.DeepEqual(mod, tc.ExpectedModify) {
			t.Errorf("Case %q: Modify did not match\n%#v\n\nGot:\n%#v", tc.Name, tc.ExpectedModify, mod)
		}
		if !reflect.DeepEqual(rem, tc.ExpectedRemainder) {
			t.Errorf("Case %q: Remainder did not match\n%#v\n\nGot:\n%#v", tc.Name, tc.ExpectedRemainder, rem)
		}
	}
}

func testAccCheckParamaterGroupDisappears(v *rds.DBParameterGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_db_parameter_group" {
				continue
			}
			_, err := conn.DeleteDBParameterGroup(&rds.DeleteDBParameterGroupInput{
				DBParameterGroupName: v.DBParameterGroupName,
			})
			return err
		}
		return nil
	}
}

func testAccCheckParameterGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_db_parameter_group" {
			continue
		}

		// Try to find the Group
		resp, err := conn.DescribeDBParameterGroups(
			&rds.DescribeDBParameterGroupsInput{
				DBParameterGroupName: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if len(resp.DBParameterGroups) != 0 &&
				*resp.DBParameterGroups[0].DBParameterGroupName == rs.Primary.ID {
				return fmt.Errorf("DB Parameter Group still exists")
			}
		}

		// Verify the error
		newerr, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if newerr.Code() != "DBParameterGroupNotFound" {
			return err
		}
	}

	return nil
}

func testAccCheckParameterGroupAttributes(v *rds.DBParameterGroup, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if *v.DBParameterGroupName != name {
			return fmt.Errorf("Bad Parameter Group name, expected (%s), got (%s)", name, *v.DBParameterGroupName)
		}

		if *v.DBParameterGroupFamily != "mysql5.6" {
			return fmt.Errorf("bad family: %#v", v.DBParameterGroupFamily)
		}

		return nil
	}
}

func testAccCheckParameterGroupExists(rName string, v *rds.DBParameterGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rName]
		if !ok {
			return fmt.Errorf("Not found: %s", rName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DB Parameter Group ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

		opts := rds.DescribeDBParameterGroupsInput{
			DBParameterGroupName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeDBParameterGroups(&opts)

		if err != nil {
			return err
		}

		if len(resp.DBParameterGroups) != 1 ||
			*resp.DBParameterGroups[0].DBParameterGroupName != rs.Primary.ID {
			return fmt.Errorf("DB Parameter Group not found")
		}

		*v = *resp.DBParameterGroups[0]

		return nil
	}
}

func testAccCheckParameterNotUserDefined(rName, paramName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rName]
		if !ok {
			return fmt.Errorf("Not found: %s", rName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DB Parameter Group ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

		opts := rds.DescribeDBParametersInput{
			DBParameterGroupName: aws.String(rs.Primary.ID),
			Source:               aws.String("user"),
		}

		userDefined := false
		err := conn.DescribeDBParametersPages(&opts, func(page *rds.DescribeDBParametersOutput, lastPage bool) bool {
			for _, param := range page.Parameters {
				if *param.ParameterName == paramName {
					userDefined = true
					return false
				}
			}
			return true
		})

		if userDefined {
			return fmt.Errorf("DB Parameter is user defined")
		}

		return err
	}
}

func testAccParameterGroupConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_db_parameter_group" "test" {
  name   = %[1]q
  family = "mysql5.6"

  parameter {
    name  = "character_set_server"
    value = "utf8"
  }

  parameter {
    name  = "character_set_client"
    value = "utf8"
  }

  parameter {
    name  = "character_set_results"
    value = "utf8"
  }
}
`, rName)
}

func testAccParameterGroupCaseWithMixedParametersConfig(rName string) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMySQL(), fmt.Sprintf(`
resource "aws_db_parameter_group" "test" {
  name   = %[1]q
  family = "mysql5.6"

  parameter {
    name         = "character_set_server"
    value        = "utf8mb4"
    apply_method = "pending-reboot"
  }
  parameter {
    name         = "innodb_large_prefix"
    value        = 1
    apply_method = "pending-reboot"
  }
  parameter {
    name         = "innodb_file_format"
    value        = "Barracuda"
    apply_method = "pending-reboot"
  }
  parameter {
    name         = "innodb_log_file_size"
    value        = 2147483648
    apply_method = "pending-reboot"
  }
  parameter {
    name         = "sql_mode"
    value        = "STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_AUTO_CREATE_USER,NO_ENGINE_SUBSTITUTION"
    apply_method = "pending-reboot"
  }
  parameter {
    name         = "innodb_log_buffer_size"
    value        = 268435456
    apply_method = "pending-reboot"
  }
  parameter {
    name         = "max_allowed_packet"
    value        = 268435456
    apply_method = "pending-reboot"
  }
}
`, rName))
}

func testAccParameterGroupWithApplyMethodConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_db_parameter_group" "test" {
  name   = %[1]q
  family = "mysql5.6"

  parameter {
    name  = "character_set_server"
    value = "utf8"
  }

  parameter {
    name         = "character_set_client"
    value        = "utf8"
    apply_method = "pending-reboot"
  }

  tags = {
    foo = "test"
  }
}
`, rName)
}

func testAccParameterGroupAddParametersConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_db_parameter_group" "test" {
  name   = %[1]q
  family = "mysql5.6"

  parameter {
    name  = "character_set_server"
    value = "utf8"
  }

  parameter {
    name  = "character_set_client"
    value = "utf8"
  }

  parameter {
    name  = "character_set_results"
    value = "utf8"
  }

  parameter {
    name  = "collation_server"
    value = "utf8_unicode_ci"
  }

  parameter {
    name  = "collation_connection"
    value = "utf8_unicode_ci"
  }
}
`, rName)
}

func testAccParameterGroupOnlyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_db_parameter_group" "test" {
  name        = %[1]q
  family      = "mysql5.6"
  description = "Test parameter group for terraform"
}
`, rName)
}

func createParameterGroupsExceedDefaultAWSLimit(rName string) string {
	return fmt.Sprintf(`
resource "aws_db_parameter_group" "test" {
  name        = %[1]q
  family      = "mysql5.6"
  description = "RDS default parameter group: Exceed default AWS parameter group limit of twenty"

  parameter {
    name  = "binlog_cache_size"
    value = 131072
  }

  parameter {
    name  = "character_set_client"
    value = "utf8"
  }

  parameter {
    name  = "character_set_connection"
    value = "utf8"
  }

  parameter {
    name  = "character_set_database"
    value = "utf8"
  }

  parameter {
    name  = "character_set_filesystem"
    value = "utf8"
  }

  parameter {
    name  = "character_set_results"
    value = "utf8"
  }

  parameter {
    name  = "character_set_server"
    value = "utf8"
  }

  parameter {
    name  = "collation_connection"
    value = "utf8_general_ci"
  }

  parameter {
    name  = "collation_server"
    value = "utf8_general_ci"
  }

  parameter {
    name  = "event_scheduler"
    value = "on"
  }

  parameter {
    name  = "innodb_buffer_pool_dump_at_shutdown"
    value = 1
  }

  parameter {
    name  = "innodb_file_format"
    value = "barracuda"
  }

  parameter {
    name  = "innodb_flush_log_at_trx_commit"
    value = 0
  }

  parameter {
    name  = "innodb_io_capacity"
    value = 2000
  }

  parameter {
    name  = "innodb_io_capacity_max"
    value = 3000
  }

  parameter {
    name  = "innodb_lock_wait_timeout"
    value = 120
  }

  parameter {
    name  = "innodb_max_dirty_pages_pct"
    value = 90
  }

  parameter {
    name         = "innodb_open_files"
    value        = 4000
    apply_method = "pending-reboot"
  }

  parameter {
    name         = "innodb_read_io_threads"
    value        = 64
    apply_method = "pending-reboot"
  }

  parameter {
    name  = "innodb_thread_concurrency"
    value = 0
  }

  parameter {
    name         = "innodb_write_io_threads"
    value        = 64
    apply_method = "pending-reboot"
  }

  parameter {
    name  = "join_buffer_size"
    value = 16777216
  }

  parameter {
    name  = "key_buffer_size"
    value = 67108864
  }

  parameter {
    name  = "log_bin_trust_function_creators"
    value = 1
  }

  parameter {
    name  = "log_warnings"
    value = 2
  }

  parameter {
    name  = "log_output"
    value = "FILE"
  }

  parameter {
    name  = "max_allowed_packet"
    value = 1073741824
  }

  parameter {
    name  = "max_connect_errors"
    value = 100
  }

  parameter {
    name  = "max_connections"
    value = 3200
  }

  parameter {
    name  = "max_heap_table_size"
    value = 67108864
  }

  parameter {
    name         = "performance_schema"
    value        = 1
    apply_method = "pending-reboot"
  }

  parameter {
    name         = "performance_schema_users_size"
    value        = 1048576
    apply_method = "pending-reboot"
  }

  parameter {
    name  = "query_cache_limit"
    value = 2097152
  }

  parameter {
    name  = "query_cache_min_res_unit"
    value = 512
  }

  parameter {
    name  = "query_cache_size"
    value = 67108864
  }

  parameter {
    name  = "slow_query_log"
    value = 1
  }

  parameter {
    name  = "sort_buffer_size"
    value = 16777216
  }

  parameter {
    name  = "sync_binlog"
    value = 0
  }

  parameter {
    name  = "table_open_cache"
    value = 4096
  }

  parameter {
    name  = "tmp_table_size"
    value = 67108864
  }

  parameter {
    name  = "tx_isolation"
    value = "repeatable-read"
  }
}
`, rName)
}

func updateParameterGroupsExceedDefaultAWSLimit(rName string) string {
	return fmt.Sprintf(`
resource "aws_db_parameter_group" "test" {
  name        = %[1]q
  family      = "mysql5.6"
  description = "Updated RDS default parameter group: Exceed default AWS parameter group limit of twenty"

  parameter {
    name  = "binlog_cache_size"
    value = 131072
  }

  parameter {
    name  = "character_set_client"
    value = "utf8"
  }

  parameter {
    name  = "character_set_connection"
    value = "utf8"
  }

  parameter {
    name  = "character_set_database"
    value = "utf8"
  }

  parameter {
    name  = "character_set_filesystem"
    value = "utf8"
  }

  parameter {
    name  = "character_set_results"
    value = "utf8"
  }

  parameter {
    name  = "character_set_server"
    value = "utf8"
  }

  parameter {
    name  = "collation_connection"
    value = "utf8_general_ci"
  }

  parameter {
    name  = "collation_server"
    value = "utf8_general_ci"
  }

  parameter {
    name  = "event_scheduler"
    value = "on"
  }

  parameter {
    name  = "innodb_buffer_pool_dump_at_shutdown"
    value = 1
  }

  parameter {
    name  = "innodb_file_format"
    value = "barracuda"
  }

  parameter {
    name  = "innodb_flush_log_at_trx_commit"
    value = 0
  }

  parameter {
    name  = "innodb_io_capacity"
    value = 2000
  }

  parameter {
    name  = "innodb_io_capacity_max"
    value = 3000
  }

  parameter {
    name  = "innodb_lock_wait_timeout"
    value = 120
  }

  parameter {
    name  = "innodb_max_dirty_pages_pct"
    value = 90
  }

  parameter {
    name         = "innodb_open_files"
    value        = 4000
    apply_method = "pending-reboot"
  }

  parameter {
    name         = "innodb_read_io_threads"
    value        = 64
    apply_method = "pending-reboot"
  }

  parameter {
    name  = "innodb_thread_concurrency"
    value = 0
  }

  parameter {
    name         = "innodb_write_io_threads"
    value        = 64
    apply_method = "pending-reboot"
  }

  parameter {
    name  = "join_buffer_size"
    value = 16777216
  }

  parameter {
    name  = "key_buffer_size"
    value = 67108864
  }

  parameter {
    name  = "log_bin_trust_function_creators"
    value = 1
  }

  parameter {
    name  = "log_warnings"
    value = 2
  }

  parameter {
    name  = "log_output"
    value = "FILE"
  }

  parameter {
    name  = "max_allowed_packet"
    value = 1073741824
  }

  parameter {
    name  = "max_connect_errors"
    value = 100
  }

  parameter {
    name  = "max_connections"
    value = 3200
  }

  parameter {
    name  = "max_heap_table_size"
    value = 67108864
  }

  parameter {
    name         = "performance_schema"
    value        = 1
    apply_method = "pending-reboot"
  }

  parameter {
    name         = "performance_schema_users_size"
    value        = 1048576
    apply_method = "pending-reboot"
  }

  parameter {
    name  = "query_cache_limit"
    value = 2097152
  }

  parameter {
    name  = "query_cache_min_res_unit"
    value = 512
  }

  parameter {
    name  = "query_cache_size"
    value = 67108864
  }

  parameter {
    name  = "slow_query_log"
    value = 1
  }

  parameter {
    name  = "sort_buffer_size"
    value = 16777216
  }

  parameter {
    name  = "sync_binlog"
    value = 0
  }

  parameter {
    name  = "table_open_cache"
    value = 4096
  }

  parameter {
    name  = "tmp_table_size"
    value = 67108864
  }

  parameter {
    name  = "tx_isolation"
    value = "repeatable-read"
  }
}
`, rName)
}

func testAccParameterGroupIncludeDefaultConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_db_parameter_group" "test" {
  name   = %[1]q
  family = "postgres9.4"

  parameter {
    name         = "client_encoding"
    value        = "UTF8"
    apply_method = "pending-reboot"
  }
}
`, rName)
}

func testAccParameterGroupUpdateParametersInitialConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_db_parameter_group" "test" {
  name   = %[1]q
  family = "mysql5.6"

  parameter {
    name  = "character_set_server"
    value = "utf8"
  }

  parameter {
    name  = "character_set_client"
    value = "utf8"
  }

  parameter {
    name  = "character_set_results"
    value = "utf8"
  }
}
`, rName)
}

func testAccParameterGroupUpdateParametersUpdatedConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_db_parameter_group" "test" {
  name   = %[1]q
  family = "mysql5.6"

  parameter {
    name  = "character_set_server"
    value = "ascii"
  }

  parameter {
    name  = "character_set_client"
    value = "utf8"
  }

  parameter {
    name  = "character_set_results"
    value = "ascii"
  }
}
`, rName)
}

func testAccParameterGroupUpperCaseConfig(rName, paramName string) string {
	return fmt.Sprintf(`
resource "aws_db_parameter_group" "test" {
  name   = %[1]q
  family = "mysql5.6"

  parameter {
    name  = %[2]q
    value = "LEAST({DBInstanceClassMemory/6000000},10)"
  }
}
`, rName, paramName)
}

const testAccDBParameterGroupConfig_namePrefix = `
resource "aws_db_parameter_group" "test" {
  name_prefix = "tf-test-"
  family      = "mysql5.6"

  parameter {
    name  = "sync_binlog"
    value = 0
  }
}
`

const testAccDBParameterGroupConfig_generatedName = `
resource "aws_db_parameter_group" "test" {
  family = "mysql5.6"

  parameter {
    name  = "sync_binlog"
    value = 0
  }
}
`
