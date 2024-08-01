// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datazone_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datazone"
	awstypes "github.com/aws/aws-sdk-go-v2/service/datazone/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfdatazone "github.com/hashicorp/terraform-provider-aws/internal/service/datazone"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDataZoneGlossary_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var glossary datazone.GetGlossaryOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	token := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	pName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_datazone_glossary.test"
	projectName := "aws_datazone_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataZoneEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlossaryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlossaryConfig_basic(rName, token, pName, dName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlossaryExists(ctx, resourceName, &glossary),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "owning_project_identifier", projectName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "domain_identifier", projectName, "domain_identifier"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAuthorizerGlossaryImportStateIdFunc(resourceName),
			},
		},
	})
}
func TestAccDataZoneGlossary_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var glossary, glossary2 datazone.GetGlossaryOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	token := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	pName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_datazone_glossary.test"
	projectName := "aws_datazone_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataZoneEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlossaryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlossaryConfig_basic(rName, token, pName, dName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlossaryExists(ctx, resourceName, &glossary),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "desc"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "owning_project_identifier", projectName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ENABLED"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_identifier", projectName, "domain_identifier"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAuthorizerGlossaryImportStateIdFunc(resourceName),
			},
			{
				Config: testAccGlossaryConfig_update(rName, token, pName, dName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlossaryExists(ctx, resourceName, &glossary2),
					testAccCheckGlossaryNotRecreated(&glossary, &glossary2),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "owning_project_identifier", projectName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "DISABLED"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_identifier", projectName, "domain_identifier"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAuthorizerGlossaryImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccDataZoneGlossary_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var glossary datazone.GetGlossaryOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	token := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	pName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datazone_glossary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataZoneEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlossaryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlossaryConfig_basic(rName, token, pName, dName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlossaryExists(ctx, resourceName, &glossary),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfdatazone.ResourceGlossary, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckGlossaryDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DataZoneClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_datazone_glossary" {
				continue
			}
			_, err := tfdatazone.FindGlossaryByID(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["domain_identifier"])
			if errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsA[*awstypes.AccessDeniedException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.DataZone, create.ErrActionCheckingDestroyed, tfdatazone.ResNameGlossary, rs.Primary.ID, err)
			}

			return create.Error(names.DataZone, create.ErrActionCheckingDestroyed, tfdatazone.ResNameGlossary, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckGlossaryExists(ctx context.Context, name string, glossary *datazone.GetGlossaryOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.ResNameGlossary, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.ResNameGlossary, name, errors.New("not set"))
		}
		if rs.Primary.Attributes["domain_identifier"] == "" {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.ResNameProject, name, errors.New("domain identifier not set"))
		}
		conn := acctest.Provider.Meta().(*conns.AWSClient).DataZoneClient(ctx)
		resp, err := tfdatazone.FindGlossaryByID(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["domain_identifier"])

		if err != nil {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.ResNameGlossary, rs.Primary.ID, err)
		}

		*glossary = *resp

		return nil
	}
}

func testAccCheckGlossaryNotRecreated(before, after *datazone.GetGlossaryOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.Id), aws.ToString(after.Id); before != after {
			return create.Error(names.DataZone, create.ErrActionCheckingNotRecreated, tfdatazone.ResNameGlossary, before, errors.New("recreated"))
		}

		return nil
	}
}

func testAccAuthorizerGlossaryImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return strings.Join([]string{rs.Primary.Attributes["domain_identifier"], rs.Primary.ID, rs.Primary.Attributes["owning_project_identifier"]}, ","), nil
	}
}

func testAccGlossaryConfig_basic(rName, token, dName, pName string) string {
	return acctest.ConfigCompose(testAccProjectConfig_basic(pName, dName), fmt.Sprintf(`
resource "aws_datazone_glossary" "test" {
  description               = "desc"
  name                      = %[1]q
  owning_project_identifier = aws_datazone_project.test.id
  status                    = "ENABLED"
  domain_identifier         = aws_datazone_project.test.domain_identifier
}
`, rName, token))
}

func testAccGlossaryConfig_update(rName, token, dName, pName string) string {
	return acctest.ConfigCompose(testAccProjectConfig_basic(pName, dName), fmt.Sprintf(`
resource "aws_datazone_glossary" "test" {
  description               = "description"
  name                      = %[1]q
  owning_project_identifier = aws_datazone_project.test.id
  status                    = "DISABLED"
  domain_identifier         = aws_datazone_project.test.domain_identifier
}
`, rName, token))
}
