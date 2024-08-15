// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package m2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/m2"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfm2 "github.com/hashicorp/terraform-provider-aws/internal/service/m2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccM2Environment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	var environment m2.GetEnvironmentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_m2_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.M2EndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.M2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_basic(rName, "bluage"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &environment),
					resource.TestCheckNoResourceAttr(resourceName, "apply_changes_during_maintenance_window"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "m2", regexache.MustCompile(`env/+.`)),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "engine_type", "bluage"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngineVersion),
					resource.TestCheckResourceAttrSet(resourceName, "environment_id"),
					resource.TestCheckNoResourceAttr(resourceName, "force_update"),
					resource.TestCheckResourceAttr(resourceName, "high_availability_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, "M2.m5.large"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrKMSKeyID),
					resource.TestCheckResourceAttrSet(resourceName, "load_balancer_arn"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPreferredMaintenanceWindow),
					resource.TestCheckResourceAttr(resourceName, names.AttrPubliclyAccessible, acctest.CtFalse),
					acctest.CheckResourceAttrGreaterThanValue(resourceName, "security_group_ids.#", 0),
					acctest.CheckResourceAttrGreaterThanValue(resourceName, "subnet_ids.#", 0),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("storage_configuration"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccM2Environment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	var environment m2.GetEnvironmentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_m2_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.M2)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.M2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_basic(rName, "bluage"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &environment),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfm2.ResourceEnvironment, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccM2Environment_full(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_m2_environment.test"
	var environment m2.GetEnvironmentOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.M2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_full(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &environment),
					resource.TestCheckNoResourceAttr(resourceName, "apply_changes_during_maintenance_window"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "m2", regexache.MustCompile(`env/+.`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Test-1"),
					resource.TestCheckResourceAttr(resourceName, "engine_type", "microfocus"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "8.0.10"),
					resource.TestCheckResourceAttrSet(resourceName, "environment_id"),
					resource.TestCheckNoResourceAttr(resourceName, "force_update"),
					resource.TestCheckResourceAttr(resourceName, "high_availability_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "high_availability_config.0.desired_capacity", "5"),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, "M2.m5.large"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrKMSKeyID),
					resource.TestCheckResourceAttrSet(resourceName, "load_balancer_arn"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPreferredMaintenanceWindow),
					resource.TestCheckResourceAttr(resourceName, names.AttrPubliclyAccessible, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", acctest.Ct2),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("storage_configuration"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccM2Environment_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	var environment m2.GetEnvironmentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_m2_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.M2EndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.M2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_update(rName, "M2.m5.large", 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &environment),
					resource.TestCheckResourceAttr(resourceName, "high_availability_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "high_availability_config.0.desired_capacity", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, "M2.m5.large"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPreferredMaintenanceWindow, "sat:03:35-sat:05:35"),
				),
			},
			{
				Config: testAccEnvironmentConfig_update(rName, "M2.m6i.large", 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &environment),
					resource.TestCheckResourceAttr(resourceName, "high_availability_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "high_availability_config.0.desired_capacity", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, "M2.m6i.large"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPreferredMaintenanceWindow, "sat:03:35-sat:05:35"),
				),
			},
		},
	})
}

func TestAccM2Environment_efs(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	var environment m2.GetEnvironmentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_m2_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.M2EndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.M2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_efsComplete(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &environment),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("storage_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"efs": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									names.AttrFileSystemID: knownvalue.NotNull(), // TODO: should be Pair, depends on https://github.com/hashicorp/terraform-plugin-testing/pull/330
									"mount_point":          knownvalue.StringExact("/m2/mount/efsexample"),
								}),
							}),
							"fsx": knownvalue.ListExact([]knownvalue.Check{}),
						}),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccM2Environment_fsx(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var environment m2.GetEnvironmentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_m2_environment.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.M2EndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.M2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_fsxComplete(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &environment),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("storage_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"efs": knownvalue.ListExact([]knownvalue.Check{}),
							"fsx": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									names.AttrFileSystemID: knownvalue.NotNull(), // TODO: should be Pair, depends on https://github.com/hashicorp/terraform-plugin-testing/pull/330
									"mount_point":          knownvalue.StringExact("/m2/mount/fsxexample"),
								}),
							}),
						}),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckEnvironmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).M2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_m2_environment" {
				continue
			}

			_, err := tfm2.FindEnvironmentByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Mainframe Modernization Environment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckEnvironmentExists(ctx context.Context, n string, v *m2.GetEnvironmentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).M2Client(ctx)

		output, err := tfm2.FindEnvironmentByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).M2Client(ctx)

	input := &m2.ListEnvironmentsInput{}
	_, err := conn.ListEnvironments(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccEnvironmentConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = [aws_vpc.test.cidr_block]
  }
}
`, rName))
}

func testAccEnvironmentConfig_basic(rName, engineType string) string {
	return fmt.Sprintf(`
