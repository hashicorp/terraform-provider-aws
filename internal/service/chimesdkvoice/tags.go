package chimesdkvoice

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/chimesdkvoice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"reflect"
	"time"
)

const retryTimeout = 30 * time.Second

func ResourceTags() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTagsCreate,
		ReadWithoutTimeout:   resourceTagsList,
		DeleteWithoutTimeout: resourceTagsDelete,
		UpdateWithoutTimeout: resourceTagsUpdate,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"resource_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"tags": {
				Type:     schema.TypeMap,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func expandTags(l map[string]interface{}) []*chimesdkvoice.Tag {

	tags := make([]*chimesdkvoice.Tag, 0)

	for k, v := range l {
		tag := &chimesdkvoice.Tag{
			Key:   aws.String(k),
			Value: aws.String(v.(string)),
		}
		tags = append(tags, tag)
	}
	return tags
}

func flattenTags(tags []*chimesdkvoice.Tag) map[string]interface{} {
	rawTags := make(map[string]interface{}, 0)

	for _, tag := range tags {
		rawTags[aws.StringValue(tag.Key)] = *tag.Value
	}
	return rawTags
}

func resourceTagsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceConn()
	arn := d.Get("resource_arn").(string)

	currentTags, _ := ListTagsForResource(ctx, arn, meta)

	var tagsToDelete []*string

	for _, tag := range currentTags {
		tagsToDelete = append(tagsToDelete, tag.Key)
	}

	in := &chimesdkvoice.UntagResourceInput{
		ResourceARN: aws.String(arn),
		TagKeys:     tagsToDelete,
	}

	_, err := conn.UntagResourceWithContext(ctx, in)

	if tfawserr.ErrCodeEquals(err, chimesdkvoice.ErrCodeBadRequestException) {
		return nil
	}

	return nil
}

func ListTagsForResource(ctx context.Context, arn string, meta interface{}) ([]*chimesdkvoice.Tag, error) {
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceConn()

	in := &chimesdkvoice.ListTagsForResourceInput{
		ResourceARN: aws.String(arn),
	}

	response := &chimesdkvoice.ListTagsForResourceOutput{}

	err := tfresource.Retry(ctx, retryTimeout, func() *retry.RetryError {
		var err error
		response, err = conn.ListTagsForResourceWithContext(ctx, in)
		if err != nil {
			return retry.NonRetryableError(err)
		} else if len(response.Tags) == 0 {
			return retry.RetryableError(errors.New("no tag found yet"))
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return response.Tags, nil
}

func resourceTagsList(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	arn := d.Id()
	tags, _ := ListTagsForResource(ctx, arn, meta)
	d.Set("resource_arn", d.Id())
	d.Set("tags", flattenTags(tags))

	return nil
}

func validateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}, new_tags_map map[string]interface{}) diag.Diagnostics {
	arn := d.Get("resource_arn").(string)

	tfresource.Retry(ctx, retryTimeout, func() *retry.RetryError {
		tags, _ := ListTagsForResource(ctx, arn, meta)

		eq := reflect.DeepEqual(tags, new_tags_map)
		if !eq {
			return retry.RetryableError(errors.New("not updated yet"))
		} else {
			return nil
		}
	})
	return nil
}

func resourceTagsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	arn := d.Get("resource_arn").(string)
	currentTags, _ := ListTagsForResource(ctx, arn, meta)
	tagsToDelete := make(map[string]interface{}, 0)

	newTags := d.Get("tags")
	newTagsMap := newTags.(map[string]interface{})

	for _, tag := range currentTags {
		_, ok := newTagsMap[*tag.Key]
		if !ok {
			tagsToDelete[aws.StringValue(tag.Key)] = tag.Value
		}
	}

	d.Set("tags", tagsToDelete)
	resourceTagsDelete(ctx, d, meta)

	d.Set("tags", newTagsMap)
	resourceTagsCreate(ctx, d, meta)

	validateUpdate(ctx, d, meta, newTagsMap)

	return resourceTagsList(ctx, d, meta)
}

func validateTagResource(ctx context.Context, d *schema.ResourceData, meta interface{}, tagsMap map[string]interface{}) bool {
	resourceArn := d.Get("resource_arn").(string)

	err := tfresource.Retry(ctx, retryTimeout, func() *retry.RetryError {
		tags, _ := ListTagsForResource(ctx, resourceArn, meta)
		eq := reflect.DeepEqual(flattenTags(tags), tagsMap)
		if !eq {
			return retry.RetryableError(errors.New("tags not updated yet"))
		} else {
			return nil
		}
	})

	if err != nil {
		return false
	}
	return true
}

func resourceTagsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceConn()

	resourceArn := d.Get("resource_arn").(string)
	tags := d.Get("tags")
	tagsMap := tags.(map[string]interface{})

	in := &chimesdkvoice.TagResourceInput{
		ResourceARN: aws.String(resourceArn),
		Tags:        expandTags(tagsMap),
	}

	_, err := conn.TagResource(in)

	if err != nil {
		return diag.Errorf("Error creating tag resource %s", err)
	}

	d.SetId(resourceArn)

	// Validate that tags were successfully created
	tagsCreated := validateTagResource(ctx, d, meta, tagsMap)

	if !tagsCreated {
		return diag.Errorf("Failed to create tag resource")
	}

	return resourceTagsList(ctx, d, meta)
}
