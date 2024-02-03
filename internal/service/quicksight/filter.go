// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"regexp"

	"github.com/aws/aws-sdk-go/service/quicksight"
)

func filterUser(user *quicksight.User, filter *dataSourceQuicksightUserFilter) (bool, error) {
	if !filter.Active.IsNull() {
		if *user.Active != filter.Active.ValueBool() {
			return true, nil
		}
	}

	if !filter.EmailRegex.IsNull() {
		regex, err := regexp.Compile(filter.EmailRegex.ValueString())
		if err != nil {
			return false, err
		}

		if !regex.MatchString(*user.Email) {
			return true, nil
		}
	}

	if !filter.IdentityType.IsNull() {
		if *user.IdentityType != filter.IdentityType.ValueString() {
			return true, nil
		}
	}

	if !filter.UserNameRegex.IsNull() {
		regex, err := regexp.Compile(filter.UserNameRegex.ValueString())
		if err != nil {
			return false, err
		}

		if !regex.MatchString(*user.UserName) {
			return true, nil
		}
	}

	if !filter.UserRole.IsNull() {
		if *user.Role != filter.UserRole.ValueString() {
			return true, nil
		}
	}

	return false, nil
}
