package servicecatalog

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	ProductReadyTimeout  = 3 * time.Minute
	ProductDeleteTimeout = 3 * time.Minute

	TagOptionReadyTimeout  = 3 * time.Minute
	TagOptionDeleteTimeout = 3 * time.Minute

	PortfolioShareCreateTimeout = 3 * time.Minute

	OrganizationsAccessStableTimeout = 3 * time.Minute
	ConstraintReadyTimeout           = 3 * time.Minute
	ConstraintDeleteTimeout          = 3 * time.Minute

	ProductPortfolioAssociationReadyTimeout  = 3 * time.Minute
	ProductPortfolioAssociationDeleteTimeout = 3 * time.Minute

	ServiceActionReadyTimeout  = 3 * time.Minute
	ServiceActionDeleteTimeout = 3 * time.Minute

	BudgetResourceAssociationReadyTimeout  = 3 * time.Minute
	BudgetResourceAssociationDeleteTimeout = 3 * time.Minute

	TagOptionResourceAssociationReadyTimeout  = 3 * time.Minute
	TagOptionResourceAssociationDeleteTimeout = 3 * time.Minute

	ProvisioningArtifactReadyTimeout   = 3 * time.Minute
	ProvisioningArtifactDeletedTimeout = 3 * time.Minute

	PrincipalPortfolioAssociationReadyTimeout  = 3 * time.Minute
	PrincipalPortfolioAssociationDeleteTimeout = 3 * time.Minute

	LaunchPathsReadyTimeout = 3 * time.Minute

	ProvisionedProductReadyTimeout  = 30 * time.Minute
	ProvisionedProductUpdateTimeout = 30 * time.Minute
	ProvisionedProductDeleteTimeout = 30 * time.Minute

	RecordReadyTimeout = 30 * time.Minute

	PortfolioConstraintsReadyTimeout = 3 * time.Minute

	MinTimeout                 = 2 * time.Second
	NotFoundChecks             = 5
	ContinuousTargetOccurrence = 2

	StatusNotFound    = "NOT_FOUND"
	StatusUnavailable = "UNAVAILABLE"

	// AWS documentation is wrong, says that status will be "AVAILABLE" but it is actually "CREATED"
	StatusCreated = "CREATED"

	OrganizationAccessStatusError = "ERROR"
)

func WaitProductReady(conn *servicecatalog.ServiceCatalog, acceptLanguage, productID string) (*servicecatalog.DescribeProductAsAdminOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{servicecatalog.StatusCreating, StatusNotFound, StatusUnavailable},
		Target:                    []string{servicecatalog.StatusAvailable, StatusCreated},
		Refresh:                   StatusProduct(conn, acceptLanguage, productID),
		Timeout:                   ProductReadyTimeout,
		ContinuousTargetOccurence: ContinuousTargetOccurrence,
		NotFoundChecks:            NotFoundChecks,
		MinTimeout:                MinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*servicecatalog.DescribeProductAsAdminOutput); ok {
		return output, err
	}

	return nil, err
}

func WaitProductDeleted(conn *servicecatalog.ServiceCatalog, acceptLanguage, productID string) (*servicecatalog.DescribeProductAsAdminOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{servicecatalog.StatusCreating, servicecatalog.StatusAvailable, StatusCreated, StatusUnavailable},
		Target:  []string{StatusNotFound},
		Refresh: StatusProduct(conn, acceptLanguage, productID),
		Timeout: ProductDeleteTimeout,
	}

	_, err := stateConf.WaitForState()

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return nil, nil
	}

	return nil, err
}

func WaitTagOptionReady(conn *servicecatalog.ServiceCatalog, id string) (*servicecatalog.TagOptionDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{StatusNotFound, StatusUnavailable},
		Target:  []string{servicecatalog.StatusAvailable},
		Refresh: StatusTagOption(conn, id),
		Timeout: TagOptionReadyTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*servicecatalog.TagOptionDetail); ok {
		return output, err
	}

	return nil, err
}

func WaitTagOptionDeleted(conn *servicecatalog.ServiceCatalog, id string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{servicecatalog.StatusAvailable},
		Target:  []string{StatusNotFound, StatusUnavailable},
		Refresh: StatusTagOption(conn, id),
		Timeout: TagOptionDeleteTimeout,
	}

	_, err := stateConf.WaitForState()

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return nil
	}

	return err
}

