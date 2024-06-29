// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/redshiftserverless"
	"github.com/aws/aws-sdk-go-v2/service/redshiftserverless/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfredshiftserverless "github.com/hashicorp/terraform-provider-aws/internal/service/redshiftserverless"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRedshiftServerlessSnapshotCopyConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var snapshotCopyConfiguration types.SnapshotCopyConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_redshiftserverless_snapshot_copy_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.RedshiftServerlessEndpointID)
			testAccPreCheckSnapshotCopyConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotCopyConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotCopyConfigurationConfig_basic(rName, acctest.AlternateRegion()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotCopyConfigurationExists(ctx, resourceName, &snapshotCopyConfiguration),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "redshift-serverless", regexache.MustCompile(`snapshotcopyconfiguration/+.`)),
					resource.TestCheckResourceAttr(resourceName, "namespace_name", rName),
					resource.TestCheckResourceAttr(resourceName, "destination_region", acctest.AlternateRegion()),
					resource.TestCheckResourceAttr(resourceName, "destination_kms_key_id", "AWS_OWNED_KMS_KEY"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_retention_period", "-1"),
				),
			},
			{
				Config: testAccSnapshotCopyConfigurationConfig_retention(rName, acctest.AlternateRegion(), 100),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotCopyConfigurationExists(ctx, resourceName, &snapshotCopyConfiguration),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "redshift-serverless", regexache.MustCompile(`snapshotcopyconfiguration/+.`)),
					resource.TestCheckResourceAttr(resourceName, "namespace_name", rName),
					resource.TestCheckResourceAttr(resourceName, "destination_region", acctest.AlternateRegion()),
					resource.TestCheckResourceAttr(resourceName, "destination_kms_key_id", "AWS_OWNED_KMS_KEY"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_retention_period", "100"),
				),
			},
			{
				Config: testAccSnapshotCopyConfigurationConfig_basic(rName, acctest.AlternateRegion()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotCopyConfigurationExists(ctx, resourceName, &snapshotCopyConfiguration),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "redshift-serverless", regexache.MustCompile(`snapshotcopyconfiguration/+.`)),
					resource.TestCheckResourceAttr(resourceName, "namespace_name", rName),
					resource.TestCheckResourceAttr(resourceName, "destination_region", acctest.AlternateRegion()),
					resource.TestCheckResourceAttr(resourceName, "destination_kms_key_id", "AWS_OWNED_KMS_KEY"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_retention_period", "-1"),
				),
			},
		},
	})
}

func TestAccRedshiftServerlessSnapshotCopyConfiguration_retention(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var snapshotCopyConfiguration types.SnapshotCopyConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_redshiftserverless_snapshot_copy_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.RedshiftServerlessEndpointID)
			testAccPreCheckSnapshotCopyConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotCopyConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotCopyConfigurationConfig_retention(rName, acctest.AlternateRegion(), 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotCopyConfigurationExists(ctx, resourceName, &snapshotCopyConfiguration),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "redshift-serverless", regexache.MustCompile(`snapshotcopyconfiguration/+.`)),
					resource.TestCheckResourceAttr(resourceName, "namespace_name", rName),
					resource.TestCheckResourceAttr(resourceName, "destination_region", acctest.AlternateRegion()),
					resource.TestCheckResourceAttr(resourceName, "destination_kms_key_id", "AWS_OWNED_KMS_KEY"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_retention_period", "1"),
				),
			},
			{
				Config: testAccSnapshotCopyConfigurationConfig_retention(rName, acctest.AlternateRegion(), 100),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotCopyConfigurationExists(ctx, resourceName, &snapshotCopyConfiguration),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "redshift-serverless", regexache.MustCompile(`snapshotcopyconfiguration/+.`)),
					resource.TestCheckResourceAttr(resourceName, "namespace_name", rName),
					resource.TestCheckResourceAttr(resourceName, "destination_region", acctest.AlternateRegion()),
					resource.TestCheckResourceAttr(resourceName, "destination_kms_key_id", "AWS_OWNED_KMS_KEY"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_retention_period", "100"),
				),
			},
		},
	})
}

