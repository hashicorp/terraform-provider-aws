package logging

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-log/internal/fieldutils"
	"github.com/hashicorp/terraform-plugin-log/internal/hclogutils"
)

const logMaskingReplacementString = "***"

// ShouldOmit takes a log's *string message and slices of fields,
// and determines, based on the LoggerOpts configuration, if the
// log should be omitted (i.e. prevent it to be printed on the final writer).
func (lo LoggerOpts) ShouldOmit(msg *string, fieldMaps ...map[string]interface{}) bool {
	// Omit log if any of the configured keys is found in the given fields
	if len(lo.OmitLogWithFieldKeys) > 0 {
		fieldsKeys := fieldutils.FieldMapsToKeys(fieldMaps...)
		if argKeysContain(fieldsKeys, lo.OmitLogWithFieldKeys) {
			return true
		}
	}

	// Omit log if any of the configured regexp matches the log message
	if len(lo.OmitLogWithMessageRegexes) > 0 {
		for _, r := range lo.OmitLogWithMessageRegexes {
			if r.MatchString(*msg) {
				return true
			}
		}
	}

	// Omit log if any of the configured strings is contained in the log message
	if len(lo.OmitLogWithMessageStrings) > 0 {
		for _, s := range lo.OmitLogWithMessageStrings {
			if strings.Contains(*msg, s) {
				return true
			}
		}
	}

	return false
}

// ApplyMask takes a log's *string message and slices of fields,
// and applies masking to fields keys' values and/or to log message,
// based on the LoggerOpts configuration.
//
// Note that the given input is changed-in-place by this method.
func (lo LoggerOpts) ApplyMask(msg *string, fieldMaps ...map[string]interface{}) {
	// Replace any log field value with the corresponding field key equal to the configured strings
	if len(lo.MaskFieldValuesWithFieldKeys) > 0 {
		for _, k := range lo.MaskFieldValuesWithFieldKeys {
			for _, f := range fieldMaps {
				for fk := range f {
					if k == fk {
						f[k] = logMaskingReplacementString
					}
				}
			}
		}
	}

	// Replace any part of any log field matching any of the configured regexp
	if len(lo.MaskAllFieldValuesRegexes) > 0 {
		for _, r := range lo.MaskAllFieldValuesRegexes {
			for _, f := range fieldMaps {
				for fk, fv := range f {
					// Can apply the regexp replacement, only if the field value is indeed a string
					fvStr, ok := fv.(string)
					if ok {
						f[fk] = r.ReplaceAllString(fvStr, logMaskingReplacementString)
					}
				}
			}
		}
	}

	// Replace any part of any log field matching any of the configured strings
	if len(lo.MaskAllFieldValuesStrings) > 0 {
		for _, s := range lo.MaskAllFieldValuesStrings {
			for _, f := range fieldMaps {
				for fk, fv := range f {
					// Can apply the regexp replacement, only if the field value is indeed a string
					fvStr, ok := fv.(string)
					if ok {
						f[fk] = strings.ReplaceAll(fvStr, s, logMaskingReplacementString)
					}
				}
			}
		}
	}

	// Replace any part of the log message matching any of the configured regexp
	if len(lo.MaskMessageRegexes) > 0 {
		for _, r := range lo.MaskMessageRegexes {
			*msg = r.ReplaceAllString(*msg, logMaskingReplacementString)
		}
	}

	// Replace any part of the log message equal to any of the configured strings
	if len(lo.MaskMessageStrings) > 0 {
		for _, s := range lo.MaskMessageStrings {
			*msg = strings.ReplaceAll(*msg, s, logMaskingReplacementString)
		}
	}
}

func OmitOrMask(tfLoggerOpts LoggerOpts, msg *string, additionalFields []map[string]interface{}) ([]interface{}, bool) {
	additionalFieldsMap := fieldutils.MergeFieldMaps(additionalFields...)

	// Apply the provider root LoggerOpts to determine if this log should be omitted
	if tfLoggerOpts.ShouldOmit(msg, tfLoggerOpts.Fields, additionalFieldsMap) {
		return nil, true
	}

	// Apply the provider root LoggerOpts to apply masking to this log
	tfLoggerOpts.ApplyMask(msg, tfLoggerOpts.Fields, additionalFieldsMap)

	return hclogutils.FieldMapsToArgs(tfLoggerOpts.Fields, additionalFieldsMap), false
}

func argKeysContain(haystack []string, needles []string) bool {
	for _, h := range haystack {
		for _, n := range needles {
			if n == h {
				return true
			}
		}
	}

	return false
}