resource "aws_m2_environment" "test" {
  name          = %[1]q
  engine_type   = %[2]q
  instance_type = "M2.m5.large"
}
`, rName, engineType)
}

func testAccEnvironmentConfig_full(rName string) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName), fmt.Sprintf(`
resource "aws_m2_environment" "test" {
  description    = "Test-1"
  engine_type    = "microfocus"
  engine_version = "8.0.10"

  high_availability_config {
    desired_capacity = 5
  }

  instance_type      = "M2.m5.large"
  kms_key_id         = aws_kms_key.test.arn
  name               = %[1]q
  security_group_ids = [aws_security_group.test.id]
  subnet_ids         = aws_subnet.test[*].id
}

resource "aws_kms_key" "test" {
  description = %[1]q
}
`, rName))
}

func testAccEnvironmentConfig_update(rName, instanceType string, desiredCapacity int32) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName), fmt.Sprintf(`
resource "aws_m2_environment" "test" {
  name               = %[1]q
  engine_type        = "bluage"
  engine_version     = "3.7.0"
  instance_type      = %[2]q
  security_group_ids = [aws_security_group.test.id]
  subnet_ids         = aws_subnet.test[*].id

  preferred_maintenance_window = "sat:03:35-sat:05:35"

  high_availability_config {
    desired_capacity = %[3]d
  }
}
`, rName, instanceType, desiredCapacity))
}

func testAccEnvironmentConfig_efsComplete(rName string) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName),
		fmt.Sprintf(`
resource "aws_m2_environment" "test" {
  name               = %[1]q
  engine_type        = "bluage"
  engine_version     = "3.7.0"
  instance_type      = "M2.m5.large"
  security_group_ids = [aws_security_group.test.id]
  subnet_ids         = aws_subnet.test[*].id

  storage_configuration {
    efs {
      file_system_id = aws_efs_file_system.test.id
      mount_point    = "/m2/mount/efsexample"
    }
  }

  depends_on = [aws_efs_mount_target.test]
}

resource "aws_efs_file_system" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_efs_access_point" "test" {
  file_system_id = aws_efs_file_system.test.id

  root_directory {
    path = "/"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_efs_mount_target" "test" {
  count = 2

  file_system_id  = aws_efs_file_system.test.id
  subnet_id       = aws_subnet.test[count.index].id
  security_groups = [aws_security_group.test.id]
}

`, rName))
}

func testAccEnvironmentConfig_fsxComplete(rName string) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName), fmt.Sprintf(`
resource "aws_m2_environment" "test" {
  name               = %[1]q
  engine_type        = "bluage"
  engine_version     = "3.7.0"
  instance_type      = "M2.m5.large"
  security_group_ids = [aws_security_group.test.id]
  subnet_ids         = aws_subnet.test[*].id

  storage_configuration {
    fsx {
      file_system_id = aws_fsx_lustre_file_system.test.id
      mount_point    = "/m2/mount/fsxexample"
    }
  }
}

resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity   = 1200
  subnet_ids         = [aws_subnet.test[0].id]
  security_group_ids = [aws_security_group.test.id]

  tags = {
    Name = %[1]q
  }
}
`, rName))
}
