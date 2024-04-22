// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
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

func waitProductReady(ctx context.Context, conn *servicecatalog.ServiceCatalog, acceptLanguage, productID string, timeout time.Duration) (*servicecatalog.DescribeProductAsAdminOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{servicecatalog.StatusCreating, StatusNotFound, StatusUnavailable},
		Target:                    []string{servicecatalog.StatusAvailable, StatusCreated},
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

func waitProductDeleted(ctx context.Context, conn *servicecatalog.ServiceCatalog, acceptLanguage, productID string, timeout time.Duration) (*servicecatalog.DescribeProductAsAdminOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{servicecatalog.StatusCreating, servicecatalog.StatusAvailable, StatusCreated, StatusUnavailable},
		Target:  []string{StatusNotFound},
		Refresh: StatusProduct(ctx, conn, acceptLanguage, productID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*servicecatalog.DescribeProductAsAdminOutput); ok {
		return output, err
	}

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return nil, nil
	}

	return nil, err
}

func WaitTagOptionReady(ctx context.Context, conn *servicecatalog.ServiceCatalog, id string, timeout time.Duration) (*servicecatalog.TagOptionDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{StatusNotFound, StatusUnavailable},
		Target:  []string{servicecatalog.StatusAvailable},
		Refresh: StatusTagOption(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*servicecatalog.TagOptionDetail); ok {
		return output, err
	}

	return nil, err
}

func WaitTagOptionDeleted(ctx context.Context, conn *servicecatalog.ServiceCatalog, id string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{servicecatalog.StatusAvailable},
		Target:  []string{StatusNotFound, StatusUnavailable},
		Refresh: StatusTagOption(ctx, conn, id),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return nil
	}

	return err
}

