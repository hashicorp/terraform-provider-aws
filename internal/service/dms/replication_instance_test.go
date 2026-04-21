// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dms_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdms "github.com/hashicorp/terraform-provider-aws/internal/service/dms"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDMSReplicationInstance_basic(t *testing.T) {
	ctx := acctest.Context(t)
	// NOTE: Using larger dms.c4.large here for AWS GovCloud (US) support
	replicationInstanceClass := "dms.t3.micro"
	if acctest.Partition() == endpoints.AwsUsGovPartitionID {
		replicationInstanceClass = "dms.c4.large"
	}
	resourceName := "aws_dms_replication_instance.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationInstanceConfig_replicationInstanceClass(rName, replicationInstanceClass),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationInstanceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAllocatedStorage, "50"),
					resource.TestCheckResourceAttr(resourceName, names.AttrAutoMinorVersionUpgrade, acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngineVersion),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrKMSKeyARN),
					resource.TestCheckResourceAttr(resourceName, "multi_az", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "network_type", "IPV4"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPreferredMaintenanceWindow),
					resource.TestCheckResourceAttr(resourceName, names.AttrPubliclyAccessible, acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, "replication_instance_arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_instance_class", replicationInstanceClass),
					resource.TestCheckResourceAttr(resourceName, "replication_instance_id", rName),
					resource.TestCheckResourceAttr(resourceName, "replication_instance_private_ips.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_instance_public_ips.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "replication_subnet_group_id"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately},
			},
		},
	})
}

func TestAccDMSReplicationInstance_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	// NOTE: Using larger dms.c4.large here for AWS GovCloud (US) support
	replicationInstanceClass := "dms.t3.micro"
	if acctest.Partition() == endpoints.AwsUsGovPartitionID {
		replicationInstanceClass = "dms.c4.large"
	}
	resourceName := "aws_dms_replication_instance.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationInstanceConfig_replicationInstanceClass(rName, replicationInstanceClass),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfdms.ResourceReplicationInstance(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDMSReplicationInstance_allocatedStorage(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_replication_instance.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationInstanceConfig_allocatedStorage(rName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAllocatedStorage, "5"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately},
			},
			{
				Config: testAccReplicationInstanceConfig_allocatedStorage(rName, 6),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAllocatedStorage, "6"),
				),
			},
		},
	})
}

func TestAccDMSReplicationInstance_autoMinorVersionUpgrade(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_replication_instance.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationInstanceConfig_autoMinorVersionUpgrade(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAutoMinorVersionUpgrade, acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately},
			},
			{
				Config: testAccReplicationInstanceConfig_autoMinorVersionUpgrade(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAutoMinorVersionUpgrade, acctest.CtFalse),
				),
			},
			{
				Config: testAccReplicationInstanceConfig_autoMinorVersionUpgrade(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAutoMinorVersionUpgrade, acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccDMSReplicationInstance_availabilityZone(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_availability_zones.available"
	resourceName := "aws_dms_replication_instance.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationInstanceConfig_availabilityZone(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAvailabilityZone, dataSourceName, "names.0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately},
			},
		},
	})
}

func TestAccDMSReplicationInstance_engineVersion(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_replication_instance.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationInstanceConfig_engineVersion(rName, "3.5.2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "3.5.2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrAllowMajorVersionUpgrade, names.AttrApplyImmediately},
			},
			{
				Config: testAccReplicationInstanceConfig_engineVersion(rName, "3.5.3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "3.5.3"),
				),
			},
		},
	})
}

func TestAccDMSReplicationInstance_kmsKeyARN(t *testing.T) {
	ctx := acctest.Context(t)
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_dms_replication_instance.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationInstanceConfig_kmsKeyARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyARN, kmsKeyResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately},
			},
		},
	})
}

func TestAccDMSReplicationInstance_multiAz(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_replication_instance.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationInstanceConfig_multiAz(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "multi_az", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately},
			},
			{
				Config: testAccReplicationInstanceConfig_multiAz(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "multi_az", acctest.CtFalse),
				),
			},
			{
				Config: testAccReplicationInstanceConfig_multiAz(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "multi_az", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccDMSReplicationInstance_networkType(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_replication_instance.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationInstanceConfig_networkType(rName, "IPV4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "network_type", "IPV4"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately},
			},
			{
				Config: testAccReplicationInstanceConfig_networkType(rName, "DUAL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "network_type", "DUAL"),
				),
			},
		},
	})
}

func TestAccDMSReplicationInstance_preferredMaintenanceWindow(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_replication_instance.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationInstanceConfig_preferredMaintenanceWindow(rName, "sun:00:30-sun:02:30"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrPreferredMaintenanceWindow, "sun:00:30-sun:02:30"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately},
			},
			{
				Config: testAccReplicationInstanceConfig_preferredMaintenanceWindow(rName, "mon:00:30-mon:02:30"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrPreferredMaintenanceWindow, "mon:00:30-mon:02:30"),
				),
			},
		},
	})
}

