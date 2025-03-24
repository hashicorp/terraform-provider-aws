// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccDomain_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, rName),
					resource.TestCheckResourceAttr(resourceName, "auth_mode", "IAM"),
					resource.TestCheckResourceAttr(resourceName, "app_network_access_type", "PublicInternetOnly"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_space_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.auto_mount_home_efs", "Enabled"),
					resource.TestCheckResourceAttrPair(resourceName, "default_user_settings.0.execution_role", "aws_iam_role.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "sagemaker", regexache.MustCompile(`domain/.+`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "tag_propagation", "DISABLED"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, "aws_vpc.test", names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrURL),
					resource.TestCheckResourceAttrSet(resourceName, "home_efs_file_system_id"),
					resource.TestCheckResourceAttr(resourceName, "domain_settings.#", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
		},
	})
}

func testAccDomain_domainSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_domainSettings(rName, "DISABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "domain_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "domain_settings.0.execution_role_identity_config", "DISABLED"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
			{
				Config: testAccDomainConfig_domainSettings(rName, "USER_PROFILE_NAME"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "domain_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "domain_settings.0.execution_role_identity_config", "USER_PROFILE_NAME"),
				),
			},
		},
	})
}

func testAccDomain_domainSettingsDockerSettingsUpdated(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_domainSettingsDockerSettings(rName, "DISABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "domain_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "domain_settings.0.docker_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "domain_settings.0.docker_settings.0.enable_docker_access", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "domain_settings.0.docker_settings.0.vpc_only_trusted_accounts.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
			{
				Config: testAccDomainConfig_domainSettingsDockerSettings(rName, "ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "domain_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "domain_settings.0.docker_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "domain_settings.0.docker_settings.0.enable_docker_access", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "domain_settings.0.docker_settings.0.vpc_only_trusted_accounts.#", "1"),
				),
			},
		},
	})
}

func testAccDomain_kms(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_kms(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, "aws_kms_key.test", names.AttrKeyID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
		},
	})
}

func testAccDomain_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
			{
				Config: testAccDomainConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccDomainConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccDomain_defaultUserSettingsSecurityGroupUpdated(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_defaultUserSettingsSecurityGroupUpdated(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.security_groups.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
			{
				Config: testAccDomainConfig_defaultUserSettingsSecurityGroupUpdated(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.security_groups.#", "2"),
				),
			},
		},
	})
}

func testAccDomain_sharingSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_sharingSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.sharing_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.sharing_settings.0.notebook_output_option", "Allowed"),
					resource.TestCheckResourceAttrPair(resourceName, "default_user_settings.0.sharing_settings.0.s3_kms_key_id", "aws_kms_key.test", names.AttrKeyID),
					resource.TestCheckResourceAttrSet(resourceName, "default_user_settings.0.sharing_settings.0.s3_output_path"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
		},
	})
}

func testAccDomain_canvasAppSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_canvasAppSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.canvas_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.canvas_app_settings.0.time_series_forecasting_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.canvas_app_settings.0.time_series_forecasting_settings.0.status", "DISABLED"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
		},
	})
}

func testAccDomain_modelRegisterSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_modelRegisterSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.canvas_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.canvas_app_settings.0.model_register_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.canvas_app_settings.0.model_register_settings.0.status", "DISABLED"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
		},
	})
}

func testAccDomain_generativeAiSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_generativeAiSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.canvas_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.canvas_app_settings.0.generative_ai_settings.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "default_user_settings.0.canvas_app_settings.0.generative_ai_settings.0.amazon_bedrock_role_arn", "aws_iam_role.test", names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
		},
	})
}

func testAccDomain_kendraSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_kendraSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.canvas_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.canvas_app_settings.0.kendra_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.canvas_app_settings.0.kendra_settings.0.status", "DISABLED"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
		},
	})
}

func testAccDomain_directDeploySettings(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_directDeploySettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.canvas_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.canvas_app_settings.0.direct_deploy_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.canvas_app_settings.0.direct_deploy_settings.0.status", "DISABLED"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
		},
	})
}

func testAccDomain_emrServerlessSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_emrServerlessSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.canvas_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.canvas_app_settings.0.emr_serverless_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.canvas_app_settings.0.emr_serverless_settings.0.status", "ENABLED"),
					resource.TestCheckResourceAttrPair(resourceName, "default_user_settings.0.canvas_app_settings.0.emr_serverless_settings.0.execution_role_arn", "aws_iam_role.test", names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
		},
	})
}

func testAccDomain_identityProviderOAuthSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_identityProviderOAuthSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.canvas_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.canvas_app_settings.0.identity_provider_oauth_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.canvas_app_settings.0.identity_provider_oauth_settings.0.status", "DISABLED"),
					resource.TestCheckResourceAttrPair(resourceName, "default_user_settings.0.canvas_app_settings.0.identity_provider_oauth_settings.0.secret_arn", "aws_secretsmanager_secret.test", names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
		},
	})
}

func testAccDomain_workspaceSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_workspaceSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.canvas_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.canvas_app_settings.0.workspace_settings.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "default_user_settings.0.canvas_app_settings.0.workspace_settings.0.s3_artifact_path"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
		},
	})
}

func testAccDomain_tensorboardAppSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_tensorBoardAppSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.tensor_board_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.tensor_board_app_settings.0.default_resource_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.tensor_board_app_settings.0.default_resource_spec.0.instance_type", "ml.t3.micro"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
		},
	})
}

func testAccDomain_tensorboardAppSettingsWithImage(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_tensorBoardAppSettingsImage(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.tensor_board_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.tensor_board_app_settings.0.default_resource_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.tensor_board_app_settings.0.default_resource_spec.0.instance_type", "ml.t3.micro"),
					resource.TestCheckResourceAttrPair(resourceName, "default_user_settings.0.tensor_board_app_settings.0.default_resource_spec.0.sagemaker_image_arn", "aws_sagemaker_image.test", names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
		},
	})
}

func testAccDomain_rSessionAppSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_rSessionAppSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.r_session_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.r_session_app_settings.0.default_resource_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.r_session_app_settings.0.default_resource_spec.0.instance_type", "ml.t3.micro"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
		},
	})
}

