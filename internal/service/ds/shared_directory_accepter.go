package ds

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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
	conn := meta.(*conns.AWSClient).DSConn()

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

	_, err = waitDirectoryShared(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return create.DiagError(names.DS, create.ErrActionWaitingForCreation, ResNameSharedDirectoryAccepter, d.Id(), err)
	}

	return resourceSharedDirectoryAccepterRead(ctx, d, meta)
}

func resourceSharedDirectoryAccepterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DSConn()

	dir, err := FindDirectoryByID(ctx, conn, d.Id())

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
	conn := meta.(*conns.AWSClient).DSConn()

	log.Printf("[DEBUG] Deleting Directory Service Directory: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, directoryApplicationDeauthorizedPropagationTimeout, func() (interface{}, error) {
		return conn.DeleteDirectoryWithContext(ctx, &directoryservice.DeleteDirectoryInput{
			DirectoryId: aws.String(d.Id()),
		})
	}, directoryservice.ErrCodeClientException, "authorized applications")

	if tfawserr.ErrCodeEquals(err, directoryservice.ErrCodeEntityDoesNotExistException) {
		return nil
	}

	if err != nil {
		return create.DiagError(names.DS, create.ErrActionDeleting, ResNameSharedDirectoryAccepter, d.Id(), err)
	}

	if _, err := waitDirectoryDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.DiagError(names.DS, create.ErrActionWaitingForDeletion, ResNameSharedDirectoryAccepter, d.Id(), err)
	}

	return nil
}
