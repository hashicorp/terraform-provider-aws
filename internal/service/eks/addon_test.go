package eks_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfeks "github.com/hashicorp/terraform-provider-aws/internal/service/eks"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccEKSAddon_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var addon eks.Addon
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	clusterResourceName := "aws_eks_cluster.test"
	addonResourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(ctx, t); testAccPreCheckAddon(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, eks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAddonDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAddonConfig_basic(rName, addonName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAddonExists(ctx, addonResourceName, &addon),
					resource.TestCheckResourceAttr(addonResourceName, "addon_name", addonName),
					resource.TestCheckResourceAttrSet(addonResourceName, "addon_version"),
					acctest.MatchResourceAttrRegionalARN(addonResourceName, "arn", "eks", regexp.MustCompile(fmt.Sprintf("addon/%s/%s/.+$", rName, addonName))),
					resource.TestCheckResourceAttrPair(addonResourceName, "cluster_name", clusterResourceName, "name"),
					resource.TestCheckResourceAttr(addonResourceName, "configuration_values", ""),
					resource.TestCheckNoResourceAttr(addonResourceName, "preserve"),
					resource.TestCheckResourceAttr(addonResourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      addonResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEKSAddon_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var addon eks.Addon
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(ctx, t); testAccPreCheckAddon(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, eks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAddonDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAddonConfig_basic(rName, addonName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfeks.ResourceAddon(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEKSAddon_Disappears_cluster(t *testing.T) {
	ctx := acctest.Context(t)
	var addon eks.Addon
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_addon.test"
	clusterResourceName := "aws_eks_cluster.test"
	addonName := "vpc-cni"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(ctx, t); testAccPreCheckAddon(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, eks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAddonDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAddonConfig_basic(rName, addonName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfeks.ResourceCluster(), clusterResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEKSAddon_addonVersion(t *testing.T) {
	ctx := acctest.Context(t)
	var addon1, addon2 eks.Addon
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"
	addonVersion1 := "v1.8.0-eksbuild.1"
	addonVersion2 := "v1.9.0-eksbuild.1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(ctx, t); testAccPreCheckAddon(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, eks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAddonDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAddonConfig_version(rName, addonName, addonVersion1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon1),
					resource.TestCheckResourceAttr(resourceName, "addon_version", addonVersion1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"resolve_conflicts"},
			},
			{
				Config: testAccAddonConfig_version(rName, addonName, addonVersion2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon2),
					resource.TestCheckResourceAttr(resourceName, "addon_version", addonVersion2),
				),
			},
		},
	})
}

func TestAccEKSAddon_preserve(t *testing.T) {
	ctx := acctest.Context(t)
	var addon eks.Addon
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(ctx, t); testAccPreCheckAddon(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, eks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAddonDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAddonConfig_preserve(rName, addonName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon),
					resource.TestCheckResourceAttr(resourceName, "preserve", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"preserve"},
			},
		},
	})
}

func TestAccEKSAddon_resolveConflicts(t *testing.T) {
	ctx := acctest.Context(t)
	var addon1, addon2, addon3 eks.Addon
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(ctx, t); testAccPreCheckAddon(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, eks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAddonDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAddonConfig_resolveConflicts(rName, addonName, eks.ResolveConflictsNone),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon1),
					resource.TestCheckResourceAttr(resourceName, "resolve_conflicts", eks.ResolveConflictsNone),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"resolve_conflicts"},
			},
			{
				Config: testAccAddonConfig_resolveConflicts(rName, addonName, eks.ResolveConflictsOverwrite),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon2),
					resource.TestCheckResourceAttr(resourceName, "resolve_conflicts", eks.ResolveConflictsOverwrite),
				),
			},
			{
				Config: testAccAddonConfig_resolveConflicts(rName, addonName, eks.ResolveConflictsPreserve),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon3),
					resource.TestCheckResourceAttr(resourceName, "resolve_conflicts", eks.ResolveConflictsPreserve),
				),
			},
		},
	})
}

