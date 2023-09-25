// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/types/nullable"
)

func ExpandLifecycleRuleAbortIncompleteMultipartUpload(m map[string]interface{}) *s3.AbortIncompleteMultipartUpload {
	if len(m) == 0 {
		return nil
	}

	result := &s3.AbortIncompleteMultipartUpload{}

	if v, ok := m["days_after_initiation"].(int); ok {
		result.DaysAfterInitiation = aws.Int64(int64(v))
	}

	return result
}

func ExpandLifecycleRuleExpiration(l []interface{}) (*s3.LifecycleExpiration, error) {
	if len(l) == 0 {
		return nil, nil
	}

	result := &s3.LifecycleExpiration{}

	if l[0] == nil {
		return result, nil
	}

	m := l[0].(map[string]interface{})

	if v, ok := m["date"].(string); ok && v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return nil, fmt.Errorf("parsing S3 Bucket Lifecycle Rule Expiration date: %w", err)
		}
		result.Date = aws.Time(t)
	}

	if v, ok := m["days"].(int); ok && v > 0 {
		result.Days = aws.Int64(int64(v))
	}

	// This cannot be specified with Days or Date
	if v, ok := m["expired_object_delete_marker"].(bool); ok && result.Date == nil && result.Days == nil {
		result.ExpiredObjectDeleteMarker = aws.Bool(v)
	}

	return result, nil
}

// ExpandLifecycleRuleFilter ensures a Filter can have only 1 of prefix, tag, or and
func ExpandLifecycleRuleFilter(ctx context.Context, l []interface{}) *s3.LifecycleRuleFilter {
	if len(l) == 0 {
		return nil
	}

	result := &s3.LifecycleRuleFilter{}

	if l[0] == nil {
		return result
	}

	m := l[0].(map[string]interface{})

	if v, ok := m["and"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.And = ExpandLifecycleRuleFilterAndOperator(ctx, v[0].(map[string]interface{}))
	}

	if v, null, _ := nullable.Int(m["object_size_greater_than"].(string)).Value(); !null && v >= 0 {
		result.ObjectSizeGreaterThan = aws.Int64(v)
	}

	if v, null, _ := nullable.Int(m["object_size_less_than"].(string)).Value(); !null && v > 0 {
		result.ObjectSizeLessThan = aws.Int64(v)
	}

	if v, ok := m["tag"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.Tag = ExpandLifecycleRuleFilterTag(v[0].(map[string]interface{}))
	}

	// Per AWS S3 API, "A Filter must have exactly one of Prefix, Tag, or And specified";
	// Specifying more than one of the listed parameters results in a MalformedXML error.
	// In practice, this also includes ObjectSizeGreaterThan and ObjectSizeLessThan.
	if v, ok := m["prefix"].(string); ok && result.And == nil && result.Tag == nil && result.ObjectSizeGreaterThan == nil && result.ObjectSizeLessThan == nil {
		result.Prefix = aws.String(v)
	}

	return result
}

func ExpandLifecycleRuleFilterAndOperator(ctx context.Context, m map[string]interface{}) *s3.LifecycleRuleAndOperator {
	if len(m) == 0 {
		return nil
	}

	result := &s3.LifecycleRuleAndOperator{}

	if v, ok := m["object_size_greater_than"].(int); ok && v > 0 {
		result.ObjectSizeGreaterThan = aws.Int64(int64(v))
	}

	if v, ok := m["object_size_less_than"].(int); ok && v > 0 {
		result.ObjectSizeLessThan = aws.Int64(int64(v))
	}

	if v, ok := m["prefix"].(string); ok {
		result.Prefix = aws.String(v)
	}

	if v, ok := m["tags"].(map[string]interface{}); ok && len(v) > 0 {
		tags := Tags(tftags.New(ctx, v).IgnoreAWS())
		if len(tags) > 0 {
			result.Tags = tags
		}
	}

	return result
}

