//go:generate go run generators/listtags/main.go
//go:generate go run generators/servicetags/main.go
//go:generate go run generators/updatetags/main.go

package keyvaluetags

import (
	"strings"
)

const (
	AwsTagKeyPrefix              = `aws:`
	ElasticbeanstalkTagKeyPrefix = `elasticbeanstalk:`
	NameTagKey                   = `Name`
	RdsTagKeyPrefix              = `rds:`
)

// KeyValueTags is a standard implementation for AWS key-value resource tags.
// The AWS Go SDK is split into multiple service packages, each service with
// its own Go struct type representing a resource tag. To standardize logic
// across all these Go types, we convert them into this Go type.
type KeyValueTags map[string]*string

// IgnoreAws returns non-AWS tag keys.
func (tags KeyValueTags) IgnoreAws() KeyValueTags {
	result := make(KeyValueTags)

	for k, v := range tags {
		if !strings.HasPrefix(k, AwsTagKeyPrefix) {
			result[k] = v
		}
	}

	return result
}

// IgnoreElasticbeanstalk returns non-AWS and non-Elasticbeanstalk tag keys.
func (tags KeyValueTags) IgnoreElasticbeanstalk() KeyValueTags {
	result := make(KeyValueTags)

	for k, v := range tags {
		if strings.HasPrefix(k, AwsTagKeyPrefix) {
			continue
		}

		if strings.HasPrefix(k, ElasticbeanstalkTagKeyPrefix) {
			continue
		}

		if k == NameTagKey {
			continue
		}

		result[k] = v
	}

	return result
}

// IgnoreRDS returns non-AWS and non-RDS tag keys.
func (tags KeyValueTags) IgnoreRds() KeyValueTags {
	result := make(KeyValueTags)

	for k, v := range tags {
		if strings.HasPrefix(k, AwsTagKeyPrefix) {
			continue
		}

		if strings.HasPrefix(k, RdsTagKeyPrefix) {
			continue
		}

		result[k] = v
	}

	return result
}

// Ignore returns non-matching tag keys.
func (tags KeyValueTags) Ignore(ignoreTags KeyValueTags) KeyValueTags {
	result := make(KeyValueTags)

	for k, v := range tags {
		if _, ok := ignoreTags[k]; ok {
			continue
		}

		result[k] = v
	}

	return result
}

// Keys returns tag keys.
func (tags KeyValueTags) Keys() []string {
	result := make([]string, 0, len(tags))

	for k := range tags {
		result = append(result, k)
	}

	return result
}

// Map returns tag keys mapped to their values.
func (tags KeyValueTags) Map() map[string]string {
	result := make(map[string]string, len(tags))

	for k, v := range tags {
		result[k] = *v
	}

	return result
}

// Merge adds missing and updates existing tags.
func (tags KeyValueTags) Merge(mergeTags KeyValueTags) KeyValueTags {
	result := make(KeyValueTags)

	for k, v := range tags {
		result[k] = v
	}

	for k, v := range mergeTags {
		result[k] = v
	}

	return result
}

// Removed returns tags removed.
func (tags KeyValueTags) Removed(newTags KeyValueTags) KeyValueTags {
	result := make(KeyValueTags)

	for k, v := range tags {
		if _, ok := newTags[k]; !ok {
			result[k] = v
		}
	}

	return result
}

// Updated returns tags added and updated.
func (tags KeyValueTags) Updated(newTags KeyValueTags) KeyValueTags {
	result := make(KeyValueTags)

	for k, newV := range newTags {
		if oldV, ok := tags[k]; !ok || *oldV != *newV {
			result[k] = newV
		}
	}

	return result
}

// New creates KeyValueTags from common Terraform Provider SDK types.
// Supports map[string]string, map[string]*string, map[string]interface{}, and []interface{}.
// When passed []interface{}, all elements are treated as keys and assigned nil values.
func New(i interface{}) KeyValueTags {
	switch value := i.(type) {
	case map[string]string:
		kvtm := make(KeyValueTags, len(value))

		for k, v := range value {
			str := v // Prevent referencing issues
			kvtm[k] = &str
		}

		return kvtm
	case map[string]*string:
		return KeyValueTags(value)
	case map[string]interface{}:
		kvtm := make(KeyValueTags, len(value))

		for k, v := range value {
			str := v.(string)
			kvtm[k] = &str
		}

		return kvtm
	case []interface{}:
		kvtm := make(KeyValueTags, len(value))

		for _, v := range value {
			kvtm[v.(string)] = nil
		}

		return kvtm
	default:
		return make(KeyValueTags)
	}
}
