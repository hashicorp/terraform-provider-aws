// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emr_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEMRReleaseLabels_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceResourceName := "data.aws_emr_release_labels.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccReleaseLabelsDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceResourceName, "release_labels.#"),
				),
			},
		},
	})
}

func TestAccEMRReleaseLabels_prefix(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceResourceName := "data.aws_emr_release_labels.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccReleaseLabelsDataSourceConfig_prefix(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceResourceName, "release_labels.#"),
				),
			},
		},
	})
}

func TestAccEMRReleaseLabels_application(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceResourceName := "data.aws_emr_release_labels.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccReleaseLabelsDataSourceConfig_application(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceResourceName, "release_labels.#"),
				),
			},
		},
	})
}

func TestAccEMRReleaseLabels_full(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceResourceName := "data.aws_emr_release_labels.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccReleaseLabelsDataSourceConfig_full(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceResourceName, "release_labels.#"),
				),
			},
		},
	})
}

func TestAccEMRReleaseLabels_empty(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceResourceName := "data.aws_emr_release_labels.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccReleaseLabelsDataSourceConfig_empty(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceResourceName, "release_labels.#", acctest.Ct0),
				),
			},
		},
	})
}

func testAccReleaseLabelsDataSourceConfig_basic() string {
	return `
data "aws_emr_release_labels" "test" {}
`
}

func testAccReleaseLabelsDataSourceConfig_prefix() string {
	return `
data "aws_emr_release_labels" "test" {
  filters {
    prefix = "emr-6"
  }
}
`
}

func testAccReleaseLabelsDataSourceConfig_application() string {
	return `
data "aws_emr_release_labels" "test" {
  filters {
    application = "Spark@3.1.2"
  }
}
`
}

func testAccReleaseLabelsDataSourceConfig_full() string {
	return `
data "aws_emr_release_labels" "test" {
  filters {
    application = "Spark@3.1.2"
    prefix      = "emr-6"
  }
}
`
}

func testAccReleaseLabelsDataSourceConfig_empty() string {
	return `
data "aws_emr_release_labels" "test" {
  filters {
    prefix = "emr-0"
  }
}
`
}
