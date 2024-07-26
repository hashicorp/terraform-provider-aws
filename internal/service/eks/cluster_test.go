// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfeks "github.com/hashicorp/terraform-provider-aws/internal/service/eks"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	clusterVersionUpgradeInitial = "1.27"
	clusterVersionUpgradeUpdated = "1.28"
)

func TestAccEKSCluster_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster types.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "access_config.#", acctest.Ct1),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "eks", regexache.MustCompile(fmt.Sprintf("cluster/%s$", rName))),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_self_managed_addons", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "certificate_authority.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_authority.0.data"),
					resource.TestCheckNoResourceAttr(resourceName, "cluster_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttr(resourceName, "enabled_cluster_log_types.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "encryption_config.#", acctest.Ct0),
					resource.TestMatchResourceAttr(resourceName, names.AttrEndpoint, regexache.MustCompile(`^https://`)),
					resource.TestCheckResourceAttr(resourceName, "identity.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "identity.0.oidc.#", acctest.Ct1),
					resource.TestMatchResourceAttr(resourceName, "identity.0.oidc.0.issuer", regexache.MustCompile(`^https://`)),
					resource.TestCheckResourceAttr(resourceName, "kubernetes_network_config.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "kubernetes_network_config.0.service_ipv4_cidr"),
					resource.TestCheckResourceAttr(resourceName, "kubernetes_network_config.0.ip_family", "ipv4"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "outpost_config.#", acctest.Ct0),
					resource.TestMatchResourceAttr(resourceName, "platform_version", regexache.MustCompile(`^eks\.\d+$`)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.ClusterStatusActive)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestMatchResourceAttr(resourceName, names.AttrVersion, regexache.MustCompile(`^\d+\.\d+$`)),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.endpoint_private_access", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.endpoint_public_access", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnet_ids.#", acctest.Ct2),
					resource.TestMatchResourceAttr(resourceName, "vpc_config.0.vpc_id", regexache.MustCompile(`^vpc-.+`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bootstrap_self_managed_addons"},
			},
		},
	})
}

func TestAccEKSCluster_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster types.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfeks.ResourceCluster(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEKSCluster_AccessConfig_create(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster types.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_accessConfig(rName, types.AuthenticationModeConfigMap),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "access_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_config.0.authentication_mode", string(types.AuthenticationModeConfigMap)),
					resource.TestCheckResourceAttr(resourceName, "access_config.0.bootstrap_cluster_creator_admin_permissions", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       false,
				ImportStateVerifyIgnore: []string{"bootstrap_self_managed_addons"},
			},
		},
	})
}

func TestAccEKSCluster_AccessConfig_update(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster types.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "access_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_config.0.authentication_mode", string(types.AuthenticationModeConfigMap)),
					resource.TestCheckResourceAttr(resourceName, "access_config.0.bootstrap_cluster_creator_admin_permissions", acctest.CtTrue),
				),
			},
			{
				Config: testAccClusterConfig_accessConfig(rName, types.AuthenticationModeConfigMap),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "access_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_config.0.authentication_mode", string(types.AuthenticationModeConfigMap)),
					resource.TestCheckResourceAttr(resourceName, "access_config.0.bootstrap_cluster_creator_admin_permissions", acctest.CtTrue),
				),
			},
			{
				Config: testAccClusterConfig_accessConfig(rName, types.AuthenticationModeApiAndConfigMap),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "access_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_config.0.authentication_mode", string(types.AuthenticationModeApiAndConfigMap)),
					resource.TestCheckResourceAttr(resourceName, "access_config.0.bootstrap_cluster_creator_admin_permissions", acctest.CtTrue),
				),
			},
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "access_config.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccEKSCluster_BootstrapSelfManagedAddons_update(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster types.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_bootstrapSelfManagedAddons(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_self_managed_addons", acctest.CtFalse),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bootstrap_self_managed_addons"},
			},
			{
				Config: testAccClusterConfig_bootstrapSelfManagedAddons(rName, true),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_self_managed_addons", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccEKSCluster_BootstrapSelfManagedAddons_migrate(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster1, cluster2 types.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.EKSServiceID),
		CheckDestroy: testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.56.1",
					},
				},
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster1),
					resource.TestCheckNoResourceAttr(resourceName, "bootstrap_self_managed_addons"),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster2),
					testAccCheckClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_self_managed_addons", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccEKSCluster_Encryption_create(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster types.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_cluster.test"
	kmsKeyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_encryption(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "encryption_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "encryption_config.0.provider.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_config.0.provider.0.key_arn", kmsKeyResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "encryption_config.0.resources.#", acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bootstrap_self_managed_addons"},
			},
		},
	})
}

