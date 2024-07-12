// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package grafana

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/grafana"
	awstypes "github.com/aws/aws-sdk-go-v2/service/grafana/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_grafana_workspace_api_key", name="Workspace API Key")
func resourceWorkspaceAPIKey() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceWorkspaceAPIKeyCreate,
		ReadWithoutTimeout:   schema.NoopContext,
		UpdateWithoutTimeout: schema.NoopContext,
		DeleteWithoutTimeout: resourceWorkspaceAPIKeyDelete,

		Schema: map[string]*schema.Schema{
			names.AttrKey: {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"key_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"key_role": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.Role](),
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
	conn := meta.(*conns.AWSClient).GrafanaClient(ctx)

	keyName := d.Get("key_name").(string)
	workspaceID := d.Get("workspace_id").(string)
	id := workspaceAPIKeyCreateResourceID(workspaceID, keyName)
	input := &grafana.CreateWorkspaceApiKeyInput{
		KeyName:       aws.String(keyName),
		KeyRole:       aws.String(d.Get("key_role").(string)),
		SecondsToLive: aws.Int32(int32(d.Get("seconds_to_live").(int))),
		WorkspaceId:   aws.String(workspaceID),
	}

	output, err := conn.CreateWorkspaceApiKey(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Grafana Workspace API Key (%s): %s", id, err)
	}

	d.SetId(id)
	d.Set(names.AttrKey, output.Key)

	return diags
}

func resourceWorkspaceAPIKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GrafanaClient(ctx)

	workspaceID, keyName, err := workspaceAPIKeyParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Grafana Workspace API Key: %s", d.Id())
	_, err = conn.DeleteWorkspaceApiKey(ctx, &grafana.DeleteWorkspaceApiKeyInput{
		KeyName:     aws.String(keyName),
		WorkspaceId: aws.String(workspaceID),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Grafana Workspace API Key (%s): %s", d.Id(), err)
	}

	return diags
}

const workspaceAPIKeyResourceIDSeparator = "/"

func workspaceAPIKeyCreateResourceID(workspaceID, keyName string) string {
	parts := []string{workspaceID, keyName}
	id := strings.Join(parts, workspaceAPIKeyResourceIDSeparator)

	return id
}

func workspaceAPIKeyParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, workspaceAPIKeyResourceIDSeparator)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected workspace-id%[2]skey-name", id, workspaceAPIKeyResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}
