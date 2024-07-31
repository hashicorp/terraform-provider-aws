// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package identitystore

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/identitystore"
	"github.com/aws/aws-sdk-go-v2/service/identitystore/document"
	"github.com/aws/aws-sdk-go-v2/service/identitystore/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_identitystore_group")
func ResourceGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGroupCreate,
		ReadWithoutTimeout:   resourceGroupRead,
		UpdateWithoutTimeout: resourceGroupUpdate,
		DeleteWithoutTimeout: resourceGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			names.AttrDisplayName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"external_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrIssuer: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"identity_store_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

const (
	ResNameGroup = "Group"
)

func resourceGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).IdentityStoreClient(ctx)

	identityStoreId := d.Get("identity_store_id").(string)

	input := &identitystore.CreateGroupInput{
		IdentityStoreId: aws.String(identityStoreId),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrDisplayName); ok {
		input.DisplayName = aws.String(v.(string))
	}

	out, err := conn.CreateGroup(ctx, input)
	if err != nil {
		return create.AppendDiagError(diags, names.IdentityStore, create.ErrActionCreating, ResNameGroup, d.Get("identity_store_id").(string), err)
	}

	if out == nil || out.GroupId == nil {
		return create.AppendDiagError(diags, names.IdentityStore, create.ErrActionCreating, ResNameGroup, d.Get("identity_store_id").(string), errors.New("empty output"))
	}

	d.SetId(fmt.Sprintf("%s/%s", aws.ToString(out.IdentityStoreId), aws.ToString(out.GroupId)))

	return append(diags, resourceGroupRead(ctx, d, meta)...)
}

func resourceGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).IdentityStoreClient(ctx)

	identityStoreId, groupId, err := resourceGroupParseID(d.Id())
	if err != nil {
		return create.AppendDiagError(diags, names.IdentityStore, create.ErrActionReading, ResNameGroup, d.Id(), err)
	}

	out, err := FindGroupByTwoPartKey(ctx, conn, identityStoreId, groupId)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IdentityStore Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.IdentityStore, create.ErrActionReading, ResNameGroup, d.Id(), err)
	}

	d.Set(names.AttrDescription, out.Description)
	d.Set(names.AttrDisplayName, out.DisplayName)
	if err := d.Set("external_ids", flattenExternalIds(out.ExternalIds)); err != nil {
		return create.AppendDiagError(diags, names.IdentityStore, create.ErrActionSetting, ResNameGroup, d.Id(), err)
	}
	d.Set("group_id", out.GroupId)
	d.Set("identity_store_id", out.IdentityStoreId)

	return diags
}

func resourceGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).IdentityStoreClient(ctx)

	in := &identitystore.UpdateGroupInput{
		GroupId:         aws.String(d.Get("group_id").(string)),
		IdentityStoreId: aws.String(d.Get("identity_store_id").(string)),
		Operations:      nil,
	}

	if d.HasChange(names.AttrDescription) {
		in.Operations = append(in.Operations, types.AttributeOperation{
			AttributePath:  aws.String(names.AttrDescription),
			AttributeValue: document.NewLazyDocument(d.Get(names.AttrDescription).(string)),
		})
	}

	if d.HasChange(names.AttrDisplayName) {
		in.Operations = append(in.Operations, types.AttributeOperation{
			AttributePath:  aws.String("displayName"),
			AttributeValue: document.NewLazyDocument(d.Get(names.AttrDisplayName).(string)),
		})
	}

	_, err := conn.UpdateGroup(ctx, in)

	if err != nil {
		return create.AppendDiagError(diags, names.IdentityStore, create.ErrActionUpdating, ResNameGroup, d.Id(), err)
	}

	return append(diags, resourceGroupRead(ctx, d, meta)...)
}

func resourceGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).IdentityStoreClient(ctx)

	log.Printf("[INFO] Deleting IdentityStore Group %s", d.Id())
	_, err := conn.DeleteGroup(ctx, &identitystore.DeleteGroupInput{
		IdentityStoreId: aws.String(d.Get("identity_store_id").(string)),
		GroupId:         aws.String(d.Get("group_id").(string)),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.IdentityStore, create.ErrActionDeleting, ResNameGroup, d.Id(), err)
	}

	return diags
}

func FindGroupByTwoPartKey(ctx context.Context, conn *identitystore.Client, identityStoreID, groupID string) (*identitystore.DescribeGroupOutput, error) {
	in := &identitystore.DescribeGroupInput{
		GroupId:         aws.String(groupID),
		IdentityStoreId: aws.String(identityStoreID),
	}

	out, err := conn.DescribeGroup(ctx, in)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.GroupId == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func resourceGroupParseID(id string) (identityStoreId, groupId string, err error) {
	parts := strings.Split(id, "/")

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		err = errors.New("expected a resource id in the form: identity-store-id/group-id")
		return
	}

	return parts[0], parts[1], nil
}
