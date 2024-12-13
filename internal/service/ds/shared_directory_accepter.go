// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ds

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directoryservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directoryservice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_directory_service_shared_directory_accepter", name="Shared Directory Accepter")
func resourceSharedDirectoryAccepter() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSharedDirectoryAccepterCreate,
		ReadWithoutTimeout:   resourceSharedDirectoryAccepterRead,
		DeleteWithoutTimeout: resourceSharedDirectoryAccepterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"method": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"notes": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrOwnerAccountID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_directory_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"shared_directory_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceSharedDirectoryAccepterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DSClient(ctx)

	sharedDirectoryID := d.Get("shared_directory_id").(string)
	input := &directoryservice.AcceptSharedDirectoryInput{
		SharedDirectoryId: aws.String(sharedDirectoryID),
	}

	output, err := conn.AcceptSharedDirectory(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "accepting Directory Service Shared Directory (%s): %s", sharedDirectoryID, err)
	}

	d.SetId(sharedDirectoryID)
	d.Set("notes", output.SharedDirectory.ShareNotes) // only available in response to create

	if _, err := waitSharedDirectoryAccepted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Directory Service Shared Directory (%s) accept: %s", d.Id(), err)
	}

	return append(diags, resourceSharedDirectoryAccepterRead(ctx, d, meta)...)
}

func resourceSharedDirectoryAccepterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DSClient(ctx)

	dir, err := findSharedDirectoryAccepterByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Directory Service Shared Directory Accepter (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Directory Service Shared Directory Accepter (%s): %s", d.Id(), err)
	}

	d.Set("method", dir.ShareMethod)
	d.Set(names.AttrOwnerAccountID, dir.OwnerDirectoryDescription.AccountId)
	d.Set("owner_directory_id", dir.OwnerDirectoryDescription.DirectoryId)
	d.Set("shared_directory_id", dir.DirectoryId)

	return diags
}

func resourceSharedDirectoryAccepterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DSClient(ctx)

	log.Printf("[DEBUG] Deleting Directory Service Shared Directory Accepter: %s", d.Id())
	_, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.ClientException](ctx, directoryApplicationDeauthorizedPropagationTimeout, func() (interface{}, error) {
		return conn.DeleteDirectory(ctx, &directoryservice.DeleteDirectoryInput{
			DirectoryId: aws.String(d.Id()),
		})
	}, "authorized applications")

	if errs.IsA[*awstypes.EntityDoesNotExistException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Directory Service Shared Directory Accepter (%s): %s", d.Id(), err)
	}

	if _, err := waitDirectoryDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Directory Service Shared Directory Accepter (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findSharedDirectoryAccepterByID(ctx context.Context, conn *directoryservice.Client, id string) (*awstypes.DirectoryDescription, error) { // nosemgrep:ci.ds-in-func-name
	output, err := findDirectoryByID(ctx, conn, id)

	if err != nil {
		return nil, err
	}

	if output.OwnerDirectoryDescription == nil {
		return nil, tfresource.NewEmptyResultError(id)
	}

	return output, nil
}

func statusDirectoryShareStatus(ctx context.Context, conn *directoryservice.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findDirectoryByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ShareStatus), nil
	}
}

func waitSharedDirectoryAccepted(ctx context.Context, conn *directoryservice.Client, id string, timeout time.Duration) (*awstypes.SharedDirectory, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ShareStatusPendingAcceptance, awstypes.ShareStatusSharing),
		Target:                    enum.Slice(awstypes.ShareStatusShared),
		Refresh:                   statusDirectoryShareStatus(ctx, conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.SharedDirectory); ok {
		return output, err
	}

	return nil, err
}
