// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package wafv2

// Schema attribute name constants used across package
const (
	attrActionAllow                       = "allow"
	attrActionBlock                       = "block"
	attrActionCaptcha                     = "captcha"
	attrActionChallenge                   = "challenge"
	attrActionCount                       = "count"
	attrActionNone                        = "none"
	attrASNMatchStatement                 = "asn_match_statement"
	attrByteMatchStatement                = "byte_match_statement"
	attrCustomRequestHandling             = "custom_request_handling"
	attrFallbackBehavior                  = "fallback_behavior"
	attrFieldToMatch                      = "field_to_match"
	attrGeoMatchStatement                 = "geo_match_statement"
	attrIPSetReferenceStatement           = "ip_set_reference_statement"
	attrLabelMatchStatement               = "label_match_statement"
	attrOversizeHandling                  = "oversize_handling"
	attrRegexMatchStatement               = "regex_match_statement"
	attrRegexPatternSetReferenceStatement = "regex_pattern_set_reference_statement"
	attrSizeConstraintStatement           = "size_constraint_statement"
	attrSQLiMatchStatement                = "sqli_match_statement"
	attrStatement                         = "statement"
	attrTextTransformation                = "text_transformation"
	attrXSSMatchStatement                 = "xss_match_statement"
)
