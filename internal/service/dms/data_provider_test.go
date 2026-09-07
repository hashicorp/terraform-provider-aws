// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dms_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/databasemigrationservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdms "github.com/hashicorp/terraform-provider-aws/internal/service/dms"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestDataProviderSettingsRoundTrip(t *testing.T) {
	t.Parallel()

	certificateARN := aws.String("arn:aws:dms:us-west-2:123456789012:cert:test")
	accessRoleARN := aws.String("arn:aws:iam::123456789012:role/test")
	testCases := map[string]awstypes.DataProviderSettings{
		"doc_db": &awstypes.DataProviderSettingsMemberDocDbSettings{
			Value: awstypes.DocDbDataProviderSettings{
				CertificateArn: certificateARN,
				DatabaseName:   aws.String("documentdb"),
				Port:           aws.Int32(27017),
				ServerName:     aws.String("docdb.example.com"),
				SslMode:        awstypes.DmsSslModeValueVerifyFull,
			},
		},
		"ibm_db2_luw": &awstypes.DataProviderSettingsMemberIbmDb2LuwSettings{
			Value: awstypes.IbmDb2LuwDataProviderSettings{
				CertificateArn:      certificateARN,
				DatabaseName:        aws.String("db2luw"),
				EncryptionAlgorithm: aws.Int32(1),
				Port:                aws.Int32(50000),
				S3AccessRoleArn:     accessRoleARN,
				S3Path:              aws.String("s3://test/luw"),
				SecurityMechanism:   aws.Int32(9),
				ServerName:          aws.String("db2luw.example.com"),
				SslMode:             awstypes.DmsSslModeValueVerifyCa,
			},
		},
		"ibm_db2_zos": &awstypes.DataProviderSettingsMemberIbmDb2zOsSettings{
			Value: awstypes.IbmDb2zOsDataProviderSettings{
				CertificateArn:  certificateARN,
				DatabaseName:    aws.String("db2zos"),
				Port:            aws.Int32(446),
				S3AccessRoleArn: accessRoleARN,
				S3Path:          aws.String("s3://test/zos"),
				ServerName:      aws.String("db2zos.example.com"),
				SslMode:         awstypes.DmsSslModeValueVerifyCa,
			},
		},
		"maria_db": &awstypes.DataProviderSettingsMemberMariaDbSettings{
			Value: awstypes.MariaDbDataProviderSettings{
				CertificateArn:  certificateARN,
				Port:            aws.Int32(3306),
				S3AccessRoleArn: accessRoleARN,
				S3Path:          aws.String("s3://test/mariadb"),
				ServerName:      aws.String("mariadb.example.com"),
				SslMode:         awstypes.DmsSslModeValueRequire,
			},
		},
		"microsoft_sql_server": &awstypes.DataProviderSettingsMemberMicrosoftSqlServerSettings{
			Value: awstypes.MicrosoftSqlServerDataProviderSettings{
				CertificateArn:  certificateARN,
				DatabaseName:    aws.String("sqlserver"),
				Port:            aws.Int32(1433),
				S3AccessRoleArn: accessRoleARN,
				S3Path:          aws.String("s3://test/sqlserver"),
				ServerName:      aws.String("sqlserver.example.com"),
				SslMode:         awstypes.DmsSslModeValueRequire,
			},
		},
		"mongo_db": &awstypes.DataProviderSettingsMemberMongoDbSettings{
			Value: awstypes.MongoDbDataProviderSettings{
				AuthMechanism:  awstypes.AuthMechanismValueScramSha1,
				AuthSource:     aws.String("admin"),
				AuthType:       awstypes.AuthTypeValuePassword,
				CertificateArn: certificateARN,
				DatabaseName:   aws.String("mongodb"),
				Port:           aws.Int32(27017),
				ServerName:     aws.String("mongodb.example.com"),
				SslMode:        awstypes.DmsSslModeValueRequire,
			},
		},
		"mysql": &awstypes.DataProviderSettingsMemberMySqlSettings{
			Value: awstypes.MySqlDataProviderSettings{
				CertificateArn:  certificateARN,
				Port:            aws.Int32(3306),
				S3AccessRoleArn: accessRoleARN,
				S3Path:          aws.String("s3://test/mysql"),
				ServerName:      aws.String("mysql.example.com"),
				SslMode:         awstypes.DmsSslModeValueVerifyFull,
			},
		},
		"oracle": &awstypes.DataProviderSettingsMemberOracleSettings{
			Value: awstypes.OracleDataProviderSettings{
				AsmServer:                            aws.String("asm.example.com"),
				CertificateArn:                       certificateARN,
				DatabaseName:                         aws.String("oracle"),
				Port:                                 aws.Int32(1521),
				S3AccessRoleArn:                      accessRoleARN,
				S3Path:                               aws.String("s3://test/oracle"),
				SecretsManagerOracleAsmAccessRoleArn: accessRoleARN,
				SecretsManagerOracleAsmSecretId:      aws.String("arn:aws:secretsmanager:us-west-2:123456789012:secret:asm"),
				SecretsManagerSecurityDbEncryptionAccessRoleArn: accessRoleARN,
				SecretsManagerSecurityDbEncryptionSecretId:      aws.String("arn:aws:secretsmanager:us-west-2:123456789012:secret:tde"),
				ServerName: aws.String("oracle.example.com"),
				SslMode:    awstypes.DmsSslModeValueVerifyFull,
			},
		},
		"postgresql": &awstypes.DataProviderSettingsMemberPostgreSqlSettings{
			Value: awstypes.PostgreSqlDataProviderSettings{
				CertificateArn:  certificateARN,
				DatabaseName:    aws.String("postgresql"),
				Port:            aws.Int32(5432),
				S3AccessRoleArn: accessRoleARN,
				S3Path:          aws.String("s3://test/postgresql"),
				ServerName:      aws.String("postgresql.example.com"),
				SslMode:         awstypes.DmsSslModeValueRequire,
			},
		},
		"redshift": &awstypes.DataProviderSettingsMemberRedshiftSettings{
			Value: awstypes.RedshiftDataProviderSettings{
				DatabaseName:    aws.String("redshift"),
				Port:            aws.Int32(5439),
				S3AccessRoleArn: accessRoleARN,
				S3Path:          aws.String("s3://test/redshift"),
				ServerName:      aws.String("redshift.example.com"),
			},
		},
		"sybase_ase": &awstypes.DataProviderSettingsMemberSybaseAseSettings{
			Value: awstypes.SybaseAseDataProviderSettings{
				CertificateArn:  certificateARN,
				DatabaseName:    aws.String("sybase"),
				EncryptPassword: aws.Bool(false),
				Port:            aws.Int32(5000),
				ServerName:      aws.String("sybase.example.com"),
				SslMode:         awstypes.DmsSslModeValueNone,
			},
		},
		"optional_fields_omitted": &awstypes.DataProviderSettingsMemberMySqlSettings{
			Value: awstypes.MySqlDataProviderSettings{
				Port:       aws.Int32(3306),
				ServerName: aws.String("mysql.example.com"),
			},
		},
	}
	for name, settings := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()

			var model struct {
				Settings fwtypes.ListNestedObjectValueOf[tfdms.DataProviderSettingsModel] `tfsdk:"settings"`
			}
			input := awstypes.DataProvider{Settings: settings}
			if diags := flex.Flatten(ctx, input, &model); diags.HasError() {
				t.Fatalf("flattening: %s", diags)
			}
			var expanded databasemigrationservice.CreateDataProviderInput
			if diags := flex.Expand(ctx, model, &expanded); diags.HasError() {
				t.Fatalf("expanding: %s", diags)
			}
			if !reflect.DeepEqual(expanded.Settings, settings) {
				t.Errorf("round trip: got %#v, want %#v", expanded.Settings, settings)
			}

			// Reusing a populated model must clear every former union member.
			for nextName, next := range testCases {
				t.Run("switch_to_"+nextName, func(t *testing.T) {
					t.Parallel()
					previous, diags := model.Settings.ToPtr(t.Context())
					if diags.HasError() {
						t.Fatal(diags)
					}
					if diags := flex.Flatten(t.Context(), next, previous); diags.HasError() {
						t.Fatalf("flattening replacement: %s", diags)
					}
					result, diags := previous.Expand(t.Context())
					if diags.HasError() {
						t.Fatalf("expanding replacement: %s", diags)
					}
					if !reflect.DeepEqual(result, next) {
						t.Errorf("engine switch: got %#v, want %#v", result, next)
					}
				})
			}
		})
	}
}

