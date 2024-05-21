// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appfabric_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/appfabric/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"

	tfappfabric "github.com/hashicorp/terraform-provider-aws/internal/service/appfabric"
)

func TestAccAppFabricIngestionDestination_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_appfabric_ingestion_destination.test"
	var ingestiondestination types.IngestionDestination

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIngestionDestinationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIngestionDestinationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIngestionDestinationExists(ctx, resourceName, &ingestiondestination),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.destination.0.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.destination.0.s3_bucket.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.destination.0.s3_bucket.0.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.destination.0.s3_bucket.0.bucket_name", "appfabric-test-audit-logs-658465413021"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.destination.0.s3_bucket.0.prefix", "AuditLog"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"processing_configuration"},
			},
		},
	})
}

func TestAccAppFabricIngestionDestination_firehose(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_appfabric_ingestion_destination.test"
	var ingestiondestination types.IngestionDestination

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIngestionDestinationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIngestionDestinationConfig_firehose(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIngestionDestinationExists(ctx, resourceName, &ingestiondestination),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.destination.0.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.destination.0.firehose_stream.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.destination.0.firehose_stream.0.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.destination.0.firehose_stream.0.stream_name", "KDS-HTP-uMu4o"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"processing_configuration"},
			},
		},
	})
}

func TestAccAppFabricIngestionDestination_destinationUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_appfabric_ingestion_destination.test"
	var ingestiondestination types.IngestionDestination

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIngestionDestinationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIngestionDestinationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIngestionDestinationExists(ctx, resourceName, &ingestiondestination),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.destination.0.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.destination.0.s3_bucket.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.destination.0.s3_bucket.0.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.destination.0.s3_bucket.0.bucket_name", "appfabric-test-audit-logs-658465413021"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.destination.0.s3_bucket.0.prefix", "AuditLog"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"processing_configuration"},
			},
			{
				Config: testAccIngestionDestinationConfig_destinationUpdate(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIngestionDestinationExists(ctx, resourceName, &ingestiondestination),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.destination.0.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.destination.0.firehose_stream.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.destination.0.firehose_stream.0.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.destination.0.firehose_stream.0.stream_name", "KDS-HTP-uMu4o"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"processing_configuration"},
			},
		},
	})
}

func TestAccAppFabricIngestionDestination_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var ingestiondestination types.IngestionDestination
	resourceName := "aws_appfabric_ingestion_destination.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIngestionDestinationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIngestionDestinationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIngestionDestinationExists(ctx, resourceName, &ingestiondestination),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfappfabric.ResourceIngestionDestination, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckIngestionDestinationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppFabricClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appfabric_ingestion_destination" {
				continue
			}

			_, err := tfappfabric.FindIngestionDestinationByID(ctx, conn, rs.Primary.Attributes[names.AttrARN], rs.Primary.Attributes["app_bundle_identifier"], rs.Primary.Attributes["ingestion_identifier"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Ingestion Destination %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckIngestionDestinationExists(ctx context.Context, name string, ingestiondestination *types.IngestionDestination) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.AppFabric, create.ErrActionCheckingExistence, tfappfabric.ResNameIngestionDestination, name, errors.New("not found"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppFabricClient(ctx)

		output, err := tfappfabric.FindIngestionDestinationByID(ctx, conn, rs.Primary.Attributes[names.AttrARN], rs.Primary.Attributes["app_bundle_identifier"], rs.Primary.Attributes["ingestion_identifier"])

		if err != nil {
			return err
		}

		*ingestiondestination = *output

		return nil
	}
}

func testAccIngestionDestinationConfig_basic() string {
	return `
resource "aws_appfabric_ingestion_destination" "test" {
  app_bundle_identifier             = "arn:aws:appfabric:us-east-1:825402724285:appbundle/04a7c9f5-14e1-4697-8b43-551186ff0a6e"
  ingestion_identifier              = "arn:aws:appfabric:us-east-1:825402724285:appbundle/04a7c9f5-14e1-4697-8b43-551186ff0a6e/ingestion/0379c487-af21-4419-80c2-92968007de1c"
  processing_configuration {
	audit_log {
		format = "json"
		schema = "raw"	
	}
  }
  destination_configuration {
    audit_log {
		destination {
			s3_bucket {
				bucket_name = "appfabric-test-audit-logs-658465413021"
				prefix = "AuditLog"
			}
		}
    }
  }
}
`
}

func testAccIngestionDestinationConfig_firehose() string {
	return `
resource "aws_appfabric_ingestion_destination" "test" {
  app_bundle_identifier             = "arn:aws:appfabric:us-east-1:825402724285:appbundle/04a7c9f5-14e1-4697-8b43-551186ff0a6e"
  ingestion_identifier              = "arn:aws:appfabric:us-east-1:825402724285:appbundle/04a7c9f5-14e1-4697-8b43-551186ff0a6e/ingestion/63a391c6-2165-453d-94ba-d8da0c59c8be"
  processing_configuration {
	audit_log {
		format = "json"
		schema = "ocsf"	
	}
  }
  destination_configuration {
    audit_log {
		destination {
			firehose_stream {
				stream_name = "KDS-HTP-uMu4o"
			}
		}
    }
  }
}
`
}

func testAccIngestionDestinationConfig_destinationUpdate() string {
	return `
resource "aws_appfabric_ingestion_destination" "test" {
	app_bundle_identifier             = "arn:aws:appfabric:us-east-1:825402724285:appbundle/04a7c9f5-14e1-4697-8b43-551186ff0a6e"
	ingestion_identifier              = "arn:aws:appfabric:us-east-1:825402724285:appbundle/04a7c9f5-14e1-4697-8b43-551186ff0a6e/ingestion/0379c487-af21-4419-80c2-92968007de1c"
	processing_configuration {
	audit_log {
		format = "json"
		schema = "ocsf"	
	}
	}
	destination_configuration {
	audit_log {
		destination {
			firehose_stream {
				stream_name = "KDS-HTP-uMu4o"
			}
		}
	}
	}
}
`
}
