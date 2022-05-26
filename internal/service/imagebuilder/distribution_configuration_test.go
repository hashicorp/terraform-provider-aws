package imagebuilder_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfimagebuilder "github.com/hashicorp/terraform-provider-aws/internal/service/imagebuilder"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(imagebuilder.EndpointsID, testAccErrorCheckSkip)
}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"You have reached the maximum allowed number of license configurations created in one day",
	)
}

func TestAccImageBuilderDistributionConfiguration_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfigurationConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "imagebuilder", fmt.Sprintf("distribution-configuration/%s", rName)),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_created"),
					resource.TestCheckResourceAttr(resourceName, "date_updated", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func TestAccImageBuilderDistributionConfiguration_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfigurationConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfimagebuilder.ResourceDistributionConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccImageBuilderDistributionConfiguration_description(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfigurationConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDistributionConfigurationConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func TestAccImageBuilderDistributionConfiguration_distribution(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.FactoriesMultipleRegion(nil, 2),
		CheckDestroy:      testAccCheckDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfigurationConfig_2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "2"),
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

func TestAccImageBuilderDistributionConfiguration_DistributionAMIDistribution_amiTags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfigurationConfig_amiTags(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "distribution.*", map[string]string{
						"ami_distribution_configuration.#":               "1",
						"ami_distribution_configuration.0.ami_tags.%":    "1",
						"ami_distribution_configuration.0.ami_tags.key1": "value1",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDistributionConfigurationConfig_amiTags(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "distribution.*", map[string]string{
						"ami_distribution_configuration.#":               "1",
						"ami_distribution_configuration.0.ami_tags.%":    "1",
						"ami_distribution_configuration.0.ami_tags.key2": "value2",
					}),
				),
			},
		},
	})
}

func TestAccImageBuilderDistributionConfiguration_DistributionAMIDistribution_description(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfigurationConfig_amiDescription(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "distribution.*", map[string]string{
						"ami_distribution_configuration.#":             "1",
						"ami_distribution_configuration.0.description": "description1",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDistributionConfigurationConfig_amiDescription(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "distribution.*", map[string]string{
						"ami_distribution_configuration.#":             "1",
						"ami_distribution_configuration.0.description": "description2",
					}),
				),
			},
		},
	})
}

func TestAccImageBuilderDistributionConfiguration_DistributionAMIDistribution_kmsKeyID(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	kmsKeyResourceName := "aws_kms_key.test"
	kmsKeyResourceName2 := "aws_kms_key.test2"
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfigurationConfig_amiKMSKeyID1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "distribution.*.ami_distribution_configuration.0.kms_key_id", kmsKeyResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDistributionConfigurationConfig_amiKMSKeyID2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "distribution.*.ami_distribution_configuration.0.kms_key_id", kmsKeyResourceName2, "arn"),
				),
			},
		},
	})
}

func TestAccImageBuilderDistributionConfiguration_DistributionAMIDistributionLaunchPermission_userGroups(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfigurationConfig_amiLaunchPermissionUserGroups(rName, "all"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "distribution.*.ami_distribution_configuration.0.launch_permission.0.user_groups.*", "all"),
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

func TestAccImageBuilderDistributionConfiguration_DistributionAMIDistributionLaunchPermission_userIDs(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfigurationConfig_amiLaunchPermissionUserIDs(rName, "111111111111"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "distribution.*.ami_distribution_configuration.0.launch_permission.0.user_ids.*", "111111111111"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDistributionConfigurationConfig_amiLaunchPermissionUserIDs(rName, "222222222222"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "distribution.*.ami_distribution_configuration.0.launch_permission.0.user_ids.*", "222222222222"),
				),
			},
		},
	})
}

