// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalog"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalog/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

const (
	BudgetResourceAssociationDeleteTimeout    = 3 * time.Minute
	BudgetResourceAssociationReadTimeout      = 10 * time.Minute
	BudgetResourceAssociationReadyTimeout     = 3 * time.Minute
	ConstraintDeleteTimeout                   = 3 * time.Minute
	ConstraintReadTimeout                     = 10 * time.Minute
	ConstraintReadyTimeout                    = 3 * time.Minute
	ConstraintUpdateTimeout                   = 3 * time.Minute
	LaunchPathsReadyTimeout                   = 3 * time.Minute
	OrganizationsAccessStableTimeout          = 3 * time.Minute
	PortfolioConstraintsReadyTimeout          = 3 * time.Minute
	PortfolioCreateTimeout                    = 30 * time.Minute
	PortfolioDeleteTimeout                    = 30 * time.Minute
	PortfolioReadTimeout                      = 10 * time.Minute
	PortfolioShareCreateTimeout               = 3 * time.Minute
	PortfolioShareDeleteTimeout               = 3 * time.Minute
	PortfolioShareReadTimeout                 = 10 * time.Minute
	PortfolioShareUpdateTimeout               = 3 * time.Minute
	PortfolioUpdateTimeout                    = 30 * time.Minute
	ProductDeleteTimeout                      = 5 * time.Minute
	ProductPortfolioAssociationDeleteTimeout  = 3 * time.Minute
	ProductPortfolioAssociationReadTimeout    = 10 * time.Minute
	ProductPortfolioAssociationReadyTimeout   = 3 * time.Minute
	ProductReadTimeout                        = 10 * time.Minute
	ProductReadyTimeout                       = 5 * time.Minute
	ProductUpdateTimeout                      = 5 * time.Minute
	ProvisionedProductDeleteTimeout           = 30 * time.Minute
	ProvisionedProductReadTimeout             = 10 * time.Minute
	ProvisionedProductReadyTimeout            = 30 * time.Minute
	ProvisionedProductUpdateTimeout           = 30 * time.Minute
	ProvisioningArtifactDeleteTimeout         = 3 * time.Minute
	ProvisioningArtifactReadTimeout           = 10 * time.Minute
	ProvisioningArtifactReadyTimeout          = 3 * time.Minute
	ProvisioningArtifactUpdateTimeout         = 3 * time.Minute
	ServiceActionDeleteTimeout                = 3 * time.Minute
	ServiceActionReadTimeout                  = 10 * time.Minute
	ServiceActionReadyTimeout                 = 3 * time.Minute
	ServiceActionUpdateTimeout                = 3 * time.Minute
	TagOptionDeleteTimeout                    = 3 * time.Minute
	TagOptionReadTimeout                      = 10 * time.Minute
	TagOptionReadyTimeout                     = 3 * time.Minute
	TagOptionResourceAssociationDeleteTimeout = 3 * time.Minute
	TagOptionResourceAssociationReadTimeout   = 10 * time.Minute
	TagOptionResourceAssociationReadyTimeout  = 3 * time.Minute
	TagOptionUpdateTimeout                    = 3 * time.Minute

	MinTimeout                 = 2 * time.Second
	NotFoundChecks             = 5
	ContinuousTargetOccurrence = 2

	StatusNotFound    = "NOT_FOUND"
	StatusUnavailable = "UNAVAILABLE"

	// AWS documentation is wrong, says that status will be "AVAILABLE" but it is actually "CREATED"
	StatusCreated = "CREATED"

	OrganizationAccessStatusError = "ERROR"
)

func waitProductReady(ctx context.Context, conn *servicecatalog.Client, acceptLanguage, productID string, timeout time.Duration) (*servicecatalog.DescribeProductAsAdminOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{string(types.StatusCreating), StatusNotFound, StatusUnavailable},
		Target:                    []string{string(types.StatusAvailable), StatusCreated},
		Refresh:                   StatusProduct(ctx, conn, acceptLanguage, productID),
		Timeout:                   timeout,
		ContinuousTargetOccurence: ContinuousTargetOccurrence,
		NotFoundChecks:            NotFoundChecks,
		MinTimeout:                MinTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*servicecatalog.DescribeProductAsAdminOutput); ok {
		return output, err
	}

	return nil, err
}

