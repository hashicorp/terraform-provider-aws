package s3

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func ExpandAccessControlTranslation(l []interface{}) *s3.AccessControlTranslation {
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

func ExpandEncryptionConfiguration(l []interface{}) *s3.EncryptionConfiguration {
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

func ExpandDeleteMarkerReplication(l []interface{}) *s3.DeleteMarkerReplication {
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

func ExpandDestination(l []interface{}) *s3.Destination {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &s3.Destination{}

	if v, ok := tfMap["access_control_translation"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.AccessControlTranslation = ExpandAccessControlTranslation(v)
	}

	if v, ok := tfMap["account"].(string); ok && v != "" {
		result.Account = aws.String(v)
	}

	if v, ok := tfMap["bucket"].(string); ok && v != "" {
		result.Bucket = aws.String(v)
	}

	if v, ok := tfMap["encryption_configuration"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.EncryptionConfiguration = ExpandEncryptionConfiguration(v)
	}

	if v, ok := tfMap["metrics"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.Metrics = ExpandMetrics(v)
	}

	if v, ok := tfMap["replication_time"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.ReplicationTime = ExpandReplicationTime(v)
	}

	if v, ok := tfMap["storage_class"].(string); ok && v != "" {
		result.StorageClass = aws.String(v)
	}

	return result
}

func ExpandExistingObjectReplication(l []interface{}) *s3.ExistingObjectReplication {
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

func ExpandFilter(l []interface{}) *s3.ReplicationRuleFilter {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &s3.ReplicationRuleFilter{}

	if v, ok := tfMap["and"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.And = ExpandReplicationRuleAndOperator(v)
	}

	if v, ok := tfMap["prefix"].(string); ok && v != "" {
		result.Prefix = aws.String(v)
	}

	if v, ok := tfMap["tag"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		tags := Tags(tftags.New(v[0]).IgnoreAWS())
		if len(tags) > 0 {
			result.Tag = tags[0]
		}
	}

	return result
}

func ExpandMetrics(l []interface{}) *s3.Metrics {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &s3.Metrics{}

	if v, ok := tfMap["event_threshold"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.EventThreshold = ExpandReplicationTimeValue(v)
	}

	if v, ok := tfMap["status"].(string); ok && v != "" {
		result.Status = aws.String(v)
	}

	return result
}

func ExpandReplicationRuleAndOperator(l []interface{}) *s3.ReplicationRuleAndOperator {
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

func ExpandReplicationTime(l []interface{}) *s3.ReplicationTime {
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
		result.Time = ExpandReplicationTimeValue(v)
	}

	return result
}

func ExpandReplicationTimeValue(l []interface{}) *s3.ReplicationTimeValue {
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

func ExpandReplicaModifications(l []interface{}) *s3.ReplicaModifications {
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

func ExpandRules(l []interface{}) []*s3.ReplicationRule {
	var rules []*s3.ReplicationRule

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}
		rule := &s3.ReplicationRule{}

		if v, ok := tfMap["delete_marker_replication"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			rule.DeleteMarkerReplication = ExpandDeleteMarkerReplication(v)
		}

		if v, ok := tfMap["destination"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			rule.Destination = ExpandDestination(v)
		}

		if v, ok := tfMap["existing_object_replication"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			rule.ExistingObjectReplication = ExpandExistingObjectReplication(v)
		}

		if v, ok := tfMap["id"].(string); ok && v != "" {
			rule.ID = aws.String(v)
		}

		if v, ok := tfMap["source_selection_criteria"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			rule.SourceSelectionCriteria = ExpandSourceSelectionCriteria(v)
		}

		if v, ok := tfMap["status"].(string); ok && v != "" {
			rule.Status = aws.String(v)
		}

		if v, ok := tfMap["filter"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			// XML schema V2
			rule.Filter = ExpandFilter(v)
			rule.Priority = aws.Int64(int64(tfMap["priority"].(int)))
		} else {
			// XML schema V1
			rule.Prefix = aws.String(tfMap["prefix"].(string))
		}

		rules = append(rules, rule)
	}

	return rules
}

func ExpandSourceSelectionCriteria(l []interface{}) *s3.SourceSelectionCriteria {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &s3.SourceSelectionCriteria{}

	if v, ok := tfMap["replica_modifications"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.ReplicaModifications = ExpandReplicaModifications(v)
	}

	if v, ok := tfMap["sse_kms_encrypted_objects"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.SseKmsEncryptedObjects = ExpandSseKmsEncryptedObjects(v)
	}

	return result
}

func ExpandSseKmsEncryptedObjects(l []interface{}) *s3.SseKmsEncryptedObjects {
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

func ExpandTag(l []interface{}) *s3.Tag {
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

func FlattenAccessControlTranslation(act *s3.AccessControlTranslation) []interface{} {
	if act == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if act.Owner != nil {
		m["owner"] = aws.StringValue(act.Owner)
	}

	return []interface{}{m}
}

func FlattenEncryptionConfiguration(ec *s3.EncryptionConfiguration) []interface{} {
	if ec == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if ec.ReplicaKmsKeyID != nil {
		m["replica_kms_key_id"] = aws.StringValue(ec.ReplicaKmsKeyID)
	}

	return []interface{}{m}
}

func FlattenDeleteMarkerReplication(dmr *s3.DeleteMarkerReplication) []interface{} {
	if dmr == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if dmr.Status != nil {
		m["status"] = aws.StringValue(dmr.Status)
	}

	return []interface{}{m}
}

func FlattenDestination(dest *s3.Destination) []interface{} {
	if dest == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if dest.AccessControlTranslation != nil {
		m["access_control_translation"] = FlattenAccessControlTranslation(dest.AccessControlTranslation)
	}

	if dest.Account != nil {
		m["account"] = aws.StringValue(dest.Account)
	}

	if dest.Bucket != nil {
		m["bucket"] = aws.StringValue(dest.Bucket)
	}

	if dest.EncryptionConfiguration != nil {
		m["encryption_configuration"] = FlattenEncryptionConfiguration(dest.EncryptionConfiguration)
	}

	if dest.Metrics != nil {
		m["metrics"] = FlattenMetrics(dest.Metrics)
	}

	if dest.ReplicationTime != nil {
		m["replication_time"] = FlattenReplicationTime(dest.ReplicationTime)
	}

	if dest.StorageClass != nil {
		m["storage_class"] = aws.StringValue(dest.StorageClass)
	}

	return []interface{}{m}
}

func FlattenExistingObjectReplication(eor *s3.ExistingObjectReplication) []interface{} {
	if eor == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if eor.Status != nil {
		m["status"] = aws.StringValue(eor.Status)
	}

	return []interface{}{m}
}

func FlattenFilter(filter *s3.ReplicationRuleFilter) []interface{} {
	if filter == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if filter.And != nil {
		m["and"] = FlattenReplicationRuleAndOperator(filter.And)
	}

	if filter.Prefix != nil {
		m["prefix"] = aws.StringValue(filter.Prefix)
	}

	if filter.Tag != nil {
		tag := KeyValueTags([]*s3.Tag{filter.Tag}).IgnoreAWS().Map()
		m["tag"] = []interface{}{tag}
	}

	return []interface{}{m}
}

func FlattenMetrics(metrics *s3.Metrics) []interface{} {
	if metrics == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if metrics.EventThreshold != nil {
		m["event_threshold"] = FlattenReplicationTimeValue(metrics.EventThreshold)
	}

	if metrics.Status != nil {
		m["status"] = aws.StringValue(metrics.Status)
	}

	return []interface{}{m}
}

func FlattenReplicationTime(rt *s3.ReplicationTime) []interface{} {
	if rt == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if rt.Status != nil {
		m["status"] = aws.StringValue(rt.Status)
	}

	if rt.Time != nil {
		m["time"] = FlattenReplicationTimeValue(rt.Time)
	}

	return []interface{}{m}

}

func FlattenReplicationTimeValue(rtv *s3.ReplicationTimeValue) []interface{} {
	if rtv == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if rtv.Minutes != nil {
		m["minutes"] = int(aws.Int64Value(rtv.Minutes))
	}

	return []interface{}{m}
}

func FlattenRules(rules []*s3.ReplicationRule) []interface{} {
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
			m["delete_marker_replication"] = FlattenDeleteMarkerReplication(rule.DeleteMarkerReplication)
		}

		if rule.Destination != nil {
			m["destination"] = FlattenDestination(rule.Destination)
		}

		if rule.ExistingObjectReplication != nil {
			m["existing_object_replication"] = FlattenExistingObjectReplication(rule.ExistingObjectReplication)
		}

		if rule.Filter != nil {
			m["filter"] = FlattenFilter(rule.Filter)
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
			m["source_selection_criteria"] = FlattenSourceSelectionCriteria(rule.SourceSelectionCriteria)
		}

		if rule.Status != nil {
			m["status"] = aws.StringValue(rule.Status)
		}

		results = append(results, m)
	}

	return results
}

func FlattenReplicaModifications(rc *s3.ReplicaModifications) []interface{} {
	if rc == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if rc.Status != nil {
		m["status"] = aws.StringValue(rc.Status)
	}

	return []interface{}{m}
}

func FlattenReplicationRuleAndOperator(op *s3.ReplicationRuleAndOperator) []interface{} {
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

func FlattenSourceSelectionCriteria(ssc *s3.SourceSelectionCriteria) []interface{} {
	if ssc == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if ssc.ReplicaModifications != nil {
		m["replica_modifications"] = FlattenReplicaModifications(ssc.ReplicaModifications)
	}

	if ssc.SseKmsEncryptedObjects != nil {
		m["sse_kms_encrypted_objects"] = FlattenSseKmsEncryptedObjects(ssc.SseKmsEncryptedObjects)
	}

	return []interface{}{m}
}

func FlattenSseKmsEncryptedObjects(objects *s3.SseKmsEncryptedObjects) []interface{} {
	if objects == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if objects.Status != nil {
		m["status"] = aws.StringValue(objects.Status)
	}

	return []interface{}{m}
}