func TestAccImageBuilderDistributionConfiguration_DistributionAMIDistributionLaunchPermission_organizationARNs(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	organizationResourceName := "aws_organizations_organization.test"
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckOrganizationsAccount(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfigurationConfig_amiLaunchPermissionOrganizationARNs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "distribution.*.ami_distribution_configuration.0.launch_permission.0.organization_arns.*", organizationResourceName, "arn"),
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

func TestAccImageBuilderDistributionConfiguration_DistributionAMIDistributionLaunchPermission_ouARNs(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	organizationalUnitResourceName := "aws_organizations_organizational_unit.test"

	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckOrganizationsAccount(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfigurationConfig_amiLaunchPermissionOrganizationalUnitARNs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "distribution.*.ami_distribution_configuration.0.launch_permission.0.organizational_unit_arns.*", organizationalUnitResourceName, "arn"),
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

func TestAccImageBuilderDistributionConfiguration_DistributionAMIDistribution_name(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfigurationConfig_amiName(rName, "name1-{{ imagebuilder:buildDate }}"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "distribution.*", map[string]string{
						"ami_distribution_configuration.#":      "1",
						"ami_distribution_configuration.0.name": "name1-{{ imagebuilder:buildDate }}",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDistributionConfigurationConfig_amiName(rName, "name2-{{ imagebuilder:buildDate }}"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "distribution.*", map[string]string{
						"ami_distribution_configuration.#":      "1",
						"ami_distribution_configuration.0.name": "name2-{{ imagebuilder:buildDate }}",
					}),
				),
			},
		},
	})
}

func TestAccImageBuilderDistributionConfiguration_DistributionAMIDistribution_targetAccountIDs(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfigurationConfig_amiTargetAccountIDs(rName, "111111111111"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "distribution.*.ami_distribution_configuration.0.target_account_ids.*", "111111111111"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDistributionConfigurationConfig_amiTargetAccountIDs(rName, "222222222222"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "distribution.*.ami_distribution_configuration.0.target_account_ids.*", "222222222222"),
				),
			},
		},
	})
}

func TestAccImageBuilderDistributionConfiguration_DistributionContainerDistribution_containerTags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfigurationConfig_containerTags(rName, "tag1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "distribution.0.container_distribution_configuration.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "distribution.0.container_distribution_configuration.0.container_tags.*", "tag1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDistributionConfigurationConfig_containerTags(rName, "tag2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "distribution.0.container_distribution_configuration.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "distribution.0.container_distribution_configuration.0.container_tags.*", "tag2"),
				),
			},
		},
	})
}

func TestAccImageBuilderDistributionConfiguration_DistributionContainerDistribution_description(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfigurationConfig_containerDescription(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "distribution.0.container_distribution_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "distribution.0.container_distribution_configuration.0.description", "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDistributionConfigurationConfig_containerDescription(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "distribution.0.container_distribution_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "distribution.0.container_distribution_configuration.0.description", "description2"),
				),
			},
		},
	})
}

func TestAccImageBuilderDistributionConfiguration_DistributionContainerDistribution_targetRepository(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfigurationConfig_containerTargetRepository(rName, "repository1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "distribution.0.container_distribution_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "distribution.0.container_distribution_configuration.0.target_repository.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "distribution.0.container_distribution_configuration.0.target_repository.0.repository_name", "repository1"),
					resource.TestCheckResourceAttr(resourceName, "distribution.0.container_distribution_configuration.0.target_repository.0.service", "ECR"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDistributionConfigurationConfig_containerTargetRepository(rName, "repository2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "distribution.0.container_distribution_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "distribution.0.container_distribution_configuration.0.target_repository.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "distribution.0.container_distribution_configuration.0.target_repository.0.repository_name", "repository2"),
					resource.TestCheckResourceAttr(resourceName, "distribution.0.container_distribution_configuration.0.target_repository.0.service", "ECR"),
				),
			},
		},
	})
}

func TestAccImageBuilderDistributionConfiguration_Distribution_launchTemplateConfiguration(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	launchTemplateResourceName := "aws_launch_template.test"
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfigurationConfig_launchTemplateIDDefault(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "distribution.0.launch_template_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "distribution.0.launch_template_configuration.0.default", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "distribution.0.launch_template_configuration.0.launch_template_id", launchTemplateResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDistributionConfigurationConfig_launchTemplateIDNonDefault(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "distribution.0.launch_template_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "distribution.0.launch_template_configuration.0.default", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "distribution.0.launch_template_configuration.0.launch_template_id", launchTemplateResourceName, "id"),
				),
			},
			{
				Config: testAccDistributionConfigurationConfig_launchTemplateIDAccountID(rName, "111111111111"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "distribution.0.launch_template_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "distribution.0.launch_template_configuration.0.default", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "distribution.0.launch_template_configuration.0.launch_template_id", launchTemplateResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "distribution.0.launch_template_configuration.0.account_id", "111111111111"),
				),
			},
		},
	})
}

