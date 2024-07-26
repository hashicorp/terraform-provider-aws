// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emrserverless_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/emrserverless/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfemrserverless "github.com/hashicorp/terraform-provider-aws/internal/service/emrserverless"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEMRServerlessApplication_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var application types.Application
	resourceName := "aws_emrserverless_application.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "emr-serverless", regexache.MustCompile(`/applications/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "hive"),
					resource.TestCheckResourceAttr(resourceName, "architecture", "X86_64"),
					resource.TestCheckResourceAttr(resourceName, "release_label", "emr-6.6.0"),
					resource.TestCheckResourceAttr(resourceName, "auto_start_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "auto_start_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "auto_stop_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "auto_stop_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "auto_stop_configuration.0.idle_timeout_minutes", "15"),
					resource.TestCheckResourceAttr(resourceName, "initial_capacity.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "image_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func TestAccEMRServerlessApplication_arch(t *testing.T) {
	ctx := acctest.Context(t)
	var application types.Application
	resourceName := "aws_emrserverless_application.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_arch(rName, "ARM64"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "architecture", "ARM64"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccApplicationConfig_arch(rName, "X86_64"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "architecture", "X86_64"),
				),
			},
		},
	})
}

func TestAccEMRServerlessApplication_releaseLabel(t *testing.T) {
	ctx := acctest.Context(t)
	var application types.Application
	resourceName := "aws_emrserverless_application.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_releaseLabel(rName, "emr-6.10.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "release_label", "emr-6.10.0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccApplicationConfig_releaseLabel(rName, "emr-6.11.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "release_label", "emr-6.11.0"),
				),
			},
		},
	})
}

func TestAccEMRServerlessApplication_initialCapacity(t *testing.T) {
	ctx := acctest.Context(t)
	var application types.Application
	resourceName := "aws_emrserverless_application.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_initialCapacity(rName, "2 vCPU"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "initial_capacity.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_capacity.0.initial_capacity_type", "HiveDriver"),
					resource.TestCheckResourceAttr(resourceName, "initial_capacity.0.initial_capacity_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_capacity.0.initial_capacity_config.0.worker_count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_capacity.0.initial_capacity_config.0.worker_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_capacity.0.initial_capacity_config.0.worker_configuration.0.cpu", "2 vCPU"),
					resource.TestCheckResourceAttr(resourceName, "initial_capacity.0.initial_capacity_config.0.worker_configuration.0.memory", "10 GB"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccApplicationConfig_initialCapacity(rName, "4 vCPU"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "initial_capacity.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_capacity.0.initial_capacity_type", "HiveDriver"),
					resource.TestCheckResourceAttr(resourceName, "initial_capacity.0.initial_capacity_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_capacity.0.initial_capacity_config.0.worker_count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_capacity.0.initial_capacity_config.0.worker_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_capacity.0.initial_capacity_config.0.worker_configuration.0.cpu", "4 vCPU"),
					resource.TestCheckResourceAttr(resourceName, "initial_capacity.0.initial_capacity_config.0.worker_configuration.0.memory", "10 GB"),
				),
			},
		},
	})
}

func TestAccEMRServerlessApplication_imageConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	var application types.Application
	resourceName := "aws_emrserverless_application.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	firstVersionRegex := regexache.MustCompile(`1\.0\.0`)
	secondVersionRegex := regexache.MustCompile(`1\.0\.1`)

	firstImageConfig, err := testAccApplicationConfig_imageConfiguration(rName, "1.0.0", "1.0.1", "1.0.0")
	if err != nil {
		t.Error(err)
	}

	secondImageConfig, err := testAccApplicationConfig_imageConfiguration(rName, "1.0.0", "1.0.1", "1.0.1")
	if err != nil {
		t.Error(err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: firstImageConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "image_configuration.#", acctest.Ct1),
					resource.TestMatchResourceAttr(resourceName, "image_configuration.0.image_uri", firstVersionRegex),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: secondImageConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "image_configuration.#", acctest.Ct1),
					resource.TestMatchResourceAttr(resourceName, "image_configuration.0.image_uri", secondVersionRegex),
				),
			},
		},
	})
}

func TestAccEMRServerlessApplication_interactiveConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var application types.Application
	resourceName := "aws_emrserverless_application.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_interactiveConfiguration(rName, true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "interactive_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "interactive_configuration.0.livy_endpoint_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "interactive_configuration.0.studio_enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccApplicationConfig_interactiveConfiguration(rName, true, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "interactive_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "interactive_configuration.0.livy_endpoint_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "interactive_configuration.0.studio_enabled", acctest.CtFalse),
				),
			},
			{
				Config: testAccApplicationConfig_interactiveConfiguration(rName, false, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "interactive_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "interactive_configuration.0.livy_endpoint_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "interactive_configuration.0.studio_enabled", acctest.CtTrue),
				),
			},
			{
				Config: testAccApplicationConfig_interactiveConfiguration(rName, false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "interactive_configuration.#", acctest.Ct1),
					resource.TestCheckNoResourceAttr(resourceName, "interactive_configuration.0.livy_endpoint_enabled"),
					resource.TestCheckNoResourceAttr(resourceName, "interactive_configuration.0.studio_enabled"),
				),
			},
		},
	})
}