func waitProductDeleted(ctx context.Context, conn *servicecatalog.Client, acceptLanguage, productID string, timeout time.Duration) (*servicecatalog.DescribeProductAsAdminOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{string(types.StatusCreating), string(types.StatusAvailable), StatusCreated, StatusUnavailable},
		Target:  []string{StatusNotFound},
		Refresh: StatusProduct(ctx, conn, acceptLanguage, productID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*servicecatalog.DescribeProductAsAdminOutput); ok {
		return output, err
	}

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, nil
	}

	return nil, err
}

func WaitTagOptionReady(ctx context.Context, conn *servicecatalog.Client, id string, timeout time.Duration) (*types.TagOptionDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{StatusNotFound, StatusUnavailable},
		Target:  []string{string(types.StatusAvailable)},
		Refresh: StatusTagOption(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.TagOptionDetail); ok {
		return output, err
	}

	return nil, err
}

func WaitTagOptionDeleted(ctx context.Context, conn *servicecatalog.Client, id string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{string(types.StatusAvailable)},
		Target:  []string{StatusNotFound, StatusUnavailable},
		Refresh: StatusTagOption(ctx, conn, id),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil
	}

	return err
}

func WaitPortfolioShareReady(ctx context.Context, conn *servicecatalog.Client, portfolioID, shareType, principalID string, acceptRequired bool, timeout time.Duration) (*types.PortfolioShareDetail, error) {
	targets := []string{string(types.ShareStatusCompleted)}

	if !acceptRequired {
		targets = append(targets, string(types.ShareStatusInProgress))
	}

	stateConf := &retry.StateChangeConf{
		Pending: []string{string(types.ShareStatusNotStarted), string(types.ShareStatusInProgress), StatusNotFound, StatusUnavailable},
		Target:  targets,
		Refresh: StatusPortfolioShare(ctx, conn, portfolioID, shareType, principalID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.PortfolioShareDetail); ok {
		return output, err
	}

	return nil, err
}

func WaitPortfolioShareCreatedWithToken(ctx context.Context, conn *servicecatalog.Client, token string, acceptRequired bool, timeout time.Duration) (*servicecatalog.DescribePortfolioShareStatusOutput, error) {
	targets := []string{string(types.ShareStatusCompleted)}

	if !acceptRequired {
		targets = append(targets, string(types.ShareStatusInProgress))
	}

	stateConf := &retry.StateChangeConf{
		Pending: []string{string(types.ShareStatusNotStarted), string(types.ShareStatusInProgress), StatusNotFound, StatusUnavailable},
		Target:  targets,
		Refresh: StatusPortfolioShareWithToken(ctx, conn, token),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*servicecatalog.DescribePortfolioShareStatusOutput); ok {
		return output, err
	}

	return nil, err
}

func WaitPortfolioShareDeleted(ctx context.Context, conn *servicecatalog.Client, portfolioID, shareType, principalID string, timeout time.Duration) (*types.PortfolioShareDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{string(types.ShareStatusNotStarted), string(types.ShareStatusInProgress), string(types.ShareStatusCompleted), StatusUnavailable},
		Target:  []string{},
		Refresh: StatusPortfolioShare(ctx, conn, portfolioID, shareType, principalID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.PortfolioShareDetail); ok {
		return output, err
	}

	return nil, err
}

func WaitPortfolioShareDeletedWithToken(ctx context.Context, conn *servicecatalog.Client, token string, timeout time.Duration) (*servicecatalog.DescribePortfolioShareStatusOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{string(types.ShareStatusNotStarted), string(types.ShareStatusInProgress), StatusNotFound, StatusUnavailable},
		Target:  []string{string(types.ShareStatusCompleted)},
		Refresh: StatusPortfolioShareWithToken(ctx, conn, token),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*servicecatalog.DescribePortfolioShareStatusOutput); ok {
		return output, err
	}

	return nil, err
}

func WaitOrganizationsAccessStable(ctx context.Context, conn *servicecatalog.Client, timeout time.Duration) (string, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{string(types.AccessStatusUnderChange), StatusNotFound, StatusUnavailable},
		Target:  []string{string(types.AccessStatusEnabled), string(types.AccessStatusDisabled)},
		Refresh: StatusOrganizationsAccess(ctx, conn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*servicecatalog.GetAWSOrganizationsAccessStatusOutput); ok {
		return aws.ToString(output.AccessStatus), err
	}

	return "", err
}