func TestAccEKSCluster_Encryption_update(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster1, cluster2 types.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_cluster.test"
	kmsKeyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "encryption_config.#", acctest.Ct0),
				),
			},
			{
				Config: testAccClusterConfig_encryption(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster2),
					testAccCheckClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "encryption_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "encryption_config.0.provider.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_config.0.provider.0.key_arn", kmsKeyResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "encryption_config.0.resources.#", acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bootstrap_self_managed_addons"},
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/19968.
func TestAccEKSCluster_Encryption_versionUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster1, cluster2 types.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_cluster.test"
	kmsKeyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_encryptionVersion(rName, clusterVersionUpgradeInitial),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "encryption_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "encryption_config.0.provider.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_config.0.provider.0.key_arn", kmsKeyResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "encryption_config.0.resources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, clusterVersionUpgradeInitial),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bootstrap_self_managed_addons"},
			},
			{
				Config: testAccClusterConfig_encryptionVersion(rName, clusterVersionUpgradeUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster2),
					testAccCheckClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "encryption_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "encryption_config.0.provider.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_config.0.provider.0.key_arn", kmsKeyResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "encryption_config.0.resources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, clusterVersionUpgradeUpdated),
				),
			},
		},
	})
}

func TestAccEKSCluster_version(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster1, cluster2 types.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_version(rName, clusterVersionUpgradeInitial),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, clusterVersionUpgradeInitial),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bootstrap_self_managed_addons"},
			},
			{
				Config: testAccClusterConfig_version(rName, clusterVersionUpgradeUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster2),
					testAccCheckClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, clusterVersionUpgradeUpdated),
				),
			},
		},
	})
}

func TestAccEKSCluster_logging(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster1, cluster2 types.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_logging(rName, []string{"api"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "enabled_cluster_log_types.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cluster_log_types.*", "api"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bootstrap_self_managed_addons"},
			},
			{
				Config: testAccClusterConfig_logging(rName, []string{"api", "audit"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster2),
					testAccCheckClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "enabled_cluster_log_types.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cluster_log_types.*", "api"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cluster_log_types.*", "audit"),
				),
			},
			// Disable all log types.
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster2),
					testAccCheckClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "enabled_cluster_log_types.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccEKSCluster_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster1, cluster2, cluster3 types.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bootstrap_self_managed_addons"},
			},
			{
				Config: testAccClusterConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccClusterConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster3),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccEKSCluster_VPC_securityGroupIDs(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster types.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_vpcSecurityGroupIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bootstrap_self_managed_addons"},
			},
		},
	})
}

func TestAccEKSCluster_VPC_securityGroupIDsAndSubnetIDs_update(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster1, cluster2 types.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_vpcSecurityGroupIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnet_ids.#", acctest.Ct2),
				),
			},
			{
				Config: testAccClusterConfig_vpcSecurityGroupIDsAndSubnetIDsUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster2),
					testAccCheckClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnet_ids.#", acctest.Ct2),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bootstrap_self_managed_addons"},
			},
		},
	})
}

