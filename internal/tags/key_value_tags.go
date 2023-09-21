// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tags

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/go-cty/cty"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	awsTagKeyPrefix                             = `aws:` // nosemgrep:ci.aws-in-const-name,ci.aws-in-var-name
	ElasticbeanstalkTagKeyPrefix                = `elasticbeanstalk:`
	NameTagKey                                  = `Name`
	ServerlessApplicationRepositoryTagKeyPrefix = `serverlessrepo:`
)

// DefaultConfig contains tags to default across all resources.
type DefaultConfig struct {
	Tags KeyValueTags
}

// IgnoreConfig contains various options for removing resource tags.
type IgnoreConfig struct {
	Keys        KeyValueTags
	KeyPrefixes KeyValueTags
}

// KeyValueTags is a standard implementation for AWS key-value resource tags.
// The AWS Go SDK is split into multiple service packages, each service with
// its own Go struct type representing a resource tag. To standardize logic
// across all these Go types, we convert them into this Go type.
type KeyValueTags map[string]*TagData

// GetTags is convenience method that returns the DefaultConfig's Tags, if any
func (dc *DefaultConfig) GetTags() KeyValueTags {
	if dc == nil {
		return nil
	}

	return dc.Tags
}

// MergeTags returns the result of keyvaluetags.Merge() on the given
// DefaultConfig.Tags with KeyValueTags provided as an argument,
// overriding the value of any tag with a matching key.
func (dc *DefaultConfig) MergeTags(tags KeyValueTags) KeyValueTags {
	if dc == nil || dc.Tags == nil {
		return tags
	}

	return dc.Tags.Merge(tags)
}

// TagsEqual returns true if the given configuration's Tags
// are equal to those passed in as an argument;
// otherwise returns false
func (dc *DefaultConfig) TagsEqual(tags KeyValueTags) bool {
	if dc == nil || dc.Tags == nil {
		return tags == nil
	}

	if tags == nil {
		return false
	}

	if len(tags) == 0 {
		return len(dc.Tags) == 0
	}

	return dc.Tags.ContainsAll(tags)
}

// IgnoreAWS returns non-AWS tag keys.
func (tags KeyValueTags) IgnoreAWS() KeyValueTags { // nosemgrep:ci.aws-in-func-name
	result := make(KeyValueTags)

	for k, v := range tags {
		if !strings.HasPrefix(k, awsTagKeyPrefix) {
			result[k] = v
		}
	}

	return result
}

// IgnoreConfig returns any tags not removed by a given configuration.
func (tags KeyValueTags) IgnoreConfig(config *IgnoreConfig) KeyValueTags {
	if config == nil {
		return tags
	}

	result := tags.IgnorePrefixes(config.KeyPrefixes)
	result = result.Ignore(config.Keys)

	return result
}

