// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package servicecatalog

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/servicecatalog"
	awstypes "github.com/aws/aws-sdk-go-v2/service/servicecatalog/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
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

	minTimeout                 = 2 * time.Second
	notFoundChecks             = 5
	continuousTargetOccurrence = 2

	statusNotFound    = "NOT_FOUND"
	statusUnavailable = "UNAVAILABLE"

	// AWS documentation is wrong, says that status will be "AVAILABLE" but it is actually "CREATED"
	statusCreated = "CREATED"

	organizationAccessStatusError = "ERROR"
)

func waitProductReady(ctx context.Context, conn *servicecatalog.Client, acceptLanguage, productID string, timeout time.Duration) (*servicecatalog.DescribeProductAsAdminOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.StatusCreating, statusNotFound, statusUnavailable),
		Target:                    enum.Slice(awstypes.StatusAvailable, statusCreated),
		Refresh:                   statusProduct(conn, acceptLanguage, productID),
		Timeout:                   timeout,
		ContinuousTargetOccurence: continuousTargetOccurrence,
		NotFoundChecks:            notFoundChecks,
		MinTimeout:                minTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*servicecatalog.DescribeProductAsAdminOutput); ok {
		return output, err
	}

	return nil, err
}

func waitProductDeleted(ctx context.Context, conn *servicecatalog.Client, acceptLanguage, productID string, timeout time.Duration) (*servicecatalog.DescribeProductAsAdminOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StatusCreating, awstypes.StatusAvailable, statusCreated, statusUnavailable),
		Target:  []string{statusNotFound},
		Refresh: statusProduct(conn, acceptLanguage, productID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*servicecatalog.DescribeProductAsAdminOutput); ok {
		return output, err
	}

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, nil
	}

	return nil, err
}

func waitTagOptionReady(ctx context.Context, conn *servicecatalog.Client, id string, timeout time.Duration) (*awstypes.TagOptionDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusNotFound, statusUnavailable},
		Target:  enum.Slice(awstypes.StatusAvailable),
		Refresh: statusTagOption(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TagOptionDetail); ok {
		return output, err
	}

	return nil, err
}

func waitTagOptionDeleted(ctx context.Context, conn *servicecatalog.Client, id string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StatusAvailable),
		Target:  []string{statusNotFound, statusUnavailable},
		Refresh: statusTagOption(conn, id),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil
	}

	return err
}

func waitPortfolioShareReady(ctx context.Context, conn *servicecatalog.Client, portfolioID, shareType, principalID string, acceptRequired bool, timeout time.Duration) (*awstypes.PortfolioShareDetail, error) {
	targets := enum.Slice(awstypes.ShareStatusCompleted)

	if !acceptRequired {
		targets = append(targets, string(awstypes.ShareStatusInProgress))
	}

	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ShareStatusNotStarted, awstypes.ShareStatusInProgress, statusNotFound, statusUnavailable),
		Target:  targets,
		Refresh: statusPortfolioShare(conn, portfolioID, shareType, principalID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.PortfolioShareDetail); ok {
		return output, err
	}

	return nil, err
}

func waitPortfolioShareCreatedWithToken(ctx context.Context, conn *servicecatalog.Client, token string, waitForAcceptance bool, timeout time.Duration) (*servicecatalog.DescribePortfolioShareStatusOutput, error) {
	targets := enum.Slice(awstypes.ShareStatusCompleted)

	if !waitForAcceptance {
		targets = append(targets, string(awstypes.ShareStatusInProgress))
	}

	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ShareStatusNotStarted, awstypes.ShareStatusInProgress, statusNotFound, statusUnavailable),
		Target:  targets,
		Refresh: statusPortfolioShareWithToken(conn, token),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*servicecatalog.DescribePortfolioShareStatusOutput); ok {
		return output, err
	}

	return nil, err
}

