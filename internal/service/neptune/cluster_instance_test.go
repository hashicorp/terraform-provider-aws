// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptune_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/service/neptune"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfneptune "github.com/hashicorp/terraform-provider-aws/internal/service/neptune"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNeptuneClusterInstance_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_cluster_instance.cluster_instances"
	clusterResourceName := "aws_neptune_cluster.test"
	parameterGroupResourceName := "aws_neptune_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAddress),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "rds", fmt.Sprintf("db:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrAutoMinorVersionUpgrade, acctest.CtTrue),
					resource.TestMatchResourceAttr(resourceName, names.AttrAvailabilityZone, regexache.MustCompile(fmt.Sprintf("^%s[a-z]{1}$", acctest.Region()))),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrClusterIdentifier, clusterResourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "dbi_resource_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAddress),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "neptune"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngineVersion),
					resource.TestCheckResourceAttr(resourceName, names.AttrIdentifier, rName),
					resource.TestCheckResourceAttr(resourceName, "identifier_prefix", ""),
					resource.TestCheckResourceAttrPair(resourceName, "instance_class", "data.aws_neptune_orderable_db_instance.test", "instance_class"),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyARN, ""),
					resource.TestCheckResourceAttrPair(resourceName, "neptune_parameter_group_name", parameterGroupResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "neptune_subnet_group_name", "default"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, strconv.Itoa(tfneptune.DefaultPort)),
					resource.TestCheckResourceAttrSet(resourceName, "preferred_backup_window"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPreferredMaintenanceWindow),
					resource.TestCheckResourceAttr(resourceName, "promotion_tier", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, names.AttrPubliclyAccessible, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageEncrypted, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, "standard"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "writer", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterInstanceConfig_modified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrAutoMinorVersionUpgrade, acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccNeptuneClusterInstance_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_cluster_instance.cluster_instances"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfneptune.ResourceClusterInstance(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccNeptuneClusterInstance_identifierGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_cluster_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
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
			},
		},
	})
}

func TestAccNeptuneClusterInstance_identifierPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_cluster_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
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
			},
		},
	})
}

func TestAccNeptuneClusterInstance_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_cluster_instance.cluster_instances"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
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

func TestAccNeptuneClusterInstance_withAZ(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_cluster_instance.cluster_instances"
	availabiltyZonesDataSourceName := "data.aws_availability_zones.available"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_az(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAvailabilityZone, availabiltyZonesDataSourceName, "names.0"),
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

func TestAccNeptuneClusterInstance_withSubnetGroup(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_cluster_instance.test"
	subnetGroupResourceName := "aws_neptune_subnet_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_subnetGroup(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "neptune_subnet_group_name", subnetGroupResourceName, names.AttrName),
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

func TestAccNeptuneClusterInstance_kmsKey(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_cluster_instance.cluster_instances"
	kmsKeyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_kmsKey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyARN, kmsKeyResourceName, names.AttrARN),
				),
			},
		},
	})
}

func testAccCheckClusterInstanceExists(ctx context.Context, n string, v *neptune.DBInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Neptune Cluster Instance ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NeptuneConn(ctx)

		output, err := tfneptune.FindDBInstanceByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckClusterInstanceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).NeptuneConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_neptune_cluster_instance" {
				continue
			}

			_, err := tfneptune.FindDBInstanceByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Neptune Cluster Instance %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccClusterInstanceConfig_baseSansCluster(rName string) string {
	return fmt.Sprintf(`
data "aws_neptune_orderable_db_instance" "test" {
  engine         = "neptune"
  engine_version = aws_neptune_cluster.test.engine_version
  license_model  = "amazon-license"

  preferred_instance_classes = ["db.t3.medium", "db.r5.large", "db.r4.large"]
}

resource "aws_neptune_parameter_group" "test" {
  name   = %[1]q
  family = "neptune1.3"

  parameter {
    name  = "neptune_query_timeout"
    value = "25"
  }
}
`, rName)
}

