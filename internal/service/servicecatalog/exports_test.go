// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog

// Exports for use in tests only.
var (
	ResourceBudgetResourceAssociation     = resourceBudgetResourceAssociation
	ResourceConstraint                    = resourceConstraint
	ResourcePortfolio                     = resourcePortfolio
	ResourcePortfolioShare                = resourcePortfolioShare
	ResourceProduct                       = resourceProduct
	ResourceProductPortfolioAssociation   = resourceProductPortfolioAssociation
	ResourceProvisionedProduct            = resourceProvisionedProduct
	ResourceProvisioningArtifact          = resourceProvisioningArtifact
	ResourcePrincipalPortfolioAssociation = resourcePrincipalPortfolioAssociation
	ResourceServiceAction                 = resourceServiceAction
	ResourceTagOption                     = resourceTagOption
	ResourceTagOptionResourceAssociation  = resourceTagOptionResourceAssociation

	FindPortfolioByID                 = findPortfolioByID
	FindPortfolioShare                = findPortfolioShare
	FindPrincipalPortfolioAssociation = findPrincipalPortfolioAssociation

	BudgetResourceAssociationParseID             = budgetResourceAssociationParseID
	ProductPortfolioAssociationParseID           = productPortfolioAssociationParseID
	ProvisioningArtifactParseID                  = provisioningArtifactParseID
	PrincipalPortfolioAssociationParseResourceID = principalPortfolioAssociationParseResourceID
	TagOptionResourceAssociationParseID          = tagOptionResourceAssociationParseID

	AcceptLanguageEnglish = acceptLanguageEnglish
	StatusCreated         = statusCreated

	WaitBudgetResourceAssociationDeleted    = waitBudgetResourceAssociationDeleted
	WaitBudgetResourceAssociationReady      = waitBudgetResourceAssociationReady
	WaitOrganizationsAccessStable           = waitOrganizationsAccessStable
	WaitProductPortfolioAssociationDeleted  = waitProductPortfolioAssociationDeleted
	WaitProductPortfolioAssociationReady    = waitProductPortfolioAssociationReady
	WaitProvisionedProductReady             = waitProvisionedProductReady
	WaitTagOptionResourceAssociationDeleted = waitTagOptionResourceAssociationDeleted
	WaitTagOptionResourceAssociationReady   = waitTagOptionResourceAssociationReady
)