func testAccDomain_rStudioServerProAppSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_rStudioServerProAppSettings(rName, "ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.r_studio_server_pro_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.r_studio_server_pro_app_settings.0.access_status", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.r_studio_server_pro_app_settings.0.user_group", "R_STUDIO_ADMIN"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
			{
				Config: testAccDomainConfig_rStudioServerProAppSettings(rName, "DISABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.r_studio_server_pro_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.r_studio_server_pro_app_settings.0.access_status", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.r_studio_server_pro_app_settings.0.user_group", "R_STUDIO_ADMIN"),
				),
			},
			{
				Config: testAccDomainConfig_rStudioServerProAppSettings(rName, "ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.r_studio_server_pro_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.r_studio_server_pro_app_settings.0.access_status", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.r_studio_server_pro_app_settings.0.user_group", "R_STUDIO_ADMIN"),
				),
			},
		},
	})
}

func testAccDomain_rStudioServerProDomainSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_rStudioServerProDomainSettings(rName, "https://connect.domain.com", "https://package.domain.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "domain_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "domain_settings.0.r_studio_server_pro_domain_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "domain_settings.0.r_studio_server_pro_domain_settings.0.r_studio_connect_url", "https://connect.domain.com"),
					resource.TestCheckResourceAttr(resourceName, "domain_settings.0.r_studio_server_pro_domain_settings.0.r_studio_package_manager_url", "https://package.domain.com"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
		},
	})
}

func testAccDomain_rStudioDomainDisabledNetworkUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_rStudioDomainDisabledNetworkUpdate(rName, "PublicInternetOnly"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "domain_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "app_network_access_type", "PublicInternetOnly"),
				),
			},
			{
				Config: testAccDomainConfig_rStudioDomainDisabledNetworkUpdate(rName, "VpcOnly"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "domain_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "app_network_access_type", "VpcOnly")),
			},
		},
	})
}

func testAccDomain_kernelGatewayAppSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_kernelGatewayAppSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.kernel_gateway_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.kernel_gateway_app_settings.0.default_resource_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.kernel_gateway_app_settings.0.default_resource_spec.0.instance_type", "ml.t3.micro"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
		},
	})
}

func testAccDomain_codeEditorAppSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_codeEditorAppSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.code_editor_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.code_editor_app_settings.0.default_resource_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.code_editor_app_settings.0.default_resource_spec.0.instance_type", "ml.t3.micro"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
		},
	})
}

func testAccDomain_codeEditorAppSettings_customImage(t *testing.T) {
	ctx := acctest.Context(t)
	if os.Getenv("SAGEMAKER_IMAGE_VERSION_BASE_IMAGE") == "" {
		t.Skip("Environment variable SAGEMAKER_IMAGE_VERSION_BASE_IMAGE is not set")
	}

	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"
	baseImage := os.Getenv("SAGEMAKER_IMAGE_VERSION_BASE_IMAGE")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_codeEditorAppSettingsCustomImage(rName, baseImage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.code_editor_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.code_editor_app_settings.0.default_resource_spec.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.code_editor_app_settings.0.custom_image.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "default_user_settings.0.code_editor_app_settings.0.custom_image.0.app_image_config_name", "aws_sagemaker_app_image_config.test", "app_image_config_name"),
					resource.TestCheckResourceAttrPair(resourceName, "default_user_settings.0.code_editor_app_settings.0.custom_image.0.image_name", "aws_sagemaker_image.test", "image_name"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
		},
	})
}

func testAccDomain_codeEditorAppSettings_defaultResourceSpecAndCustomImage(t *testing.T) {
	ctx := acctest.Context(t)
	if os.Getenv("SAGEMAKER_IMAGE_VERSION_BASE_IMAGE") == "" {
		t.Skip("Environment variable SAGEMAKER_IMAGE_VERSION_BASE_IMAGE is not set")
	}

	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"
	baseImage := os.Getenv("SAGEMAKER_IMAGE_VERSION_BASE_IMAGE")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_codeEditorAppSettingsDefaultResourceSpecAndCustomImage(rName, baseImage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.code_editor_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.code_editor_app_settings.0.default_resource_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.code_editor_app_settings.0.custom_image.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "default_user_settings.0.code_editor_app_settings.0.default_resource_spec.0.sagemaker_image_version_arn", "aws_sagemaker_image_version.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "default_user_settings.0.code_editor_app_settings.0.custom_image.0.app_image_config_name", "aws_sagemaker_app_image_config.test", "app_image_config_name"),
					resource.TestCheckResourceAttrPair(resourceName, "default_user_settings.0.code_editor_app_settings.0.custom_image.0.image_name", "aws_sagemaker_image.test", "image_name"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
		},
	})
}

func testAccDomain_jupyterLabAppSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_jupyterLabAppSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.jupyter_lab_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.jupyter_lab_app_settings.0.default_resource_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.jupyter_lab_app_settings.0.default_resource_spec.0.instance_type", "ml.t3.micro"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
		},
	})
}

func testAccDomain_jupyterLabAppSettingsEMRSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_jupyterLabAppSettingsEMRSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.jupyter_lab_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.jupyter_lab_app_settings.0.default_resource_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.jupyter_lab_app_settings.0.default_resource_spec.0.instance_type", "ml.t3.micro"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.jupyter_lab_app_settings.0.emr_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.jupyter_lab_app_settings.0.emr_settings.0.assumable_role_arns.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "default_user_settings.0.jupyter_lab_app_settings.0.emr_settings.0.assumable_role_arns.*", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.jupyter_lab_app_settings.0.emr_settings.0.execution_role_arns.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "default_user_settings.0.jupyter_lab_app_settings.0.emr_settings.0.execution_role_arns.*", "aws_iam_role.test", names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
		},
	})
}

func testAccDomain_jupyterLabAppSettingsAppLifecycle(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_jupyterLabAppSettingsAppLifecycle(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.jupyter_lab_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.jupyter_lab_app_settings.0.app_lifecycle_management.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.jupyter_lab_app_settings.0.app_lifecycle_management.0.idle_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.jupyter_lab_app_settings.0.app_lifecycle_management.0.idle_settings.0.idle_timeout_in_minutes", "75"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.jupyter_lab_app_settings.0.app_lifecycle_management.0.idle_settings.0.min_idle_timeout_in_minutes", "60"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.jupyter_lab_app_settings.0.app_lifecycle_management.0.idle_settings.0.max_idle_timeout_in_minutes", "90"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.jupyter_lab_app_settings.0.app_lifecycle_management.0.idle_settings.0.lifecycle_management", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.jupyter_lab_app_settings.0.default_resource_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.jupyter_lab_app_settings.0.default_resource_spec.0.instance_type", "ml.t3.micro"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
		},
	})
}

