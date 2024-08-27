// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfglue "github.com/hashicorp/terraform-provider-aws/internal/service/glue"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccGlueCrawler_dynamoDBTarget(t *testing.T) {
	ctx := acctest.Context(t)
	var crawler awstypes.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrawlerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrawlerConfig_dynamoDBTarget(rName, "table1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "classifiers.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrConfiguration, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.0.path", fmt.Sprintf("%s-%s", rName, "table1")),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.0.scan_all", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrRole, rName),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, ""),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.delete_behavior", "DEPRECATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.update_behavior", "UPDATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "table_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccCrawlerConfig_dynamoDBTarget(rName, "table2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "classifiers.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrConfiguration, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.0.path", fmt.Sprintf("%s-%s", rName, "table2")),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.0.scan_all", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrRole, rName),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, ""),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.delete_behavior", "DEPRECATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.update_behavior", "UPDATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "table_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func TestAccGlueCrawler_DynamoDBTarget_scanAll(t *testing.T) {
	ctx := acctest.Context(t)
	var crawler awstypes.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrawlerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrawlerConfig_dynamoDBTargetScanAll(rName, "table1", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.0.path", fmt.Sprintf("%s-%s", rName, "table1")),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.0.scan_all", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCrawlerConfig_dynamoDBTargetScanAll(rName, "table1", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.0.path", fmt.Sprintf("%s-%s", rName, "table1")),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.0.scan_all", acctest.CtTrue),
				),
			},
			{
				Config: testAccCrawlerConfig_dynamoDBTargetScanAll(rName, "table1", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.0.path", fmt.Sprintf("%s-%s", rName, "table1")),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.0.scan_all", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccGlueCrawler_DynamoDBTarget_scanRate(t *testing.T) {
	ctx := acctest.Context(t)
	var crawler awstypes.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrawlerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrawlerConfig_dynamoDBTargetScanRate(rName, "table1", 0.5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.0.path", fmt.Sprintf("%s-%s", rName, "table1")),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.0.scan_rate", "0.5"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCrawlerConfig_dynamoDBTargetScanRate(rName, "table1", 1.5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.0.path", fmt.Sprintf("%s-%s", rName, "table1")),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.0.scan_rate", "1.5"),
				),
			},
			{
				Config: testAccCrawlerConfig_dynamoDBTargetScanRate(rName, "table1", 0.5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.0.path", fmt.Sprintf("%s-%s", rName, "table1")),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.0.scan_rate", "0.5"),
				),
			},
		},
	})
}

func TestAccGlueCrawler_jdbcTarget(t *testing.T) {
	ctx := acctest.Context(t)
	var crawler awstypes.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	jdbcConnectionUrl := fmt.Sprintf("jdbc:mysql://%s/testdatabase", acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrawlerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrawlerConfig_jdbcTarget(rName, jdbcConnectionUrl, "database-name/%"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "classifiers.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrConfiguration, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.exclusions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.path", "database-name/%"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.enable_additional_metadata.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrRole, rName),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, ""),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.delete_behavior", "DEPRECATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.update_behavior", "UPDATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "table_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccCrawlerConfig_jdbcTarget(rName, jdbcConnectionUrl, "database-name/table-name"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "classifiers.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrConfiguration, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.exclusions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.path", "database-name/table-name"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.enable_additional_metadata.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrRole, rName),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, ""),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.delete_behavior", "DEPRECATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.update_behavior", "UPDATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "table_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCrawlerConfig_jdbcTargetMetadata(rName, jdbcConnectionUrl, "database-name/table-name"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "classifiers.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrConfiguration, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.exclusions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.path", "database-name/table-name"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.enable_additional_metadata.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrRole, rName),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, ""),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.delete_behavior", "DEPRECATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.update_behavior", "UPDATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "table_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
		},
	})
}

func TestAccGlueCrawler_JDBCTarget_exclusions(t *testing.T) {
	ctx := acctest.Context(t)
	var crawler awstypes.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	jdbcConnectionUrl := fmt.Sprintf("jdbc:mysql://%s/testdatabase", acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrawlerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrawlerConfig_jdbcTargetExclusions2(rName, jdbcConnectionUrl, "exclusion1", "exclusion2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.exclusions.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.exclusions.0", "exclusion1"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.exclusions.1", "exclusion2"),
				),
			},
			{
				Config: testAccCrawlerConfig_jdbcTargetExclusions1(rName, jdbcConnectionUrl, "exclusion1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.exclusions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.exclusions.0", "exclusion1"),
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

func TestAccGlueCrawler_JDBCTarget_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	var crawler awstypes.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"
	jdbcConnectionUrl := fmt.Sprintf("jdbc:mysql://%s/testdatabase", acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrawlerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrawlerConfig_jdbcTargetMultiple(rName, jdbcConnectionUrl, "database-name/table1", "database-name/table2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.exclusions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.path", "database-name/table1"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.1.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.1.exclusions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.1.path", "database-name/table2"),
				),
			},
			{
				Config: testAccCrawlerConfig_jdbcTarget(rName, jdbcConnectionUrl, "database-name/table1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.exclusions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.path", "database-name/table1"),
				),
			},
			{
				Config: testAccCrawlerConfig_jdbcTargetMultiple(rName, jdbcConnectionUrl, "database-name/table1", "database-name/table2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.#", acctest.Ct2),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.exclusions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.path", "database-name/table1"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.1.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.1.exclusions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.1.path", "database-name/table2"),
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

func TestAccGlueCrawler_mongoDBTarget(t *testing.T) {
	ctx := acctest.Context(t)
	var crawler awstypes.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"
	connectionURL := "mongodb://" + net.JoinHostPort(acctest.RandomDomainName(), "27017") + "/testdatabase"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrawlerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrawlerConfig_mongoDBTarget(rName, connectionURL, "database-name/%"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.scan_all", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.path", "database-name/%"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCrawlerConfig_mongoDBTarget(rName, connectionURL, "database-name/table-name"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.scan_all", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.path", "database-name/table-name"),
				),
			},
		},
	})
}

func TestAccGlueCrawler_MongoDBTargetScan_all(t *testing.T) {
	ctx := acctest.Context(t)
	var crawler awstypes.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"
	connectionURL := "mongodb://" + net.JoinHostPort(acctest.RandomDomainName(), "27017") + "/testdatabase"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrawlerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrawlerConfig_mongoDBTargetScanAll(rName, connectionURL, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.scan_all", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.path", "database-name/table-name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCrawlerConfig_mongoDBTargetScanAll(rName, connectionURL, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.scan_all", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.path", "database-name/table-name"),
				),
			},
			{
				Config: testAccCrawlerConfig_mongoDBTargetScanAll(rName, connectionURL, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.scan_all", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.path", "database-name/table-name"),
				),
			},
		},
	})
}