func TestDataProviderSettingsExpandInvalid(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	testCases := map[string]tfdms.DataProviderSettingsModel{
		"missing": {},
		"empty": {
			MySQLSettings: fwtypes.NewListNestedObjectValueOfEmpty[tfdms.DataProviderMySQLSettingsModel](ctx),
		},
		"unknown": {
			MySQLSettings: fwtypes.NewListNestedObjectValueOfUnknown[tfdms.DataProviderMySQLSettingsModel](ctx),
		},
		"multiple_engines": {
			MySQLSettings:  fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfdms.DataProviderMySQLSettingsModel{}),
			OracleSettings: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfdms.DataProviderOracleSettingsModel{}),
		},
		"multiple_blocks": {
			MySQLSettings: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfdms.DataProviderMySQLSettingsModel{{}, {}}),
		},
		"null_object": {
			MySQLSettings: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfdms.DataProviderMySQLSettingsModel{nil}),
		},
	}
	for name, model := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got, diags := model.Expand(t.Context())
			if !diags.HasError() {
				t.Fatalf("expected an error, got %v", got)
			}
			if got != nil {
				t.Errorf("expected no expanded settings on error, got %v", got)
			}
		})
	}
}

func TestDataProviderSettingsFlattenInvalid(t *testing.T) {
	t.Parallel()
	testCases := map[string]any{
		"nil":            nil,
		"unrelated_type": "mysql",
		"unknown_member": awstypes.UnknownUnionMember{Tag: "future_engine", Value: []byte("{}")},
	}
	for name, value := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			model := tfdms.DataProviderSettingsModel{
				MySQLSettings: fwtypes.NewListNestedObjectValueOfPtrMust(t.Context(), &tfdms.DataProviderMySQLSettingsModel{
					ServerName: types.StringValue("mysql.example.com"),
				}),
			}
			previous := model.MySQLSettings
			if diags := model.Flatten(t.Context(), value); !diags.HasError() {
				t.Fatal("expected an error for unsupported settings")
			}
			if !model.MySQLSettings.Equal(previous) {
				t.Fatal("failed flatten changed existing settings")
			}
		})
	}
}

