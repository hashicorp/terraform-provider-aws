// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfeks "github.com/hashicorp/terraform-provider-aws/internal/service/eks"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEKSAddon_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var addon types.Addon
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	clusterResourceName := "aws_eks_cluster.test"
	addonResourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t); testAccPreCheckAddon(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAddonDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAddonConfig_basic(rName, addonName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAddonExists(ctx, addonResourceName, &addon),
					resource.TestCheckResourceAttr(addonResourceName, "addon_name", addonName),
					resource.TestCheckResourceAttrSet(addonResourceName, "addon_version"),
					acctest.MatchResourceAttrRegionalARN(addonResourceName, names.AttrARN, "eks", regexache.MustCompile(fmt.Sprintf("addon/%s/%s/.+$", rName, addonName))),
					resource.TestCheckResourceAttrPair(addonResourceName, names.AttrClusterName, clusterResourceName, names.AttrName),
					resource.TestCheckResourceAttr(addonResourceName, "configuration_values", ""),
					resource.TestCheckNoResourceAttr(addonResourceName, "preserve"),
					resource.TestCheckResourceAttr(addonResourceName, acctest.CtTagsPercent, acctest.Ct0),
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
	var addon types.Addon
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t); testAccPreCheckAddon(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
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
	var addon types.Addon
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_addon.test"
	clusterResourceName := "aws_eks_cluster.test"
	addonName := "vpc-cni"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t); testAccPreCheckAddon(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
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
	var addon1, addon2 types.Addon
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"
	addonVersion1 := "v1.14.1-eksbuild.1"
	addonVersion2 := "v1.15.3-eksbuild.1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t); testAccPreCheckAddon(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
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
				ImportStateVerifyIgnore: []string{"resolve_conflicts_on_create", "resolve_conflicts_on_update"},
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
	var addon types.Addon
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t); testAccPreCheckAddon(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAddonDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAddonConfig_preserve(rName, addonName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon),
					resource.TestCheckResourceAttr(resourceName, "preserve", acctest.CtTrue),
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

func TestAccEKSAddon_deprecated(t *testing.T) {
	ctx := acctest.Context(t)
	var addon1, addon2, addon3 types.Addon
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t); testAccPreCheckAddon(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAddonDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAddonConfig_deprecated(rName, addonName, string(types.ResolveConflictsNone)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon1),
					resource.TestCheckResourceAttr(resourceName, "resolve_conflicts", string(types.ResolveConflictsNone)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"resolve_conflicts"},
			},
			{
				Config: testAccAddonConfig_deprecated(rName, addonName, string(types.ResolveConflictsOverwrite)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon2),
					resource.TestCheckResourceAttr(resourceName, "resolve_conflicts", string(types.ResolveConflictsOverwrite)),
				),
			},
			{
				Config: testAccAddonConfig_deprecated(rName, addonName, string(types.ResolveConflictsPreserve)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon3),
					resource.TestCheckResourceAttr(resourceName, "resolve_conflicts", string(types.ResolveConflictsPreserve)),
				),
			},
		},
	})
}

func TestAccEKSAddon_resolveConflicts(t *testing.T) {
	ctx := acctest.Context(t)
	var addon1, addon2, addon3 types.Addon
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t); testAccPreCheckAddon(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAddonDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAddonConfig_resolveConflicts(rName, addonName, string(types.ResolveConflictsNone), string(types.ResolveConflictsNone)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon1),
					resource.TestCheckResourceAttr(resourceName, "resolve_conflicts_on_create", string(types.ResolveConflictsNone)),
					resource.TestCheckResourceAttr(resourceName, "resolve_conflicts_on_update", string(types.ResolveConflictsNone)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"resolve_conflicts_on_create", "resolve_conflicts_on_update"},
			},
			{
				Config: testAccAddonConfig_resolveConflicts(rName, addonName, string(types.ResolveConflictsOverwrite), string(types.ResolveConflictsOverwrite)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon2),
					resource.TestCheckResourceAttr(resourceName, "resolve_conflicts_on_create", string(types.ResolveConflictsOverwrite)),
					resource.TestCheckResourceAttr(resourceName, "resolve_conflicts_on_update", string(types.ResolveConflictsOverwrite)),
				),
			},
			{
				Config: testAccAddonConfig_resolveConflicts(rName, addonName, string(types.ResolveConflictsOverwrite), string(types.ResolveConflictsPreserve)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon3),
					resource.TestCheckResourceAttr(resourceName, "resolve_conflicts_on_create", string(types.ResolveConflictsOverwrite)),
					resource.TestCheckResourceAttr(resourceName, "resolve_conflicts_on_update", string(types.ResolveConflictsPreserve)),
				),
			},
		},
	})
}

