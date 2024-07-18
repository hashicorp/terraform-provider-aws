// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package synthetics

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/synthetics"
	awstypes "github.com/aws/aws-sdk-go-v2/service/synthetics/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_synthetics_group_association", name="Group Association")
func ResourceGroupAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGroupAssociationCreate,
		ReadWithoutTimeout:   resourceGroupAssociationRead,
		DeleteWithoutTimeout: resourceGroupAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"canary_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"group_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrGroupName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceGroupAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SyntheticsClient(ctx)

	canaryArn := d.Get("canary_arn").(string)
	groupName := d.Get(names.AttrGroupName).(string)

	in := &synthetics.AssociateResourceInput{
		ResourceArn:     aws.String(canaryArn),
		GroupIdentifier: aws.String(groupName),
	}

	out, err := conn.AssociateResource(ctx, in)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "associating canary (%s) with group (%s): %s", canaryArn, groupName, err)
	}

	if out == nil {
		return sdkdiag.AppendErrorf(diags, "associating canary (%s) with group (%s): Empty output", canaryArn, groupName)
	}

	d.SetId(GroupAssociationCreateResourceID(canaryArn, groupName))

	return append(diags, resourceGroupAssociationRead(ctx, d, meta)...)
}

func resourceGroupAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SyntheticsClient(ctx)

	canaryArn, groupName, err := GroupAssociationParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	group, err := FindAssociatedGroup(ctx, conn, canaryArn, groupName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Synthetics Group Association between canary (%s) and group (%s) not found, removing from state", canaryArn, groupName)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Synthetics Group Associations for canary (%s): %s", canaryArn, err)
	}

	d.Set("canary_arn", canaryArn)
	d.Set("group_arn", group.Arn)
	d.Set("group_id", group.Id)
	d.Set(names.AttrGroupName, group.Name)

	return diags
}

func resourceGroupAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SyntheticsClient(ctx)

	log.Printf("[DEBUG] Deleting Synthetics Group Association %s", d.Id())

	canaryArn, groupName, err := GroupAssociationParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	in := &synthetics.DisassociateResourceInput{
		ResourceArn:     aws.String(canaryArn),
		GroupIdentifier: aws.String(groupName),
	}

	_, err = conn.DisassociateResource(ctx, in)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "does not exist in group") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disassociating Synthetics Group (%s) from canary (%s): %s", groupName, canaryArn, err)
	}

	return diags
}

const groupAssociationResourceIDSeparator = ","

func GroupAssociationCreateResourceID(canaryArn, groupName string) string {
	parts := []string{canaryArn, groupName}
	id := strings.Join(parts, groupAssociationResourceIDSeparator)

	return id
}

func GroupAssociationParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, groupAssociationResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected CanaryArn%[2]sGroupName", id, groupAssociationResourceIDSeparator)
}
