package s3

import (
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/experimental/nullable"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func ExpandReplicationRuleDestinationAccessControlTranslation(l []interface{}) *s3.AccessControlTranslation {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &s3.AccessControlTranslation{}

	if v, ok := tfMap["owner"].(string); ok && v != "" {
		result.Owner = aws.String(v)
	}

	return result
}

func ExpandReplicationRuleDestinationEncryptionConfiguration(l []interface{}) *s3.EncryptionConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &s3.EncryptionConfiguration{}

	if v, ok := tfMap["replica_kms_key_id"].(string); ok && v != "" {
		result.ReplicaKmsKeyID = aws.String(v)
	}

	return result
}

func ExpandReplicationRuleDeleteMarkerReplication(l []interface{}) *s3.DeleteMarkerReplication {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &s3.DeleteMarkerReplication{}

	if v, ok := tfMap["status"].(string); ok && v != "" {
		result.Status = aws.String(v)
	}

	return result
}

func ExpandReplicationRuleDestination(l []interface{}) *s3.Destination {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &s3.Destination{}

	if v, ok := tfMap["access_control_translation"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.AccessControlTranslation = ExpandReplicationRuleDestinationAccessControlTranslation(v)
	}

	if v, ok := tfMap["account"].(string); ok && v != "" {
		result.Account = aws.String(v)
	}

	if v, ok := tfMap["bucket"].(string); ok && v != "" {
		result.Bucket = aws.String(v)
	}

	if v, ok := tfMap["encryption_configuration"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.EncryptionConfiguration = ExpandReplicationRuleDestinationEncryptionConfiguration(v)
	}

	if v, ok := tfMap["metrics"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.Metrics = ExpandReplicationRuleDestinationMetrics(v)
	}

	if v, ok := tfMap["replication_time"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.ReplicationTime = ExpandReplicationRuleDestinationReplicationTime(v)
	}

	if v, ok := tfMap["storage_class"].(string); ok && v != "" {
		result.StorageClass = aws.String(v)
	}

	return result
}

func ExpandReplicationRuleExistingObjectReplication(l []interface{}) *s3.ExistingObjectReplication {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &s3.ExistingObjectReplication{}

	if v, ok := tfMap["status"].(string); ok && v != "" {
		result.Status = aws.String(v)
	}

	return result
}

