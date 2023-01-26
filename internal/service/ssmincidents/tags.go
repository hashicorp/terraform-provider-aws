package ssmincidents

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssmincidents"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// lists all tags for a particular resource
func listResourceTags(context context.Context, client *ssmincidents.Client, arn string) (tftags.KeyValueTags, error) {
	input := &ssmincidents.ListTagsForResourceInput{
		ResourceArn: aws.String(arn),
	}

	output, err := client.ListTagsForResource(context, input)

	if err != nil {
		return tftags.New(nil), err
	}

	return tftags.New(output.Tags), nil
}

// gets all tags via get request and sets them in Resource Data ssmincidents resource
func setResourceDataTags(context context.Context, d *schema.ResourceData, meta interface{}, client *ssmincidents.Client, resourceName string) diag.Diagnostics {
	tags, err := listResourceTags(context, client, d.Id())
	if err != nil {
		return create.DiagError(names.SSMIncidents, create.ErrActionReading, resourceName, d.Id(), err)
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return create.DiagError(names.SSMIncidents, create.ErrActionSetting, resourceName, d.Id(), err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return create.DiagError(names.SSMIncidents, create.ErrActionSetting, resourceName, d.Id(), err)
	}

	return nil
}

// makes api calls to update Resource Data Tags
func updateResourceTags(context context.Context, client *ssmincidents.Client, d *schema.ResourceData) error {
	old, new := d.GetChange("tags_all")

	oldTags := tftags.New(old)
	newTags := tftags.New(new)

	allNewTagsMap := flex.ExpandStringValueMap(new.(map[string]interface{}))

	if err := updateResourceTag(context, client, d.Id(), oldTags.Removed(newTags), oldTags.Updated(newTags)); err != nil {
		return err
	}

	// provider level tags cannot have "" as value
	// resource level tags can have "" as value but this change is not recorded by d.GetChange("tags_all")
	// so we have to look specifically for any tags updated with "" as the value

	old, new = d.GetChange("tags")

	oldTags = tftags.New(old)
	newTags = tftags.New(new)

	toUpdate := make(map[string]string)

	for k, v := range oldTags.Updated(newTags).Map() {
		if v == "" {
			toUpdate[k] = v
			allNewTagsMap[k] = v
		}
	}

	// since we are adding an extra tag to tags_all not initially detected by terraform
	// we must set tags_all to what is properly expected in create/update function so that
	// terraform plan is consistent to what we receive with terraform refresh/the update function
	d.Set("tags_all", allNewTagsMap)

	empty := tftags.KeyValueTags{}
	if err := updateResourceTag(context, client, d.Id(), empty, tftags.New(toUpdate)); err != nil {
		return err
	}

	return nil
}

func updateResourceTag(context context.Context, client *ssmincidents.Client, arn string, removedTags, addedTags tftags.KeyValueTags) error {
	if len(removedTags) > 0 {
		input := &ssmincidents.UntagResourceInput{
			ResourceArn: aws.String(arn),
			TagKeys:     removedTags.Keys(),
		}
		if _, err := client.UntagResource(context, input); err != nil {
			return err
		}
	}

	if len(addedTags) > 0 {
		input := &ssmincidents.TagResourceInput{
			ResourceArn: aws.String(arn),
			Tags:        addedTags.IgnoreAWS().Map(),
		}

		if _, err := client.TagResource(context, input); err != nil {
			return err
		}
	}

	return nil
}
