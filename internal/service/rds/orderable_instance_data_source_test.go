package rds_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccRDSOrderableInstanceDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_rds_orderable_db_instance.test"
	engine := "mysql"
	licenseModel := "general-public-license"
	storageType := "standard"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccOrderableInstancePreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccOrderableInstanceDataSourceConfig_basic(engine, licenseModel, storageType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "engine", engine),
					resource.TestCheckResourceAttr(dataSourceName, "license_model", licenseModel),
					resource.TestCheckResourceAttr(dataSourceName, "storage_type", storageType),
					resource.TestCheckResourceAttrPair(dataSourceName, "engine_version", "data.aws_rds_engine_version.default", "version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "instance_class", "data.aws_rds_orderable_db_instance.dynamic", "instance_class"),
				),
			},
		},
	})
}

func TestAccRDSOrderableInstanceDataSource_preferredClass(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	dataSourceName := "data.aws_rds_orderable_db_instance.test"
	preferredClass := "db.t3.micro"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccOrderableInstancePreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccOrderableInstanceDataSourceConfig_preferredClass(preferredClass),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "instance_class", preferredClass),
				),
			},
		},
	})
}

func TestAccRDSOrderableInstanceDataSource_preferredVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	dataSourceName := "data.aws_rds_orderable_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccOrderableInstancePreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccOrderableInstanceDataSourceConfig_preferredVersion(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "engine_version", "data.aws_rds_engine_version.default", "version"),
				),
			},
		},
	})
}

func TestAccRDSOrderableInstanceDataSource_preferredClassAndVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	dataSourceName := "data.aws_rds_orderable_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccOrderableInstancePreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccOrderableInstanceDataSourceConfig_preferredClassAndVersion(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "instance_class", "data.aws_rds_orderable_db_instance.dynamic", "instance_class"),
					resource.TestCheckResourceAttrPair(dataSourceName, "engine_version", "data.aws_rds_engine_version.default", "version"),
				),
			},
		},
	})
}

func TestAccRDSOrderableInstanceDataSource_supportsEnhancedMonitoring(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	dataSourceName := "data.aws_rds_orderable_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccOrderableInstancePreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccOrderableInstanceDataSourceConfig_supportsEnhancedMonitoring(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "supports_enhanced_monitoring", "true"),
				),
			},
		},
	})
}

func TestAccRDSOrderableInstanceDataSource_supportsIAMDatabaseAuthentication(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	dataSourceName := "data.aws_rds_orderable_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccOrderableInstancePreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccOrderableInstanceDataSourceConfig_supportsIAMDatabaseAuthentication(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "supports_iam_database_authentication", "true"),
				),
			},
		},
	})
}

func TestAccRDSOrderableInstanceDataSource_supportsIops(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	dataSourceName := "data.aws_rds_orderable_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccOrderableInstancePreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccOrderableInstanceDataSourceConfig_supportsIops(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "supports_iops", "true"),
				),
			},
		},
	})
}

func TestAccRDSOrderableInstanceDataSource_supportsKerberosAuthentication(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	dataSourceName := "data.aws_rds_orderable_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccOrderableInstancePreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccOrderableInstanceDataSourceConfig_supportsKerberosAuthentication(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "supports_kerberos_authentication", "true"),
				),
			},
		},
	})
}

func TestAccRDSOrderableInstanceDataSource_supportsPerformanceInsights(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	dataSourceName := "data.aws_rds_orderable_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccOrderableInstancePreCheck(t)
			testAccPerformanceInsightsDefaultVersionPreCheck(t, "mysql")
		},
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccOrderableInstanceDataSourceConfig_supportsPerformanceInsights(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "supports_performance_insights", "true"),
				),
			},
		},
	})
}

func TestAccRDSOrderableInstanceDataSource_supportsStorageAutoScaling(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	dataSourceName := "data.aws_rds_orderable_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccOrderableInstancePreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccOrderableInstanceDataSourceConfig_supportsStorageAutoScaling(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "supports_storage_autoscaling", "true"),
				),
			},
		},
	})
}

func TestAccRDSOrderableInstanceDataSource_supportsStorageEncryption(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	dataSourceName := "data.aws_rds_orderable_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccOrderableInstancePreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccOrderableInstanceDataSourceConfig_supportsStorageEncryption(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "supports_storage_encryption", "true"),
				),
			},
		},
	})
}

func testAccOrderableInstancePreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

	input := &rds.DescribeOrderableDBInstanceOptionsInput{
		Engine:          aws.String("mysql"),
		EngineVersion:   aws.String("8.0.20"),
		DBInstanceClass: aws.String("db.m5.xlarge"),
	}

	_, err := conn.DescribeOrderableDBInstanceOptions(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccOrderableInstanceDataSourceConfig_basic(engine, license, storage string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = %[1]q
}

data "aws_rds_orderable_db_instance" "dynamic" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  preferred_instance_classes = [%[2]s]
}