func TestAccGlueCrawler_MongoDBTarget_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	var crawler awstypes.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"
	connectionURL := "mongodb://" + net.JoinHostPort(acctest.RandomDomainName(), "27017") + "/testdatabase"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrawlerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrawlerConfig_mongoDBMultiple(rName, connectionURL, "database-name/table1", "database-name/table2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.scan_all", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.path", "database-name/table1"),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.1.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.1.scan_all", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.1.path", "database-name/table2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCrawlerConfig_mongoDBTarget(rName, connectionURL, "database-name/%"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.scan_all", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.path", "database-name/%"),
				),
			},
			{
				Config: testAccCrawlerConfig_mongoDBMultiple(rName, connectionURL, "database-name/table1", "database-name/table2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.scan_all", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.path", "database-name/table1"),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.1.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.1.scan_all", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.1.path", "database-name/table2"),
				),
			},
		},
	})
}

func TestAccGlueCrawler_deltaTarget(t *testing.T) {
	ctx := acctest.Context(t)
	var crawler awstypes.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"
	connectionURL := "mongodb://" + net.JoinHostPort(acctest.RandomDomainName(), "27017") + "/testdatabase"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrawlerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrawlerConfig_deltaTarget(rName, connectionURL, "s3://table1", "null"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "delta_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "delta_target.0.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "delta_target.0.create_native_delta_table", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "delta_target.0.delta_tables.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "delta_target.0.delta_tables.*", "s3://table1"),
					resource.TestCheckResourceAttr(resourceName, "delta_target.0.write_manifest", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCrawlerConfig_deltaTarget(rName, connectionURL, "s3://table2", acctest.CtTrue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "delta_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "delta_target.0.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "delta_target.0.create_native_delta_table", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "delta_target.0.delta_tables.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "delta_target.0.delta_tables.*", "s3://table2"),
					resource.TestCheckResourceAttr(resourceName, "delta_target.0.write_manifest", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccGlueCrawler_hudiTarget(t *testing.T) {
	ctx := acctest.Context(t)
	var crawler awstypes.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"
	connectionURL := "mongodb://" + net.JoinHostPort(acctest.RandomDomainName(), "27017") + "/testdatabase"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrawlerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrawlerConfig_hudiTarget(rName, connectionURL, "s3://table1", 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "hudi_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "hudi_target.0.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "hudi_target.0.maximum_traversal_depth", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "hudi_target.0.paths.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "hudi_target.0.paths.*", "s3://table1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCrawlerConfig_hudiTarget(rName, connectionURL, "s3://table2", 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "hudi_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "hudi_target.0.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "hudi_target.0.maximum_traversal_depth", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "hudi_target.0.paths.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "hudi_target.0.paths.*", "s3://table2"),
				),
			},
		},
	})
}

func TestAccGlueCrawler_icebergTarget(t *testing.T) {
	ctx := acctest.Context(t)
	var crawler awstypes.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"
	connectionURL := "mongodb://" + net.JoinHostPort(acctest.RandomDomainName(), "27017") + "/testdatabase"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrawlerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrawlerConfig_icebergTarget(rName, connectionURL, "s3://table1", 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "iceberg_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "iceberg_target.0.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "iceberg_target.0.maximum_traversal_depth", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "iceberg_target.0.paths.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "iceberg_target.0.paths.*", "s3://table1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCrawlerConfig_icebergTarget(rName, connectionURL, "s3://table2", 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "iceberg_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "iceberg_target.0.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "iceberg_target.0.maximum_traversal_depth", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "iceberg_target.0.paths.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "iceberg_target.0.paths.*", "s3://table2"),
				),
			},
		},
	})
}

