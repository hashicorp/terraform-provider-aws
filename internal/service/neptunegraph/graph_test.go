// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptunegraph_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/neptunegraph"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
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
					resource.TestCheckResourceAttr(resourceName, "public_connectivity", "false"),
					resource.TestCheckResourceAttr(resourceName, "replica_count", "0"),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
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
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "true"),
				),
			},
			{
				Config:      testAccGraphConfig_deletionProtection(rName, true),
				Destroy:     true,
				ExpectError: regexache.MustCompile(`Graph .* cannot be deleted because deletion protection is enabled.`),
			},
			{
				Config: testAccGraphConfig_deletionProtection(rName, false),
			},
			{
				Config:  testAccGraphConfig_deletionProtection(rName, false),
				Destroy: true,
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
				return nil
			}
			if err != nil {
				return create.Error(names.NeptuneGraph, create.ErrActionCheckingDestroyed, tfneptunegraph.ResNameGraph, rs.Primary.ID, err)
			}

			return create.Error(names.NeptuneGraph, create.ErrActionCheckingDestroyed, tfneptunegraph.ResNameGraph, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckGraphExists(ctx context.Context, name string, graph *neptunegraph.GetGraphOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.NeptuneGraph, create.ErrActionCheckingExistence, tfneptunegraph.ResNameGraph, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.NeptuneGraph, create.ErrActionCheckingExistence, tfneptunegraph.ResNameGraph, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NeptuneGraphClient(ctx)

		resp, err := tfneptunegraph.FindGraphByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.NeptuneGraph, create.ErrActionCheckingExistence, tfneptunegraph.ResNameGraph, rs.Primary.ID, err)
		}

		*graph = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).NeptuneGraphClient(ctx)

	input := &neptunegraph.ListGraphsInput{}

	_, err := conn.ListGraphs(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckGraphNotRecreated(before, after *neptunegraph.GetGraphOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.Id), aws.ToString(after.Id); before != after {
			return create.Error(names.NeptuneGraph, create.ErrActionCheckingNotRecreated, tfneptunegraph.ResNameGraph, before, errors.New("recreated"))
		}

		return nil
	}
}

func testAccGraphConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_neptunegraph_graph" "test" {
    graph_name = %[1]q
    provisioned_memory = 16
    public_connectivity = false
    replica_count = 0
    deletion_protection = false
}
`, rName)
}

func testAccGraphConfig_vectorSearch(rName string, dimensions int) string {
	return fmt.Sprintf(`
resource "aws_neptunegraph_graph" "test" {
    graph_name = %[1]q
    provisioned_memory = 16
    public_connectivity = false
    replica_count = 0
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
    graph_name = %[1]q
    provisioned_memory = 16
    public_connectivity = false
    replica_count = 0
    deletion_protection = false
    kms_key_identifier = aws_kms_key.test.arn
}
`, rName)
}

func testAccGraphConfig_deletionProtection(rName string, dp bool) string {
	return fmt.Sprintf(`
resource "aws_neptunegraph_graph" "test" {
    graph_name = %[1]q
    provisioned_memory = 16
    public_connectivity = false
    replica_count = 0
    deletion_protection = %[2]t
}
`, rName, dp)
}

func testAccGraphConfig_nameGenerated() string {
	return fmt.Sprintf(`
resource "aws_neptunegraph_graph" "test" {
    provisioned_memory = 16
    deletion_protection = false
}
`)
}

func testAccGraphConfig_namePrefix(namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_neptunegraph_graph" "test" {
    provisioned_memory = 16
    deletion_protection = false
	graph_name_prefix = %[1]q
}
`, namePrefix)
}

func testAccGraphConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_neptunegraph_graph" "test" {
	provisioned_memory = 16
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
	provisioned_memory = 16
	deletion_protection = false
	tags = {
    	%[2]q = %[3]q
    	%[4]q = %[5]q
	}
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