func TestAccEKSAddon_serviceAccountRoleARN(t *testing.T) {
	ctx := acctest.Context(t)
	var addon eks.Addon
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_addon.test"
	serviceRoleResourceName := "aws_iam_role.test-service-role"
	addonName := "vpc-cni"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(ctx, t); testAccPreCheckAddon(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, eks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAddonDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAddonConfig_serviceAccountRoleARN(rName, addonName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon),
					resource.TestCheckResourceAttrPair(resourceName, "service_account_role_arn", serviceRoleResourceName, "arn"),
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

func TestAccEKSAddon_configurationValues(t *testing.T) {
	ctx := acctest.Context(t)
	var addon eks.Addon
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_addon.test"
	configurationValues := "{\"env\": {\"WARM_ENI_TARGET\":\"2\",\"ENABLE_POD_ENI\":\"true\"},\"resources\": {\"limits\":{\"cpu\":\"100m\",\"memory\":\"100Mi\"},\"requests\":{\"cpu\":\"100m\",\"memory\":\"100Mi\"}}}"
	updateConfigurationValues := "{\"env\": {\"WARM_ENI_TARGET\":\"2\",\"ENABLE_POD_ENI\":\"true\"},\"resources\": {\"limits\":{\"cpu\":\"200m\",\"memory\":\"150Mi\"},\"requests\":{\"cpu\":\"200m\",\"memory\":\"150Mi\"}}}"
	emptyConfigurationValues := "{}"
	invalidConfigurationValues := "{\"env\": {\"INVALID_FIELD\":\"2\"}}"
	addonName := "vpc-cni"
	addonVersion := "v1.10.4-eksbuild.1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(ctx, t); testAccPreCheckAddon(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, eks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAddonDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAddonConfig_configurationValues(rName, addonName, addonVersion, configurationValues, eks.ResolveConflictsOverwrite),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon),
					resource.TestCheckResourceAttr(resourceName, "configuration_values", configurationValues),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"resolve_conflicts"},
			},
			{
				Config: testAccAddonConfig_configurationValues(rName, addonName, addonVersion, updateConfigurationValues, eks.ResolveConflictsOverwrite),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon),
					resource.TestCheckResourceAttr(resourceName, "configuration_values", updateConfigurationValues),
				),
			},
			{
				Config: testAccAddonConfig_configurationValues(rName, addonName, addonVersion, emptyConfigurationValues, eks.ResolveConflictsOverwrite),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon),
					resource.TestCheckResourceAttr(resourceName, "configuration_values", emptyConfigurationValues),
				),
			},
			{
				Config:      testAccAddonConfig_configurationValues(rName, addonName, addonVersion, invalidConfigurationValues, eks.ResolveConflictsOverwrite),
				ExpectError: regexp.MustCompile(`InvalidParameterException: ConfigurationValue provided in request is not supported`),
			},
		},
	})
}

func TestAccEKSAddon_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var addon1, addon2, addon3 eks.Addon
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(ctx, t); testAccPreCheckAddon(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, eks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAddonDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAddonConfig_tags1(rName, addonName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon1),
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
				Config: testAccAddonConfig_tags2(rName, addonName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAddonConfig_tags1(rName, addonName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccEKSAddon_DefaultTags_providerOnly(t *testing.T) {
	ctx := acctest.Context(t)
	var addon eks.Addon
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, eks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAddonDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1("providerkey1", "providervalue1"),
					testAccAddonConfig_basic(rName, addonName),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey1", "providervalue1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags2("providerkey1", "providervalue1", "providerkey2", "providervalue2"),
					testAccAddonConfig_basic(rName, addonName),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey1", "providervalue1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey2", "providervalue2"),
				),
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1("providerkey1", "value1"),
					testAccAddonConfig_basic(rName, addonName),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey1", "value1"),
				),
			},
		},
	})
}