func TestAccGlueCrawler_s3Target(t *testing.T) {
	ctx := acctest.Context(t)
	var crawler awstypes.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrawlerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrawlerConfig_s3Target(rName, "bucket1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "classifiers.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrConfiguration, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrRole, rName),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.exclusions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.path", fmt.Sprintf("s3://%s-bucket1", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, ""),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.delete_behavior", "DEPRECATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.update_behavior", "UPDATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "table_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccCrawlerConfig_s3Target(rName, "bucket2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "classifiers.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrConfiguration, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrRole, rName),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.exclusions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.path", fmt.Sprintf("s3://%s-bucket2", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, ""),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.delete_behavior", "DEPRECATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.update_behavior", "UPDATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "table_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func TestAccGlueCrawler_S3Target_connectionName(t *testing.T) {
	ctx := acctest.Context(t)
	var crawler awstypes.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"
	connectionName := "aws_glue_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrawlerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrawlerConfig_s3TargetConnectionName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "s3_target.0.connection_name", connectionName, names.AttrName),
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

func TestAccGlueCrawler_S3Target_sampleSize(t *testing.T) {
	ctx := acctest.Context(t)
	var crawler awstypes.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrawlerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrawlerConfig_s3TargetSampleSize(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.sample_size", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCrawlerConfig_s3TargetSampleSize(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.sample_size", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccGlueCrawler_S3Target_exclusions(t *testing.T) {
	ctx := acctest.Context(t)
	var crawler awstypes.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrawlerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrawlerConfig_s3TargetExclusions2(rName, "exclusion1", "exclusion2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.exclusions.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.exclusions.0", "exclusion1"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.exclusions.1", "exclusion2"),
				),
			},
			{
				Config: testAccCrawlerConfig_s3TargetExclusions1(rName, "exclusion1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.exclusions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.exclusions.0", "exclusion1"),
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

func TestAccGlueCrawler_S3Target_eventqueue(t *testing.T) {
	ctx := acctest.Context(t)
	var crawler awstypes.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrawlerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrawlerConfig_s3TargetEventQueue(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", acctest.Ct1),
					acctest.CheckResourceAttrRegionalARN(resourceName, "s3_target.0.event_queue_arn", "sqs", rName),
					resource.TestCheckResourceAttr(resourceName, "recrawl_policy.0.recrawl_behavior", "CRAWL_EVENT_MODE"),
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

func TestAccGlueCrawler_CatalogTarget_dlqeventqueue(t *testing.T) {
	ctx := acctest.Context(t)
	var crawler awstypes.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrawlerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrawlerConfig_catalogTargetDlqEventQueue(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "catalog_target.0.event_queue_arn", "aws_sqs_queue.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "catalog_target.0.dlq_event_queue_arn", "aws_sqs_queue.test_dlq", names.AttrARN),
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

func TestAccGlueCrawler_S3Target_dlqeventqueue(t *testing.T) {
	ctx := acctest.Context(t)
	var crawler awstypes.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrawlerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrawlerConfig_s3TargetDlqEventQueue(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "s3_target.0.event_queue_arn", "aws_sqs_queue.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "s3_target.0.dlq_event_queue_arn", "aws_sqs_queue.test_dlq", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "recrawl_policy.0.recrawl_behavior", "CRAWL_EVENT_MODE"),
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

func TestAccGlueCrawler_S3Target_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	var crawler awstypes.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrawlerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrawlerConfig_s3TargetMultiple(rName, "bucket1", "bucket2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.exclusions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.path", fmt.Sprintf("s3://%s-bucket1", rName)),
					resource.TestCheckResourceAttr(resourceName, "s3_target.1.exclusions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3_target.1.path", fmt.Sprintf("s3://%s-bucket2", rName)),
				),
			},
			{
				Config: testAccCrawlerConfig_s3Target(rName, "bucket1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.exclusions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.path", fmt.Sprintf("s3://%s-bucket1", rName)),
				),
			},
			{
				Config: testAccCrawlerConfig_s3TargetMultiple(rName, "bucket1", "bucket2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.exclusions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.path", fmt.Sprintf("s3://%s-bucket1", rName)),
					resource.TestCheckResourceAttr(resourceName, "s3_target.1.exclusions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3_target.1.path", fmt.Sprintf("s3://%s-bucket2", rName)),
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

func TestAccGlueCrawler_catalogTarget(t *testing.T) {
	ctx := acctest.Context(t)
	var crawler awstypes.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrawlerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrawlerConfig_catalogTarget(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "classifiers.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrRole, rName),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.0.database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.0.tables.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.0.tables.0", fmt.Sprintf("%s_table_0", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, ""),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.delete_behavior", "LOG"),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.update_behavior", "UPDATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "table_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrConfiguration, "{\"Version\":1.0,\"Grouping\":{\"TableGroupingPolicy\":\"CombineCompatibleSchemas\"}}"),
				),
			},
			{
				Config: testAccCrawlerConfig_catalogTarget(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "classifiers.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrRole, rName),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.0.database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.0.tables.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.0.tables.0", fmt.Sprintf("%s_table_0", rName)),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.0.tables.1", fmt.Sprintf("%s_table_1", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, ""),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.delete_behavior", "LOG"),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.update_behavior", "UPDATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "table_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrConfiguration, "{\"Version\":1.0,\"Grouping\":{\"TableGroupingPolicy\":\"CombineCompatibleSchemas\"}}"),
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

func TestAccGlueCrawler_CatalogTarget_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	var crawler awstypes.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrawlerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrawlerConfig_catalogTarget(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.0.database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.0.tables.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.0.tables.0", fmt.Sprintf("%s_table_0", rName)),
				),
			},
			{
				Config: testAccCrawlerConfig_catalogTargetMultiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.0.database_name", fmt.Sprintf("%s_database_0", rName)),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.1.database_name", fmt.Sprintf("%s_database_1", rName)),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.0.tables.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.0.tables.0", fmt.Sprintf("%s_table_0", rName)),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.1.tables.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.1.tables.0", fmt.Sprintf("%s_table_1", rName)),
				),
			},
			{
				Config: testAccCrawlerConfig_catalogTarget(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.0.database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.0.tables.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.0.tables.0", fmt.Sprintf("%s_table_0", rName)),
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

func TestAccGlueCrawler_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var crawler awstypes.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrawlerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrawlerConfig_s3Target(rName, "bucket1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfglue.ResourceCrawler(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccGlueCrawler_classifiers(t *testing.T) {
	ctx := acctest.Context(t)
	var crawler awstypes.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrawlerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrawlerConfig_classifiersSingle(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "classifiers.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "classifiers.0", rName+acctest.Ct1),
				),
			},
			{
				Config: testAccCrawlerConfig_classifiersMultiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "classifiers.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "classifiers.0", rName+acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "classifiers.1", rName+acctest.Ct2),
				),
			},
			{
				Config: testAccCrawlerConfig_classifiersSingle(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "classifiers.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "classifiers.0", rName+acctest.Ct1),
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

func TestAccGlueCrawler_Configuration(t *testing.T) {
	ctx := acctest.Context(t)
	var crawler awstypes.Crawler
	configuration1 := `{"Version": 1.0, "CrawlerOutput": {"Tables": { "AddOrUpdateBehavior": "MergeNewColumns" }}}`
	configuration2 := `{"Version": 1.0, "CrawlerOutput": {"Partitions": { "AddOrUpdateBehavior": "InheritFromTable" }}}`
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrawlerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrawlerConfig_configuration(rName, configuration1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					testAccCheckCrawlerConfiguration(&crawler, configuration1),
				),
			},
			{
				Config: testAccCrawlerConfig_configuration(rName, configuration2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					testAccCheckCrawlerConfiguration(&crawler, configuration2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCrawlerConfig_configuration(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, names.AttrConfiguration, ""),
				),
			},
		},
	})
}

func TestAccGlueCrawler_description(t *testing.T) {
	ctx := acctest.Context(t)
	var crawler awstypes.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrawlerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrawlerConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
				),
			},
			{
				Config: testAccCrawlerConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
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

func TestAccGlueCrawler_RoleARN_noPath(t *testing.T) {
	ctx := acctest.Context(t)
	var crawler awstypes.Crawler
	iamRoleResourceName := "aws_iam_role.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrawlerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrawlerConfig_roleARNNoPath(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRole, iamRoleResourceName, names.AttrName),
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

func TestAccGlueCrawler_RoleARN_path(t *testing.T) {
	ctx := acctest.Context(t)
	var crawler awstypes.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrawlerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrawlerConfig_roleARNPath(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, names.AttrRole, fmt.Sprintf("path/%s", rName)),
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

func TestAccGlueCrawler_RoleName_path(t *testing.T) {
	ctx := acctest.Context(t)
	var crawler awstypes.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrawlerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrawlerConfig_roleNamePath(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, names.AttrRole, fmt.Sprintf("path/%s", rName)),
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

func TestAccGlueCrawler_schedule(t *testing.T) {
	ctx := acctest.Context(t)
	var crawler awstypes.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrawlerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrawlerConfig_schedule(rName, "cron(0 1 * * ? *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, "cron(0 1 * * ? *)"),
				),
			},
			{
				Config: testAccCrawlerConfig_schedule(rName, "cron(0 2 * * ? *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, "cron(0 2 * * ? *)"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCrawlerConfig_s3Target(rName, "bucket1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, ""),
				),
			},
		},
	})
}

func TestAccGlueCrawler_schemaChangePolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var crawler awstypes.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrawlerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrawlerConfig_schemaChangePolicy(rName, string(awstypes.DeleteBehaviorDeleteFromDatabase), string(awstypes.UpdateBehaviorUpdateInDatabase)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.delete_behavior", string(awstypes.DeleteBehaviorDeleteFromDatabase)),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.update_behavior", string(awstypes.UpdateBehaviorUpdateInDatabase)),
				),
			},
			{
				Config: testAccCrawlerConfig_schemaChangePolicy(rName, string(awstypes.DeleteBehaviorLog), string(awstypes.UpdateBehaviorLog)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.delete_behavior", string(awstypes.DeleteBehaviorLog)),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.update_behavior", string(awstypes.UpdateBehaviorLog)),
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

func TestAccGlueCrawler_tablePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var crawler awstypes.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrawlerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrawlerConfig_tablePrefix(rName, "prefix1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "table_prefix", "prefix1"),
				),
			},
			{
				Config: testAccCrawlerConfig_tablePrefix(rName, "prefix2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "table_prefix", "prefix2"),
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

func TestAccGlueCrawler_removeTablePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var crawler awstypes.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrawlerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrawlerConfig_tablePrefix(rName, names.AttrPrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "table_prefix", names.AttrPrefix),
				),
			},
			{
				Config: testAccCrawlerConfig_tablePrefix(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "table_prefix", ""),
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

func TestAccGlueCrawler_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var crawler1, crawler2, crawler3 awstypes.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrawlerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrawlerConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler1),
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
				Config: testAccCrawlerConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccCrawlerConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler3),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccGlueCrawler_security(t *testing.T) {
	ctx := acctest.Context(t)
	var crawler awstypes.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrawlerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrawlerConfig_securityConfiguration(rName, "security_configuration1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "security_configuration", "security_configuration1"),
				),
			},
			{
				Config: testAccCrawlerConfig_securityConfiguration(rName, "security_configuration2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "security_configuration", "security_configuration2"),
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

func TestAccGlueCrawler_lineage(t *testing.T) {
	ctx := acctest.Context(t)
	var crawler awstypes.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrawlerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrawlerConfig_lineage(rName, "ENABLE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "lineage_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "lineage_configuration.0.crawler_lineage_settings", "ENABLE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCrawlerConfig_lineage(rName, "DISABLE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "lineage_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "lineage_configuration.0.crawler_lineage_settings", "DISABLE")),
			},
			{
				Config: testAccCrawlerConfig_lineage(rName, "ENABLE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "lineage_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "lineage_configuration.0.crawler_lineage_settings", "ENABLE"),
				),
			},
		},
	})
}

