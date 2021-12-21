package eks_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/eks"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfeks "github.com/hashicorp/terraform-provider-aws/internal/service/eks"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccEKSClusterRegistration_basic(t *testing.T) {
	var cluster eks.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_cluster_registration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, eks.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClusterRegistrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterRegistrationBaseConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "connector_config.0.provider", "OTHER"),
					resource.TestCheckResourceAttrPair(resourceName, "connector_config.0.role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccEKSClusterRegistration_disappears(t *testing.T) {
	var cluster eks.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_cluster_registration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, eks.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClusterRegistrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterRegistrationBaseConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					acctest.CheckResourceDisappears(acctest.Provider, tfeks.ResourceClusterRegistration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEKSClusterRegistration_tags(t *testing.T) {
	var cluster1, cluster2, cluster3 eks.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_cluster_registration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, eks.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClusterRegistrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterRegistrationTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
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
				Config: testAccClusterRegistrationTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccClusterRegistrationTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccEKSClusterRegistration_InvalidClusterProvider(t *testing.T) {
	cluName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	regName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_eks_cluster_registration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, eks.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClusterRegistrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_Required(cluName),
			},
			{
				Config: testAccClusterRegistrationBaseConfig(regName),
			},
			{
				ResourceName:  resourceName,
				ImportState:   true,
				ImportStateId: cluName,
				ExpectError:   regexp.MustCompile(`has not been registered`),
			},
		},
	})
}

func testAccCheckClusterRegistrationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EKSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_eks_cluster_registration" {
			continue
		}

		_, err := tfeks.FindClusterByName(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EKS Cluster Registration %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccClusterRegistrationBaseIAMConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Sid" : "SSMAccess",
        "Effect" : "Allow",
        "Principal" : {
          "Service" : [
            "ssm.${data.aws_partition.current.dns_suffix}"
          ]
        },
        "Action" : "sts:AssumeRole"
      }
    ]
  })
}

data aws_iam_policy_document test {
  statement {
    sid = "SsmControlChannel"

    actions = [
      "ssmmessages:CreateControlChannel"
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:eks:*:*:cluster/*"
    ]
  }

  statement {
    sid = "ssmDataplaneOperations"

    actions = [
      "ssmmessages:CreateDataChannel",
      "ssmmessages:OpenDataChannel",
      "ssmmessages:OpenControlChannel"
    ]

    resources = [
      "*"
    ]
  }
}

resource "aws_iam_policy" "test" {
  name   = %[1]q
  path   = "/"
  policy = data.aws_iam_policy_document.test.json
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.test.arn
}
`, rName)
}

func testAccClusterRegistrationBaseConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterRegistrationBaseIAMConfig(rName),
		fmt.Sprintf(`
resource "aws_eks_cluster_registration" "test" {
  name = %[1]q

  connector_config {
    provider = "OTHER"
    role_arn = aws_iam_role.test.arn
  }

  depends_on = [
    "aws_iam_role_policy_attachment.test",
  ]
}
`, rName))
}

func testAccClusterRegistrationTags1Config(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccClusterRegistrationBaseIAMConfig(rName), fmt.Sprintf(`
resource "aws_eks_cluster_registration" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }

  connector_config {
    provider = "OTHER"
    role_arn = aws_iam_role.test.arn
  }

  depends_on = [
    "aws_iam_role_policy_attachment.test",
  ]
}
`, rName, tagKey1, tagValue1))
}

func testAccClusterRegistrationTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccClusterRegistrationBaseIAMConfig(rName), fmt.Sprintf(`
resource "aws_eks_cluster_registration" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  connector_config {
    provider = "OTHER"
    role_arn = aws_iam_role.test.arn
  }

  depends_on = [
    "aws_iam_role_policy_attachment.test",
  ]
}
  `, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