func ExpandReplicationRuleFilter(l []interface{}) *s3.ReplicationRuleFilter {
	if len(l) == 0 {
		return nil
	}

	result := &s3.ReplicationRuleFilter{}

	// Support the empty filter block in terraform i.e. 'filter {}',
	// which is also supported by the API even though the docs note that
	// one of Prefix, Tag, or And is required.
	if l[0] == nil {
		return result
	}

	tfMap := l[0].(map[string]interface{})

	if v, ok := tfMap["and"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.And = ExpandReplicationRuleFilterAndOperator(v)
	}

	if v, ok := tfMap["tag"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.Tag = ExpandReplicationRuleFilterTag(v)
	}

	// Per AWS S3 API, "A Filter must have exactly one of Prefix, Tag, or And specified";
	// Specifying more than one of the listed parameters results in a MalformedXML error.
	// If a filter is specified as filter { prefix = "" } in Terraform, we should send the prefix value
	// in the API request even if it is an empty value, else Terraform will report non-empty plans.
	// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/23487
	if v, ok := tfMap["prefix"].(string); ok && result.And == nil && result.Tag == nil {
		result.Prefix = aws.String(v)
	}

	return result
}

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
			return nil, fmt.Errorf("error parsing S3 Bucket Lifecycle Rule Expiration date: %w", err)
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
func ExpandLifecycleRuleFilter(l []interface{}) *s3.LifecycleRuleFilter {
	if len(l) == 0 {
		return nil
	}

	result := &s3.LifecycleRuleFilter{}

	if l[0] == nil {
		return result
	}

	m := l[0].(map[string]interface{})

	if v, ok := m["and"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.And = ExpandLifecycleRuleFilterAndOperator(v[0].(map[string]interface{}))
	}

	if v, null, _ := nullable.Int(m["object_size_greater_than"].(string)).Value(); !null && v > 0 {
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

func ExpandLifecycleRuleFilterAndOperator(m map[string]interface{}) *s3.LifecycleRuleAndOperator {
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
		tags := Tags(tftags.New(v).IgnoreAWS())
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
				return nil, fmt.Errorf("error parsing S3 Bucket Lifecycle Rule Transition date: %w", err)
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

func ExpandLifecycleRules(l []interface{}) ([]*s3.LifecycleRule, error) {
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
			result.Filter = ExpandLifecycleRuleFilter(v)
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

func ExpandReplicationRuleDestinationMetrics(l []interface{}) *s3.Metrics {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &s3.Metrics{}

	if v, ok := tfMap["event_threshold"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.EventThreshold = ExpandReplicationRuleDestinationReplicationTimeValue(v)
	}

	if v, ok := tfMap["status"].(string); ok && v != "" {
		result.Status = aws.String(v)
	}

	return result
}

func ExpandReplicationRuleFilterAndOperator(l []interface{}) *s3.ReplicationRuleAndOperator {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &s3.ReplicationRuleAndOperator{}

	if v, ok := tfMap["prefix"].(string); ok && v != "" {
		result.Prefix = aws.String(v)
	}

	if v, ok := tfMap["tags"].(map[string]interface{}); ok && len(v) > 0 {
		tags := Tags(tftags.New(v).IgnoreAWS())
		if len(tags) > 0 {
			result.Tags = tags
		}
	}

	return result
}

func ExpandReplicationRuleDestinationReplicationTime(l []interface{}) *s3.ReplicationTime {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &s3.ReplicationTime{}

	if v, ok := tfMap["status"].(string); ok && v != "" {
		result.Status = aws.String(v)
	}

	if v, ok := tfMap["time"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.Time = ExpandReplicationRuleDestinationReplicationTimeValue(v)
	}

	return result
}

func ExpandReplicationRuleDestinationReplicationTimeValue(l []interface{}) *s3.ReplicationTimeValue {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &s3.ReplicationTimeValue{}

	if v, ok := tfMap["minutes"].(int); ok {
		result.Minutes = aws.Int64(int64(v))
	}

	return result
}

func ExpandSourceSelectionCriteriaReplicaModifications(l []interface{}) *s3.ReplicaModifications {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &s3.ReplicaModifications{}

	if v, ok := tfMap["status"].(string); ok && v != "" {
		result.Status = aws.String(v)
	}

	return result
}

func ExpandReplicationRules(l []interface{}) []*s3.ReplicationRule {
	var rules []*s3.ReplicationRule

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}
		rule := &s3.ReplicationRule{}

		if v, ok := tfMap["delete_marker_replication"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			rule.DeleteMarkerReplication = ExpandReplicationRuleDeleteMarkerReplication(v)
		}

		if v, ok := tfMap["destination"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			rule.Destination = ExpandReplicationRuleDestination(v)
		}

		if v, ok := tfMap["existing_object_replication"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			rule.ExistingObjectReplication = ExpandReplicationRuleExistingObjectReplication(v)
		}

		if v, ok := tfMap["id"].(string); ok && v != "" {
			rule.ID = aws.String(v)
		}

		if v, ok := tfMap["source_selection_criteria"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			rule.SourceSelectionCriteria = ExpandReplicationRuleSourceSelectionCriteria(v)
		}

		if v, ok := tfMap["status"].(string); ok && v != "" {
			rule.Status = aws.String(v)
		}

		// Support the empty filter block in terraform i.e. 'filter {}',
		// which implies the replication rule does not require a specific filter,
		// by expanding the "filter" array even if the first element is nil.
		if v, ok := tfMap["filter"].([]interface{}); ok && len(v) > 0 {
			// XML schema V2
			rule.Filter = ExpandReplicationRuleFilter(v)
			rule.Priority = aws.Int64(int64(tfMap["priority"].(int)))
		} else {
			// XML schema V1
			rule.Prefix = aws.String(tfMap["prefix"].(string))
		}

		rules = append(rules, rule)
	}

	return rules
}

func ExpandReplicationRuleSourceSelectionCriteria(l []interface{}) *s3.SourceSelectionCriteria {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &s3.SourceSelectionCriteria{}

	if v, ok := tfMap["replica_modifications"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.ReplicaModifications = ExpandSourceSelectionCriteriaReplicaModifications(v)
	}

	if v, ok := tfMap["sse_kms_encrypted_objects"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.SseKmsEncryptedObjects = ExpandSourceSelectionCriteriaSSEKMSEncryptedObjects(v)
	}

	return result
}

func ExpandSourceSelectionCriteriaSSEKMSEncryptedObjects(l []interface{}) *s3.SseKmsEncryptedObjects {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &s3.SseKmsEncryptedObjects{}

	if v, ok := tfMap["status"].(string); ok && v != "" {
		result.Status = aws.String(v)
	}

	return result
}

func ExpandReplicationRuleFilterTag(l []interface{}) *s3.Tag {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &s3.Tag{}

	if v, ok := tfMap["key"].(string); ok && v != "" {
		result.Key = aws.String(v)
	}

	if v, ok := tfMap["value"].(string); ok && v != "" {
		result.Value = aws.String(v)
	}

	return result
}

func FlattenReplicationRuleDestinationAccessControlTranslation(act *s3.AccessControlTranslation) []interface{} {
	if act == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if act.Owner != nil {
		m["owner"] = aws.StringValue(act.Owner)
	}

	return []interface{}{m}
}

func FlattenReplicationRuleDestinationEncryptionConfiguration(ec *s3.EncryptionConfiguration) []interface{} {
	if ec == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if ec.ReplicaKmsKeyID != nil {
		m["replica_kms_key_id"] = aws.StringValue(ec.ReplicaKmsKeyID)
	}

	return []interface{}{m}
}

func FlattenReplicationRuleDeleteMarkerReplication(dmr *s3.DeleteMarkerReplication) []interface{} {
	if dmr == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if dmr.Status != nil {
		m["status"] = aws.StringValue(dmr.Status)
	}

	return []interface{}{m}
}

func FlattenReplicationRuleDestination(dest *s3.Destination) []interface{} {
	if dest == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if dest.AccessControlTranslation != nil {
		m["access_control_translation"] = FlattenReplicationRuleDestinationAccessControlTranslation(dest.AccessControlTranslation)
	}

	if dest.Account != nil {
		m["account"] = aws.StringValue(dest.Account)
	}

	if dest.Bucket != nil {
		m["bucket"] = aws.StringValue(dest.Bucket)
	}

	if dest.EncryptionConfiguration != nil {
		m["encryption_configuration"] = FlattenReplicationRuleDestinationEncryptionConfiguration(dest.EncryptionConfiguration)
	}

	if dest.Metrics != nil {
		m["metrics"] = FlattenReplicationRuleDestinationMetrics(dest.Metrics)
	}

	if dest.ReplicationTime != nil {
		m["replication_time"] = FlattenReplicationRuleDestinationReplicationTime(dest.ReplicationTime)
	}

	if dest.StorageClass != nil {
		m["storage_class"] = aws.StringValue(dest.StorageClass)
	}

	return []interface{}{m}
}

func FlattenReplicationRuleExistingObjectReplication(eor *s3.ExistingObjectReplication) []interface{} {
	if eor == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if eor.Status != nil {
		m["status"] = aws.StringValue(eor.Status)
	}

	return []interface{}{m}
}

func FlattenReplicationRuleFilter(filter *s3.ReplicationRuleFilter) []interface{} {
	if filter == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if filter.And != nil {
		m["and"] = FlattenReplicationRuleFilterAndOperator(filter.And)
	}

	if filter.Prefix != nil {
		m["prefix"] = aws.StringValue(filter.Prefix)
	}

	if filter.Tag != nil {
		m["tag"] = FlattenReplicationRuleFilterTag(filter.Tag)
	}

	return []interface{}{m}
}

func FlattenLifecycleRules(rules []*s3.LifecycleRule) []interface{} {
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
			m["filter"] = FlattenLifecycleRuleFilter(rule.Filter)
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

func FlattenLifecycleRuleFilter(filter *s3.LifecycleRuleFilter) []interface{} {
	if filter == nil {
		return nil
	}

	m := make(map[string]interface{})

	if filter.And != nil {
		m["and"] = FlattenLifecycleRuleFilterAndOperator(filter.And)
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

func FlattenLifecycleRuleFilterAndOperator(andOp *s3.LifecycleRuleAndOperator) []interface{} {
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
		m["tags"] = KeyValueTags(andOp.Tags).IgnoreAWS().Map()
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

func FlattenReplicationRuleDestinationMetrics(metrics *s3.Metrics) []interface{} {
	if metrics == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if metrics.EventThreshold != nil {
		m["event_threshold"] = FlattenReplicationRuleDestinationReplicationTimeValue(metrics.EventThreshold)
	}

	if metrics.Status != nil {
		m["status"] = aws.StringValue(metrics.Status)
	}

	return []interface{}{m}
}

func FlattenReplicationRuleDestinationReplicationTime(rt *s3.ReplicationTime) []interface{} {
	if rt == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if rt.Status != nil {
		m["status"] = aws.StringValue(rt.Status)
	}

	if rt.Time != nil {
		m["time"] = FlattenReplicationRuleDestinationReplicationTimeValue(rt.Time)
	}

	return []interface{}{m}

}

func FlattenReplicationRuleDestinationReplicationTimeValue(rtv *s3.ReplicationTimeValue) []interface{} {
	if rtv == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if rtv.Minutes != nil {
		m["minutes"] = int(aws.Int64Value(rtv.Minutes))
	}

	return []interface{}{m}
}

func FlattenReplicationRules(rules []*s3.ReplicationRule) []interface{} {
	if len(rules) == 0 {
		return []interface{}{}
	}

	var results []interface{}

	for _, rule := range rules {
		if rule == nil {
			continue
		}

		m := make(map[string]interface{})

		if rule.DeleteMarkerReplication != nil {
			m["delete_marker_replication"] = FlattenReplicationRuleDeleteMarkerReplication(rule.DeleteMarkerReplication)
		}

		if rule.Destination != nil {
			m["destination"] = FlattenReplicationRuleDestination(rule.Destination)
		}

		if rule.ExistingObjectReplication != nil {
			m["existing_object_replication"] = FlattenReplicationRuleExistingObjectReplication(rule.ExistingObjectReplication)
		}

		if rule.Filter != nil {
			m["filter"] = FlattenReplicationRuleFilter(rule.Filter)
		}

		if rule.ID != nil {
			m["id"] = aws.StringValue(rule.ID)
		}

		if rule.Prefix != nil {
			m["prefix"] = aws.StringValue(rule.Prefix)
		}

		if rule.Priority != nil {
			m["priority"] = int(aws.Int64Value(rule.Priority))
		}

		if rule.SourceSelectionCriteria != nil {
			m["source_selection_criteria"] = FlattenReplicationRuleSourceSelectionCriteria(rule.SourceSelectionCriteria)
		}

		if rule.Status != nil {
			m["status"] = aws.StringValue(rule.Status)
		}

		results = append(results, m)
	}

	return results
}

func FlattenSourceSelectionCriteriaReplicaModifications(rc *s3.ReplicaModifications) []interface{} {
	if rc == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if rc.Status != nil {
		m["status"] = aws.StringValue(rc.Status)
	}

	return []interface{}{m}
}

func FlattenReplicationRuleFilterAndOperator(op *s3.ReplicationRuleAndOperator) []interface{} {
	if op == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if op.Prefix != nil {
		m["prefix"] = aws.StringValue(op.Prefix)
	}

	if op.Tags != nil {
		m["tags"] = KeyValueTags(op.Tags).IgnoreAWS().Map()
	}

	return []interface{}{m}

}

func FlattenReplicationRuleFilterTag(tag *s3.Tag) []interface{} {
	if tag == nil {
		return []interface{}{}
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

func FlattenReplicationRuleSourceSelectionCriteria(ssc *s3.SourceSelectionCriteria) []interface{} {
	if ssc == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if ssc.ReplicaModifications != nil {
		m["replica_modifications"] = FlattenSourceSelectionCriteriaReplicaModifications(ssc.ReplicaModifications)
	}

	if ssc.SseKmsEncryptedObjects != nil {
		m["sse_kms_encrypted_objects"] = FlattenSourceSelectionCriteriaSSEKMSEncryptedObjects(ssc.SseKmsEncryptedObjects)
	}

	return []interface{}{m}
}

func FlattenSourceSelectionCriteriaSSEKMSEncryptedObjects(objects *s3.SseKmsEncryptedObjects) []interface{} {
	if objects == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if objects.Status != nil {
		m["status"] = aws.StringValue(objects.Status)
	}

	return []interface{}{m}
}
