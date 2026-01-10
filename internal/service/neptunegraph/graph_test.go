// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptunegraph_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/neptunegraph"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfneptunegraph "github.com/hashicorp/terraform-provider-aws/internal/service/neptunegraph"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNeptuneGraphGraph_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var graph neptunegraph.GetGraphOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_neptunegraph_graph.test"
	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneGraphServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGraphExists(ctx, t, resourceName, &graph),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_neptunegraph_graph.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AttrID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGraphExists(ctx, t, resourceName, &graph),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_neptunegraph_graph.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AttrID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphConfig_vectorSearch(rName, 128),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGraphExists(ctx, t, resourceName, &graph),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_neptunegraph_graph.test"
	keyResourceName := "aws_kms_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AttrID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphConfig_kmsKey(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGraphExists(ctx, t, resourceName, &graph),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_neptunegraph_graph.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AttrID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphConfig_deletionProtection(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGraphExists(ctx, t, resourceName, &graph),
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
					testAccCheckGraphExists(ctx, t, resourceName, &graph),
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

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AttrID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphConfig_nameGenerated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGraphExists(ctx, t, resourceName, &graph),
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

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AttrID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphConfig_namePrefix("tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphExists(ctx, t, resourceName, &graph),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_neptunegraph_graph.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AttrID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphExists(ctx, t, resourceName, &graph),
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
					testAccCheckGraphExists(ctx, t, resourceName, &graph),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccGraphConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphExists(ctx, t, resourceName, &graph),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckGraphDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).NeptuneGraphClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_neptunegraph_graph" {
				continue
			}

			_, err := tfneptunegraph.FindGraphByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
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

func testAccCheckGraphExists(ctx context.Context, t *testing.T, n string, v *neptunegraph.GetGraphOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).NeptuneGraphClient(ctx)

		output, err := tfneptunegraph.FindGraphByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).NeptuneGraphClient(ctx)
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
  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = true

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

func TestAccNeptuneGraphGraph_importFromS3(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var graph neptunegraph.GetGraphOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_neptunegraph_graph.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneGraphServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"time": {
				Source:            "hashicorp/time",
				VersionConstraint: "0.12.1",
			},
		},
		CheckDestroy: testAccCheckGraphDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphConfig_importFromS3(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGraphExists(ctx, t, resourceName, &graph),
					resource.TestCheckResourceAttr(resourceName, "import_task.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "import_task.0.format", "CSV"),
					resource.TestCheckResourceAttrSet(resourceName, "import_task.0.source"),
					resource.TestCheckResourceAttrSet(resourceName, "import_task.0.role_arn"),
				),
			},
			{
				Config: testAccGraphConfig_importFromS3_removed(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGraphExists(ctx, t, resourceName, &graph),
					resource.TestCheckResourceAttr(resourceName, "import_task.#", "0"),
				),
			},
			{
				Config: testAccGraphConfig_importFromS3_modified(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGraphExists(ctx, t, resourceName, &graph),
					resource.TestCheckResourceAttr(resourceName, "import_task.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "import_task.0.format", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "import_task.0.fail_on_error", "true"),
				),
			},
		},
	})
}

func testAccGraphConfig_importFromS3(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

# Create minimal Neptune CSV data
resource "aws_s3_object" "vertices" {
  bucket  = aws_s3_bucket.test.bucket
  key     = "vertices/vertices.csv"
  content = "~id,~label\nv1,airport\nv2,airport\n"
}

resource "aws_s3_object" "edges" {
  bucket  = aws_s3_bucket.test.bucket
  key     = "edges/edges.csv"
  content = "~id,~from,~to,~label\ne1,v1,v2,route\n"
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
		Service = "neptune-graph.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = [
        "s3:Get*",
        "s3:List*"
      ]
      Resource = [
        "${aws_s3_bucket.test.arn}/*",
        aws_s3_bucket.test.arn
      ]
    }]
  })
}

# Allow time for IAM role to propagate
resource "time_sleep" "wait" {
  depends_on = [aws_iam_role_policy.test]
  create_duration = "10s"
}

resource "aws_neptunegraph_graph" "test" {
  graph_name          = %[1]q
  deletion_protection = false
  public_connectivity = false
  replica_count       = 0

  import_task {
    source   = "s3://${aws_s3_bucket.test.bucket}/"
    role_arn = aws_iam_role.test.arn
    format   = "CSV"
  }

  depends_on = [
    time_sleep.wait,
    aws_s3_object.vertices,
    aws_s3_object.edges
  ]
}
`, rName)
}

func testAccGraphConfig_importFromS3_removed(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
		Service = "neptune-graph.amazonaws.com"
      }
    }]
  })
}

resource "aws_neptunegraph_graph" "test" {
  graph_name          = %[1]q
  provisioned_memory  = 128
  deletion_protection = false
  public_connectivity = false
  replica_count       = 0
}
`, rName)
}

