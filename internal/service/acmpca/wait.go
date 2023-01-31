package acmpca

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// waitCertificateAuthorityCreated waits for a CertificateAuthority to return Active or PendingCertificate
func waitCertificateAuthorityCreated(ctx context.Context, conn *acmpca.ACMPCA, arn string, timeout time.Duration) (*acmpca.CertificateAuthority, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"", acmpca.CertificateAuthorityStatusCreating},
		Target:  []string{acmpca.CertificateAuthorityStatusActive, acmpca.CertificateAuthorityStatusPendingCertificate},
		Refresh: statusCertificateAuthority(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*acmpca.CertificateAuthority); ok {
		return v, err
	}

	return nil, err
}

const (
	certificateAuthorityActiveTimeout = 1 * time.Minute
)
