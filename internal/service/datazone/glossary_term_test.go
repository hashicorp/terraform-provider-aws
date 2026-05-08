// Copyright IBM Corp. 2014, 2026
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
	"github.com/aws/aws-sdk-go-v2/service/datazone/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfdatazone "github.com/hashicorp/terraform-provider-aws/internal/service/datazone"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDataZoneGlossaryTerm_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var glossaryterm datazone.GetGlossaryTermOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resourceName := "aws_datazone_glossary_term.test"
	glossaryName := "aws_datazone_glossary.test"
	glossarySecond := "aws_datazone_glossary_term.second"

	domainName := "aws_datazone_domain.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataZoneEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlossaryTermDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlossaryTermConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlossaryTermExists(ctx, t, resourceName, &glossaryterm),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "domain_identifier", domainName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "glossary_identifier", glossaryName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "long_description", "long_description"),
					resource.TestCheckResourceAttr(resourceName, "short_description", "short_desc"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "term_relations.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "term_relations.0.classifies.0", glossarySecond, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "term_relations.0.is_a.0", glossarySecond, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "user"},
				ImportStateIdFunc:       testAccAuthorizerGlossaryTermImportStateIdFunc(resourceName),
			},
		},
	})
}
func TestAccDataZoneGlossaryTerm_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var glossaryterm1, glossaryterm2 datazone.GetGlossaryTermOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resourceName := "aws_datazone_glossary_term.test"
	domainName := "aws_datazone_domain.test"
	glossaryName := "aws_datazone_glossary.test"
	glossarySecond := "aws_datazone_glossary_term.second"
	glossaryThird := "aws_datazone_glossary_term.third"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataZoneEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlossaryTermDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlossaryTermConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlossaryTermExists(ctx, t, resourceName, &glossaryterm1),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "domain_identifier", domainName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "glossary_identifier", glossaryName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "long_description", "long_description"),
					resource.TestCheckResourceAttr(resourceName, "short_description", "short_desc"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),
					resource.TestCheckResourceAttr(resourceName, "term_relations.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "term_relations.0.classifies.0", glossarySecond, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "term_relations.0.is_a.0", glossarySecond, names.AttrID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "user"},
				ImportStateIdFunc:       testAccAuthorizerGlossaryTermImportStateIdFunc(resourceName),
			},
			{
				Config: testAccGlossaryTermConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlossaryTermExists(ctx, t, resourceName, &glossaryterm2),
					testAccCheckGlossaryTermNotRecreated(&glossaryterm1, &glossaryterm2),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "glossary_identifier"),
					resource.TestCheckResourceAttr(resourceName, "long_description", "long"),
					resource.TestCheckResourceAttr(resourceName, "short_description", "short"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),
					resource.TestCheckResourceAttr(resourceName, "term_relations.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "term_relations.0.classifies.0", glossaryThird, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "term_relations.0.is_a.0", glossaryThird, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAuthorizerGlossaryTermImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccDataZoneGlossaryTerm_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var glossaryterm datazone.GetGlossaryTermOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_datazone_glossary_term.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataZoneEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlossaryTermDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlossaryTermConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlossaryTermExists(ctx, t, resourceName, &glossaryterm),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfdatazone.ResourceGlossaryTerm, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckGlossaryTermDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DataZoneClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_datazone_glossary_term" {
				continue
			}
			input := datazone.GetGlossaryTermInput{
				Identifier:       aws.String(rs.Primary.ID),
				DomainIdentifier: aws.String(rs.Primary.Attributes["domain_identifier"]),
			}
			_, err := conn.GetGlossaryTerm(ctx, &input)

			if errs.IsA[*types.ResourceNotFoundException](err) || errs.IsA[*types.AccessDeniedException](err) {
				continue
			}

			if err != nil {
				return create.Error(names.DataZone, create.ErrActionCheckingDestroyed, tfdatazone.ResNameGlossaryTerm, rs.Primary.ID, err)
			}

			return create.Error(names.DataZone, create.ErrActionCheckingDestroyed, tfdatazone.ResNameGlossaryTerm, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccAuthorizerGlossaryTermImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return strings.Join([]string{rs.Primary.Attributes["domain_identifier"], rs.Primary.ID, rs.Primary.Attributes["glossary_identifier"]}, ","), nil
	}
}

func testAccCheckGlossaryTermExists(ctx context.Context, t *testing.T, name string, glossaryterm *datazone.GetGlossaryTermOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.ResNameGlossaryTerm, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.ResNameGlossaryTerm, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).DataZoneClient(ctx)
		resp, err := tfdatazone.FindGlossaryTermByID(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["domain_identifier"])
		if err != nil {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.ResNameGlossaryTerm, rs.Primary.ID, err)
		}

		*glossaryterm = *resp

		return nil
	}
}

func testAccCheckGlossaryTermNotRecreated(before, after *datazone.GetGlossaryTermOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.Id), aws.ToString(after.Id); before != after {
			return create.Error(names.DataZone, create.ErrActionCheckingNotRecreated, tfdatazone.ResNameGlossaryTerm, before, errors.New("recreated"))
		}

		return nil
	}
}

func testAccGlossaryTermConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccGlossaryConfig_basic(rName), fmt.Sprintf(`
resource "aws_datazone_glossary_term" "test" {
  domain_identifier   = aws_datazone_domain.test.id
  glossary_identifier = aws_datazone_glossary.test.id
  long_description    = "long_description"
  name                = %[1]q
  short_description   = "short_desc"
  status              = "ENABLED"
  term_relations {
    classifies = [aws_datazone_glossary_term.second.id]
    is_a       = [aws_datazone_glossary_term.second.id]
  }
}

resource "aws_datazone_glossary_term" "second" {
  domain_identifier   = aws_datazone_domain.test.id
  glossary_identifier = aws_datazone_glossary.test.id
  long_description    = "long_description"
  name                = "%[1]s-2"
  short_description   = "short_desc"
  status              = "ENABLED"
}
`, rName))
}

func testAccGlossaryTermConfig_update(rName string) string {
	return acctest.ConfigCompose(testAccGlossaryConfig_basic(rName), fmt.Sprintf(`
resource "aws_datazone_glossary_term" "test" {
  domain_identifier   = aws_datazone_domain.test.id
  glossary_identifier = aws_datazone_glossary.test.id
  long_description    = "long"
  name                = %[1]q
  short_description   = "short"
  status              = "ENABLED"
  term_relations {
    classifies = [aws_datazone_glossary_term.third.id]
    is_a       = [aws_datazone_glossary_term.third.id]
  }
}

resource "aws_datazone_glossary_term" "second" {
  domain_identifier   = aws_datazone_domain.test.id
  glossary_identifier = aws_datazone_glossary.test.id
  long_description    = "long_description"
  name                = "%[1]s-2"
  short_description   = "short_desc"
  status              = "ENABLED"
}

resource "aws_datazone_glossary_term" "third" {
  domain_identifier   = aws_datazone_domain.test.id
  glossary_identifier = aws_datazone_glossary.test.id
  long_description    = "long_description"
  name                = "%[1]s-3"
  short_description   = "short_desc"
  status              = "ENABLED"
}
`, rName))
}