func TestAccGlueCrawler_lakeformation(t *testing.T) {
	ctx := acctest.Context(t)
	var crawler awstypes.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrawlerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrawlerConfig_lakeformation(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "lake_formation_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "lake_formation_configuration.0.use_lake_formation_credentials", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCrawlerConfig_lakeformation(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "lake_formation_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "lake_formation_configuration.0.use_lake_formation_credentials", acctest.CtFalse)),
			},
		},
	})
}

func TestAccGlueCrawler_reCrawlPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var crawler awstypes.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrawlerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrawlerConfig_recrawlPolicy(rName, "CRAWL_EVERYTHING"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "recrawl_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "recrawl_policy.0.recrawl_behavior", "CRAWL_EVERYTHING"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCrawlerConfig_recrawlPolicy(rName, "CRAWL_NEW_FOLDERS_ONLY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "recrawl_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "recrawl_policy.0.recrawl_behavior", "CRAWL_NEW_FOLDERS_ONLY")),
			},
			{
				Config: testAccCrawlerConfig_recrawlPolicy(rName, "CRAWL_EVERYTHING"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(ctx, resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "recrawl_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "recrawl_policy.0.recrawl_behavior", "CRAWL_EVERYTHING"),
				),
			},
		},
	})
}

func testAccCheckCrawlerExists(ctx context.Context, resourceName string, crawler *awstypes.Crawler) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueClient(ctx)
		output, err := tfglue.FindCrawlerByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*crawler = *output

		return nil
	}
}

func testAccCheckCrawlerDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_glue_crawler" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).GlueClient(ctx)
			_, err := tfglue.FindCrawlerByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Glue Crawler %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckCrawlerConfiguration(crawler *awstypes.Crawler, acctestJSON string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiJSON := aws.ToString(crawler.Configuration)
		apiJSONBuffer := bytes.NewBufferString("")
		if err := json.Compact(apiJSONBuffer, []byte(apiJSON)); err != nil {
			return fmt.Errorf("unable to compact API configuration JSON: %s", err)
		}

		acctestJSONBuffer := bytes.NewBufferString("")
		if err := json.Compact(acctestJSONBuffer, []byte(acctestJSON)); err != nil {
			return fmt.Errorf("unable to compact acceptance test configuration JSON: %s", err)
		}

		if !verify.JSONBytesEqual(apiJSONBuffer.Bytes(), acctestJSONBuffer.Bytes()) {
			return fmt.Errorf("expected configuration JSON to match %v, received JSON: %v", acctestJSON, apiJSON)
		}
		return nil
	}
}

func testAccCrawlerConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.assume.json
}

data "aws_iam_policy_document" "assume" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["glue.amazonaws.com"]
    }
  }
}

data "aws_iam_policy" "AWSGlueServiceRole" {
  arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSGlueServiceRole"
}

resource "aws_iam_role_policy_attachment" "test-AWSGlueServiceRole" {
  policy_arn = data.aws_iam_policy.AWSGlueServiceRole.arn
  role       = aws_iam_role.test.name
}

data "aws_iam_policy" "AmazonDynamoDBReadOnlyAccess" {
  arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonDynamoDBReadOnlyAccess"
}

resource "aws_iam_role_policy_attachment" "test-AmazonDynamoDBReadOnlyAccess" {
  policy_arn = data.aws_iam_policy.AmazonDynamoDBReadOnlyAccess.arn
  role       = aws_iam_role.test.name
}

resource "aws_iam_role_policy" "LakeFormationDataAccess" {
  role = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "LakeFormationDataAccess",
      "Effect": "Allow",
      "Action": [
        "lakeformation:GetDataAccess"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}
`, rName)
}

func testAccCrawlerConfig_classifiersSingle(rName string) string {
	return acctest.ConfigCompose(testAccCrawlerConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_classifier" "test1" {
  name = %[2]q

  grok_classifier {
    classification = "example"
    grok_pattern   = "example"
  }
}

resource "aws_glue_classifier" "test2" {
  name = %[3]q

  grok_classifier {
    classification = "example"
    grok_pattern   = "example"
  }
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  classifiers   = [aws_glue_classifier.test1.id]
  name          = %[4]q
  database_name = aws_glue_catalog_database.test.name
  role          = aws_iam_role.test.name

  s3_target {
    path = "s3://bucket-name"
  }
}
`, rName, rName+acctest.Ct1, rName+acctest.Ct2, rName))
}