func TestAccDMSReplicationInstance_publiclyAccessible(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_replication_instance.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationInstanceConfig_publiclyAccessible(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrPubliclyAccessible, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "replication_instance_public_ips.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately},
			},
		},
	})
}

func TestAccDMSReplicationInstance_replicationInstanceClass(t *testing.T) {
	ctx := acctest.Context(t)
	// NOTE: Using larger dms.c4.(x)?large here for AWS GovCloud (US) support
	replicationInstanceClass1 := "dms.c4.large"
	replicationInstanceClass2 := "dms.c4.xlarge"
	resourceName := "aws_dms_replication_instance.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationInstanceConfig_replicationInstanceClass(rName, replicationInstanceClass1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "replication_instance_class", replicationInstanceClass1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately},
			},
			{
				Config: testAccReplicationInstanceConfig_replicationInstanceClass(rName, replicationInstanceClass2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "replication_instance_class", replicationInstanceClass2),
				),
			},
		},
	})
}

func TestAccDMSReplicationInstance_vpcSecurityGroupIDs(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_replication_instance.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationInstanceConfig_vpcSecurityGroupIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately},
			},
		},
	})
}

func TestAccDMSReplicationInstance_kerberosAuthenticationSettings(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_replication_instance.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationInstanceConfig_kerberosAuthenticationSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "dns_name_servers", ""),
					resource.TestCheckResourceAttr(resourceName, "kerberos_authentication_settings.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately},
			},
		},
	})
}

func testAccCheckReplicationInstanceExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).DMSClient(ctx)

		_, err := tfdms.FindReplicationInstanceByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckReplicationInstanceDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DMSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dms_replication_instance" {
				continue
			}

			_, err := tfdms.FindReplicationInstanceByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("DMS Replication Instance %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

// Ideally we'd like to be able to leverage the "default" replication subnet group.
// However, it may not exist or may include deleted subnets.
func testAccReplicationInstanceConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_dms_replication_subnet_group" "test" {
  replication_subnet_group_id          = %[1]q
  replication_subnet_group_description = "testing"
  subnet_ids                           = aws_subnet.test[*].id
}
`, rName))
}

func testAccReplicationInstanceConfig_allocatedStorage(rName string, allocatedStorage int) string {
	return acctest.ConfigCompose(testAccReplicationInstanceConfig_base(rName), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_dms_replication_instance" "test" {
  allocated_storage           = %[2]d
  apply_immediately           = true
  replication_instance_class  = data.aws_partition.current.partition == "aws" ? "dms.t3.micro" : "dms.c4.large"
  replication_instance_id     = %[1]q
  replication_subnet_group_id = aws_dms_replication_subnet_group.test.id
}
`, rName, allocatedStorage))
}

func testAccReplicationInstanceConfig_autoMinorVersionUpgrade(rName string, autoMinorVersionUpgrade bool) string {
	return acctest.ConfigCompose(testAccReplicationInstanceConfig_base(rName), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_dms_replication_instance" "test" {
  apply_immediately           = true
  auto_minor_version_upgrade  = %[2]t
  replication_instance_class  = data.aws_partition.current.partition == "aws" ? "dms.t3.micro" : "dms.c4.large"
  replication_instance_id     = %[1]q
  replication_subnet_group_id = aws_dms_replication_subnet_group.test.id
}
`, rName, autoMinorVersionUpgrade))
}

func testAccReplicationInstanceConfig_availabilityZone(rName string) string {
	return acctest.ConfigCompose(testAccReplicationInstanceConfig_base(rName), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_dms_replication_instance" "test" {
  apply_immediately           = true
  availability_zone           = data.aws_availability_zones.available.names[0]
  replication_instance_class  = data.aws_partition.current.partition == "aws" ? "dms.t3.micro" : "dms.c4.large"
  replication_instance_id     = %[1]q
  replication_subnet_group_id = aws_dms_replication_subnet_group.test.id
}
`, rName))
}

func testAccReplicationInstanceConfig_engineVersion(rName, engineVersion string) string {
	return acctest.ConfigCompose(testAccReplicationInstanceConfig_base(rName), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_dms_replication_instance" "test" {
  apply_immediately           = true
  allow_major_version_upgrade = true
  engine_version              = %[2]q
  replication_instance_class  = data.aws_partition.current.partition == "aws" ? "dms.t3.micro" : "dms.c4.large"
  replication_instance_id     = %[1]q
  replication_subnet_group_id = aws_dms_replication_subnet_group.test.id
}
`, rName, engineVersion))
}

func testAccReplicationInstanceConfig_kmsKeyARN(rName string) string {
	return acctest.ConfigCompose(testAccReplicationInstanceConfig_base(rName), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_dms_replication_instance" "test" {
  apply_immediately           = true
  kms_key_arn                 = aws_kms_key.test.arn
  replication_instance_class  = data.aws_partition.current.partition == "aws" ? "dms.t3.micro" : "dms.c4.large"
  replication_instance_id     = %[1]q
  replication_subnet_group_id = aws_dms_replication_subnet_group.test.id
}
`, rName))
}