func TestAccEMRServerlessApplication_maxCapacity(t *testing.T) {
	ctx := acctest.Context(t)
	var application types.Application
	resourceName := "aws_emrserverless_application.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_maxCapacity(rName, "2 vCPU"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "maximum_capacity.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "maximum_capacity.0.cpu", "2 vCPU"),
					resource.TestCheckResourceAttr(resourceName, "maximum_capacity.0.memory", "10 GB"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccApplicationConfig_maxCapacity(rName, "4 vCPU"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "maximum_capacity.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "maximum_capacity.0.cpu", "4 vCPU"),
					resource.TestCheckResourceAttr(resourceName, "maximum_capacity.0.memory", "10 GB")),
			},
		},
	})
}

func TestAccEMRServerlessApplication_network(t *testing.T) {
	ctx := acctest.Context(t)
	var application types.Application
	resourceName := "aws_emrserverless_application.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_network(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.security_group_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.subnet_ids.#", acctest.Ct2),
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

func TestAccEMRServerlessApplication_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var application types.Application
	resourceName := "aws_emrserverless_application.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfemrserverless.ResourceApplication(), resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfemrserverless.ResourceApplication(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEMRServerlessApplication_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var application types.Application
	resourceName := "aws_emrserverless_application.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
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
				Config: testAccApplicationConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccApplicationConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckApplicationExists(ctx context.Context, resourceName string, application *types.Application) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EMRServerlessClient(ctx)

		output, err := tfemrserverless.FindApplicationByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("EMR Serverless Application (%s) not found", rs.Primary.ID)
		}

		*application = *output

		return nil
	}
}

func testAccCheckApplicationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EMRServerlessClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_emrserverless_application" {
				continue
			}

			_, err := tfemrserverless.FindApplicationByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EMR Serverless Application %s still exists", rs.Primary.ID)
		}
		return nil
	}
}

func testAccApplicationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_emrserverless_application" "test" {
  name          = %[1]q
  release_label = "emr-6.6.0"
  type          = "hive"
}
`, rName)
}

func testAccApplicationConfig_releaseLabel(rName string, rl string) string {
	return fmt.Sprintf(`
resource "aws_emrserverless_application" "test" {
  name          = %[1]q
  release_label = %[2]q
  type          = "spark"
}
`, rName, rl)
}

func testAccApplicationConfig_initialCapacity(rName, cpu string) string {
	return fmt.Sprintf(`
resource "aws_emrserverless_application" "test" {
  name          = %[1]q
  release_label = "emr-6.6.0"
  type          = "hive"

  initial_capacity {
    initial_capacity_type = "HiveDriver"

    initial_capacity_config {
      worker_count = 1
      worker_configuration {
        cpu    = %[2]q
        memory = "10 GB"
      }
    }
  }
}
`, rName, cpu)
}

func testAccApplicationConfig_interactiveConfiguration(rName string, livyEndpointEnabled, studioEnabled bool) string {
	return fmt.Sprintf(`
resource "aws_emrserverless_application" "test" {
  name          = %[1]q
  release_label = "emr-7.1.0"
  type          = "spark"
  interactive_configuration {
    livy_endpoint_enabled = %[2]t
    studio_enabled        = %[3]t
  }
}
`, rName, livyEndpointEnabled, studioEnabled)
}

func testAccApplicationConfig_maxCapacity(rName, cpu string) string {
	return fmt.Sprintf(`
resource "aws_emrserverless_application" "test" {
  name          = %[1]q
  release_label = "emr-6.6.0"
  type          = "hive"

  maximum_capacity {
    cpu    = %[2]q
    memory = "10 GB"
  }
}
`, rName, cpu)
}

func testAccApplicationConfig_network(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_security_group" "test" {
  name        = %[1]q
  description = "Allow all inbound traffic"
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    self      = true
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_emrserverless_application" "test" {
  name          = %[1]q
  release_label = "emr-6.6.0"
  type          = "hive"

  network_configuration {
    security_group_ids = [aws_security_group.test.id]
    subnet_ids         = aws_subnet.test[*].id
  }
}
`, rName))
}

func testAccApplicationConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_emrserverless_application" "test" {
  name          = %[1]q
  release_label = "emr-6.6.0"
  type          = "hive"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccApplicationConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_emrserverless_application" "test" {
  name          = %[1]q
  release_label = "emr-6.6.0"
  type          = "hive"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccApplicationConfig_arch(rName, arch string) string {
	return fmt.Sprintf(`
resource "aws_emrserverless_application" "test" {
  name          = %[1]q
  release_label = "emr-6.6.0"
  type          = "hive"
  architecture  = %[2]q
}
`, rName, arch)
}