func WaitPortfolioShareReady(conn *servicecatalog.ServiceCatalog, portfolioID, shareType, principalID string, acceptRequired bool) (*servicecatalog.PortfolioShareDetail, error) {
	targets := []string{servicecatalog.ShareStatusCompleted}

	if !acceptRequired {
		targets = append(targets, servicecatalog.ShareStatusInProgress)
	}

	stateConf := &resource.StateChangeConf{
		Pending: []string{servicecatalog.ShareStatusNotStarted, servicecatalog.ShareStatusInProgress, StatusNotFound, StatusUnavailable},
		Target:  targets,
		Refresh: StatusPortfolioShare(conn, portfolioID, shareType, principalID),
		Timeout: PortfolioShareCreateTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*servicecatalog.PortfolioShareDetail); ok {
		return output, err
	}

	return nil, err
}

func WaitPortfolioShareCreatedWithToken(conn *servicecatalog.ServiceCatalog, token string, acceptRequired bool) (*servicecatalog.DescribePortfolioShareStatusOutput, error) {
	targets := []string{servicecatalog.ShareStatusCompleted}

	if !acceptRequired {
		targets = append(targets, servicecatalog.ShareStatusInProgress)
	}

	stateConf := &resource.StateChangeConf{
		Pending: []string{servicecatalog.ShareStatusNotStarted, servicecatalog.ShareStatusInProgress, StatusNotFound, StatusUnavailable},
		Target:  targets,
		Refresh: StatusPortfolioShareWithToken(conn, token),
		Timeout: PortfolioShareCreateTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*servicecatalog.DescribePortfolioShareStatusOutput); ok {
		return output, err
	}

	return nil, err
}

func WaitPortfolioShareDeleted(conn *servicecatalog.ServiceCatalog, portfolioID, shareType, principalID string) (*servicecatalog.PortfolioShareDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{servicecatalog.ShareStatusNotStarted, servicecatalog.ShareStatusInProgress, servicecatalog.ShareStatusCompleted, StatusUnavailable},
		Target:  []string{StatusNotFound},
		Refresh: StatusPortfolioShare(conn, portfolioID, shareType, principalID),
		Timeout: PortfolioShareCreateTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if tfresource.NotFound(err) {
		return nil, nil
	}

	if output, ok := outputRaw.(*servicecatalog.PortfolioShareDetail); ok {
		return output, err
	}

	return nil, err
}

func WaitPortfolioShareDeletedWithToken(conn *servicecatalog.ServiceCatalog, token string) (*servicecatalog.DescribePortfolioShareStatusOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{servicecatalog.ShareStatusNotStarted, servicecatalog.ShareStatusInProgress, StatusNotFound, StatusUnavailable},
		Target:  []string{servicecatalog.ShareStatusCompleted},
		Refresh: StatusPortfolioShareWithToken(conn, token),
		Timeout: PortfolioShareCreateTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*servicecatalog.DescribePortfolioShareStatusOutput); ok {
		return output, err
	}

	return nil, err
}

func WaitOrganizationsAccessStable(conn *servicecatalog.ServiceCatalog) (string, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{servicecatalog.AccessStatusUnderChange, StatusNotFound, StatusUnavailable},
		Target:  []string{servicecatalog.AccessStatusEnabled, servicecatalog.AccessStatusDisabled},
		Refresh: StatusOrganizationsAccess(conn),
		Timeout: OrganizationsAccessStableTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*servicecatalog.GetAWSOrganizationsAccessStatusOutput); ok {
		return aws.StringValue(output.AccessStatus), err
	}

	return "", err
}

func WaitConstraintReady(conn *servicecatalog.ServiceCatalog, acceptLanguage, id string) (*servicecatalog.DescribeConstraintOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{StatusNotFound, servicecatalog.StatusCreating, StatusUnavailable},
		Target:                    []string{servicecatalog.StatusAvailable},
		Refresh:                   StatusConstraint(conn, acceptLanguage, id),
		Timeout:                   ConstraintReadyTimeout,
		ContinuousTargetOccurence: ContinuousTargetOccurrence,
		NotFoundChecks:            NotFoundChecks,
		MinTimeout:                MinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*servicecatalog.DescribeConstraintOutput); ok {
		return output, err
	}

	return nil, err
}

func WaitConstraintDeleted(conn *servicecatalog.ServiceCatalog, acceptLanguage, id string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{servicecatalog.StatusAvailable, servicecatalog.StatusCreating},
		Target:  []string{StatusNotFound},
		Refresh: StatusConstraint(conn, acceptLanguage, id),
		Timeout: ConstraintDeleteTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}

func WaitProductPortfolioAssociationReady(conn *servicecatalog.ServiceCatalog, acceptLanguage, portfolioID, productID string) (*servicecatalog.PortfolioDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{StatusNotFound, StatusUnavailable},
		Target:                    []string{servicecatalog.StatusAvailable},
		Refresh:                   StatusProductPortfolioAssociation(conn, acceptLanguage, portfolioID, productID),
		Timeout:                   ProductPortfolioAssociationReadyTimeout,
		ContinuousTargetOccurence: ContinuousTargetOccurrence,
		NotFoundChecks:            NotFoundChecks,
		MinTimeout:                MinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*servicecatalog.PortfolioDetail); ok {
		return output, err
	}

	return nil, err
}

