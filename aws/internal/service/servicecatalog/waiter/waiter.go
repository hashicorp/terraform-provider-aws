package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

const (
	ProductReadyTimeout  = 3 * time.Minute
	ProductDeleteTimeout = 3 * time.Minute

	TagOptionReadyTimeout  = 3 * time.Minute
	TagOptionDeleteTimeout = 3 * time.Minute

	PortfolioShareCreateTimeout = 3 * time.Minute

	OrganizationsAccessStableTimeout = 3 * time.Minute

	StatusNotFound    = "NOT_FOUND"
	StatusUnavailable = "UNAVAILABLE"

	// AWS documentation is wrong, says that status will be "AVAILABLE" but it is actually "CREATED"
	ProductStatusCreated = "CREATED"

	OrganizationAccessStatusError = "ERROR"
)

func ProductReady(conn *servicecatalog.ServiceCatalog, acceptLanguage, productID string) (*servicecatalog.DescribeProductAsAdminOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{servicecatalog.StatusCreating, StatusNotFound, StatusUnavailable},
		Target:  []string{servicecatalog.StatusAvailable, ProductStatusCreated},
		Refresh: ProductStatus(conn, acceptLanguage, productID),
		Timeout: ProductReadyTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*servicecatalog.DescribeProductAsAdminOutput); ok {
		return output, err
	}

	return nil, err
}

func ProductDeleted(conn *servicecatalog.ServiceCatalog, acceptLanguage, productID string) (*servicecatalog.DescribeProductAsAdminOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{servicecatalog.StatusCreating, servicecatalog.StatusAvailable, ProductStatusCreated, StatusUnavailable},
		Target:  []string{StatusNotFound},
		Refresh: ProductStatus(conn, acceptLanguage, productID),
		Timeout: ProductDeleteTimeout,
	}

	_, err := stateConf.WaitForState()

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return nil, nil
	}

	return nil, err
}

func TagOptionReady(conn *servicecatalog.ServiceCatalog, id string) (*servicecatalog.TagOptionDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{StatusNotFound, StatusUnavailable},
		Target:  []string{servicecatalog.StatusAvailable},
		Refresh: TagOptionStatus(conn, id),
		Timeout: TagOptionReadyTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*servicecatalog.TagOptionDetail); ok {
		return output, err
	}

	return nil, err
}

func TagOptionDeleted(conn *servicecatalog.ServiceCatalog, id string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{servicecatalog.StatusAvailable},
		Target:  []string{StatusNotFound, StatusUnavailable},
		Refresh: TagOptionStatus(conn, id),
		Timeout: TagOptionDeleteTimeout,
	}

	_, err := stateConf.WaitForState()

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return nil
	}

	return err
}

func PortfolioShareReady(conn *servicecatalog.ServiceCatalog, portfolioID, shareType, principalID string, acceptRequired bool) (*servicecatalog.PortfolioShareDetail, error) {
	targets := []string{servicecatalog.ShareStatusCompleted}

	if !acceptRequired {
		targets = append(targets, servicecatalog.ShareStatusInProgress)
	}

	stateConf := &resource.StateChangeConf{
		Pending: []string{servicecatalog.ShareStatusNotStarted, servicecatalog.ShareStatusInProgress, StatusNotFound, StatusUnavailable},
		Target:  targets,
		Refresh: PortfolioShareStatus(conn, portfolioID, shareType, principalID),
		Timeout: PortfolioShareCreateTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*servicecatalog.PortfolioShareDetail); ok {
		return output, err
	}

	return nil, err
}

func PortfolioShareCreatedWithToken(conn *servicecatalog.ServiceCatalog, token string, acceptRequired bool) (*servicecatalog.DescribePortfolioShareStatusOutput, error) {
	targets := []string{servicecatalog.ShareStatusCompleted}

	if !acceptRequired {
		targets = append(targets, servicecatalog.ShareStatusInProgress)
	}

	stateConf := &resource.StateChangeConf{
		Pending: []string{servicecatalog.ShareStatusNotStarted, servicecatalog.ShareStatusInProgress, StatusNotFound, StatusUnavailable},
		Target:  targets,
		Refresh: PortfolioShareStatusWithToken(conn, token),
		Timeout: PortfolioShareCreateTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*servicecatalog.DescribePortfolioShareStatusOutput); ok {
		return output, err
	}

	return nil, err
}

func PortfolioShareDeleted(conn *servicecatalog.ServiceCatalog, portfolioID, shareType, principalID string) (*servicecatalog.PortfolioShareDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{servicecatalog.ShareStatusNotStarted, servicecatalog.ShareStatusInProgress, servicecatalog.ShareStatusCompleted, StatusUnavailable},
		Target:  []string{StatusNotFound},
		Refresh: PortfolioShareStatus(conn, portfolioID, shareType, principalID),
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

func PortfolioShareDeletedWithToken(conn *servicecatalog.ServiceCatalog, token string) (*servicecatalog.DescribePortfolioShareStatusOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{servicecatalog.ShareStatusNotStarted, servicecatalog.ShareStatusInProgress, StatusNotFound, StatusUnavailable},
		Target:  []string{servicecatalog.ShareStatusCompleted},
		Refresh: PortfolioShareStatusWithToken(conn, token),
		Timeout: PortfolioShareCreateTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*servicecatalog.DescribePortfolioShareStatusOutput); ok {
		return output, err
	}

	return nil, err
}

func OrganizationsAccessStable(conn *servicecatalog.ServiceCatalog) (string, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{servicecatalog.AccessStatusUnderChange, StatusNotFound, StatusUnavailable},
		Target:  []string{servicecatalog.AccessStatusEnabled, servicecatalog.AccessStatusDisabled},
		Refresh: OrganizationsAccessStatus(conn),
		Timeout: OrganizationsAccessStableTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*servicecatalog.GetAWSOrganizationsAccessStatusOutput); ok {
		return aws.StringValue(output.AccessStatus), err
	}

	return "", err
}
