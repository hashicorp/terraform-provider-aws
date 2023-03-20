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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_oam_link")
func ResourceLink() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLinkCreate,
		ReadWithoutTimeout:   resourceLinkRead,
		UpdateWithoutTimeout: resourceLinkUpdate,
		DeleteWithoutTimeout: resourceLinkDelete,

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
			"label": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"label_template": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"link_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_types": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				MaxItems: 50,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(ResourceTypeValues(types.ResourceType("").Values()), false),
				},
				Set: schema.HashString,
			},
			"sink_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sink_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameLink = "Link"
)

func resourceLinkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ObservabilityAccessManagerClient()

	in := &oam.CreateLinkInput{
		LabelTemplate:  aws.String(d.Get("label_template").(string)),
		ResourceTypes:  ExpandResourceTypes(d.Get("resource_types").(*schema.Set).List()),
		SinkIdentifier: aws.String(d.Get("sink_identifier").(string)),
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(ctx, d.Get("tags").(map[string]interface{})))

	if len(tags) > 0 {
		in.Tags = Tags(tags.IgnoreAWS())
	}

	out, err := conn.CreateLink(ctx, in)
	if err != nil {
		return create.DiagError(names.ObservabilityAccessManager, create.ErrActionCreating, ResNameLink, d.Get("sink_identifier").(string), err)
	}

	if out == nil || out.Id == nil {
		return create.DiagError(names.ObservabilityAccessManager, create.ErrActionCreating, ResNameLink, d.Get("sink_identifier").(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.Arn))

	return resourceLinkRead(ctx, d, meta)
}

func resourceLinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ObservabilityAccessManagerClient()

	out, err := findLinkByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ObservabilityAccessManager Link (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.ObservabilityAccessManager, create.ErrActionReading, ResNameLink, d.Id(), err)
	}

	d.Set("arn", out.Arn)
	d.Set("label", out.Label)
	d.Set("label_template", out.LabelTemplate)
	d.Set("link_id", out.Id)
	d.Set("resource_types", flex.FlattenStringValueList(out.ResourceTypes))
	d.Set("sink_arn", out.SinkArn)
	if _, ok := d.GetOk("sink_identifier"); !ok {
		d.Set("sink_identifier", out.SinkArn)
	}

	tags, err := ListTags(ctx, conn, d.Id())
	if err != nil {
		return create.DiagError(names.ObservabilityAccessManager, create.ErrActionReading, ResNameLink, d.Id(), err)
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return create.DiagError(names.ObservabilityAccessManager, create.ErrActionSetting, ResNameLink, d.Id(), err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return create.DiagError(names.ObservabilityAccessManager, create.ErrActionSetting, ResNameLink, d.Id(), err)
	}

	return nil
}

func resourceLinkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ObservabilityAccessManagerClient()

	update := false

	in := &oam.UpdateLinkInput{
		Identifier: aws.String(d.Id()),
	}

	if d.HasChanges("resource_types") {
		in.ResourceTypes = ExpandResourceTypes(d.Get("resource_types").(*schema.Set).List())
		update = true
	}

	if d.HasChange("tags_all") {
		log.Printf("[DEBUG] Updating ObservabilityAccessManager Link Tags (%s): %#v", d.Id(), d.Get("tags_all"))
		oldTags, newTags := d.GetChange("tags_all")
		if err := UpdateTags(ctx, conn, d.Get("arn").(string), oldTags, newTags); err != nil {
			return create.DiagError(names.ObservabilityAccessManager, create.ErrActionUpdating, ResNameLink, d.Id(), err)
		}
	}

	if update {
		log.Printf("[DEBUG] Updating ObservabilityAccessManager Link (%s): %#v", d.Id(), in)
		_, err := conn.UpdateLink(ctx, in)
		if err != nil {
			return create.DiagError(names.ObservabilityAccessManager, create.ErrActionUpdating, ResNameLink, d.Id(), err)
		}
	}

	return resourceLinkRead(ctx, d, meta)
}

func resourceLinkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ObservabilityAccessManagerClient()

	log.Printf("[INFO] Deleting ObservabilityAccessManager Link %s", d.Id())

	_, err := conn.DeleteLink(ctx, &oam.DeleteLinkInput{
		Identifier: aws.String(d.Id()),
	})

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return create.DiagError(names.ObservabilityAccessManager, create.ErrActionDeleting, ResNameLink, d.Id(), err)
	}

	return nil
}

func findLinkByID(ctx context.Context, conn *oam.Client, id string) (*oam.GetLinkOutput, error) {
	in := &oam.GetLinkInput{
		Identifier: aws.String(id),
	}
	out, err := conn.GetLink(ctx, in)
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

	if out == nil || out.Arn == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func ExpandResourceTypes(resourceTypeList []interface{}) []types.ResourceType {
	if len(resourceTypeList) == 0 {
		return nil
	}

	var resourceTypes []types.ResourceType

	for _, resourceTypeString := range resourceTypeList {
		if resourceTypeString == nil {
			continue
		}

		resourceType := types.ResourceType(resourceTypeString.(string))
		resourceTypes = append(resourceTypes, resourceType)
	}

	return resourceTypes
}

func ResourceTypeValues(resourceTypes []types.ResourceType) []string {
	var out []string

	for _, v := range resourceTypes {
		out = append(out, string(v))
	}

	return out
}