func testAccCrawlerConfig_classifiersMultiple(rName string) string {
	return acctest.ConfigCompose(testAccCrawlerConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_classifier" "test1" {
  name = %[2]q

  grok_classifier {
    classification = "example"
    grok_pattern   = "example"
  }
}

resource "aws_glue_classifier" "test2" {
  name = %[3]q

  grok_classifier {
    classification = "example"
    grok_pattern   = "example"
  }
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  classifiers   = [aws_glue_classifier.test1.id, aws_glue_classifier.test2.id]
  name          = %[4]q
  database_name = aws_glue_catalog_database.test.name
  role          = aws_iam_role.test.name

  s3_target {
    path = "s3://bucket-name"
  }
}
`, rName, rName+acctest.Ct1, rName+acctest.Ct2, rName))
}

func testAccCrawlerConfig_configuration(rName, configuration string) string {
	return acctest.ConfigCompose(testAccCrawlerConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  configuration = %[2]s
  database_name = aws_glue_catalog_database.test.name
  name          = %[3]q
  role          = aws_iam_role.test.name

  s3_target {
    path = "s3://bucket-name"
  }
}
`, rName, strconv.Quote(configuration), rName))
}

func testAccCrawlerConfig_description(rName, description string) string {
	return acctest.ConfigCompose(testAccCrawlerConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  description   = %[2]q
  name          = %[3]q
  role          = aws_iam_role.test.name

  s3_target {
    path = "s3://bucket-name"
  }
}
`, rName, description, rName))
}

func testAccCrawlerConfig_dynamoDBTarget(rName, suffix string) string {
	return acctest.ConfigCompose(testAccCrawlerConfig_base(rName), fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = "%[1]s-%[2]s"
  read_capacity  = 1
  write_capacity = 1
  hash_key       = %[1]q

  attribute {
    name = %[1]q
    type = "S"
  }
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q
  role          = aws_iam_role.test.name

  dynamodb_target {
    path = aws_dynamodb_table.test.name
  }
}
`, rName, suffix))
}

func testAccCrawlerConfig_dynamoDBTargetScanAll(rName, suffix string, scanAll bool) string {
	return acctest.ConfigCompose(testAccCrawlerConfig_base(rName), fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = "%[1]s-%[2]s"
  read_capacity  = 1
  write_capacity = 1
  hash_key       = %[1]q

  attribute {
    name = %[1]q
    type = "S"
  }
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q
  role          = aws_iam_role.test.name

  dynamodb_target {
    path     = aws_dynamodb_table.test.name
    scan_all = %[3]t
  }
}
`, rName, suffix, scanAll))
}

func testAccCrawlerConfig_dynamoDBTargetScanRate(rName, suffix string, scanRate float64) string {
	return acctest.ConfigCompose(testAccCrawlerConfig_base(rName), fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = "%[1]s-%[2]s"
  read_capacity  = 1
  write_capacity = 1
  hash_key       = %[1]q

  attribute {
    name = %[1]q
    type = "S"
  }
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q
  role          = aws_iam_role.test.name

  dynamodb_target {
    path      = aws_dynamodb_table.test.name
    scan_rate = %[3]g
  }
}
`, rName, suffix, scanRate))
}

func testAccCrawlerConfig_jdbcTarget(rName, jdbcConnectionUrl, path string) string {
	return acctest.ConfigCompose(testAccCrawlerConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_connection" "test" {
  name = %[1]q

  connection_properties = {
    JDBC_CONNECTION_URL = %[2]q
    PASSWORD            = "testpassword"
    USERNAME            = "testusername"
  }
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q
  role          = aws_iam_role.test.name

  jdbc_target {
    connection_name = aws_glue_connection.test.name
    path            = %[3]q
  }
}
`, rName, jdbcConnectionUrl, path))
}

func testAccCrawlerConfig_jdbcTargetMetadata(rName, jdbcConnectionUrl, path string) string {
	return acctest.ConfigCompose(testAccCrawlerConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_connection" "test" {
  name = %[1]q

  connection_properties = {
    JDBC_CONNECTION_URL = %[1]q
    PASSWORD            = "testpassword"
    USERNAME            = "testusername"
  }
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q
  role          = aws_iam_role.test.name

  jdbc_target {
    connection_name            = aws_glue_connection.test.name
    path                       = %[3]q
    enable_additional_metadata = ["COMMENTS"]
  }
}
`, rName, jdbcConnectionUrl, path))
}

func testAccCrawlerConfig_jdbcTargetExclusions1(rName, jdbcConnectionUrl, exclusion1 string) string {
	return acctest.ConfigCompose(testAccCrawlerConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_connection" "test" {
  name = %[1]q

  connection_properties = {
    JDBC_CONNECTION_URL = %[2]q
    PASSWORD            = "testpassword"
    USERNAME            = "testusername"
  }
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q
  role          = aws_iam_role.test.name

  jdbc_target {
    connection_name = aws_glue_connection.test.name
    exclusions      = [%[3]q]
    path            = "database-name/table1"
  }
}
`, rName, jdbcConnectionUrl, exclusion1))
}

func testAccCrawlerConfig_jdbcTargetExclusions2(rName, jdbcConnectionUrl, exclusion1, exclusion2 string) string {
	return acctest.ConfigCompose(testAccCrawlerConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_connection" "test" {
  name = %[1]q

  connection_properties = {
    JDBC_CONNECTION_URL = %[2]q
    PASSWORD            = "testpassword"
    USERNAME            = "testusername"
  }
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q
  role          = aws_iam_role.test.name

  jdbc_target {
    connection_name = aws_glue_connection.test.name
    exclusions      = [%[3]q, %[4]q]
    path            = "database-name/table1"
  }
}
`, rName, jdbcConnectionUrl, exclusion1, exclusion2))
}

func testAccCrawlerConfig_jdbcTargetMultiple(rName, jdbcConnectionUrl, path1, path2 string) string {
	return acctest.ConfigCompose(testAccCrawlerConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_connection" "test" {
  name = %[1]q

  connection_properties = {
    JDBC_CONNECTION_URL = %[2]q
    PASSWORD            = "testpassword"
    USERNAME            = "testusername"
  }
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q
  role          = aws_iam_role.test.name

  jdbc_target {
    connection_name = aws_glue_connection.test.name
    path            = %[3]q
  }

  jdbc_target {
    connection_name = aws_glue_connection.test.name
    path            = %[4]q
  }
}
`, rName, jdbcConnectionUrl, path1, path2))
}

func testAccCrawlerConfig_roleARNNoPath(rName string) string {
	return acctest.ConfigCompose(testAccCrawlerConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q
  role          = aws_iam_role.test.arn

  s3_target {
    path = "s3://bucket-name"
  }
}
`, rName))
}

func testAccCrawlerConfig_roleARNPath(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/path/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "glue.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test-AWSGlueServiceRole" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSGlueServiceRole"
  role       = aws_iam_role.test.name
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q
  role          = aws_iam_role.test.arn

  s3_target {
    path = "s3://bucket-name"
  }
}
`, rName)
}

func testAccCrawlerConfig_roleNamePath(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/path/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "glue.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test-AWSGlueServiceRole" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSGlueServiceRole"
  role       = aws_iam_role.test.name
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q
  role          = "${replace(aws_iam_role.test.path, "/^\\//", "")}${aws_iam_role.test.name}"

  s3_target {
    path = "s3://${aws_s3_bucket.test.bucket}"
  }
}
`, rName)
}

