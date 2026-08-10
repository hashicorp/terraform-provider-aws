// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkfirewall_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkfirewall/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfnetworkfirewall "github.com/hashicorp/terraform-provider-aws/internal/service/networkfirewall"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func checkContainerAssociationARN(name string) knownvalue.Check {
	return tfknownvalue.RegionalARNExact("network-firewall", "container-association/"+name)
}

func checkContainerAssociationARNAlternateRegion(name string) knownvalue.Check {
	return tfknownvalue.RegionalARNAlternateRegionExact("network-firewall", "container-association/"+name)
}

func TestAccNetworkFirewallContainerAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v networkfirewall.DescribeContainerAssociationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_container_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerAssociationConfig_basicEKS(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckContainerAssociationExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("container_association_arn"), checkContainerAssociationARN(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("container_association_name"), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(awstypes.ContainerMonitoringTypeEks)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("container_monitoring_configuration"), knownvalue.ListSizeExact(1)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "container_association_arn",
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "container_association_arn"),
				ImportStateVerifyIgnore:              []string{"update_token"},
			},
		},
	})
}

func TestAccNetworkFirewallContainerAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var v networkfirewall.DescribeContainerAssociationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_container_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerAssociationConfig_basicECS(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckContainerAssociationExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfnetworkfirewall.ResourceContainerAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccNetworkFirewallContainerAssociation_description(t *testing.T) {
	ctx := acctest.Context(t)

	var v networkfirewall.DescribeContainerAssociationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_container_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerAssociationConfig_description(rName, "initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckContainerAssociationExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("initial description")),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "container_association_arn",
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "container_association_arn"),
				ImportStateVerifyIgnore:              []string{"update_token"},
			},
			{
				Config: testAccContainerAssociationConfig_description(rName, "updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckContainerAssociationExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("updated description")),
				},
			},
		},
	})
}

func TestAccNetworkFirewallContainerAssociation_attributeFilters(t *testing.T) {
	ctx := acctest.Context(t)

	var v networkfirewall.DescribeContainerAssociationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_container_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerAssociationConfig_attributeFilters(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckContainerAssociationExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName,
						tfjsonpath.New("container_monitoring_configuration").AtSliceIndex(0).AtMapKey("attribute_filter").AtSliceIndex(0).AtMapKey(names.AttrKey),
						knownvalue.StringExact("ecs.instance-type"),
					),
					statecheck.ExpectKnownValue(resourceName,
						tfjsonpath.New("container_monitoring_configuration").AtSliceIndex(0).AtMapKey("attribute_filter").AtSliceIndex(0).AtMapKey(names.AttrValue),
						knownvalue.StringExact("c5.xlarge"),
					),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "container_association_arn",
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "container_association_arn"),
				ImportStateVerifyIgnore:              []string{"update_token"},
			},
		},
	})
}

func TestAccNetworkFirewallContainerAssociation_ecs(t *testing.T) {
	ctx := acctest.Context(t)

	var v networkfirewall.DescribeContainerAssociationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_container_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerAssociationConfig_basicECS(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckContainerAssociationExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(awstypes.ContainerMonitoringTypeEcs)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "container_association_arn",
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "container_association_arn"),
				ImportStateVerifyIgnore:              []string{"update_token"},
			},
		},
	})
}

func TestAccNetworkFirewallContainerAssociation_tags(t *testing.T) {
	ctx := acctest.Context(t)

	var v networkfirewall.DescribeContainerAssociationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_container_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerAssociationConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckContainerAssociationExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "container_association_arn",
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "container_association_arn"),
				ImportStateVerifyIgnore:              []string{"update_token"},
			},
			{
				Config: testAccContainerAssociationConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckContainerAssociationExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
			{
				Config: testAccContainerAssociationConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckContainerAssociationExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
		},
	})
}

func testAccCheckContainerAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).NetworkFirewallClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networkfirewall_container_association" {
				continue
			}

			_, err := tfnetworkfirewall.FindContainerAssociationByARN(ctx, conn, rs.Primary.Attributes["container_association_arn"])
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}
			return fmt.Errorf("NetworkFirewall Container Association %s still exists", rs.Primary.Attributes["container_association_arn"])
		}

		return nil
	}
}

