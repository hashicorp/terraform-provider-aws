// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
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

func TestAccEKSFargateProfile_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var fargateProfile types.FargateProfile
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	eksClusterResourceName := "aws_eks_cluster.test"
	iamRoleResourceName := "aws_iam_role.pod"
	resourceName := "aws_eks_fargate_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartition(t, names.StandardPartitionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFargateProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFargateProfileConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFargateProfileExists(ctx, resourceName, &fargateProfile),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "eks", regexache.MustCompile(fmt.Sprintf("fargateprofile/%[1]s/%[1]s/.+", rName))),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrClusterName, eksClusterResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "fargate_profile_name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "pod_execution_role_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "selector.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.FargateProfileStatusActive)),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", acctest.Ct2),
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

func TestAccEKSFargateProfile_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var fargateProfile types.FargateProfile
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_fargate_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartition(t, names.StandardPartitionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFargateProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFargateProfileConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFargateProfileExists(ctx, resourceName, &fargateProfile),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfeks.ResourceFargateProfile(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEKSFargateProfile_Multi_profile(t *testing.T) {
	ctx := acctest.Context(t)
	var fargateProfile types.FargateProfile
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName1 := "aws_eks_fargate_profile.test.0"
	resourceName2 := "aws_eks_fargate_profile.test.1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartition(t, names.StandardPartitionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFargateProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFargateProfileConfig_multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFargateProfileExists(ctx, resourceName1, &fargateProfile),
					testAccCheckFargateProfileExists(ctx, resourceName2, &fargateProfile),
				),
			},
		},
	})
}

func TestAccEKSFargateProfile_Selector_labels(t *testing.T) {
	ctx := acctest.Context(t)
	var fargateProfile1 types.FargateProfile
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_fargate_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartition(t, names.StandardPartitionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFargateProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFargateProfileConfig_selectorLabels1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFargateProfileExists(ctx, resourceName, &fargateProfile1),
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

func TestAccEKSFargateProfile_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var fargateProfile1, fargateProfile2, fargateProfile3 types.FargateProfile
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_fargate_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartition(t, names.StandardPartitionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFargateProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFargateProfileConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFargateProfileExists(ctx, resourceName, &fargateProfile1),
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
				Config: testAccFargateProfileConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFargateProfileExists(ctx, resourceName, &fargateProfile2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccFargateProfileConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFargateProfileExists(ctx, resourceName, &fargateProfile3),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckFargateProfileExists(ctx context.Context, n string, v *types.FargateProfile) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EKS Fargate Profile ID is set")
		}

		clusterName, fargateProfileName, err := tfeks.FargateProfileParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EKSClient(ctx)

		output, err := tfeks.FindFargateProfileByTwoPartKey(ctx, conn, clusterName, fargateProfileName)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckFargateProfileDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EKSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_eks_fargate_profile" {
				continue
			}

			clusterName, fargateProfileName, err := tfeks.FargateProfileParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfeks.FindFargateProfileByTwoPartKey(ctx, conn, clusterName, fargateProfileName)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EKS Fargate Profile %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccFargateProfileConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_partition" "current" {}

resource "aws_iam_role" "cluster" {
  name = "%[1]s-cluster"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "eks.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role_policy_attachment" "cluster-AmazonEKSClusterPolicy" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.cluster.name
}

resource "aws_iam_role" "pod" {
  name = "%[1]s-pod"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "eks-fargate-pods.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role_policy_attachment" "pod-AmazonEKSFargatePodExecutionRolePolicy" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonEKSFargatePodExecutionRolePolicy"
  role       = aws_iam_role.pod.name
}

resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name                          = %[1]q
    "kubernetes.io/cluster/%[1]s" = "shared"
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_main_route_table_association" "test" {
  route_table_id = aws_route_table.public.id
  vpc_id         = aws_vpc.test.id
}

resource "aws_subnet" "private" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index + 2)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name                          = %[1]q
    "kubernetes.io/cluster/%[1]s" = "shared"
  }
}

resource "aws_eip" "private" {
  count      = 2
  depends_on = [aws_internet_gateway.test]

  domain = "vpc"

  tags = {
    Name = %[1]q
  }
}

