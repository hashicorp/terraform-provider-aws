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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ssmcontacts_contact", name="Context")
// @Tags(identifierAttribute="id")
func ResourceContact() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceContactCreate,
		ReadWithoutTimeout:   resourceContactRead,
		UpdateWithoutTimeout: resourceContactUpdate,
		DeleteWithoutTimeout: resourceContactDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

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
			names.AttrType: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameContact = "Contact"
)

func resourceContactCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*conns.AWSClient).SSMContactsClient(ctx)

	input := &ssmcontacts.CreateContactInput{
		Alias:       aws.String(d.Get(names.AttrAlias).(string)),
		DisplayName: aws.String(d.Get(names.AttrDisplayName).(string)),
		Plan:        &types.Plan{Stages: []types.Stage{}},
		Tags:        getTagsIn(ctx),
		Type:        types.ContactType(d.Get(names.AttrType).(string)),
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

func resourceContactRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

func resourceContactUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMContactsClient(ctx)

	if d.HasChanges(names.AttrDisplayName) {
		in := &ssmcontacts.UpdateContactInput{
			ContactId:   aws.String(d.Id()),
			DisplayName: aws.String(d.Get(names.AttrDisplayName).(string)),
		}

		_, err := conn.UpdateContact(ctx, in)
		if err != nil {
			return create.AppendDiagError(diags, names.SSMContacts, create.ErrActionUpdating, ResNameContact, d.Id(), err)
		}
	}

	return append(diags, resourceContactRead(ctx, d, meta)...)
}

func resourceContactDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