// At the time of writing, the AWS EMR Serverless API returns a 500 error if you try to create an EMR Serverless
// application with an image from a public emr repo, and so we need to build an image and put it in a temporary
// repo in order to run the test
func testAccApplicationConfig_imageConfiguration(rName, firstImageVersion, secondImageVersion, selectedImageVersion string) (string, error) {
	if firstImageVersion == secondImageVersion {
		return "", fmt.Errorf("firstImageVersion and secondImageVersion cannot be equal. Was given %[1]q for both", firstImageVersion)
	}

	if selectedImageVersion != firstImageVersion && selectedImageVersion != secondImageVersion {
		return "", fmt.Errorf("selectedImageVersion must be equal to firstImageVersion or secondImageVersion (%[1]q or %[2]q). Was given %[3]q", firstImageVersion, secondImageVersion, selectedImageVersion)
	}

	selectedVersionResourceName := "test_version1"
	if selectedImageVersion != firstImageVersion {
		selectedVersionResourceName = "test_version2"
	}

	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_partition" "current" {}

resource "aws_ecr_repository" "test" {
  name         = %[1]q
  force_delete = true
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block              = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  map_public_ip_on_launch = true
  vpc_id                  = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_security_group" "test" {
  vpc_id = aws_vpc.test.id

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    protocol    = "-1"
    to_port     = 0
  }

  ingress {
    from_port = 0
    protocol  = -1
    self      = true
    to_port   = 0
  }
}

resource "aws_iam_role" "test" {
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}"
      }
      Sid = ""
    }]
  })
  name = %[1]q
}

resource "aws_iam_role_policy_attachment" "AmazonSSMManagedInstanceCore" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSSMManagedInstanceCore"
  role       = aws_iam_role.test.name
}

resource "aws_iam_role_policy_attachment" "EC2InstanceProfileForImageBuilder" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/EC2InstanceProfileForImageBuilder"
  role       = aws_iam_role.test.name
}

resource "aws_iam_role_policy_attachment" "EC2InstanceProfileForImageBuilderECRContainerBuilds" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/EC2InstanceProfileForImageBuilderECRContainerBuilds"
  role       = aws_iam_role.test.name
}

resource "aws_iam_instance_profile" "test" {
  name = aws_iam_role.test.name
  role = aws_iam_role.test.name

  depends_on = [
    aws_iam_role_policy_attachment.AmazonSSMManagedInstanceCore,
    aws_iam_role_policy_attachment.EC2InstanceProfileForImageBuilderECRContainerBuilds
  ]
}

resource "aws_imagebuilder_container_recipe" "test_version1" {
  name              = "%[1]s_version1"
  container_type    = "DOCKER"
  parent_image      = "public.ecr.aws/emr-serverless/hive/emr-6.9.0"
  version           = %[3]q
  platform_override = "Linux"

  component {
    component_arn = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:component/hello-world-linux/x.x.x"
  }

  dockerfile_template_data = <<EOF
FROM {{{ imagebuilder:parentImage }}}
EOF

  target_repository {
    repository_name = aws_ecr_repository.test.name
    service         = "ECR"
  }
}

resource "aws_imagebuilder_container_recipe" "test_version2" {
  name              = "%[1]s_version2"
  container_type    = "DOCKER"
  parent_image      = "public.ecr.aws/emr-serverless/hive/emr-6.9.0"
  version           = %[4]q
  platform_override = "Linux"

  component {
    component_arn = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:component/hello-world-linux/x.x.x"
  }

  dockerfile_template_data = <<EOF
FROM {{{ imagebuilder:parentImage }}}
EOF

  target_repository {
    repository_name = aws_ecr_repository.test.name
    service         = "ECR"
  }
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
  security_group_ids    = [aws_default_security_group.test.id]
  subnet_id             = aws_subnet.test.id

  depends_on = [aws_default_route_table.test]
}

resource "aws_imagebuilder_image" "test_version1" {
  container_recipe_arn             = aws_imagebuilder_container_recipe.test_version1.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
}

resource "aws_imagebuilder_image" "test_version2" {
  container_recipe_arn             = aws_imagebuilder_container_recipe.test_version2.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
}

resource "aws_emrserverless_application" "test" {
  name          = %[1]q
  release_label = "emr-6.9.0"
  type          = "hive"

  image_configuration {
    image_uri = "${aws_ecr_repository.test.repository_url}:${replace(aws_imagebuilder_image.%[2]s.version, "/", "-")}"
  }
}
`, rName, selectedVersionResourceName, firstImageVersion, secondImageVersion), nil
}