func ExpandLifecycleRuleFilterTag(m map[string]interface{}) *s3.Tag {
	if len(m) == 0 {
		return nil
	}

	result := &s3.Tag{}

	if key, ok := m["key"].(string); ok {
		result.Key = aws.String(key)
	}

	if value, ok := m["value"].(string); ok {
		result.Value = aws.String(value)
	}

	return result
}

func ExpandLifecycleRuleNoncurrentVersionExpiration(m map[string]interface{}) *s3.NoncurrentVersionExpiration {
	if len(m) == 0 {
		return nil
	}

	result := &s3.NoncurrentVersionExpiration{}

	if v, null, _ := nullable.Int(m["newer_noncurrent_versions"].(string)).Value(); !null && v > 0 {
		result.NewerNoncurrentVersions = aws.Int64(v)
	}

	if v, ok := m["noncurrent_days"].(int); ok {
		result.NoncurrentDays = aws.Int64(int64(v))
	}

	return result
}

func ExpandLifecycleRuleNoncurrentVersionTransitions(l []interface{}) []*s3.NoncurrentVersionTransition {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	var results []*s3.NoncurrentVersionTransition

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		transition := &s3.NoncurrentVersionTransition{}

		if v, null, _ := nullable.Int(tfMap["newer_noncurrent_versions"].(string)).Value(); !null && v > 0 {
			transition.NewerNoncurrentVersions = aws.Int64(v)
		}

		if v, ok := tfMap["noncurrent_days"].(int); ok {
			transition.NoncurrentDays = aws.Int64(int64(v))
		}

		if v, ok := tfMap["storage_class"].(string); ok && v != "" {
			transition.StorageClass = aws.String(v)
		}

		results = append(results, transition)
	}

	return results
}

func ExpandLifecycleRuleTransitions(l []interface{}) ([]*s3.Transition, error) {
	if len(l) == 0 || l[0] == nil {
		return nil, nil
	}

	var results []*s3.Transition

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		transition := &s3.Transition{}

		if v, ok := tfMap["date"].(string); ok && v != "" {
			t, err := time.Parse(time.RFC3339, v)
			if err != nil {
				return nil, fmt.Errorf("parsing S3 Bucket Lifecycle Rule Transition date: %w", err)
			}
			transition.Date = aws.Time(t)
		}

		// Only one of "date" and "days" can be configured
		// so only set the transition.Days value when transition.Date is nil
		// By default, tfMap["days"] = 0 if not explicitly configured in terraform.
		if v, ok := tfMap["days"].(int); ok && v >= 0 && transition.Date == nil {
			transition.Days = aws.Int64(int64(v))
		}

		if v, ok := tfMap["storage_class"].(string); ok && v != "" {
			transition.StorageClass = aws.String(v)
		}

		results = append(results, transition)
	}

	return results, nil
}

func ExpandLifecycleRules(ctx context.Context, l []interface{}) ([]*s3.LifecycleRule, error) {
	if len(l) == 0 || l[0] == nil {
		return nil, nil
	}

	var results []*s3.LifecycleRule

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		result := &s3.LifecycleRule{}

		if v, ok := tfMap["abort_incomplete_multipart_upload"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			result.AbortIncompleteMultipartUpload = ExpandLifecycleRuleAbortIncompleteMultipartUpload(v[0].(map[string]interface{}))
		}

		if v, ok := tfMap["expiration"].([]interface{}); ok && len(v) > 0 {
			expiration, err := ExpandLifecycleRuleExpiration(v)
			if err != nil {
				return nil, err
			}
			result.Expiration = expiration
		}

		if v, ok := tfMap["filter"].([]interface{}); ok && len(v) > 0 {
			result.Filter = ExpandLifecycleRuleFilter(ctx, v)
		}

		if v, ok := tfMap["prefix"].(string); ok && result.Filter == nil {
			// If neither the filter block nor the prefix are specified,
			// apply the Default behavior from v3.x of the provider;
			// otherwise, set the prefix as specified in Terraform.
			if v == "" {
				result.SetFilter(&s3.LifecycleRuleFilter{
					Prefix: aws.String(v),
				})
			} else {
				result.Prefix = aws.String(v)
			}
		}

		if v, ok := tfMap["id"].(string); ok {
			result.ID = aws.String(v)
		}

		if v, ok := tfMap["noncurrent_version_expiration"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			result.NoncurrentVersionExpiration = ExpandLifecycleRuleNoncurrentVersionExpiration(v[0].(map[string]interface{}))
		}

		if v, ok := tfMap["noncurrent_version_transition"].(*schema.Set); ok && v.Len() > 0 {
			result.NoncurrentVersionTransitions = ExpandLifecycleRuleNoncurrentVersionTransitions(v.List())
		}

		if v, ok := tfMap["status"].(string); ok && v != "" {
			result.Status = aws.String(v)
		}

		if v, ok := tfMap["transition"].(*schema.Set); ok && v.Len() > 0 {
			transitions, err := ExpandLifecycleRuleTransitions(v.List())
			if err != nil {
				return nil, err
			}
			result.Transitions = transitions
		}

		results = append(results, result)
	}

	return results, nil
}

