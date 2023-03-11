package oam

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/oam"
	"github.com/aws/aws-sdk-go-v2/service/oam/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_oam_sink")
func ResourceSink() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSinkCreate,
		ReadWithoutTimeout:   resourceSinkRead,
		UpdateWithoutTimeout: resourceSinkUpdate,
		DeleteWithoutTimeout: resourceSinkDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(1 * time.Minute),
			Delete: schema.DefaultTimeout(1 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"sink_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameSink = "Sink"
)

func resourceSinkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ObservabilityAccessManagerClient()

	in := &oam.CreateSinkInput{
		Name: aws.String(d.Get("name").(string)),
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(ctx, d.Get("tags").(map[string]interface{})))

	if len(tags) > 0 {
		in.Tags = Tags(tags.IgnoreAWS())
	}

	out, err := conn.CreateSink(ctx, in)
	if err != nil {
		return create.DiagError(names.ObservabilityAccessManager, create.ErrActionCreating, ResNameSink, d.Get("name").(string), err)
	}

	if out == nil {
		return create.DiagError(names.ObservabilityAccessManager, create.ErrActionCreating, ResNameSink, d.Get("name").(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.Arn))

	return resourceSinkRead(ctx, d, meta)
}

func resourceSinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ObservabilityAccessManagerClient()

	out, err := findSinkByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ObservabilityAccessManager Sink (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.ObservabilityAccessManager, create.ErrActionReading, ResNameSink, d.Id(), err)
	}

	d.Set("arn", out.Arn)
	d.Set("name", out.Name)
	d.Set("sink_id", out.Id)

	tags, err := ListTags(ctx, conn, d.Id())
	if err != nil {
		return create.DiagError(names.ObservabilityAccessManager, create.ErrActionReading, ResNameSink, d.Id(), err)
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return create.DiagError(names.ObservabilityAccessManager, create.ErrActionSetting, ResNameSink, d.Id(), err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return create.DiagError(names.ObservabilityAccessManager, create.ErrActionSetting, ResNameSink, d.Id(), err)
	}

	return nil
}

func resourceSinkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ObservabilityAccessManagerClient()

	if d.HasChange("tags_all") {
		log.Printf("[DEBUG] Updating ObservabilityAccessManager Sink Tags (%s): %#v", d.Id(), d.Get("tags_all"))
		oldTags, newTags := d.GetChange("tags_all")
		if err := UpdateTags(ctx, conn, d.Get("arn").(string), oldTags, newTags); err != nil {
			return create.DiagError(names.ObservabilityAccessManager, create.ErrActionUpdating, ResNameSink, d.Id(), err)
		}

		return resourceSinkRead(ctx, d, meta)
	}

	return nil
}

func resourceSinkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ObservabilityAccessManagerClient()

	log.Printf("[INFO] Deleting ObservabilityAccessManager Sink %s", d.Id())

	_, err := conn.DeleteSink(ctx, &oam.DeleteSinkInput{
		Identifier: aws.String(d.Id()),
	})

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return create.DiagError(names.ObservabilityAccessManager, create.ErrActionDeleting, ResNameSink, d.Id(), err)
	}

	return nil
}

func findSinkByID(ctx context.Context, conn *oam.Client, id string) (*oam.GetSinkOutput, error) {
	in := &oam.GetSinkInput{
		Identifier: aws.String(id),
	}
	out, err := conn.GetSink(ctx, in)
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &resource.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}
