// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apprunner_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/apprunner"
	"github.com/aws/aws-sdk-go-v2/service/apprunner/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapprunner "github.com/hashicorp/terraform-provider-aws/internal/service/apprunner"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppRunnerService_ImageRepository_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_imageRepository(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrServiceName, rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "apprunner", regexache.MustCompile(fmt.Sprintf(`service/%s/.+`, rName))),
					acctest.MatchResourceAttrRegionalARN(resourceName, "auto_scaling_configuration_arn", "apprunner", regexache.MustCompile(`autoscalingconfiguration/DefaultConfiguration/1/.+`)),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.0.protocol", string(types.HealthCheckProtocolTcp)),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.0.path", "/"),
					// Only check the following attribute values for health_check and instance configurations
					// are set as their defaults differ in the API documentation and API itself
					resource.TestCheckResourceAttrSet(resourceName, "health_check_configuration.0.interval"),
					resource.TestCheckResourceAttrSet(resourceName, "health_check_configuration.0.timeout"),
					resource.TestCheckResourceAttrSet(resourceName, "health_check_configuration.0.healthy_threshold"),
					resource.TestCheckResourceAttrSet(resourceName, "health_check_configuration.0.unhealthy_threshold"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "instance_configuration.0.cpu"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_configuration.0.memory"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.instance_role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.egress_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.egress_configuration.0.egress_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.egress_configuration.0.vpc_connector_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.ingress_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.ingress_configuration.0.is_publicly_accessible", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.ip_address_type", "IPV4"),
					resource.TestCheckResourceAttr(resourceName, "observability_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "service_id"),
					resource.TestCheckResourceAttrSet(resourceName, "service_url"),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.0.authentication_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.0.auto_deployments_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.0.code_repository.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.0.image_repository.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.0.image_repository.0.image_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.0.image_repository.0.image_configuration.0.port", "80"),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.0.image_repository.0.image_configuration.0.runtime_environment_secrets.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.0.image_repository.0.image_configuration.0.runtime_environment_variables.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.0.image_repository.0.image_identifier", "public.ecr.aws/nginx/nginx:latest"),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.0.image_repository.0.image_repository_type", string(types.ImageRepositoryTypeEcrPublic)),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.ServiceStatusRunning)),
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

func TestAccAppRunnerService_ImageRepository_autoScaling(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_service.test"
	autoScalingResourceName := "aws_apprunner_auto_scaling_configuration_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_imageRepository(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName),
				),
			},
			{
				Config: testAccServiceConfig_ImageRepository_autoScalingConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "auto_scaling_configuration_arn", autoScalingResourceName, names.AttrARN),
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

func TestAccAppRunnerService_ImageRepository_encryption(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_service.test"
	kmsResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_ImageRepository_encryptionConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_configuration.0.kms_key", kmsResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Test resource recreation; EncryptionConfiguration (or lack thereof) Forces New resource
				Config: testAccServiceConfig_imageRepository(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccAppRunnerService_ImageRepository_healthCheck(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_ImageRepository_healthCheckConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.0.healthy_threshold", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.0.interval", "5"),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.0.protocol", string(types.HealthCheckProtocolTcp)),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.0.timeout", "5"),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.0.unhealthy_threshold", "5"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccServiceConfig_ImageRepository_updateHealthCheckConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.0.healthy_threshold", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.0.interval", "5"),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.0.protocol", string(types.HealthCheckProtocolTcp)),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.0.timeout", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.0.unhealthy_threshold", acctest.Ct4),
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

func TestAccAppRunnerService_ImageRepository_instance_NoInstanceRole(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_ImageRepository_InstanceConfiguration_noInstanceRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.cpu", "1024"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.instance_role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.memory", "3072"),
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

func TestAccAppRunnerService_ImageRepository_instance_Update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_service.test"
	roleResourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_ImageRepository_instanceConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.cpu", "1024"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_configuration.0.instance_role_arn", roleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.memory", "3072"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccServiceConfig_ImageRepository_updateInstanceConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.cpu", "2048"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_configuration.0.instance_role_arn", roleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.memory", "4096"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccServiceConfig_imageRepository(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.cpu", "2048"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.memory", "4096"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_configuration.0.instance_role_arn"), // The IAM Role is not unset
				),
			},
		},
	})
}
func TestAccAppRunnerService_ImageRepository_instance_Update1(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_service.test"
	roleResourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_ImageRepository_instanceConfiguration1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.cpu", "256"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_configuration.0.instance_role_arn", roleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.memory", "512"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccServiceConfig_ImageRepository_updateInstanceConfiguration1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.cpu", "4096"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_configuration.0.instance_role_arn", roleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.memory", "12288"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccServiceConfig_imageRepository(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.cpu", "4096"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.memory", "12288"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_configuration.0.instance_role_arn"), // The IAM Role is not unset
				),
			},
		},
	})
}

