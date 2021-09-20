package verify


import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tftags "github.com/hashicorp/terraform-provider-aws/aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

// SetTagsDiff sets the new plan difference with the result of
// merging resource tags on to those defined at the provider-level;
// returns an error if unsuccessful or if the resource tags are identical
// to those configured at the provider-level to avoid non-empty plans
// after resource READ operations as resource and provider-level tags
// will be indistinguishable when returned from an AWS API.
func SetTagsDiff(_ context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	resourceTags := tftags.New(diff.Get("tags").(map[string]interface{}))

	if defaultTagsConfig.TagsEqual(resourceTags) {
		return fmt.Errorf(`"tags" are identical to those in the "default_tags" configuration block of the provider: please de-duplicate and try again`)
	}

	allTags := defaultTagsConfig.MergeTags(resourceTags).IgnoreConfig(ignoreTagsConfig)

	// To ensure "tags_all" is correctly computed, we explicitly set the attribute diff
	// when the merger of resource-level tags onto provider-level tags results in n > 0 tags,
	// otherwise we mark the attribute as "Computed" only when their is a known diff (excluding an empty map)
	// or a change for "tags_all".
	// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/18366
	// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/19005
	if len(allTags) > 0 {
		if err := diff.SetNew("tags_all", allTags.Map()); err != nil {
			return fmt.Errorf("error setting new tags_all diff: %w", err)
		}
	} else if len(diff.Get("tags_all").(map[string]interface{})) > 0 {
		if err := diff.SetNewComputed("tags_all"); err != nil {
			return fmt.Errorf("error setting tags_all to computed: %w", err)
		}
	} else if diff.HasChange("tags_all") {
		if err := diff.SetNewComputed("tags_all"); err != nil {
			return fmt.Errorf("error setting tags_all to computed: %w", err)
		}
	}

	return nil
}

