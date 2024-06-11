// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ds

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directoryservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directoryservice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResNameSharedDirectoryAccepter = "Shared Directory Accepter"
)

// @SDKResource("aws_directory_service_shared_directory_accepter")
func ResourceSharedDirectoryAccepter() *schema.Resource {
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

	input := directoryservice.AcceptSharedDirectoryInput{
		SharedDirectoryId: aws.String(d.Get("shared_directory_id").(string)),
	}

	log.Printf("[DEBUG] Accepting shared directory: %+v", input)

	output, err := conn.AcceptSharedDirectory(ctx, &input)

	if err != nil {
		return create.AppendDiagError(diags, names.DS, create.ErrActionCreating, ResNameSharedDirectoryAccepter, d.Get("shared_directory_id").(string), err)
	}

	if output == nil || output.SharedDirectory == nil {
		return create.AppendDiagError(diags, names.DS, create.ErrActionCreating, ResNameSharedDirectoryAccepter, d.Get("shared_directory_id").(string), errors.New("empty output"))
	}

	d.SetId(d.Get("shared_directory_id").(string))

	d.Set("notes", output.SharedDirectory.ShareNotes) // only available in response to create

	_, err = waitDirectoryShared(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return create.AppendDiagError(diags, names.DS, create.ErrActionWaitingForCreation, ResNameSharedDirectoryAccepter, d.Id(), err)
	}

	return append(diags, resourceSharedDirectoryAccepterRead(ctx, d, meta)...)
}

func resourceSharedDirectoryAccepterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).DSClient(ctx)

	dir, err := FindDirectoryByID(ctx, conn, d.Id())

	if err != nil {
		return create.AppendDiagError(diags, names.DS, create.ErrActionReading, ResNameSharedDirectoryAccepter, d.Id(), err)
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

	log.Printf("[DEBUG] Deleting Directory Service Directory: %s", d.Id())
	_, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.ClientException](ctx, directoryApplicationDeauthorizedPropagationTimeout, func() (interface{}, error) {
		return conn.DeleteDirectory(ctx, &directoryservice.DeleteDirectoryInput{
			DirectoryId: aws.String(d.Id()),
		})
	}, "authorized applications")

	if errs.IsA[*awstypes.EntityDoesNotExistException](err) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.DS, create.ErrActionDeleting, ResNameSharedDirectoryAccepter, d.Id(), err)
	}

	if _, err := waitDirectoryDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.AppendDiagError(diags, names.DS, create.ErrActionWaitingForDeletion, ResNameSharedDirectoryAccepter, d.Id(), err)
	}

	return diags
}
