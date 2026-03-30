// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package eks_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfeks "github.com/hashicorp/terraform-provider-aws/internal/service/eks"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEKSCapability_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var capability awstypes.Capability
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_eks_capability.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapabilityDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCapabilityConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapabilityExists(ctx, t, resourceName, &capability),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("eks", regexache.MustCompile(`capability/.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("capability_name"), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("delete_propagation_policy"), tfknownvalue.StringExact(awstypes.CapabilityDeletePropagationPolicyRetain)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(awstypes.CapabilityTypeKro)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ",", names.AttrClusterName, "capability_name"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccEKSCapability_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var capability awstypes.Capability
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_eks_capability.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapabilityDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCapabilityConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapabilityExists(ctx, t, resourceName, &capability),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfeks.ResourceCapability, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccEKSCapability_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var capability awstypes.Capability
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_eks_capability.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapabilityDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCapabilityConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapabilityExists(ctx, t, resourceName, &capability),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ",", names.AttrClusterName, "capability_name"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccCapabilityConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapabilityExists(ctx, t, resourceName, &capability),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
			{
				Config: testAccCapabilityConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapabilityExists(ctx, t, resourceName, &capability),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
		},
	})
}

func TestAccEKSCapability_ArgoCD_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var capability awstypes.Capability
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_eks_capability.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapabilityDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCapabilityConfig_argoCDBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapabilityExists(ctx, t, resourceName, &capability),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ",", names.AttrClusterName, "capability_name"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccEKSCapability_ArgoCD_rbac(t *testing.T) {
	ctx := acctest.Context(t)
	var capability awstypes.Capability
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_eks_capability.test"
	userID := acctest.SkipIfEnvVarNotSet(t, "AWS_IDENTITY_STORE_USER_ID")
	groupID := acctest.SkipIfEnvVarNotSet(t, "AWS_IDENTITY_STORE_GROUP_ID")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapabilityDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCapabilityConfig_argoCDRBAC1(rName, userID, groupID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapabilityExists(ctx, t, resourceName, &capability),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ",", names.AttrClusterName, "capability_name"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccCapabilityConfig_argoCDRBAC2(rName, userID, groupID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapabilityExists(ctx, t, resourceName, &capability),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func testAccCheckCapabilityExists(ctx context.Context, t *testing.T, n string, v *awstypes.Capability) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).EKSClient(ctx)

		output, err := tfeks.FindCapabilityByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrClusterName], rs.Primary.Attributes["capability_name"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckCapabilityDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EKSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_eks_capability" {
				continue
			}

			_, err := tfeks.FindCapabilityByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrClusterName], rs.Primary.Attributes["capability_name"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EKS Capability (%s,%s) still exists", rs.Primary.Attributes[names.AttrClusterName], rs.Primary.Attributes["capability_name"])
		}

		return nil
	}
}

func testAccCapabilityConfig_base(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(rName), fmt.Sprintf(`
resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.cluster.arn

  access_config {
    authentication_mode                         = "API"
    bootstrap_cluster_creator_admin_permissions = true
  }

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.cluster_AmazonEKSClusterPolicy]
}

resource "aws_iam_role" "capability" {
  name = "%[1]s-capability"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        Service = "capabilities.eks.amazonaws.com"
      }
      Action = [
        "sts:AssumeRole",
        "sts:TagSession"
      ]
    }]
  })
}

resource "aws_iam_role_policy_attachment" "capability" {
  role       = aws_iam_role.capability.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AdministratorAccess"
}
`, rName))
}

func testAccCapabilityConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccCapabilityConfig_base(rName), fmt.Sprintf(`
resource "aws_eks_capability" "test" {
  cluster_name              = aws_eks_cluster.test.name
  capability_name           = %[1]q
  type                      = "KRO"
  role_arn                  = aws_iam_role.capability.arn
  delete_propagation_policy = "RETAIN"

  depends_on = [aws_iam_role_policy_attachment.capability]
}
`, rName))
}

func testAccCapabilityConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccCapabilityConfig_base(rName), fmt.Sprintf(`
resource "aws_eks_capability" "test" {
  cluster_name              = aws_eks_cluster.test.name
  capability_name           = %[1]q
  type                      = "KRO"
  role_arn                  = aws_iam_role.capability.arn
  delete_propagation_policy = "RETAIN"

  tags = {
    %[2]q = %[3]q
  }

  depends_on = [aws_iam_role_policy_attachment.capability]
}
`, rName, tagKey1, tagValue1))
}

func testAccCapabilityConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccCapabilityConfig_base(rName), fmt.Sprintf(`
resource "aws_eks_capability" "test" {
  cluster_name              = aws_eks_cluster.test.name
  capability_name           = %[1]q
  type                      = "KRO"
  role_arn                  = aws_iam_role.capability.arn
  delete_propagation_policy = "RETAIN"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  depends_on = [aws_iam_role_policy_attachment.capability]
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccCapabilityConfig_argoCDBasic(rName string) string {
	return acctest.ConfigCompose(testAccCapabilityConfig_base(rName), fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_eks_capability" "test" {
  cluster_name              = aws_eks_cluster.test.name
  capability_name           = %[1]q
  type                      = "ARGOCD"
  role_arn                  = aws_iam_role.capability.arn
  delete_propagation_policy = "RETAIN"

  configuration {
    argo_cd {
      aws_idc {
        idc_instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
      }
    }
  }

  depends_on = [aws_iam_role_policy_attachment.capability]
}
`, rName))
}

func testAccCapabilityConfig_argoCDRBAC1(rName, userID, groupName string) string {
	return acctest.ConfigCompose(testAccCapabilityConfig_base(rName), fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_eks_capability" "test" {
  cluster_name              = aws_eks_cluster.test.name
  capability_name           = %[1]q
  type                      = "ARGOCD"
  role_arn                  = aws_iam_role.capability.arn
  delete_propagation_policy = "RETAIN"

  configuration {
    argo_cd {
      aws_idc {
        idc_instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
        idc_region       = %[2]q
      }

      rbac_role_mapping {
        role = "ADMIN"

        identity {
          type = "SSO_USER"
          id   = %[3]q
        }
      }

      rbac_role_mapping {
        role = "VIEWER"

        identity {
          type = "SSO_GROUP"
          id   = %[4]q
        }
      }
    }
  }

  depends_on = [aws_iam_role_policy_attachment.capability]
}
`, rName, acctest.Region(), userID, groupName))
}

func testAccCapabilityConfig_argoCDRBAC2(rName, userID, groupName string) string {
	return acctest.ConfigCompose(testAccCapabilityConfig_base(rName), fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_eks_capability" "test" {
  cluster_name              = aws_eks_cluster.test.name
  capability_name           = %[1]q
  type                      = "ARGOCD"
  role_arn                  = aws_iam_role.capability.arn
  delete_propagation_policy = "RETAIN"

  configuration {
    argo_cd {
      aws_idc {
        idc_instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
        idc_region       = %[2]q
      }

      rbac_role_mapping {
        role = "EDITOR"

        identity {
          type = "SSO_USER"
          id   = %[3]q
        }
      }

      rbac_role_mapping {
        role = "ADMIN"

        identity {
          type = "SSO_GROUP"
          id   = %[4]q
        }
      }
    }
  }

  depends_on = [aws_iam_role_policy_attachment.capability]
}
`, rName, acctest.Region(), userID, groupName))
}
