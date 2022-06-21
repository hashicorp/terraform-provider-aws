package s3_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
)

func TestHostedZoneIDForRegion(t *testing.T) {
	if r, _ := tfs3.HostedZoneIDForRegion(endpoints.UsEast1RegionID); r != "Z3AQBSTGFYJSTF" {
		t.Fatalf("bad: %s", r)
	}
	if r, _ := tfs3.HostedZoneIDForRegion(endpoints.ApSoutheast2RegionID); r != "Z1WCIGYICN2BYD" {
		t.Fatalf("bad: %s", r)
	}

	// Bad input should be error
	if r, err := tfs3.HostedZoneIDForRegion("not-a-region"); err == nil {
		t.Fatalf("bad: %s", r)
	}
}
