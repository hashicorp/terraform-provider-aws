// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package docdb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/docdb/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdocdb "github.com/hashicorp/terraform-provider-aws/internal/service/docdb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDocDBClusterInstance_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DBInstance
	resourceName := "aws_docdb_cluster_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, resourceName, &v),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrApplyImmediately),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "rds", regexache.MustCompile(fmt.Sprintf("db:%s", rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrAutoMinorVersionUpgrade, acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttrSet(resourceName, "ca_cert_identifier"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrClusterIdentifier),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_snapshot", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, "db_subnet_group_name"),
					resource.TestCheckResourceAttrSet(resourceName, "dbi_resource_id"),
					resource.TestCheckNoResourceAttr(resourceName, "enable_performance_insights"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEndpoint),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "docdb"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngineVersion),
					resource.TestCheckResourceAttr(resourceName, names.AttrIdentifier, rName),
					resource.TestCheckResourceAttr(resourceName, "identifier_prefix", ""),
					resource.TestCheckResourceAttrSet(resourceName, "instance_class"),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
					resource.TestCheckNoResourceAttr(resourceName, "performance_insights_kms_key_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPort),
					resource.TestCheckResourceAttrSet(resourceName, "preferred_backup_window"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPreferredMaintenanceWindow),
					resource.TestCheckResourceAttr(resourceName, "promotion_tier", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPubliclyAccessible),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStorageEncrypted),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "writer", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
				},
			},
			{
				Config: testAccClusterInstanceConfig_modified(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrApplyImmediately, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrAutoMinorVersionUpgrade, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "promotion_tier", acctest.Ct3),
				),
			},
		},
	})
}

func TestAccDocDBClusterInstance_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DBInstance
	resourceName := "aws_docdb_cluster_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdocdb.ResourceClusterInstance(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDocDBClusterInstance_identifierGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DBInstance
	resourceName := "aws_docdb_cluster_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_identifierGenerated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrNameGeneratedWithPrefix(resourceName, names.AttrIdentifier, "tf-"),
					resource.TestCheckResourceAttr(resourceName, "identifier_prefix", "tf-"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
				},
			},
		},
	})
}

func TestAccDocDBClusterInstance_identifierPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DBInstance
	resourceName := "aws_docdb_cluster_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_identifierPrefix(rName, "tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrIdentifier, "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, "identifier_prefix", "tf-acc-test-prefix-"),
				),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
				},
			},
		},
	})
}

func TestAccDocDBClusterInstance_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DBInstance
	resourceName := "aws_docdb_cluster_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
				},
			},
			{
				Config: testAccClusterInstanceConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccClusterInstanceConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccDocDBClusterInstance_performanceInsights(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DBInstance
	resourceName := "aws_docdb_cluster_instance.test"
	rNamePrefix := acctest.ResourcePrefix
	rName := sdkacctest.RandomWithPrefix(rNamePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_performanceInsights(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "enable_performance_insights"),
					resource.TestCheckResourceAttrSet(resourceName, "performance_insights_kms_key_id"),
				),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					"enable_performance_insights",
					"performance_insights_kms_key_id",
				},
			},
		},
	})
}

func TestAccDocDBClusterInstance_az(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DBInstance
	resourceName := "aws_docdb_cluster_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_az(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAvailabilityZone),
				),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
				},
			},
		},
	})
}

func TestAccDocDBClusterInstance_kmsKey(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DBInstance
	resourceName := "aws_docdb_cluster_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_kmsKey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, "aws_kms_key.test", names.AttrARN),
				),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
				},
			},
		},
	})
}

func TestAccDocDBClusterInstance_copyTagsToSnapshot(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DBInstance
	resourceName := "aws_docdb_cluster_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_copyTagsToSnapshot(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_snapshot", acctest.CtTrue),
				),
			},
			{
				Config: testAccClusterInstanceConfig_copyTagsToSnapshot(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_snapshot", acctest.CtFalse),
				),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
				},
			},
		},
	})
}

func testAccCheckClusterInstanceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DocDBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_docdb_cluster_instance" {
				continue
			}

			_, err := tfdocdb.FindDBInstanceByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("DocumentDB Cluster Instance %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckClusterInstanceExists(ctx context.Context, n string, v *awstypes.DBInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DocDBClient(ctx)

		output, err := tfdocdb.FindDBInstanceByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccClusterInstanceConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_docdb_cluster" "test" {
  cluster_identifier  = %[1]q
  master_password     = "avoid-plaintext-passwords"
  master_username     = "tfacctest"
  skip_final_snapshot = true
}

data "aws_docdb_orderable_db_instance" "test" {
  engine                     = aws_docdb_cluster.test.engine
  preferred_instance_classes = ["db.t3.medium", "db.4tg.medium", "db.r5.large", "db.r6g.large"]
}
`, rName)
}

func testAccClusterInstanceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccClusterInstanceConfig_base(rName), fmt.Sprintf(`
resource "aws_docdb_cluster_instance" "test" {
  identifier         = %[1]q
  cluster_identifier = aws_docdb_cluster.test.id
  instance_class     = data.aws_docdb_orderable_db_instance.test.instance_class
}
`, rName))
}

func testAccClusterInstanceConfig_modified(rName string) string {
	return acctest.ConfigCompose(testAccClusterInstanceConfig_base(rName), fmt.Sprintf(`
