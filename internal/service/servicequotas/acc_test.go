package servicequotas_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicequotas"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

const (
	setQuotaServiceCode = "vpc"
	setQuotaQuotaCode   = "L-F678F1CE"
	setQuotaQuotaName   = "VPCs per Region"

	unsetQuotaServiceCode = "s3"
	unsetQuotaQuotaCode   = "L-FAABEEBA"
	unsetQuotaQuotaName   = "Access Points"
)

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceQuotasConn

	input := &servicequotas.ListServicesInput{}

	_, err := conn.ListServices(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

// nosemgrep: servicequotas-in-func-name
func preCheckServiceQuotaSet(serviceCode, quotaCode string, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceQuotasConn

	input := &servicequotas.GetServiceQuotaInput{
		QuotaCode:   aws.String(quotaCode),
		ServiceCode: aws.String(serviceCode),
	}

	_, err := conn.GetServiceQuota(input)
	if tfawserr.ErrCodeEquals(err, servicequotas.ErrCodeNoSuchResourceException) {
		t.Fatalf("The Service Quota (%s/%s) has never been set. This test can only be run with a quota that has previously been set. Please update the test to check a new quota.", serviceCode, quotaCode)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error getting Service Quota (%s/%s) : %s", serviceCode, quotaCode, err)
	}
}

func preCheckServiceQuotaUnset(serviceCode, quotaCode string, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceQuotasConn

	input := &servicequotas.GetServiceQuotaInput{
		QuotaCode:   aws.String(quotaCode),
		ServiceCode: aws.String(serviceCode),
	}

	_, err := conn.GetServiceQuota(input)
	if err == nil {
		t.Fatalf("The Service Quota (%s/%s) has been set. This test can only be run with a quota that has never been set. Please update the test to check a new quota.", serviceCode, quotaCode)
	}
	if !tfawserr.ErrCodeEquals(err, servicequotas.ErrCodeNoSuchResourceException) {
		t.Fatalf("unexpected PreCheck error getting Service Quota (%s/%s) : %s", serviceCode, quotaCode, err)
	}
}
