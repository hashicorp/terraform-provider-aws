// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptunegraph_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/neptunegraph"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfneptunegraph "github.com/hashicorp/terraform-provider-aws/internal/service/neptunegraph"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNeptuneGraphGraph_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var graph neptunegraph.GetGraphOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptunegraph_graph.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneGraphServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGraphExists(ctx, resourceName, &graph),
					resource.TestCheckResourceAttr(resourceName, "graph_name", rName),
					resource.TestCheckResourceAttr(resourceName, "provisioned_memory", "16"),
					resource.TestCheckResourceAttr(resourceName, "public_connectivity", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "replica_count", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, acctest.CtFalse),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "neptune-graph", regexache.MustCompile(`graph/.+$`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"graph_name_prefix"},
			},
		},
	})
}

func TestAccNeptuneGraphGraph_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var graph neptunegraph.GetGraphOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptunegraph_graph.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AttrID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGraphExists(ctx, resourceName, &graph),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfneptunegraph.ResourceGraph, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccNeptuneGraphGraph_vectorSearch(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var graph neptunegraph.GetGraphOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptunegraph_graph.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AttrID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphConfig_vectorSearch(rName, 128),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGraphExists(ctx, resourceName, &graph),
					resource.TestCheckResourceAttr(resourceName, "vector_search_configuration.0.vector_search_dimension", "128"),
				),
			},
		},
	})
}

func TestAccNeptuneGraphGraph_kmsKey(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var graph neptunegraph.GetGraphOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptunegraph_graph.test"
	keyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AttrID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphConfig_kmsKey(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGraphExists(ctx, resourceName, &graph),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_identifier", keyResourceName, names.AttrARN),
				),
			},
		},
	})
}

func TestAccNeptuneGraphGraph_deletionProtection(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var graph neptunegraph.GetGraphOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptunegraph_graph.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AttrID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphConfig_deletionProtection(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGraphExists(ctx, resourceName, &graph),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, acctest.CtTrue),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccGraphConfig_deletionProtection(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGraphExists(ctx, resourceName, &graph),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, acctest.CtFalse),
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

func TestAccNeptuneGraphGraph_nameGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var graph neptunegraph.GetGraphOutput
	resourceName := "aws_neptunegraph_graph.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AttrID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphConfig_nameGenerated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGraphExists(ctx, resourceName, &graph),
					resource.TestCheckResourceAttrSet(resourceName, "graph_name"),
				),
			},
		},
	})
}

func TestAccNeptuneGraphGraph_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var graph neptunegraph.GetGraphOutput
	resourceName := "aws_neptunegraph_graph.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AttrID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphConfig_namePrefix("tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphExists(ctx, resourceName, &graph),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, "graph_name", "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, "graph_name_prefix", "tf-acc-test-prefix-"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"graph_name_prefix"},
			},
		},
	})
}

func TestAccNeptuneGraphGraph_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var graph neptunegraph.GetGraphOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptunegraph_graph.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AttrID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphExists(ctx, resourceName, &graph),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"graph_name_prefix"},
			},
			{
				Config: testAccGraphConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphExists(ctx, resourceName, &graph),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccGraphConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphExists(ctx, resourceName, &graph),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckGraphDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).NeptuneGraphClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_neptunegraph_graph" {
				continue
			}

			_, err := tfneptunegraph.FindGraphByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Neptune Graph Graph %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckGraphExists(ctx context.Context, n string, v *neptunegraph.GetGraphOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NeptuneGraphClient(ctx)

		output, err := tfneptunegraph.FindGraphByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).NeptuneGraphClient(ctx)
	var input neptunegraph.ListGraphsInput

	_, err := conn.ListGraphs(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccGraphConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_neptunegraph_graph" "test" {
  graph_name          = %[1]q
  provisioned_memory  = 16
  public_connectivity = false
  replica_count       = 0
  deletion_protection = false
}
`, rName)
}

func testAccGraphConfig_vectorSearch(rName string, dimensions int) string {
	return fmt.Sprintf(`
resource "aws_neptunegraph_graph" "test" {
  graph_name          = %[1]q
  provisioned_memory  = 16
  public_connectivity = false
  replica_count       = 0
  deletion_protection = false
  vector_search_configuration {
    vector_search_dimension = %[2]d
  }
}
`, rName, dimensions)
}

func testAccGraphConfig_kmsKey(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}

resource "aws_iam_role" "kms_admin" {
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      }
      Action = "sts:AssumeRole"
    }]
  })

  managed_policy_arns = ["arn:${data.aws_partition.current.partition}:iam::aws:policy/NeptuneConsoleFullAccess"]

  inline_policy {
    name = "kms-perms-for-neptune-analytics"
    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [{
        Action = [
          "kms:Encrypt",
          "kms:Decrypt",
          "kms:GenerateDataKey",
          "kms:ReEncryptTo",
          "kms:GenerateDataKeyWithoutPlaintext",
          "kms:CreateGrant",
          "kms:ReEncryptFrom",
          "kms:DescribeKey"
        ]
        Effect   = "Allow"
        Resource = "*"
      }]
    })
  }
}

resource "aws_kms_key" "test" {
  description = %[1]q

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable Permissions for root principal",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      },
      "Action": "kms:*",
      "Resource": "*"
    },
    {
      "Sid": "Allow use of the key for Neptune Analytics",
      "Effect": "Allow",
      "Principal": {
        "AWS": "${aws_iam_role.kms_admin.arn}"
      },
      "Action": [
        "kms:Encrypt",
        "kms:Decrypt",
        "kms:GenerateDataKey",
        "kms:ReEncryptTo",
        "kms:GenerateDataKeyWithoutPlaintext",
        "kms:CreateGrant",
        "kms:ReEncryptFrom",
        "kms:DescribeKey"
      ],
      "Resource": "*",
      "Condition": {
        "StringEquals": {
          "kms:ViaService": "neptune-graph.amazonaws.com"
        }
      }
    },
    {
      "Sid": "Deny use of the key for non Neptune Analytics",
      "Effect": "Deny",
      "Principal": {
        "AWS": "${aws_iam_role.kms_admin.arn}"
      },
      "Action": [
        "kms:*"
      ],
      "Resource": "*",
      "Condition": {
        "StringNotEquals": {
          "kms:ViaService": "neptune-graph.amazonaws.com"
        }
      }
    }
  ]
}
POLICY
}

resource "aws_neptunegraph_graph" "test" {
  graph_name          = %[1]q
  provisioned_memory  = 16
  public_connectivity = false
  replica_count       = 0
  deletion_protection = false
  kms_key_identifier  = aws_kms_key.test.arn
}
`, rName)
}

func testAccGraphConfig_deletionProtection(rName string, dp bool) string {
	return fmt.Sprintf(`
resource "aws_neptunegraph_graph" "test" {
  graph_name          = %[1]q
  provisioned_memory  = 16
  public_connectivity = false
  replica_count       = 0
  deletion_protection = %[2]t
}
`, rName, dp)
}

func testAccGraphConfig_nameGenerated() string {
	return `
resource "aws_neptunegraph_graph" "test" {
  provisioned_memory  = 16
  deletion_protection = false
}
`
}

func testAccGraphConfig_namePrefix(namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_neptunegraph_graph" "test" {
  provisioned_memory  = 16
  deletion_protection = false
  graph_name_prefix   = %[1]q
}
`, namePrefix)
}

func testAccGraphConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_neptunegraph_graph" "test" {
  provisioned_memory  = 16
  deletion_protection = false
  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccGraphConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_neptunegraph_graph" "test" {
  provisioned_memory  = 16
  deletion_protection = false
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
