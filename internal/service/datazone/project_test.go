// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datazone_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datazone"
	"github.com/aws/aws-sdk-go-v2/service/datazone/types"
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

func TestAccDataZoneProject_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var project datazone.GetProjectOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datazone_project.test"
	domainName := "aws_datazone_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_basic(rName, dName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttrPair(resourceName, "domain_identifier", domainName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "glossary_terms.#"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "desc"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_at"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       testAccAuthorizerImportStateIdFunc(resourceName),
				ImportStateVerifyIgnore: []string{"skip_deletion_check", "project_status"},
			},
		},
	})
}
func TestAccDataZoneProject_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var project datazone.GetProjectOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datazone_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.DataZoneEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_basic(rName, dName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfdatazone.ResourceProject, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
func testAccCheckProjectDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DataZoneClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_datazone_project" {
				continue
			}
			t := rs.Primary.Attributes["domain_identifier"]

			input := &datazone.GetProjectInput{
				DomainIdentifier: &t,
				Identifier:       aws.String(rs.Primary.ID),
			}
			_, err := conn.GetProject(ctx, input)
			if errs.IsA[*types.AccessDeniedException](err) || errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.DataZone, create.ErrActionCheckingDestroyed, tfdatazone.ResNameProject, rs.Primary.ID, err)
			}
		}
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_datazone_domain" {
				continue
			}
			_, err := conn.DeleteDomain(ctx, &datazone.DeleteDomainInput{
				Identifier: aws.String(rs.Primary.Attributes["domain_identifier"]),
			})

			if err != nil {
				return create.Error(names.DataZone, create.ErrActionCheckingDestroyed, tfdatazone.ResNameDomain, rs.Primary.ID, err)
			}
		}
		return nil
	}
}

func testAccCheckProjectExists(ctx context.Context, name string, project *datazone.GetProjectOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.ResNameProject, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.ResNameProject, name, errors.New("not set"))
		}
		if rs.Primary.Attributes["domain_identifier"] == "" {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.ResNameProject, name, errors.New("domain identifier not set"))
		}
		t := rs.Primary.Attributes["domain_identifier"]
		conn := acctest.Provider.Meta().(*conns.AWSClient).DataZoneClient(ctx)
		resp, err := conn.GetProject(ctx, &datazone.GetProjectInput{
			DomainIdentifier: &t,
			Identifier:       &rs.Primary.ID,
		})

		if err != nil && !errs.IsA[*types.ResourceNotFoundException](err) {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.ResNameProject, rs.Primary.ID, err)
		}

		*project = *resp

		return nil
	}
}

func testAccCheckProjectNotRecreated(before, after *datazone.GetProjectOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.Id), aws.ToString(after.Id); before != after {
			return create.Error(names.DataZone, create.ErrActionCheckingNotRecreated, tfdatazone.ResNameProject, before, errors.New("recreated"))
		}

		return nil
	}
}
func TestAccDataZoneProject_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2 datazone.GetProjectOutput
	pName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datazone_project.test"
	domainName := "aws_datazone_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_basic(pName, dName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttrPair(resourceName, "domain_identifier", domainName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "glossary_terms.#"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "desc"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, pName),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_at"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       testAccAuthorizerImportStateIdFunc(resourceName),
				ImportStateVerifyIgnore: []string{"skip_deletion_check", "project_status"},
			},
			{
				Config: testAccProjectConfigBasicUpdate(pName, dName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &v2),
					testAccCheckProjectNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttrPair(resourceName, "domain_identifier", domainName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "glossary_terms.#"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, pName),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_at"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       testAccAuthorizerImportStateIdFunc(resourceName),
				ImportStateVerifyIgnore: []string{"project_status", "skip_deletion_check"},
			},
		},
	})
}

func testAccAuthorizerImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s:%s", rs.Primary.Attributes["domain_identifier"], rs.Primary.ID), nil
	}
}

func testAccProjectConfig_basic(pName, dName string) string {
	return acctest.ConfigCompose(testAccDomainConfig_basic(dName), fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_datazone_project" "test" {
  domain_identifier   = aws_datazone_domain.test.id
  glossary_terms      = ["2N8w6XJCwZf"]
  name                = %[1]q
  description         = "desc"
  skip_deletion_check = true
}
`, pName))
}
func testAccProjectConfigBasicUpdate(pName, dName string) string {
	return acctest.ConfigCompose(testAccDomainConfig_basic(dName), fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_datazone_project" "test" {
  domain_identifier   = aws_datazone_domain.test.id
  glossary_terms      = ["2N8w6XJCwZf"]
  name                = %[1]q
  description         = "description"
  skip_deletion_check = true
}
`, pName))
}