func TestAccEKSCluster_VPC_endpointPrivateAccess(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster1, cluster2, cluster3 types.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_vpcEndpointPrivateAccess(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.endpoint_private_access", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bootstrap_self_managed_addons"},
			},
			{
				Config: testAccClusterConfig_vpcEndpointPrivateAccess(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster2),
					testAccCheckClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.endpoint_private_access", acctest.CtFalse),
				),
			},
			{
				Config: testAccClusterConfig_vpcEndpointPrivateAccess(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster3),
					testAccCheckClusterNotRecreated(&cluster2, &cluster3),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.endpoint_private_access", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccEKSCluster_VPC_endpointPublicAccess(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster1, cluster2, cluster3 types.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_vpcEndpointPublicAccess(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.endpoint_public_access", acctest.CtFalse),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bootstrap_self_managed_addons"},
			},
			{
				Config: testAccClusterConfig_vpcEndpointPublicAccess(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster2),
					testAccCheckClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.endpoint_public_access", acctest.CtTrue),
				),
			},
			{
				Config: testAccClusterConfig_vpcEndpointPublicAccess(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster3),
					testAccCheckClusterNotRecreated(&cluster2, &cluster3),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.endpoint_public_access", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccEKSCluster_VPC_publicAccessCIDRs(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster types.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_vpcPublicAccessCIDRs(rName, `["1.2.3.4/32", "5.6.7.8/32"]`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.public_access_cidrs.#", acctest.Ct2),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bootstrap_self_managed_addons"},
			},
			{
				Config: testAccClusterConfig_vpcPublicAccessCIDRs(rName, `["4.3.2.1/32", "8.7.6.5/32"]`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.public_access_cidrs.#", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccEKSCluster_Network_serviceIPv4CIDR(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster1, cluster2 types.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccClusterConfig_networkServiceIPv4CIDR(rName, `"10.0.0.0/11"`),
				ExpectError: regexache.MustCompile(`expected .* to contain a network Value with between`),
			},
			{
				Config:      testAccClusterConfig_networkServiceIPv4CIDR(rName, `"10.0.0.0/25"`),
				ExpectError: regexache.MustCompile(`expected .* to contain a network Value with between`),
			},
			{
				Config:      testAccClusterConfig_networkServiceIPv4CIDR(rName, `"9.0.0.0/16"`),
				ExpectError: regexache.MustCompile(`must be within`),
			},
			{
				Config:      testAccClusterConfig_networkServiceIPv4CIDR(rName, `"172.14.0.0/24"`),
				ExpectError: regexache.MustCompile(`must be within`),
			},
			{
				Config:      testAccClusterConfig_networkServiceIPv4CIDR(rName, `"192.167.0.0/24"`),
				ExpectError: regexache.MustCompile(`must be within`),
			},
			{
				Config: testAccClusterConfig_networkServiceIPv4CIDR(rName, `"192.168.0.0/24"`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "kubernetes_network_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "kubernetes_network_config.0.service_ipv4_cidr", "192.168.0.0/24"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bootstrap_self_managed_addons"},
			},
			{
				Config:             testAccClusterConfig_networkServiceIPv4CIDR(rName, `"192.168.0.0/24"`),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				Config: testAccClusterConfig_networkServiceIPv4CIDR(rName, `"192.168.1.0/24"`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster2),
					testAccCheckClusterRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "kubernetes_network_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "kubernetes_network_config.0.service_ipv4_cidr", "192.168.1.0/24"),
				),
			},
		},
	})
}

func TestAccEKSCluster_Network_ipFamily(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster1, cluster2 types.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_networkIPFamily(rName, `"ipv6"`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "kubernetes_network_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "kubernetes_network_config.0.ip_family", "ipv6"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bootstrap_self_managed_addons"},
			},
			{
				Config:             testAccClusterConfig_networkIPFamily(rName, `"ipv6"`),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				Config: testAccClusterConfig_networkIPFamily(rName, `"ipv4"`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster2),
					testAccCheckClusterRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "kubernetes_network_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "kubernetes_network_config.0.ip_family", "ipv4"),
				),
			},
		},
	})
}

func TestAccEKSCluster_Outpost_create(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster types.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_cluster.test"
	controlPlaneInstanceType := "m5d.large"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_outpost(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestMatchResourceAttr(resourceName, "cluster_id", regexache.MustCompile(`^[0-9A-Fa-f]{8}\b-[0-9A-Fa-f]{4}\b-[0-9A-Fa-f]{4}\b-[0-9A-Fa-f]{4}\b-[0-9A-Fa-f]{12}$`)),
					resource.TestCheckResourceAttr(resourceName, "outpost_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "outpost_config.0.control_plane_instance_type", controlPlaneInstanceType),
					resource.TestCheckResourceAttr(resourceName, "outpost_config.0.outpost_arns.#", acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bootstrap_self_managed_addons"},
			},
		},
	})
}