func TestAccEKSAddon_DefaultTags_updateToProviderOnly(t *testing.T) {
	ctx := acctest.Context(t)
	var addon eks.Addon
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, eks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAddonDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAddonConfig_tags1(rName, addonName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1"),
				),
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1("key1", "value1"),
					testAccAddonConfig_basic(rName, addonName),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1"),
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

func TestAccEKSAddon_DefaultTags_updateToResourceOnly(t *testing.T) {
	ctx := acctest.Context(t)
	var addon eks.Addon
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, eks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAddonDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1("key1", "value1"),
					testAccAddonConfig_basic(rName, addonName),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1"),
				),
			},
			{
				Config: testAccAddonConfig_tags1(rName, addonName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1"),
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

func TestAccEKSAddon_DefaultTagsProviderAndResource_nonOverlappingTag(t *testing.T) {
	ctx := acctest.Context(t)
	var addon eks.Addon
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, eks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAddonDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1("providerkey1", "providervalue1"),
					testAccAddonConfig_tags1(rName, addonName, "resourcekey1", "resourcevalue1"),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.resourcekey1", "resourcevalue1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey1", "providervalue1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.resourcekey1", "resourcevalue1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1("providerkey1", "providervalue1"),
					testAccAddonConfig_tags2(rName, addonName, "resourcekey1", "resourcevalue1", "resourcekey2", "resourcevalue2"),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.resourcekey1", "resourcevalue1"),
					resource.TestCheckResourceAttr(resourceName, "tags.resourcekey2", "resourcevalue2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey1", "providervalue1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.resourcekey1", "resourcevalue1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.resourcekey2", "resourcevalue2"),
				),
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1("providerkey2", "providervalue2"),
					testAccAddonConfig_tags1(rName, addonName, "resourcekey3", "resourcevalue3"),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.resourcekey3", "resourcevalue3"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey2", "providervalue2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.resourcekey3", "resourcevalue3"),
				),
			},
		},
	})
}

func TestAccEKSAddon_DefaultTagsProviderAndResource_overlappingTag(t *testing.T) {
	ctx := acctest.Context(t)
	var addon eks.Addon
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, eks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAddonDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1("overlapkey1", "providervalue1"),
					testAccAddonConfig_tags1(rName, addonName, "overlapkey1", "resourcevalue1"),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.overlapkey1", "resourcevalue1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags2("overlapkey1", "providervalue1", "overlapkey2", "providervalue2"),
					testAccAddonConfig_tags2(rName, addonName, "overlapkey1", "resourcevalue1", "overlapkey2", "resourcevalue2"),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.overlapkey1", "resourcevalue1"),
					resource.TestCheckResourceAttr(resourceName, "tags.overlapkey2", "resourcevalue2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.overlapkey1", "resourcevalue1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.overlapkey2", "resourcevalue2"),
				),
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1("overlapkey1", "providervalue1"),
					testAccAddonConfig_tags1(rName, addonName, "overlapkey1", "resourcevalue2"),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.overlapkey1", "resourcevalue2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.overlapkey1", "resourcevalue2"),
				),
			},
		},
	})
}

func TestAccEKSAddon_DefaultTagsProviderAndResource_duplicateTag(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	addonName := "vpc-cni"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, eks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1("overlapkey", "overlapvalue"),
					testAccAddonConfig_tags1(rName, addonName, "overlapkey", "overlapvalue"),
				),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`"tags" are identical to those in the "default_tags" configuration block`),
			},
		},
	})
}

func TestAccEKSAddon_defaultAndIgnoreTags(t *testing.T) {
	ctx := acctest.Context(t)
	var addon eks.Addon
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, eks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAddonDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAddonConfig_tags1(rName, addonName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon),
					testAccCheckAddonUpdateTags(ctx, &addon, nil, map[string]string{"defaultkey1": "defaultvalue1"}),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultAndIgnoreTagsKeyPrefixes1("defaultkey1", "defaultvalue1", "defaultkey"),
					testAccAddonConfig_tags1(rName, addonName, "key1", "value1"),
				),
				PlanOnly: true,
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultAndIgnoreTagsKeys1("defaultkey1", "defaultvalue1"),
					testAccAddonConfig_tags1(rName, addonName, "key1", "value1"),
				),
				PlanOnly: true,
			},
		},
	})
}

func TestAccEKSAddon_ignoreTags(t *testing.T) {
	ctx := acctest.Context(t)
	var addon eks.Addon
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, eks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAddonDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAddonConfig_tags1(rName, addonName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon),
					testAccCheckAddonUpdateTags(ctx, &addon, nil, map[string]string{"ignorekey1": "ignorevalue1"}),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigIgnoreTagsKeyPrefixes1("ignorekey"),
					testAccAddonConfig_tags1(rName, addonName, "key1", "value1"),
				),
				PlanOnly: true,
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigIgnoreTagsKeys("ignorekey1"),
					testAccAddonConfig_tags1(rName, addonName, "key1", "value1"),
				),
				PlanOnly: true,
			},
		},
	})
}