func WaitConstraintReady(ctx context.Context, conn *servicecatalog.Client, acceptLanguage, id string, timeout time.Duration) (*servicecatalog.DescribeConstraintOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{StatusNotFound, string(types.StatusCreating), StatusUnavailable},
		Target:                    []string{string(types.StatusAvailable)},
		Refresh:                   StatusConstraint(ctx, conn, acceptLanguage, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: ContinuousTargetOccurrence,
		NotFoundChecks:            NotFoundChecks,
		MinTimeout:                MinTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*servicecatalog.DescribeConstraintOutput); ok {
		return output, err
	}

	return nil, err
}

func WaitConstraintDeleted(ctx context.Context, conn *servicecatalog.Client, acceptLanguage, id string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{string(types.StatusAvailable), string(types.StatusCreating)},
		Target:  []string{StatusNotFound},
		Refresh: StatusConstraint(ctx, conn, acceptLanguage, id),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func WaitProductPortfolioAssociationReady(ctx context.Context, conn *servicecatalog.Client, acceptLanguage, portfolioID, productID string, timeout time.Duration) (*types.PortfolioDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{StatusNotFound, StatusUnavailable},
		Target:                    []string{string(types.StatusAvailable)},
		Refresh:                   StatusProductPortfolioAssociation(ctx, conn, acceptLanguage, portfolioID, productID),
		Timeout:                   timeout,
		ContinuousTargetOccurence: ContinuousTargetOccurrence,
		NotFoundChecks:            NotFoundChecks,
		MinTimeout:                MinTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.PortfolioDetail); ok {
		return output, err
	}

	return nil, err
}

func WaitProductPortfolioAssociationDeleted(ctx context.Context, conn *servicecatalog.Client, acceptLanguage, portfolioID, productID string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{string(types.StatusAvailable)},
		Target:  []string{StatusNotFound, StatusUnavailable},
		Refresh: StatusProductPortfolioAssociation(ctx, conn, acceptLanguage, portfolioID, productID),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func WaitServiceActionReady(ctx context.Context, conn *servicecatalog.Client, acceptLanguage, id string, timeout time.Duration) (*types.ServiceActionDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{StatusNotFound, StatusUnavailable},
		Target:  []string{string(types.StatusAvailable)},
		Refresh: StatusServiceAction(ctx, conn, acceptLanguage, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.ServiceActionDetail); ok {
		return output, err
	}

	return nil, err
}

func WaitServiceActionDeleted(ctx context.Context, conn *servicecatalog.Client, acceptLanguage, id string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{string(types.StatusAvailable)},
		Target:  []string{StatusNotFound, StatusUnavailable},
		Refresh: StatusServiceAction(ctx, conn, acceptLanguage, id),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil
	}

	return err
}

func WaitBudgetResourceAssociationReady(ctx context.Context, conn *servicecatalog.Client, budgetName, resourceID string, timeout time.Duration) (*types.BudgetDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{StatusNotFound, StatusUnavailable},
		Target:  []string{string(types.StatusAvailable)},
		Refresh: StatusBudgetResourceAssociation(ctx, conn, budgetName, resourceID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.BudgetDetail); ok {
		return output, err
	}

	return nil, err
}