func waitPortfolioShareDeleted(ctx context.Context, conn *servicecatalog.Client, portfolioID, shareType, principalID string, timeout time.Duration) (*awstypes.PortfolioShareDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ShareStatusNotStarted, awstypes.ShareStatusInProgress, awstypes.ShareStatusCompleted, statusUnavailable),
		Target:  []string{},
		Refresh: statusPortfolioShare(conn, portfolioID, shareType, principalID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.PortfolioShareDetail); ok {
		return output, err
	}

	return nil, err
}

func waitPortfolioShareDeletedWithToken(ctx context.Context, conn *servicecatalog.Client, token string, timeout time.Duration) (*servicecatalog.DescribePortfolioShareStatusOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ShareStatusNotStarted, awstypes.ShareStatusInProgress, statusNotFound, statusUnavailable),
		Target:  enum.Slice(awstypes.ShareStatusCompleted),
		Refresh: statusPortfolioShareWithToken(conn, token),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*servicecatalog.DescribePortfolioShareStatusOutput); ok {
		return output, err
	}

	return nil, err
}

func waitOrganizationsAccessStable(ctx context.Context, conn *servicecatalog.Client, timeout time.Duration) (string, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AccessStatusUnderChange, statusNotFound, statusUnavailable),
		Target:  enum.Slice(awstypes.AccessStatusEnabled, awstypes.AccessStatusDisabled),
		Refresh: statusOrganizationsAccess(conn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*servicecatalog.GetAWSOrganizationsAccessStatusOutput); ok {
		return string(output.AccessStatus), err
	}

	return "", err
}

func waitConstraintReady(ctx context.Context, conn *servicecatalog.Client, acceptLanguage, id string, timeout time.Duration) (*servicecatalog.DescribeConstraintOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(statusNotFound, awstypes.StatusCreating, statusUnavailable),
		Target:                    enum.Slice(awstypes.StatusAvailable),
		Refresh:                   statusConstraint(conn, acceptLanguage, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: continuousTargetOccurrence,
		NotFoundChecks:            notFoundChecks,
		MinTimeout:                minTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*servicecatalog.DescribeConstraintOutput); ok {
		return output, err
	}

	return nil, err
}

func waitConstraintDeleted(ctx context.Context, conn *servicecatalog.Client, acceptLanguage, id string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StatusAvailable, awstypes.StatusCreating),
		Target:  []string{statusNotFound},
		Refresh: statusConstraint(conn, acceptLanguage, id),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitProductPortfolioAssociationReady(ctx context.Context, conn *servicecatalog.Client, acceptLanguage, portfolioID, productID string, timeout time.Duration) (*awstypes.PortfolioDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusNotFound, statusUnavailable},
		Target:                    enum.Slice(awstypes.StatusAvailable),
		Refresh:                   statusProductPortfolioAssociation(conn, acceptLanguage, portfolioID, productID),
		Timeout:                   timeout,
		ContinuousTargetOccurence: continuousTargetOccurrence,
		NotFoundChecks:            notFoundChecks,
		MinTimeout:                minTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.PortfolioDetail); ok {
		return output, err
	}

	return nil, err
}

func waitProductPortfolioAssociationDeleted(ctx context.Context, conn *servicecatalog.Client, acceptLanguage, portfolioID, productID string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StatusAvailable),
		Target:  []string{statusNotFound, statusUnavailable},
		Refresh: statusProductPortfolioAssociation(conn, acceptLanguage, portfolioID, productID),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitServiceActionReady(ctx context.Context, conn *servicecatalog.Client, acceptLanguage, id string, timeout time.Duration) (*awstypes.ServiceActionDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusNotFound, statusUnavailable},
		Target:  enum.Slice(awstypes.StatusAvailable),
		Refresh: statusServiceAction(conn, acceptLanguage, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ServiceActionDetail); ok {
		return output, err
	}

	return nil, err
}