func TestAccAppRunnerService_ImageRepository_networkConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_service.test"
	vpcConnectorResourceName := "aws_apprunner_vpc_connector.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_ImageRepository_networkConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.egress_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.egress_configuration.0.egress_type", "VPC"),
					resource.TestCheckResourceAttrPair(resourceName, "network_configuration.0.egress_configuration.0.vpc_connector_arn", vpcConnectorResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.ingress_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.ingress_configuration.0.is_publicly_accessible", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, "service_url"),
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

func TestAccAppRunnerService_ImageRepository_observabilityConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_service.test"
	observabilityConfigurationResourceName := "aws_apprunner_observability_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_ImageRepository_observabilityConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "observability_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "observability_configuration.0.observability_configuration_arn", observabilityConfigurationResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "observability_configuration.0.observability_enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccServiceConfig_ImageRepository_observabilityConfiguration_disabled(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "observability_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "observability_configuration.0.observability_enabled", acctest.CtFalse),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/19469
func TestAccAppRunnerService_ImageRepository_runtimeEnvironmentVars(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_ImageRepository_runtimeEnvVars(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.0.image_repository.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.0.image_repository.0.image_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.0.image_repository.0.image_configuration.0.runtime_environment_variables.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.0.image_repository.0.image_configuration.0.runtime_environment_variables.APP_NAME", rName),
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

func TestAccAppRunnerService_ImageRepository_runtimeEnvironmentSecrets(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_service.test"
	ssmParameterResourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_ImageRepository_runtimeEnvSecrets(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.0.image_repository.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.0.image_repository.0.image_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.0.image_repository.0.image_configuration.0.runtime_environment_secrets.%", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "source_configuration.0.image_repository.0.image_configuration.0.runtime_environment_secrets.SSM_PARAMETER", ssmParameterResourceName, names.AttrARN),
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

func TestAccAppRunnerService_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_imageRepository(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapprunner.ResourceService(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAppRunnerService_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName),
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
				Config: testAccServiceConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccServiceConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckServiceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_apprunner_service" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).AppRunnerClient(ctx)

			_, err := tfapprunner.FindServiceByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("App Runner Service %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckServiceExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppRunnerClient(ctx)

		_, err := tfapprunner.FindServiceByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AppRunnerClient(ctx)

	input := &apprunner.ListServicesInput{}

	_, err := conn.ListServices(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccServiceConfig_imageRepository(rName string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_service" "test" {
  service_name = %[1]q
  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "80"
      }
      image_identifier      = "public.ecr.aws/nginx/nginx:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }
}
`, rName)
}

func testAccServiceConfig_ImageRepository_runtimeEnvVars(rName string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_service" "test" {
  service_name = %[1]q
  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "80"
        runtime_environment_variables = {
          APP_NAME = %[1]q
        }
      }
      image_identifier      = "public.ecr.aws/nginx/nginx:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }
}
`, rName)
}

func testAccServiceConfig_ImageRepository_runtimeEnvSecrets(rName string) string {
	return acctest.ConfigCompose(
		testAccIAMRole(rName),
		fmt.Sprintf(`
resource "aws_ssm_parameter" "test" {
  name  = %[1]q
  type  = "String"
  value = "test"
}

resource "aws_iam_role_policy" "test_policy" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "ssm:GetParameters",
        ]
        Effect = "Allow"
        Resource = [
          aws_ssm_parameter.test.arn
        ]
      },
    ]
  })
}

resource "aws_apprunner_service" "test" {
  service_name = %[1]q
  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "80"
        runtime_environment_secrets = {
          SSM_PARAMETER = aws_ssm_parameter.test.arn
        }
      }
      image_identifier      = "public.ecr.aws/nginx/nginx:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }
  instance_configuration {
    cpu               = "1 vCPU"
    instance_role_arn = aws_iam_role.test.arn
    memory            = "3 GB"
  }
}
`, rName))
}

func testAccServiceConfig_ImageRepository_autoScalingConfiguration(rName string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_auto_scaling_configuration_version" "test" {
  auto_scaling_configuration_name = %[1]q
}

resource "aws_apprunner_service" "test" {
  auto_scaling_configuration_arn = aws_apprunner_auto_scaling_configuration_version.test.arn

  service_name = %[1]q

  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "80"
      }
      image_identifier      = "public.ecr.aws/nginx/nginx:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }
}
`, rName)
}

func testAccServiceConfig_ImageRepository_encryptionConfiguration(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
  description             = %[1]q
}

resource "aws_apprunner_service" "test" {
  service_name = %[1]q

  encryption_configuration {
    kms_key = aws_kms_key.test.arn
  }

  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "80"
      }
      image_identifier      = "public.ecr.aws/nginx/nginx:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }
}
`, rName)
}

func testAccServiceConfig_ImageRepository_healthCheckConfiguration(rName string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_service" "test" {
  service_name = %[1]q

  health_check_configuration {
    healthy_threshold = 2
    timeout           = 5
  }

  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "80"
      }
      image_identifier      = "public.ecr.aws/nginx/nginx:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }
}
`, rName)
}

func testAccServiceConfig_ImageRepository_updateHealthCheckConfiguration(rName string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_service" "test" {
  service_name = %[1]q

  health_check_configuration {
    healthy_threshold   = 2
    timeout             = 10
    unhealthy_threshold = 4
  }

  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "80"
      }
      image_identifier      = "public.ecr.aws/nginx/nginx:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }
}
`, rName)
}

