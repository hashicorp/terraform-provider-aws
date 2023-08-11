// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package grafana

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/managedgrafana"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_grafana_workspace_api_key")
func ResourceWorkspaceAPIKey() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceWorkspaceAPIKeyCreate,
		ReadWithoutTimeout:   schema.NoopContext,
		UpdateWithoutTimeout: schema.NoopContext,
		DeleteWithoutTimeout: resourceWorkspaceAPIKeyDelete,

		Schema: map[string]*schema.Schema{
			"key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"key_role": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(managedgrafana.Role_Values(), false),
			},
			"seconds_to_live": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(1, 2592000),
			},
			"workspace_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceWorkspaceAPIKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GrafanaConn(ctx)

	keyName := d.Get("key_name").(string)
	workspaceID := d.Get("workspace_id").(string)
	id := WorkspaceAPIKeyCreateResourceID(workspaceID, keyName)
	input := &managedgrafana.CreateWorkspaceApiKeyInput{
		KeyName:       aws.String(keyName),
		KeyRole:       aws.String(d.Get("key_role").(string)),
		SecondsToLive: aws.Int64(int64(d.Get("seconds_to_live").(int))),
		WorkspaceId:   aws.String(workspaceID),
	}

	log.Printf("[DEBUG] Creating Grafana Workspace API Key: %s", input)
	output, err := conn.CreateWorkspaceApiKeyWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Grafana Workspace API Key (%s): %s", id, err)
	}

	d.SetId(id)
	d.Set("key", output.Key)

	return diags
}

func resourceWorkspaceAPIKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GrafanaConn(ctx)

	workspaceID, keyName, err := WorkspaceAPIKeyParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Grafana Workspace API Key (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Deleting Grafana Workspace API Key: %s", d.Id())
	_, err = conn.DeleteWorkspaceApiKeyWithContext(ctx, &managedgrafana.DeleteWorkspaceApiKeyInput{
		KeyName:     aws.String(keyName),
		WorkspaceId: aws.String(workspaceID),
	})

	if tfawserr.ErrCodeEquals(err, managedgrafana.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Grafana Workspace API Key (%s): %s", d.Id(), err)
	}

	return diags
}

const workspaceAPIKeyIDSeparator = "/"

func WorkspaceAPIKeyCreateResourceID(workspaceID, keyName string) string {
	parts := []string{workspaceID, keyName}
	id := strings.Join(parts, workspaceAPIKeyIDSeparator)

	return id
}

func WorkspaceAPIKeyParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, workspaceAPIKeyIDSeparator)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected workspace-id%[2]skey-name", id, workspaceAPIKeyIDSeparator)
	}

	return parts[0], parts[1], nil
}