func TestDataProviderSettingsRequiredBlock(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	testCases := map[string]struct {
		value     fwtypes.ListNestedObjectValueOf[tfdms.DataProviderSettingsModel]
		wantError bool
	}{
		"missing":  {value: fwtypes.NewListNestedObjectValueOfNull[tfdms.DataProviderSettingsModel](ctx), wantError: true},
		"empty":    {value: fwtypes.NewListNestedObjectValueOfEmpty[tfdms.DataProviderSettingsModel](ctx), wantError: true},
		"unknown":  {value: fwtypes.NewListNestedObjectValueOfUnknown[tfdms.DataProviderSettingsModel](ctx)},
		"one":      {value: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfdms.DataProviderSettingsModel{})},
		"multiple": {value: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfdms.DataProviderSettingsModel{{}, {}}), wantError: true},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var response validator.ListResponse
			request := validator.ListRequest{
				Path:           path.Root("settings"),
				PathExpression: path.MatchRoot("settings"),
				ConfigValue:    testCase.value.ListValue,
			}
			for _, v := range tfdms.DataProviderSettingsBlock(t.Context()).Validators {
				v.ValidateList(t.Context(), request, &response)
			}
			if got := response.Diagnostics.HasError(); got != testCase.wantError {
				t.Errorf("HasError() = %t, want %t: %s", got, testCase.wantError, response.Diagnostics)
			}
		})
	}
}

