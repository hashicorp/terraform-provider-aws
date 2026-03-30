// Copyright IBM Corp. 2014, 2026
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
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_datazone_project.test"
	domainName := "aws_datazone_domain.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProjectExists(ctx, t, resourceName, &project),
					resource.TestCheckResourceAttrPair(resourceName, "domain_identifier", domainName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "failure_reasons.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "glossary_terms.#", "0"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedAt),
					acctest.CheckResourceAttrRFC3339(resourceName, "last_updated_at"),
					// resource.TestCheckResourceAttr(resourceName, "project_status", string(types.ProjectStatusActive)),
					resource.TestCheckResourceAttr(resourceName, "skip_deletion_check", acctest.CtTrue),
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

func TestAccDataZoneProject_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var project datazone.GetProjectOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_datazone_project.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.DataZoneEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProjectExists(ctx, t, resourceName, &project),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfdatazone.ResourceProject, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDataZoneProject_description(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2 datazone.GetProjectOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_datazone_project.test"
	domainName := "aws_datazone_domain.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_description(rName, "desc"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProjectExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttrPair(resourceName, "domain_identifier", domainName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "glossary_terms.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "desc"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedAt),
					acctest.CheckResourceAttrRFC3339(resourceName, "last_updated_at"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       testAccAuthorizerImportStateIdFunc(resourceName),
				ImportStateVerifyIgnore: []string{"project_status", "skip_deletion_check"},
			},
			{
				Config: testAccProjectConfig_description(rName, "updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProjectExists(ctx, t, resourceName, &v2),
					testAccCheckProjectNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttrPair(resourceName, "domain_identifier", domainName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "glossary_terms.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "updated"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedAt),
					acctest.CheckResourceAttrRFC3339(resourceName, "last_updated_at"),
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

func testAccCheckProjectDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DataZoneClient(ctx)
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
			input := datazone.DeleteDomainInput{
				Identifier: aws.String(rs.Primary.Attributes["domain_identifier"]),
			}
			_, err := conn.DeleteDomain(ctx, &input)

			if err != nil {
				return create.Error(names.DataZone, create.ErrActionCheckingDestroyed, tfdatazone.ResNameProject, rs.Primary.ID, err)
			}
		}
		return nil
	}
}

func testAccCheckProjectExists(ctx context.Context, t *testing.T, name string, project *datazone.GetProjectOutput) resource.TestCheckFunc {
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
		domainIdentifier := rs.Primary.Attributes["domain_identifier"]
		conn := acctest.ProviderMeta(ctx, t).DataZoneClient(ctx)
		input := datazone.GetProjectInput{
			DomainIdentifier: &domainIdentifier,
			Identifier:       &rs.Primary.ID,
		}
		resp, err := conn.GetProject(ctx, &input)

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

func testAccAuthorizerImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s:%s", rs.Primary.Attributes["domain_identifier"], rs.Primary.ID), nil
	}
}

func testAccProjectConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccDomainConfig_basic(rName), fmt.Sprintf(`
resource "aws_datazone_project" "test" {
  domain_identifier   = aws_datazone_domain.test.id
  name                = %[1]q
  skip_deletion_check = true
}
`, rName))
}

func testAccProjectConfig_description(rName, description string) string {
	return acctest.ConfigCompose(testAccDomainConfig_basic(rName), fmt.Sprintf(`
resource "aws_datazone_project" "test" {
  domain_identifier   = aws_datazone_domain.test.id
  name                = %[1]q
  description         = %[2]q
  skip_deletion_check = true
}
`, rName, description))
}