func TestAccEKSAddon_serviceAccountRoleARN(t *testing.T) {
	ctx := acctest.Context(t)
	var addon types.Addon
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_addon.test"
	serviceRoleResourceName := "aws_iam_role.test-service-role"
	addonName := "vpc-cni"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t); testAccPreCheckAddon(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAddonDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAddonConfig_serviceAccountRoleARN(rName, addonName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon),
					resource.TestCheckResourceAttrPair(resourceName, "service_account_role_arn", serviceRoleResourceName, names.AttrARN),
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
	var addon types.Addon
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_addon.test"
	configurationValues := "{\"env\": {\"WARM_ENI_TARGET\":\"2\",\"ENABLE_POD_ENI\":\"true\"},\"resources\": {\"limits\":{\"cpu\":\"100m\",\"memory\":\"100Mi\"},\"requests\":{\"cpu\":\"100m\",\"memory\":\"100Mi\"}}}"
	updateConfigurationValues := "{\"env\": {\"WARM_ENI_TARGET\":\"2\",\"ENABLE_POD_ENI\":\"true\"},\"resources\": {\"limits\":{\"cpu\":\"200m\",\"memory\":\"150Mi\"},\"requests\":{\"cpu\":\"200m\",\"memory\":\"150Mi\"}}}"
	emptyConfigurationValues := "{}"
	invalidConfigurationValues := "{\"env\": {\"INVALID_FIELD\":\"2\"}}"
	addonName := "vpc-cni"
	addonVersion := "v1.15.3-eksbuild.1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t); testAccPreCheckAddon(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAddonDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAddonConfig_configurationValues(rName, addonName, addonVersion, configurationValues, string(types.ResolveConflictsOverwrite)),
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
				Config: testAccAddonConfig_configurationValues(rName, addonName, addonVersion, updateConfigurationValues, string(types.ResolveConflictsOverwrite)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon),
					resource.TestCheckResourceAttr(resourceName, "configuration_values", updateConfigurationValues),
				),
			},
			{
				Config: testAccAddonConfig_configurationValues(rName, addonName, addonVersion, emptyConfigurationValues, string(types.ResolveConflictsOverwrite)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon),
					resource.TestCheckResourceAttr(resourceName, "configuration_values", emptyConfigurationValues),
				),
			},
			{
				Config:      testAccAddonConfig_configurationValues(rName, addonName, addonVersion, invalidConfigurationValues, string(types.ResolveConflictsOverwrite)),
				ExpectError: regexache.MustCompile(`InvalidParameterException: ConfigurationValue provided in request is not supported`),
			},
		},
	})
}

func TestAccEKSAddon_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var addon1, addon2, addon3 types.Addon
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t); testAccPreCheckAddon(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAddonDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAddonConfig_tags1(rName, addonName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon1),
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
				Config: testAccAddonConfig_tags2(rName, addonName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccAddonConfig_tags1(rName, addonName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, resourceName, &addon3),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckAddonExists(ctx context.Context, n string, v *types.Addon) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		clusterName, addonName, err := tfeks.AddonParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EKSClient(ctx)

		output, err := tfeks.FindAddonByTwoPartKey(ctx, conn, clusterName, addonName)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckAddonDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EKSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_eks_addon" {
				continue
			}

			clusterName, addonName, err := tfeks.AddonParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfeks.FindAddonByTwoPartKey(ctx, conn, clusterName, addonName)

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
	conn := acctest.Provider.Meta().(*conns.AWSClient).EKSClient(ctx)

	input := &eks.DescribeAddonVersionsInput{}

	_, err := conn.DescribeAddonVersions(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
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
  cluster_name                = aws_eks_cluster.test.name
  addon_name                  = %[2]q
  addon_version               = %[3]q
  resolve_conflicts_on_create = "OVERWRITE"
  resolve_conflicts_on_update = "OVERWRITE"
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

func testAccAddonConfig_deprecated(rName, addonName, resolveConflicts string) string {
	return acctest.ConfigCompose(testAccAddonConfig_base(rName), fmt.Sprintf(`
resource "aws_eks_addon" "test" {
  cluster_name      = aws_eks_cluster.test.name
  addon_name        = %[2]q
  resolve_conflicts = %[3]q
}
`, rName, addonName, resolveConflicts))
}

func testAccAddonConfig_resolveConflicts(rName, addonName, resolveConflictsOnCreate, resolveConflictsOnUpdate string) string {
	return acctest.ConfigCompose(testAccAddonConfig_base(rName), fmt.Sprintf(`
resource "aws_eks_addon" "test" {
  cluster_name                = aws_eks_cluster.test.name
  addon_name                  = %[2]q
  resolve_conflicts_on_create = %[3]q
  resolve_conflicts_on_update = %[4]q
}
`, rName, addonName, resolveConflictsOnCreate, resolveConflictsOnUpdate))
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
