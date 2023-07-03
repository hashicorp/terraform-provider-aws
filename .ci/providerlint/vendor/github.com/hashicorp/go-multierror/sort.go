// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package multierror

// Len implements sort.Interface function for length
func (err Error) Len() int {
	return len(err.Errors)
}

// Swap implements sort.Interface function for swapping elements
func (err Error) Swap(i, j int) {
	err.Errors[i], err.Errors[j] = err.Errors[j], err.Errors[i]
}

// Less implements sort.Interface function for determining order
func (err Error) Less(i, j int) bool {
	return err.Errors[i].Error() < err.Errors[j].Error()
}
