// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package grafana

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/managedgrafana"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_grafana_role_association")
func ResourceRoleAssociation() *schema.Resource {
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
			"role": {
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
	conn := meta.(*conns.AWSClient).GrafanaConn(ctx)

	role := d.Get("role").(string)
	workspaceID := d.Get("workspace_id").(string)

	updateInstructions := make([]*managedgrafana.UpdateInstruction, 0)
	if v, ok := d.GetOk("user_ids"); ok && v.(*schema.Set).Len() > 0 {
		typeSsoUser := managedgrafana.UserTypeSsoUser
		updateInstructions = populateUpdateInstructions(role, flex.ExpandStringSet(v.(*schema.Set)), managedgrafana.UpdateActionAdd, typeSsoUser, updateInstructions)
	}

	if v, ok := d.GetOk("group_ids"); ok && v.(*schema.Set).Len() > 0 {
		typeSsoUser := managedgrafana.UserTypeSsoGroup
		updateInstructions = populateUpdateInstructions(role, flex.ExpandStringSet(v.(*schema.Set)), managedgrafana.UpdateActionAdd, typeSsoUser, updateInstructions)
	}

	input := &managedgrafana.UpdatePermissionsInput{
		UpdateInstructionBatch: updateInstructions,
		WorkspaceId:            aws.String(workspaceID),
	}

	log.Printf("[DEBUG] Creating Grafana Workspace Role Association: %s", input)
	response, err := conn.UpdatePermissionsWithContext(ctx, input)

	for _, updateError := range response.Errors {
		return sdkdiag.AppendErrorf(diags, "creating Grafana Workspace Role Association: %s", aws.StringValue(updateError.Message))
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Grafana Workspace Role Association: %s", err)
	}

	if d.Id() == "" {
		d.SetId(fmt.Sprintf("%s/%s", workspaceID, role))
	}

	return append(diags, resourceRoleAssociationRead(ctx, d, meta)...)
}

func populateUpdateInstructions(role string, list []*string, action string, typeSsoUser string, updateInstructions []*managedgrafana.UpdateInstruction) []*managedgrafana.UpdateInstruction {
	users := make([]*managedgrafana.User, len(list))
	for i := 0; i < len(users); i++ {
		users[i] = &managedgrafana.User{
			Id:   list[i],
			Type: aws.String(typeSsoUser),
		}
	}
	updateInstructions = append(updateInstructions, &managedgrafana.UpdateInstruction{
		Action: aws.String(action),
		Role:   aws.String(role),
		Users:  users,
	})

	return updateInstructions
}

func resourceRoleAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GrafanaConn(ctx)

	roleAssociations, err := FindRoleAssociationsByRoleAndWorkspaceID(ctx, conn, d.Get("role").(string), d.Get("workspace_id").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Grafana Workspace Role Association %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Grafana Workspace Role Association (%s): %s", d.Id(), err)
	}

	d.Set("group_ids", roleAssociations[managedgrafana.UserTypeSsoGroup])
	d.Set("user_ids", roleAssociations[managedgrafana.UserTypeSsoUser])

	return diags
}

func resourceRoleAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GrafanaConn(ctx)

	updateInstructions := make([]*managedgrafana.UpdateInstruction, 0)
	if v, ok := d.GetOk("user_ids"); ok && v.(*schema.Set).Len() > 0 {
		typeSsoUser := managedgrafana.UserTypeSsoUser
		updateInstructions = populateUpdateInstructions(d.Get("role").(string), flex.ExpandStringSet(v.(*schema.Set)), managedgrafana.UpdateActionRevoke, typeSsoUser, updateInstructions)
	}

	if v, ok := d.GetOk("group_ids"); ok && v.(*schema.Set).Len() > 0 {
		typeSsoUser := managedgrafana.UserTypeSsoGroup
		updateInstructions = populateUpdateInstructions(d.Get("role").(string), flex.ExpandStringSet(v.(*schema.Set)), managedgrafana.UpdateActionRevoke, typeSsoUser, updateInstructions)
	}

	input := &managedgrafana.UpdatePermissionsInput{
		UpdateInstructionBatch: updateInstructions,
		WorkspaceId:            aws.String(d.Get("workspace_id").(string)),
	}

	log.Printf("[DEBUG] Deleting Grafana Workspace Role Association: %s", input)
	_, err := conn.UpdatePermissionsWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Grafana Workspace Role Association: %s", err)
	}

	return diags
}