func testAccCrawlerConfig_s3Target(rName, path string) string {
	return acctest.ConfigCompose(testAccCrawlerConfig_base(rName), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = "%[1]s-%[2]s"
  force_destroy = true
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q
  role          = aws_iam_role.test.name

  s3_target {
    path = "s3://${aws_s3_bucket.test.bucket}"
  }
}
`, rName, path))
}

func testAccCrawlerConfig_s3TargetExclusions1(rName, exclusion1 string) string {
	return acctest.ConfigCompose(testAccCrawlerConfig_base(rName), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q
  role          = aws_iam_role.test.name

  s3_target {
    exclusions = [%[2]q]
    path       = "s3://${aws_s3_bucket.test.bucket}"
  }
}
`, rName, exclusion1))
}

func testAccCrawlerConfig_s3TargetConnectionName(rName string) string {
	return acctest.ConfigCompose(testAccCrawlerConfig_base(rName), acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    from_port = 1
    protocol  = "tcp"
    self      = true
    to_port   = 65535
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_glue_catalog_database" "test" {
  name = "%[1]s"
}

resource "aws_glue_connection" "test" {
  connection_properties = {
    JDBC_ENFORCE_SSL = false
  }

  connection_type = "NETWORK"

  name = "%[1]s"

  physical_connection_requirements {
    availability_zone      = aws_subnet.test[0].availability_zone
    security_group_id_list = [aws_security_group.test.id]
    subnet_id              = aws_subnet.test[0].id
  }
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  name          = "%[1]s"
  role          = aws_iam_role.test.name

  s3_target {
    connection_name = aws_glue_connection.test.name
    path            = "s3://${aws_s3_bucket.test.bucket}"
  }
}
`, rName))
}

func testAccCrawlerConfig_s3TargetExclusions2(rName, exclusion1, exclusion2 string) string {
	return acctest.ConfigCompose(testAccCrawlerConfig_base(rName), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q
  role          = aws_iam_role.test.name

  s3_target {
    exclusions = [%[2]q, %[3]q]
    path       = "s3://${aws_s3_bucket.test.bucket}"
  }
}
`, rName, exclusion1, exclusion2))
}

func testAccCrawlerConfig_s3TargetEventQueue(rName string) string {
	return acctest.ConfigCompose(testAccCrawlerConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_sqs_queue" "test" {
  name = %[1]q

  visibility_timeout_seconds = 3600
}

resource "aws_iam_role_policy" "test_sqs" {
  role = aws_iam_role.test.name

  policy = data.aws_iam_policy_document.role_test_sqs.json
}

data "aws_iam_policy_document" "role_test_sqs" {
  statement {
    effect = "Allow"

    actions = [
      "sqs:DeleteMessage",
      "sqs:GetQueueUrl",
      "sqs:ListDeadLetterSourceQueues",
      "sqs:DeleteMessageBatch",
      "sqs:ReceiveMessage",
      "sqs:GetQueueAttributes",
      "sqs:ListQueueTags",
      "sqs:SetQueueAttributes",
      "sqs:PurgeQueue",
    ]

    resources = [
      aws_sqs_queue.test.arn,
    ]
  }
}

resource "aws_glue_crawler" "test" {
  depends_on = [
    aws_iam_role_policy_attachment.test-AWSGlueServiceRole,
    aws_iam_role_policy.test_sqs,
  ]

  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q
  role          = aws_iam_role.test.name

  s3_target {
    path = "s3://${aws_s3_bucket.test.bucket}"

    event_queue_arn = aws_sqs_queue.test.arn
  }

  recrawl_policy {
    recrawl_behavior = "CRAWL_EVENT_MODE"
  }
}
`, rName))
}

func testAccCrawlerConfig_catalogTargetDlqEventQueue(rName string) string {
	return acctest.ConfigCompose(testAccCrawlerConfig_base(rName), fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name = %[1]q

  visibility_timeout_seconds = 3600
}

resource "aws_sqs_queue" "test_dlq" {
  name = "%[1]sdlq"

  visibility_timeout_seconds = 3600
}

resource "aws_iam_role_policy" "test_sqs" {
  role = aws_iam_role.test.name

  policy = data.aws_iam_policy_document.role_test_sqs.json
}

data "aws_iam_policy_document" "role_test_sqs" {
  statement {
    effect = "Allow"

    actions = [
      "sqs:DeleteMessage",
      "sqs:GetQueueUrl",
      "sqs:ListDeadLetterSourceQueues",
      "sqs:DeleteMessageBatch",
      "sqs:ReceiveMessage",
      "sqs:GetQueueAttributes",
      "sqs:ListQueueTags",
      "sqs:SetQueueAttributes",
      "sqs:PurgeQueue",
    ]

    resources = [
      aws_sqs_queue.test_dlq.arn,
      aws_sqs_queue.test.arn,
    ]
  }
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_s3_bucket" "default" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_glue_catalog_table" "test" {
  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q
  table_type    = "EXTERNAL_TABLE"

  storage_descriptor {
    location = "s3://${aws_s3_bucket.default.bucket}"
  }
}

resource "aws_lakeformation_permissions" "test" {
  permissions = ["ALL"]
  principal   = aws_iam_role.test.arn

  table {
    database_name = aws_glue_catalog_database.test.name
    name          = aws_glue_catalog_table.test.name
  }
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole, aws_iam_role_policy.test_sqs]

  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q
  role          = aws_iam_role.test.name

  schema_change_policy {
    delete_behavior = "LOG"
  }

  catalog_target {
    database_name       = aws_glue_catalog_database.test.name
    tables              = [aws_glue_catalog_table.test.name]
    event_queue_arn     = aws_sqs_queue.test.arn
    dlq_event_queue_arn = aws_sqs_queue.test_dlq.arn
  }

  configuration = <<EOF
{
  "Version": 1,
  "Grouping": {
    "TableGroupingPolicy": "CombineCompatibleSchemas"
  }
}
EOF
}
`, rName))
}

