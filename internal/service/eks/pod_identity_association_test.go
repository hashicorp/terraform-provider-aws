// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfeks "github.com/hashicorp/terraform-provider-aws/internal/service/eks"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEKSPodIdentityAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var podidentityassociation types.PodIdentityAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_pod_identity_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EKSEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPodIdentityAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPodIdentityAssociationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPodIdentityAssociationExists(ctx, resourceName, &podidentityassociation),
					resource.TestCheckResourceAttrSet(resourceName, "cluster_name"),
					resource.TestCheckResourceAttrSet(resourceName, "namespace"),
					resource.TestCheckResourceAttrSet(resourceName, "role_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "service_account"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccCheckPodIdentityAssociationImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEKSPodIdentityAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var podidentityassociation types.PodIdentityAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_pod_identity_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EKSEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPodIdentityAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPodIdentityAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPodIdentityAssociationExists(ctx, resourceName, &podidentityassociation),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfeks.ResourcePodIdentityAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEKSPodIdentityAssociation_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var podidentityassociation types.PodIdentityAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_pod_identity_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EKSEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPodIdentityAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPodIdentityAssociationConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPodIdentityAssociationExists(ctx, resourceName, &podidentityassociation),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccCheckPodIdentityAssociationImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPodIdentityAssociationConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPodIdentityAssociationExists(ctx, resourceName, &podidentityassociation),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccPodIdentityAssociationConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPodIdentityAssociationExists(ctx, resourceName, &podidentityassociation),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccEKSPodIdentityAssociation_updateRoleARN(t *testing.T) {
	ctx := acctest.Context(t)
	var podidentityassociation types.PodIdentityAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_pod_identity_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EKSEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPodIdentityAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPodIdentityAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPodIdentityAssociationExists(ctx, resourceName, &podidentityassociation),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccCheckPodIdentityAssociationImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPodIdentityAssociationConfig_updatedRoleARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPodIdentityAssociationExists(ctx, resourceName, &podidentityassociation),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test2", "arn"),
				),
			},
		},
	})
}

func testAccCheckPodIdentityAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EKSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_eks_pod_identity_association" {
				continue
			}

			_, err := tfeks.FindPodIdentityAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["association_id"], rs.Primary.Attributes["cluster_name"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EKS Pod Identity Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckPodIdentityAssociationExists(ctx context.Context, n string, v *types.PodIdentityAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return create.Error(names.EKS, create.ErrActionCheckingExistence, tfeks.ResNamePodIdentityAssociation, n, errors.New("not found"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EKSClient(ctx)

		output, err := tfeks.FindPodIdentityAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["association_id"], rs.Primary.Attributes["cluster_name"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckPodIdentityAssociationImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		return fmt.Sprintf("%s,%s", rs.Primary.Attributes["cluster_name"], rs.Primary.Attributes["association_id"]), nil
	}
}

func testAccPodIdentityAssociationConfig_clusterBase(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "cluster" {
  name_prefix = %[1]q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "eks.amazonaws.com"
      },
      "Action": [
        "sts:AssumeRole",
        "sts:TagSession"
      ]
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "cluster" {
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
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name                          = %[1]q
    "kubernetes.io/cluster/%[1]s" = "shared"
  }
}

resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.cluster.arn

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.cluster]
}
`, rName))
}

func testAccPodIdentityAssociationConfig_podIdentityRoleBase(rName string) string {
	return acctest.ConfigCompose(
		fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "pods.eks.amazonaws.com"
      },
      "Action": [
        "sts:AssumeRole",
        "sts:TagSession"
      ]
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonS3ReadOnlyAccess"
  role       = aws_iam_role.test.name
}
`, rName))
}

func testAccPodIdentityAssociationConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccPodIdentityAssociationConfig_clusterBase(rName),
		testAccPodIdentityAssociationConfig_podIdentityRoleBase(rName),
		fmt.Sprintf(`
resource "aws_eks_pod_identity_association" "test" {
  cluster_name    = aws_eks_cluster.test.name
  namespace       = %[1]q
  service_account = "%[1]s-sa"
  role_arn        = aws_iam_role.test.arn
}
`, rName))
}

func testAccPodIdentityAssociationConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccPodIdentityAssociationConfig_clusterBase(rName),
		testAccPodIdentityAssociationConfig_podIdentityRoleBase(rName),
		fmt.Sprintf(`
resource "aws_eks_pod_identity_association" "test" {
  cluster_name    = aws_eks_cluster.test.name
  namespace       = %[1]q
  service_account = "%[1]s-sa"
  role_arn        = aws_iam_role.test.arn

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccPodIdentityAssociationConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccPodIdentityAssociationConfig_clusterBase(rName),
		testAccPodIdentityAssociationConfig_podIdentityRoleBase(rName),
		fmt.Sprintf(`
resource "aws_eks_pod_identity_association" "test" {
  cluster_name    = aws_eks_cluster.test.name
  namespace       = %[1]q
  service_account = "%[1]s-sa"
  role_arn        = aws_iam_role.test.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccPodIdentityAssociationConfig_updatedRoleARN(rName string) string {
	return acctest.ConfigCompose(
		testAccPodIdentityAssociationConfig_clusterBase(rName),
		testAccPodIdentityAssociationConfig_podIdentityRoleBase(rName),
		fmt.Sprintf(`
resource "aws_iam_role" "test2" {
  name = "%[1]s-2"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "pods.eks.amazonaws.com"
      },
      "Action": [
        "sts:AssumeRole",
        "sts:TagSession"
      ]
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "test2" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonS3ReadOnlyAccess"
  role       = aws_iam_role.test2.name
}

resource "aws_eks_pod_identity_association" "test" {
  cluster_name    = aws_eks_cluster.test.name
  namespace       = %[1]q
  service_account = "%[1]s-sa"
  role_arn        = aws_iam_role.test2.arn
}
`, rName))
}
