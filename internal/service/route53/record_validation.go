// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"
	"fmt"
	"strings"

	awstypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// validateTXTRecordStringLength is a CustomizeDiff function that checks each
// quoted string segment within a TXT or SPF record value does not exceed the
// 255-character limit imposed by RFC 4408 and the Route 53 API.
func validateTXTRecordStringLength(_ context.Context, d *schema.ResourceDiff, _ any) error {
	rrType := awstypes.RRType(d.Get(names.AttrType).(string))
	if rrType != awstypes.RRTypeTxt && rrType != awstypes.RRTypeSpf {
		return nil
	}

	records, ok := d.GetOk("records")
	if !ok {
		return nil
	}

	for _, v := range records.(*schema.Set).List() {
		if err := validateTXTRecordStringSegments(v.(string)); err != nil {
			return err
		}
	}

	return nil
}

// validateTXTRecordStringSegments checks that each quoted-string segment of a
// TXT record value is at most 255 characters. Users who need longer values
// must split them into multiple quoted strings using the quote-space-quote
// separator in their record value. Segments that are already split this way
// pass validation.
func validateTXTRecordStringSegments(value string) error {
	// Split on `" "` which is how users represent multi-string TXT entries in
	// Terraform. Each segment between the quote-space-quote boundaries maps to
	// an individual quoted string in the Route 53 API and is subject to the
	// 255-character limit.
	segments := strings.Split(value, "\" \"")
	for _, seg := range segments {
		if len(seg) > 255 {
			return fmt.Errorf(
				"TXT record value contains a string segment longer than 255 characters (%d): "+
					"per RFC 4408, each quoted string in a TXT record must be 255 characters or fewer, "+
					"split the value into multiple strings using the '\\\" \\\"' separator in your record value "+
					"(e.g. \"first_255_chars\\\" \\\"remaining_chars\")", len(seg))
		}
	}
	return nil
}