func WaitBudgetResourceAssociationDeleted(ctx context.Context, conn *servicecatalog.Client, budgetName, resourceID string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{string(types.StatusAvailable)},
		Target:  []string{StatusNotFound, StatusUnavailable},
		Refresh: StatusBudgetResourceAssociation(ctx, conn, budgetName, resourceID),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func WaitTagOptionResourceAssociationReady(ctx context.Context, conn *servicecatalog.Client, tagOptionID, resourceID string, timeout time.Duration) (*types.ResourceDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{StatusNotFound, StatusUnavailable},
		Target:  []string{string(types.StatusAvailable)},
		Refresh: StatusTagOptionResourceAssociation(ctx, conn, tagOptionID, resourceID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.ResourceDetail); ok {
		return output, err
	}

	return nil, err
}

func WaitTagOptionResourceAssociationDeleted(ctx context.Context, conn *servicecatalog.Client, tagOptionID, resourceID string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{string(types.StatusAvailable)},
		Target:  []string{StatusNotFound, StatusUnavailable},
		Refresh: StatusTagOptionResourceAssociation(ctx, conn, tagOptionID, resourceID),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func WaitProvisioningArtifactReady(ctx context.Context, conn *servicecatalog.Client, id, productID string, timeout time.Duration) (*servicecatalog.DescribeProvisioningArtifactOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{string(types.StatusCreating), StatusNotFound, StatusUnavailable},
		Target:                    []string{string(types.StatusAvailable), StatusCreated},
		Refresh:                   StatusProvisioningArtifact(ctx, conn, id, productID),
		Timeout:                   timeout,
		ContinuousTargetOccurence: ContinuousTargetOccurrence,
		NotFoundChecks:            NotFoundChecks,
		MinTimeout:                MinTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*servicecatalog.DescribeProvisioningArtifactOutput); ok {
		return output, err
	}

	return nil, err
}

func WaitProvisioningArtifactDeleted(ctx context.Context, conn *servicecatalog.Client, id, productID string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{string(types.StatusCreating), string(types.StatusAvailable), StatusCreated, StatusUnavailable},
		Target:  []string{StatusNotFound},
		Refresh: StatusProvisioningArtifact(ctx, conn, id, productID),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil
	}

	return err
}

func WaitLaunchPathsReady(ctx context.Context, conn *servicecatalog.Client, acceptLanguage, productID string, timeout time.Duration) ([]*types.LaunchPathSummary, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{StatusNotFound},
		Target:                    []string{string(types.StatusAvailable)},
		Refresh:                   StatusLaunchPaths(ctx, conn, acceptLanguage, productID),
		Timeout:                   timeout,
		ContinuousTargetOccurence: ContinuousTargetOccurrence,
		NotFoundChecks:            NotFoundChecks,
		MinTimeout:                MinTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.([]*types.LaunchPathSummary); ok {
		return output, err
	}

	return nil, err
}

func WaitProvisionedProductReady(ctx context.Context, conn *servicecatalog.Client, acceptLanguage, id, name string, timeout time.Duration) (*servicecatalog.DescribeProvisionedProductOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{string(types.ProvisionedProductStatusUnderChange), string(types.ProvisionedProductStatusPlanInProgress)},
		Target:                    []string{string(types.ProvisionedProductStatusAvailable)},
		Refresh:                   StatusProvisionedProduct(ctx, conn, acceptLanguage, id, name),
		Timeout:                   timeout,
		ContinuousTargetOccurence: ContinuousTargetOccurrence,
		NotFoundChecks:            NotFoundChecks,
		MinTimeout:                MinTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*servicecatalog.DescribeProvisionedProductOutput); ok {
		if detail := output.ProvisionedProductDetail; detail != nil {
			var foo *retry.UnexpectedStateError
			if errors.As(err, &foo) {
				// The statuses `ERROR` and `TAINTED` are equivalent: the application of the requested change has failed.
				// The difference is that, in the case of `TAINTED`, there is a previous version to roll back to.
				status := aws.ToString(detail.Status)
				if status == string(types.ProvisionedProductStatusError) || status == string(types.ProvisionedProductStatusTainted) {
					return output, errors.New(aws.ToString(detail.StatusMessage))
				}
			}
		}
		return output, err
	}

	return nil, err
}

func WaitProvisionedProductTerminated(ctx context.Context, conn *servicecatalog.Client, acceptLanguage, id, name string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{string(types.ProvisionedProductStatusAvailable), string(types.ProvisionedProductStatusUnderChange)},
		Target:  []string{},
		Refresh: StatusProvisionedProduct(ctx, conn, acceptLanguage, id, name),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func WaitPortfolioConstraintsReady(ctx context.Context, conn *servicecatalog.Client, acceptLanguage, portfolioID, productID string, timeout time.Duration) ([]*types.ConstraintDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{StatusNotFound},
		Target:  []string{string(types.StatusAvailable)},
		Refresh: StatusPortfolioConstraints(ctx, conn, acceptLanguage, portfolioID, productID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.([]*types.ConstraintDetail); ok {
		return output, err
	}

	return nil, err
}