func testAccReplicationInstanceConfig_multiAz(rName string, multiAz bool) string {
	return acctest.ConfigCompose(testAccReplicationInstanceConfig_base(rName), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_dms_replication_instance" "test" {
  apply_immediately           = true
  multi_az                    = %[2]t
  replication_instance_class  = data.aws_partition.current.partition == "aws" ? "dms.t3.micro" : "dms.c4.large"
  replication_instance_id     = %[1]q
  replication_subnet_group_id = aws_dms_replication_subnet_group.test.id
}
`, rName, multiAz))
}

func testAccReplicationInstanceConfig_networkType(rName, networkType string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnetsIPv6(rName, 2), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_dms_replication_subnet_group" "test" {
  replication_subnet_group_id          = %[1]q
  replication_subnet_group_description = "testing"
  subnet_ids                           = aws_subnet.test[*].id
}

resource "aws_dms_replication_instance" "test" {
  apply_immediately           = true
  network_type                = %[2]q
  replication_instance_class  = data.aws_partition.current.partition == "aws" ? "dms.t3.micro" : "dms.c4.large"
  replication_instance_id     = %[1]q
  replication_subnet_group_id = aws_dms_replication_subnet_group.test.id
}
`, rName, networkType))
}

func testAccReplicationInstanceConfig_preferredMaintenanceWindow(rName, preferredMaintenanceWindow string) string {
	return acctest.ConfigCompose(testAccReplicationInstanceConfig_base(rName), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_dms_replication_instance" "test" {
  apply_immediately            = true
  preferred_maintenance_window = %[2]q
  replication_instance_class   = data.aws_partition.current.partition == "aws" ? "dms.t3.micro" : "dms.c4.large"
  replication_instance_id      = %[1]q
  replication_subnet_group_id  = aws_dms_replication_subnet_group.test.id
}
`, rName, preferredMaintenanceWindow))
}

func testAccReplicationInstanceConfig_publiclyAccessible(rName string, publiclyAccessible bool) string {
	return acctest.ConfigCompose(testAccReplicationInstanceConfig_base(rName), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_dms_replication_instance" "test" {
  apply_immediately           = true
  publicly_accessible         = %[2]t
  replication_instance_class  = data.aws_partition.current.partition == "aws" ? "dms.t3.micro" : "dms.c4.large"
  replication_instance_id     = %[1]q
  replication_subnet_group_id = aws_dms_replication_subnet_group.test.id

  depends_on = [aws_internet_gateway.test]
}
`, rName, publiclyAccessible))
}

func testAccReplicationInstanceConfig_replicationInstanceClass(rName, replicationInstanceClass string) string {
	return acctest.ConfigCompose(testAccReplicationInstanceConfig_base(rName), fmt.Sprintf(`
resource "aws_dms_replication_instance" "test" {
  apply_immediately           = true
  replication_instance_class  = %[1]q
  replication_instance_id     = %[2]q
  replication_subnet_group_id = aws_dms_replication_subnet_group.test.id
}
`, replicationInstanceClass, rName))
}

func testAccReplicationInstanceConfig_vpcSecurityGroupIDs(rName string) string {
	return acctest.ConfigCompose(testAccReplicationInstanceConfig_base(rName), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_dms_replication_instance" "test" {
  apply_immediately           = true
  replication_instance_class  = data.aws_partition.current.partition == "aws" ? "dms.t3.micro" : "dms.c4.large"
  replication_instance_id     = %[1]q
  replication_subnet_group_id = aws_dms_replication_subnet_group.test.replication_subnet_group_id
  vpc_security_group_ids      = [aws_security_group.test.id]
}
`, rName))
}

func testAccReplicationInstanceConfig_kerberosAuthenticationSettings(rName string) string {
	return acctest.ConfigCompose(testAccReplicationInstanceConfig_base(rName),
		testAccReplicationInstanceConfig_secretBase(rName), fmt.Sprintf(`
resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "test" {
  subnet_id      = aws_subnet.test[0].id
  route_table_id = aws_route_table.test.id
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_dms_replication_instance" "test" {
  apply_immediately           = true
  replication_instance_class  = data.aws_partition.current.partition == "aws" ? "dms.t3.micro" : "dms.c4.large"
  replication_instance_id     = %[1]q
  replication_subnet_group_id = aws_dms_replication_subnet_group.test.replication_subnet_group_id
  vpc_security_group_ids      = [aws_security_group.test.id]

  kerberos_authentication_settings {
    key_cache_secret_iam_arn = aws_iam_role.test.arn
    key_cache_secret_id      = aws_secretsmanager_secret.test.id
    krb5_file_contents       = file("test-fixtures/krb5.conf")
  }
}
`, rName))
}

func testAccReplicationInstanceConfig_secretBase(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}
data "aws_partition" "current" {}

resource "aws_secretsmanager_secret" "test" {
  name                    = %[1]q
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_binary = filebase64("test-fixtures/keytab.krb")
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "dms.${data.aws_region.current.region}.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name   = %[1]q
  role   = aws_iam_role.test.id
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
        "Action": "secretsmanager:*",
        "Effect": "Allow",
        "Resource": "*"
    }
  ]
}
EOF
}
`, rName)
}
