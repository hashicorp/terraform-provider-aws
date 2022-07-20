package sagemaker_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
)

func testAccDomain_basic(t *testing.T) {
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "domain_name", rName),
					resource.TestCheckResourceAttr(resourceName, "auth_mode", "IAM"),
					resource.TestCheckResourceAttr(resourceName, "app_network_access_type", "PublicInternetOnly"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "default_user_settings.0.execution_role", "aws_iam_role.test", "arn"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "sagemaker", regexp.MustCompile(`domain/.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", "aws_vpc.test", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "url"),
					resource.TestCheckResourceAttrSet(resourceName, "home_efs_file_system_id"),
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

func testAccDomain_kms(t *testing.T) {
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_kms(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", "aws_kms_key.test", "arn"),
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
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_policy"},
			},
			{
				Config: testAccDomainConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccDomainConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccDomain_securityGroup(t *testing.T) {
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_securityGroup1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
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
				Config: testAccDomainConfig_securityGroup2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.security_groups.#", "2"),
				),
			},
		},
	})
}

func testAccDomain_sharingSettings(t *testing.T) {
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_sharingSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.sharing_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.sharing_settings.0.notebook_output_option", "Allowed"),
					resource.TestCheckResourceAttrPair(resourceName, "default_user_settings.0.sharing_settings.0.s3_kms_key_id", "aws_kms_key.test", "arn"),
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

func testAccDomain_tensorboardAppSettings(t *testing.T) {
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_tensorBoardAppSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
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
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_tensorBoardAppSettingsImage(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.tensor_board_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.tensor_board_app_settings.0.default_resource_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.tensor_board_app_settings.0.default_resource_spec.0.instance_type", "ml.t3.micro"),
					resource.TestCheckResourceAttrPair(resourceName, "default_user_settings.0.tensor_board_app_settings.0.default_resource_spec.0.sagemaker_image_arn", "aws_sagemaker_image.test", "arn"),
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

func testAccDomain_kernelGatewayAppSettings(t *testing.T) {
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_kernelGatewayAppSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
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

func testAccDomain_kernelGatewayAppSettings_lifecycleConfig(t *testing.T) {
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_kernelGatewayAppSettingsLifecycle(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.kernel_gateway_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.kernel_gateway_app_settings.0.lifecycle_config_arns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.kernel_gateway_app_settings.0.default_resource_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.kernel_gateway_app_settings.0.default_resource_spec.0.instance_type", "ml.t3.micro"),
					resource.TestCheckResourceAttrPair(resourceName, "default_user_settings.0.kernel_gateway_app_settings.0.default_resource_spec.0.lifecycle_config_arn", "aws_sagemaker_studio_lifecycle_config.test", "arn"),
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

	if os.Getenv("SAGEMAKER_IMAGE_VERSION_BASE_IMAGE") == "" {
		t.Skip("Environment variable SAGEMAKER_IMAGE_VERSION_BASE_IMAGE is not set")
	}

	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"
	baseImage := os.Getenv("SAGEMAKER_IMAGE_VERSION_BASE_IMAGE")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_kernelGatewayAppSettingsCustomImage(rName, baseImage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
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

	if os.Getenv("SAGEMAKER_IMAGE_VERSION_BASE_IMAGE") == "" {
		t.Skip("Environment variable SAGEMAKER_IMAGE_VERSION_BASE_IMAGE is not set")
	}

	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"
	baseImage := os.Getenv("SAGEMAKER_IMAGE_VERSION_BASE_IMAGE")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_kernelGatewayAppSettingsDefaultResourceSpecAndCustomImage(rName, baseImage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.kernel_gateway_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.kernel_gateway_app_settings.0.default_resource_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.kernel_gateway_app_settings.0.custom_image.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "default_user_settings.0.kernel_gateway_app_settings.0.default_resource_spec.0.sagemaker_image_version_arn", "aws_sagemaker_image_version.test", "arn"),
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
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_jupyterServerAppSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
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

func testAccDomain_disappears(t *testing.T) {
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					acctest.CheckResourceDisappears(acctest.Provider, tfsagemaker.ResourceDomain(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccDomain_defaultUserSettingsUpdated(t *testing.T) {
	var domain sagemaker.DescribeDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "domain_name", rName),
					resource.TestCheckResourceAttr(resourceName, "auth_mode", "IAM"),
					resource.TestCheckResourceAttr(resourceName, "app_network_access_type", "PublicInternetOnly"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "default_user_settings.0.execution_role", "aws_iam_role.test", "arn"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "sagemaker", regexp.MustCompile(`domain/.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", "aws_vpc.test", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "url"),
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
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.sharing_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.sharing_settings.0.notebook_output_option", "Allowed"),
					resource.TestCheckResourceAttrPair(resourceName, "default_user_settings.0.sharing_settings.0.s3_kms_key_id", "aws_kms_key.test", "arn"),
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

func testAccCheckDomainDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_domain" {
			continue
		}

		domain, err := tfsagemaker.FindDomainByName(conn, rs.Primary.ID)

		if tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading SageMaker Domain (%s): %w", rs.Primary.ID, err)
		}

		domainArn := aws.StringValue(domain.DomainArn)
		domainID, err := tfsagemaker.DecodeDomainID(domainArn)
		if err != nil {
			return err
		}

		if domainID == rs.Primary.ID {
			return fmt.Errorf("sagemaker domain %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckDomainExists(n string, codeRepo *sagemaker.DescribeDomainOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker domain ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn
		resp, err := tfsagemaker.FindDomainByName(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*codeRepo = *resp

		return nil
	}
}

func testAccDomainBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.0.1.0/24"

  tags = {
    Name = %[1]q
  }
}

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
`, rName)
}

func testAccDomainConfig_basic(rName string) string {
	return testAccDomainBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = [aws_subnet.test.id]

  default_user_settings {
    execution_role = aws_iam_role.test.arn
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName)
}

func testAccDomainConfig_kms(rName string) string {
	return testAccDomainBaseConfig(rName) + fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "Terraform acc test"
  deletion_window_in_days = 7
}

resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = [aws_subnet.test.id]
  kms_key_id  = aws_kms_key.test.arn

  default_user_settings {
    execution_role = aws_iam_role.test.arn
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName)
}

func testAccDomainConfig_securityGroup1(rName string) string {
	return testAccDomainBaseConfig(rName) + fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = "%[1]s"
}

resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = [aws_subnet.test.id]

  default_user_settings {
    execution_role  = aws_iam_role.test.arn
    security_groups = [aws_security_group.test.id]
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName)
}

func testAccDomainConfig_securityGroup2(rName string) string {
	return testAccDomainBaseConfig(rName) + fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_security_group" "test2" {
  name = "%[1]s-2"
}

resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = [aws_subnet.test.id]

  default_user_settings {
    execution_role  = aws_iam_role.test.arn
    security_groups = [aws_security_group.test.id, aws_security_group.test2.id]
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName)
}

func testAccDomainConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return testAccDomainBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = [aws_subnet.test.id]

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
`, rName, tagKey1, tagValue1)
}

func testAccDomainConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccDomainBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = [aws_subnet.test.id]

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
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccDomainConfig_sharingSettings(rName string) string {
	return testAccDomainBaseConfig(rName) + fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}


resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = [aws_subnet.test.id]

  default_user_settings {
    execution_role = aws_iam_role.test.arn

    sharing_settings {
      notebook_output_option = "Allowed"
      s3_kms_key_id          = aws_kms_key.test.arn
      s3_output_path         = "s3://${aws_s3_bucket.test.bucket}/sharing"
    }
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName)
}

func testAccDomainConfig_tensorBoardAppSettings(rName string) string {
	return testAccDomainBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = [aws_subnet.test.id]

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
`, rName)
}

func testAccDomainConfig_tensorBoardAppSettingsImage(rName string) string {
	return testAccDomainBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_image" "test" {
  image_name = %[1]q
  role_arn   = aws_iam_role.test.arn
}

resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = [aws_subnet.test.id]

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
`, rName)
}

func testAccDomainConfig_jupyterServerAppSettings(rName string) string {
	return testAccDomainBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = [aws_subnet.test.id]

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
`, rName)
}

func testAccDomainConfig_kernelGatewayAppSettings(rName string) string {
	return testAccDomainBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = [aws_subnet.test.id]

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
`, rName)
}

func testAccDomainConfig_kernelGatewayAppSettingsLifecycle(rName string) string {
	return testAccDomainBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_studio_lifecycle_config" "test" {
  studio_lifecycle_config_name     = %[1]q
  studio_lifecycle_config_app_type = "JupyterServer"
  studio_lifecycle_config_content  = base64encode("echo Hello")
}

resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = [aws_subnet.test.id]

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
`, rName)
}

func testAccDomainConfig_kernelGatewayAppSettingsCustomImage(rName, baseImage string) string {
	return testAccDomainBaseConfig(rName) + fmt.Sprintf(`
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
  subnet_ids  = [aws_subnet.test.id]

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
`, rName, baseImage)
}

func testAccDomainConfig_kernelGatewayAppSettingsDefaultResourceSpecAndCustomImage(rName, baseImage string) string {
	return testAccDomainBaseConfig(rName) + fmt.Sprintf(`
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
  subnet_ids  = [aws_subnet.test.id]

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
`, rName, baseImage)
}
