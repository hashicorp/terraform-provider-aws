// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package identitystore

import (
	"context"
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
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_identitystore_group", name="Group")
func resourceGroup() *schema.Resource {
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

func resourceGroupCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IdentityStoreClient(ctx)

	displayName := d.Get(names.AttrDisplayName).(string)
	identityStoreID := d.Get("identity_store_id").(string)
	input := &identitystore.CreateGroupInput{
		DisplayName:     aws.String(displayName),
		IdentityStoreId: aws.String(identityStoreID),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IdentityStore Group (%s): %s", displayName, err)
	}

	d.SetId(groupCreateResourceID(identityStoreID, aws.ToString(output.GroupId)))

	return append(diags, resourceGroupRead(ctx, d, meta)...)
}

func resourceGroupRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IdentityStoreClient(ctx)

	identityStoreID, groupID, err := groupParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	out, err := findGroupByTwoPartKey(ctx, conn, identityStoreID, groupID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IdentityStore Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IdentityStore Group (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrDescription, out.Description)
	d.Set(names.AttrDisplayName, out.DisplayName)
	if err := d.Set("external_ids", flattenExternalIDs(out.ExternalIds)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting external_ids: %s", err)
	}
	d.Set("group_id", out.GroupId)
	d.Set("identity_store_id", out.IdentityStoreId)

	return diags
}

func resourceGroupUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IdentityStoreClient(ctx)

	identityStoreID, groupID, err := groupParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &identitystore.UpdateGroupInput{
		GroupId:         aws.String(groupID),
		IdentityStoreId: aws.String(identityStoreID),
	}

	if d.HasChange(names.AttrDescription) {
		input.Operations = append(input.Operations, types.AttributeOperation{
			AttributePath:  aws.String(names.AttrDescription),
			AttributeValue: document.NewLazyDocument(d.Get(names.AttrDescription).(string)),
		})
	}

	if d.HasChange(names.AttrDisplayName) {
		input.Operations = append(input.Operations, types.AttributeOperation{
			AttributePath:  aws.String("displayName"),
			AttributeValue: document.NewLazyDocument(d.Get(names.AttrDisplayName).(string)),
		})
	}

	_, err = conn.UpdateGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating IdentityStore Group (%s): %s", d.Id(), err)
	}

	return append(diags, resourceGroupRead(ctx, d, meta)...)
}

func resourceGroupDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IdentityStoreClient(ctx)

	identityStoreID, groupID, err := groupParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting IdentityStore Group: %s", d.Id())
	_, err = conn.DeleteGroup(ctx, &identitystore.DeleteGroupInput{
		GroupId:         aws.String(groupID),
		IdentityStoreId: aws.String(identityStoreID),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IdentityStore Group (%s): %s", d.Id(), err)
	}

	return diags
}

const groupResourceIDSeparator = "/"

func groupCreateResourceID(identityStoreID, groupID string) string {
	parts := []string{identityStoreID, groupID}
	id := strings.Join(parts, groupResourceIDSeparator)

	return id
}

func groupParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, groupResourceIDSeparator)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected identity-store-id%[2]sgroup-id", id, groupResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findGroupByTwoPartKey(ctx context.Context, conn *identitystore.Client, identityStoreID, groupID string) (*identitystore.DescribeGroupOutput, error) {
	input := &identitystore.DescribeGroupInput{
		GroupId:         aws.String(groupID),
		IdentityStoreId: aws.String(identityStoreID),
	}

	return findGroup(ctx, conn, input)
}

func findGroup(ctx context.Context, conn *identitystore.Client, input *identitystore.DescribeGroupInput) (*identitystore.DescribeGroupOutput, error) {
	output, err := conn.DescribeGroup(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
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
