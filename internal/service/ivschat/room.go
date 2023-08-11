// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ivschat

import (
	"context"
	"errors"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ivschat"
	"github.com/aws/aws-sdk-go-v2/service/ivschat/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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

// @SDKResource("aws_ivschat_room", name="Room")
// @Tags(identifierAttribute="id")
func ResourceRoom() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRoomCreate,
		ReadWithoutTimeout:   resourceRoomRead,
		UpdateWithoutTimeout: resourceRoomUpdate,
		DeleteWithoutTimeout: resourceRoomDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"logging_configuration_identifiers": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},
			"maximum_message_length": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(1, 500),
			},
			"maximum_message_rate_per_second": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(1, 10),
			},
			"message_review_handler": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"fallback_result": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice(fallbackResultValues(types.FallbackResult("").Values()), false),
						},
						"uri": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-_]{0,128}$`), "must contain only alphanumeric, hyphen, and underscore characters, with max length of 128 characters"),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameRoom = "Room"
)

func resourceRoomCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IVSChatClient(ctx)

	in := &ivschat.CreateRoomInput{
		Tags: getTagsIn(ctx),
	}

	if v, ok := d.GetOk("logging_configuration_identifiers"); ok {
		in.LoggingConfigurationIdentifiers = flex.ExpandStringValueList(v.([]interface{}))
	}

	if v, ok := d.GetOk("maximum_message_length"); ok {
		in.MaximumMessageLength = int32(v.(int))
	}

	if v, ok := d.GetOk("maximum_message_rate_per_second"); ok {
		in.MaximumMessageRatePerSecond = int32(v.(int))
	}

	if v, ok := d.GetOk("message_review_handler"); ok && len(v.([]interface{})) > 0 {
		in.MessageReviewHandler = expandMessageReviewHandler(v.([]interface{}))
	}

	if v, ok := d.GetOk("name"); ok {
		in.Name = aws.String(v.(string))
	}

	out, err := conn.CreateRoom(ctx, in)
	if err != nil {
		return create.DiagError(names.IVSChat, create.ErrActionCreating, ResNameRoom, d.Get("name").(string), err)
	}

	if out == nil {
		return create.DiagError(names.IVSChat, create.ErrActionCreating, ResNameRoom, d.Get("name").(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.Arn))

	if _, err := waitRoomCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.DiagError(names.IVSChat, create.ErrActionWaitingForCreation, ResNameRoom, d.Id(), err)
	}

	return resourceRoomRead(ctx, d, meta)
}

func resourceRoomRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IVSChatClient(ctx)

	out, err := findRoomByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IVSChat Room (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.IVSChat, create.ErrActionReading, ResNameRoom, d.Id(), err)
	}

	d.Set("arn", out.Arn)

	if err := d.Set("logging_configuration_identifiers", flex.FlattenStringValueList(out.LoggingConfigurationIdentifiers)); err != nil {
		return create.DiagError(names.IVSChat, create.ErrActionSetting, ResNameRoom, d.Id(), err)
	}

	d.Set("maximum_message_length", out.MaximumMessageLength)
	d.Set("maximum_message_rate_per_second", out.MaximumMessageRatePerSecond)

	if err := d.Set("message_review_handler", flattenMessageReviewHandler(out.MessageReviewHandler)); err != nil {
		return create.DiagError(names.IVSChat, create.ErrActionSetting, ResNameRoom, d.Id(), err)
	}

	d.Set("name", out.Name)

	return nil
}

func resourceRoomUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IVSChatClient(ctx)

	update := false

	in := &ivschat.UpdateRoomInput{
		Identifier: aws.String(d.Id()),
	}

	if d.HasChanges("logging_configuration_identifiers") {
		in.LoggingConfigurationIdentifiers = flex.ExpandStringValueList(d.Get("logging_configuration_identifiers").([]interface{}))
		update = true
	}

	if d.HasChanges("maximum_message_length") {
		in.MaximumMessageLength = int32(d.Get("maximum_message_length").(int))
		update = true
	}

	if d.HasChanges("maximum_message_rate_per_second") {
		in.MaximumMessageRatePerSecond = int32(d.Get("maximum_message_rate_per_second").(int))
		update = true
	}

	if d.HasChanges("message_review_handler") {
		in.MessageReviewHandler = expandMessageReviewHandler(d.Get("message_review_handler").([]interface{}))
		update = true
	}

	if d.HasChanges("name") {
		in.Name = aws.String(d.Get("name").(string))
		update = true
	}

	if !update {
		return nil
	}

	log.Printf("[DEBUG] Updating IVSChat Room (%s): %#v", d.Id(), in)
	out, err := conn.UpdateRoom(ctx, in)
	if err != nil {
		return create.DiagError(names.IVSChat, create.ErrActionUpdating, ResNameRoom, d.Id(), err)
	}

	if _, err := waitRoomUpdated(ctx, conn, aws.ToString(out.Arn), d.Timeout(schema.TimeoutUpdate), in); err != nil {
		return create.DiagError(names.IVSChat, create.ErrActionWaitingForUpdate, ResNameRoom, d.Id(), err)
	}

	return resourceRoomRead(ctx, d, meta)
}

func resourceRoomDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IVSChatClient(ctx)

	log.Printf("[INFO] Deleting IVSChat Room %s", d.Id())

	_, err := conn.DeleteRoom(ctx, &ivschat.DeleteRoomInput{
		Identifier: aws.String(d.Id()),
	})

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return create.DiagError(names.IVSChat, create.ErrActionDeleting, ResNameRoom, d.Id(), err)
	}

	if _, err := waitRoomDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.DiagError(names.IVSChat, create.ErrActionWaitingForDeletion, ResNameRoom, d.Id(), err)
	}

	return nil
}

func flattenMessageReviewHandler(apiObject *types.MessageReviewHandler) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.FallbackResult; v != "" {
		m["fallback_result"] = v
	}

	if v := apiObject.Uri; v != nil {
		m["uri"] = aws.ToString(v)
	}

	return []interface{}{m}
}

func expandMessageReviewHandler(vSettings []interface{}) *types.MessageReviewHandler {
	if len(vSettings) == 0 || vSettings[0] == nil {
		return nil
	}

	tfMap := vSettings[0].(map[string]interface{})

	messageReviewHandler := &types.MessageReviewHandler{}

	if v, ok := tfMap["fallback_result"].(string); ok && v != "" {
		messageReviewHandler.FallbackResult = types.FallbackResult(v)
	}

	if v, ok := tfMap["uri"].(string); ok {
		messageReviewHandler.Uri = aws.String(v)
	}

	return messageReviewHandler
}

func fallbackResultValues(in []types.FallbackResult) []string {
	var out []string

	for _, v := range in {
		out = append(out, string(v))
	}

	return out
}
