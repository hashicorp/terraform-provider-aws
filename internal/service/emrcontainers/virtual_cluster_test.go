package emrcontainers_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/emrcontainers"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfemrcontainers "github.com/hashicorp/terraform-provider-aws/internal/service/emrcontainers"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccEMRContainersVirtualCluster_basic(t *testing.T) {
	var v emrcontainers.VirtualCluster
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_emrcontainers_virtual_cluster.test"
	testExternalProviders := map[string]resource.ExternalProvider{
		"kubernetes": {
			Source:            "hashicorp/kubernetes",
			VersionConstraint: "~> 2.3",
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckIAMServiceLinkedRole(t, "/aws-service-role/emr-containers.amazonaws.com")
		},
		ErrorCheck:        acctest.ErrorCheck(t, emrcontainers.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		ExternalProviders: testExternalProviders,
		CheckDestroy:      testAccCheckVirtualClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVirtualClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "container_provider.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "container_provider.0.id", rName),
					resource.TestCheckResourceAttr(resourceName, "container_provider.0.info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "container_provider.0.info.0.eks_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "container_provider.0.info.0.eks_info.0.namespace", "default"),
					resource.TestCheckResourceAttr(resourceName, "container_provider.0.type", "EKS"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			//
			// virtual_cluster_test.go:29: Step 2/2 error running import: exit status 1
			//
			//   Error: Invalid provider configuration
			//
			//     on /var/folders/lx/48ng4y950gv10_x6x1jwk05w0000gq/T/plugintest806912737/work384085989/terraform_plugin_test.tf line 180:
			//    180: provider "kubernetes" {
			//
			//   The configuration for provider["registry.terraform.io/hashicorp/kubernetes"]
			//   depends on values that cannot be determined until apply.
			//
			// {
			// 	ResourceName:      resourceName,
			// 	ImportState:       true,
			// 	ImportStateVerify: true,
			// },
		},
	})
}

