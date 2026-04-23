// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package secretsmanager

import (
	"testing"
	"time"
)

func TestResourceSecretVersion_Timeouts(t *testing.T) {
	r := resourceSecretVersion()

	if r.Timeouts == nil {
		t.Fatal("expected resource timeouts to be configured")
	}

	if got, want := r.Timeouts.Create, 2*time.Minute; got == nil || *got != want {
		t.Fatalf("expected create timeout %s, got %#v", want, got)
	}

	if got, want := r.Timeouts.Delete, 2*time.Minute; got == nil || *got != want {
		t.Fatalf("expected delete timeout %s, got %#v", want, got)
	}

	if got := r.Timeouts.Update; got != nil {
		t.Fatalf("expected update timeout to be unset, got %#v", got)
	}

	if got := r.Timeouts.Read; got != nil {
		t.Fatalf("expected read timeout to be unset, got %#v", got)
	}
}