func testAccDomain_jupyterLabAppSettingsBuiltInLifecycle(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfigjupyterLabAppSettingsBuiltInLifecycle(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.jupyter_lab_app_settings.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "default_user_settings.0.jupyter_lab_app_settings.0.built_in_lifecycle_config_arn", "aws_sagemaker_studio_lifecycle_config.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.jupyter_lab_app_settings.0.default_resource_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.jupyter_lab_app_settings.0.default_resource_spec.0.instance_type", "ml.t3.micro"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
		},
	})
}

func testAccDomain_kernelGatewayAppSettings_lifecycleConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_kernelGatewayAppSettingsLifecycle(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.kernel_gateway_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.kernel_gateway_app_settings.0.lifecycle_config_arns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.kernel_gateway_app_settings.0.default_resource_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.kernel_gateway_app_settings.0.default_resource_spec.0.instance_type", "ml.t3.micro"),
					resource.TestCheckResourceAttrPair(resourceName, "default_user_settings.0.kernel_gateway_app_settings.0.default_resource_spec.0.lifecycle_config_arn", "aws_sagemaker_studio_lifecycle_config.test", names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
		},
	})
}

func testAccDomain_kernelGatewayAppSettings_customImage(t *testing.T) {
	ctx := acctest.Context(t)
	if os.Getenv("SAGEMAKER_IMAGE_VERSION_BASE_IMAGE") == "" {
		t.Skip("Environment variable SAGEMAKER_IMAGE_VERSION_BASE_IMAGE is not set")
	}

	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"
	baseImage := os.Getenv("SAGEMAKER_IMAGE_VERSION_BASE_IMAGE")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_kernelGatewayAppSettingsCustomImage(rName, baseImage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.kernel_gateway_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.kernel_gateway_app_settings.0.default_resource_spec.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.kernel_gateway_app_settings.0.custom_image.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "default_user_settings.0.kernel_gateway_app_settings.0.custom_image.0.app_image_config_name", "aws_sagemaker_app_image_config.test", "app_image_config_name"),
					resource.TestCheckResourceAttrPair(resourceName, "default_user_settings.0.kernel_gateway_app_settings.0.custom_image.0.image_name", "aws_sagemaker_image.test", "image_name"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
		},
	})
}

func testAccDomain_kernelGatewayAppSettings_defaultResourceSpecAndCustomImage(t *testing.T) {
	ctx := acctest.Context(t)
	if os.Getenv("SAGEMAKER_IMAGE_VERSION_BASE_IMAGE") == "" {
		t.Skip("Environment variable SAGEMAKER_IMAGE_VERSION_BASE_IMAGE is not set")
	}

	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"
	baseImage := os.Getenv("SAGEMAKER_IMAGE_VERSION_BASE_IMAGE")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_kernelGatewayAppSettingsDefaultResourceSpecAndCustomImage(rName, baseImage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.kernel_gateway_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.kernel_gateway_app_settings.0.default_resource_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.kernel_gateway_app_settings.0.custom_image.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "default_user_settings.0.kernel_gateway_app_settings.0.default_resource_spec.0.sagemaker_image_version_arn", "aws_sagemaker_image_version.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "default_user_settings.0.kernel_gateway_app_settings.0.custom_image.0.app_image_config_name", "aws_sagemaker_app_image_config.test", "app_image_config_name"),
					resource.TestCheckResourceAttrPair(resourceName, "default_user_settings.0.kernel_gateway_app_settings.0.custom_image.0.image_name", "aws_sagemaker_image.test", "image_name"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
		},
	})
}

func testAccDomain_jupyterServerAppSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_jupyterServerAppSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.jupyter_server_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.jupyter_server_app_settings.0.default_resource_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.jupyter_server_app_settings.0.default_resource_spec.0.instance_type", "system"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
		},
	})
}

func testAccDomain_jupyterServerAppSettings_code(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_jupyterServerAppSettingsCode(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.jupyter_server_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.jupyter_server_app_settings.0.default_resource_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.jupyter_server_app_settings.0.default_resource_spec.0.instance_type", "system"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.jupyter_server_app_settings.0.code_repository.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "default_user_settings.0.jupyter_server_app_settings.0.code_repository.*", map[string]string{
						"repository_url": "https://github.com/hashicorp/terraform-provider-aws.git",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
		},
	})
}

func testAccDomain_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsagemaker.ResourceDomain(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccDomain_defaultUserSettingsUpdated(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, rName),
					resource.TestCheckResourceAttr(resourceName, "auth_mode", "IAM"),
					resource.TestCheckResourceAttr(resourceName, "app_network_access_type", "PublicInternetOnly"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "default_user_settings.0.execution_role", "aws_iam_role.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "sagemaker", regexache.MustCompile(`domain/.+`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, "aws_vpc.test", names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrURL),
					resource.TestCheckResourceAttrSet(resourceName, "home_efs_file_system_id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
			{
				Config: testAccDomainConfig_sharingSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "app_network_access_type", "VpcOnly"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.sharing_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.sharing_settings.0.notebook_output_option", "Allowed"),
					resource.TestCheckResourceAttrPair(resourceName, "default_user_settings.0.sharing_settings.0.s3_kms_key_id", "aws_kms_key.test", names.AttrKeyID),
					resource.TestCheckResourceAttrSet(resourceName, "default_user_settings.0.sharing_settings.0.s3_output_path"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
		},
	})
}

func testAccDomain_spaceSettingsKernelGatewayAppSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_defaultSpaceKernelGatewayAppSettings(rName, "ml.t3.micro"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.kernel_gateway_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.kernel_gateway_app_settings.0.default_resource_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.kernel_gateway_app_settings.0.default_resource_spec.0.instance_type", "ml.t3.micro"),
					resource.TestCheckResourceAttr(resourceName, "default_space_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_space_settings.0.kernel_gateway_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_space_settings.0.kernel_gateway_app_settings.0.default_resource_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_space_settings.0.kernel_gateway_app_settings.0.default_resource_spec.0.instance_type", "ml.t3.micro"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
			{
				Config: testAccDomainConfig_defaultSpaceKernelGatewayAppSettings(rName, "ml.t3.small"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.kernel_gateway_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.kernel_gateway_app_settings.0.default_resource_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.kernel_gateway_app_settings.0.default_resource_spec.0.instance_type", "ml.t3.small"),
					resource.TestCheckResourceAttr(resourceName, "default_space_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_space_settings.0.kernel_gateway_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_space_settings.0.kernel_gateway_app_settings.0.default_resource_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_space_settings.0.kernel_gateway_app_settings.0.default_resource_spec.0.instance_type", "ml.t3.small"),
				),
			},
		},
	})
}

