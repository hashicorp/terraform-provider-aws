// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/action/timeouts"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type ActionWithTimeouts struct {
	defaultInvokeTimeout time.Duration
}

func (w *ActionWithTimeouts) SetDefaultInvokeTimeout(timeout time.Duration) {
	w.defaultInvokeTimeout = timeout
}

// InvokeTimeout returns any configured Invoke timeout value or the default value.
func (w *ActionWithTimeouts) InvokeTimeout(ctx context.Context, timeouts timeouts.Value) time.Duration {
	timeout, diags := timeouts.Invoke(ctx, w.defaultInvokeTimeout)

	if errors := diags.Errors(); len(errors) > 0 {
		tflog.Warn(ctx, "reading configured Invoke timeout", map[string]any{
			"summary": errors[0].Summary(),
			"detail":  errors[0].Detail(),
		})
	}

	return timeout
}
