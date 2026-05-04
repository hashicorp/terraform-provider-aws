// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecs_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECSDaemonTaskDefinitionsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ecs_daemon_task_definitions.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonTaskDefinitionsDataSourceConfig_basic(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("daemon_task_definitions"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrARN:         tfknownvalue.RegionalARNRegexp("ecs", regexache.MustCompile(`daemon-task-definition/`+rName+`.*:\d+$`)),
							names.AttrStatus:      knownvalue.StringExact("ACTIVE"),
							"registered_at":       knownvalue.NotNull(),
							"registered_by":       knownvalue.NotNull(),
							"delete_requested_at": knownvalue.Null(),
						}),
					})),
				},
			},
		},
	})
}

func TestAccECSDaemonTaskDefinitionsDataSource_familyPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ecs_daemon_task_definitions.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonTaskDefinitionsDataSourceConfig_familyPrefix(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("daemon_task_definitions"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrARN:         tfknownvalue.RegionalARNRegexp("ecs", regexache.MustCompile(`daemon-task-definition/`+rName+`.*:\d+$`)),
							names.AttrStatus:      knownvalue.StringExact("ACTIVE"),
							"registered_at":       knownvalue.NotNull(),
							"registered_by":       knownvalue.NotNull(),
							"delete_requested_at": knownvalue.Null(),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrARN:         tfknownvalue.RegionalARNRegexp("ecs", regexache.MustCompile(`daemon-task-definition/`+rName+`.*:\d+$`)),
							names.AttrStatus:      knownvalue.StringExact("ACTIVE"),
							"registered_at":       knownvalue.NotNull(),
							"registered_by":       knownvalue.NotNull(),
							"delete_requested_at": knownvalue.Null(),
						}),
					})),
				},
			},
		},
	})
}

func TestAccECSDaemonTaskDefinitionsDataSource_family(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ecs_daemon_task_definitions.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonTaskDefinitionsDataSourceConfig_family(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("daemon_task_definitions"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrARN:         tfknownvalue.RegionalARNRegexp("ecs", regexache.MustCompile(`daemon-task-definition/`+rName+`.*:\d+$`)),
							names.AttrStatus:      knownvalue.StringExact("ACTIVE"),
							"registered_at":       knownvalue.NotNull(),
							"registered_by":       knownvalue.NotNull(),
							"delete_requested_at": knownvalue.Null(),
						}),
					})),
				},
			},
		},
	})
}

func TestAccECSDaemonTaskDefinitionsDataSource_revision(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ecs_daemon_task_definitions.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonTaskDefinitionsDataSourceConfig_revision(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("daemon_task_definitions"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrARN:         tfknownvalue.RegionalARNRegexp("ecs", regexache.MustCompile(`daemon-task-definition/`+rName+`.*:\d+$`)),
							names.AttrStatus:      knownvalue.StringExact("ACTIVE"),
							"registered_at":       knownvalue.NotNull(),
							"registered_by":       knownvalue.NotNull(),
							"delete_requested_at": knownvalue.Null(),
						}),
					})),
				},
			},
		},
	})
}

func TestAccECSDaemonTaskDefinitionsDataSource_sort(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ecs_daemon_task_definitions.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonTaskDefinitionsDataSourceConfig_sort(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("daemon_task_definitions"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrARN:         tfknownvalue.RegionalARNRegexp("ecs", regexache.MustCompile(`daemon-task-definition/`+rName+`.*:\d+$`)),
							names.AttrStatus:      knownvalue.StringExact("ACTIVE"),
							"registered_at":       knownvalue.NotNull(),
							"registered_by":       knownvalue.NotNull(),
							"delete_requested_at": knownvalue.Null(),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrARN:         tfknownvalue.RegionalARNRegexp("ecs", regexache.MustCompile(`daemon-task-definition/`+rName+`.*:\d+$`)),
							names.AttrStatus:      knownvalue.StringExact("ACTIVE"),
							"registered_at":       knownvalue.NotNull(),
							"registered_by":       knownvalue.NotNull(),
							"delete_requested_at": knownvalue.Null(),
						}),
					})),
				},
			},
		},
	})
}

// Config generators

func testAccDaemonTaskDefinitionsDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_daemon_task_definition" "test" {
  family = %[1]q

  container_definition {
    name      = "app"
    image     = "nginx:latest"
    cpu       = 256
    memory    = 512
    essential = true
  }
}

data "aws_ecs_daemon_task_definitions" "test" {
  family     = aws_ecs_daemon_task_definition.test.family
  status     = "ACTIVE"
  depends_on = [aws_ecs_daemon_task_definition.test]
}
`, rName)
}

func testAccDaemonTaskDefinitionsDataSourceConfig_familyPrefix(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_daemon_task_definition" "test" {
  family = %[1]q

  container_definition {
    name      = "app"
    image     = "nginx:latest"
    cpu       = 256
    memory    = 512
    essential = true
  }
}

resource "aws_ecs_daemon_task_definition" "other" {
  family = "%[1]s-other"

  container_definition {
    name      = "app"
    image     = "nginx:latest"
    cpu       = 256
    memory    = 512
    essential = true
  }
}

data "aws_ecs_daemon_task_definitions" "test" {
  family_prefix = %[1]q
  depends_on    = [aws_ecs_daemon_task_definition.test, aws_ecs_daemon_task_definition.other]
}
`, rName)
}

func testAccDaemonTaskDefinitionsDataSourceConfig_family(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_daemon_task_definition" "test" {
  family = %[1]q

  container_definition {
    name      = "app"
    image     = "nginx:latest"
    cpu       = 256
    memory    = 512
    essential = true
  }
}

data "aws_ecs_daemon_task_definitions" "test" {
  family     = aws_ecs_daemon_task_definition.test.family
  depends_on = [aws_ecs_daemon_task_definition.test]
}
`, rName)
}

func testAccDaemonTaskDefinitionsDataSourceConfig_revision(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_daemon_task_definition" "test" {
  family = %[1]q

  container_definition {
    name      = "app"
    image     = "nginx:latest"
    cpu       = 256
    memory    = 512
    essential = true
  }
}

data "aws_ecs_daemon_task_definitions" "test" {
  family_prefix = %[1]q
  revision      = "LAST_REGISTERED"
  depends_on    = [aws_ecs_daemon_task_definition.test]
}
`, rName)
}

func testAccDaemonTaskDefinitionsDataSourceConfig_sort(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_daemon_task_definition" "test1" {
  family = "%[1]s-1"

  container_definition {
    name      = "app"
    image     = "nginx:latest"
    cpu       = 256
    memory    = 512
    essential = true
  }
}

resource "aws_ecs_daemon_task_definition" "test2" {
  family = "%[1]s-2"

  container_definition {
    name      = "app"
    image     = "nginx:latest"
    cpu       = 256
    memory    = 512
    essential = true
  }
}

data "aws_ecs_daemon_task_definitions" "test" {
  family_prefix = %[1]q
  sort          = "ASC"
  depends_on    = [aws_ecs_daemon_task_definition.test1, aws_ecs_daemon_task_definition.test2]
}
`, rName)
}
