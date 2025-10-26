// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssmcontacts

import (
	"context"
	"errors"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssmcontacts"
	"github.com/aws/aws-sdk-go-v2/service/ssmcontacts/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ssmcontacts_contact", name="Contact")
// @ArnIdentity
// @Tags(identifierAttribute="arn")
// @Testing(skipEmptyTags=true, skipNullTags=true)
// @Testing(identityRegionOverrideTest=false)
// @Testing(serialize=true)
// @Testing(preIdentityVersion="v6.15.0")
func ResourceContact() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceContactCreate,
		ReadWithoutTimeout:   resourceContactRead,
		UpdateWithoutTimeout: resourceContactUpdate,
		DeleteWithoutTimeout: resourceContactDelete,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrAlias: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrDisplayName: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"rotation_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			names.AttrType: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

const (
	ResNameContact = "Contact"
)

func resourceContactCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*conns.AWSClient).SSMContactsClient(ctx)

	contactType := types.ContactType(d.Get(names.AttrType).(string))

	input := &ssmcontacts.CreateContactInput{
		Alias:       aws.String(d.Get(names.AttrAlias).(string)),
		DisplayName: aws.String(d.Get(names.AttrDisplayName).(string)),
		Tags:        getTagsIn(ctx),
		Type:        contactType,
	}

	if contactType == types.ContactTypeOncallSchedule {
		plan := &types.Plan{}
		if v, ok := d.GetOk("rotation_ids"); ok {
			plan.RotationIds = flex.ExpandStringValueList(v.([]any))
		}
		input.Plan = plan
	} else {
		input.Plan = &types.Plan{Stages: []types.Stage{}}
	}

	output, err := client.CreateContact(ctx, input)
	if err != nil {
		return create.AppendDiagError(diags, names.SSMContacts, create.ErrActionCreating, ResNameContact, d.Get(names.AttrAlias).(string), err)
	}

	if output == nil {
		return create.AppendDiagError(diags, names.SSMContacts, create.ErrActionCreating, ResNameContact, d.Get(names.AttrAlias).(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(output.ContactArn))

	return append(diags, resourceContactRead(ctx, d, meta)...)
}

func resourceContactRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMContactsClient(ctx)

	out, err := findContactByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SSMContacts Contact (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.SSMContacts, create.ErrActionReading, ResNameContact, d.Id(), err)
	}

	if err := setContactResourceData(d, out); err != nil {
		return create.AppendDiagError(diags, names.SSMContacts, create.ErrActionSetting, ResNameContact, d.Id(), err)
	}

	return diags
}

func resourceContactUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMContactsClient(ctx)

	if d.HasChanges(names.AttrDisplayName, "rotation_ids") {
		contactType := types.ContactType(d.Get(names.AttrType).(string))

		in := &ssmcontacts.UpdateContactInput{
			ContactId: aws.String(d.Id()),
		}

		if d.HasChange(names.AttrDisplayName) {
			in.DisplayName = aws.String(d.Get(names.AttrDisplayName).(string))
		}

		if d.HasChange("rotation_ids") && contactType == types.ContactTypeOncallSchedule {
			plan := &types.Plan{}
			if v, ok := d.GetOk("rotation_ids"); ok {
				plan.RotationIds = flex.ExpandStringValueList(v.([]any))
			}
			in.Plan = plan
		}

		_, err := conn.UpdateContact(ctx, in)
		if err != nil {
			return create.AppendDiagError(diags, names.SSMContacts, create.ErrActionUpdating, ResNameContact, d.Id(), err)
		}
	}

	return append(diags, resourceContactRead(ctx, d, meta)...)
}

func resourceContactDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMContactsClient(ctx)

	log.Printf("[INFO] Deleting SSMContacts Contact %s", d.Id())

	_, err := conn.DeleteContact(ctx, &ssmcontacts.DeleteContactInput{
		ContactId: aws.String(d.Id()),
	})

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return diags
		}

		return create.AppendDiagError(diags, names.SSMContacts, create.ErrActionDeleting, ResNameContact, d.Id(), err)
	}
	return diags
}