func WaitPortfolioShareReady(ctx context.Context, conn *servicecatalog.ServiceCatalog, portfolioID, shareType, principalID string, acceptRequired bool, timeout time.Duration) (*servicecatalog.PortfolioShareDetail, error) {
	targets := []string{servicecatalog.ShareStatusCompleted}

	if !acceptRequired {
		targets = append(targets, servicecatalog.ShareStatusInProgress)
	}

	stateConf := &retry.StateChangeConf{
		Pending: []string{servicecatalog.ShareStatusNotStarted, servicecatalog.ShareStatusInProgress, StatusNotFound, StatusUnavailable},
		Target:  targets,
		Refresh: StatusPortfolioShare(ctx, conn, portfolioID, shareType, principalID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*servicecatalog.PortfolioShareDetail); ok {
		return output, err
	}

	return nil, err
}

func WaitPortfolioShareCreatedWithToken(ctx context.Context, conn *servicecatalog.ServiceCatalog, token string, acceptRequired bool, timeout time.Duration) (*servicecatalog.DescribePortfolioShareStatusOutput, error) {
	targets := []string{servicecatalog.ShareStatusCompleted}

	if !acceptRequired {
		targets = append(targets, servicecatalog.ShareStatusInProgress)
	}

	stateConf := &retry.StateChangeConf{
		Pending: []string{servicecatalog.ShareStatusNotStarted, servicecatalog.ShareStatusInProgress, StatusNotFound, StatusUnavailable},
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

func WaitPortfolioShareDeleted(ctx context.Context, conn *servicecatalog.ServiceCatalog, portfolioID, shareType, principalID string, timeout time.Duration) (*servicecatalog.PortfolioShareDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{servicecatalog.ShareStatusNotStarted, servicecatalog.ShareStatusInProgress, servicecatalog.ShareStatusCompleted, StatusUnavailable},
		Target:  []string{},
		Refresh: StatusPortfolioShare(ctx, conn, portfolioID, shareType, principalID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*servicecatalog.PortfolioShareDetail); ok {
		return output, err
	}

	return nil, err
}

func WaitPortfolioShareDeletedWithToken(ctx context.Context, conn *servicecatalog.ServiceCatalog, token string, timeout time.Duration) (*servicecatalog.DescribePortfolioShareStatusOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{servicecatalog.ShareStatusNotStarted, servicecatalog.ShareStatusInProgress, StatusNotFound, StatusUnavailable},
		Target:  []string{servicecatalog.ShareStatusCompleted},
		Refresh: StatusPortfolioShareWithToken(ctx, conn, token),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*servicecatalog.DescribePortfolioShareStatusOutput); ok {
		return output, err
	}

	return nil, err
}

func WaitOrganizationsAccessStable(ctx context.Context, conn *servicecatalog.ServiceCatalog, timeout time.Duration) (string, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{servicecatalog.AccessStatusUnderChange, StatusNotFound, StatusUnavailable},
		Target:  []string{servicecatalog.AccessStatusEnabled, servicecatalog.AccessStatusDisabled},
		Refresh: StatusOrganizationsAccess(ctx, conn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*servicecatalog.GetAWSOrganizationsAccessStatusOutput); ok {
		return aws.StringValue(output.AccessStatus), err
	}

	return "", err
}

func WaitConstraintReady(ctx context.Context, conn *servicecatalog.ServiceCatalog, acceptLanguage, id string, timeout time.Duration) (*servicecatalog.DescribeConstraintOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{StatusNotFound, servicecatalog.StatusCreating, StatusUnavailable},
		Target:                    []string{servicecatalog.StatusAvailable},
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

func WaitConstraintDeleted(ctx context.Context, conn *servicecatalog.ServiceCatalog, acceptLanguage, id string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{servicecatalog.StatusAvailable, servicecatalog.StatusCreating},
		Target:  []string{StatusNotFound},
		Refresh: StatusConstraint(ctx, conn, acceptLanguage, id),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func WaitProductPortfolioAssociationReady(ctx context.Context, conn *servicecatalog.ServiceCatalog, acceptLanguage, portfolioID, productID string, timeout time.Duration) (*servicecatalog.PortfolioDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{StatusNotFound, StatusUnavailable},
		Target:                    []string{servicecatalog.StatusAvailable},
		Refresh:                   StatusProductPortfolioAssociation(ctx, conn, acceptLanguage, portfolioID, productID),
		Timeout:                   timeout,
		ContinuousTargetOccurence: ContinuousTargetOccurrence,
		NotFoundChecks:            NotFoundChecks,
		MinTimeout:                MinTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*servicecatalog.PortfolioDetail); ok {
		return output, err
	}

	return nil, err
}

func WaitProductPortfolioAssociationDeleted(ctx context.Context, conn *servicecatalog.ServiceCatalog, acceptLanguage, portfolioID, productID string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{servicecatalog.StatusAvailable},
		Target:  []string{StatusNotFound, StatusUnavailable},
		Refresh: StatusProductPortfolioAssociation(ctx, conn, acceptLanguage, portfolioID, productID),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func WaitServiceActionReady(ctx context.Context, conn *servicecatalog.ServiceCatalog, acceptLanguage, id string, timeout time.Duration) (*servicecatalog.ServiceActionDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{StatusNotFound, StatusUnavailable},
		Target:  []string{servicecatalog.StatusAvailable},
		Refresh: StatusServiceAction(ctx, conn, acceptLanguage, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*servicecatalog.ServiceActionDetail); ok {
		return output, err
	}

	return nil, err
}

func WaitServiceActionDeleted(ctx context.Context, conn *servicecatalog.ServiceCatalog, acceptLanguage, id string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{servicecatalog.StatusAvailable},
		Target:  []string{StatusNotFound, StatusUnavailable},
		Refresh: StatusServiceAction(ctx, conn, acceptLanguage, id),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return nil
	}

	return err
}

func WaitBudgetResourceAssociationReady(ctx context.Context, conn *servicecatalog.ServiceCatalog, budgetName, resourceID string, timeout time.Duration) (*servicecatalog.BudgetDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{StatusNotFound, StatusUnavailable},
		Target:  []string{servicecatalog.StatusAvailable},
		Refresh: StatusBudgetResourceAssociation(ctx, conn, budgetName, resourceID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*servicecatalog.BudgetDetail); ok {
		return output, err
	}

	return nil, err
}

func WaitBudgetResourceAssociationDeleted(ctx context.Context, conn *servicecatalog.ServiceCatalog, budgetName, resourceID string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{servicecatalog.StatusAvailable},
		Target:  []string{StatusNotFound, StatusUnavailable},
		Refresh: StatusBudgetResourceAssociation(ctx, conn, budgetName, resourceID),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func WaitTagOptionResourceAssociationReady(ctx context.Context, conn *servicecatalog.ServiceCatalog, tagOptionID, resourceID string, timeout time.Duration) (*servicecatalog.ResourceDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{StatusNotFound, StatusUnavailable},
		Target:  []string{servicecatalog.StatusAvailable},
		Refresh: StatusTagOptionResourceAssociation(ctx, conn, tagOptionID, resourceID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*servicecatalog.ResourceDetail); ok {
		return output, err
	}

	return nil, err
}

func WaitTagOptionResourceAssociationDeleted(ctx context.Context, conn *servicecatalog.ServiceCatalog, tagOptionID, resourceID string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{servicecatalog.StatusAvailable},
		Target:  []string{StatusNotFound, StatusUnavailable},
		Refresh: StatusTagOptionResourceAssociation(ctx, conn, tagOptionID, resourceID),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func WaitProvisioningArtifactReady(ctx context.Context, conn *servicecatalog.ServiceCatalog, id, productID string, timeout time.Duration) (*servicecatalog.DescribeProvisioningArtifactOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{servicecatalog.StatusCreating, StatusNotFound, StatusUnavailable},
		Target:                    []string{servicecatalog.StatusAvailable, StatusCreated},
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

func WaitProvisioningArtifactDeleted(ctx context.Context, conn *servicecatalog.ServiceCatalog, id, productID string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{servicecatalog.StatusCreating, servicecatalog.StatusAvailable, StatusCreated, StatusUnavailable},
		Target:  []string{StatusNotFound},
		Refresh: StatusProvisioningArtifact(ctx, conn, id, productID),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return nil
	}

	return err
}

func WaitLaunchPathsReady(ctx context.Context, conn *servicecatalog.ServiceCatalog, acceptLanguage, productID string, timeout time.Duration) ([]*servicecatalog.LaunchPathSummary, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{StatusNotFound},
		Target:                    []string{servicecatalog.StatusAvailable},
		Refresh:                   StatusLaunchPaths(ctx, conn, acceptLanguage, productID),
		Timeout:                   timeout,
		ContinuousTargetOccurence: ContinuousTargetOccurrence,
		NotFoundChecks:            NotFoundChecks,
		MinTimeout:                MinTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.([]*servicecatalog.LaunchPathSummary); ok {
		return output, err
	}

	return nil, err
}

func WaitProvisionedProductReady(ctx context.Context, conn *servicecatalog.ServiceCatalog, acceptLanguage, id, name string, timeout time.Duration) (*servicecatalog.DescribeProvisionedProductOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{servicecatalog.ProvisionedProductStatusUnderChange, servicecatalog.ProvisionedProductStatusPlanInProgress},
		Target:                    []string{servicecatalog.ProvisionedProductStatusAvailable},
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
				status := aws.StringValue(detail.Status)
				if status == servicecatalog.ProvisionedProductStatusError || status == servicecatalog.ProvisionedProductStatusTainted {
					return output, errors.New(aws.StringValue(detail.StatusMessage))
				}
			}
		}
		return output, err
	}

	return nil, err
}

func WaitProvisionedProductTerminated(ctx context.Context, conn *servicecatalog.ServiceCatalog, acceptLanguage, id, name string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{servicecatalog.ProvisionedProductStatusAvailable, servicecatalog.ProvisionedProductStatusUnderChange},
		Target:  []string{},
		Refresh: StatusProvisionedProduct(ctx, conn, acceptLanguage, id, name),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func WaitPortfolioConstraintsReady(ctx context.Context, conn *servicecatalog.ServiceCatalog, acceptLanguage, portfolioID, productID string, timeout time.Duration) ([]*servicecatalog.ConstraintDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{StatusNotFound},
		Target:  []string{servicecatalog.StatusAvailable},
		Refresh: StatusPortfolioConstraints(ctx, conn, acceptLanguage, portfolioID, productID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.([]*servicecatalog.ConstraintDetail); ok {
		return output, err
	}

	return nil, err
}