func FlattenLifecycleRules(ctx context.Context, rules []*s3.LifecycleRule) []interface{} {
	if len(rules) == 0 {
		return []interface{}{}
	}

	var results []interface{}

	for _, rule := range rules {
		if rule == nil {
			continue
		}

		m := make(map[string]interface{})

		if rule.AbortIncompleteMultipartUpload != nil {
			m["abort_incomplete_multipart_upload"] = FlattenLifecycleRuleAbortIncompleteMultipartUpload(rule.AbortIncompleteMultipartUpload)
		}

		if rule.Expiration != nil {
			m["expiration"] = FlattenLifecycleRuleExpiration(rule.Expiration)
		}

		if rule.Filter != nil {
			m["filter"] = FlattenLifecycleRuleFilter(ctx, rule.Filter)
		}

		if rule.ID != nil {
			m["id"] = aws.StringValue(rule.ID)
		}

		if rule.NoncurrentVersionExpiration != nil {
			m["noncurrent_version_expiration"] = FlattenLifecycleRuleNoncurrentVersionExpiration(rule.NoncurrentVersionExpiration)
		}

		if rule.NoncurrentVersionTransitions != nil {
			m["noncurrent_version_transition"] = FlattenLifecycleRuleNoncurrentVersionTransitions(rule.NoncurrentVersionTransitions)
		}

		if rule.Prefix != nil {
			m["prefix"] = aws.StringValue(rule.Prefix)
		}

		if rule.Status != nil {
			m["status"] = aws.StringValue(rule.Status)
		}

		if rule.Transitions != nil {
			m["transition"] = FlattenLifecycleRuleTransitions(rule.Transitions)
		}

		results = append(results, m)
	}

	return results
}

func FlattenLifecycleRuleAbortIncompleteMultipartUpload(u *s3.AbortIncompleteMultipartUpload) []interface{} {
	if u == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if u.DaysAfterInitiation != nil {
		m["days_after_initiation"] = int(aws.Int64Value(u.DaysAfterInitiation))
	}

	return []interface{}{m}
}

func FlattenLifecycleRuleExpiration(expiration *s3.LifecycleExpiration) []interface{} {
	if expiration == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if expiration.Days != nil {
		m["days"] = int(aws.Int64Value(expiration.Days))
	}

	if expiration.Date != nil {
		m["date"] = expiration.Date.Format(time.RFC3339)
	}

	if expiration.ExpiredObjectDeleteMarker != nil {
		m["expired_object_delete_marker"] = aws.BoolValue(expiration.ExpiredObjectDeleteMarker)
	}

	return []interface{}{m}
}

