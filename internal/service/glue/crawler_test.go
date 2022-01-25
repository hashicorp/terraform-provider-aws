package glue_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfglue "github.com/hashicorp/terraform-provider-aws/internal/service/glue"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func TestAccGlueCrawler_dynamoDBTarget(t *testing.T) {
	var crawler glue.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_DynamodbTarget(rName, "table1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "classifiers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "configuration", ""),
					resource.TestCheckResourceAttr(resourceName, "database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.0.path", "table1"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.0.scan_all", "true"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "role", rName),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "schedule", ""),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.delete_behavior", "DEPRECATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.update_behavior", "UPDATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "table_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccGlueCrawlerConfig_DynamodbTarget(rName, "table2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "classifiers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "configuration", ""),
					resource.TestCheckResourceAttr(resourceName, "database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.0.path", "table2"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.0.scan_all", "true"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "role", rName),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "schedule", ""),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.delete_behavior", "DEPRECATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.update_behavior", "UPDATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "table_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
	var crawler glue.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_DynamodbTargetScanAll(rName, "table1", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.0.path", "table1"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.0.scan_all", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGlueCrawlerConfig_DynamodbTargetScanAll(rName, "table1", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.0.path", "table1"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.0.scan_all", "true"),
				),
			},
			{
				Config: testAccGlueCrawlerConfig_DynamodbTargetScanAll(rName, "table1", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.0.path", "table1"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.0.scan_all", "false"),
				),
			},
		},
	})
}

func TestAccGlueCrawler_DynamoDBTarget_scanRate(t *testing.T) {
	var crawler glue.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_DynamodbTargetScanRate(rName, "table1", 0.5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.0.path", "table1"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.0.scan_rate", "0.5"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGlueCrawlerConfig_DynamodbTargetScanRate(rName, "table1", 1.5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.0.path", "table1"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.0.scan_rate", "1.5"),
				),
			},
			{
				Config: testAccGlueCrawlerConfig_DynamodbTargetScanRate(rName, "table1", 0.5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.0.path", "table1"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.0.scan_rate", "0.5"),
				),
			},
		},
	})
}

func TestAccGlueCrawler_jdbcTarget(t *testing.T) {
	var crawler glue.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	jdbcConnectionUrl := fmt.Sprintf("jdbc:mysql://%s/testdatabase", acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_JdbcTarget(rName, jdbcConnectionUrl, "database-name/%"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "classifiers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "configuration", ""),
					resource.TestCheckResourceAttr(resourceName, "database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.exclusions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.path", "database-name/%"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "role", rName),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "schedule", ""),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.delete_behavior", "DEPRECATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.update_behavior", "UPDATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "table_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccGlueCrawlerConfig_JdbcTarget(rName, jdbcConnectionUrl, "database-name/table-name"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "classifiers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "configuration", ""),
					resource.TestCheckResourceAttr(resourceName, "database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.exclusions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.path", "database-name/table-name"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "role", rName),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "schedule", ""),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.delete_behavior", "DEPRECATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.update_behavior", "UPDATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "table_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccGlueCrawler_JDBCTarget_exclusions(t *testing.T) {
	var crawler glue.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	jdbcConnectionUrl := fmt.Sprintf("jdbc:mysql://%s/testdatabase", acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_JdbcTarget_Exclusions2(rName, jdbcConnectionUrl, "exclusion1", "exclusion2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.exclusions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.exclusions.0", "exclusion1"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.exclusions.1", "exclusion2"),
				),
			},
			{
				Config: testAccGlueCrawlerConfig_JdbcTarget_Exclusions1(rName, jdbcConnectionUrl, "exclusion1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.exclusions.#", "1"),
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
	var crawler glue.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	jdbcConnectionUrl := fmt.Sprintf("jdbc:mysql://%s/testdatabase", acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_JdbcTarget_Multiple(rName, jdbcConnectionUrl, "database-name/table1", "database-name/table2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.exclusions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.path", "database-name/table1"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.1.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.1.exclusions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.1.path", "database-name/table2"),
				),
			},
			{
				Config: testAccGlueCrawlerConfig_JdbcTarget(rName, jdbcConnectionUrl, "database-name/table1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.exclusions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.path", "database-name/table1"),
				),
			},
			{
				Config: testAccGlueCrawlerConfig_JdbcTarget_Multiple(rName, jdbcConnectionUrl, "database-name/table1", "database-name/table2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.#", "2"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.exclusions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.path", "database-name/table1"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.1.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.1.exclusions.#", "0"),
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
	var crawler glue.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	connectionUrl := fmt.Sprintf("mongodb://%s:27017/testdatabase", acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfigMongoDBTarget(rName, connectionUrl, "database-name/%"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.scan_all", "true"),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.path", "database-name/%"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGlueCrawlerConfigMongoDBTarget(rName, connectionUrl, "database-name/table-name"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.scan_all", "true"),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.path", "database-name/table-name"),
				),
			},
		},
	})
}

