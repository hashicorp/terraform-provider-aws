// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// WithTimeouts is intended to be embedded in resources which use the special "timeouts" nested block.
// See https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts.
type WithTimeouts struct {
	defaultCreateTimeout, defaultReadTimeout, defaultUpdateTimeout, defaultDeleteTimeout time.Duration
}

// SetDefaultCreateTimeout sets the resource's default Create timeout value.
func (w *WithTimeouts) SetDefaultCreateTimeout(timeout time.Duration) {
	w.defaultCreateTimeout = timeout
}

// SetDefaultReadTimeout sets the resource's default Read timeout value.
func (w *WithTimeouts) SetDefaultReadTimeout(timeout time.Duration) {
	w.defaultReadTimeout = timeout
}

// SetDefaultUpdateTimeout sets the resource's default Update timeout value.
func (w *WithTimeouts) SetDefaultUpdateTimeout(timeout time.Duration) {
	w.defaultUpdateTimeout = timeout
}

// SetDefaultDeleteTimeout sets the resource's default Delete timeout value.
func (w *WithTimeouts) SetDefaultDeleteTimeout(timeout time.Duration) {
	w.defaultDeleteTimeout = timeout
}

// CreateTimeout returns any configured Create timeout value or the default value.
func (w *WithTimeouts) CreateTimeout(ctx context.Context, timeouts timeouts.Value) time.Duration {
	timeout, diags := timeouts.Create(ctx, w.defaultCreateTimeout)

	if errors := diags.Errors(); len(errors) > 0 {
		tflog.Warn(ctx, "reading configured Create timeout", map[string]interface{}{
			"summary": errors[0].Summary(),
			"detail":  errors[0].Detail(),
		})

		return w.defaultCreateTimeout
	}

	return timeout
}

// ReadTimeout returns any configured Read timeout value or the default value.
func (w *WithTimeouts) ReadTimeout(ctx context.Context, timeouts timeouts.Value) time.Duration {
	timeout, diags := timeouts.Read(ctx, w.defaultReadTimeout)

	if errors := diags.Errors(); len(errors) > 0 {
		tflog.Warn(ctx, "reading configured Read timeout", map[string]interface{}{
			"summary": errors[0].Summary(),
			"detail":  errors[0].Detail(),
		})

		return w.defaultReadTimeout
	}

	return timeout
}

// UpdateTimeout returns any configured Update timeout value or the default value.
func (w *WithTimeouts) UpdateTimeout(ctx context.Context, timeouts timeouts.Value) time.Duration {
	timeout, diags := timeouts.Update(ctx, w.defaultUpdateTimeout)

	if errors := diags.Errors(); len(errors) > 0 {
		tflog.Warn(ctx, "reading configured Update timeout", map[string]interface{}{
			"summary": errors[0].Summary(),
			"detail":  errors[0].Detail(),
		})

		return w.defaultUpdateTimeout
	}

	return timeout
}

// DeleteTimeout returns any configured Delete timeout value or the default value.
func (w *WithTimeouts) DeleteTimeout(ctx context.Context, timeouts timeouts.Value) time.Duration {
	timeout, diags := timeouts.Delete(ctx, w.defaultDeleteTimeout)

	if errors := diags.Errors(); len(errors) > 0 {
		tflog.Warn(ctx, "reading configured Delete timeout", map[string]interface{}{
			"summary": errors[0].Summary(),
			"detail":  errors[0].Detail(),
		})

		return w.defaultDeleteTimeout
	}

	return timeout
}