func FlattenLifecycleRuleFilter(ctx context.Context, filter *s3.LifecycleRuleFilter) []interface{} {
	if filter == nil {
		return nil
	}

	m := make(map[string]interface{})

	if filter.And != nil {
		m["and"] = FlattenLifecycleRuleFilterAndOperator(ctx, filter.And)
	}

	if filter.ObjectSizeGreaterThan != nil {
		m["object_size_greater_than"] = strconv.FormatInt(aws.Int64Value(filter.ObjectSizeGreaterThan), 10)
	}

	if filter.ObjectSizeLessThan != nil {
		m["object_size_less_than"] = strconv.FormatInt(aws.Int64Value(filter.ObjectSizeLessThan), 10)
	}

	if filter.Prefix != nil {
		m["prefix"] = aws.StringValue(filter.Prefix)
	}

	if filter.Tag != nil {
		m["tag"] = FlattenLifecycleRuleFilterTag(filter.Tag)
	}

	return []interface{}{m}
}

func FlattenLifecycleRuleFilterAndOperator(ctx context.Context, andOp *s3.LifecycleRuleAndOperator) []interface{} {
	if andOp == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if andOp.ObjectSizeGreaterThan != nil {
		m["object_size_greater_than"] = int(aws.Int64Value(andOp.ObjectSizeGreaterThan))
	}

	if andOp.ObjectSizeLessThan != nil {
		m["object_size_less_than"] = int(aws.Int64Value(andOp.ObjectSizeLessThan))
	}

	if andOp.Prefix != nil {
		m["prefix"] = aws.StringValue(andOp.Prefix)
	}

	if andOp.Tags != nil {
		m["tags"] = KeyValueTags(ctx, andOp.Tags).IgnoreAWS().Map()
	}

	return []interface{}{m}
}

func FlattenLifecycleRuleFilterTag(tag *s3.Tag) []interface{} {
	if tag == nil {
		return nil
	}

	m := make(map[string]interface{})

	if tag.Key != nil {
		m["key"] = aws.StringValue(tag.Key)
	}

	if tag.Value != nil {
		m["value"] = aws.StringValue(tag.Value)
	}

	return []interface{}{m}
}

func FlattenLifecycleRuleNoncurrentVersionExpiration(expiration *s3.NoncurrentVersionExpiration) []interface{} {
	if expiration == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if expiration.NewerNoncurrentVersions != nil {
		m["newer_noncurrent_versions"] = strconv.FormatInt(aws.Int64Value(expiration.NewerNoncurrentVersions), 10)
	}

	if expiration.NoncurrentDays != nil {
		m["noncurrent_days"] = int(aws.Int64Value(expiration.NoncurrentDays))
	}

	return []interface{}{m}
}

func FlattenLifecycleRuleNoncurrentVersionTransitions(transitions []*s3.NoncurrentVersionTransition) []interface{} {
	if len(transitions) == 0 {
		return []interface{}{}
	}

	var results []interface{}

	for _, transition := range transitions {
		if transition == nil {
			continue
		}

		m := make(map[string]interface{})

		if transition.NewerNoncurrentVersions != nil {
			m["newer_noncurrent_versions"] = strconv.FormatInt(aws.Int64Value(transition.NewerNoncurrentVersions), 10)
		}

		if transition.NoncurrentDays != nil {
			m["noncurrent_days"] = int(aws.Int64Value(transition.NoncurrentDays))
		}

		if transition.StorageClass != nil {
			m["storage_class"] = aws.StringValue(transition.StorageClass)
		}

		results = append(results, m)
	}

	return results
}

func FlattenLifecycleRuleTransitions(transitions []*s3.Transition) []interface{} {
	if len(transitions) == 0 {
		return []interface{}{}
	}

	var results []interface{}

	for _, transition := range transitions {
		if transition == nil {
			continue
		}

		m := make(map[string]interface{})

		if transition.Date != nil {
			m["date"] = transition.Date.Format(time.RFC3339)
		}

		if transition.Days != nil {
			m["days"] = int(aws.Int64Value(transition.Days))
		}

		if transition.StorageClass != nil {
			m["storage_class"] = aws.StringValue(transition.StorageClass)
		}

		results = append(results, m)
	}

	return results
}