func TestAccEKSCluster_Outpost_placement(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster types.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_cluster.test"
	controlPlaneInstanceType := "m5d.large"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_outpostPlacement(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestMatchResourceAttr(resourceName, "cluster_id", regexache.MustCompile(`^[0-9A-Fa-f]{8}\b-[0-9A-Fa-f]{4}\b-[0-9A-Fa-f]{4}\b-[0-9A-Fa-f]{4}\b-[0-9A-Fa-f]{12}$`)),
					resource.TestCheckResourceAttr(resourceName, "outpost_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "outpost_config.0.control_plane_instance_type", controlPlaneInstanceType),
					resource.TestCheckResourceAttr(resourceName, "outpost_config.0.outpost_arns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "outpost_config.0.control_plane_placement.#", acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bootstrap_self_managed_addons"},
			},
		},
	})
}

func testAccCheckClusterExists(ctx context.Context, n string, v *types.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EKSClient(ctx)

		output, err := tfeks.FindClusterByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckClusterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_eks_cluster" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).EKSClient(ctx)

			_, err := tfeks.FindClusterByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EKS Cluster %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckClusterRecreated(i, j *types.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToTime(i.CreatedAt).Equal(aws.ToTime(j.CreatedAt)) {
			return errors.New("EKS Cluster was not recreated")
		}

		return nil
	}
}

func testAccCheckClusterNotRecreated(i, j *types.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.ToTime(i.CreatedAt).Equal(aws.ToTime(j.CreatedAt)) {
			return errors.New("EKS Cluster was recreated")
		}

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EKSClient(ctx)

	input := &eks.ListClustersInput{}

	_, err := conn.ListClusters(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccClusterConfig_base(rName string) string {
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

  assign_generated_ipv6_cidr_block = true

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

  ipv6_cidr_block                 = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, count.index)
  assign_ipv6_address_on_creation = true

  tags = {
    Name                          = %[1]q
    "kubernetes.io/cluster/%[1]s" = "shared"
  }
}
`, rName))
}

func testAccClusterConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(rName), fmt.Sprintf(`
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

func testAccClusterConfig_accessConfig(rName string, authenticationMode types.AuthenticationMode) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(rName), fmt.Sprintf(`
resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  access_config {
    bootstrap_cluster_creator_admin_permissions = true
    authentication_mode                         = %[2]q
  }

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName, authenticationMode))
}

func testAccClusterConfig_bootstrapSelfManagedAddons(rName string, bootstrapSelfManagedAddons bool) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(rName), fmt.Sprintf(`
resource "aws_eks_cluster" "test" {
  name                          = %[1]q
  role_arn                      = aws_iam_role.test.arn
  bootstrap_self_managed_addons = %[2]t

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName, bootstrapSelfManagedAddons))
}

func testAccClusterConfig_version(rName, version string) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(rName), fmt.Sprintf(`
resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn
  version  = %[2]q

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName, version))
}

func testAccClusterConfig_logging(rName string, logTypes []string) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(rName), fmt.Sprintf(`
resource "aws_eks_cluster" "test" {
  name                      = %[1]q
  role_arn                  = aws_iam_role.test.arn
  enabled_cluster_log_types = ["%[2]v"]

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName, strings.Join(logTypes, "\", \"")))
}

func testAccClusterConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(rName), fmt.Sprintf(`
resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  tags = {
    %[2]q = %[3]q
  }

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName, tagKey1, tagValue1))
}

func testAccClusterConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(rName), fmt.Sprintf(`
resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccClusterConfig_encryption(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(rName), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  encryption_config {
    resources = ["secrets"]

    provider {
      key_arn = aws_kms_key.test.arn
    }
  }

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName))
}

func testAccClusterConfig_encryptionVersion(rName, version string) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(rName), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn
  version  = %[2]q

  encryption_config {
    resources = ["secrets"]

    provider {
      key_arn = aws_kms_key.test.arn
    }
  }

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName, version))
}