func TestAccGlueCrawler_MongoDBTargetScan_all(t *testing.T) {
	var crawler glue.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	connectionUrl := fmt.Sprintf("mongodb://%s:27017/testdatabase", acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfigMongoDBTargetScanAll(rName, connectionUrl, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.scan_all", "false"),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.path", "database-name/table-name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGlueCrawlerConfigMongoDBTargetScanAll(rName, connectionUrl, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.scan_all", "true"),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.path", "database-name/table-name"),
				),
			},
			{
				Config: testAccGlueCrawlerConfigMongoDBTargetScanAll(rName, connectionUrl, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.scan_all", "false"),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.path", "database-name/table-name"),
				),
			},
		},
	})
}

func TestAccGlueCrawler_MongoDBTarget_multiple(t *testing.T) {
	var crawler glue.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	connectionUrl := fmt.Sprintf("mongodb://%s:27017/testdatabase", acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfigMongoDBMultiple(rName, connectionUrl, "database-name/table1", "database-name/table2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.scan_all", "true"),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.path", "database-name/table1"),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.1.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.1.scan_all", "true"),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.1.path", "database-name/table2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGlueCrawlerConfigMongoDBTarget(rName, connectionUrl, "database-name/%"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.scan_all", "true"),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.path", "database-name/%"),
				),
			},
			{
				Config: testAccGlueCrawlerConfigMongoDBMultiple(rName, connectionUrl, "database-name/table1", "database-name/table2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.scan_all", "true"),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.0.path", "database-name/table1"),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.1.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.1.scan_all", "true"),
					resource.TestCheckResourceAttr(resourceName, "mongodb_target.1.path", "database-name/table2"),
				),
			},
		},
	})
}

func TestAccGlueCrawler_deltaTarget(t *testing.T) {
	var crawler glue.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	connectionUrl := fmt.Sprintf("mongodb://%s:27017/testdatabase", acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfigDeltaTarget(rName, connectionUrl, "s3://table1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "delta_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "delta_target.0.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "delta_target.0.delta_tables.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "delta_target.0.delta_tables.*", "s3://table1"),
					resource.TestCheckResourceAttr(resourceName, "delta_target.0.write_manifest", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGlueCrawlerConfigDeltaTarget(rName, connectionUrl, "s3://table2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "delta_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "delta_target.0.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "delta_target.0.delta_tables.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "delta_target.0.delta_tables.*", "s3://table2"),
					resource.TestCheckResourceAttr(resourceName, "delta_target.0.write_manifest", "false"),
				),
			},
		},
	})
}

func TestAccGlueCrawler_s3Target(t *testing.T) {
	var crawler glue.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_S3Target(rName, "s3://bucket1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "classifiers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "configuration", ""),
					resource.TestCheckResourceAttr(resourceName, "database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "role", rName),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.exclusions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.path", "s3://bucket1"),
					resource.TestCheckResourceAttr(resourceName, "schedule", ""),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.delete_behavior", "DEPRECATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.update_behavior", "UPDATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "table_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccGlueCrawlerConfig_S3Target(rName, "s3://bucket2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "classifiers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "configuration", ""),
					resource.TestCheckResourceAttr(resourceName, "database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "role", rName),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.exclusions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.path", "s3://bucket2"),
					resource.TestCheckResourceAttr(resourceName, "schedule", ""),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.delete_behavior", "DEPRECATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.update_behavior", "UPDATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "table_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
	var crawler glue.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"
	connectionName := "aws_glue_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_S3Target_ConnectionName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "s3_target.0.connection_name", connectionName, "name"),
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
	var crawler glue.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_S3TargetSampleSize(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.sample_size", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGlueCrawlerConfig_S3TargetSampleSize(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.sample_size", "2"),
				),
			},
		},
	})
}