func testAccClusterInstanceConfig_base(rName string) string {
	return acctest.ConfigCompose(testAccClusterInstanceConfig_baseSansCluster(rName), acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier                   = %[1]q
  availability_zones                   = slice(data.aws_availability_zones.available.names, 0, min(3, length(data.aws_availability_zones.available.names)))
  engine                               = "neptune"
  neptune_cluster_parameter_group_name = "default.neptune1.3"
  skip_final_snapshot                  = true
}
`, rName))
}

func testAccClusterInstanceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccClusterInstanceConfig_base(rName), fmt.Sprintf(`
resource "aws_neptune_cluster_instance" "cluster_instances" {
  identifier                   = %[1]q
  cluster_identifier           = aws_neptune_cluster.test.id
  instance_class               = data.aws_neptune_orderable_db_instance.test.instance_class
  engine_version               = data.aws_neptune_orderable_db_instance.test.engine_version
  neptune_parameter_group_name = aws_neptune_parameter_group.test.name
  promotion_tier               = "3"
}
`, rName))
}

func testAccClusterInstanceConfig_modified(rName string) string {
	return acctest.ConfigCompose(testAccClusterInstanceConfig_base(rName), fmt.Sprintf(`
resource "aws_neptune_cluster_instance" "cluster_instances" {
  identifier                   = %[1]q
  cluster_identifier           = aws_neptune_cluster.test.id
  instance_class               = data.aws_neptune_orderable_db_instance.test.instance_class
  engine_version               = data.aws_neptune_orderable_db_instance.test.engine_version
  neptune_parameter_group_name = aws_neptune_parameter_group.test.name
  auto_minor_version_upgrade   = false
  promotion_tier               = "3"
}
`, rName))
}

func testAccClusterInstanceConfig_identifierGenerated(rName string) string {
	return acctest.ConfigCompose(testAccClusterInstanceConfig_base(rName), `
resource "aws_neptune_cluster_instance" "test" {
  cluster_identifier = aws_neptune_cluster.test.id
  instance_class     = data.aws_neptune_orderable_db_instance.test.instance_class
  engine_version     = data.aws_neptune_orderable_db_instance.test.engine_version

  neptune_parameter_group_name = aws_neptune_parameter_group.test.name
}
`)
}

func testAccClusterInstanceConfig_identifierPrefix(rName, prefix string) string {
	return acctest.ConfigCompose(testAccClusterInstanceConfig_base(rName), fmt.Sprintf(`
resource "aws_neptune_cluster_instance" "test" {
  identifier_prefix  = %[1]q
  cluster_identifier = aws_neptune_cluster.test.id
  instance_class     = data.aws_neptune_orderable_db_instance.test.instance_class
  engine_version     = data.aws_neptune_orderable_db_instance.test.engine_version

  neptune_parameter_group_name = aws_neptune_parameter_group.test.name
}
`, prefix))
}

func testAccClusterInstanceConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccClusterInstanceConfig_base(rName), fmt.Sprintf(`
resource "aws_neptune_cluster_instance" "cluster_instances" {
  identifier                   = %[1]q
  cluster_identifier           = aws_neptune_cluster.test.id
  instance_class               = data.aws_neptune_orderable_db_instance.test.instance_class
  engine_version               = data.aws_neptune_orderable_db_instance.test.engine_version
  neptune_parameter_group_name = aws_neptune_parameter_group.test.name
  promotion_tier               = "3"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccClusterInstanceConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccClusterInstanceConfig_base(rName), fmt.Sprintf(`
resource "aws_neptune_cluster_instance" "cluster_instances" {
  identifier                   = %[1]q
  cluster_identifier           = aws_neptune_cluster.test.id
  instance_class               = data.aws_neptune_orderable_db_instance.test.instance_class
  engine_version               = data.aws_neptune_orderable_db_instance.test.engine_version
  neptune_parameter_group_name = aws_neptune_parameter_group.test.name
  promotion_tier               = "3"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccClusterInstanceConfig_az(rName string) string {
	return acctest.ConfigCompose(testAccClusterInstanceConfig_base(rName), fmt.Sprintf(`
resource "aws_neptune_cluster_instance" "cluster_instances" {
  identifier                   = %[1]q
  cluster_identifier           = aws_neptune_cluster.test.id
  instance_class               = data.aws_neptune_orderable_db_instance.test.instance_class
  engine_version               = data.aws_neptune_orderable_db_instance.test.engine_version
  neptune_parameter_group_name = aws_neptune_parameter_group.test.name
  promotion_tier               = "3"
  availability_zone            = data.aws_availability_zones.available.names[0]
}
`, rName))
}

func testAccClusterInstanceConfig_subnetGroup(rName string) string {
	return acctest.ConfigCompose(testAccClusterInstanceConfig_baseSansCluster(rName), acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_neptune_cluster_instance" "test" {
  identifier         = %[1]q
  cluster_identifier = aws_neptune_cluster.test.id
  instance_class     = data.aws_neptune_orderable_db_instance.test.instance_class
  engine_version     = data.aws_neptune_orderable_db_instance.test.engine_version

  neptune_parameter_group_name = aws_neptune_parameter_group.test.name
}

resource "aws_neptune_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_neptune_cluster" "test" {
  cluster_identifier                   = %[1]q
  neptune_subnet_group_name            = aws_neptune_subnet_group.test.name
  neptune_cluster_parameter_group_name = "default.neptune1.3"
  skip_final_snapshot                  = true
}
`, rName))
}

func testAccClusterInstanceConfig_kmsKey(rName string) string {
	return acctest.ConfigCompose(testAccClusterInstanceConfig_baseSansCluster(rName), acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
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

resource "aws_neptune_cluster_instance" "cluster_instances" {
  identifier                   = %[1]q
  cluster_identifier           = aws_neptune_cluster.test.id
  instance_class               = data.aws_neptune_orderable_db_instance.test.instance_class
  engine_version               = data.aws_neptune_orderable_db_instance.test.engine_version
  neptune_parameter_group_name = aws_neptune_parameter_group.test.name
}

resource "aws_neptune_cluster" "test" {
  cluster_identifier  = %[1]q
  availability_zones  = slice(data.aws_availability_zones.available.names, 0, min(3, length(data.aws_availability_zones.available.names)))
  skip_final_snapshot = true
  storage_encrypted   = true
  kms_key_arn         = aws_kms_key.test.arn

  neptune_cluster_parameter_group_name = "default.neptune1.3"
}
`, rName))
}
