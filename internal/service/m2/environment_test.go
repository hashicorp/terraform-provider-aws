// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package m2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/m2"
	"github.com/aws/aws-sdk-go-v2/service/m2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/names"

	tfm2 "github.com/hashicorp/terraform-provider-aws/internal/service/m2"
)

const (
	testEngineType          = "bluage"
	testEngineVersion       = "3.7.0"
	testEngineUpdateVersion = "3.8.0"
)

func TestAccM2Environment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
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
		ErrorCheck:               acctest.ErrorCheck(t, names.M2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_basic(rName, testEngineType, testEngineVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &environment),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
					resource.TestCheckResourceAttr(resourceName, "engine_type", testEngineType),
					resource.TestCheckResourceAttr(resourceName, "engine_version", testEngineVersion),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "m2", regexache.MustCompile(`env/+.`)),
					resource.TestCheckResourceAttr(resourceName, "high_availability_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "efs_mount.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "fsx_mount.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "M2.m5.large"),
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

func TestAccM2Environment_update(t *testing.T) {
	ctx := acctest.Context(t)
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
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
		ErrorCheck:               acctest.ErrorCheck(t, names.M2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_highAvailability(rName, testEngineType, testEngineVersion, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &environment),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
					resource.TestCheckResourceAttr(resourceName, "engine_type", testEngineType),
					resource.TestCheckResourceAttr(resourceName, "engine_version", testEngineVersion),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "m2", regexache.MustCompile(`env/+.`)),
					resource.TestCheckResourceAttr(resourceName, "high_availability_config.0.desired_capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "efs_mount.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "fsx_mount.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "M2.m5.large"),
				),
			},
			{
				Config: testAccEnvironmentConfig_update(rName, testEngineType, testEngineVersion, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &environment),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
					resource.TestCheckResourceAttr(resourceName, "engine_type", testEngineType),
					resource.TestCheckResourceAttr(resourceName, "engine_version", testEngineVersion),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "m2", regexache.MustCompile(`env/+.`)),
					resource.TestCheckResourceAttr(resourceName, "high_availability_config.0.desired_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "efs_mount.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "fsx_mount.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "M2.m6i.large"),
					resource.TestCheckResourceAttr(resourceName, "preferred_maintenance_window", "sat:03:35-sat:05:35"),
				),
			},
		},
	})
}

