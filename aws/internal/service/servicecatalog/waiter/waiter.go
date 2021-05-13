package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	ProductReadyTimeout  = 3 * time.Minute
	ProductDeleteTimeout = 3 * time.Minute

	TagOptionReadyTimeout  = 3 * time.Minute
	TagOptionDeleteTimeout = 3 * time.Minute

	StatusNotFound    = "NOT_FOUND"
	StatusUnavailable = "UNAVAILABLE"

	// AWS documentation is wrong, says that status will be "AVAILABLE" but it is actually "CREATED"
	ProductStatusCreated = "CREATED"
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
		Timeout: ProductReadyTimeout,
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
