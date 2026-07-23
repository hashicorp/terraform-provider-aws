// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMCustomPolicySimulationDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_iam_custom_policy_simulation.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomPolicySimulationDataSourceConfig_basic,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("all_allowed"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("results"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"action_name":         knownvalue.StringExact("s3:GetObject"),
							"allowed":             knownvalue.Bool(true),
							"decision":            knownvalue.StringExact("allowed"),
							names.AttrResourceARN: knownvalue.StringExact("*"),
						}),
					})),
				},
			},
		},
	})
}

func TestAccIAMCustomPolicySimulationDataSource_implicitDeny(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_iam_custom_policy_simulation.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomPolicySimulationDataSourceConfig_implicitDeny,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("all_allowed"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("results"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"action_name": knownvalue.StringExact("s3:DeleteBucket"),
							"allowed":     knownvalue.Bool(false),
							"decision":    knownvalue.StringExact("implicitDeny"),
						}),
					})),
				},
			},
		},
	})
}

func TestAccIAMCustomPolicySimulationDataSource_explicitDeny(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_iam_custom_policy_simulation.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomPolicySimulationDataSourceConfig_explicitDeny,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("all_allowed"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("results"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"action_name": knownvalue.StringExact("s3:GetObject"),
							"allowed":     knownvalue.Bool(false),
							"decision":    knownvalue.StringExact("explicitDeny"),
						}),
					})),
				},
			},
		},
	})
}

func TestAccIAMCustomPolicySimulationDataSource_multipleActions(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_iam_custom_policy_simulation.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomPolicySimulationDataSourceConfig_multipleActions,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("all_allowed"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("results"), knownvalue.ListSizeExact(2)),
				},
			},
		},
	})
}

func TestAccIAMCustomPolicySimulationDataSource_withContext(t *testing.T) {
	ctx := acctest.Context(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomPolicySimulationDataSourceConfig_withContext,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.aws_iam_custom_policy_simulation.match", tfjsonpath.New("all_allowed"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("data.aws_iam_custom_policy_simulation.no_match", tfjsonpath.New("all_allowed"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("data.aws_iam_custom_policy_simulation.no_context", tfjsonpath.New("all_allowed"), knownvalue.Bool(false)),
				},
			},
		},
	})
}

func TestAccIAMCustomPolicySimulationDataSource_permissionsBoundary(t *testing.T) {
	ctx := acctest.Context(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomPolicySimulationDataSourceConfig_permissionsBoundary,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.aws_iam_custom_policy_simulation.without_boundary", tfjsonpath.New("all_allowed"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("data.aws_iam_custom_policy_simulation.with_boundary", tfjsonpath.New("all_allowed"), knownvalue.Bool(false)),
				},
			},
		},
	})
}

const testAccCustomPolicySimulationDataSourceConfig_basic = `
data "aws_iam_custom_policy_simulation" "test" {
  action_names = ["s3:GetObject"]

  policies_json = [jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = "s3:GetObject"
      Resource = "*"
    }]
  })]
}
`

const testAccCustomPolicySimulationDataSourceConfig_implicitDeny = `
data "aws_iam_custom_policy_simulation" "test" {
  action_names = ["s3:DeleteBucket"]

  policies_json = [jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = "s3:GetObject"
      Resource = "*"
    }]
  })]
}
`

const testAccCustomPolicySimulationDataSourceConfig_explicitDeny = `
data "aws_iam_custom_policy_simulation" "test" {
  action_names = ["s3:GetObject"]

  policies_json = [jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = "s3:*"
        Resource = "*"
      },
      {
        Effect   = "Deny"
        Action   = "s3:GetObject"
        Resource = "*"
      },
    ]
  })]
}
`

const testAccCustomPolicySimulationDataSourceConfig_multipleActions = `
data "aws_iam_custom_policy_simulation" "test" {
  action_names = ["s3:GetObject", "s3:DeleteBucket"]

  policies_json = [jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = "s3:GetObject"
      Resource = "*"
    }]
  })]
}
`

const testAccCustomPolicySimulationDataSourceConfig_withContext = `
data "aws_iam_custom_policy_simulation" "match" {
  action_names = ["s3:GetObject"]

  policies_json = [jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = "s3:GetObject"
      Resource = "*"
      Condition = {
        StringEquals = {
          "s3:prefix" = "test"
        }
      }
    }]
  })]

  context {
    key    = "s3:prefix"
    type   = "string"
    values = ["test"]
  }
}

data "aws_iam_custom_policy_simulation" "no_match" {
  action_names = ["s3:GetObject"]

  policies_json = [jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = "s3:GetObject"
      Resource = "*"
      Condition = {
        StringEquals = {
          "s3:prefix" = "test"
        }
      }
    }]
  })]

  context {
    key    = "s3:prefix"
    type   = "string"
    values = ["wrong"]
  }
}

data "aws_iam_custom_policy_simulation" "no_context" {
  action_names = ["s3:GetObject"]

  policies_json = [jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = "s3:GetObject"
      Resource = "*"
      Condition = {
        StringEquals = {
          "s3:prefix" = "test"
        }
      }
    }]
  })]
}
`

const testAccCustomPolicySimulationDataSourceConfig_permissionsBoundary = `
data "aws_iam_custom_policy_simulation" "without_boundary" {
  action_names = ["s3:GetObject"]

  policies_json = [jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = "s3:*"
      Resource = "*"
    }]
  })]
}

data "aws_iam_custom_policy_simulation" "with_boundary" {
  action_names = ["s3:GetObject"]

  policies_json = [jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = "s3:*"
      Resource = "*"
    }]
  })]

  permissions_boundary_policies_json = [jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = "ec2:*"
      Resource = "*"
    }]
  })]
}
`