func TestAccGlueCrawler_S3Target_exclusions(t *testing.T) {
	var crawler glue.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_S3Target_Exclusions2(rName, "exclusion1", "exclusion2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.exclusions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.exclusions.0", "exclusion1"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.exclusions.1", "exclusion2"),
				),
			},
			{
				Config: testAccGlueCrawlerConfig_S3Target_Exclusions1(rName, "exclusion1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.exclusions.#", "1"),
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
	var crawler glue.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_S3Target_EventQueue(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", "1"),
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

func TestAccGlueCrawler_S3Target_dlqeventqueue(t *testing.T) {
	var crawler glue.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_S3Target_DlqEventQueue(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", "1"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "s3_target.0.event_queue_arn", "sqs", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "s3_target.0.dlq_event_queue_arn", "sqs", fmt.Sprintf("%sdlq", rName)),
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
	var crawler glue.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_S3Target_Multiple(rName, "s3://bucket1", "s3://bucket2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.exclusions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.path", "s3://bucket1"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.1.exclusions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.1.path", "s3://bucket2"),
				),
			},
			{
				Config: testAccGlueCrawlerConfig_S3Target(rName, "s3://bucket1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.exclusions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.path", "s3://bucket1"),
				),
			},
			{
				Config: testAccGlueCrawlerConfig_S3Target_Multiple(rName, "s3://bucket1", "s3://bucket2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.exclusions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.path", "s3://bucket1"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.1.exclusions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.1.path", "s3://bucket2"),
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
	var crawler glue.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_CatalogTarget(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "classifiers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "role", rName),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.0.database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.0.tables.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.0.tables.0", fmt.Sprintf("%s_table_0", rName)),
					resource.TestCheckResourceAttr(resourceName, "schedule", ""),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.delete_behavior", "LOG"),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.update_behavior", "UPDATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "table_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "configuration", "{\"Version\":1.0,\"Grouping\":{\"TableGroupingPolicy\":\"CombineCompatibleSchemas\"}}"),
				),
			},
			{
				Config: testAccGlueCrawlerConfig_CatalogTarget(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "classifiers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "role", rName),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.0.database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.0.tables.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.0.tables.0", fmt.Sprintf("%s_table_0", rName)),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.0.tables.1", fmt.Sprintf("%s_table_1", rName)),
					resource.TestCheckResourceAttr(resourceName, "schedule", ""),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.delete_behavior", "LOG"),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.update_behavior", "UPDATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "table_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "configuration", "{\"Version\":1.0,\"Grouping\":{\"TableGroupingPolicy\":\"CombineCompatibleSchemas\"}}"),
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
	var crawler glue.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_CatalogTarget(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.0.database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.0.tables.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.0.tables.0", fmt.Sprintf("%s_table_0", rName)),
				),
			},
			{
				Config: testAccGlueCrawlerConfig_CatalogTarget_Multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.0.database_name", fmt.Sprintf("%s_database_0", rName)),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.1.database_name", fmt.Sprintf("%s_database_1", rName)),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.0.tables.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.0.tables.0", fmt.Sprintf("%s_table_0", rName)),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.1.tables.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.1.tables.0", fmt.Sprintf("%s_table_1", rName)),
				),
			},
			{
				Config: testAccGlueCrawlerConfig_CatalogTarget(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("crawler/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.0.database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "catalog_target.0.tables.#", "1"),
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
	var crawler glue.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_S3Target(rName, "s3://bucket1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					acctest.CheckResourceDisappears(acctest.Provider, tfglue.ResourceCrawler(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccGlueCrawler_classifiers(t *testing.T) {
	var crawler glue.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_Classifiers_Single(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "classifiers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "classifiers.0", rName+"1"),
				),
			},
			{
				Config: testAccGlueCrawlerConfig_Classifiers_Multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "classifiers.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "classifiers.0", rName+"1"),
					resource.TestCheckResourceAttr(resourceName, "classifiers.1", rName+"2"),
				),
			},
			{
				Config: testAccGlueCrawlerConfig_Classifiers_Single(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "classifiers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "classifiers.0", rName+"1"),
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
	var crawler glue.Crawler
	configuration1 := `{"Version": 1.0, "CrawlerOutput": {"Tables": { "AddOrUpdateBehavior": "MergeNewColumns" }}}`
	configuration2 := `{"Version": 1.0, "CrawlerOutput": {"Partitions": { "AddOrUpdateBehavior": "InheritFromTable" }}}`
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_Configuration(rName, configuration1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					testAccCheckCrawlerConfiguration(&crawler, configuration1),
				),
			},
			{
				Config: testAccGlueCrawlerConfig_Configuration(rName, configuration2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					testAccCheckCrawlerConfiguration(&crawler, configuration2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGlueCrawlerConfig_Configuration(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "configuration", ""),
				),
			},
		},
	})
}

