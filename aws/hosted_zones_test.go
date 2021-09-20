package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
)

func TestHostedZoneIDForRegion(t *testing.T) {
	if r, _ := HostedZoneIDForRegion(endpoints.UsEast1RegionID); r != "Z3AQBSTGFYJSTF" {
		t.Fatalf("bad: %s", r)
	}
	if r, _ := HostedZoneIDForRegion(endpoints.ApSoutheast2RegionID); r != "Z1WCIGYICN2BYD" {
		t.Fatalf("bad: %s", r)
	}

	// Bad input should be error
	if r, err := HostedZoneIDForRegion("not-a-region"); err == nil {
		t.Fatalf("bad: %s", r)
	}
}
