package aws

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/rds"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func init() {
	resource.AddTestSweepers("aws_db_parameter_group", &resource.Sweeper{
		Name: "aws_db_parameter_group",
		F:    testSweepRdsDbParameterGroups,
		Dependencies: []string{
			"aws_db_instance",
		},
	})
}

func testSweepRdsDbParameterGroups(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).rdsconn

	err = conn.DescribeDBParameterGroupsPages(&rds.DescribeDBParameterGroupsInput{}, func(out *rds.DescribeDBParameterGroupsOutput, lastPage bool) bool {
		for _, dbpg := range out.DBParameterGroups {
			if dbpg == nil {
				continue
			}

			input := &rds.DeleteDBParameterGroupInput{
				DBParameterGroupName: dbpg.DBParameterGroupName,
			}
			name := aws.StringValue(dbpg.DBParameterGroupName)

			if strings.HasPrefix(name, "default.") {
				log.Printf("[INFO] Skipping DB Parameter Group: %s", name)
				continue
			}

			log.Printf("[INFO] Deleting DB Parameter Group: %s", name)

			_, err := conn.DeleteDBParameterGroup(input)

			if err != nil {
				log.Printf("[ERROR] Failed to delete DB Parameter Group %s: %s", name, err)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping RDS DB Parameter Group sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error retrieving DB Parameter Groups: %s", err)
	}

	return nil
}

func TestAccAWSDBParameterGroup_basic(t *testing.T) {
	var v rds.DBParameterGroup
	resourceName := "aws_db_parameter_group.test"
	groupName := fmt.Sprintf("parameter-group-test-terraform-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBParameterGroupConfig(groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBParameterGroupExists(resourceName, &v),
					testAccCheckAWSDBParameterGroupAttributes(&v, groupName),
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
				Config: testAccAWSDBParameterGroupAddParametersConfig(groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBParameterGroupExists(resourceName, &v),
					testAccCheckAWSDBParameterGroupAttributes(&v, groupName),
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
				Config: testAccAWSDBParameterGroupConfig(groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBParameterGroupExists(resourceName, &v),
					testAccCheckAWSDBParameterGroupAttributes(&v, groupName),
					testAccCheckAWSDBParameterNotUserDefined(resourceName, "collation_connection"),
					testAccCheckAWSDBParameterNotUserDefined(resourceName, "collation_server"),
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

func TestAccAWSDBParameterGroup_caseWithMixedParameters(t *testing.T) {
	groupName := fmt.Sprintf("parameter-group-test-terraform-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBParameterGroupConfigCaseWithMixedParameters(groupName),
				Check:  resource.ComposeTestCheckFunc(),
			},
		},
	})
}

func TestAccAWSDBParameterGroup_limit(t *testing.T) {
	var v rds.DBParameterGroup
	resourceName := "aws_db_parameter_group.test"
	groupName := fmt.Sprintf("parameter-group-test-terraform-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: createAwsDbParameterGroupsExceedDefaultAwsLimit(groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBParameterGroupExists(resourceName, &v),
					testAccCheckAWSDBParameterGroupAttributes(&v, groupName),
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
				Config: updateAwsDbParameterGroupsExceedDefaultAwsLimit(groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBParameterGroupExists(resourceName, &v),
					testAccCheckAWSDBParameterGroupAttributes(&v, groupName),
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

func TestAccAWSDBParameterGroup_Disappears(t *testing.T) {
	var v rds.DBParameterGroup
	resourceName := "aws_db_parameter_group.test"
	groupName := fmt.Sprintf("parameter-group-test-terraform-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBParameterGroupConfig(groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBParameterGroupExists(resourceName, &v),
					testAccCheckAWSDbParamaterGroupDisappears(&v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSDBParameterGroup_namePrefix(t *testing.T) {
	var v rds.DBParameterGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDBParameterGroupConfig_namePrefix,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBParameterGroupExists("aws_db_parameter_group.test", &v),
					resource.TestMatchResourceAttr("aws_db_parameter_group.test", "name", regexp.MustCompile("^tf-test-")),
				),
			},
		},
	})
}

func TestAccAWSDBParameterGroup_generatedName(t *testing.T) {
	var v rds.DBParameterGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDBParameterGroupConfig_generatedName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBParameterGroupExists("aws_db_parameter_group.test", &v),
				),
			},
		},
	})
}