data "aws_rds_orderable_db_instance" "test" {
  instance_class = data.aws_rds_orderable_db_instance.dynamic.instance_class
  engine         = data.aws_rds_engine_version.default.engine
  engine_version = data.aws_rds_engine_version.default.version
  license_model  = %[3]q
  storage_type   = %[4]q
}
`, engine, mySQLPreferredInstanceClasses, license, storage)
}

func testAccOrderableInstanceDataSourceConfig_preferredClass(preferredClass string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine         = data.aws_rds_engine_version.default.engine
  engine_version = data.aws_rds_engine_version.default.version
  license_model  = "general-public-license"

  preferred_instance_classes = [
    "db.xyz.xlarge",
    %[1]q,
    "db.t3.small",
  ]
}
`, preferredClass)
}

func testAccOrderableInstanceDataSourceConfig_preferredVersion() string {
	return `
data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine        = data.aws_rds_engine_version.default.engine
  license_model = "general-public-license"
  storage_type  = "standard"

  preferred_engine_versions = [
    "18.42.32",
    data.aws_rds_engine_version.default.version,
    "not.a.version",
  ]
}
`
}

func testAccOrderableInstanceDataSourceConfig_preferredClassAndVersion() string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "dynamic" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  preferred_instance_classes = [%[1]s]
}

data "aws_rds_orderable_db_instance" "test" {
  engine        = data.aws_rds_engine_version.default.engine
  license_model = "general-public-license"

  preferred_instance_classes = [
    "db.xyz.xlarge",
    data.aws_rds_orderable_db_instance.dynamic.instance_class,
    "db.t3.small",
  ]
  preferred_engine_versions = [
    "18.42.32",
    data.aws_rds_engine_version.default.version,
    "not.a.version",
  ]
}
`, mySQLPreferredInstanceClasses)
}

func testAccOrderableInstanceDataSourceConfig_supportsEnhancedMonitoring() string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                       = data.aws_rds_engine_version.default.engine
  license_model                = "general-public-license"
  storage_type                 = "standard"
  supports_enhanced_monitoring = true

  preferred_engine_versions  = ["8.0.25", "8.0.26", data.aws_rds_engine_version.default.version]
  preferred_instance_classes = [%[1]s]
}
`, mySQLPreferredInstanceClasses)
}

func testAccOrderableInstanceDataSourceConfig_supportsIAMDatabaseAuthentication() string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                               = data.aws_rds_engine_version.default.engine
  license_model                        = "general-public-license"
  storage_type                         = "standard"
  supports_iam_database_authentication = true

  preferred_engine_versions  = ["8.0.25", "8.0.26", data.aws_rds_engine_version.default.version]
  preferred_instance_classes = [%[1]s]
}
`, mySQLPreferredInstanceClasses)
}

func testAccOrderableInstanceDataSourceConfig_supportsIops() string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine        = data.aws_rds_engine_version.default.engine
  license_model = "general-public-license"
  supports_iops = true

  preferred_engine_versions  = ["8.0.20", "8.0.19", data.aws_rds_engine_version.default.version]
  preferred_instance_classes = [%[1]s]
}
`, mySQLPreferredInstanceClasses)
}

func testAccOrderableInstanceDataSourceConfig_supportsKerberosAuthentication() string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "postgres"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                           = data.aws_rds_engine_version.default.engine
  license_model                    = "postgresql-license"
  storage_type                     = "standard"
  supports_kerberos_authentication = true

  preferred_engine_versions  = ["14.1", "13.5", data.aws_rds_engine_version.default.version]
  preferred_instance_classes = [%[1]s]
}
`, postgresPreferredInstanceClasses)
}

func testAccOrderableInstanceDataSourceConfig_supportsPerformanceInsights() string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                        = data.aws_rds_engine_version.default.engine
  license_model                 = "general-public-license"
  supports_performance_insights = true

  preferred_engine_versions  = ["8.0.25", "8.0.26", data.aws_rds_engine_version.default.version]
  preferred_instance_classes = [%[1]s]
}
`, mySQLPreferredInstanceClasses)
}

func testAccOrderableInstanceDataSourceConfig_supportsStorageAutoScaling() string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                       = data.aws_rds_engine_version.default.engine
  license_model                = "general-public-license"
  supports_storage_autoscaling = true

  preferred_engine_versions  = ["8.0.20", "8.0.19", data.aws_rds_engine_version.default.version]
  preferred_instance_classes = [%[1]s]
}
`, mySQLPreferredInstanceClasses)
}

func testAccOrderableInstanceDataSourceConfig_supportsStorageEncryption() string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                      = data.aws_rds_engine_version.default.engine
  license_model               = "general-public-license"
  storage_type                = "standard"
  supports_storage_encryption = true

  preferred_engine_versions  = ["8.0.25", "8.0.26", data.aws_rds_engine_version.default.version]
  preferred_instance_classes = [%[1]s]
}
`, mySQLPreferredInstanceClasses)
}
