// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalog"
	awstypes "github.com/aws/aws-sdk-go-v2/service/servicecatalog/types"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_servicecatalog_budget_resource_association", &resource.Sweeper{
		Name:         "aws_servicecatalog_budget_resource_association",
		Dependencies: []string{},
		F:            sweepBudgetResourceAssociations,
	})

	resource.AddTestSweepers("aws_servicecatalog_constraint", &resource.Sweeper{
		Name:         "aws_servicecatalog_constraint",
		Dependencies: []string{},
		F:            sweepConstraints,
	})

	resource.AddTestSweepers("aws_servicecatalog_principal_portfolio_association", &resource.Sweeper{
		Name:         "aws_servicecatalog_principal_portfolio_association",
		Dependencies: []string{},
		F:            sweepPrincipalPortfolioAssociations,
	})

	resource.AddTestSweepers("aws_servicecatalog_product_portfolio_association", &resource.Sweeper{
		Name:         "aws_servicecatalog_product_portfolio_association",
		Dependencies: []string{},
		F:            sweepProductPortfolioAssociations,
	})

	resource.AddTestSweepers("aws_servicecatalog_product", &resource.Sweeper{
		Name: "aws_servicecatalog_product",
		Dependencies: []string{
			"aws_servicecatalog_provisioning_artifact",
		},
		F: sweepProducts,
	})

	resource.AddTestSweepers("aws_servicecatalog_provisioned_product", &resource.Sweeper{
		Name:         "aws_servicecatalog_provisioned_product",
		Dependencies: []string{},
		F:            sweepProvisionedProducts,
	})

	resource.AddTestSweepers("aws_servicecatalog_provisioning_artifact", &resource.Sweeper{
		Name:         "aws_servicecatalog_provisioning_artifact",
		Dependencies: []string{},
		F:            sweepProvisioningArtifacts,
	})

	resource.AddTestSweepers("aws_servicecatalog_service_action", &resource.Sweeper{
		Name:         "aws_servicecatalog_service_action",
		Dependencies: []string{},
		F:            sweepServiceActions,
	})

	resource.AddTestSweepers("aws_servicecatalog_tag_option_resource_association", &resource.Sweeper{
		Name:         "aws_servicecatalog_tag_option_resource_association",
		Dependencies: []string{},
		F:            sweepTagOptionResourceAssociations,
	})

	resource.AddTestSweepers("aws_servicecatalog_tag_option", &resource.Sweeper{
		Name:         "aws_servicecatalog_tag_option",
		Dependencies: []string{},
		F:            sweepTagOptions,
	})
}