func TestAccImageBuilderDistributionConfiguration_Distribution_licenseARNs(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	licenseConfigurationResourceName := "aws_licensemanager_license_configuration.test"
	licenseConfigurationResourceName2 := "aws_licensemanager_license_configuration.test2"
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckIAMServiceLinkedRole(t, "/aws-service-role/license-manager.amazonaws.com")
		},
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfigurationConfig_licenseARNs1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "distribution.*.license_configuration_arns.*", licenseConfigurationResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDistributionConfigurationConfig_licenseARNs2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "distribution.*.license_configuration_arns.*", licenseConfigurationResourceName2, "id"),
				),
			},
		},
	})
}

func TestAccImageBuilderDistributionConfiguration_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfigurationConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDistributionConfigurationConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccDistributionConfigurationConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckDistributionConfigurationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ImageBuilderConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_imagebuilder_distribution_configuration" {
			continue
		}

		input := &imagebuilder.GetDistributionConfigurationInput{
			DistributionConfigurationArn: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetDistributionConfiguration(input)

		if tfawserr.ErrCodeEquals(err, imagebuilder.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error getting Image Builder Distribution Configuration (%s): %w", rs.Primary.ID, err)
		}

		if output != nil {
			return fmt.Errorf("Image Builder Distribution Configuration (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckDistributionConfigurationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ImageBuilderConn

		input := &imagebuilder.GetDistributionConfigurationInput{
			DistributionConfigurationArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetDistributionConfiguration(input)

		if err != nil {
			return fmt.Errorf("error getting Image Builder Distribution Configuration (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccDistributionConfigurationConfig_description(rName string, description string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_imagebuilder_distribution_configuration" "test" {
  description = %[2]q
  name        = %[1]q

  distribution {
    ami_distribution_configuration {
      name = "{{ imagebuilder:buildDate }}"
    }

    region = data.aws_region.current.name
  }
}
`, rName, description)
}

func testAccDistributionConfigurationConfig_2(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_region" "alternate" {
  provider = awsalternate
}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    ami_distribution_configuration {
      name = "{{ imagebuilder:buildDate }}"
    }

    region = data.aws_region.current.name
  }

  distribution {
    ami_distribution_configuration {
      name = "{{ imagebuilder:buildDate }}"
    }

    region = data.aws_region.alternate.name
  }
}
`, rName))
}

func testAccDistributionConfigurationConfig_amiTags(rName string, amiTagKey string, amiTagValue string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    ami_distribution_configuration {
      ami_tags = {
        %[2]q = %[3]q
      }
    }

    region = data.aws_region.current.name
  }
}
`, rName, amiTagKey, amiTagValue)
}

func testAccDistributionConfigurationConfig_amiDescription(rName string, description string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    ami_distribution_configuration {
      description = %[2]q
    }

    region = data.aws_region.current.name
  }
}
`, rName, description)
}

func testAccDistributionConfigurationConfig_amiKMSKeyID1(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

data "aws_region" "current" {}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    ami_distribution_configuration {
      kms_key_id = aws_kms_key.test.arn
    }

    region = data.aws_region.current.name
  }
}
`, rName)
}

func testAccDistributionConfigurationConfig_amiKMSKeyID2(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test2" {
  deletion_window_in_days = 7
}

data "aws_region" "current" {}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    ami_distribution_configuration {
      kms_key_id = aws_kms_key.test2.arn
    }

    region = data.aws_region.current.name
  }
}
`, rName)
}

func testAccDistributionConfigurationConfig_amiLaunchPermissionUserGroups(rName string, userGroup string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    ami_distribution_configuration {
      launch_permission {
        user_groups = [%[2]q]
      }
    }

    region = data.aws_region.current.name
  }
}
`, rName, userGroup)
}

func testAccDistributionConfigurationConfig_amiLaunchPermissionUserIDs(rName string, userId string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    ami_distribution_configuration {
      launch_permission {
        user_ids = [%[2]q]
      }
    }

    region = data.aws_region.current.name
  }
}
`, rName, userId)
}

func testAccDistributionConfigurationConfig_amiLaunchPermissionOrganizationARNs(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}
data "aws_partition" "current" {}

resource "aws_organizations_organization" "test" {}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q
  distribution {
    ami_distribution_configuration {
      launch_permission {
        organization_arns = [aws_organizations_organization.test.arn]
      }
    }
    region = data.aws_region.current.name
  }
}
`, rName)
}

func testAccDistributionConfigurationConfig_amiLaunchPermissionOrganizationalUnitARNs(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}
data "aws_partition" "current" {}

resource "aws_organizations_organization" "test" {}

resource "aws_organizations_organizational_unit" "test" {
  name      = %[1]q
  parent_id = aws_organizations_organization.test.roots[0].id
}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q
  distribution {
    ami_distribution_configuration {
      launch_permission {
        organizational_unit_arns = [aws_organizations_organizational_unit.test.arn]
      }
    }
    region = data.aws_region.current.name
  }
}
  `, rName)
}