func testAccCheckAddonExists(ctx context.Context, n string, v *eks.Addon) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EKS Add-On ID is set")
		}

		clusterName, addonName, err := tfeks.AddonParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EKSConn()

		output, err := tfeks.FindAddonByClusterNameAndAddonName(ctx, conn, clusterName, addonName)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckAddonDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EKSConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_eks_addon" {
				continue
			}

			clusterName, addonName, err := tfeks.AddonParseResourceID(rs.Primary.ID)

			if err != nil {
				return err
			}

			_, err = tfeks.FindAddonByClusterNameAndAddonName(ctx, conn, clusterName, addonName)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EKS Add-On %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccPreCheckAddon(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EKSConn()

	input := &eks.DescribeAddonVersionsInput{}

	_, err := conn.DescribeAddonVersionsWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckAddonUpdateTags(ctx context.Context, addon *eks.Addon, oldTags, newTags map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EKSConn()

		return tfeks.UpdateTags(ctx, conn, aws.StringValue(addon.AddonArn), oldTags, newTags)
	}
}

func testAccAddonConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "eks.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "test-AmazonEKSClusterPolicy" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.test.name
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name                          = %[1]q
    "kubernetes.io/cluster/%[1]s" = "shared"
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name                          = %[1]q
    "kubernetes.io/cluster/%[1]s" = "shared"
  }
}

resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName))
}

func testAccAddonConfig_basic(rName, addonName string) string {
	return acctest.ConfigCompose(testAccAddonConfig_base(rName), fmt.Sprintf(`
resource "aws_eks_addon" "test" {
  cluster_name = aws_eks_cluster.test.name
  addon_name   = %[2]q
}
`, rName, addonName))
}

func testAccAddonConfig_version(rName, addonName, addonVersion string) string {
	return acctest.ConfigCompose(testAccAddonConfig_base(rName), fmt.Sprintf(`
resource "aws_eks_addon" "test" {
  cluster_name      = aws_eks_cluster.test.name
  addon_name        = %[2]q
  addon_version     = %[3]q
  resolve_conflicts = "OVERWRITE"
}
`, rName, addonName, addonVersion))
}

func testAccAddonConfig_preserve(rName, addonName string) string {
	return acctest.ConfigCompose(testAccAddonConfig_base(rName), fmt.Sprintf(`
resource "aws_eks_addon" "test" {
  cluster_name = aws_eks_cluster.test.name
  addon_name   = %[2]q
  preserve     = true
}
`, rName, addonName))
}

func testAccAddonConfig_resolveConflicts(rName, addonName, resolveConflicts string) string {
	return acctest.ConfigCompose(testAccAddonConfig_base(rName), fmt.Sprintf(`
resource "aws_eks_addon" "test" {
  cluster_name      = aws_eks_cluster.test.name
  addon_name        = %[2]q
  resolve_conflicts = %[3]q
}
`, rName, addonName, resolveConflicts))
}

func testAccAddonConfig_serviceAccountRoleARN(rName, addonName string) string {
	return acctest.ConfigCompose(testAccAddonConfig_base(rName), fmt.Sprintf(`
resource "aws_iam_role" "test-service-role" {
  name               = "test-service-role"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_eks_addon" "test" {
  cluster_name             = aws_eks_cluster.test.name
  addon_name               = %[2]q
  service_account_role_arn = aws_iam_role.test-service-role.arn
}
`, rName, addonName))
}

func testAccAddonConfig_tags1(rName, addonName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccAddonConfig_base(rName), fmt.Sprintf(`
resource "aws_eks_addon" "test" {
  cluster_name = aws_eks_cluster.test.name
  addon_name   = %[2]q

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, addonName, tagKey1, tagValue1))
}

func testAccAddonConfig_tags2(rName, addonName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccAddonConfig_base(rName), fmt.Sprintf(`
resource "aws_eks_addon" "test" {
  cluster_name = aws_eks_cluster.test.name
  addon_name   = %[2]q

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, addonName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccAddonConfig_configurationValues(rName, addonName, addonVersion, configurationValues, resolveConflicts string) string {
	return acctest.ConfigCompose(testAccAddonConfig_base(rName), fmt.Sprintf(`
resource "aws_eks_addon" "test" {
  cluster_name         = aws_eks_cluster.test.name
  addon_name           = %[2]q
  addon_version        = %[3]q
  configuration_values = %[4]q
  resolve_conflicts    = %[5]q
}
`, rName, addonName, addonVersion, configurationValues, resolveConflicts))
}