func TestAccGlueCrawler_description(t *testing.T) {
	var crawler glue.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_Description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				Config: testAccGlueCrawlerConfig_Description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
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
	var crawler glue.Crawler
	iamRoleResourceName := "aws_iam_role.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_Role_ARN_NoPath(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttrPair(resourceName, "role", iamRoleResourceName, "name"),
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
	var crawler glue.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_Role_ARN_Path(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "role", fmt.Sprintf("path/%s", rName)),
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
	var crawler glue.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_Role_Name_Path(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "role", fmt.Sprintf("path/%s", rName)),
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
	var crawler glue.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_Schedule(rName, "cron(0 1 * * ? *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "schedule", "cron(0 1 * * ? *)"),
				),
			},
			{
				Config: testAccGlueCrawlerConfig_Schedule(rName, "cron(0 2 * * ? *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "schedule", "cron(0 2 * * ? *)"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGlueCrawlerConfig_S3Target(rName, "s3://bucket-name"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "schedule", ""),
				),
			},
		},
	})
}

func TestAccGlueCrawler_schemaChangePolicy(t *testing.T) {
	var crawler glue.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_SchemaChangePolicy(rName, glue.DeleteBehaviorDeleteFromDatabase, glue.UpdateBehaviorUpdateInDatabase),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.delete_behavior", glue.DeleteBehaviorDeleteFromDatabase),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.update_behavior", glue.UpdateBehaviorUpdateInDatabase),
				),
			},
			{
				Config: testAccGlueCrawlerConfig_SchemaChangePolicy(rName, glue.DeleteBehaviorLog, glue.UpdateBehaviorLog),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.delete_behavior", glue.DeleteBehaviorLog),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.update_behavior", glue.UpdateBehaviorLog),
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
	var crawler glue.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_TablePrefix(rName, "prefix1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "table_prefix", "prefix1"),
				),
			},
			{
				Config: testAccGlueCrawlerConfig_TablePrefix(rName, "prefix2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
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
	var crawler glue.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_TablePrefix(rName, "prefix"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "table_prefix", "prefix"),
				),
			},
			{
				Config: testAccGlueCrawlerConfig_TablePrefix(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
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
	var crawler1, crawler2, crawler3 glue.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler1),
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
				Config: testAccGlueCrawlerConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccGlueCrawlerConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccGlueCrawler_security(t *testing.T) {
	var crawler glue.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_SecurityConfiguration(rName, "security_configuration1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "security_configuration", "security_configuration1"),
				),
			},
			{
				Config: testAccGlueCrawlerConfig_SecurityConfiguration(rName, "security_configuration2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
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
	var crawler glue.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerLineageConfig(rName, "ENABLE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "lineage_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "lineage_configuration.0.crawler_lineage_settings", "ENABLE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGlueCrawlerLineageConfig(rName, "DISABLE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "lineage_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "lineage_configuration.0.crawler_lineage_settings", "DISABLE")),
			},
			{
				Config: testAccGlueCrawlerLineageConfig(rName, "ENABLE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "lineage_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "lineage_configuration.0.crawler_lineage_settings", "ENABLE"),
				),
			},
		},
	})
}

func TestAccGlueCrawler_reCrawlPolicy(t *testing.T) {
	var crawler glue.Crawler
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerRecrawlPolicyConfig(rName, "CRAWL_EVERYTHING"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "recrawl_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "recrawl_policy.0.recrawl_behavior", "CRAWL_EVERYTHING"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGlueCrawlerRecrawlPolicyConfig(rName, "CRAWL_NEW_FOLDERS_ONLY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "recrawl_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "recrawl_policy.0.recrawl_behavior", "CRAWL_NEW_FOLDERS_ONLY")),
			},
			{
				Config: testAccGlueCrawlerRecrawlPolicyConfig(rName, "CRAWL_EVERYTHING"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "recrawl_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "recrawl_policy.0.recrawl_behavior", "CRAWL_EVERYTHING"),
				),
			},
		},
	})
}

func testAccCheckCrawlerExists(resourceName string, crawler *glue.Crawler) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		glueConn := acctest.Provider.Meta().(*conns.AWSClient).GlueConn
		out, err := glueConn.GetCrawler(&glue.GetCrawlerInput{
			Name: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if out.Crawler == nil {
			return fmt.Errorf("no Glue Crawler found")
		}

		*crawler = *out.Crawler

		return nil
	}
}

func testAccCheckCrawlerDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_glue_crawler" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueConn
		output, err := conn.GetCrawler(&glue.GetCrawlerInput{
			Name: aws.String(rs.Primary.ID),
		})

		if err != nil {
			if tfawserr.ErrMessageContains(err, glue.ErrCodeEntityNotFoundException, "") {
				return nil
			}
			return err
		}

		crawler := output.Crawler
		if crawler != nil && aws.StringValue(crawler.Name) == rs.Primary.ID {
			return fmt.Errorf("Glue Crawler %s still exists", rs.Primary.ID)
		}

		return nil
	}

	return nil
}

func testAccCheckCrawlerConfiguration(crawler *glue.Crawler, acctestJSON string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiJSON := aws.StringValue(crawler.Configuration)
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

func testAccGlueCrawlerConfig_Base(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name               = %q
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

func testAccGlueCrawlerConfig_Classifiers_Single(rName string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %q
}

resource "aws_glue_classifier" "test1" {
  name = %q

  grok_classifier {
    classification = "example"
    grok_pattern   = "example"
  }
}

resource "aws_glue_classifier" "test2" {
  name = %q

  grok_classifier {
    classification = "example"
    grok_pattern   = "example"
  }
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  classifiers   = [aws_glue_classifier.test1.id]
  name          = %q
  database_name = aws_glue_catalog_database.test.name
  role          = aws_iam_role.test.name

  s3_target {
    path = "s3://bucket-name"
  }
}
`, rName, rName+"1", rName+"2", rName)
}

func testAccGlueCrawlerConfig_Classifiers_Multiple(rName string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %q
}

resource "aws_glue_classifier" "test1" {
  name = %q

  grok_classifier {
    classification = "example"
    grok_pattern   = "example"
  }
}

resource "aws_glue_classifier" "test2" {
  name = %q

  grok_classifier {
    classification = "example"
    grok_pattern   = "example"
  }
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  classifiers   = [aws_glue_classifier.test1.id, aws_glue_classifier.test2.id]
  name          = %q
  database_name = aws_glue_catalog_database.test.name
  role          = aws_iam_role.test.name

  s3_target {
    path = "s3://bucket-name"
  }
}
`, rName, rName+"1", rName+"2", rName)
}

func testAccGlueCrawlerConfig_Configuration(rName, configuration string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %q
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  configuration = %s
  database_name = aws_glue_catalog_database.test.name
  name          = %q
  role          = aws_iam_role.test.name

  s3_target {
    path = "s3://bucket-name"
  }
}
`, rName, strconv.Quote(configuration), rName)
}

func testAccGlueCrawlerConfig_Description(rName, description string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %q
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  description   = %q
  name          = %q
  role          = aws_iam_role.test.name

  s3_target {
    path = "s3://bucket-name"
  }
}
`, rName, description, rName)
}

func testAccGlueCrawlerConfig_DynamodbTarget(rName, path string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q
  role          = aws_iam_role.test.name

  dynamodb_target {
    path = %[2]q
  }
}
`, rName, path)
}

func testAccGlueCrawlerConfig_DynamodbTargetScanAll(rName, path string, scanAll bool) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q
  role          = aws_iam_role.test.name

  dynamodb_target {
    path     = %[2]q
    scan_all = %[3]t
  }
}
`, rName, path, scanAll)
}

func testAccGlueCrawlerConfig_DynamodbTargetScanRate(rName, path string, scanRate float64) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q
  role          = aws_iam_role.test.name

  dynamodb_target {
    path      = %[2]q
    scan_rate = %[3]g
  }
}
`, rName, path, scanRate)
}

func testAccGlueCrawlerConfig_JdbcTarget(rName, jdbcConnectionUrl, path string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
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
    connection_name = aws_glue_connection.test.name
    path            = %[3]q
  }
}
`, rName, jdbcConnectionUrl, path)
}

func testAccGlueCrawlerConfig_JdbcTarget_Exclusions1(rName, jdbcConnectionUrl, exclusion1 string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
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
`, rName, jdbcConnectionUrl, exclusion1)
}

func testAccGlueCrawlerConfig_JdbcTarget_Exclusions2(rName, jdbcConnectionUrl, exclusion1, exclusion2 string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
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
`, rName, jdbcConnectionUrl, exclusion1, exclusion2)
}