func testAccDistributionConfigurationConfig_amiName(rName string, name string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    ami_distribution_configuration {
      name = %[2]q
    }

    region = data.aws_region.current.name
  }
}
`, rName, name)
}

func testAccDistributionConfigurationConfig_amiTargetAccountIDs(rName string, targetAccountId string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    ami_distribution_configuration {
      target_account_ids = [%[2]q]
    }

    region = data.aws_region.current.name
  }
}
`, rName, targetAccountId)
}

func testAccDistributionConfigurationConfig_containerTargetRepository(rName string, repositoryName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    container_distribution_configuration {
      target_repository {
        repository_name = %[2]q
        service         = "ECR"
      }
    }

    region = data.aws_region.current.name
  }
}
`, rName, repositoryName)
}

func testAccDistributionConfigurationConfig_containerDescription(rName string, description string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    container_distribution_configuration {
      target_repository {
        repository_name = "repository-name"
        service         = "ECR"
      }

      description = %[2]q
    }

    region = data.aws_region.current.name
  }
}
`, rName, description)
}

func testAccDistributionConfigurationConfig_containerTags(rName string, containerTag string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    container_distribution_configuration {
      target_repository {
        repository_name = "repository-name"
        service         = "ECR"
      }

      container_tags = [%[2]q]
    }

    region = data.aws_region.current.name
  }
}
`, rName, containerTag)
}

func testAccDistributionConfigurationConfig_launchTemplateIDDefault(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_caller_identity" "current" {}

resource "aws_launch_template" "test" {
  instance_type = "t2.micro"
  name          = %[1]q
}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    launch_template_configuration {
      default            = true
      launch_template_id = aws_launch_template.test.id
      account_id         = data.aws_caller_identity.current.account_id
    }

    region = data.aws_region.current.name
  }
}
`, rName)
}

func testAccDistributionConfigurationConfig_launchTemplateIDNonDefault(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_caller_identity" "current" {}

resource "aws_launch_template" "test" {
  instance_type = "t2.micro"
  name          = %[1]q
}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    launch_template_configuration {
      default            = false
      launch_template_id = aws_launch_template.test.id
      account_id         = data.aws_caller_identity.current.account_id
    }

    region = data.aws_region.current.name
  }
}
`, rName)
}

func testAccDistributionConfigurationConfig_launchTemplateIDAccountID(rName string, accountId string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_launch_template" "test" {
  instance_type = "t2.micro"
  name          = %[1]q
}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    launch_template_configuration {
      default            = false
      launch_template_id = aws_launch_template.test.id
      account_id         = %[2]q
    }

    ami_distribution_configuration {
      launch_permission {
        user_ids = [%[2]q]
      }
    }

    region = data.aws_region.current.name
  }
}
  `, rName, accountId)
}

func testAccDistributionConfigurationConfig_licenseARNs1(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_licensemanager_license_configuration" "test" {
  name                  = %[1]q
  license_counting_type = "Socket"
}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    license_configuration_arns = [aws_licensemanager_license_configuration.test.id]
    region                     = data.aws_region.current.name
  }
}
`, rName)
}

func testAccDistributionConfigurationConfig_licenseARNs2(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_licensemanager_license_configuration" "test2" {
  name                  = %[1]q
  license_counting_type = "Socket"
}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    license_configuration_arns = [aws_licensemanager_license_configuration.test2.id]
    region                     = data.aws_region.current.name
  }
}
`, rName)
}

func testAccDistributionConfigurationConfig_name(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    ami_distribution_configuration {
      name = "{{ imagebuilder:buildDate }}"
    }

    region = data.aws_region.current.name
  }
}
`, rName)
}

func testAccDistributionConfigurationConfig_tags1(rName string, tagKey1 string, tagValue1 string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    ami_distribution_configuration {
      name = "{{ imagebuilder:buildDate }}"
    }

    region = data.aws_region.current.name
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccDistributionConfigurationConfig_tags2(rName string, tagKey1 string, tagValue1 string, tagKey2 string, tagValue2 string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    ami_distribution_configuration {
      name = "{{ imagebuilder:buildDate }}"
    }

    region = data.aws_region.current.name
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
