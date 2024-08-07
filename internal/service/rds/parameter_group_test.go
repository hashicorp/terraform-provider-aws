// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSParameterGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.DBParameterGroup
	resourceName := "aws_db_parameter_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, resourceName, &v),
					testAccCheckParameterGroupAttributes(&v, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrFamily, "mysql5.6"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_results",
						names.AttrValue: "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_server",
						names.AttrValue: "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_client",
						names.AttrValue: "utf8",
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "rds", regexache.MustCompile(fmt.Sprintf("pg:%s$", rName))),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccParameterGroupConfig_addParameters(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, resourceName, &v),
					testAccCheckParameterGroupAttributes(&v, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrFamily, "mysql5.6"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "collation_connection",
						names.AttrValue: "utf8_unicode_ci",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_results",
						names.AttrValue: "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_server",
						names.AttrValue: "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "collation_server",
						names.AttrValue: "utf8_unicode_ci",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_client",
						names.AttrValue: "utf8",
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "rds", regexache.MustCompile(fmt.Sprintf("pg:%s$", rName))),
				),
			},
			{
				Config: testAccParameterGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, resourceName, &v),
					testAccCheckParameterGroupAttributes(&v, rName),
					testAccCheckParameterNotUserDefined(ctx, resourceName, "collation_connection"),
					testAccCheckParameterNotUserDefined(ctx, resourceName, "collation_server"),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_results",
						names.AttrValue: "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_server",
						names.AttrValue: "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_client",
						names.AttrValue: "utf8",
					}),
				),
			},
		},
	})
}

func TestAccRDSParameterGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.DBParameterGroup
	resourceName := "aws_db_parameter_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfrds.ResourceParameterGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRDSParameterGroup_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.DBParameterGroup
	resourceName := "aws_db_parameter_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccParameterGroupConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccParameterGroupConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccRDSParameterGroup_caseWithMixedParameters(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupConfig_caseWithMixedParameters(rName),
				Check:  resource.ComposeTestCheckFunc(),
			},
		},
	})
}