func testAccGlueCrawlerConfig_JdbcTarget_Multiple(rName, jdbcConnectionUrl, path1, path2 string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
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
`, rName, jdbcConnectionUrl, path1, path2)
}

func testAccGlueCrawlerConfig_Role_ARN_NoPath(rName string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
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

func testAccGlueCrawlerConfig_Role_ARN_Path(rName string) string {
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

func testAccGlueCrawlerConfig_Role_Name_Path(rName string) string {
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
  role          = "${replace(aws_iam_role.test.path, "/^\\//", "")}${aws_iam_role.test.name}"

  s3_target {
    path = "s3://bucket-name"
  }
}
`, rName)
}

func testAccGlueCrawlerConfig_S3Target(rName, path string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q
  role          = aws_iam_role.test.name

  s3_target {
    path = %[2]q
  }
}
`, rName, path)
}

func testAccGlueCrawlerConfig_S3Target_Exclusions1(rName, exclusion1 string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
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
    path       = "s3://bucket1"
  }
}
`, rName, exclusion1)
}

func testAccGlueCrawlerConfig_S3Target_ConnectionName(rName string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
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
    Name = "terraform-testacc-glue-connection-base"
  }
}

resource "aws_security_group" "test" {
  name   = "%[1]s"
  vpc_id = aws_vpc.test.id

  ingress {
    from_port = 1
    protocol  = "tcp"
    self      = true
    to_port   = 65535
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "terraform-testacc-glue-connection-base"
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

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  name          = "%[1]s"
  role          = aws_iam_role.test.name

  s3_target {
    connection_name = aws_glue_connection.test.name
    path            = "s3://bucket1"
  }
}
`, rName)
}

func testAccGlueCrawlerConfig_S3Target_Exclusions2(rName, exclusion1, exclusion2 string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
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
    path       = "s3://bucket1"
  }
}
`, rName, exclusion1, exclusion2)
}

func testAccGlueCrawlerConfig_S3Target_EventQueue(rName string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
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
`, rName)
}

func testAccGlueCrawlerConfig_S3Target_DlqEventQueue(rName string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
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
`, rName)
}

func testAccGlueCrawlerConfig_S3Target_Multiple(rName, path1, path2 string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_crawler" "test" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q
  role          = aws_iam_role.test.name

  s3_target {
    path = %[2]q
  }

  s3_target {
    path = %[3]q
  }
}
`, rName, path1, path2)
}

func testAccGlueCrawlerConfig_CatalogTarget(rName string, tableCount int) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
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
`, rName, tableCount)
}

func testAccGlueCrawlerConfig_CatalogTarget_Multiple(rName string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
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
`, rName)
}

func testAccGlueCrawlerConfig_Schedule(rName, schedule string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
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
`, rName, schedule)
}

func testAccGlueCrawlerConfig_SchemaChangePolicy(rName, deleteBehavior, updateBehavior string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
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
`, rName, deleteBehavior, updateBehavior)
}

func testAccGlueCrawlerConfig_TablePrefix(rName, tablePrefix string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
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
`, rName, tablePrefix)
}

func testAccGlueCrawlerConfigTags1(rName, tagKey1, tagValue1 string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
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
    path = "s3://bucket-name"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccGlueCrawlerConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
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
    path = "s3://bucket-name"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccGlueCrawlerConfig_SecurityConfiguration(rName, securityConfiguration string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
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
`, rName, securityConfiguration)
}

func testAccGlueCrawlerConfigMongoDBTarget(rName, connectionUrl, path string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
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
`, rName, connectionUrl, path)
}

func testAccGlueCrawlerConfigMongoDBTargetScanAll(rName, connectionUrl string, scan bool) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
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
`, rName, connectionUrl, scan)
}

func testAccGlueCrawlerConfigMongoDBMultiple(rName, connectionUrl, path1, path2 string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
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
`, rName, connectionUrl, path1, path2)
}

func testAccGlueCrawlerConfigDeltaTarget(rName, connectionUrl, tableName string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
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

  delta_target {
    connection_name = aws_glue_connection.test.name
    delta_tables    = [%[3]q]
    write_manifest  = false
  }
}
`, rName, connectionUrl, tableName)
}

func testAccGlueCrawlerLineageConfig(rName, lineageConfig string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
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
`, rName, lineageConfig)
}

func testAccGlueCrawlerRecrawlPolicyConfig(rName, policy string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
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
    path = "s3://bucket-name"
  }
}
`, rName, policy)
}

func testAccGlueCrawlerConfig_S3TargetSampleSize(rName string, size int) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
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
    path        = "s3://bucket1"
  }
}
`, rName, size)
}
