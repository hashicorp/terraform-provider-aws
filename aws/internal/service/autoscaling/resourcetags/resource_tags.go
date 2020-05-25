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
	Value             string
	PropagateAtLaunch bool
}

// ResourceTags is a standard implementation for AWS Auto Scaling resource tags.
type ResourceTags struct {
	ResourceID string
	Tags       map[string]*TagValue
}

func (tags *ResourceTags) AutoscalingTags() []*autoscaling.Tag {
	result := make([]*autoscaling.Tag, 0, len(tags.Tags))

	for k, v := range tags.Tags {
		tag := &autoscaling.Tag{
			Key:               aws.String(k),
			PropagateAtLaunch: aws.Bool(v.PropagateAtLaunch),
			ResourceId:        aws.String(tags.ResourceID),
			ResourceType:      aws.String("auto-scaling-group"),
			Value:             aws.String(v.Value),
		}

		result = append(result, tag)
	}

	return result
}

// IgnoreAws returns non-AWS tag keys.
func (tags *ResourceTags) IgnoreAws() *ResourceTags {
	result := &ResourceTags{
		Tags: make(map[string]*TagValue, len(tags.Tags)),
	}

	for k, v := range tags.Tags {
		if !strings.HasPrefix(k, keyvaluetags.AwsTagKeyPrefix) {
			result.Tags[k] = v
		}
	}

	return result
}

// Keys returns tag keys.
func (tags *ResourceTags) Keys() []string {
	result := make([]string, 0, len(tags.Tags))

	for k := range tags.Tags {
		result = append(result, k)
	}

	return result
}

// New creates ResourceTags from common Terraform Provider SDK types.
func New(resourceID string, i interface{}) (*ResourceTags, error) {
	switch values := i.(type) {
	case []interface{}:
		// The list of tags described by ListSchema().
		tags := &ResourceTags{
			ResourceID: resourceID,
			Tags:       make(map[string]*TagValue, len(values)),
		}

		for _, value := range values {
			m := value.(map[string]interface{})

			key, ok := m["key"].(string)
			if !ok || key == "" {
				return nil, fmt.Errorf("missing Auto Scaling tag key")
			}

			value, ok := m["key"].(string)
			if !ok {
				return nil, fmt.Errorf("invalid Auto Scaling value for tag key (%s)", key)
			}

			var propagateAtLaunch bool
			var err error

			if v, ok := m["propagate_at_launch"].(bool); ok {
				propagateAtLaunch = v
			} else if v, ok := m["propagate_at_launch"].(string); ok {
				if propagateAtLaunch, err = strconv.ParseBool(v); err != nil {
					return nil, fmt.Errorf("invalid Auto Scaling propagate_at_launch value for tag key (%s): %w", key, err)
				}
			}

			tags.Tags[key] = &TagValue{
				Value:             value,
				PropagateAtLaunch: propagateAtLaunch,
			}
		}

		return tags, nil
	case *schema.Set:
		// The set of tags described by SetSchema().
		return New(resourceID, values.List())
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
