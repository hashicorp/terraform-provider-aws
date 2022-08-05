package ds

import (
	"context"
	"errors"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResNameSharedDirectoryAccepter = "Shared Directory Accepter"
)

func ResourceSharedDirectoryAccepter() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSharedDirectoryAccepterCreate,
		ReadContext:   resourceSharedDirectoryAccepterRead,
		DeleteContext: resourceSharedDirectoryAccepterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
			"owner_account_id": {
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
	conn := meta.(*conns.AWSClient).DSConn

	input := directoryservice.AcceptSharedDirectoryInput{
		SharedDirectoryId: aws.String(d.Get("shared_directory_id").(string)),
	}

	log.Printf("[DEBUG] Accepting shared directory: %s", input)

	output, err := conn.AcceptSharedDirectoryWithContext(ctx, &input)

	if err != nil {
		return create.DiagError(names.DS, create.ErrActionCreating, ResNameSharedDirectoryAccepter, d.Get("shared_directory_id").(string), err)
	}

	if output == nil || output.SharedDirectory == nil {
		return create.DiagError(names.DS, create.ErrActionCreating, ResNameSharedDirectoryAccepter, d.Get("shared_directory_id").(string), errors.New("empty output"))
	}

	d.SetId(d.Get("shared_directory_id").(string))

	d.Set("notes", output.SharedDirectory.ShareNotes) // only available in response to create

	_, err = waitDirectoryShared(ctx, conn, d.Id())

	if err != nil {
		return create.DiagError(names.DS, create.ErrActionWaitingForCreation, ResNameSharedDirectoryAccepter, d.Id(), err)
	}

	return resourceSharedDirectoryAccepterRead(ctx, d, meta)
}

func resourceSharedDirectoryAccepterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DSConn

	dir, err := findDirectoryByID(conn, d.Id())

	if err != nil {
		return create.DiagError(names.DS, create.ErrActionReading, ResNameSharedDirectoryAccepter, d.Id(), err)
	}

	d.Set("method", dir.ShareMethod)
	d.Set("owner_account_id", dir.OwnerDirectoryDescription.AccountId)
	d.Set("owner_directory_id", dir.OwnerDirectoryDescription.DirectoryId)
	d.Set("shared_directory_id", dir.DirectoryId)
	return nil
}

func resourceSharedDirectoryAccepterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DSConn

	input := &directoryservice.DeleteDirectoryInput{
		DirectoryId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Directory Service Directory: (%s)", d.Id())
	err := resource.Retry(directoryApplicationDeauthorizedPropagationTimeout, func() *resource.RetryError {
		_, err := conn.DeleteDirectory(input)

		if tfawserr.ErrCodeEquals(err, directoryservice.ErrCodeEntityDoesNotExistException) {
			return nil
		}
		if tfawserr.ErrMessageContains(err, directoryservice.ErrCodeClientException, "authorized applications") {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DeleteDirectory(input)
	}

	if err != nil {
		return create.DiagError(names.DS, create.ErrActionDeleting, ResNameSharedDirectoryAccepter, d.Id(), err)
	}

	err = waitDirectoryDeleted(conn, d.Id())

	if err != nil {
		return create.DiagError(names.DS, create.ErrActionWaitingForDeletion, ResNameSharedDirectoryAccepter, d.Id(), err)
	}

	return nil
}