// IgnoreElasticbeanstalk returns non-AWS and non-Elasticbeanstalk tag keys.
func (tags KeyValueTags) IgnoreElasticbeanstalk() KeyValueTags {
	result := make(KeyValueTags)

	for k, v := range tags {
		if strings.HasPrefix(k, awsTagKeyPrefix) {
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

// IgnorePrefixes returns non-matching tag key prefixes.
func (tags KeyValueTags) IgnorePrefixes(ignoreTagPrefixes KeyValueTags) KeyValueTags {
	result := make(KeyValueTags)

	for k, v := range tags {
		var ignore bool

		for ignoreTagPrefix := range ignoreTagPrefixes {
			if strings.HasPrefix(k, ignoreTagPrefix) {
				ignore = true
				break
			}
		}

		if ignore {
			continue
		}

		result[k] = v
	}

	return result
}

// IgnoreServerlessApplicationRepository returns non-AWS and non-ServerlessApplicationRepository tag keys.
func (tags KeyValueTags) IgnoreServerlessApplicationRepository() KeyValueTags {
	result := make(KeyValueTags)

	for k, v := range tags {
		if strings.HasPrefix(k, awsTagKeyPrefix) {
			continue
		}

		if strings.HasPrefix(k, ServerlessApplicationRepositoryTagKeyPrefix) {
			continue
		}

		result[k] = v
	}

	return result
}

// IgnoreSystem returns non-system tag keys.
// The ignored keys vary on the specified service.
func (tags KeyValueTags) IgnoreSystem(serviceName string) KeyValueTags {
	switch serviceName {
	case names.ElasticBeanstalk:
		return tags.IgnoreElasticbeanstalk()
	case names.ServerlessRepo:
		return tags.IgnoreServerlessApplicationRepository()
	default:
		return tags.IgnoreAWS()
	}
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

// KeyAdditionalBoolValue returns the boolean value of an additional tag field.
// If the key or additional field is not found, returns nil.
func (tags KeyValueTags) KeyAdditionalBoolValue(key string, fieldName string) *bool {
	tag, ok := tags[key]

	if !ok || tag == nil || tag.AdditionalBoolFields == nil {
		return nil
	}

	if v, ok := tag.AdditionalBoolFields[fieldName]; ok {
		return v
	}

	return nil
}

// KeyAdditionalStringValue returns the string value of an additional tag field.
// If the key or additional field is not found, returns nil.
func (tags KeyValueTags) KeyAdditionalStringValue(key string, fieldName string) *string {
	tag, ok := tags[key]

	if !ok || tag == nil || tag.AdditionalStringFields == nil {
		return nil
	}

	if v, ok := tag.AdditionalStringFields[fieldName]; ok {
		return v
	}

	return nil
}

// KeyExists returns true if a tag key exists.
// If the key is not found, returns nil.
func (tags KeyValueTags) KeyExists(key string) bool {
	if _, ok := tags[key]; ok {
		return true
	}

	return false
}

// KeyTagData returns all tag key data.
// If the key is not found, returns nil.
// Use KeyExists to determine if key is present.
func (tags KeyValueTags) KeyTagData(key string) *TagData {
	if v, ok := tags[key]; ok {
		return v
	}

	return nil
}

// KeyValue returns a tag key value.
// If the key is not found, returns nil.
// Use KeyExists to determine if key is present.
func (tags KeyValueTags) KeyValue(key string) *string {
	v, ok := tags[key]

	if !ok || v == nil {
		return nil
	}

	return v.Value
}

// Keys returns tag keys.
func (tags KeyValueTags) Keys() []string {
	result := make([]string, 0, len(tags))

	for k := range tags {
		result = append(result, k)
	}

	return result
}

// ListofMap returns a list of flattened tags.
// Compatible with setting Terraform state for strongly typed configuration blocks.
func (tags KeyValueTags) ListofMap() []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(tags))

	for k, v := range tags {
		m := map[string]interface{}{
			"key":   k,
			"value": "",
		}

		if v == nil {
			result = append(result, m)
			continue
		}

		if v.Value != nil {
			m["value"] = *v.Value
		}

		for k, v := range v.AdditionalBoolFields {
			m[ToSnakeCase(k)] = false

			if v != nil {
				m[ToSnakeCase(k)] = *v
			}
		}

		for k, v := range v.AdditionalStringFields {
			m[ToSnakeCase(k)] = ""

			if v != nil {
				m[ToSnakeCase(k)] = *v
			}
		}

		result = append(result, m)
	}

	return result
}

// Map returns tag keys mapped to their values.
func (tags KeyValueTags) Map() map[string]string {
	result := make(map[string]string, len(tags))

	for k, v := range tags {
		if v == nil || v.Value == nil {
			result[k] = ""
			continue
		}

		result[k] = *v.Value
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

// Only returns matching tag keys.
func (tags KeyValueTags) Only(onlyTags KeyValueTags) KeyValueTags {
	result := make(KeyValueTags)

	for k, v := range tags {
		if _, ok := onlyTags[k]; !ok {
			continue
		}

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
		if oldV, ok := tags[k]; !ok || !oldV.Equal(newV) {
			result[k] = newV
		}
	}

	return result
}

// Chunks returns a slice of KeyValueTags, each of the specified size.
func (tags KeyValueTags) Chunks(size int) []KeyValueTags {
	result := []KeyValueTags{}

	i := 0
	var chunk KeyValueTags
	for k, v := range tags {
		if i%size == 0 {
			chunk = make(KeyValueTags)
			result = append(result, chunk)
		}

		chunk[k] = v

		i++
	}

	return result
}

// ContainsAll returns whether or not all the target tags are contained.
func (tags KeyValueTags) ContainsAll(target KeyValueTags) bool {
	for key, value := range target {
		if v, ok := tags[key]; !ok || !v.Equal(value) {
			return false
		}
	}

	return true
}

func (tags KeyValueTags) Difference(target KeyValueTags) KeyValueTags {
	result := make(KeyValueTags)

	for k, v := range tags {
		if val, ok := target[k]; !ok || (v.ValueString() != val.ValueString()) {
			result[k] = v
		}
	}

	return result
}

func (tags KeyValueTags) HasZeroValue() bool {
	for _, v := range tags {
		if v.ValueString() == "" {
			return true
		}
	}

	return false
}

// Equal returns whether or two sets of key-value tags are equal.
func (tags KeyValueTags) Equal(other KeyValueTags) bool {
	if tags == nil && other == nil {
		return true
	}

	if tags == nil || other == nil {
		return false
	}

	if len(tags) != len(other) {
		return false
	}

	for k, v := range tags {
		o, ok := other[k]
		if !ok {
			return false
		}

		if !v.Equal(o) {
			return false
		}
	}

	return true
}

func (tags KeyValueTags) DeepEqual(target KeyValueTags) bool {
	return reflect.DeepEqual(tags, target)
}

// Hash returns a stable hash value.
// The returned value may be negative (i.e. not suitable for a 'Set' function).
func (tags KeyValueTags) Hash() int {
	hash := 0

	for k, v := range tags {
		if v == nil || v.Value == nil {
			hash = hash ^ create.StringHashcode(k)
			continue
		}

		hash = hash ^ create.StringHashcode(fmt.Sprintf("%s-%s", k, *v.Value))
	}

	return hash
}

// RemoveDefaultConfig returns tags not present in a DefaultConfig object
// in addition to tags with key/value pairs that override those in a DefaultConfig;
// however, if all tags present in the DefaultConfig object are equivalent to those
// in the given KeyValueTags, then the KeyValueTags are returned, effectively
// bypassing the need to remove differing tags.
func (tags KeyValueTags) RemoveDefaultConfig(dc *DefaultConfig) KeyValueTags {
	if dc == nil || dc.Tags == nil {
		return tags
	}

	result := make(KeyValueTags)

	for k, v := range tags {
		if defaultVal, ok := dc.Tags[k]; !ok || !v.Equal(defaultVal) {
			result[k] = v
		}
	}

	return result
}

// String returns the default string representation of the KeyValueTags.
func (tags KeyValueTags) String() string {
	var builder strings.Builder

	keys := tags.Keys()
	sort.Strings(keys)

	builder.WriteString("map[")
	for i, k := range keys {
		if i > 0 {
			builder.WriteString(" ")
		}
		fmt.Fprintf(&builder, "%s:%s", k, tags[k].String())
	}
	builder.WriteString("]")

	return builder.String()
}

// URLEncode returns the KeyValueTags encoded as URL Query parameters.
func (tags KeyValueTags) URLEncode() string {
	values := url.Values{}

	for k, v := range tags {
		if v == nil || v.Value == nil {
			continue
		}

		values.Add(k, *v.Value)
	}

	return values.Encode()
}

// URLQueryString returns the KeyValueTags formatted as URL Query parameters without encoding.
func (tags KeyValueTags) URLQueryString() string {
	keys := make([]string, 0, len(tags))
	for k, v := range tags {
		if v == nil || v.Value == nil {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf strings.Builder
	for _, k := range keys {
		if buf.Len() > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(k)
		buf.WriteByte('=')
		buf.WriteString(*tags[k].Value)
	}

	return buf.String()
}

// New creates KeyValueTags from common types or returns an empty KeyValueTags.
//
// Supports various Terraform Plugin SDK and Terraform Plugin Framework types including
// map[string]string, map[string]*string, map[string]interface{}, []interface{}, and types.Map.
// When passed []interface{}, all elements are treated as keys and assigned nil values.
// When passed KeyValueTags or its underlying type implementation, returns itself.
func New(ctx context.Context, i interface{}) KeyValueTags {
	switch value := i.(type) {
	case KeyValueTags:
		return make(KeyValueTags).Merge(value)
	case map[string]*TagData:
		return make(KeyValueTags).Merge(KeyValueTags(value))
	case map[string]string:
		kvtm := make(KeyValueTags, len(value))

		for k, v := range value {
			v := v // Prevent referencing issues
			kvtm[k] = &TagData{Value: &v}
		}

		return kvtm
	case map[string]*string:
		kvtm := make(KeyValueTags, len(value))

		for k, v := range value {
			v := v

			if v == nil {
				kvtm[k] = nil
				continue
			}

			kvtm[k] = &TagData{Value: v}
		}

		return kvtm
	case map[string]interface{}:
		kvtm := make(KeyValueTags, len(value))

		for k, v := range value {
			kvtm[k] = &TagData{}

			v, ok := v.(string)

			if ok {
				kvtm[k].Value = &v
			}
		}

		return kvtm
	case []string:
		kvtm := make(KeyValueTags, len(value))

		for _, v := range value {
			kvtm[v] = nil
		}

		return kvtm
	case []interface{}:
		kvtm := make(KeyValueTags, len(value))

		for _, v := range value {
			kvtm[v.(string)] = nil
		}

		return kvtm
	case types.Map:
		return New(ctx, flex.ExpandFrameworkStringValueMap(ctx, value))
	default:
		return make(KeyValueTags)
	}
}

// TagData represents the data associated with a resource tag key.
// Almost exclusively for AWS services, this is just a tag value,
// however there are services that attach additional data to tags.
// An example is autoscaling with the PropagateAtLaunch field.
type TagData struct {
	// Additional boolean field names and values associated with this tag.
	// Each service is responsible for properly handling this data.
	AdditionalBoolFields map[string]*bool

	// Additional string field names and values associated with this tag.
	// Each service is responsible for properly handling this data.
	AdditionalStringFields map[string]*string

	// Tag value.
	Value *string
}

func (td *TagData) ValueString() string {
	if td.Value == nil {
		return ""
	}

	return *td.Value
}

func (td *TagData) Equal(other *TagData) bool {
	if td == nil && other == nil {
		return true
	}

	if td == nil || other == nil {
		return false
	}

	if !reflect.DeepEqual(td.AdditionalBoolFields, other.AdditionalBoolFields) {
		return false
	}

	if !reflect.DeepEqual(td.AdditionalStringFields, other.AdditionalStringFields) {
		return false
	}

	if !reflect.DeepEqual(td.Value, other.Value) {
		return false
	}

	return true
}

func (td *TagData) String() string {
	if td == nil {
		return ""
	}

	var fields []string

	if len(td.AdditionalBoolFields) > 0 {
		var additionalBoolFields []string

		for k, v := range td.AdditionalBoolFields {
			additionalBoolField := fmt.Sprintf("%s:", k)

			if v != nil {
				additionalBoolField += fmt.Sprintf("%t", *v)
			}

			additionalBoolFields = append(additionalBoolFields, additionalBoolField)
		}

		fields = append(fields, fmt.Sprintf("AdditionalBoolFields: map[%s]", strings.Join(additionalBoolFields, " ")))
	}

	if len(td.AdditionalStringFields) > 0 {
		var additionalStringFields []string

		for k, v := range td.AdditionalStringFields {
			additionalStringField := fmt.Sprintf("%s:", k)

			if v != nil {
				additionalStringField += *v
			}

			additionalStringFields = append(additionalStringFields, additionalStringField)
		}

		fields = append(fields, fmt.Sprintf("AdditionalStringFields: map[%s]", strings.Join(additionalStringFields, " ")))
	}

	if td.Value != nil {
		fields = append(fields, fmt.Sprintf("Value: %s", *td.Value))
	}

	return fmt.Sprintf("TagData{%s}", strings.Join(fields, ", "))
}

// schemaResourceData is an interface that implements functions from schema.ResourceData
type schemaResourceData interface {
	GetRawConfig() cty.Value
	GetRawPlan() cty.Value
	GetRawState() cty.Value
}

// tagSource is an enum that identifiers the source of the tag
type tagSource int

const (
	configuration tagSource = iota
	plan
	state
)

// configTag contains the value and source of the incoming tag
type configTag struct {
	value  string
	source tagSource
}

// ResolveDuplicates resolves differences between incoming tags, defaultTags, and ignoreConfig
func (tags KeyValueTags) ResolveDuplicates(ctx context.Context, defaultConfig *DefaultConfig, ignoreConfig *IgnoreConfig, d schemaResourceData) KeyValueTags {
	// remove default config.
	t := tags.RemoveDefaultConfig(defaultConfig)

	cf := d.GetRawConfig()
	configExists := !cf.IsNull() && cf.IsKnown()

	result := make(map[string]string)
	for k, v := range t {
		result[k] = v.ValueString()
	}

	configTags := make(map[string]configTag)
	if configExists {
		c := cf.GetAttr(names.AttrTags)

		// if the config is null just return the incoming tags
		// no duplicates to calculate
		if c.IsNull() {
			return t
		}

		if !c.IsNull() && c.IsKnown() {
			normalizeTagsFromRaw(c.AsValueMap(), configTags, configuration)
		}
	}

	if pl := d.GetRawPlan(); !pl.IsNull() && pl.IsKnown() {
		c := pl.GetAttr(names.AttrTags)
		if !c.IsNull() && c.IsKnown() {
			normalizeTagsFromRaw(c.AsValueMap(), configTags, plan)
		}
	}

	if st := d.GetRawState(); !st.IsNull() && st.IsKnown() {
		c := st.GetAttr(names.AttrTags)
		if !c.IsNull() {
			normalizeTagsFromRaw(c.AsValueMap(), configTags, state)
		}
	}

	for k, v := range configTags {
		if _, ok := result[k]; !ok {
			if defaultConfig != nil {
				if val, ok := defaultConfig.Tags[k]; ok && val.ValueString() == v.value {
					// config does not exist during a refresh.
					// set duplicate values from other sources for refresh diff calculation
					if !configExists {
						result[k] = v.value
					} else {
						if v.source == configuration {
							result[k] = v.value
						}
					}
				}
			}
		}
	}

	return New(ctx, result).IgnoreConfig(ignoreConfig)
}

// ResolveDuplicatesFramework resolves differences between incoming tags, defaultTags, and ignoreConfig
func (tags KeyValueTags) ResolveDuplicatesFramework(ctx context.Context, defaultConfig *DefaultConfig, ignoreConfig *IgnoreConfig, resp *resource.ReadResponse, diags fwdiag.Diagnostics) KeyValueTags {
	// remove default config.
	t := tags.RemoveDefaultConfig(defaultConfig)

	var tagsAll types.Map
	diags.Append(resp.State.GetAttribute(ctx, path.Root("tags"), &tagsAll)...)

	if diags.HasError() {
		return KeyValueTags{}
	}

	result := make(map[string]string)
	for k, v := range t {
		result[k] = v.ValueString()
	}

	for k, v := range tagsAll.Elements() {
		if _, ok := result[k]; !ok {
			if defaultConfig != nil {
				s, err := strconv.Unquote(v.String()) // TODO rework to use Framework Map.Equals() value

				if err != nil {
					diags.AddError(
						"unable to normalize string",
						"unable to normalize string default value",
					)
				}

				if val, ok := defaultConfig.Tags[k]; ok && val.ValueString() == s {
					result[k] = s
				}
			}
		}
	}

	return New(ctx, result).IgnoreConfig(ignoreConfig)
}

// ToSnakeCase converts a string to snake case.
//
// For example, AWS Go SDK field names are in PascalCase,
// while Terraform schema attribute names are in snake_case.
func ToSnakeCase(str string) string {
	result := regexache.MustCompile("(.)([A-Z][a-z]+)").ReplaceAllString(str, "${1}_${2}")
	result = regexache.MustCompile("([0-9a-z])([A-Z])").ReplaceAllString(result, "${1}_${2}")
	return strings.ToLower(result)
}

func normalizeTagsFromRaw(m map[string]cty.Value, incoming map[string]configTag, source tagSource) {
	for k, v := range m {
		if !v.IsNull() {
			if _, ok := incoming[k]; !ok {
				incoming[k] = configTag{
					value:  v.AsString(),
					source: source,
				}
			}
		}
	}
}