func testAccCrawlerConfig_s3TargetDlqEventQueue(rName string) string {
	return acctest.ConfigCompose(testAccCrawlerConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_sqs_queue" "test" {
  name = %[1]q

  visibility_timeout_seconds = 3600
}

resource "aws_sqs_queue" "test_dlq" {
  name = "%[1]sdlq"

  visibility_timeout_seconds = 3600
}

resource "aws_iam_role_policy" "test_sqs" {
  role = aws_iam_role.test.name

  policy = data.aws_iam_policy_document.role_test_sqs.json
}

data "aws_iam_policy_document" "role_test_sqs" {
  statement {
    effect = "Allow"

    actions = [
      "sqs:DeleteMessage",
      "sqs:GetQueueUrl",
      "sqs:ListDeadLetterSourceQueues",
      "sqs:DeleteMessageBatch",
      "sqs:ReceiveMessage",
      "sqs:GetQueueAttributes",
      "sqs:ListQueueTags",
      "sqs:SetQueueAttributes",
      "sqs:PurgeQueue",
    ]

    resources = [
      aws_sqs_queue.test_dlq.arn,
      aws_sqs_queue.test.arn,
    ]
  }
}

resource "aws_glue_crawler" "test" {
  depends_on = [
    aws_iam_role_policy_attachment.test-AWSGlueServiceRole,
    aws_iam_role_policy.test_sqs,
  ]

  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q
  role          = aws_iam_role.test.name

  s3_target {
    path = "s3://${aws_s3_bucket.test.bucket}"

    event_queue_arn     = aws_sqs_queue.test.arn
    dlq_event_queue_arn = aws_sqs_queue.test_dlq.arn
  }

  recrawl_policy {
    recrawl_behavior = "CRAWL_EVENT_MODE"
  }
}
`, rName))
}

func testAccCrawlerConfig_s3TargetMultiple(rName, path1, path2 string) string {
	return acctest.ConfigCompose(testAccCrawlerConfig_base(rName), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = "%[1]s-%[2]s"
  force_destroy = true
}

resource "aws_s3_bucket" "test2" {
  bucket        = "%[1]s-%[3]s"
  force_destroy = true
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q
  role          = aws_iam_role.test.name

  s3_target {
    path = "s3://${aws_s3_bucket.test.bucket}"
  }

  s3_target {
    path = "s3://${aws_s3_bucket.test2.bucket}"
  }
}
`, rName, path1, path2))
}

func testAccCrawlerConfig_catalogTarget(rName string, tableCount int) string {
	return acctest.ConfigCompose(testAccCrawlerConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_s3_bucket" "default" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_glue_catalog_table" "test" {
  count = %[2]d

  database_name = aws_glue_catalog_database.test.name
  name          = "%[1]s_table_${count.index}"
  table_type    = "EXTERNAL_TABLE"

  storage_descriptor {
    location = "s3://${aws_s3_bucket.default.bucket}"
  }
}

resource "aws_lakeformation_permissions" "test" {
  count = %[2]d

  permissions = ["ALL"]
  principal   = aws_iam_role.test.arn

  table {
    database_name = aws_glue_catalog_database.test.name
    name          = aws_glue_catalog_table.test[count.index].name
  }
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q
  role          = aws_iam_role.test.name

  schema_change_policy {
    delete_behavior = "LOG"
  }

  catalog_target {
    database_name = aws_glue_catalog_database.test.name
    tables        = flatten([aws_glue_catalog_table.test[*].name])
  }

  configuration = <<EOF
{
  "Version": 1,
  "Grouping": {
    "TableGroupingPolicy": "CombineCompatibleSchemas"
  }
}
EOF
}
`, rName, tableCount))
}

func testAccCrawlerConfig_catalogTargetMultiple(rName string) string {
	return acctest.ConfigCompose(testAccCrawlerConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  count = 2
  name  = "%[1]s_database_${count.index}"
}

resource "aws_glue_catalog_table" "test" {
  count         = 2
  database_name = aws_glue_catalog_database.test[count.index].name
  name          = "%[1]s_table_${count.index}"
  table_type    = "EXTERNAL_TABLE"

  storage_descriptor {
    location = "s3://${aws_s3_bucket.default.bucket}"
  }
}

resource "aws_lakeformation_permissions" "test" {
  count = 2

  permissions = ["ALL"]
  principal   = aws_iam_role.test.arn

  table {
    database_name = aws_glue_catalog_database.test[count.index].name
    name          = aws_glue_catalog_table.test[count.index].name
  }
}

resource "aws_s3_bucket" "default" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test[0].name
  name          = %[1]q
  role          = aws_iam_role.test.name

  schema_change_policy {
    delete_behavior = "LOG"
  }

  catalog_target {
    database_name = aws_glue_catalog_database.test[0].name
    tables        = [aws_glue_catalog_table.test[0].name]
  }

  catalog_target {
    database_name = aws_glue_catalog_database.test[1].name
    tables        = [aws_glue_catalog_table.test[1].name]
  }

  configuration = <<EOF
{
  "Version": 1,
  "Grouping": {
    "TableGroupingPolicy": "CombineCompatibleSchemas"
  }
}
EOF
}
`, rName))
}

func testAccCrawlerConfig_schedule(rName, schedule string) string {
	return acctest.ConfigCompose(testAccCrawlerConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q
  role          = aws_iam_role.test.name
  schedule      = %[2]q

  s3_target {
    path = "s3://bucket-name"
  }
}
`, rName, schedule))
}

func testAccCrawlerConfig_schemaChangePolicy(rName, deleteBehavior, updateBehavior string) string {
	return acctest.ConfigCompose(testAccCrawlerConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q
  role          = aws_iam_role.test.name

  s3_target {
    path = "s3://bucket-name"
  }

  schema_change_policy {
    delete_behavior = %[2]q
    update_behavior = %[3]q
  }
}
`, rName, deleteBehavior, updateBehavior))
}

func testAccCrawlerConfig_tablePrefix(rName, tablePrefix string) string {
	return acctest.ConfigCompose(testAccCrawlerConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q
  role          = aws_iam_role.test.name
  table_prefix  = %[2]q

  s3_target {
    path = "s3://bucket-name"
  }
}
`, rName, tablePrefix))
}

func testAccCrawlerConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccCrawlerConfig_base(rName), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q
  role          = aws_iam_role.test.name
  table_prefix  = %[1]q

  s3_target {
    path = "s3://${aws_s3_bucket.test.bucket}"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccCrawlerConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccCrawlerConfig_base(rName), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q
  role          = aws_iam_role.test.name
  table_prefix  = %[1]q

  s3_target {
    path = "s3://${aws_s3_bucket.test.bucket}"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccCrawlerConfig_securityConfiguration(rName, securityConfiguration string) string {
	return acctest.ConfigCompose(testAccCrawlerConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_security_configuration" "test" {
  name = %[2]q

  encryption_configuration {
    cloudwatch_encryption {
      cloudwatch_encryption_mode = "DISABLED"
    }

    job_bookmarks_encryption {
      job_bookmarks_encryption_mode = "DISABLED"
    }

    s3_encryption {
      s3_encryption_mode = "DISABLED"
    }
  }
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name          = aws_glue_catalog_database.test.name
  name                   = %[1]q
  role                   = aws_iam_role.test.name
  security_configuration = aws_glue_security_configuration.test.name

  s3_target {
    path = "s3://bucket-name"
  }
}
`, rName, securityConfiguration))
}