func TestAccRDSParameterGroup_limit(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.DBParameterGroup
	resourceName := "aws_db_parameter_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupConfig_exceedDefaultLimit(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, resourceName, &v),
					testAccCheckParameterGroupAttributes(&v, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrFamily, "mysql5.6"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "RDS default parameter group: Exceed default AWS parameter group limit of twenty"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_server",
						names.AttrValue: "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_client",
						names.AttrValue: "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "collation_server",
						names.AttrValue: "utf8_general_ci",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "collation_connection",
						names.AttrValue: "utf8_general_ci",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "join_buffer_size",
						names.AttrValue: "16777216",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "key_buffer_size",
						names.AttrValue: "67108864",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "max_connections",
						names.AttrValue: "3200",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "max_heap_table_size",
						names.AttrValue: "67108864",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "performance_schema",
						names.AttrValue: acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "performance_schema_users_size",
						names.AttrValue: "1048576",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "query_cache_limit",
						names.AttrValue: "2097152",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "query_cache_size",
						names.AttrValue: "67108864",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "sort_buffer_size",
						names.AttrValue: "16777216",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "table_open_cache",
						names.AttrValue: "4096",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "tmp_table_size",
						names.AttrValue: "67108864",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "binlog_cache_size",
						names.AttrValue: "131072",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "innodb_flush_log_at_trx_commit",
						names.AttrValue: acctest.Ct0,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "innodb_open_files",
						names.AttrValue: "4000",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "innodb_read_io_threads",
						names.AttrValue: "64",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "innodb_thread_concurrency",
						names.AttrValue: acctest.Ct0,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "innodb_write_io_threads",
						names.AttrValue: "64",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_connection",
						names.AttrValue: "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_database",
						names.AttrValue: "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_filesystem",
						names.AttrValue: "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_results",
						names.AttrValue: "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "event_scheduler",
						names.AttrValue: "on",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "innodb_buffer_pool_dump_at_shutdown",
						names.AttrValue: acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "innodb_file_format",
						names.AttrValue: "barracuda",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "innodb_io_capacity",
						names.AttrValue: "2000",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "innodb_io_capacity_max",
						names.AttrValue: "3000",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "innodb_lock_wait_timeout",
						names.AttrValue: "120",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "innodb_max_dirty_pages_pct",
						names.AttrValue: "90",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "log_bin_trust_function_creators",
						names.AttrValue: acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "log_warnings",
						names.AttrValue: acctest.Ct2,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "log_output",
						names.AttrValue: "FILE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "max_allowed_packet",
						names.AttrValue: "1073741824",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "max_connect_errors",
						names.AttrValue: "100",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "query_cache_min_res_unit",
						names.AttrValue: "512",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "slow_query_log",
						names.AttrValue: acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "sync_binlog",
						names.AttrValue: acctest.Ct0,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "tx_isolation",
						names.AttrValue: "repeatable-read",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccParameterGroupConfig_updateExceedDefaultLimit(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, resourceName, &v),
					testAccCheckParameterGroupAttributes(&v, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrFamily, "mysql5.6"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Updated RDS default parameter group: Exceed default AWS parameter group limit of twenty"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_server",
						names.AttrValue: "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_client",
						names.AttrValue: "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "collation_server",
						names.AttrValue: "utf8_general_ci",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "collation_connection",
						names.AttrValue: "utf8_general_ci",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "join_buffer_size",
						names.AttrValue: "16777216",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "key_buffer_size",
						names.AttrValue: "67108864",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "max_connections",
						names.AttrValue: "3200",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "max_heap_table_size",
						names.AttrValue: "67108864",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "performance_schema",
						names.AttrValue: acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "performance_schema_users_size",
						names.AttrValue: "1048576",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "query_cache_limit",
						names.AttrValue: "2097152",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "query_cache_size",
						names.AttrValue: "67108864",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "sort_buffer_size",
						names.AttrValue: "16777216",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "table_open_cache",
						names.AttrValue: "4096",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "tmp_table_size",
						names.AttrValue: "67108864",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "binlog_cache_size",
						names.AttrValue: "131072",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "innodb_flush_log_at_trx_commit",
						names.AttrValue: acctest.Ct0,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "innodb_open_files",
						names.AttrValue: "4000",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "innodb_read_io_threads",
						names.AttrValue: "64",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "innodb_thread_concurrency",
						names.AttrValue: acctest.Ct0,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "innodb_write_io_threads",
						names.AttrValue: "64",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_connection",
						names.AttrValue: "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_database",
						names.AttrValue: "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_filesystem",
						names.AttrValue: "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_results",
						names.AttrValue: "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "event_scheduler",
						names.AttrValue: "on",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "innodb_buffer_pool_dump_at_shutdown",
						names.AttrValue: acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "innodb_file_format",
						names.AttrValue: "barracuda",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "innodb_io_capacity",
						names.AttrValue: "2000",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "innodb_io_capacity_max",
						names.AttrValue: "3000",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "innodb_lock_wait_timeout",
						names.AttrValue: "120",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "innodb_max_dirty_pages_pct",
						names.AttrValue: "90",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "log_bin_trust_function_creators",
						names.AttrValue: acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "log_warnings",
						names.AttrValue: acctest.Ct2,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "log_output",
						names.AttrValue: "FILE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "max_allowed_packet",
						names.AttrValue: "1073741824",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "max_connect_errors",
						names.AttrValue: "100",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "query_cache_min_res_unit",
						names.AttrValue: "512",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "slow_query_log",
						names.AttrValue: acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "sync_binlog",
						names.AttrValue: acctest.Ct0,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "tx_isolation",
						names.AttrValue: "repeatable-read",
					}),
				),
			},
		},
	})
}

func TestAccRDSParameterGroup_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.DBParameterGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDBParameterGroupConfig_namePrefix,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, "aws_db_parameter_group.test", &v),
					resource.TestMatchResourceAttr("aws_db_parameter_group.test", names.AttrName, regexache.MustCompile("^tf-test-")),
				),
			},
		},
	})
}

func TestAccRDSParameterGroup_generatedName(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.DBParameterGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDBParameterGroupConfig_generatedName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, "aws_db_parameter_group.test", &v),
				),
			},
		},
	})
}

