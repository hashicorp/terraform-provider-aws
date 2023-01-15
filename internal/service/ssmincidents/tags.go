package ssmincidents

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssmincidents"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// lists all tags for a particular resource
func listResourceTags(ctx context.Context, conn *ssmincidents.Client, arn string) (tftags.KeyValueTags, error) {
	input := &ssmincidents.ListTagsForResourceInput{
		ResourceArn: aws.String(arn),
	}

	output, err := conn.ListTagsForResource(ctx, input)

	if err != nil {
		return tftags.New(nil), err
	}

	return tftags.New(output.Tags), nil
}

// gets all tags via get request and sets them in terraform state for a ssmincidents resource
func GetSetResourceTags(ctx context.Context, d *schema.ResourceData, meta interface{}, conn *ssmincidents.Client, resourceName string) diag.Diagnostics {
	tags, err := listResourceTags(ctx, conn, d.Id())
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

func UpdateResourceTags(ctx context.Context, conn *ssmincidents.Client, d *schema.ResourceData) error {
	o, n := d.GetChange("tags_all")

	oldTags := tftags.New(o)
	newTags := tftags.New(n)

	allNewTagsMap := CastInterfaceMapToStringMap(n.(map[string]interface{}))

	if err := updateResourceTag(ctx, conn, d.Id(), oldTags.Removed(newTags), oldTags.Updated(newTags)); err != nil {
		return err
	}

	// provider level tags cannot have "" as value
	// resource level tags can have "" as value but this change is not recorded by d.GetChange("tags_all")
	// so we have to look specifically for any tags updated with "" as the value

	o, n = d.GetChange("tags")

	oldTags = tftags.New(o)
	newTags = tftags.New(n)

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
	if err := updateResourceTag(ctx, conn, d.Id(), empty, tftags.New(toUpdate)); err != nil {
		return err
	}

	return nil
}

func updateResourceTag(ctx context.Context, conn *ssmincidents.Client, arn string, removedTags, updatedTags tftags.KeyValueTags) error {
	if len(removedTags) > 0 {
		input := &ssmincidents.UntagResourceInput{
			ResourceArn: aws.String(arn),
			TagKeys:     removedTags.Keys(),
		}
		if _, err := conn.UntagResource(ctx, input); err != nil {
			return err
		}
	}

	if len(updatedTags) > 0 {
		input := &ssmincidents.TagResourceInput{
			ResourceArn: aws.String(arn),
			Tags:        updatedTags.IgnoreAWS().Map(),
		}

		if _, err := conn.TagResource(ctx, input); err != nil {
			return err
		}
	}

	return nil
}
