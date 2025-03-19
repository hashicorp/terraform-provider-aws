// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sesv2

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sesv2_contact_list", name="Contact List")
// @Tags(identifierAttribute="arn")
// @Testing(serialize=true)
func resourceContactList() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceContactListCreate,
		ReadWithoutTimeout:   resourceContactListRead,
		UpdateWithoutTimeout: resourceContactListUpdate,
		DeleteWithoutTimeout: resourceContactListDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"contact_list_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"created_timestamp": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"last_updated_timestamp": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"topic": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"default_subscription_status": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.SubscriptionStatus](),
						},
						names.AttrDescription: {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrDisplayName: {
							Type:     schema.TypeString,
							Required: true,
						},
						"topic_name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

const (
	resNameContactList = "Contact List"
)

func resourceContactListCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	in := &sesv2.CreateContactListInput{
		ContactListName: aws.String(d.Get("contact_list_name").(string)),
		Tags:            getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		in.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("topic"); ok && v.(*schema.Set).Len() > 0 {
		in.Topics = expandTopics(v.(*schema.Set).List())
	}

	out, err := conn.CreateContactList(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionCreating, resNameContactList, d.Get("contact_list_name").(string), err)
	}

	if out == nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionCreating, resNameContactList, d.Get("contact_list_name").(string), errors.New("empty output"))
	}

	d.SetId(d.Get("contact_list_name").(string))

	return append(diags, resourceContactListRead(ctx, d, meta)...)
}

func resourceContactListRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	out, err := findContactListByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SESV2 ContactList (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionReading, resNameContactList, d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition(ctx),
		Service:   "ses",
		Region:    meta.(*conns.AWSClient).Region(ctx),
		AccountID: meta.(*conns.AWSClient).AccountID(ctx),
		Resource:  fmt.Sprintf("contact-list/%s", d.Id()),
	}.String()

	d.Set(names.AttrARN, arn)
	d.Set("contact_list_name", out.ContactListName)
	d.Set("created_timestamp", aws.ToTime(out.CreatedTimestamp).Format(time.RFC3339))
	d.Set(names.AttrDescription, out.Description)
	d.Set("last_updated_timestamp", aws.ToTime(out.LastUpdatedTimestamp).Format(time.RFC3339))

	if err := d.Set("topic", flattenTopics(out.Topics)); err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionSetting, resNameContactList, d.Id(), err)
	}

	return diags
}

func resourceContactListUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	in := &sesv2.UpdateContactListInput{
		ContactListName: aws.String(d.Id()),
	}

	if d.HasChanges(names.AttrDescription, "topic") {
		in.Description = aws.String(d.Get(names.AttrDescription).(string))
		in.Topics = expandTopics(d.Get("topic").(*schema.Set).List())

		log.Printf("[DEBUG] Updating SESV2 ContactList (%s): %#v", d.Id(), in)
		if _, err := conn.UpdateContactList(ctx, in); err != nil {
			return create.AppendDiagError(diags, names.SESV2, create.ErrActionUpdating, resNameContactList, d.Id(), err)
		}
	}

	return append(diags, resourceContactListRead(ctx, d, meta)...)
}

func resourceContactListDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	log.Printf("[INFO] Deleting SESV2 ContactList %s", d.Id())

	_, err := conn.DeleteContactList(ctx, &sesv2.DeleteContactListInput{
		ContactListName: aws.String(d.Id()),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionDeleting, resNameContactList, d.Id(), err)
	}

	return diags
}

func findContactListByID(ctx context.Context, conn *sesv2.Client, id string) (*sesv2.GetContactListOutput, error) {
	input := &sesv2.GetContactListInput{
		ContactListName: aws.String(id),
	}

	return findContactList(ctx, conn, input)
}

func findContactList(ctx context.Context, conn *sesv2.Client, input *sesv2.GetContactListInput) (*sesv2.GetContactListOutput, error) {
	output, err := conn.GetContactList(ctx, input)

	if errs.IsA[*types.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func expandTopics(tfList []any) []types.Topic {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.Topic

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)

		if !ok {
			continue
		}

		apiObject := expandTopic(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandTopic(tfMap map[string]any) *types.Topic {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.Topic{}

	if v, ok := tfMap["default_subscription_status"].(string); ok && v != "" {
		apiObject.DefaultSubscriptionStatus = types.SubscriptionStatus(v)
	}

	if v, ok := tfMap[names.AttrDescription].(string); ok && v != "" {
		apiObject.Description = aws.String(v)
	}

	if v, ok := tfMap[names.AttrDisplayName].(string); ok && v != "" {
		apiObject.DisplayName = aws.String(v)
	}

	if v, ok := tfMap["topic_name"].(string); ok && v != "" {
		apiObject.TopicName = aws.String(v)
	}

	return apiObject
}

func flattenTopics(apiObjects []types.Topic) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenTopic(&apiObject))
	}

	return tfList
}

func flattenTopic(apiObject *types.Topic) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"default_subscription_status": string(apiObject.DefaultSubscriptionStatus),
	}

	if v := apiObject.Description; v != nil {
		tfMap[names.AttrDescription] = aws.ToString(v)
	}

	if v := apiObject.DisplayName; v != nil {
		tfMap[names.AttrDisplayName] = aws.ToString(v)
	}

	if v := apiObject.TopicName; v != nil {
		tfMap["topic_name"] = aws.ToString(v)
	}

	return tfMap
}