func TestAccRDSParameterGroup_withApplyMethod(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.DBParameterGroup
	resourceName := "aws_db_parameter_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupConfig_applyMethod(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, resourceName, &v),
					testAccCheckParameterGroupAttributes(&v, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrFamily, "mysql5.6"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Managed by Terraform"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_server",
						names.AttrValue: "utf8",
						"apply_method":  "immediate",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_client",
						names.AttrValue: "utf8",
						"apply_method":  "pending-reboot",
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
	ctx := acctest.Context(t)
	var v types.DBParameterGroup
	resourceName := "aws_db_parameter_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupConfig_only(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, resourceName, &v),
					testAccCheckParameterGroupAttributes(&v, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrFamily, "mysql5.6"),
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
	ctx := acctest.Context(t)
	var v types.DBParameterGroup
	resourceName := "aws_db_parameter_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupConfig_includeDefault(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrFamily, "postgres9.4"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrParameter},
			},
		},
	})
}

func TestAccRDSParameterGroup_updateParameters(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.DBParameterGroup
	resourceName := "aws_db_parameter_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupConfig_updateParametersInitial(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, resourceName, &v),
					testAccCheckParameterGroupAttributes(&v, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrFamily, "mysql5.6"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_results",
						names.AttrValue: "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_server",
						names.AttrValue: "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_client",
						names.AttrValue: "utf8",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccParameterGroupConfig_updateParametersUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, resourceName, &v),
					testAccCheckParameterGroupAttributes(&v, rName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_results",
						names.AttrValue: "ascii",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_server",
						names.AttrValue: "ascii",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_client",
						names.AttrValue: "utf8",
					}),
				),
			},
		},
	})
}

func TestAccRDSParameterGroup_caseParameters(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.DBParameterGroup
	resourceName := "aws_db_parameter_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupConfig_upperCase(rName, "Max_connections"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, resourceName, &v),
					testAccCheckParameterGroupAttributes(&v, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrFamily, "mysql5.6"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "Max_connections",
						names.AttrValue: "LEAST({DBInstanceClassMemory/6000000},10)",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"parameter.0.name"},
			},
			{
				Config: testAccParameterGroupConfig_upperCase(rName, "max_connections"),
			},
		},
	})
}

