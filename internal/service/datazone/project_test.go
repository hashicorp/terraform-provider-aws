// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datazone_test

// **PLEASE DELETE THIS AND ALL TIP COMMENTS BEFORE SUBMITTING A PR FOR REVIEW!**
//
// TIP: ==== INTRODUCTION ====
// Thank you for trying the skaff tool!
//
// You have opted to include these helpful comments. They all include "TIP:"
// to help you find and remove them when you're done with them.
//
// While some aspects of this file are customized to your input, the
// scaffold tool does *not* look at the AWS API and ensure it has correct
// function, structure, and variable names. It makes guesses based on
// commonalities. You will need to make significant adjustments.
//
// In other words, as generated, this is a rough outline of the work you will
// need to do. If something doesn't make sense for your situation, get rid of
// it.

import (
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	//
	// The provider linter wants your imports to be in two groups: first,
	// standard library (i.e., "fmt" or "strings"), second, everything else.
	//
	// Also, AWS Go SDK v2 may handle nested structures differently than v1,
	// using the services/datazone/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// types.<Type Name>.
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
	"github.com/hashicorp/terraform-provider-aws/names"

	// TIP: You will often need to import the package that this test file lives
	// in. Since it is in the "test" context, it must import the package to use
	// any normal context constants, variables, or functions.
	tfdatazone "github.com/hashicorp/terraform-provider-aws/internal/service/datazone"
)

func TestAccDataZoneProject_basic(t *testing.T) {
	ctx := acctest.Context(t)
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var project datazone.GetProjectOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix) // Name for project
	dName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix) // Name for domain
	resourceName := "aws_datazone_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			//acctest.PreCheckPartitionHasService(t, names.DataZoneEndpointID)
			//testAccPreCheckProject(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_basic(rName, dName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					// inputed
					resource.TestCheckResourceAttrSet(resourceName, "domain_id"), // find a domain identifier
					//resource.TestCheckResourceAttrSet(resourceName, "glossary_terms.#"),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "skip_deletion_check"),
					// computed
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					//resource.TestCheckResourceAttrSet(resourceName, "failure_reasons.#"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_at"),
					// should this ever be empty?? resource.TestCheckResourceAttrSet(resourceName, "project_status"),
					//resource.TestCheckResourceAttrSet(resourceName, "result_metadata"),
				),
			},
			{
				ResourceName:       resourceName,
				ImportState:        true,
				ImportStateVerify:  true,
				ExpectNonEmptyPlan: true,

				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
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
	dName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix) // Name for domain
	resourceName := "aws_datazone_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AttrDomainName)
			testAccPreCheckProject(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_basic(rName, dName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					// TIP: The Plugin-Framework disappears helper is similar to the Plugin-SDK version,
					// but expects a new resource factory function as the third argument. To expose this
					// private function to the testing package, you may need to add a line like the following
					// to exports_test.go:
					//
					//   var ResourceProject = newResourceProject
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
			t := rs.Primary.Attributes["domain_id"]

			input := &datazone.GetProjectInput{
				DomainIdentifier: &t,
				Identifier:       aws.String(rs.Primary.ID),
			}
			_, err := conn.GetProject(ctx, input)
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.DataZone, create.ErrActionCheckingDestroyed, tfdatazone.ResNameProject, rs.Primary.ID, err)
			}

			return create.Error(names.DataZone, create.ErrActionCheckingDestroyed, tfdatazone.ResNameProject, rs.Primary.ID, errors.New("not destroyed"))
		}

		return testAccCheckProjectDomainDestroy(ctx, s)
	}
}
func testAccCheckProjectDomainDestroy(ctx context.Context, s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DataZoneClient(ctx)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_datazone_domain" {
			continue
		}
		in := &datazone.GetDomainInput{
			Identifier: aws.String(rs.Primary.ID),
		}
		del := &datazone.DeleteDomainInput{
			Identifier: aws.String(rs.Primary.ID),
		}
		_, err := conn.DeleteDomain(ctx, del)

		if err != nil {
			return create.Error(names.DataZone, create.ErrActionCheckingDestroyed, tfdatazone.ResNameDomain, rs.Primary.ID, err)

		}

		_, err = conn.GetDomain(ctx, in)

		if tfdatazone.IsResourceMissing(err) {
			return nil
		}

		if err != nil {
			return create.Error(names.DataZone, create.ErrActionCheckingDestroyed, tfdatazone.ResNameDomain, rs.Primary.ID, err)
		}

		return create.Error(names.DataZone, create.ErrActionCheckingDestroyed, tfdatazone.ResNameDomain, rs.Primary.ID, errors.New("not destroyed"))
	}

	return nil
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
		if rs.Primary.Attributes["domain_id"] == "" {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.ResNameProject, name, errors.New("domain identifer not set"))

		}
		t := rs.Primary.Attributes["domain_id"]

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataZoneClient(ctx)
		resp, err := conn.GetProject(ctx, &datazone.GetProjectInput{
			DomainIdentifier: &t,
			Identifier:       &rs.Primary.ID,
		})

		if err != nil {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.ResNameProject, rs.Primary.ID, err)
		}

		*project = *resp

		return nil
	}
}
func testAccPreCheckProject(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DataZoneClient(ctx)

	input := &datazone.ListProjectsInput{}
	_, err := conn.ListProjects(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
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
func TestAccCheckProjectUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var project datazone.GetProjectOutput
	pName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix) // Name for datazone project
	dName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix) // Name for datazone domain
	resourceName := "aws_datazone_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			//acctest.PreCheckPartitionHasService(t, names.DataZoneEndpointID)
			//testAccPreCheckProject(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_basic(pName, dName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					// inputed
					resource.TestCheckResourceAttrSet(resourceName, "domain_id"), // find a domain identifier
					//resource.TestCheckResourceAttrSet(resourceName, "glossary_terms.#"),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "skip_deletion_check"),
					// computed
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					//resource.TestCheckResourceAttrSet(resourceName, "failure_reasons.#"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_at"),
					// should this ever be empty?? resource.TestCheckResourceAttrSet(resourceName, "project_status"),
					//resource.TestCheckResourceAttrSet(resourceName, "result_metadata"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				//ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
			{
				Config: testAccProjectConfig_basic(pName, dName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					// inputed
					resource.TestCheckResourceAttrSet(resourceName, "domain_id"), // find a domain identifier
					//resource.TestCheckResourceAttrSet(resourceName, "glossary_terms.#"),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "skip_deletion_check"),
					// computed
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					//resource.TestCheckResourceAttrSet(resourceName, "failure_reasons.#"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_at"),
					// should this ever be empty?? resource.TestCheckResourceAttrSet(resourceName, "project_status"),
					//resource.TestCheckResourceAttrSet(resourceName, "result_metadata"),
				),
			},
		},
	})

}

func testAccProjectConfig_basic(pName, dName string) string {
	return acctest.ConfigCompose(testAccDomainConfig_basic(dName), fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_datazone_project" "test" {  
  domain_id = aws_datazone_domain.test.id
  name              = %[1]q
  description       = "desc"
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
  domain_id = aws_datazone_domain.test.id
  name              = %[1]q
  description       = "description"
  skip_deletion_check = true
}
`, pName))
}

//   glossary_terms    = ["2N8w6XJCwZf"]
