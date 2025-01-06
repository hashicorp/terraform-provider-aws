// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package grafana

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/grafana"
	awstypes "github.com/aws/aws-sdk-go-v2/service/grafana/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_grafana_role_association", name="Workspace Role Association")
func resourceRoleAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRoleAssociationUpsert,
		ReadWithoutTimeout:   resourceRoleAssociationRead,
		UpdateWithoutTimeout: resourceRoleAssociationUpsert,
		DeleteWithoutTimeout: resourceRoleAssociationDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrRole: {
				Type:     schema.TypeString,
				Required: true,
			},
			"user_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"workspace_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceRoleAssociationUpsert(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GrafanaClient(ctx)

	role := awstypes.Role(d.Get(names.AttrRole).(string))
	workspaceID := d.Get("workspace_id").(string)
	id := fmt.Sprintf("%s/%s", workspaceID, role)

	updateInstructions := make([]awstypes.UpdateInstruction, 0)
	if v, ok := d.GetOk("user_ids"); ok && v.(*schema.Set).Len() > 0 {
		updateInstructions = populateUpdateInstructions(role, flex.ExpandStringSet(v.(*schema.Set)), awstypes.UpdateActionAdd, awstypes.UserTypeSsoUser, updateInstructions)
	}

	if v, ok := d.GetOk("group_ids"); ok && v.(*schema.Set).Len() > 0 {
		updateInstructions = populateUpdateInstructions(role, flex.ExpandStringSet(v.(*schema.Set)), awstypes.UpdateActionAdd, awstypes.UserTypeSsoGroup, updateInstructions)
	}

	input := &grafana.UpdatePermissionsInput{
		UpdateInstructionBatch: updateInstructions,
		WorkspaceId:            aws.String(workspaceID),
	}

	output, err := conn.UpdatePermissions(ctx, input)

	if err == nil {
		err = updatesError(output.Errors)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Grafana Workspace Role Association (%s): %s", id, err)
	}

	if d.IsNewResource() {
		d.SetId(id)
	}

	return append(diags, resourceRoleAssociationRead(ctx, d, meta)...)
}

func resourceRoleAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GrafanaClient(ctx)

	role := awstypes.Role(d.Get(names.AttrRole).(string))
	workspaceID := d.Get("workspace_id").(string)
	roleAssociations, err := findRoleAssociationsByTwoPartKey(ctx, conn, role, workspaceID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Grafana Workspace Role Association %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Grafana Workspace Role Association (%s): %s", d.Id(), err)
	}

	d.Set("group_ids", roleAssociations[awstypes.UserTypeSsoGroup])
	d.Set("user_ids", roleAssociations[awstypes.UserTypeSsoUser])

	return diags
}

func resourceRoleAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GrafanaClient(ctx)

	role := awstypes.Role(d.Get(names.AttrRole).(string))
	workspaceID := d.Get("workspace_id").(string)

	updateInstructions := make([]awstypes.UpdateInstruction, 0)
	if v, ok := d.GetOk("user_ids"); ok && v.(*schema.Set).Len() > 0 {
		updateInstructions = populateUpdateInstructions(role, flex.ExpandStringSet(v.(*schema.Set)), awstypes.UpdateActionRevoke, awstypes.UserTypeSsoUser, updateInstructions)
	}

	if v, ok := d.GetOk("group_ids"); ok && v.(*schema.Set).Len() > 0 {
		updateInstructions = populateUpdateInstructions(role, flex.ExpandStringSet(v.(*schema.Set)), awstypes.UpdateActionRevoke, awstypes.UserTypeSsoGroup, updateInstructions)
	}

	input := &grafana.UpdatePermissionsInput{
		UpdateInstructionBatch: updateInstructions,
		WorkspaceId:            aws.String(workspaceID),
	}

	log.Printf("[DEBUG] Deleting Grafana Workspace Role Association: %s", d.Id())
	output, err := conn.UpdatePermissions(ctx, input)

	if err == nil {
		err = updatesError(output.Errors)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Grafana Workspace Role Association: %s", err)
	}

	return diags
}

func findRoleAssociationsByTwoPartKey(ctx context.Context, conn *grafana.Client, role awstypes.Role, workspaceID string) (map[awstypes.UserType][]string, error) {
	input := &grafana.ListPermissionsInput{
		WorkspaceId: aws.String(workspaceID),
	}
	output := make(map[awstypes.UserType][]string, 0)

	pages := grafana.NewListPermissionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.Permissions {
			if v.Role == role {
				userType := v.User.Type
				output[userType] = append(output[userType], aws.ToString(v.User.Id))
			}
		}
	}

	if len(output) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func populateUpdateInstructions(role awstypes.Role, list []*string, action awstypes.UpdateAction, typeSSOUser awstypes.UserType, updateInstructions []awstypes.UpdateInstruction) []awstypes.UpdateInstruction {
	users := make([]awstypes.User, len(list))
	for i := 0; i < len(users); i++ {
		users[i] = awstypes.User{
			Id:   list[i],
			Type: typeSSOUser,
		}
	}
	updateInstructions = append(updateInstructions, awstypes.UpdateInstruction{
		Action: action,
		Role:   role,
		Users:  users,
	})

	return updateInstructions
}
