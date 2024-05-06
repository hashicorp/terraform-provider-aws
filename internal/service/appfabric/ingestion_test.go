// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appfabric_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appfabric"
	"github.com/aws/aws-sdk-go-v2/service/appfabric/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/names"

	tfappfabric "github.com/hashicorp/terraform-provider-aws/internal/service/appfabric"
)

func TestAccAppFabricIngestion_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var ingestion appfabric.GetIngestionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appfabric_ingestion.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIngestionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIngestionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIngestionExists(ctx, resourceName, &ingestion),
					resource.TestCheckResourceAttrSet(resourceName, "app"),
					resource.TestCheckResourceAttrSet(resourceName, "app_bundle_identifier"),
					resource.TestCheckResourceAttrSet(resourceName, "ingestion_type"),
					resource.TestCheckResourceAttrSet(resourceName, "tenant_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccIngestionImportStateIDFunc(ctx, resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppFabricIngestion_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var ingestion appfabric.GetIngestionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appfabric_ingestion.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIngestionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIngestionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIngestionExists(ctx, resourceName, &ingestion),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfappfabric.ResourceIngestion, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckIngestionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppFabricClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appfabric_ingestion" {
				continue
			}
			_, err := conn.GetIngestion(ctx, &appfabric.GetIngestionInput{
				AppBundleIdentifier: aws.String(rs.Primary.Attributes["app_bundle_identifier"]),
				IngestionIdentifier: aws.String(rs.Primary.Attributes["arn"]),
			})
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.AppFabric, create.ErrActionCheckingDestroyed, tfappfabric.ResNameIngestion, rs.Primary.ID, err)
			}
			return create.Error(names.AppFabric, create.ErrActionCheckingDestroyed, tfappfabric.ResNameIngestion, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckIngestionExists(ctx context.Context, name string, ingestion *appfabric.GetIngestionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.AppFabric, create.ErrActionCheckingExistence, tfappfabric.ResNameIngestion, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.AppFabric, create.ErrActionCheckingExistence, tfappfabric.ResNameIngestion, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppFabricClient(ctx)
		resp, err := conn.GetIngestion(ctx, &appfabric.GetIngestionInput{
			AppBundleIdentifier: aws.String(rs.Primary.Attributes["app_bundle_identifier"]),
			IngestionIdentifier: aws.String(rs.Primary.Attributes["arn"]),
		})

		if err != nil {
			return create.Error(names.AppFabric, create.ErrActionCheckingExistence, tfappfabric.ResNameIngestion, rs.Primary.ID, err)
		}

		*ingestion = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AppFabricClient(ctx)

	input := &appfabric.ListAppBundlesInput{}
	_, err := conn.ListAppBundles(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckIngestionNotRecreated(before, after *appfabric.GetIngestionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.Ingestion.Arn), aws.ToString(after.Ingestion.Arn); before != after {
			return create.Error(names.AppFabric, create.ErrActionCheckingNotRecreated, tfappfabric.ResNameIngestion, before, errors.New("recreated"))
		}

		return nil
	}
}

func testAccIngestionImportStateIDFunc(ctx context.Context, resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return "", errors.New("No Ingestion ID set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppFabricClient(ctx)
		appBundleIdentifier := rs.Primary.Attributes["app_bundle_identifier"]
		ingestionARN := rs.Primary.Attributes["arn"]

		_, err := conn.GetIngestion(ctx, &appfabric.GetIngestionInput{
			AppBundleIdentifier: aws.String(appBundleIdentifier),
			IngestionIdentifier: aws.String(ingestionARN),
		})

		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%s,%s", appBundleIdentifier, ingestionARN), nil
	}
}

func testAccIngestionConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_appfabric_ingestion" "test" {
	app = "OKTA"
	app_bundle_identifier = "arn:aws:appfabric:us-east-1:637423205184:appbundle/a9b91477-8831-43c0-970c-95bdc3b06633"
	tenant_id = "dev-22002358.okta.com"
	ingestion_type = "auditLog"
	tags = {
		Name = "AppFabricTesting"
	}
}
`)
}