func testAccGraphConfig_importFromS3_modified(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_object" "vertices" {
  bucket  = aws_s3_bucket.test.bucket
  key     = "vertices/vertices.csv"
  content = "~id,~label\nv1,airport\nv2,airport\n"
}

resource "aws_s3_object" "edges" {
  bucket  = aws_s3_bucket.test.bucket
  key     = "edges/edges.csv"
  content = "~id,~from,~to,~label\ne1,v1,v2,route\n"
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
		Service = "neptune-graph.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = [
        "s3:Get*",
        "s3:List*"
      ]
      Resource = [
        "${aws_s3_bucket.test.arn}/*",
        aws_s3_bucket.test.arn
      ]
    }]
  })
}

resource "time_sleep" "wait" {
  depends_on = [aws_iam_role_policy.test]
  create_duration = "10s"
}

resource "aws_neptunegraph_graph" "test" {
  graph_name          = %[1]q
  deletion_protection = false
  public_connectivity = false
  replica_count       = 0

  import_task {
    source        = "s3://${aws_s3_bucket.test.bucket}/"
    role_arn      = aws_iam_role.test.arn
    format        = "CSV"
    fail_on_error = true
  }

  depends_on = [
    time_sleep.wait,
    aws_s3_object.vertices,
    aws_s3_object.edges
  ]
}
`, rName)
}

func TestAccNeptuneGraphGraph_importFromNeptuneDB(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var graph neptunegraph.GetGraphOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_neptunegraph_graph.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneGraphServiceID, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"null": {
				Source:            "hashicorp/null",
				VersionConstraint: "3.2.2",
			},
			"time": {
				Source:            "hashicorp/time",
				VersionConstraint: "0.12.1",
			},
		},
		CheckDestroy: testAccCheckGraphDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphConfig_importFromNeptuneDB(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGraphExists(ctx, t, resourceName, &graph),
					resource.TestCheckResourceAttr(resourceName, "import_task.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "import_task.0.import_options.neptune.s3_export_path"),
					resource.TestCheckResourceAttrSet(resourceName, "import_task.0.import_options.neptune.s3_export_kms_key_id"),
					resource.TestCheckResourceAttrSet(resourceName, "import_task.0.source"),
					resource.TestCheckResourceAttrSet(resourceName, "import_task.0.role_arn"),
				),
			},
		},
	})
}

func testAccGraphConfig_importFromNeptuneDB(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}
data "aws_partition" "current" {}
data "aws_caller_identity" "current" {}
data "aws_neptune_engine_version" "latest" {
    latest = true
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_object" "vertices" {
  bucket  = aws_s3_bucket.test.bucket
  key     = "vertices.csv"
  content = "~id,~label\nv1,airport\nv2,airport\n"
}

resource "aws_s3_object" "edges" {
  bucket  = aws_s3_bucket.test.bucket
  key     = "edges.csv"
  content = "~id,~from,~to,~label\ne1,v1,v2,route\n"
}

resource "aws_iam_role" "neptune_load" {
  name = "%[1]s-load"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "rds.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy" "neptune_load" {
  role = aws_iam_role.neptune_load.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = [
        "s3:Get*",
        "s3:List*"
      ]
      Resource = [
        "${aws_s3_bucket.test.arn}/*",
        aws_s3_bucket.test.arn
      ]
    }]
  })
}

resource "aws_default_vpc" "test" {}

data "aws_subnets" "test" {
  filter {
    name   = "vpc-id"
    values = [aws_default_vpc.test.id]
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_default_vpc.test.id

  ingress {
    from_port   = 8182
    to_port     = 8182
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_neptune_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = data.aws_subnets.test.ids
}

resource "aws_neptune_cluster" "test" {
  cluster_identifier          = %[1]q
  engine                      = "neptune"
  engine_version              = data.aws_neptune_engine_version.latest.version_actual
  skip_final_snapshot         = true
  neptune_subnet_group_name   = aws_neptune_subnet_group.test.name
  apply_immediately           = true
  iam_roles                   = [aws_iam_role.neptune_load.arn]
  vpc_security_group_ids      = [aws_security_group.test.id]
  iam_database_authentication_enabled = true
}

resource "aws_neptune_cluster_instance" "test" {
  identifier                 = %[1]q
  cluster_identifier         = aws_neptune_cluster.test.id
  instance_class             = "db.t4g.medium"
  apply_immediately          = true
  publicly_accessible        = true
}

resource "time_sleep" "neptune_ready" {
  depends_on      = [aws_neptune_cluster_instance.test]
  create_duration = "60s"
}

resource "null_resource" "bulk_load" {
  depends_on = [time_sleep.neptune_ready]

  provisioner "local-exec" {
    command = <<-EOT
      LOAD_ID=$(aws neptunedata start-loader-job \
        --source s3://${aws_s3_bucket.test.bucket}/ \
        --format csv \
        --s3-bucket-region ${data.aws_region.current.region} \
        --iam-role-arn ${aws_iam_role.neptune_load.arn} \
        --region ${data.aws_region.current.region} \
        --endpoint-url https://${aws_neptune_cluster.test.endpoint}:${aws_neptune_cluster.test.port} \
        --query 'payload.loadId' \
        --output text)
      
      echo "Waiting for load job $LOAD_ID to complete..."
      while true; do
        STATUS=$(aws neptunedata get-loader-job-status \
          --load-id "$LOAD_ID" \
          --endpoint-url https://${aws_neptune_cluster.test.endpoint}:${aws_neptune_cluster.test.port} \
          --region ${data.aws_region.current.region} \
          --query 'payload.overallStatus.status' \
          --output text)
        
        if [ "$STATUS" = "LOAD_COMPLETED" ]; then
          echo "Load completed successfully"
          break
        elif [ "$STATUS" = "LOAD_FAILED" ] || [ "$STATUS" = "LOAD_CANCELLED" ]; then
          echo "Load failed with status: $STATUS"
          exit 1
        fi
        
        echo "Current status: $STATUS, waiting..."
        sleep 10
      done
    EOT
  }

  triggers = {
    cluster_id = aws_neptune_cluster.test.id
    bucket     = aws_s3_bucket.test.id
  }
}

resource "aws_iam_role" "graph_import" {
  name = "%[1]s-graph"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = [
			"neptune-graph.amazonaws.com",
			"export.rds.amazonaws.com"
		]
      }
    }]
  })
}

resource "aws_iam_role_policy" "graph_import" {
  role = aws_iam_role.graph_import.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
        {
            Effect = "Allow",
            Action = [
                "iam:PassRole"
            ],
            Resource = aws_iam_role.graph_import.arn
        },
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject*",
          "s3:ListBucket",
          "s3:PutObject*",
          "s3:GetBucketLocation",
          "s3:DeleteObject*"
        ]
        Resource = [
          "${aws_s3_bucket.test.arn}/*",
          aws_s3_bucket.test.arn
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "kms:ListGrants",
          "kms:CreateGrant",
          "kms:RevokeGrant",
          "kms:Encrypt",
          "kms:Decrypt",
          "kms:ReEncrypt*",
          "kms:DescribeKey",
          "kms:GenerateDataKey"
        ]
        Resource = aws_kms_key.test.arn
      },
      {
        Effect = "Allow"
        Action = [
          "rds:DescribeDBClusters",
          "rds:StartExportTask"
        ]
        Resource = aws_neptune_cluster.test.arn
      },
      {
        Effect = "Allow"
        Action = [
          "rds:DescribeExportTasks",
          "rds:CancelExportTask"
        ]
        Resource = "*"
      }    
    ]
  })
}

resource "time_sleep" "iam_propagate" {
  depends_on      = [aws_iam_role_policy.graph_import]
  create_duration = "10s"
}

resource "aws_kms_key" "test" {
  description             = "KMS key for Neptune DB export encryption"
  deletion_window_in_days = 7

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow Neptune Graph to use the key"
        Effect = "Allow"
        Principal = {
          Service = "neptune-graph.amazonaws.com"
        }
        Action = [
          "kms:Decrypt",
          "kms:DescribeKey",
          "kms:CreateGrant"
        ]
        Resource = "*"
      },
      {
        Sid    = "Allow S3 to use the key"
        Effect = "Allow"
        Principal = {
          Service = "s3.amazonaws.com"
        }
        Action = [
          "kms:Decrypt",
          "kms:GenerateDataKey"
        ]
        Resource = "*"
      }
    ]
  })
}

resource "aws_neptunegraph_graph" "test" {
  graph_name          = %[1]q
  deletion_protection = false
  public_connectivity = false
  replica_count       = 0

  import_task {
    source   = aws_neptune_cluster.test.arn
    role_arn = aws_iam_role.graph_import.arn

    import_options {
      neptune {
        s3_export_path       = "s3://${aws_s3_bucket.test.bucket}/export/"
        s3_export_kms_key_id = aws_kms_key.test.arn
        preserve_default_vertex_labels = true
        preserve_edge_ids              = true
      }
    }
  }

  depends_on = [
    null_resource.bulk_load,
    time_sleep.iam_propagate
  ]
}`, rName)
}
