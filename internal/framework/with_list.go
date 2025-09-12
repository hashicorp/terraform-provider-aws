// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/listresource"
)

type Lister interface {
	AppendResultInterceptor(listresource.ListResultInterceptor)
}

var _ Lister = &WithList{}

type WithList struct {
	interceptors []listresource.ListResultInterceptor
}

func (w *WithList) AppendResultInterceptor(interceptor listresource.ListResultInterceptor) {
	w.interceptors = append(w.interceptors, interceptor)
}

func (w WithList) ResultInterceptors() []listresource.ListResultInterceptor {
	return w.interceptors
}