func TestParameterGroupModifyChunk(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name              string
		ChunkSize         int
		Parameters        []types.Parameter
		ExpectedModify    []types.Parameter
		ExpectedRemainder []types.Parameter
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
			Parameters: []types.Parameter{
				{
					ApplyMethod:    types.ApplyMethodImmediate,
					ParameterName:  aws.String("tx_isolation"),
					ParameterValue: aws.String("repeatable-read"),
				},
				{
					ApplyMethod:    types.ApplyMethodImmediate,
					ParameterName:  aws.String("binlog_cache_size"),
					ParameterValue: aws.String("131072"),
				},
			},
			ExpectedModify: []types.Parameter{
				{
					ApplyMethod:    types.ApplyMethodImmediate,
					ParameterName:  aws.String("tx_isolation"),
					ParameterValue: aws.String("repeatable-read"),
				},
				{
					ApplyMethod:    types.ApplyMethodImmediate,
					ParameterName:  aws.String("binlog_cache_size"),
					ParameterValue: aws.String("131072"),
				},
			},
			ExpectedRemainder: nil,
		},
		{
			Name:      "Over 3 max, 6 in",
			ChunkSize: 3,
			Parameters: []types.Parameter{
				{
					ApplyMethod:    types.ApplyMethodImmediate,
					ParameterName:  aws.String("tx_isolation"),
					ParameterValue: aws.String("repeatable-read"),
				},
				{
					ApplyMethod:    types.ApplyMethodImmediate,
					ParameterName:  aws.String("binlog_cache_size"),
					ParameterValue: aws.String("131072"),
				},
				{
					ApplyMethod:    types.ApplyMethodPendingReboot,
					ParameterName:  aws.String("innodb_read_io_threads"),
					ParameterValue: aws.String("64"),
				},
				{
					ApplyMethod:    types.ApplyMethodImmediate,
					ParameterName:  aws.String("character_set_server"),
					ParameterValue: aws.String("utf8"),
				},
				{
					ApplyMethod:    types.ApplyMethodImmediate,
					ParameterName:  aws.String("innodb_flush_log_at_trx_commit"),
					ParameterValue: aws.String(acctest.Ct0),
				},
				{
					ApplyMethod:    types.ApplyMethodImmediate,
					ParameterName:  aws.String("character_set_filesystem"),
					ParameterValue: aws.String("utf8"),
				},
			},
			ExpectedModify: []types.Parameter{
				{
					ApplyMethod:    types.ApplyMethodImmediate,
					ParameterName:  aws.String("character_set_server"),
					ParameterValue: aws.String("utf8"),
				},
				{
					ApplyMethod:    types.ApplyMethodImmediate,
					ParameterName:  aws.String("character_set_filesystem"),
					ParameterValue: aws.String("utf8"),
				},
				{
					ApplyMethod:    types.ApplyMethodImmediate,
					ParameterName:  aws.String("tx_isolation"),
					ParameterValue: aws.String("repeatable-read"),
				},
			},
			ExpectedRemainder: []types.Parameter{
				{
					ApplyMethod:    types.ApplyMethodImmediate,
					ParameterName:  aws.String("binlog_cache_size"),
					ParameterValue: aws.String("131072"),
				},
				{
					ApplyMethod:    types.ApplyMethodPendingReboot,
					ParameterName:  aws.String("innodb_read_io_threads"),
					ParameterValue: aws.String("64"),
				},
				{
					ApplyMethod:    types.ApplyMethodImmediate,
					ParameterName:  aws.String("innodb_flush_log_at_trx_commit"),
					ParameterValue: aws.String(acctest.Ct0),
				},
			},
		},
		{
			Name:      "Over 3 max, 9 in",
			ChunkSize: 3,
			Parameters: []types.Parameter{
				{
					ApplyMethod:    types.ApplyMethodPendingReboot,
					ParameterName:  aws.String("tx_isolation"),
					ParameterValue: aws.String("repeatable-read"),
				},
				{
					ApplyMethod:    types.ApplyMethodPendingReboot,
					ParameterName:  aws.String("binlog_cache_size"),
					ParameterValue: aws.String("131072"),
				},
				{
					ApplyMethod:    types.ApplyMethodPendingReboot,
					ParameterName:  aws.String("innodb_read_io_threads"),
					ParameterValue: aws.String("64"),
				},
				{
					ApplyMethod:    types.ApplyMethodPendingReboot,
					ParameterName:  aws.String("character_set_server"),
					ParameterValue: aws.String("utf8"),
				},
				{
					ApplyMethod:    types.ApplyMethodImmediate,
					ParameterName:  aws.String("innodb_flush_log_at_trx_commit"),
					ParameterValue: aws.String(acctest.Ct0),
				},
				{
					ApplyMethod:    types.ApplyMethodImmediate,
					ParameterName:  aws.String("character_set_filesystem"),
					ParameterValue: aws.String("utf8"),
				},
				{
					ApplyMethod:    types.ApplyMethodPendingReboot,
					ParameterName:  aws.String("innodb_max_dirty_pages_pct"),
					ParameterValue: aws.String("90"),
				},
				{
					ApplyMethod:    types.ApplyMethodImmediate,
					ParameterName:  aws.String("character_set_connection"),
					ParameterValue: aws.String("utf8"),
				},
				{
					ApplyMethod:    types.ApplyMethodImmediate,
					ParameterName:  aws.String("key_buffer_size"),
					ParameterValue: aws.String("67108864"),
				},
			},
			ExpectedModify: []types.Parameter{
				{
					ApplyMethod:    types.ApplyMethodImmediate,
					ParameterName:  aws.String("character_set_filesystem"),
					ParameterValue: aws.String("utf8"),
				},
				{
					ApplyMethod:    types.ApplyMethodImmediate,
					ParameterName:  aws.String("character_set_connection"),
					ParameterValue: aws.String("utf8"),
				},
				{
					ApplyMethod:    types.ApplyMethodImmediate,
					ParameterName:  aws.String("innodb_flush_log_at_trx_commit"),
					ParameterValue: aws.String(acctest.Ct0),
				},
			},
			ExpectedRemainder: []types.Parameter{
				{
					ApplyMethod:    types.ApplyMethodPendingReboot,
					ParameterName:  aws.String("tx_isolation"),
					ParameterValue: aws.String("repeatable-read"),
				},
				{
					ApplyMethod:    types.ApplyMethodPendingReboot,
					ParameterName:  aws.String("binlog_cache_size"),
					ParameterValue: aws.String("131072"),
				},
				{
					ApplyMethod:    types.ApplyMethodPendingReboot,
					ParameterName:  aws.String("innodb_read_io_threads"),
					ParameterValue: aws.String("64"),
				},
				{
					ApplyMethod:    types.ApplyMethodPendingReboot,
					ParameterName:  aws.String("character_set_server"),
					ParameterValue: aws.String("utf8"),
				},
				{
					ApplyMethod:    types.ApplyMethodPendingReboot,
					ParameterName:  aws.String("innodb_max_dirty_pages_pct"),
					ParameterValue: aws.String("90"),
				},
				{
					ApplyMethod:    types.ApplyMethodImmediate,
					ParameterName:  aws.String("key_buffer_size"),
					ParameterValue: aws.String("67108864"),
				},
			},
		},
	}

	for _, tc := range cases {
		mod, rem := tfrds.ParameterGroupModifyChunk(tc.Parameters, tc.ChunkSize)
		if !reflect.DeepEqual(mod, tc.ExpectedModify) {
			t.Errorf("Case %q: Modify did not match\n%#v\n\nGot:\n%#v", tc.Name, tc.ExpectedModify, mod)
		}
		if !reflect.DeepEqual(rem, tc.ExpectedRemainder) {
			t.Errorf("Case %q: Remainder did not match\n%#v\n\nGot:\n%#v", tc.Name, tc.ExpectedRemainder, rem)
		}
	}
}

func testAccCheckParameterGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_db_parameter_group" {
				continue
			}

			_, err := tfrds.FindDBParameterGroupByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("RDS DB Parameter Group %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckParameterGroupAttributes(v *types.DBParameterGroup, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *v.DBParameterGroupName != name {
			return fmt.Errorf("Bad Parameter Group name, expected (%s), got (%s)", name, *v.DBParameterGroupName)
		}

		family := "mysql5.6"
		if aws.ToString(v.DBParameterGroupFamily) != family {
			return fmt.Errorf("bad family, got: %s, expecting: %s", aws.ToString(v.DBParameterGroupFamily), family)
		}

		return nil
	}
}

func testAccCheckParameterGroupExists(ctx context.Context, n string, v *types.DBParameterGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSClient(ctx)

		output, err := tfrds.FindDBParameterGroupByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckParameterNotUserDefined(ctx context.Context, rName, paramName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rName]
		if !ok {
			return fmt.Errorf("Not found: %s", rName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSClient(ctx)

		input := &rds.DescribeDBParametersInput{
			DBParameterGroupName: aws.String(rs.Primary.ID),
			Source:               aws.String("user"),
		}

		userDefined := false
		pages := rds.NewDescribeDBParametersPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)

			if err != nil {
				return err
			}

			for _, param := range page.Parameters {
				if aws.ToString(param.ParameterName) == paramName {
					userDefined = true
				}
			}
		}

		if userDefined {
			return fmt.Errorf("DB Parameter is user defined")
		}

		return nil
	}
}

func testAccParameterGroupConfig_basic(rName string) string {
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

func testAccParameterGroupConfig_tags1(rName, tagKey1, tagValue1 string) string {
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

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccParameterGroupConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccParameterGroupConfig_caseWithMixedParameters(rName string) string {
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

func testAccParameterGroupConfig_applyMethod(rName string) string {
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

func testAccParameterGroupConfig_addParameters(rName string) string {
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

func testAccParameterGroupConfig_only(rName string) string {
	return fmt.Sprintf(`
resource "aws_db_parameter_group" "test" {
  name        = %[1]q
  family      = "mysql5.6"
  description = "Test parameter group for terraform"
}
`, rName)
}

func testAccParameterGroupConfig_exceedDefaultLimit(rName string) string {
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

func testAccParameterGroupConfig_updateExceedDefaultLimit(rName string) string {
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

func testAccParameterGroupConfig_includeDefault(rName string) string {
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

func testAccParameterGroupConfig_updateParametersInitial(rName string) string {
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

func testAccParameterGroupConfig_updateParametersUpdated(rName string) string {
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

func testAccParameterGroupConfig_upperCase(rName, paramName string) string {
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