func testAccCrawlerConfig_mongoDBTarget(rName, connectionUrl, path string) string {
	return acctest.ConfigCompose(testAccCrawlerConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_connection" "test" {
  name            = %[1]q
  connection_type = "MONGODB"

  connection_properties = {
    CONNECTION_URL = %[2]q
    PASSWORD       = "testpassword"
    USERNAME       = "testusername"
  }
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q
  role          = aws_iam_role.test.name

  mongodb_target {
    connection_name = aws_glue_connection.test.name
    path            = %[3]q
  }
}
`, rName, connectionUrl, path))
}

func testAccCrawlerConfig_mongoDBTargetScanAll(rName, connectionUrl string, scan bool) string {
	return acctest.ConfigCompose(testAccCrawlerConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_connection" "test" {
  name            = %[1]q
  connection_type = "MONGODB"

  connection_properties = {
    CONNECTION_URL = %[2]q
    PASSWORD       = "testpassword"
    USERNAME       = "testusername"
  }
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q
  role          = aws_iam_role.test.name

  mongodb_target {
    connection_name = aws_glue_connection.test.name
    path            = "database-name/table-name"
    scan_all        = %[3]t
  }
}
`, rName, connectionUrl, scan))
}

func testAccCrawlerConfig_mongoDBMultiple(rName, connectionUrl, path1, path2 string) string {
	return acctest.ConfigCompose(testAccCrawlerConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_connection" "test" {
  name            = %[1]q
  connection_type = "MONGODB"

  connection_properties = {
    CONNECTION_URL = %[2]q
    PASSWORD       = "testpassword"
    USERNAME       = "testusername"
  }
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q
  role          = aws_iam_role.test.name

  mongodb_target {
    connection_name = aws_glue_connection.test.name
    path            = %[3]q
  }

  mongodb_target {
    connection_name = aws_glue_connection.test.name
    path            = %[4]q
  }
}
`, rName, connectionUrl, path1, path2))
}

func testAccCrawlerConfig_deltaTarget(rName, connectionUrl, tableName, createNativeDeltaTable string) string {
	return acctest.ConfigCompose(testAccCrawlerConfig_base(rName), fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    protocol  = "tcp"
    self      = true
    from_port = 1
    to_port   = 65535
  }
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_connection" "test" {
  connection_type = "NETWORK"
  name            = "%[1]s"

  physical_connection_requirements {
    availability_zone      = aws_subnet.test.availability_zone
    security_group_id_list = [aws_security_group.test.id]
    subnet_id              = aws_subnet.test.id
  }
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q
  role          = aws_iam_role.test.name

  delta_target {
    connection_name           = aws_glue_connection.test.name
    delta_tables              = [%[3]q]
    write_manifest            = false
    create_native_delta_table = %[4]s
  }
}
`, rName, connectionUrl, tableName, createNativeDeltaTable))
}

func testAccCrawlerConfig_hudiTarget(rName, connectionUrl, tableName string, depth int) string {
	return acctest.ConfigCompose(testAccCrawlerConfig_base(rName), acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    from_port = 1
    protocol  = "tcp"
    self      = true
    to_port   = 65535
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_connection" "test" {
  connection_properties = {
    JDBC_ENFORCE_SSL = false
  }

  connection_type = "NETWORK"

  name = %[1]q

  physical_connection_requirements {
    availability_zone      = aws_subnet.test[0].availability_zone
    security_group_id_list = [aws_security_group.test.id]
    subnet_id              = aws_subnet.test[0].id
  }
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q
  role          = aws_iam_role.test.name

  hudi_target {
    connection_name         = aws_glue_connection.test.name
    paths                   = [%[3]q]
    maximum_traversal_depth = %[4]d
  }
}
`, rName, connectionUrl, tableName, depth))
}

func testAccCrawlerConfig_icebergTarget(rName, connectionUrl, tableName string, depth int) string {
	return acctest.ConfigCompose(testAccCrawlerConfig_base(rName), acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    from_port = 1
    protocol  = "tcp"
    self      = true
    to_port   = 65535
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_connection" "test" {
  connection_properties = {
    JDBC_ENFORCE_SSL = false
  }

  connection_type = "NETWORK"

  name = %[1]q

  physical_connection_requirements {
    availability_zone      = aws_subnet.test[0].availability_zone
    security_group_id_list = [aws_security_group.test.id]
    subnet_id              = aws_subnet.test[0].id
  }
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q
  role          = aws_iam_role.test.name

  iceberg_target {
    connection_name         = aws_glue_connection.test.name
    paths                   = [%[3]q]
    maximum_traversal_depth = %[4]d
  }
}
`, rName, connectionUrl, tableName, depth))
}

func testAccCrawlerConfig_lakeformation(rName string, use bool) string {
	return acctest.ConfigCompose(testAccCrawlerConfig_base(rName), fmt.Sprintf(`
resource "aws_lakeformation_resource" "test" {
  arn      = aws_s3_bucket.test.arn
  role_arn = aws_iam_role.test.arn
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  principal   = aws_iam_role.test.arn
  permissions = ["DATA_LOCATION_ACCESS"]

  data_location {
    arn = aws_s3_bucket.test.arn
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q
  role          = aws_iam_role.test.name

  lake_formation_configuration {
    use_lake_formation_credentials = %[2]t
  }

  s3_target {
    path = "s3://${aws_s3_bucket.test.bucket}"
  }
}
`, rName, use))
}

func testAccCrawlerConfig_lineage(rName, lineageConfig string) string {
	return acctest.ConfigCompose(testAccCrawlerConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q
  role          = aws_iam_role.test.name

  lineage_configuration {
    crawler_lineage_settings = %[2]q
  }

  s3_target {
    path = "s3://bucket-name"
  }
}
`, rName, lineageConfig))
}

func testAccCrawlerConfig_recrawlPolicy(rName, policy string) string {
	return acctest.ConfigCompose(testAccCrawlerConfig_base(rName), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q
  role          = aws_iam_role.test.name

  schema_change_policy {
    delete_behavior = "LOG"
    update_behavior = "LOG"
  }

  recrawl_policy {
    recrawl_behavior = %[2]q
  }

  s3_target {
    path = "s3://${aws_s3_bucket.test.bucket}"
  }
}
`, rName, policy))
}

func testAccCrawlerConfig_s3TargetSampleSize(rName string, size int) string {
	return acctest.ConfigCompose(testAccCrawlerConfig_base(rName), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q
  role          = aws_iam_role.test.name

  s3_target {
    sample_size = %[2]d
    path        = "s3://${aws_s3_bucket.test.bucket}"
  }
}
`, rName, size))
}
