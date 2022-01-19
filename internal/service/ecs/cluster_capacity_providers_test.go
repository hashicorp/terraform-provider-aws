package ecs_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ecs"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfecs "github.com/hashicorp/terraform-provider-aws/internal/service/ecs"
)

func TestAccECSClusterCapacityProviders_basic(t *testing.T) {
	var cluster ecs.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_cluster_capacity_providers.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ecs.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterCapacityProvidersConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_ecs_cluster.test", &cluster),
					resource.TestCheckResourceAttr(resourceName, "capacity_providers.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "capacity_providers.*", "FARGATE"),
					resource.TestCheckResourceAttr(resourceName, "cluster_name", rName),
					resource.TestCheckResourceAttr(resourceName, "default_capacity_provider_strategy.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "default_capacity_provider_strategy.*", map[string]string{
						"base":              "1",
						"weight":            "100",
						"capacity_provider": "FARGATE",
					}),
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

func TestAccECSClusterCapacityProviders_disappears(t *testing.T) {
	var cluster ecs.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_cluster_capacity_providers.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ecs.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterCapacityProvidersConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_ecs_cluster.test", &cluster),
					acctest.CheckResourceDisappears(acctest.Provider, tfecs.ResourceCluster(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccECSClusterCapacityProviders_defaults(t *testing.T) {
	var cluster ecs.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_cluster_capacity_providers.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ecs.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterCapacityProvidersConfig_defaults(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_ecs_cluster.test", &cluster),
					resource.TestCheckResourceAttr(resourceName, "capacity_providers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cluster_name", rName),
					resource.TestCheckResourceAttr(resourceName, "default_capacity_provider_strategy.#", "0"),
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

func TestAccECSClusterCapacityProviders_update_capacityProviders(t *testing.T) {
	var cluster ecs.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_cluster_capacity_providers.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ecs.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterCapacityProvidersConfig_withCapacityProviders1(rName, "FARGATE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_ecs_cluster.test", &cluster),
					resource.TestCheckResourceAttr(resourceName, "capacity_providers.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "capacity_providers.*", "FARGATE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterCapacityProvidersConfig_withCapacityProviders2(rName, "FARGATE", "FARGATE_SPOT"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_ecs_cluster.test", &cluster),
					resource.TestCheckResourceAttr(resourceName, "capacity_providers.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "capacity_providers.*", "FARGATE"),
					resource.TestCheckTypeSetElemAttr(resourceName, "capacity_providers.*", "FARGATE_SPOT"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterCapacityProvidersConfig_withCapacityProviders0(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_ecs_cluster.test", &cluster),
					resource.TestCheckResourceAttr(resourceName, "capacity_providers.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterCapacityProvidersConfig_withCapacityProviders1(rName, "FARGATE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_ecs_cluster.test", &cluster),
					resource.TestCheckResourceAttr(resourceName, "capacity_providers.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "capacity_providers.*", "FARGATE"),
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

func TestAccECSClusterCapacityProviders_update_defaultCapacityProviderStrategy(t *testing.T) {
	var cluster ecs.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_cluster_capacity_providers.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ecs.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterCapacityProvidersConfig_withDefaultCapacityProviderStrategy1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_ecs_cluster.test", &cluster),
					resource.TestCheckResourceAttr(resourceName, "default_capacity_provider_strategy.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "default_capacity_provider_strategy.*", map[string]string{
						"base":              "1",
						"weight":            "100",
						"capacity_provider": "FARGATE",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterCapacityProvidersConfig_withDefaultCapacityProviderStrategy2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_ecs_cluster.test", &cluster),
					resource.TestCheckResourceAttr(resourceName, "default_capacity_provider_strategy.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "default_capacity_provider_strategy.*", map[string]string{
						"base":              "1",
						"weight":            "50",
						"capacity_provider": "FARGATE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "default_capacity_provider_strategy.*", map[string]string{
						"base":              "",
						"weight":            "50",
						"capacity_provider": "FARGATE_SPOT",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterCapacityProvidersConfig_withDefaultCapacityProviderStrategy3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_ecs_cluster.test", &cluster),
					resource.TestCheckResourceAttr(resourceName, "default_capacity_provider_strategy.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "default_capacity_provider_strategy.*", map[string]string{
						"base":              "2",
						"weight":            "25",
						"capacity_provider": "FARGATE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "default_capacity_provider_strategy.*", map[string]string{
						"base":              "",
						"weight":            "75",
						"capacity_provider": "FARGATE_SPOT",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterCapacityProvidersConfig_withDefaultCapacityProviderStrategy4(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_ecs_cluster.test", &cluster),
					resource.TestCheckResourceAttr(resourceName, "default_capacity_provider_strategy.#", "0"),
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

func testAccClusterCapacityProvidersConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_cluster_capacity_providers" "test" {
  cluster_name = aws_ecs_cluster.test.name

  capacity_providers = ["FARGATE"]

  default_capacity_provider_strategy {
    base              = 1
    weight            = 100
    capacity_provider = "FARGATE"
  }
}
`, rName)
}

func testAccClusterCapacityProvidersConfig_defaults(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_cluster_capacity_providers" "test" {
  cluster_name = aws_ecs_cluster.test.name
}
`, rName)
}

func testAccClusterCapacityProvidersConfig_withCapacityProviders0(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_cluster_capacity_providers" "test" {
  cluster_name = aws_ecs_cluster.test.name

  capacity_providers = []
}
`, rName)
}

func testAccClusterCapacityProvidersConfig_withCapacityProviders1(rName, provider1 string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_cluster_capacity_providers" "test" {
  cluster_name = aws_ecs_cluster.test.name

  capacity_providers = [%[2]q]
}
`, rName, provider1)
}

func testAccClusterCapacityProvidersConfig_withCapacityProviders2(rName, provider1, provider2 string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_cluster_capacity_providers" "test" {
  cluster_name = aws_ecs_cluster.test.name

  capacity_providers = [%[2]q, %[3]q]
}
`, rName, provider1, provider2)
}

func testAccClusterCapacityProvidersConfig_withDefaultCapacityProviderStrategy1(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_cluster_capacity_providers" "test" {
  cluster_name = aws_ecs_cluster.test.name

  capacity_providers = ["FARGATE", "FARGATE_SPOT"]

  default_capacity_provider_strategy {
    base              = 1
    weight            = 100
    capacity_provider = "FARGATE"
  }
}
`, rName)
}

func testAccClusterCapacityProvidersConfig_withDefaultCapacityProviderStrategy2(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_cluster_capacity_providers" "test" {
  cluster_name = aws_ecs_cluster.test.name

  capacity_providers = ["FARGATE", "FARGATE_SPOT"]

  default_capacity_provider_strategy {
    base              = 1
    weight            = 50
    capacity_provider = "FARGATE"
  }

  default_capacity_provider_strategy {
    weight            = 50
    capacity_provider = "FARGATE_SPOT"
  }
}
`, rName)
}

func testAccClusterCapacityProvidersConfig_withDefaultCapacityProviderStrategy3(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_cluster_capacity_providers" "test" {
  cluster_name = aws_ecs_cluster.test.name

  capacity_providers = ["FARGATE", "FARGATE_SPOT"]

  default_capacity_provider_strategy {
    base              = 2
    weight            = 25
    capacity_provider = "FARGATE"
  }

  default_capacity_provider_strategy {
    weight            = 75
    capacity_provider = "FARGATE_SPOT"
  }
}
`, rName)
}

func testAccClusterCapacityProvidersConfig_withDefaultCapacityProviderStrategy4(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_cluster_capacity_providers" "test" {
  cluster_name = aws_ecs_cluster.test.name

  capacity_providers = ["FARGATE", "FARGATE_SPOT"]
}
`, rName)
}