func testAccDomain_posix(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_posix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, rName),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.custom_posix_user_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.custom_posix_user_config.0.gid", "1001"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.custom_posix_user_config.0.uid", "10000"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
		},
	})
}

func testAccDomain_spaceStorageSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_spaceStorageSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, rName),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.space_storage_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.space_storage_settings.0.default_ebs_storage_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.space_storage_settings.0.default_ebs_storage_settings.0.default_ebs_volume_size_in_gb", "10"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.space_storage_settings.0.default_ebs_storage_settings.0.maximum_ebs_volume_size_in_gb", "200"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
		},
	})
}

func testAccDomain_efs(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_efs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "default_user_settings.0.execution_role", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.custom_file_system_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.custom_file_system_config.0.efs_file_system_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "default_user_settings.0.custom_file_system_config.0.efs_file_system_config.0.file_system_id", "aws_efs_file_system.test", names.AttrID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
		},
	})
}

func testAccDomain_studioWebPortalSettings_hiddenAppTypes(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_studioWebPortalSettings_hiddenAppTypes(rName, []string{"JupyterServer", "KernelGateway"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.studio_web_portal_settings.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "default_user_settings.0.studio_web_portal_settings.0.hidden_app_types.*", "JupyterServer"),
					resource.TestCheckTypeSetElemAttr(resourceName, "default_user_settings.0.studio_web_portal_settings.0.hidden_app_types.*", "KernelGateway"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
			{
				Config: testAccDomainConfig_studioWebPortalSettings_hiddenAppTypes(rName, []string{"JupyterServer", "KernelGateway", "CodeEditor"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.studio_web_portal_settings.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "default_user_settings.0.studio_web_portal_settings.0.hidden_app_types.*", "JupyterServer"),
					resource.TestCheckTypeSetElemAttr(resourceName, "default_user_settings.0.studio_web_portal_settings.0.hidden_app_types.*", "KernelGateway"),
					resource.TestCheckTypeSetElemAttr(resourceName, "default_user_settings.0.studio_web_portal_settings.0.hidden_app_types.*", "CodeEditor"),
				),
			},
		},
	})
}

func testAccDomain_studioWebPortalSettings_hiddenInstanceTypes(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_studioWebPortalSettings_hiddenInstanceTypes(rName, []string{"ml.m5.8xlarge", "ml.m5.12xlarge"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.studio_web_portal_settings.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "default_user_settings.0.studio_web_portal_settings.0.hidden_instance_types.*", "ml.m5.8xlarge"),
					resource.TestCheckTypeSetElemAttr(resourceName, "default_user_settings.0.studio_web_portal_settings.0.hidden_instance_types.*", "ml.m5.12xlarge"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
			{
				Config: testAccDomainConfig_studioWebPortalSettings_hiddenInstanceTypes(rName, []string{"ml.m5.8xlarge", "ml.m5.12xlarge", "ml.m5.16xlarge"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.studio_web_portal_settings.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "default_user_settings.0.studio_web_portal_settings.0.hidden_instance_types.*", "ml.m5.8xlarge"),
					resource.TestCheckTypeSetElemAttr(resourceName, "default_user_settings.0.studio_web_portal_settings.0.hidden_instance_types.*", "ml.m5.12xlarge"),
					resource.TestCheckTypeSetElemAttr(resourceName, "default_user_settings.0.studio_web_portal_settings.0.hidden_instance_types.*", "ml.m5.16xlarge"),
				),
			},
		},
	})
}

func testAccDomain_studioWebPortalSettings_hiddenMlTools(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_studioWebPortalSettings_hiddenMlTools(rName, []string{"DataWrangler", "FeatureStore"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.studio_web_portal_settings.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "default_user_settings.0.studio_web_portal_settings.0.hidden_ml_tools.*", "DataWrangler"),
					resource.TestCheckTypeSetElemAttr(resourceName, "default_user_settings.0.studio_web_portal_settings.0.hidden_ml_tools.*", "FeatureStore"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
			{
				Config: testAccDomainConfig_studioWebPortalSettings_hiddenMlTools(rName, []string{"DataWrangler", "FeatureStore", "EmrClusters"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.studio_web_portal_settings.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "default_user_settings.0.studio_web_portal_settings.0.hidden_ml_tools.*", "DataWrangler"),
					resource.TestCheckTypeSetElemAttr(resourceName, "default_user_settings.0.studio_web_portal_settings.0.hidden_ml_tools.*", "FeatureStore"),
					resource.TestCheckTypeSetElemAttr(resourceName, "default_user_settings.0.studio_web_portal_settings.0.hidden_ml_tools.*", "EmrClusters"),
				),
			},
		},
	})
}

func testAccDomain_spaceSettingsJupyterLabAppSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_spaceSettingsJupyterLabAppSettings(rName, "ml.t3.micro", "ml.t3.small"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.jupyter_lab_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.jupyter_lab_app_settings.0.default_resource_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.jupyter_lab_app_settings.0.default_resource_spec.0.instance_type", "ml.t3.micro"),
					resource.TestCheckResourceAttr(resourceName, "default_space_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_space_settings.0.jupyter_lab_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_space_settings.0.jupyter_lab_app_settings.0.default_resource_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_space_settings.0.jupyter_lab_app_settings.0.default_resource_spec.0.instance_type", "ml.t3.small"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
			{
				Config: testAccDomainConfig_spaceSettingsJupyterLabAppSettings(rName, "ml.t3.small", "ml.t3.micro"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.jupyter_lab_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.jupyter_lab_app_settings.0.default_resource_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.jupyter_lab_app_settings.0.default_resource_spec.0.instance_type", "ml.t3.small"),
					resource.TestCheckResourceAttr(resourceName, "default_space_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_space_settings.0.jupyter_lab_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_space_settings.0.jupyter_lab_app_settings.0.default_resource_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_space_settings.0.jupyter_lab_app_settings.0.default_resource_spec.0.instance_type", "ml.t3.micro"),
				),
			},
		},
	})
}

func testAccDomain_spaceSettingsSpaceStorageSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_spaceSettingsSpaceStorageSettings(rName, "100", "200"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, rName),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.space_storage_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.space_storage_settings.0.default_ebs_storage_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.space_storage_settings.0.default_ebs_storage_settings.0.default_ebs_volume_size_in_gb", "10"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.space_storage_settings.0.default_ebs_storage_settings.0.maximum_ebs_volume_size_in_gb", "100"),
					resource.TestCheckResourceAttr(resourceName, "default_space_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_space_settings.0.space_storage_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_space_settings.0.space_storage_settings.0.default_ebs_storage_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_space_settings.0.space_storage_settings.0.default_ebs_storage_settings.0.default_ebs_volume_size_in_gb", "10"),
					resource.TestCheckResourceAttr(resourceName, "default_space_settings.0.space_storage_settings.0.default_ebs_storage_settings.0.maximum_ebs_volume_size_in_gb", "200"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
			{
				Config: testAccDomainConfig_spaceSettingsSpaceStorageSettings(rName, "150", "250"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, rName),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.space_storage_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.space_storage_settings.0.default_ebs_storage_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.space_storage_settings.0.default_ebs_storage_settings.0.default_ebs_volume_size_in_gb", "10"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.space_storage_settings.0.default_ebs_storage_settings.0.maximum_ebs_volume_size_in_gb", "150"),
					resource.TestCheckResourceAttr(resourceName, "default_space_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_space_settings.0.space_storage_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_space_settings.0.space_storage_settings.0.default_ebs_storage_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_space_settings.0.space_storage_settings.0.default_ebs_storage_settings.0.default_ebs_volume_size_in_gb", "10"),
					resource.TestCheckResourceAttr(resourceName, "default_space_settings.0.space_storage_settings.0.default_ebs_storage_settings.0.maximum_ebs_volume_size_in_gb", "250"),
				),
			},
		},
	})
}

func testAccDomain_spaceSettingsCustomPOSIXUserConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_spaceSettingsCustomPOSIXUserConfig(rName, "1001", "10000", "1002", "20000"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, rName),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.custom_posix_user_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.custom_posix_user_config.0.gid", "1001"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.custom_posix_user_config.0.uid", "10000"),
					resource.TestCheckResourceAttr(resourceName, "default_space_settings.0.custom_posix_user_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_space_settings.0.custom_posix_user_config.0.gid", "1002"),
					resource.TestCheckResourceAttr(resourceName, "default_space_settings.0.custom_posix_user_config.0.uid", "20000"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
			{
				Config: testAccDomainConfig_spaceSettingsCustomPOSIXUserConfig(rName, "2001", "20000", "2002", "40000"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, rName),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.custom_posix_user_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.custom_posix_user_config.0.gid", "2001"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.custom_posix_user_config.0.uid", "20000"),
					resource.TestCheckResourceAttr(resourceName, "default_space_settings.0.custom_posix_user_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_space_settings.0.custom_posix_user_config.0.gid", "2002"),
					resource.TestCheckResourceAttr(resourceName, "default_space_settings.0.custom_posix_user_config.0.uid", "40000"),
				),
			},
		},
	})
}

func testAccDomain_spaceSettingsCustomFileSystemConfigs(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_spaceSettingsCustomFileSystemConfigs(rName, "test-1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "default_user_settings.0.execution_role", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.custom_file_system_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.custom_file_system_config.0.efs_file_system_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "default_user_settings.0.custom_file_system_config.0.efs_file_system_config.0.file_system_id", "aws_efs_file_system.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "default_space_settings.0.execution_role", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_space_settings.0.custom_file_system_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_space_settings.0.custom_file_system_config.0.efs_file_system_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "default_space_settings.0.custom_file_system_config.0.efs_file_system_config.0.file_system_id", "aws_efs_file_system.test", names.AttrID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
			{
				Config: testAccDomainConfig_spaceSettingsCustomFileSystemConfigs(rName, "test-2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "default_user_settings.0.execution_role", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.custom_file_system_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.custom_file_system_config.0.efs_file_system_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "default_user_settings.0.custom_file_system_config.0.efs_file_system_config.0.file_system_id", "aws_efs_file_system.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "default_space_settings.0.execution_role", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_space_settings.0.custom_file_system_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_space_settings.0.custom_file_system_config.0.efs_file_system_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "default_space_settings.0.custom_file_system_config.0.efs_file_system_config.0.file_system_id", "aws_efs_file_system.test", names.AttrID),
				),
			},
		},
	})
}

func testAccCheckDomainDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sagemaker_domain" {
				continue
			}

			_, err := tfsagemaker.FindDomainByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return fmt.Errorf("reading SageMaker AI Domain (%s): %w", rs.Primary.ID, err)
			}

			return fmt.Errorf("sagemaker domain %q still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDomainExists(ctx context.Context, n string, codeRepo *sagemaker.DescribeDomainOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker domain ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerClient(ctx)
		resp, err := tfsagemaker.FindDomainByName(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*codeRepo = *resp

		return nil
	}
}

func testAccDomainConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/"
  assume_role_policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["sagemaker.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSageMakerFullAccess"
}
`, rName))
}

func testAccDomainConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName))
}

func testAccDomainConfig_posix(rName string) string {
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn
    custom_posix_user_config {
      gid = 1001
      uid = 10000
    }
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName))
}

func testAccDomainConfig_domainSettings(rName, config string) string {
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn
  }

  domain_settings {
    execution_role_identity_config = %[2]q
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName, config))
}

func testAccDomainConfig_domainSettingsDockerSettings(rName, config string) string {
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn
  }

  app_network_access_type = "VpcOnly"

  domain_settings {
    docker_settings {
      enable_docker_access      = %[2]q
      vpc_only_trusted_accounts = [data.aws_caller_identity.current.account_id]
    }
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName, config))
}

func testAccDomainConfig_kms(rName string) string {
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id
  kms_key_id  = aws_kms_key.test.key_id

  default_user_settings {
    execution_role = aws_iam_role.test.arn
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName))
}

func testAccDomainConfig_defaultUserSettingsSecurityGroupUpdated(rName string, sgCount int) string {
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_security_group" "test" {
  count = %[2]d

  name = "%[1]s-${count.index}"

  tags = {
    Name = "%[1]s-${count.index}"
  }
}

resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role  = aws_iam_role.test.arn
    security_groups = aws_security_group.test[*].id
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName, sgCount))
}

func testAccDomainConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccDomainConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccDomainConfig_sharingSettings(rName string) string {
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}


resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_sagemaker_domain" "test" {
  domain_name             = %[1]q
  auth_mode               = "IAM"
  vpc_id                  = aws_vpc.test.id
  subnet_ids              = aws_subnet.test[*].id
  app_network_access_type = "VpcOnly"

  default_user_settings {
    execution_role = aws_iam_role.test.arn

    sharing_settings {
      notebook_output_option = "Allowed"
      s3_kms_key_id          = aws_kms_key.test.key_id
      s3_output_path         = "s3://${aws_s3_bucket.test.bucket}/sharing"
    }
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName))
}

func testAccDomainConfig_canvasAppSettings(rName string) string {
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn

    canvas_app_settings {
      time_series_forecasting_settings {
        status = "DISABLED"
      }
    }
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName))
}

func testAccDomainConfig_modelRegisterSettings(rName string) string {
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn

    canvas_app_settings {
      model_register_settings {
        status = "DISABLED"
      }
    }
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName))
}

func testAccDomainConfig_generativeAiSettings(rName string) string {
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn

    canvas_app_settings {
      generative_ai_settings {
        amazon_bedrock_role_arn = aws_iam_role.test.arn
      }
    }
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName))
}

func testAccDomainConfig_kendraSettings(rName string) string {
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn

    canvas_app_settings {
      kendra_settings {
        status = "DISABLED"
      }
    }
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName))
}

func testAccDomainConfig_directDeploySettings(rName string) string {
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn

    canvas_app_settings {
      direct_deploy_settings {
        status = "DISABLED"
      }
    }
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName))
}

func testAccDomainConfig_emrServerlessSettings(rName string) string {
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn

    canvas_app_settings {
      emr_serverless_settings {
        execution_role_arn = aws_iam_role.test.arn
        status             = "ENABLED"
      }
    }
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName))
}

func testAccDomainConfig_identityProviderOAuthSettings(rName string) string {
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = jsonencode({ username = "example", password = "example" })
}

resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn

    canvas_app_settings {
      identity_provider_oauth_settings {
        secret_arn = aws_secretsmanager_secret.test.arn
        status     = "DISABLED"
      }
    }
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }

  depends_on = [aws_secretsmanager_secret_version.test]
}
`, rName))
}

func testAccDomainConfig_workspaceSettings(rName string) string {
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn

    canvas_app_settings {
      workspace_settings {
        s3_artifact_path = "s3://${aws_s3_bucket.test.bucket}/path"
      }
    }
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName))
}

func testAccDomainConfig_tensorBoardAppSettings(rName string) string {
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn

    tensor_board_app_settings {
      default_resource_spec {
        instance_type = "ml.t3.micro"
      }
    }
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName))
}

func testAccDomainConfig_tensorBoardAppSettingsImage(rName string) string {
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_image" "test" {
  image_name = %[1]q
  role_arn   = aws_iam_role.test.arn
}

resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn

    tensor_board_app_settings {
      default_resource_spec {
        instance_type       = "ml.t3.micro"
        sagemaker_image_arn = aws_sagemaker_image.test.arn
      }
    }
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName))
}

func testAccDomainConfig_jupyterServerAppSettings(rName string) string {
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn

    jupyter_server_app_settings {
      default_resource_spec {
        instance_type = "system"
      }
    }
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName))
}

func testAccDomainConfig_jupyterServerAppSettingsCode(rName string) string {
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn

    jupyter_server_app_settings {
      code_repository {
        repository_url = "https://github.com/hashicorp/terraform-provider-aws.git"
      }

      default_resource_spec {
        instance_type = "system"
      }
    }
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName))
}

func testAccDomainConfig_rSessionAppSettings(rName string) string {
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn

    r_session_app_settings {
      default_resource_spec {
        instance_type = "ml.t3.micro"
      }
    }
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName))
}

func testAccDomainConfig_rStudioServerProAppSettings(rName, state string) string {
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn

    r_studio_server_pro_app_settings {
      access_status = %[2]q
      user_group    = "R_STUDIO_ADMIN"
    }
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName, state))
}

func testAccDomainConfig_rStudioServerProDomainSettings(rName, connectURL string, packageURL string) string {
	return acctest.ConfigCompose(testAccDomainConfig_baseWithLicense(rName), fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  domain_settings {

    r_studio_server_pro_domain_settings {
      r_studio_connect_url         = %[2]q
      r_studio_package_manager_url = %[3]q
      domain_execution_role_arn    = aws_iam_role.test.arn
      default_resource_spec {
        instance_type = "system"
      }
    }
  }

  default_user_settings {
    execution_role = aws_iam_role.test.arn
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }

  # ignoring default image
  # it would be too hard to create the logic to find the default Rstudio image: https://docs.aws.amazon.com/sagemaker/latest/dg/rstudio-version.html
  # it changes for every region
  lifecycle {
    ignore_changes = [
      domain_settings[0].r_studio_server_pro_domain_settings[0].default_resource_spec[0]
    ]
  }
}
`, rName, connectURL, packageURL))
}

func testAccDomainConfig_rStudioDomainDisabledNetworkUpdate(rName, networkAccess string) string {
	return acctest.ConfigCompose(testAccDomainConfig_baseWithLicense(rName), fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name             = %[1]q
  auth_mode               = "IAM"
  vpc_id                  = aws_vpc.test.id
  subnet_ids              = aws_subnet.test[*].id
  app_network_access_type = %[2]q

  default_user_settings {
    execution_role = aws_iam_role.test.arn
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName, networkAccess))
}

func testAccDomainConfig_baseWithLicense(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/"
  assume_role_policy = data.aws_iam_policy_document.test.json
  inline_policy {
    name   = "GetLicense"
    policy = data.aws_iam_policy_document.license.json
  }
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["sagemaker.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

# needed for RStudio
data "aws_iam_policy_document" "license" {
  statement {
    sid    = "ReadLicense"
    effect = "Allow"
    actions = [
      "license-manager:ExtendLicenseConsumption",
      "license-manager:ListReceivedLicenses",
      "license-manager:GetLicense",
      "license-manager:CheckoutLicense",
      "license-manager:CheckInLicense",
    ]
    resources = ["*"]
  }
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSageMakerFullAccess"
}
`, rName))
}