func TestAccRedshiftServerlessSnapshotCopyConfiguration_kms(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var snapshotCopyConfiguration types.SnapshotCopyConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_redshiftserverless_snapshot_copy_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.RedshiftServerlessEndpointID)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheckSnapshotCopyConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		CheckDestroy:             testAccCheckSnapshotCopyConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotCopyConfigurationConfig_kms(rName, acctest.AlternateRegion()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotCopyConfigurationExists(ctx, resourceName, &snapshotCopyConfiguration),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "redshift-serverless", regexache.MustCompile(`snapshotcopyconfiguration/+.`)),
					resource.TestCheckResourceAttr(resourceName, "namespace_name", rName),
					resource.TestCheckResourceAttr(resourceName, "destination_region", acctest.AlternateRegion()),
					acctest.MatchResourceAttrRegionalARNRegion(resourceName, "destination_kms_key_id", "kms", acctest.AlternateRegion(), regexache.MustCompile(`key/+.`)),
				),
			},
		},
	})
}

func TestAccRedshiftServerlessSnapshotCopyConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var snapshotCopyConfiguration types.SnapshotCopyConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_redshiftserverless_snapshot_copy_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.RedshiftServerlessEndpointID)
			testAccPreCheckSnapshotCopyConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotCopyConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotCopyConfigurationConfig_basic(rName, acctest.AlternateRegion()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotCopyConfigurationExists(ctx, resourceName, &snapshotCopyConfiguration),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfredshiftserverless.ResourceSnapshotCopyConfiguration, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSnapshotCopyConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftServerlessClient(ctx)
		connNamespace := acctest.Provider.Meta().(*conns.AWSClient).RedshiftServerlessConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_redshiftserverless_snapshot_copy_configuration" {
				continue
			}

			_, errNamespace := tfredshiftserverless.FindNamespaceByName(ctx, connNamespace, rs.Primary.Attributes["namespace_name"])
			if tfresource.NotFound(errNamespace) {
				continue
			}
			if errNamespace != nil {
				return errNamespace
			}

			_, err := tfredshiftserverless.FindSnapshotCopyConfigurationByID(ctx, conn, rs.Primary.Attributes["namespace_name"], rs.Primary.Attributes["id"])
			if tfresource.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("Redshift Serverless Snapshot Copy Configuration %s still exists", rs.Primary.ID)
		}
		return nil
	}
}

func testAccCheckSnapshotCopyConfigurationExists(ctx context.Context, name string, snapshotCopyConfiguration *types.SnapshotCopyConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.RedshiftServerless, create.ErrActionCheckingExistence, tfredshiftserverless.ResNameSnapshotCopyConfiguration, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.RedshiftServerless, create.ErrActionCheckingExistence, tfredshiftserverless.ResNameSnapshotCopyConfiguration, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftServerlessClient(ctx)
		output, err := tfredshiftserverless.FindSnapshotCopyConfigurationByID(ctx, conn, rs.Primary.Attributes["namespace_name"], rs.Primary.Attributes["id"])

		if err != nil {
			return err
		}

		*snapshotCopyConfiguration = *output

		return nil
	}
}

func testAccPreCheckSnapshotCopyConfiguration(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftServerlessClient(ctx)

	input := &redshiftserverless.ListSnapshotCopyConfigurationsInput{}
	_, err := conn.ListSnapshotCopyConfigurations(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccSnapshotCopyConfigurationConfig_basic(rName string, region string) string {
	return fmt.Sprintf(`
resource "aws_redshiftserverless_namespace" "test" {
  namespace_name = %[1]q
}

resource "aws_redshiftserverless_snapshot_copy_configuration" "test" {
  namespace_name     = aws_redshiftserverless_namespace.test.namespace_name
  destination_region = %[2]q
}
`, rName, region)
}

func testAccSnapshotCopyConfigurationConfig_retention(rName string, region string, retention int) string {
	return fmt.Sprintf(`
resource "aws_redshiftserverless_namespace" "test" {
  namespace_name = %[1]q
}

resource "aws_redshiftserverless_snapshot_copy_configuration" "test" {
  namespace_name            = aws_redshiftserverless_namespace.test.namespace_name
  destination_region        = %[2]q
  snapshot_retention_period = %[3]d
}
`, rName, region, retention)
}

func testAccSnapshotCopyConfigurationConfig_kms(rName string, region string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		fmt.Sprintf(`

resource "aws_kms_key" "test" {
  provider = awsalternate
}

resource "aws_redshiftserverless_namespace" "test" {
  namespace_name = %[1]q
}

resource "aws_redshiftserverless_snapshot_copy_configuration" "test" {
  namespace_name         = aws_redshiftserverless_namespace.test.namespace_name
  destination_region     = %[2]q
  destination_kms_key_id = aws_kms_key.test.arn
}
`, rName, region))
}