func TestAccM2Environment_highAvailability(t *testing.T) {
	ctx := acctest.Context(t)
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
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
		ErrorCheck:               acctest.ErrorCheck(t, names.M2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_highAvailability(rName, testEngineType, testEngineVersion, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &environment),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
					resource.TestCheckResourceAttr(resourceName, "engine_type", testEngineType),
					resource.TestCheckResourceAttr(resourceName, "engine_version", testEngineVersion),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "m2", regexache.MustCompile(`env/+.`)),
					resource.TestCheckResourceAttr(resourceName, "high_availability_config.0.desired_capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "efs_mount.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "fsx_mount.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "M2.m5.large"),
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

func TestAccM2Environment_efs(t *testing.T) {
	ctx := acctest.Context(t)
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
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
		ErrorCheck:               acctest.ErrorCheck(t, names.M2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_efsComplete(rName, testEngineType, testEngineVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &environment),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
					resource.TestCheckResourceAttr(resourceName, "engine_type", testEngineType),
					resource.TestCheckResourceAttr(resourceName, "engine_version", testEngineVersion),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "m2", regexache.MustCompile(`env/+.`)),
					resource.TestCheckResourceAttr(resourceName, "high_availability_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "efs_mount.0.mount_point", "/m2/mount/example"),
					resource.TestCheckResourceAttr(resourceName, "fsx_mount.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "M2.m5.large"),
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
				Config: testAccEnvironmentConfig_basic(rName, testEngineType, testEngineVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &environment),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfm2.ResourceEnvironment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
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

			_, err := conn.GetEnvironment(ctx, &m2.GetEnvironmentInput{
				EnvironmentId: aws.String(rs.Primary.ID),
			})
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.M2, create.ErrActionCheckingDestroyed, tfm2.ResNameEnvironment, rs.Primary.ID, err)
			}

			return create.Error(names.M2, create.ErrActionCheckingDestroyed, tfm2.ResNameEnvironment, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckEnvironmentExists(ctx context.Context, name string, environment *m2.GetEnvironmentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.M2, create.ErrActionCheckingExistence, tfm2.ResNameEnvironment, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.M2, create.ErrActionCheckingExistence, tfm2.ResNameEnvironment, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).M2Client(ctx)
		resp, err := conn.GetEnvironment(ctx, &m2.GetEnvironmentInput{
			EnvironmentId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.M2, create.ErrActionCheckingExistence, tfm2.ResNameEnvironment, rs.Primary.ID, err)
		}

		*environment = *resp

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

func testAccCheckEnvironmentNotRecreated(before, after *m2.GetEnvironmentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.EnvironmentId), aws.ToString(after.EnvironmentId); before != after {
			return create.Error(names.M2, create.ErrActionCheckingNotRecreated, tfm2.ResNameEnvironment, before, errors.New("recreated"))
		}

		return nil
	}
}

func testAccEnvironmentConfig_basic(rName, engineType, engineVersion string) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName),
		fmt.Sprintf(`

resource "aws_m2_environment" "test" {
  name            = %[1]q
  description     = %[1]q
  engine_type     = %[2]q
  engine_version  = %[3]q
  instance_type   = "M2.m5.large"
  security_groups = [aws_security_group.test.id]
  subnet_ids      = aws_subnet.test[*].id
}
`, rName, engineType, engineVersion))
}

func testAccEnvironmentConfig_update(rName, engineType, engineVersion string, desiredCapacity int32) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName),
		fmt.Sprintf(`

resource "aws_m2_environment" "test" {
  name            = %[1]q
  description     = %[1]q
  engine_type     = %[2]q
  engine_version  = %[3]q
  instance_type   = "M2.m6i.large"
  security_groups = [aws_security_group.test.id]
  subnet_ids      = aws_subnet.test[*].id
 
  preferred_maintenance_window = "sat:03:35-sat:05:35"

  high_availability_config {
    desired_capacity = %[4]d
  }

}
`, rName, engineType, engineVersion, desiredCapacity))

}

func testAccEnvironmentConfig_highAvailability(rName, engineType, engineVersion string, desiredCapacity int32) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName),
		fmt.Sprintf(`

resource "aws_m2_environment" "test" {
  name            = %[1]q
  description     = %[1]q
  engine_type     = %[2]q
  engine_version  = %[3]q
  instance_type   = "M2.m5.large"
  security_groups = [aws_security_group.test.id]
  subnet_ids      = aws_subnet.test[*].id 

  high_availability_config {
    desired_capacity = %[4]d
  }
}
`, rName, engineType, engineVersion, desiredCapacity))
}

func testAccEnvironmentConfig_efsComplete(rName, engineType, engineVersion string) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName),
		testAccEnvironmentConfig_efs(rName),
		fmt.Sprintf(`

resource "aws_m2_environment" "test" {
  name            = %[1]q
  description     = %[1]q
  engine_type     = %[2]q
  engine_version  = %[3]q
  instance_type   = "M2.m5.large"
  security_groups = [aws_security_group.test.id]
  subnet_ids      = aws_subnet.test[*].id 

  efs_mount {
    file_system_id = aws_efs_file_system.test.id
    mount_point    = "/m2/mount/example"
  }

}
`, rName, engineType, engineVersion))
}

func testAccEnvironmentConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		fmt.Sprintf(`
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
`, rName),
		fmt.Sprintf(`
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

func testAccEnvironmentConfig_efs(rName string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %[1]q
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
`, rName)
}