func sweepBudgetResourceAssociations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.ServiceCatalogClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	pages := servicecatalog.NewListPortfoliosPaginator(conn, &servicecatalog.ListPortfoliosInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error describing Service Catalog Portfolio for %s: %w", region, err))
		}

		for _, port := range page.PortfolioDetails {
			input := &servicecatalog.ListBudgetsForResourceInput{
				ResourceId: port.Id,
			}

			pages := servicecatalog.NewListBudgetsForResourcePaginator(conn, input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if err != nil {
					errs = multierror.Append(errs, fmt.Errorf("error describing Service Catalog Portfolio Budget for %s: %w", region, err))
				}

				for _, budget := range page.Budgets {
					r := resourceBudgetResourceAssociation()
					d := r.Data(nil)
					d.SetId(budgetResourceAssociationID(aws.ToString(budget.BudgetName), aws.ToString(port.Id)))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}
			}
		}
	}

	pagesProducts := servicecatalog.NewSearchProductsAsAdminPaginator(conn, &servicecatalog.SearchProductsAsAdminInput{})
	for pagesProducts.HasMorePages() {
		page, err := pagesProducts.NextPage(ctx)

		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error describing Service Catalog Products for %s: %w", region, err))
		}

		for _, pvd := range page.ProductViewDetails {
			if pvd.ProductViewSummary == nil {
				continue
			}

			input := &servicecatalog.ListBudgetsForResourceInput{
				ResourceId: pvd.ProductViewSummary.ProductId,
			}

			pages := servicecatalog.NewListBudgetsForResourcePaginator(conn, input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if err != nil {
					errs = multierror.Append(errs, fmt.Errorf("error describing Service Catalog Budget Resource (Product) Associations for %s: %w", region, err))
				}

				for _, budget := range page.Budgets {
					r := resourceBudgetResourceAssociation()
					d := r.Data(nil)
					d.SetId(budgetResourceAssociationID(aws.ToString(budget.BudgetName), aws.ToString(pvd.ProductViewSummary.ProductId)))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}
			}
		}
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Service Catalog Budget Resource Associations for %s: %w", region, err))
	}

	if awsv2.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Service Catalog Budget Resource Associations sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepConstraints(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.ServiceCatalogClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	// no paginator or list operation for constraints directly, have to list portfolios and constraints of portfolios
	pages := servicecatalog.NewListPortfoliosPaginator(conn, &servicecatalog.ListPortfoliosInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error describing Service Catalog Portfolios for %s: %w", region, err))
		}

		for _, detail := range page.PortfolioDetails {
			input := &servicecatalog.ListConstraintsForPortfolioInput{
				PortfolioId: detail.Id,
			}

			pages := servicecatalog.NewListConstraintsForPortfolioPaginator(conn, input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if err != nil {
					errs = multierror.Append(errs, fmt.Errorf("error describing Service Catalog Constraints for %s: %w", region, err))
				}

				for _, detail := range page.ConstraintDetails {
					r := resourceConstraint()
					d := r.Data(nil)
					d.SetId(aws.ToString(detail.ConstraintId))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}
			}
		}
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Service Catalog Constraints for %s: %w", region, err))
	}

	if awsv2.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Service Catalog Constraints sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepPrincipalPortfolioAssociations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.ServiceCatalogClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	pages := servicecatalog.NewListPortfoliosPaginator(conn, &servicecatalog.ListPortfoliosInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error listing Service Catalog Principal Portfolios for %s: %w", region, err))
		}

		for _, detail := range page.PortfolioDetails {
			input := &servicecatalog.ListPrincipalsForPortfolioInput{
				PortfolioId: detail.Id,
			}

			pages := servicecatalog.NewListPrincipalsForPortfolioPaginator(conn, input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if err != nil {
					errs = multierror.Append(errs, fmt.Errorf("error listing Service Catalog Portfolios for Principals %s: %w", region, err))
					continue
				}

				for _, principal := range page.Principals {
					r := resourcePrincipalPortfolioAssociation()
					d := r.Data(nil)
					d.SetId(principalPortfolioAssociationCreateResourceID(acceptLanguageEnglish, aws.ToString(principal.PrincipalARN), aws.ToString(detail.Id), string(principal.PrincipalType)))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}
			}
		}
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Service Catalog Principal Portfolio Associations for %s: %w", region, err))
	}

	if awsv2.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Service Catalog Principal Portfolio Associations sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepProductPortfolioAssociations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.ServiceCatalogClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	// no paginator or list operation for associations directly, have to list products and associations of products
	pages := servicecatalog.NewSearchProductsAsAdminPaginator(conn, &servicecatalog.SearchProductsAsAdminInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error describing Service Catalog Products for %s: %w", region, err))
		}

		for _, detail := range page.ProductViewDetails {
			productARN, err := arn.Parse(aws.ToString(detail.ProductARN))

			if err != nil {
				errs = multierror.Append(errs, fmt.Errorf("error parsing Product ARN for %s: %w", aws.ToString(detail.ProductARN), err))
				continue
			}

			// arn:aws:catalog:us-west-2:187416307283:product/prod-t5thhvquxw2x2
			resourceParts := strings.SplitN(productARN.Resource, "/", 2)

			if len(resourceParts) != 2 {
				errs = multierror.Append(errs, fmt.Errorf("error parsing Product ARN resource for %s: %w", aws.ToString(detail.ProductARN), err))
				continue
			}

			productID := resourceParts[1]

			input := &servicecatalog.ListPortfoliosForProductInput{
				ProductId: aws.String(productID),
			}

			pages := servicecatalog.NewListPortfoliosForProductPaginator(conn, input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if err != nil {
					errs = multierror.Append(errs, fmt.Errorf("error listing Service Catalog Portfolios for Products %s: %w", region, err))
					continue
				}

				for _, detail := range page.PortfolioDetails {
					r := resourceProductPortfolioAssociation()
					d := r.Data(nil)
					d.SetId(productPortfolioAssociationCreateID(acceptLanguageEnglish, aws.ToString(detail.Id), productID))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}
			}
		}
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Service Catalog Product Portfolio Associations for %s: %w", region, err))
	}

	if awsv2.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Service Catalog Product Portfolio Associations sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepProducts(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.ServiceCatalogClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	pages := servicecatalog.NewSearchProductsAsAdminPaginator(conn, &servicecatalog.SearchProductsAsAdminInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error describing Service Catalog Products for %s: %w", region, err))
		}

		for _, pvd := range page.ProductViewDetails {
			if pvd.ProductViewSummary == nil {
				continue
			}

			id := aws.ToString(pvd.ProductViewSummary.ProductId)

			r := resourceProduct()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Service Catalog Products for %s: %w", region, err))
	}

	if awsv2.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Service Catalog Products sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepProvisionedProducts(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.ServiceCatalogClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &servicecatalog.SearchProvisionedProductsInput{
		AccessLevelFilter: &awstypes.AccessLevelFilter{
			Key:   awstypes.AccessLevelFilterKeyAccount,
			Value: aws.String("self"), // only supported value
		},
	}

	pages := servicecatalog.NewSearchProvisionedProductsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error describing Service Catalog Provisioned Products for %s: %w", region, err))
		}

		for _, detail := range page.ProvisionedProducts {
			r := resourceProvisionedProduct()
			d := r.Data(nil)
			d.SetId(aws.ToString(detail.Id))
			d.Set("ignore_errors", true)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Service Catalog Provisioned Products for %s: %w", region, err))
	}

	if awsv2.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Service Catalog Provisioned Products sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepProvisioningArtifacts(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.ServiceCatalogClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	pages := servicecatalog.NewSearchProductsAsAdminPaginator(conn, &servicecatalog.SearchProductsAsAdminInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error describing Service Catalog Products for %s: %w", region, err))
		}

		for _, pvd := range page.ProductViewDetails {
			if pvd.ProductViewSummary == nil || pvd.ProductViewSummary.ProductId == nil {
				continue
			}

			productID := aws.ToString(pvd.ProductViewSummary.ProductId)
			input := &servicecatalog.ListProvisioningArtifactsInput{
				ProductId: aws.String(productID),
			}

			// there's no paginator for ListProvisioningArtifacts
			output, err := conn.ListProvisioningArtifacts(ctx, input)
			if err != nil {
				errs = multierror.Append(errs, fmt.Errorf("error listing Service Catalog Provisioning Artifacts for product (%s): %w", productID, err))
				break
			}

			for _, pad := range output.ProvisioningArtifactDetails {
				r := resourceProvisioningArtifact()
				d := r.Data(nil)

				d.SetId(provisioningArtifactID(aws.ToString(pad.Id), productID))

				sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
			}
		}
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Service Catalog Provisioning Artifacts for %s: %w", region, err))
	}

	if awsv2.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Service Catalog Provisioning Artifacts sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepServiceActions(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.ServiceCatalogClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	pages := servicecatalog.NewListServiceActionsPaginator(conn, &servicecatalog.ListServiceActionsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error describing Service Catalog Service Actions for %s: %w", region, err))
		}

		for _, sas := range page.ServiceActionSummaries {
			id := aws.ToString(sas.Id)

			r := resourceServiceAction()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Service Catalog Service Actions for %s: %w", region, err))
	}

	if awsv2.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Service Catalog Service Actions sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepTagOptionResourceAssociations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.ServiceCatalogClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	pages := servicecatalog.NewListTagOptionsPaginator(conn, &servicecatalog.ListTagOptionsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.TagOptionNotMigratedException](err) {
			log.Printf("[WARN] Skipping Service Catalog Tag Option Resource Associations sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error describing Service Catalog Tag Options for %s: %w", region, err))
		}

		for _, tag := range page.TagOptionDetails {
			input := &servicecatalog.ListResourcesForTagOptionInput{
				TagOptionId: tag.Id,
			}

			pages := servicecatalog.NewListResourcesForTagOptionPaginator(conn, input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if errs.IsA[*awstypes.TagOptionNotMigratedException](err) {
					log.Printf("[WARN] Skipping Service Catalog Tag Option Resource Associations sweep for %s: %s", region, err)
					return nil
				}

				if err != nil {
					sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error describing Service Catalog Tag Option Resource Associations for %s: %w", region, err))
				}

				for _, resource := range page.ResourceDetails {
					r := resourceTagOptionResourceAssociation()
					d := r.Data(nil)
					d.SetId(aws.ToString(resource.Id))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}
			}
		}
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Service Catalog Tag Option Resource Associations for %s: %w", region, err))
	}

	if awsv2.SkipSweepError(sweeperErrs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Service Catalog Tag Option Resource Associations sweep for %s: %s", region, sweeperErrs)
		return nil
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepTagOptions(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.ServiceCatalogClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	pages := servicecatalog.NewListTagOptionsPaginator(conn, &servicecatalog.ListTagOptionsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.TagOptionNotMigratedException](err) {
			log.Printf("[WARN] Skipping Service Catalog Tag Options sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error describing Service Catalog Tag Options for %s: %w", region, err))
		}

		for _, tod := range page.TagOptionDetails {
			id := aws.ToString(tod.Id)

			r := resourceTagOption()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Service Catalog Tag Options for %s: %w", region, err))
	}

	if awsv2.SkipSweepError(sweeperErrs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Service Catalog Tag Options sweep for %s: %s", region, sweeperErrs)
		return nil
	}

	return sweeperErrs.ErrorOrNil()
}
