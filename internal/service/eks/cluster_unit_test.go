// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceClusterFlattenCertificate(t *testing.T) {
	t.Parallel()

	r := resourceCluster()
	d := r.Data(&terraform.InstanceState{
		ID: "example",
		Attributes: map[string]string{
			"certificate_authority.#":      "1",
			"certificate_authority.0.data": "encoded-certificate",
		},
	})
	cluster := &types.Cluster{
		CreatedAt: aws.Time(time.Unix(0, 0)),
		CertificateAuthority: &types.Certificate{
			Data: aws.String("encoded-certificate"),
			Active: &types.ActiveCertificateAuthority{
				ActivatedBy: types.CertificateAuthorityActivatedByEks,
				Id:          aws.String("ca-example"),
			},
		},
	}

	if err := resourceClusterFlatten(t.Context(), cluster, d); err != nil {
		t.Fatal(err)
	}
	state := d.State()
	for key, want := range map[string]string{
		"certificate_authority.0.active.0.activated_by": "EKS",
		"certificate_authority.0.active.0.id":           "ca-example",
		"certificate_authority.0.data":                  "encoded-certificate",
	} {
		if got := state.Attributes[key]; got != want {
			t.Errorf("%s = %q, want %q", key, got, want)
		}
	}

	d = r.Data(state)
	cluster.CertificateAuthority.Active = nil
	if err := resourceClusterFlatten(t.Context(), cluster, d); err != nil {
		t.Fatal(err)
	}
	if got := d.Get("certificate_authority.0.active.#"); got != 0 {
		t.Errorf("active certificate authority count = %v, want 0", got)
	}
	if got := d.Get("certificate_authority.0.data"); got != "encoded-certificate" {
		t.Errorf("certificate data = %v, want encoded-certificate", got)
	}
}