func WaitProductPortfolioAssociationDeleted(conn *servicecatalog.ServiceCatalog, acceptLanguage, portfolioID, productID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{servicecatalog.StatusAvailable},
		Target:  []string{StatusNotFound, StatusUnavailable},
		Refresh: StatusProductPortfolioAssociation(conn, acceptLanguage, portfolioID, productID),
		Timeout: ProductPortfolioAssociationDeleteTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}

func WaitServiceActionReady(conn *servicecatalog.ServiceCatalog, acceptLanguage, id string) (*servicecatalog.ServiceActionDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{StatusNotFound, StatusUnavailable},
		Target:  []string{servicecatalog.StatusAvailable},
		Refresh: StatusServiceAction(conn, acceptLanguage, id),
		Timeout: ServiceActionReadyTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*servicecatalog.ServiceActionDetail); ok {
		return output, err
	}

	return nil, err
}

func WaitServiceActionDeleted(conn *servicecatalog.ServiceCatalog, acceptLanguage, id string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{servicecatalog.StatusAvailable},
		Target:  []string{StatusNotFound, StatusUnavailable},
		Refresh: StatusServiceAction(conn, acceptLanguage, id),
		Timeout: ServiceActionDeleteTimeout,
	}

	_, err := stateConf.WaitForState()

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return nil
	}

	return err
}

func WaitBudgetResourceAssociationReady(conn *servicecatalog.ServiceCatalog, budgetName, resourceID string) (*servicecatalog.BudgetDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{StatusNotFound, StatusUnavailable},
		Target:  []string{servicecatalog.StatusAvailable},
		Refresh: StatusBudgetResourceAssociation(conn, budgetName, resourceID),
		Timeout: BudgetResourceAssociationReadyTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*servicecatalog.BudgetDetail); ok {
		return output, err
	}

	return nil, err
}

func WaitBudgetResourceAssociationDeleted(conn *servicecatalog.ServiceCatalog, budgetName, resourceID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{servicecatalog.StatusAvailable},
		Target:  []string{StatusNotFound, StatusUnavailable},
		Refresh: StatusBudgetResourceAssociation(conn, budgetName, resourceID),
		Timeout: BudgetResourceAssociationDeleteTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}

func WaitTagOptionResourceAssociationReady(conn *servicecatalog.ServiceCatalog, tagOptionID, resourceID string) (*servicecatalog.ResourceDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{StatusNotFound, StatusUnavailable},
		Target:  []string{servicecatalog.StatusAvailable},
		Refresh: StatusTagOptionResourceAssociation(conn, tagOptionID, resourceID),
		Timeout: TagOptionResourceAssociationReadyTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*servicecatalog.ResourceDetail); ok {
		return output, err
	}

	return nil, err
}

func WaitTagOptionResourceAssociationDeleted(conn *servicecatalog.ServiceCatalog, tagOptionID, resourceID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{servicecatalog.StatusAvailable},
		Target:  []string{StatusNotFound, StatusUnavailable},
		Refresh: StatusTagOptionResourceAssociation(conn, tagOptionID, resourceID),
		Timeout: TagOptionResourceAssociationDeleteTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}

func WaitProvisioningArtifactReady(conn *servicecatalog.ServiceCatalog, id, productID string) (*servicecatalog.DescribeProvisioningArtifactOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{servicecatalog.StatusCreating, StatusNotFound, StatusUnavailable},
		Target:                    []string{servicecatalog.StatusAvailable, StatusCreated},
		Refresh:                   StatusProvisioningArtifact(conn, id, productID),
		Timeout:                   ProvisioningArtifactReadyTimeout,
		ContinuousTargetOccurence: ContinuousTargetOccurrence,
		NotFoundChecks:            NotFoundChecks,
		MinTimeout:                MinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*servicecatalog.DescribeProvisioningArtifactOutput); ok {
		return output, err
	}

	return nil, err
}

func WaitProvisioningArtifactDeleted(conn *servicecatalog.ServiceCatalog, id, productID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{servicecatalog.StatusCreating, servicecatalog.StatusAvailable, StatusCreated, StatusUnavailable},
		Target:  []string{StatusNotFound},
		Refresh: StatusProvisioningArtifact(conn, id, productID),
		Timeout: ProvisioningArtifactDeletedTimeout,
	}

	_, err := stateConf.WaitForState()

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return err
	}

	return nil
}