func TestAccDMSDataProvider_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DataProvider
	resourceName := "aws_dms_data_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DMS)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataProviderConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataProviderExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "dms", regexache.MustCompile(`data-provider:.+$`)),
					resource.TestMatchResourceAttr(resourceName, "creation_time", regexache.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?Z$`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "postgres"),
					resource.TestCheckResourceAttr(resourceName, "virtual", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.postgresql_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.postgresql_settings.0.server_name", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.postgresql_settings.0.port", "5432"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.postgresql_settings.0.database_name", "example"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.postgresql_settings.0.ssl_mode", "none"),
					resource.TestCheckResourceAttr(resourceName, names.AttrTags+".%", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrTagsAll+".%", "0"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccDMSDataProvider_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DataProvider
	resourceName := "aws_dms_data_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DMS)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataProviderConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataProviderExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfdms.ResourceDataProvider, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDMSDataProvider_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DataProvider
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_data_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DMS)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataProviderConfig_update(rName, "first description", "example.com", "example", "none", 5432, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataProviderExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "first description"),
					resource.TestCheckResourceAttr(resourceName, "virtual", acctest.CtFalse),
				),
			},
			{
				Config: testAccDataProviderConfig_update(rName+"-updated", "second description", "updated.example.com", "updated", "require", 5433, true),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataProviderExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName+"-updated"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "second description"),
					resource.TestCheckResourceAttr(resourceName, "virtual", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "settings.0.postgresql_settings.0.server_name", "updated.example.com"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.postgresql_settings.0.port", "5433"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.postgresql_settings.0.database_name", "updated"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.postgresql_settings.0.ssl_mode", "require"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccDataProviderConfig_named(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataProviderExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "virtual", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "settings.0.postgresql_settings.0.server_name", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.postgresql_settings.0.port", "5432"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.postgresql_settings.0.database_name", "example"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.postgresql_settings.0.ssl_mode", "none"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccDMSDataProvider_engine(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DataProvider
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_data_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DMS)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataProviderConfig_named(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataProviderExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "postgres"),
				),
			},
			{
				Config: testAccDataProviderConfig_settings(rName, "mysql", "mysql_settings", 3306, `ssl_mode = "none"`),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataProviderExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "mysql"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.mysql_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.mysql_settings.0.server_name", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.mysql_settings.0.port", "3306"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.mysql_settings.0.ssl_mode", "none"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.postgresql_settings.#", "0"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccDMSDataProvider_settings(t *testing.T) {
	testCases := []struct {
		engine       string
		block        string
		port         int
		databaseName bool
		sslMode      bool
	}{
		{engine: "aurora", block: "mysql_settings", port: 3306, sslMode: true},
		{engine: "aurora-postgresql", block: "postgresql_settings", port: 5432, databaseName: true, sslMode: true},
		{engine: "db2", block: "ibm_db2_luw_settings", port: 50000, databaseName: true, sslMode: true},
		{engine: "db2-zos", block: "ibm_db2_zos_settings", port: 50000, databaseName: true, sslMode: true},
		{engine: "docdb", block: "doc_db_settings", port: 27017, databaseName: true, sslMode: true},
		{engine: "mariadb", block: "maria_db_settings", port: 3306, sslMode: true},
		{engine: "mongodb", block: "mongo_db_settings", port: 27017, databaseName: true, sslMode: true},
		{engine: "mysql", block: "mysql_settings", port: 3306, sslMode: true},
		{engine: "oracle", block: "oracle_settings", port: 1521, databaseName: true, sslMode: true},
		{engine: "postgres", block: "postgresql_settings", port: 5432, databaseName: true, sslMode: true},
		{engine: "redshift", block: "redshift_settings", port: 5439, databaseName: true},
		{engine: "sqlserver", block: "microsoft_sql_server_settings", port: 1433, databaseName: true, sslMode: true},
		{engine: "sybase", block: "sybase_ase_settings", port: 5000, databaseName: true, sslMode: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.engine, func(t *testing.T) {
			ctx := acctest.Context(t)
			var v awstypes.DataProvider
			rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
			resourceName := "aws_dms_data_provider.test"
			prefix := "settings.0." + testCase.block
			settings := ""
			checks := []resource.TestCheckFunc{
				testAccCheckDataProviderExists(ctx, t, resourceName, &v),
				resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				resource.TestCheckResourceAttr(resourceName, names.AttrEngine, testCase.engine),
				resource.TestCheckResourceAttr(resourceName, prefix+".#", "1"),
				resource.TestCheckResourceAttr(resourceName, prefix+".0.server_name", "example.com"),
				resource.TestCheckResourceAttr(resourceName, prefix+".0.port", strconv.Itoa(testCase.port)),
			}
			if testCase.databaseName {
				settings += "database_name = \"example\"\n"
				checks = append(checks, resource.TestCheckResourceAttr(resourceName, prefix+".0.database_name", "example"))
			}
			if testCase.sslMode {
				settings += "ssl_mode = \"none\"\n"
				checks = append(checks, resource.TestCheckResourceAttr(resourceName, prefix+".0.ssl_mode", "none"))
			}

			acctest.ParallelTest(ctx, t, resource.TestCase{
				PreCheck: func() {
					acctest.PreCheck(ctx, t)
					acctest.PreCheckPartitionHasService(t, names.DMS)
					testAccPreCheck(ctx, t)
				},
				ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				CheckDestroy:             testAccCheckDataProviderDestroy(ctx, t),
				Steps: []resource.TestStep{
					{
						Config: testAccDataProviderConfig_settings(rName, testCase.engine, testCase.block, testCase.port, settings),
						Check:  resource.ComposeAggregateTestCheckFunc(checks...),
					},
					{
						ResourceName:                         resourceName,
						ImportState:                          true,
						ImportStateVerify:                    true,
						ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
						ImportStateVerifyIdentifierAttribute: names.AttrARN,
					},
				},
			})
		})
	}
}

func TestAccDMSDataProvider_mongoDBSettingsUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DataProvider
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_data_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DMS)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataProviderConfig_settings(rName, "mongodb", "mongo_db_settings", 27017, `
      database_name  = "example"
      auth_source    = "example"
      auth_type      = "password"
      auth_mechanism = "scram_sha_1"
      ssl_mode       = "none"
`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataProviderExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "settings.0.mongo_db_settings.0.database_name", "example"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.mongo_db_settings.0.auth_source", "example"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.mongo_db_settings.0.auth_type", "password"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.mongo_db_settings.0.auth_mechanism", "scram_sha_1"),
				),
			},
			{
				Config: testAccDataProviderConfig_settings(rName, "mongodb", "mongo_db_settings", 27017, `
      auth_source    = "admin"
      auth_type      = "no"
      auth_mechanism = "default"
      ssl_mode       = "none"
`),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataProviderExists(ctx, t, resourceName, &v),
					resource.TestCheckNoResourceAttr(resourceName, "settings.0.mongo_db_settings.0.database_name"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.mongo_db_settings.0.auth_source", "admin"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.mongo_db_settings.0.auth_type", "no"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.mongo_db_settings.0.auth_mechanism", "default"),
					func(_ *terraform.State) error {
						settings, ok := v.Settings.(*awstypes.DataProviderSettingsMemberMongoDbSettings)
						if !ok {
							return fmt.Errorf("unexpected settings type: %T", v.Settings)
						}
						if got := aws.ToString(settings.Value.DatabaseName); got != "" {
							return fmt.Errorf("database name not cleared: %q", got)
						}
						return nil
					},
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func testAccCheckDataProviderDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dms_data_provider" {
				continue
			}

			ctx := conns.NewResourceContext(ctx, "", "", "", rs.Primary.Attributes[names.AttrRegion])
			conn := acctest.ProviderMeta(ctx, t).DMSClient(ctx)
			arn := rs.Primary.Attributes[names.AttrARN]
			_, err := tfdms.FindDataProviderByARN(ctx, conn, arn)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return create.Error(names.DMS, create.ErrActionCheckingDestroyed, "Data Provider", arn, err)
			}

			return create.Error(names.DMS, create.ErrActionCheckingDestroyed, "Data Provider", arn, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckDataProviderExists(ctx context.Context, t *testing.T, name string, v *awstypes.DataProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.DMS, create.ErrActionCheckingExistence, "Data Provider", name, errors.New("not found"))
		}

		arn := rs.Primary.Attributes[names.AttrARN]
		if arn == "" {
			return create.Error(names.DMS, create.ErrActionCheckingExistence, "Data Provider", name, errors.New("arn not set"))
		}

		ctx := conns.NewResourceContext(ctx, "", "", "", rs.Primary.Attributes[names.AttrRegion])
		conn := acctest.ProviderMeta(ctx, t).DMSClient(ctx)
		output, err := tfdms.FindDataProviderByARN(ctx, conn, arn)
		if err != nil {
			return create.Error(names.DMS, create.ErrActionCheckingExistence, "Data Provider", arn, err)
		}

		*v = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	t.Helper()

	conn := acctest.ProviderMeta(ctx, t).DMSClient(ctx)
	var input databasemigrationservice.DescribeDataProvidersInput
	_, err := conn.DescribeDataProviders(ctx, &input)
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccDataProviderConfig_basic() string {
	return `
resource "aws_dms_data_provider" "test" {
  engine = "postgres"

  settings {
    postgresql_settings {
      server_name   = "example.com"
      port          = 5432
      database_name = "example"
      ssl_mode      = "none"
    }
  }
}
`
}

func testAccDataProviderConfig_named(rName string) string {
	return testAccDataProviderConfig_settings(rName, "postgres", "postgresql_settings", 5432, `
      database_name = "example"
      ssl_mode      = "none"
`)
}

func testAccDataProviderConfig_update(rName, description, serverName, databaseName, sslMode string, port int, virtual bool) string {
	return fmt.Sprintf(`
resource "aws_dms_data_provider" "test" {
  name        = %[1]q
  description = %[2]q
  engine      = "postgres"
  virtual     = %[7]t

  settings {
    postgresql_settings {
      server_name   = %[3]q
      port          = %[6]d
      database_name = %[4]q
      ssl_mode      = %[5]q
    }
  }
}
`, rName, description, serverName, databaseName, sslMode, port, virtual)
}

func testAccDataProviderConfig_settings(rName, engine, block string, port int, settings string) string {
	return fmt.Sprintf(`
resource "aws_dms_data_provider" "test" {
  name   = %[1]q
  engine = %[2]q

  settings {
    %[3]s {
      server_name = "example.com"
      port        = %[4]d
      %[5]s
    }
  }
}
`, rName, engine, block, port, settings)
}