func testAccCheckContainerAssociationExists(ctx context.Context, t *testing.T, n string, v *networkfirewall.DescribeContainerAssociationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).NetworkFirewallClient(ctx)

		out, err := tfnetworkfirewall.FindContainerAssociationByARN(ctx, conn, rs.Primary.Attributes["container_association_arn"])
		if err != nil {
			return err
		}

		*v = *out

		return nil
	}
}

// testAccContainerAssociationConfig_baseEKS provisions the minimal VPC, IAM role, and EKS
// cluster required as a monitoring target. EKS cluster creation is slow, so this base is
// reserved for the one test that must exercise the EKS code path.
func testAccContainerAssociationConfig_baseEKS(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_service_principal" "eks" {
  service_name = "eks"
}

resource "aws_iam_role" "cluster" {
  name = %[1]q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "${data.aws_service_principal.eks.name}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "cluster_AmazonEKSClusterPolicy" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.cluster.name
}

resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.cluster.arn

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.cluster_AmazonEKSClusterPolicy]
}
`, rName))
}

// testAccContainerAssociationConfig_baseECS provisions a bare ECS cluster, which has no VPC
// or IAM dependencies. This keeps most argument-level tests fast.
func testAccContainerAssociationConfig_baseECS(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}
`, rName)
}

func testAccContainerAssociationConfig_basicEKS(rName string) string {
	return acctest.ConfigCompose(testAccContainerAssociationConfig_baseEKS(rName), fmt.Sprintf(`
resource "aws_networkfirewall_container_association" "test" {
  container_association_name = %[1]q
  type                       = "EKS"

  container_monitoring_configuration {
    cluster_arn = aws_eks_cluster.test.arn
  }
}
`, rName))
}

func testAccContainerAssociationConfig_basicECS(rName string) string {
	return acctest.ConfigCompose(testAccContainerAssociationConfig_baseECS(rName), fmt.Sprintf(`
resource "aws_networkfirewall_container_association" "test" {
  container_association_name = %[1]q
  type                       = "ECS"

  container_monitoring_configuration {
    cluster_arn = aws_ecs_cluster.test.arn
  }
}
`, rName))
}

func testAccContainerAssociationConfig_description(rName, description string) string {
	return acctest.ConfigCompose(testAccContainerAssociationConfig_baseECS(rName), fmt.Sprintf(`
resource "aws_networkfirewall_container_association" "test" {
  container_association_name = %[1]q
  type                       = "ECS"
  description                = %[2]q

  container_monitoring_configuration {
    cluster_arn = aws_ecs_cluster.test.arn
  }
}
`, rName, description))
}

func testAccContainerAssociationConfig_attributeFilters(rName string) string {
	return acctest.ConfigCompose(testAccContainerAssociationConfig_baseECS(rName), fmt.Sprintf(`
resource "aws_networkfirewall_container_association" "test" {
  container_association_name = %[1]q
  type                       = "ECS"

  container_monitoring_configuration {
    cluster_arn = aws_ecs_cluster.test.arn

    attribute_filter {
      key   = "ecs.instance-type"
      value = "c5.xlarge"
    }
  }
}
`, rName))
}

func testAccContainerAssociationConfig_tags1(rName, tag1Key, tag1Value string) string {
	return acctest.ConfigCompose(testAccContainerAssociationConfig_baseECS(rName), fmt.Sprintf(`
resource "aws_networkfirewall_container_association" "test" {
  container_association_name = %[1]q
  type                       = "ECS"

  container_monitoring_configuration {
    cluster_arn = aws_ecs_cluster.test.arn
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tag1Key, tag1Value))
}

func testAccContainerAssociationConfig_tags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return acctest.ConfigCompose(testAccContainerAssociationConfig_baseECS(rName), fmt.Sprintf(`
resource "aws_networkfirewall_container_association" "test" {
  container_association_name = %[1]q
  type                       = "ECS"

  container_monitoring_configuration {
    cluster_arn = aws_ecs_cluster.test.arn
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value))
}
