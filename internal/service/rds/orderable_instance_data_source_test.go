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
	class := "db.t2.small"
	engine := "mysql"
	engineVersion := "5.7.22"
	licenseModel := "general-public-license"
	storageType := "standard"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccOrderableInstancePreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccOrderableInstanceDataSourceConfig_basic(class, engine, engineVersion, licenseModel, storageType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "instance_class", class),
					resource.TestCheckResourceAttr(dataSourceName, "engine", engine),
					resource.TestCheckResourceAttr(dataSourceName, "engine_version", engineVersion),
					resource.TestCheckResourceAttr(dataSourceName, "license_model", licenseModel),
					resource.TestCheckResourceAttr(dataSourceName, "storage_type", storageType),
				),
			},
		},
	})
}

func TestAccRDSOrderableInstanceDataSource_preferredClass(t *testing.T) {
	dataSourceName := "data.aws_rds_orderable_db_instance.test"
	preferredClass := "db.t2.micro"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccOrderableInstancePreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: nil,
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
	dataSourceName := "data.aws_rds_orderable_db_instance.test"
	preferredVersion := "5.7.22"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccOrderableInstancePreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccOrderableInstanceDataSourceConfig_preferredVersion(preferredVersion),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "engine_version", preferredVersion),
				),
			},
		},
	})
}

func TestAccRDSOrderableInstanceDataSource_preferredClassAndVersion(t *testing.T) {
	dataSourceName := "data.aws_rds_orderable_db_instance.test"
	preferredClass := "db.m3.medium"
	preferredVersion := "5.7.22"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccOrderableInstancePreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccOrderableInstanceDataSourceConfig_preferredClassAndVersion(preferredClass, preferredVersion),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "instance_class", preferredClass),
					resource.TestCheckResourceAttr(dataSourceName, "engine_version", preferredVersion),
				),
			},
		},
	})
}

func TestAccRDSOrderableInstanceDataSource_supportsEnhancedMonitoring(t *testing.T) {
	dataSourceName := "data.aws_rds_orderable_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccOrderableInstancePreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: nil,
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
	dataSourceName := "data.aws_rds_orderable_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccOrderableInstancePreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: nil,
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
	dataSourceName := "data.aws_rds_orderable_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccOrderableInstancePreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: nil,
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
	dataSourceName := "data.aws_rds_orderable_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccOrderableInstancePreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: nil,
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
	dataSourceName := "data.aws_rds_orderable_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccOrderableInstancePreCheck(t)
			testAccRDSPerformanceInsightsDefaultVersionPreCheck(t, "mysql")
		},
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: nil,
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
	dataSourceName := "data.aws_rds_orderable_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccOrderableInstancePreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: nil,
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
	dataSourceName := "data.aws_rds_orderable_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccOrderableInstancePreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: nil,
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

func testAccOrderableInstanceDataSourceConfig_basic(class, engine, version, license, storage string) string {
	return fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  instance_class = %q
  engine         = %q
  engine_version = %q
  license_model  = %q
  storage_type   = %q
}
`, class, engine, version, license, storage)
}

func testAccOrderableInstanceDataSourceConfig_preferredClass(preferredClass string) string {
	return fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine         = "mysql"
  engine_version = "5.7.22"
  license_model  = "general-public-license"

  preferred_instance_classes = [
    "db.xyz.xlarge",
    %q,
    "db.t3.small",
  ]
}
`, preferredClass)
}

func testAccOrderableInstanceDataSourceConfig_preferredVersion(preferredVersion string) string {
	return fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine        = "mysql"
  license_model = "general-public-license"
  storage_type  = "standard"

  preferred_engine_versions = [
    "18.42.32",
    %q,
    "not.a.version",
  ]
}
`, preferredVersion)
}

func testAccOrderableInstanceDataSourceConfig_preferredClassAndVersion(preferredClass, preferredVersion string) string {
	return fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine        = "mysql"
  license_model = "general-public-license"

  preferred_instance_classes = [
    "db.xyz.xlarge",
    %q,
    "db.t3.small",
  ]
  preferred_engine_versions = [
    "18.42.32",
    %q,
    "not.a.version",
  ]
}
`, preferredClass, preferredVersion)
}

func testAccOrderableInstanceDataSourceConfig_supportsEnhancedMonitoring() string {
	return `
data "aws_rds_orderable_db_instance" "test" {
  engine                       = "mysql"
  license_model                = "general-public-license"
  storage_type                 = "standard"
  supports_enhanced_monitoring = true

  preferred_engine_versions  = ["5.6.35", "5.6.41", "5.6.44"]
  preferred_instance_classes = ["db.t2.small", "db.t3.medium", "db.t3.large"]
}
`
}

func testAccOrderableInstanceDataSourceConfig_supportsIAMDatabaseAuthentication() string {
	return `
data "aws_rds_orderable_db_instance" "test" {
  engine                               = "mysql"
  license_model                        = "general-public-license"
  storage_type                         = "standard"
  supports_iam_database_authentication = true

  preferred_engine_versions  = ["5.6.35", "5.6.41", "5.6.44"]
  preferred_instance_classes = ["db.t2.small", "db.t3.medium", "db.t3.large"]
}
`
}

func testAccOrderableInstanceDataSourceConfig_supportsIops() string {
	return `
data "aws_rds_orderable_db_instance" "test" {
  engine        = "mysql"
  license_model = "general-public-license"
  supports_iops = true

  preferred_engine_versions  = ["8.0.20", "8.0.19", "8.0.17"]
  preferred_instance_classes = ["db.t3.small", "db.t2.xlarge", "db.t2.small"]
}
`
}

func testAccOrderableInstanceDataSourceConfig_supportsKerberosAuthentication() string {
	return `
data "aws_rds_orderable_db_instance" "test" {
  engine                           = "postgres"
  license_model                    = "postgresql-license"
  storage_type                     = "standard"
  supports_kerberos_authentication = true

  preferred_engine_versions  = ["12.3", "11.1", "10.13"]
  preferred_instance_classes = ["db.m5.xlarge", "db.r5.large", "db.t3.large"]
}
`
}

func testAccOrderableInstanceDataSourceConfig_supportsPerformanceInsights() string {
	return `
data "aws_rds_orderable_db_instance" "test" {
  engine                        = "mysql"
  license_model                 = "general-public-license"
  supports_performance_insights = true

  preferred_engine_versions  = ["5.6.35", "5.6.41", "5.6.44"]
  preferred_instance_classes = ["db.t2.small", "db.t3.medium", "db.t3.large"]
}
`
}

func testAccOrderableInstanceDataSourceConfig_supportsStorageAutoScaling() string {
	return `
data "aws_rds_orderable_db_instance" "test" {
  engine                       = "mysql"
  license_model                = "general-public-license"
  supports_storage_autoscaling = true

  preferred_engine_versions  = ["8.0.20", "8.0.19", "5.7.30"]
  preferred_instance_classes = ["db.t3.medium", "db.t2.large", "db.t3.xlarge"]
}
`
}

func testAccOrderableInstanceDataSourceConfig_supportsStorageEncryption() string {
	return `
data "aws_rds_orderable_db_instance" "test" {
  engine                      = "mysql"
  license_model               = "general-public-license"
  storage_type                = "standard"
  supports_storage_encryption = true

  preferred_engine_versions  = ["5.6.35", "5.6.41", "5.6.44"]
  preferred_instance_classes = ["db.t2.small", "db.t3.medium", "db.t3.large"]
}
`
}