func testAccServiceConfig_ImageRepository_networkConfiguration(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
  name   = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_apprunner_vpc_connector" "test" {
  vpc_connector_name = %[1]q
  subnets            = aws_subnet.test[*].id
  security_groups    = [aws_security_group.test.id]
}

resource "aws_apprunner_service" "test" {
  service_name = %[1]q

  network_configuration {
    egress_configuration {
      egress_type       = "VPC"
      vpc_connector_arn = aws_apprunner_vpc_connector.test.arn
    }

    ingress_configuration {
      is_publicly_accessible = false
    }
  }

  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "80"
      }
      image_identifier      = "public.ecr.aws/nginx/nginx:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }
}
`, rName))
}

func testAccServiceConfig_ImageRepository_observabilityConfiguration(rName string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_service" "test" {
  service_name = %[1]q

  health_check_configuration {
    healthy_threshold = 2
    timeout           = 5
  }

  observability_configuration {
    observability_configuration_arn = aws_apprunner_observability_configuration.test.arn
    observability_enabled           = true
  }

  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "80"
      }
      image_identifier      = "public.ecr.aws/nginx/nginx:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }
}

resource "aws_apprunner_observability_configuration" "test" {
  observability_configuration_name = %[1]q

  trace_configuration {
    vendor = "AWSXRAY"
  }
}
`, rName)
}

func testAccServiceConfig_ImageRepository_observabilityConfiguration_disabled(rName string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_service" "test" {
  service_name = %[1]q

  health_check_configuration {
    healthy_threshold = 2
    timeout           = 5
  }

  observability_configuration {
    observability_enabled = false
  }

  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "80"
      }
      image_identifier      = "public.ecr.aws/nginx/nginx:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }
}
`, rName)
}

func testAccIAMRole(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "tasks.apprunner.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}
`, rName)
}

func testAccServiceConfig_ImageRepository_instanceConfiguration(rName string) string {
	return acctest.ConfigCompose(
		testAccIAMRole(rName),
		fmt.Sprintf(`
resource "aws_apprunner_service" "test" {
  service_name = %[1]q

  instance_configuration {
    cpu               = "1 vCPU"
    instance_role_arn = aws_iam_role.test.arn
    memory            = "3 GB"
  }

  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "80"
      }
      image_identifier      = "public.ecr.aws/nginx/nginx:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }
}
`, rName))
}

func testAccServiceConfig_ImageRepository_instanceConfiguration1(rName string) string {
	return acctest.ConfigCompose(
		testAccIAMRole(rName),
		fmt.Sprintf(`
resource "aws_apprunner_service" "test" {
  service_name = %[1]q

  instance_configuration {
    cpu               = "0.25 vCPU"
    instance_role_arn = aws_iam_role.test.arn
    memory            = "0.5 GB"
  }

  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "80"
      }
      image_identifier      = "public.ecr.aws/nginx/nginx:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }
}
`, rName))
}

func testAccServiceConfig_ImageRepository_InstanceConfiguration_noInstanceRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_service" "test" {
  service_name = %[1]q

  instance_configuration {
    cpu    = "1 vCPU"
    memory = "3 GB"
  }

  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "80"
      }
      image_identifier      = "public.ecr.aws/nginx/nginx:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }
}
`, rName)
}

func testAccServiceConfig_ImageRepository_updateInstanceConfiguration(rName string) string {
	return acctest.ConfigCompose(
		testAccIAMRole(rName),
		fmt.Sprintf(`
resource "aws_apprunner_service" "test" {
  service_name = %[1]q

  instance_configuration {
    cpu               = "2 vCPU"
    instance_role_arn = aws_iam_role.test.arn
    memory            = "4 GB"
  }

  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "80"
      }
      image_identifier      = "public.ecr.aws/nginx/nginx:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }
}
`, rName))
}

func testAccServiceConfig_ImageRepository_updateInstanceConfiguration1(rName string) string {
	return acctest.ConfigCompose(
		testAccIAMRole(rName),
		fmt.Sprintf(`
resource "aws_apprunner_service" "test" {
  service_name = %[1]q

  instance_configuration {
    cpu               = "4 vCPU"
    instance_role_arn = aws_iam_role.test.arn
    memory            = "12 GB"
  }

  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "80"
      }
      image_identifier      = "public.ecr.aws/nginx/nginx:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }
}
`, rName))
}

func testAccServiceConfig_tags1(rName string, tagKey1 string, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_service" "test" {
  service_name = %[1]q
  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "80"
      }
      image_identifier      = "public.ecr.aws/nginx/nginx:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccServiceConfig_tags2(rName string, tagKey1 string, tagValue1 string, tagKey2 string, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_service" "test" {
  service_name = %[1]q
  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "80"
      }
      image_identifier      = "public.ecr.aws/nginx/nginx:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