resource "aws_nat_gateway" "private" {
  count = 2

  allocation_id = aws_eip.private[count.index].id
  subnet_id     = aws_subnet.private[count.index].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "private" {
  count = 2

  vpc_id = aws_vpc.test.id

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.private[count.index].id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "private" {
  count = 2

  subnet_id      = aws_subnet.private[count.index].id
  route_table_id = aws_route_table.private[count.index].id
}

resource "aws_subnet" "public" {
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
    subnet_ids = aws_subnet.public[*].id
  }

  depends_on = [
    aws_iam_role_policy_attachment.cluster-AmazonEKSClusterPolicy,
    aws_main_route_table_association.test,
  ]
}
`, rName)
}

func testAccFargateProfileConfig_name(rName string) string {
	return acctest.ConfigCompose(testAccFargateProfileConfig_base(rName), fmt.Sprintf(`
resource "aws_eks_fargate_profile" "test" {
  cluster_name           = aws_eks_cluster.test.name
  fargate_profile_name   = %[1]q
  pod_execution_role_arn = aws_iam_role.pod.arn
  subnet_ids             = aws_subnet.private[*].id

  selector {
    namespace = "test"
  }

  depends_on = [
    aws_iam_role_policy_attachment.pod-AmazonEKSFargatePodExecutionRolePolicy,
    aws_route_table_association.private,
  ]
}
`, rName))
}

func testAccFargateProfileConfig_multiple(rName string) string {
	return acctest.ConfigCompose(testAccFargateProfileConfig_base(rName), fmt.Sprintf(`
resource "aws_eks_fargate_profile" "test" {
  count = 2

  cluster_name           = aws_eks_cluster.test.name
  fargate_profile_name   = "%[1]s-${count.index}"
  pod_execution_role_arn = aws_iam_role.pod.arn
  subnet_ids             = aws_subnet.private[*].id

  selector {
    namespace = "test"
  }

  depends_on = [
    aws_iam_role_policy_attachment.pod-AmazonEKSFargatePodExecutionRolePolicy,
    aws_route_table_association.private,
  ]
}
`, rName))
}

func testAccFargateProfileConfig_selectorLabels1(rName, labelKey1, labelValue1 string) string {
	return acctest.ConfigCompose(testAccFargateProfileConfig_base(rName), fmt.Sprintf(`
resource "aws_eks_fargate_profile" "test" {
  cluster_name           = aws_eks_cluster.test.name
  fargate_profile_name   = %[1]q
  pod_execution_role_arn = aws_iam_role.pod.arn
  subnet_ids             = aws_subnet.private[*].id

  selector {
    labels = {
      %[2]q = %[3]q
    }
    namespace = "test"
  }

  depends_on = [
    aws_iam_role_policy_attachment.pod-AmazonEKSFargatePodExecutionRolePolicy,
    aws_route_table_association.private,
  ]
}
`, rName, labelKey1, labelValue1))
}

func testAccFargateProfileConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccFargateProfileConfig_base(rName), fmt.Sprintf(`
resource "aws_eks_fargate_profile" "test" {
  cluster_name           = aws_eks_cluster.test.name
  fargate_profile_name   = %[1]q
  pod_execution_role_arn = aws_iam_role.pod.arn
  subnet_ids             = aws_subnet.private[*].id

  selector {
    namespace = "test"
  }

  tags = {
    %[2]q = %[3]q
  }

  depends_on = [
    aws_iam_role_policy_attachment.pod-AmazonEKSFargatePodExecutionRolePolicy,
    aws_route_table_association.private,
  ]
}
`, rName, tagKey1, tagValue1))
}

func testAccFargateProfileConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccFargateProfileConfig_base(rName), fmt.Sprintf(`
resource "aws_eks_fargate_profile" "test" {
  cluster_name           = aws_eks_cluster.test.name
  fargate_profile_name   = %[1]q
  pod_execution_role_arn = aws_iam_role.pod.arn
  subnet_ids             = aws_subnet.private[*].id

  selector {
    namespace = "test"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  depends_on = [
    aws_iam_role_policy_attachment.pod-AmazonEKSFargatePodExecutionRolePolicy,
    aws_route_table_association.private,
  ]
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
