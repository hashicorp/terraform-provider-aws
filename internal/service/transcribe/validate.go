// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transcribe

import (
	"github.com/aws/aws-sdk-go-v2/service/transcribe/types"
)

func validateLanguageCodes(t []types.LanguageCode) []string {
	var out []string

	for _, v := range t {
		out = append(out, string(v))
	}

	return out
}
