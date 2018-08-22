package aws

import (
	"testing"
)

func TestHostedZoneIDForRegion(t *testing.T) {
	if r, _ := HostedZoneIDForRegion("us-east-1"); r != "Z3AQBSTGFYJSTF" {
		t.Fatalf("bad: %s", r)
	}
	if r, _ := HostedZoneIDForRegion("ap-southeast-2"); r != "Z1WCIGYICN2BYD" {
		t.Fatalf("bad: %s", r)
	}

	// Bad input should be error
	if r, err := HostedZoneIDForRegion("not-a-region"); err == nil {
		t.Fatalf("bad: %s", r)
	}
}