func testAccDomainConfig_codeEditorAppSettings(rName string) string {
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn

    code_editor_app_settings {
      default_resource_spec {
        instance_type = "ml.t3.micro"
      }
    }
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName))
}

func testAccDomainConfig_codeEditorAppSettingsCustomImage(rName, baseImage string) string {
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_image" "test" {
  image_name = %[1]q
  role_arn   = aws_iam_role.test.arn

  depends_on = [aws_iam_role_policy_attachment.test]
}

resource "aws_sagemaker_app_image_config" "test" {
  app_image_config_name = %[1]q
}

resource "aws_sagemaker_image_version" "test" {
  image_name = aws_sagemaker_image.test.id
  base_image = %[2]q
}

resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn

    code_editor_app_settings {
      custom_image {
        app_image_config_name = aws_sagemaker_app_image_config.test.app_image_config_name
        image_name            = aws_sagemaker_image_version.test.image_name
      }
    }
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName, baseImage))
}

func testAccDomainConfig_codeEditorAppSettingsDefaultResourceSpecAndCustomImage(rName, baseImage string) string {
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_image" "test" {
  image_name = %[1]q
  role_arn   = aws_iam_role.test.arn

  depends_on = [aws_iam_role_policy_attachment.test]
}

resource "aws_sagemaker_app_image_config" "test" {
  app_image_config_name = %[1]q
}

resource "aws_sagemaker_image_version" "test" {
  image_name = aws_sagemaker_image.test.id
  base_image = %[2]q
}

resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn

    code_editor_app_settings {
      default_resource_spec {
        instance_type               = "ml.t3.micro"
        sagemaker_image_version_arn = aws_sagemaker_image_version.test.arn
      }

      custom_image {
        app_image_config_name = aws_sagemaker_app_image_config.test.app_image_config_name
        image_name            = aws_sagemaker_image_version.test.image_name
      }
    }
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName, baseImage))
}

func testAccDomainConfig_kernelGatewayAppSettings(rName string) string {
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn

    kernel_gateway_app_settings {
      default_resource_spec {
        instance_type = "ml.t3.micro"
      }
    }
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName))
}

func testAccDomainConfig_jupyterLabAppSettings(rName string) string {
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn

    jupyter_lab_app_settings {
      default_resource_spec {
        instance_type = "ml.t3.micro"
      }
    }
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName))
}

func testAccDomainConfig_jupyterLabAppSettingsEMRSettings(rName string) string {
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn

    jupyter_lab_app_settings {
      default_resource_spec {
        instance_type = "ml.t3.micro"
      }
      emr_settings {
        assumable_role_arns = [aws_iam_role.test.arn]
        execution_role_arns = [aws_iam_role.test.arn]
      }
    }
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName))
}

func testAccDomainConfig_jupyterLabAppSettingsAppLifecycle(rName string) string {
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn

    jupyter_lab_app_settings {
      app_lifecycle_management {
        idle_settings {
          idle_timeout_in_minutes     = 75
          lifecycle_management        = "ENABLED"
          max_idle_timeout_in_minutes = 90
          min_idle_timeout_in_minutes = 60
        }
      }
      default_resource_spec {
        instance_type = "ml.t3.micro"
      }
    }
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName))
}

func testAccDomainConfigjupyterLabAppSettingsBuiltInLifecycle(rName string) string {
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_studio_lifecycle_config" "test" {
  studio_lifecycle_config_name     = %[1]q
  studio_lifecycle_config_app_type = "JupyterServer"
  studio_lifecycle_config_content  = base64encode("echo Hello")
}

resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn

    jupyter_lab_app_settings {
      default_resource_spec {
        instance_type = "ml.t3.micro"
      }

      built_in_lifecycle_config_arn = aws_sagemaker_studio_lifecycle_config.test.arn
      lifecycle_config_arns         = [aws_sagemaker_studio_lifecycle_config.test.arn]
    }
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName))
}

func testAccDomainConfig_defaultSpaceKernelGatewayAppSettings(rName, instance string) string {
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn

    kernel_gateway_app_settings {
      default_resource_spec {
        instance_type = %[2]q
      }
    }
  }

  default_space_settings {
    execution_role = aws_iam_role.test.arn

    kernel_gateway_app_settings {
      default_resource_spec {
        instance_type = %[2]q
      }
    }
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName, instance))
}

func testAccDomainConfig_kernelGatewayAppSettingsLifecycle(rName string) string {
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_studio_lifecycle_config" "test" {
  studio_lifecycle_config_name     = %[1]q
  studio_lifecycle_config_app_type = "JupyterServer"
  studio_lifecycle_config_content  = base64encode("echo Hello")
}

resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn

    kernel_gateway_app_settings {
      default_resource_spec {
        instance_type        = "ml.t3.micro"
        lifecycle_config_arn = aws_sagemaker_studio_lifecycle_config.test.arn
      }

      lifecycle_config_arns = [aws_sagemaker_studio_lifecycle_config.test.arn]
    }
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName))
}

func testAccDomainConfig_kernelGatewayAppSettingsCustomImage(rName, baseImage string) string {
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_image" "test" {
  image_name = %[1]q
  role_arn   = aws_iam_role.test.arn

  depends_on = [aws_iam_role_policy_attachment.test]
}

resource "aws_sagemaker_app_image_config" "test" {
  app_image_config_name = %[1]q

  kernel_gateway_image_config {
    kernel_spec {
      name = %[1]q
    }
  }
}

resource "aws_sagemaker_image_version" "test" {
  image_name = aws_sagemaker_image.test.id
  base_image = %[2]q
}

resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn

    kernel_gateway_app_settings {
      custom_image {
        app_image_config_name = aws_sagemaker_app_image_config.test.app_image_config_name
        image_name            = aws_sagemaker_image_version.test.image_name
      }
    }
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName, baseImage))
}

func testAccDomainConfig_kernelGatewayAppSettingsDefaultResourceSpecAndCustomImage(rName, baseImage string) string {
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_image" "test" {
  image_name = %[1]q
  role_arn   = aws_iam_role.test.arn

  depends_on = [aws_iam_role_policy_attachment.test]
}

resource "aws_sagemaker_app_image_config" "test" {
  app_image_config_name = %[1]q

  kernel_gateway_image_config {
    kernel_spec {
      name = %[1]q
    }
  }
}

resource "aws_sagemaker_image_version" "test" {
  image_name = aws_sagemaker_image.test.id
  base_image = %[2]q
}

resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn

    kernel_gateway_app_settings {
      default_resource_spec {
        instance_type               = "ml.t3.micro"
        sagemaker_image_version_arn = aws_sagemaker_image_version.test.arn
      }

      custom_image {
        app_image_config_name = aws_sagemaker_app_image_config.test.app_image_config_name
        image_name            = aws_sagemaker_image_version.test.image_name
      }
    }
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName, baseImage))
}

func testAccDomainConfig_efs(rName string) string {
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_efs_mount_target" "test" {
  file_system_id = aws_efs_file_system.test.id
  subnet_id      = aws_subnet.test[0].id
}

resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn

    custom_file_system_config {
      efs_file_system_config {
        file_system_id   = aws_efs_mount_target.test.file_system_id
        file_system_path = "/"
      }
    }
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName))
}

func testAccDomainConfig_spaceStorageSettings(rName string) string {
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn
    space_storage_settings {
      default_ebs_storage_settings {
        default_ebs_volume_size_in_gb = 10
        maximum_ebs_volume_size_in_gb = 200
      }
    }
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName))
}

func testAccDomainConfig_studioWebPortalSettings_hiddenAppTypes(rName string, hiddenAppTypes []string) string {
	var hiddenAppTypesString string
	for i, appType := range hiddenAppTypes {
		if i > 0 {
			hiddenAppTypesString += ", "
		}
		hiddenAppTypesString += fmt.Sprintf("%q", appType)
	}
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn

    studio_web_portal_settings {
      hidden_app_types = [%[2]s]
    }
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName, hiddenAppTypesString))
}

func testAccDomainConfig_studioWebPortalSettings_hiddenInstanceTypes(rName string, hiddenInstanceTypes []string) string {
	var hiddenInstanceTypesString string
	for i, instanceType := range hiddenInstanceTypes {
		if i > 0 {
			hiddenInstanceTypesString += ", "
		}
		hiddenInstanceTypesString += fmt.Sprintf("%q", instanceType)
	}
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn

    studio_web_portal_settings {
      hidden_instance_types = [%[2]s]
    }
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName, hiddenInstanceTypesString))
}

func testAccDomainConfig_studioWebPortalSettings_hiddenMlTools(rName string, hiddenMlTools []string) string {
	var hiddenMlToolsString string
	for i, mlTool := range hiddenMlTools {
		if i > 0 {
			hiddenMlToolsString += ", "
		}
		hiddenMlToolsString += fmt.Sprintf("%q", mlTool)
	}
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn

    studio_web_portal_settings {
      hidden_ml_tools = [%[2]s]
    }
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName, hiddenMlToolsString))
}

func testAccDomainConfig_spaceSettingsJupyterLabAppSettings(rName string, defaultUserSettingsinstanceType string, defaultSpaceSettingsinstanceType string) string {
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn

    jupyter_lab_app_settings {
      default_resource_spec {
        instance_type = %[2]q
      }
    }
  }

  default_space_settings {
    execution_role = aws_iam_role.test.arn

    jupyter_lab_app_settings {
      default_resource_spec {
        instance_type = %[3]q
      }
    }
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName, defaultUserSettingsinstanceType, defaultSpaceSettingsinstanceType))
}

func testAccDomainConfig_spaceSettingsSpaceStorageSettings(rName string, defaultUserSettingsMaxEbsVolumeSize string, defaultSpaceSettingsMaxEbsVolumeSize string) string {
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn
    space_storage_settings {
      default_ebs_storage_settings {
        default_ebs_volume_size_in_gb = 10
        maximum_ebs_volume_size_in_gb = %[2]q
      }
    }
  }

  default_space_settings {
    execution_role = aws_iam_role.test.arn
    space_storage_settings {
      default_ebs_storage_settings {
        default_ebs_volume_size_in_gb = 10
        maximum_ebs_volume_size_in_gb = %[3]q
      }
    }
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName, defaultUserSettingsMaxEbsVolumeSize, defaultSpaceSettingsMaxEbsVolumeSize))
}

func testAccDomainConfig_spaceSettingsCustomPOSIXUserConfig(rName string, defaultUserSettingsGid string, defaultUserSettingsUid string, defaultSpaceSettingsGid string, defaultSpaceSettingsUid string) string {
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn
    custom_posix_user_config {
      gid = %[2]q
      uid = %[3]q
    }
  }

  default_space_settings {
    execution_role = aws_iam_role.test.arn
    custom_posix_user_config {
      gid = %[4]q
      uid = %[5]q
    }
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName, defaultUserSettingsGid, defaultUserSettingsUid, defaultSpaceSettingsGid, defaultSpaceSettingsUid))
}

func testAccDomainConfig_spaceSettingsCustomFileSystemConfigs(rName string, efsName string) string {
	return acctest.ConfigCompose(testAccDomainConfig_base(rName), fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %[2]q
  tags = {
    Name = %[2]q
  }
}

resource "aws_efs_mount_target" "test" {
  file_system_id = aws_efs_file_system.test.id
  subnet_id      = aws_subnet.test[0].id
}

resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn

    custom_file_system_config {
      efs_file_system_config {
        file_system_id   = aws_efs_mount_target.test.file_system_id
        file_system_path = "/"
      }
    }
  }

  default_space_settings {
    execution_role = aws_iam_role.test.arn

    custom_file_system_config {
      efs_file_system_config {
        file_system_id   = aws_efs_mount_target.test.file_system_id
        file_system_path = "/"
      }
    }
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName, efsName))
}