func waitServiceActionDeleted(ctx context.Context, conn *servicecatalog.Client, acceptLanguage, id string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StatusAvailable),
		Target:  []string{statusNotFound, statusUnavailable},
		Refresh: statusServiceAction(conn, acceptLanguage, id),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil
	}

	return err
}

func waitBudgetResourceAssociationReady(ctx context.Context, conn *servicecatalog.Client, budgetName, resourceID string, timeout time.Duration) (*awstypes.BudgetDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusNotFound, statusUnavailable},
		Target:  enum.Slice(awstypes.StatusAvailable),
		Refresh: statusBudgetResourceAssociation(conn, budgetName, resourceID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.BudgetDetail); ok {
		return output, err
	}

	return nil, err
}

func waitBudgetResourceAssociationDeleted(ctx context.Context, conn *servicecatalog.Client, budgetName, resourceID string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StatusAvailable),
		Target:  []string{statusNotFound, statusUnavailable},
		Refresh: statusBudgetResourceAssociation(conn, budgetName, resourceID),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitTagOptionResourceAssociationReady(ctx context.Context, conn *servicecatalog.Client, tagOptionID, resourceID string, timeout time.Duration) (*awstypes.ResourceDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusNotFound, statusUnavailable},
		Target:  enum.Slice(awstypes.StatusAvailable),
		Refresh: statusTagOptionResourceAssociation(conn, tagOptionID, resourceID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ResourceDetail); ok {
		return output, err
	}

	return nil, err
}

func waitTagOptionResourceAssociationDeleted(ctx context.Context, conn *servicecatalog.Client, tagOptionID, resourceID string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StatusAvailable),
		Target:  []string{statusNotFound, statusUnavailable},
		Refresh: statusTagOptionResourceAssociation(conn, tagOptionID, resourceID),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitProvisioningArtifactReady(ctx context.Context, conn *servicecatalog.Client, id, productID string, timeout time.Duration) (*servicecatalog.DescribeProvisioningArtifactOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.StatusCreating, statusNotFound, statusUnavailable),
		Target:                    enum.Slice(awstypes.StatusAvailable, statusCreated),
		Refresh:                   statusProvisioningArtifact(conn, id, productID),
		Timeout:                   timeout,
		ContinuousTargetOccurence: continuousTargetOccurrence,
		NotFoundChecks:            notFoundChecks,
		MinTimeout:                minTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*servicecatalog.DescribeProvisioningArtifactOutput); ok {
		return output, err
	}

	return nil, err
}

func waitProvisioningArtifactDeleted(ctx context.Context, conn *servicecatalog.Client, id, productID string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StatusCreating, awstypes.StatusAvailable, statusCreated, statusUnavailable),
		Target:  []string{statusNotFound},
		Refresh: statusProvisioningArtifact(conn, id, productID),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil
	}

	return err
}

func waitLaunchPathsReady(ctx context.Context, conn *servicecatalog.Client, acceptLanguage, productID string, timeout time.Duration) ([]awstypes.LaunchPathSummary, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusNotFound},
		Target:                    enum.Slice(awstypes.StatusAvailable),
		Refresh:                   statusLaunchPaths(conn, acceptLanguage, productID),
		Timeout:                   timeout,
		ContinuousTargetOccurence: continuousTargetOccurrence,
		NotFoundChecks:            notFoundChecks,
		MinTimeout:                minTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.([]awstypes.LaunchPathSummary); ok {
		return output, err
	}

	return nil, err
}

func waitPortfolioConstraintsReady(ctx context.Context, conn *servicecatalog.Client, acceptLanguage, portfolioID, productID string, timeout time.Duration) ([]awstypes.ConstraintDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusNotFound},
		Target:  enum.Slice(awstypes.StatusAvailable),
		Refresh: statusPortfolioConstraints(conn, acceptLanguage, portfolioID, productID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.([]awstypes.ConstraintDetail); ok {
		return output, err
	}

	return nil, err
}
