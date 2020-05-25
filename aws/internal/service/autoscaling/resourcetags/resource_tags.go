package resourcetags

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

type TagValue struct {
	PropagateAtLaunch bool
	Value             string
}

// ResourceTags is a standard implementation for AWS Auto Scaling resource tags.
type ResourceTags map[string]TagValue

// AutoscalingTags returns autoscaling service tags.
func (tags ResourceTags) AutoscalingTags(resourceID string) []*autoscaling.Tag {
	result := make([]*autoscaling.Tag, 0, len(tags))

	for k, v := range tags {
		tag := &autoscaling.Tag{
			Key:               aws.String(k),
			PropagateAtLaunch: aws.Bool(v.PropagateAtLaunch),
			ResourceId:        aws.String(resourceID),
			ResourceType:      aws.String("auto-scaling-group"),
			Value:             aws.String(v.Value),
		}

		result = append(result, tag)
	}

	return result
}

// AutoscalingResourceTags creates ResourceTags from autoscaling service tags.
func AutoscalingResourceTags(tags []*autoscaling.TagDescription) (ResourceTags, error) {
	m := make(map[string]TagValue, len(tags))

	for _, tag := range tags {
		m[aws.StringValue(tag.Key)] = TagValue{
			PropagateAtLaunch: aws.BoolValue(tag.PropagateAtLaunch),
			Value:             aws.StringValue(tag.Value),
		}
	}

	return New(m)
}

// AutoscalingUpdateTags updates autoscaling service tags.
func AutoscalingUpdateTags(conn *autoscaling.AutoScaling, identifier string, oldTagsMap interface{}, newTagsMap interface{}) error {
	oldTags, err := New(oldTagsMap)
	if err != nil {
		return err
	}
	newTags, err := New(newTagsMap)
	if err != nil {
		return err
	}

	if removedTags := oldTags.Removed(newTags); len(removedTags) > 0 {
		input := &autoscaling.DeleteTagsInput{
			Tags: removedTags.IgnoreAws().AutoscalingTags(identifier),
		}

		_, err := conn.DeleteTags(input)

		if err != nil {
			return fmt.Errorf("error untagging resource (%s): %w", identifier, err)
		}
	}

	if updatedTags := oldTags.Updated(newTags); len(updatedTags) > 0 {
		input := &autoscaling.CreateOrUpdateTagsInput{
			Tags: updatedTags.IgnoreAws().AutoscalingTags(identifier),
		}

		_, err := conn.CreateOrUpdateTags(input)

		if err != nil {
			return fmt.Errorf("error tagging resource (%s): %w", identifier, err)
		}
	}

	return nil
}

// IgnoreAws returns non-AWS tag keys.
func (tags ResourceTags) IgnoreAws() ResourceTags {
	result := make(ResourceTags, len(tags))

	for k, v := range tags {
		if !strings.HasPrefix(k, keyvaluetags.AwsTagKeyPrefix) {
			result[k] = v
		}
	}

	return result
}

// Keys returns tag keys.
func (tags ResourceTags) Keys() []string {
	result := make([]string, 0, len(tags))

	for k := range tags {
		result = append(result, k)
	}

	return result
}

// List returns tags as described by ListSchema().
func (tags ResourceTags) List() []interface{} {
	result := make([]interface{}, 0, len(tags))

	for k, v := range tags {
		result = append(result, map[string]interface{}{
			"key":   k,
			"value": v.Value,
			// For the list representation, all map values are strings.
			"propagate_at_launch": strconv.FormatBool(v.PropagateAtLaunch),
		})
	}

	return result
}

// Set returns tags as described by SetSchema().
func (tags ResourceTags) Set() []interface{} {
	result := make([]interface{}, 0, len(tags))

	for k, v := range tags {
		result = append(result, map[string]interface{}{
			"key":                 k,
			"value":               v.Value,
			"propagate_at_launch": v.PropagateAtLaunch,
		})
	}

	return result
}

// Removed returns tags removed.
func (tags ResourceTags) Removed(newTags ResourceTags) ResourceTags {
	result := make(ResourceTags)

	for k, v := range tags {
		if _, ok := newTags[k]; !ok {
			result[k] = v
		}
	}

	return result
}

// Updated returns tags added and updated.
func (tags ResourceTags) Updated(newTags ResourceTags) ResourceTags {
	result := make(ResourceTags)

	for k, newV := range newTags {
		if oldV, ok := tags[k]; !ok || oldV.PropagateAtLaunch != newV.PropagateAtLaunch || oldV.Value != newV.Value {
			result[k] = newV
		}
	}

	return result
}

// New creates ResourceTags from common Terraform Provider SDK types.
func New(i interface{}) (ResourceTags, error) {
	switch values := i.(type) {
	case map[string]TagValue:
		return ResourceTags(values), nil
	case []interface{}:
		// The list of tags described by ListSchema().
		tags := make(ResourceTags, len(values))

		for _, value := range values {
			m := value.(map[string]interface{})

			key, ok := m["key"].(string)
			if !ok || key == "" {
				return nil, fmt.Errorf("missing Auto Scaling tag key")
			}

			value, ok := m["value"].(string)
			if !ok {
				return nil, fmt.Errorf("invalid tag value for Auto Scaling tag key (%s)", key)
			}

			v, ok := m["propagate_at_launch"]
			if !ok {
				return nil, fmt.Errorf("missing propagate_at_launch value for Auto Scaling tag key (%s)", key)
			}

			var propagateAtLaunch bool
			var err error

			switch v := v.(type) {
			case bool:
				propagateAtLaunch = v
			case string:
				if propagateAtLaunch, err = strconv.ParseBool(v); err != nil {
					return nil, fmt.Errorf("invalid propagate_at_launch value for Auto Scaling tag key (%s): %w", key, err)
				}
			default:
				return nil, fmt.Errorf("invalid propagate_at_launch type (%T) for Auto Scaling tag key (%s)", v, key)
			}

			tags[key] = TagValue{
				PropagateAtLaunch: propagateAtLaunch,
				Value:             value,
			}
		}

		return tags, nil
	case *schema.Set:
		// The set of tags described by SetSchema().
		return New(values.List())
	default:
		return nil, fmt.Errorf("invalid Auto Scaling tags type: %T", values)
	}
}

// ListSchema returns a *schema.Schema that represents a list of Auto Scaling resource tags.
// It is conventional for an attribute of this type to be included as a top-level attribute called "tags".
func ListSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Schema{
			Type: schema.TypeMap,
			Elem: &schema.Schema{Type: schema.TypeString},
		},
	}
}

// SetSchema returns a *schema.Schema that represents a set of Auto Scaling resource tags.
// It is conventional for an attribute of this type to be included as a top-level attribute called "tag".
func SetSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"key": {
					Type:     schema.TypeString,
					Required: true,
				},

				"value": {
					Type:     schema.TypeString,
					Required: true,
				},

				"propagate_at_launch": {
					Type:     schema.TypeBool,
					Required: true,
				},
			},
		},
	}
}