resource "aws_docdb_cluster_instance" "test" {
  identifier                 = %[1]q
  cluster_identifier         = aws_docdb_cluster.test.id
  instance_class             = data.aws_docdb_orderable_db_instance.test.instance_class
  apply_immediately          = true
  auto_minor_version_upgrade = false
  promotion_tier             = 3
}
`, rName))
}

func testAccClusterInstanceConfig_identifierGenerated(rName string) string {
	return acctest.ConfigCompose(testAccClusterInstanceConfig_base(rName), `
resource "aws_docdb_cluster_instance" "test" {
  cluster_identifier = aws_docdb_cluster.test.id
  instance_class     = data.aws_docdb_orderable_db_instance.test.instance_class
}
`)
}

func testAccClusterInstanceConfig_identifierPrefix(rName, prefix string) string {
	return acctest.ConfigCompose(testAccClusterInstanceConfig_base(rName), fmt.Sprintf(`
resource "aws_docdb_cluster_instance" "test" {
  identifier_prefix  = %[1]q
  cluster_identifier = aws_docdb_cluster.test.id
  instance_class     = data.aws_docdb_orderable_db_instance.test.instance_class
}
`, prefix))
}

func testAccClusterInstanceConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccClusterInstanceConfig_base(rName), fmt.Sprintf(`
resource "aws_docdb_cluster_instance" "test" {
  identifier         = %[1]q
  cluster_identifier = aws_docdb_cluster.test.id
  instance_class     = data.aws_docdb_orderable_db_instance.test.instance_class

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccClusterInstanceConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccClusterInstanceConfig_base(rName), fmt.Sprintf(`
resource "aws_docdb_cluster_instance" "test" {
  identifier         = %[1]q
  cluster_identifier = aws_docdb_cluster.test.id
  instance_class     = data.aws_docdb_orderable_db_instance.test.instance_class

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccClusterInstanceConfig_performanceInsights(rName string) string {
	return acctest.ConfigCompose(testAccClusterInstanceConfig_base(rName), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = %[1]q

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
  	  "Sid": "Enable IAM User Permissions",
	  "Effect": "Allow",
	  "Principal": {
	    "AWS": "*"
	  },
	  "Action": "kms:*",
	  "Resource": "*"
	}
  ]
}
POLICY
}

resource "aws_docdb_cluster_instance" "test" {
  identifier                      = %[1]q
  cluster_identifier              = aws_docdb_cluster.test.id
  instance_class                  = data.aws_docdb_orderable_db_instance.test.instance_class
  promotion_tier                  = "3"
  enable_performance_insights     = true
  performance_insights_kms_key_id = aws_kms_key.test.arn
}
`, rName))
}

func testAccClusterInstanceConfig_az(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_docdb_cluster" "test" {
  cluster_identifier  = %[1]q
  availability_zones  = [data.aws_availability_zones.available.names[0], data.aws_availability_zones.available.names[1], data.aws_availability_zones.available.names[2]]
  master_password     = "avoid-plaintext-passwords"
  master_username     = "tfacctest"
  skip_final_snapshot = true
}

data "aws_docdb_orderable_db_instance" "test" {
  engine                     = aws_docdb_cluster.test.engine
  preferred_instance_classes = ["db.t3.medium", "db.4tg.medium", "db.r5.large", "db.r6g.large"]
}

resource "aws_docdb_cluster_instance" "test" {
  identifier         = %[1]q
  cluster_identifier = aws_docdb_cluster.test.id
  instance_class     = data.aws_docdb_orderable_db_instance.test.instance_class
  promotion_tier     = "3"
  availability_zone  = data.aws_availability_zones.available.names[0]
}
`, rName))
}

func testAccClusterInstanceConfig_kmsKey(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = %[1]q

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_docdb_cluster" "test" {
  cluster_identifier  = %[1]q
  master_password     = "avoid-plaintext-passwords"
  master_username     = "tfacctest"
  storage_encrypted   = true
  kms_key_id          = aws_kms_key.test.arn
  skip_final_snapshot = true
}

data "aws_docdb_orderable_db_instance" "test" {
  engine                     = aws_docdb_cluster.test.engine
  preferred_instance_classes = ["db.t3.medium", "db.4tg.medium", "db.r5.large", "db.r6g.large"]
}

resource "aws_docdb_cluster_instance" "test" {
  identifier         = %[1]q
  cluster_identifier = aws_docdb_cluster.test.id
  instance_class     = data.aws_docdb_orderable_db_instance.test.instance_class
}
`, rName))
}

func testAccClusterInstanceConfig_copyTagsToSnapshot(rName string, flag bool) string {
	return acctest.ConfigCompose(testAccClusterInstanceConfig_base(rName), fmt.Sprintf(`
resource "aws_docdb_cluster_instance" "test" {
  identifier            = %[1]q
  cluster_identifier    = aws_docdb_cluster.test.id
  copy_tags_to_snapshot = %[2]t
  instance_class        = data.aws_docdb_orderable_db_instance.test.instance_class
  promotion_tier        = "3"
}
`, rName, flag))
}
