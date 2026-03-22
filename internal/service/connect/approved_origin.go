// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_connect_approved_origin", name="Approved Origin")
func resourceApprovedOrigin() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceApprovedOriginCreate,
		ReadWithoutTimeout:   resourceApprovedOriginRead,
		DeleteWithoutTimeout: resourceApprovedOriginDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"origin": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceApprovedOriginCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID := d.Get("instance_id").(string)
	origin := d.Get("origin").(string)

	input := &connect.AssociateApprovedOriginInput{
		InstanceId: aws.String(instanceID),
		Origin:     aws.String(origin),
	}

	_, err := conn.AssociateApprovedOrigin(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect Approved Origin (%s): %s", origin, err)
	}

	d.SetId(approvedOriginCreateResourceID(instanceID, origin))

	return append(diags, resourceApprovedOriginRead(ctx, d, meta)...)
}

func resourceApprovedOriginRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID, origin, err := ApprovedOriginParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	// List all approved origins and check if ours exists
	found := false
	input := &connect.ListApprovedOriginsInput{
		InstanceId: aws.String(instanceID),
	}

	paginator := connect.NewListApprovedOriginsPaginator(conn, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Connect Approved Origins for instance (%s): %s", instanceID, err)
		}
		for _, o := range page.Origins {
			if o == origin {
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		log.Printf("[WARN] Connect Approved Origin (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	d.Set("instance_id", instanceID)
	d.Set("origin", origin)

	return diags
}

func resourceApprovedOriginDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID, origin, err := ApprovedOriginParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Connect Approved Origin: %s", d.Id())

	_, err = conn.DisassociateApprovedOrigin(ctx, &connect.DisassociateApprovedOriginInput{
		InstanceId: aws.String(instanceID),
		Origin:     aws.String(origin),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Connect Approved Origin (%s): %s", d.Id(), err)
	}

	return diags
}

const approvedOriginResourceIDSeparator = ":"

func approvedOriginCreateResourceID(instanceID, origin string) string {
	parts := []string{instanceID, origin}
	id := strings.Join(parts, approvedOriginResourceIDSeparator)
	return id
}

func ApprovedOriginParseResourceID(id string) (string, string, error) {
	// origin can contain "://" so we split on first ":" only
	idx := strings.Index(id, approvedOriginResourceIDSeparator)
	if idx == -1 {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected instanceID:origin", id)
	}
	instanceID := id[:idx]
	origin := id[idx+1:]
	if instanceID == "" || origin == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected instanceID:origin", id)
	}
	return instanceID, origin, nil
}