func WaitPrincipalPortfolioAssociationReady(conn *servicecatalog.ServiceCatalog, acceptLanguage, principalARN, portfolioID string) (*servicecatalog.Principal, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{StatusNotFound, StatusUnavailable},
		Target:                    []string{servicecatalog.StatusAvailable},
		Refresh:                   StatusPrincipalPortfolioAssociation(conn, acceptLanguage, principalARN, portfolioID),
		Timeout:                   PrincipalPortfolioAssociationReadyTimeout,
		ContinuousTargetOccurence: ContinuousTargetOccurrence,
		NotFoundChecks:            NotFoundChecks,
		MinTimeout:                MinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*servicecatalog.Principal); ok {
		return output, err
	}

	return nil, err
}

func WaitPrincipalPortfolioAssociationDeleted(conn *servicecatalog.ServiceCatalog, acceptLanguage, principalARN, portfolioID string) error {
	stateConf := &resource.StateChangeConf{
		Pending:        []string{servicecatalog.StatusAvailable},
		Target:         []string{StatusNotFound, StatusUnavailable},
		Refresh:        StatusPrincipalPortfolioAssociation(conn, acceptLanguage, principalARN, portfolioID),
		Timeout:        PrincipalPortfolioAssociationDeleteTimeout,
		NotFoundChecks: 1,
	}

	_, err := stateConf.WaitForState()

	return err
}

func WaitLaunchPathsReady(conn *servicecatalog.ServiceCatalog, acceptLanguage, productID string) ([]*servicecatalog.LaunchPathSummary, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{StatusNotFound},
		Target:                    []string{servicecatalog.StatusAvailable},
		Refresh:                   StatusLaunchPaths(conn, acceptLanguage, productID),
		Timeout:                   LaunchPathsReadyTimeout,
		ContinuousTargetOccurence: ContinuousTargetOccurrence,
		NotFoundChecks:            NotFoundChecks,
		MinTimeout:                MinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.([]*servicecatalog.LaunchPathSummary); ok {
		return output, err
	}

	return nil, err
}

func WaitProvisionedProductReady(conn *servicecatalog.ServiceCatalog, acceptLanguage, id, name string) (*servicecatalog.DescribeProvisionedProductOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{StatusNotFound, StatusUnavailable, servicecatalog.ProvisionedProductStatusUnderChange, servicecatalog.ProvisionedProductStatusPlanInProgress},
		Target:                    []string{servicecatalog.StatusAvailable},
		Refresh:                   StatusProvisionedProduct(conn, acceptLanguage, id, name),
		Timeout:                   ProvisionedProductReadyTimeout,
		ContinuousTargetOccurence: ContinuousTargetOccurrence,
		NotFoundChecks:            NotFoundChecks,
		MinTimeout:                MinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*servicecatalog.DescribeProvisionedProductOutput); ok {
		return output, err
	}

	return nil, err
}

func WaitProvisionedProductTerminated(conn *servicecatalog.ServiceCatalog, acceptLanguage, id, name string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{servicecatalog.StatusAvailable, servicecatalog.ProvisionedProductStatusUnderChange},
		Target:  []string{StatusNotFound, StatusUnavailable},
		Refresh: StatusProvisionedProduct(conn, acceptLanguage, id, name),
		Timeout: ProvisionedProductDeleteTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}

func WaitRecordReady(conn *servicecatalog.ServiceCatalog, acceptLanguage, id string) (*servicecatalog.DescribeRecordOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{StatusNotFound, StatusUnavailable, servicecatalog.ProvisionedProductStatusUnderChange, servicecatalog.ProvisionedProductStatusPlanInProgress},
		Target:                    []string{servicecatalog.RecordStatusSucceeded, servicecatalog.StatusAvailable},
		Refresh:                   StatusRecord(conn, acceptLanguage, id),
		Timeout:                   RecordReadyTimeout,
		ContinuousTargetOccurence: ContinuousTargetOccurrence,
		NotFoundChecks:            NotFoundChecks,
		MinTimeout:                MinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*servicecatalog.DescribeRecordOutput); ok {
		return output, err
	}

	return nil, err
}

func WaitPortfolioConstraintsReady(conn *servicecatalog.ServiceCatalog, acceptLanguage, portfolioID, productID string) ([]*servicecatalog.ConstraintDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{StatusNotFound},
		Target:  []string{servicecatalog.StatusAvailable},
		Refresh: StatusPortfolioConstraints(conn, acceptLanguage, portfolioID, productID),
		Timeout: PortfolioConstraintsReadyTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.([]*servicecatalog.ConstraintDetail); ok {
		return output, err
	}

	return nil, err
}