func TestAccAWSDBParameterGroup_withApplyMethod(t *testing.T) {
	var v rds.DBParameterGroup
	resourceName := "aws_db_parameter_group.test"
	groupName := fmt.Sprintf("parameter-group-test-terraform-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBParameterGroupConfigWithApplyMethod(groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBParameterGroupExists(resourceName, &v),
					testAccCheckAWSDBParameterGroupAttributes(&v, groupName),
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

func TestAccAWSDBParameterGroup_Only(t *testing.T) {
	var v rds.DBParameterGroup
	resourceName := "aws_db_parameter_group.test"
	groupName := fmt.Sprintf("parameter-group-test-terraform-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBParameterGroupOnlyConfig(groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBParameterGroupExists(resourceName, &v),
					testAccCheckAWSDBParameterGroupAttributes(&v, groupName),
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

func TestAccAWSDBParameterGroup_MatchDefault(t *testing.T) {
	var v rds.DBParameterGroup
	resourceName := "aws_db_parameter_group.test"
	groupName := fmt.Sprintf("parameter-group-test-terraform-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBParameterGroupIncludeDefaultConfig(groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBParameterGroupExists(resourceName, &v),
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

func TestAccAWSDBParameterGroup_updateParameters(t *testing.T) {
	var v rds.DBParameterGroup
	resourceName := "aws_db_parameter_group.test"
	groupName := fmt.Sprintf("parameter-group-test-terraform-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBParameterGroupUpdateParametersInitialConfig(groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBParameterGroupExists(resourceName, &v),
					testAccCheckAWSDBParameterGroupAttributes(&v, groupName),
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
				Config: testAccAWSDBParameterGroupUpdateParametersUpdatedConfig(groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBParameterGroupExists(resourceName, &v),
					testAccCheckAWSDBParameterGroupAttributes(&v, groupName),
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

func TestAccAWSDBParameterGroup_caseParameters(t *testing.T) {
	var v rds.DBParameterGroup
	resourceName := "aws_db_parameter_group.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBParameterGroupUpperCaseConfig(rName, "Max_connections"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBParameterGroupExists(resourceName, &v),
					testAccCheckAWSDBParameterGroupAttributes(&v, rName),
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
				Config: testAccAWSDBParameterGroupUpperCaseConfig(rName, "max_connections"),
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
		mod, rem := resourceDBParameterModifyChunk(tc.Parameters, tc.ChunkSize)
		if !reflect.DeepEqual(mod, tc.ExpectedModify) {
			t.Errorf("Case %q: Modify did not match\n%#v\n\nGot:\n%#v", tc.Name, tc.ExpectedModify, mod)
		}
		if !reflect.DeepEqual(rem, tc.ExpectedRemainder) {
			t.Errorf("Case %q: Remainder did not match\n%#v\n\nGot:\n%#v", tc.Name, tc.ExpectedRemainder, rem)
		}
	}
}

func testAccCheckAWSDbParamaterGroupDisappears(v *rds.DBParameterGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).rdsconn

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

func testAccCheckAWSDBParameterGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).rdsconn

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

func testAccCheckAWSDBParameterGroupAttributes(v *rds.DBParameterGroup, name string) resource.TestCheckFunc {
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

func testAccCheckAWSDBParameterGroupExists(rName string, v *rds.DBParameterGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rName]
		if !ok {
			return fmt.Errorf("Not found: %s", rName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DB Parameter Group ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).rdsconn

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

func testAccCheckAWSDBParameterNotUserDefined(rName, paramName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rName]
		if !ok {
			return fmt.Errorf("Not found: %s", rName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DB Parameter Group ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).rdsconn

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

func testAccAWSDBParameterGroupConfig(rName string) string {
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

func testAccAWSDBParameterGroupConfigCaseWithMixedParameters(rName string) string {
	return acctest.ConfigCompose(testAccAWSDBInstanceConfig_orderableClassMysql(), fmt.Sprintf(`
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

func testAccAWSDBParameterGroupConfigWithApplyMethod(rName string) string {
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

func testAccAWSDBParameterGroupAddParametersConfig(rName string) string {
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

func testAccAWSDBParameterGroupOnlyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_db_parameter_group" "test" {
  name        = %[1]q
  family      = "mysql5.6"
  description = "Test parameter group for terraform"
}
`, rName)
}

func createAwsDbParameterGroupsExceedDefaultAwsLimit(rName string) string {
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

func updateAwsDbParameterGroupsExceedDefaultAwsLimit(rName string) string {
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

func testAccAWSDBParameterGroupIncludeDefaultConfig(rName string) string {
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

func testAccAWSDBParameterGroupUpdateParametersInitialConfig(rName string) string {
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

func testAccAWSDBParameterGroupUpdateParametersUpdatedConfig(rName string) string {
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

func testAccAWSDBParameterGroupUpperCaseConfig(rName, paramName string) string {
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