func TestAccEMRContainersVirtualCluster_disappears(t *testing.T) {
	var v emrcontainers.VirtualCluster
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_emrcontainers_virtual_cluster.test"
	testExternalProviders := map[string]resource.ExternalProvider{
		"kubernetes": {
			Source:            "hashicorp/kubernetes",
			VersionConstraint: "~> 2.3",
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckIAMServiceLinkedRole(t, "/aws-service-role/emr-containers.amazonaws.com")
		},
		ErrorCheck:        acctest.ErrorCheck(t, emrcontainers.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		ExternalProviders: testExternalProviders,
		CheckDestroy:      testAccCheckVirtualClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVirtualClusterExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfemrcontainers.ResourceVirtualCluster(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEMRContainersVirtualCluster_tags(t *testing.T) {
	var v emrcontainers.VirtualCluster
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_emrcontainers_virtual_cluster.test"
	testExternalProviders := map[string]resource.ExternalProvider{
		"kubernetes": {
			Source:            "hashicorp/kubernetes",
			VersionConstraint: "~> 2.3",
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckIAMServiceLinkedRole(t, "/aws-service-role/emr-containers.amazonaws.com")
		},
		ErrorCheck:        acctest.ErrorCheck(t, emrcontainers.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		ExternalProviders: testExternalProviders,
		CheckDestroy:      testAccCheckVirtualClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualClusterConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVirtualClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccVirtualClusterConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVirtualClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccVirtualClusterConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVirtualClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckVirtualClusterExists(n string, v *emrcontainers.VirtualCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EMR Containers Virtual Cluster ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EMRContainersConn

		output, err := tfemrcontainers.FindVirtualClusterByID(context.TODO(), conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckVirtualClusterDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EMRContainersConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_emrcontainers_virtual_cluster" {
			continue
		}

		_, err := tfemrcontainers.FindVirtualClusterByID(context.TODO(), conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EMR Containers Virtual Cluster %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccVirtualClusterBase(rName string) string {
	//lintignore:AT004
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_caller_identity" "current" {}

resource "aws_iam_role" "cluster" {
  name = "%[1]s-cluster"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = [
          "eks.${data.aws_partition.current.dns_suffix}",
          "eks-nodegroup.${data.aws_partition.current.dns_suffix}",
        ]
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role_policy_attachment" "cluster-AmazonEKSClusterPolicy" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.cluster.name
}

resource "aws_iam_role" "node" {
  name = "%[1]s-node"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role_policy_attachment" "node-AmazonEKSWorkerNodePolicy" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonEKSWorkerNodePolicy"
  role       = aws_iam_role.node.name
}

resource "aws_iam_role_policy_attachment" "node-AmazonEKS_CNI_Policy" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonEKS_CNI_Policy"
  role       = aws_iam_role.node.name
}

resource "aws_iam_role_policy_attachment" "node-AmazonEC2ContainerRegistryReadOnly" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
  role       = aws_iam_role.node.name
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

resource "aws_route_table" "test" {
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
  route_table_id = aws_route_table.test.id
  vpc_id         = aws_vpc.test.id
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    protocol    = -1
    to_port     = 0
  }

  ingress {
    cidr_blocks = [aws_vpc.test.cidr_block]
    from_port   = 0
    protocol    = -1
    to_port     = 0
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone       = data.aws_availability_zones.available.names[count.index]
  cidr_block              = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  map_public_ip_on_launch = true
  vpc_id                  = aws_vpc.test.id

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

  depends_on = [
    "aws_iam_role_policy_attachment.cluster-AmazonEKSClusterPolicy",
    "aws_main_route_table_association.test",
  ]
}

resource "aws_eks_node_group" "test" {
  cluster_name    = aws_eks_cluster.test.name
  node_group_name = %[1]q
  node_role_arn   = aws_iam_role.node.arn
  subnet_ids      = aws_subnet.test[*].id

  scaling_config {
    desired_size = 1
    max_size     = 1
    min_size     = 1
  }

  depends_on = [
    kubernetes_config_map.aws_auth,
    aws_iam_role_policy_attachment.node-AmazonEKSWorkerNodePolicy,
    aws_iam_role_policy_attachment.node-AmazonEKS_CNI_Policy,
    aws_iam_role_policy_attachment.node-AmazonEC2ContainerRegistryReadOnly,
  ]
}

data "aws_eks_cluster_auth" "cluster" {
  name = aws_eks_cluster.test.id
}

provider "kubernetes" {
  host                   = aws_eks_cluster.test.endpoint
  cluster_ca_certificate = base64decode(aws_eks_cluster.test.certificate_authority[0].data)
  token                  = data.aws_eks_cluster_auth.cluster.token
}

resource "kubernetes_role" "emrcontainers_role" {
  metadata {
    name      = "emr-containers"
    namespace = "default"
  }

  rule {
    api_groups = [""]
    resources  = ["namespaces"]
    verbs      = ["get"]
  }

  rule {
    api_groups = [""]
    resources  = ["serviceaccounts", "services", "configmaps", "events", "pods", "pods/log"]
    verbs      = ["get", "list", "watch", "describe", "create", "edit", "delete", "deletecollection", "annotate", "patch", "label"]
  }

  rule {
    api_groups = [""]
    resources  = ["secrets"]
    verbs      = ["create", "patch", "delete", "watch"]
  }

  rule {
    api_groups = ["apps"]
    resources  = ["statefulsets", "deployments"]
    verbs      = ["get", "list", "watch", "describe", "create", "edit", "delete", "annotate", "patch", "label"]
  }

  rule {
    api_groups = ["batch"]
    resources  = ["jobs"]
    verbs      = ["get", "list", "watch", "describe", "create", "edit", "delete", "annotate", "patch", "label"]
  }

  rule {
    api_groups = ["extensions"]
    resources  = ["ingresses"]
    verbs      = ["get", "list", "watch", "describe", "create", "edit", "delete", "annotate", "patch", "label"]
  }

  rule {
    api_groups = ["rbac.authorization.k8s.io"]
    resources  = ["roles", "rolebindings"]
    verbs      = ["get", "list", "watch", "describe", "create", "edit", "delete", "deletecollection", "annotate", "patch", "label"]
  }
}

resource "kubernetes_role_binding" "emrcontainers_rolemapping" {
  metadata {
    name      = "emr-containers"
    namespace = "default"
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "Role"
    name      = kubernetes_role.emrcontainers_role.metadata[0].name
  }

  subject {
    kind      = "User"
    name      = "emr-containers"
    api_group = "rbac.authorization.k8s.io"
  }
}

resource "kubernetes_config_map" "aws_auth" {
  metadata {
    name      = "aws-auth"
    namespace = "kube-system"
  }

  data = {
    mapRoles = <<EOF
- rolearn: ${aws_iam_role.node.arn}
  username: system:node:{{EC2PrivateDNSName}}
  groups:
  - system:bootstrappers
  - system:nodes
- rolearn: arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role/AWSServiceRoleForAmazonEMRContainers
  username: ${kubernetes_role_binding.emrcontainers_rolemapping.subject[0].name}
EOF
  }
}
`, rName))
}

func testAccVirtualClusterConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccVirtualClusterBase(rName), fmt.Sprintf(`
resource "aws_emrcontainers_virtual_cluster" "test" {
  container_provider {
    id   = aws_eks_cluster.test.name
    type = "EKS"

    info {
      eks_info {
        namespace = "default"
      }
    }
  }

  name = %[1]q

  depends_on = [kubernetes_config_map.aws_auth]
}
`, rName))
}

func testAccVirtualClusterConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccVirtualClusterBase(rName), fmt.Sprintf(`
resource "aws_emrcontainers_virtual_cluster" "test" {
  container_provider {
    id   = aws_eks_cluster.test.name
    type = "EKS"

    info {
      eks_info {
        namespace = "default"
      }
    }
  }

  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }

  depends_on = [kubernetes_config_map.aws_auth]
}
`, rName, tagKey1, tagValue1))
}

func testAccVirtualClusterConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccVirtualClusterBase(rName), fmt.Sprintf(`
resource "aws_emrcontainers_virtual_cluster" "test" {
  container_provider {
    id   = aws_eks_cluster.test.name
    type = "EKS"

    info {
      eks_info {
        namespace = "default"
      }
    }
  }

  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  depends_on = [kubernetes_config_map.aws_auth]
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