func testAccClusterConfig_vpcSecurityGroupIDs(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(rName), fmt.Sprintf(`
resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  vpc_config {
    security_group_ids = [aws_security_group.test.id]
    subnet_ids         = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName))
}

func testAccClusterConfig_vpcSecurityGroupIDsAndSubnetIDsUpdated(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(rName), fmt.Sprintf(`
resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test2" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "new" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index + 2}.0/24"
  vpc_id            = aws_vpc.test.id

  ipv6_cidr_block                 = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, count.index + 2)
  assign_ipv6_address_on_creation = true

  tags = {
    Name                          = %[1]q
    "kubernetes.io/cluster/%[1]s" = "shared"
  }
}

resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  vpc_config {
    security_group_ids = [aws_security_group.test.id, aws_security_group.test2.id]
    subnet_ids         = aws_subnet.new[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName))
}

func testAccClusterConfig_vpcEndpointPrivateAccess(rName string, endpointPrivateAccess bool) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(rName), fmt.Sprintf(`
resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  vpc_config {
    endpoint_private_access = %[2]t
    endpoint_public_access  = true
    subnet_ids              = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName, endpointPrivateAccess))
}

func testAccClusterConfig_vpcEndpointPublicAccess(rName string, endpointPublicAccess bool) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(rName), fmt.Sprintf(`
resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  vpc_config {
    endpoint_private_access = true
    endpoint_public_access  = %[2]t
    subnet_ids              = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName, endpointPublicAccess))
}

func testAccClusterConfig_vpcPublicAccessCIDRs(rName string, publicAccessCidr string) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(rName), fmt.Sprintf(`
resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  vpc_config {
    endpoint_private_access = true
    endpoint_public_access  = true
    public_access_cidrs     = %[2]s
    subnet_ids              = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName, publicAccessCidr))
}

func testAccClusterConfig_networkServiceIPv4CIDR(rName string, serviceIpv4Cidr string) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(rName), fmt.Sprintf(`
resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  kubernetes_network_config {
    service_ipv4_cidr = %[2]s
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName, serviceIpv4Cidr))
}

func testAccClusterConfig_networkIPFamily(rName string, ipFamily string) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(rName), fmt.Sprintf(`
resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  kubernetes_network_config {
    ip_family = %[2]s
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName, ipFamily))
}

func testAccClusterConfig_outpost(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(rName), fmt.Sprintf(`
data "aws_iam_role" "test" {
  name = "AmazonEKSLocalOutpostClusterRole"
}

data "aws_outposts_outpost" "test" {
  id = "op-XXXXXXXX"
}

data "aws_subnets" test {
  filter {
    name   = "outpost-arn"
    values = [data.aws_outposts_outpost.test.arn]
  }
}

resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = data.aws_iam_role.test.arn

  outpost_config {
    control_plane_instance_type = "m5d.large"
    outpost_arns                = [data.aws_outposts_outpost.test.arn]
  }

  vpc_config {
    endpoint_private_access = true
    endpoint_public_access  = false
    subnet_ids              = [tolist(data.aws_subnets.test.ids)[0]]
  }
}
`, rName))
}

func testAccClusterConfig_outpostPlacement(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(rName), fmt.Sprintf(`
data "aws_iam_role" "test" {
  name = "AmazonEKSLocalOutpostClusterRole"
}

data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

data "aws_subnets" test {
  filter {
    name   = "outpost-arn"
    values = [data.aws_outposts_outpost.test.arn]
  }
}

resource "aws_placement_group" "test" {
  name     = %[1]q
  strategy = "cluster"
}

resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = data.aws_iam_role.test.arn

  outpost_config {
    control_plane_instance_type = "m5d.large"
    control_plane_placement {
      group_name = aws_placement_group.test.name
    }
    outpost_arns = [data.aws_outposts_outpost.test.arn]
  }

  vpc_config {
    endpoint_private_access = true
    endpoint_public_access  = false
    subnet_ids              = [tolist(data.aws_subnets.test.ids)[0]]
  }
}
`, rName))
}
