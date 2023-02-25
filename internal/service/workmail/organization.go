package workmail

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/workmail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceOrganization() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceOrganizationCreate,
		ReadWithoutTimeout:   resourceOrganizationRead,
		DeleteWithoutTimeout: resourceOrganizationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Read:   schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"alias": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_mail_domain": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"directory_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"directory_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

const (
	ResNameOrganization = "Organization"
)

func resourceOrganizationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).WorkMailClient()

	in := &workmail.CreateOrganizationInput{
		Alias:       aws.String(d.Get("alias").(string)),
		ClientToken: aws.String(resource.UniqueId()),
	}

	out, err := conn.CreateOrganization(ctx, in)
	if err != nil {
		return create.DiagError(names.WorkMail, create.ErrActionCreating, ResNameOrganization, d.Get("alias").(string), err)
	}

	if out == nil || out.OrganizationId == nil {
		return create.DiagError(names.WorkMail, create.ErrActionCreating, ResNameOrganization, d.Get("alias").(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.OrganizationId))

	if _, err := waitOrganizationCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.DiagError(names.WorkMail, create.ErrActionWaitingForCreation, ResNameOrganization, d.Id(), err)
	}

	return resourceOrganizationRead(ctx, d, meta)
}

func resourceOrganizationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).WorkMailClient()

	out, err := findOrganizationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WorkMail Organization (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.WorkMail, create.ErrActionReading, ResNameOrganization, d.Id(), err)
	}

	if !d.IsNewResource() && aws.ToString(out.State) == statusDeleted {
		log.Printf("[WARN] WorkMail Organization (%s) is already deleted, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("alias", out.Alias)
	d.Set("arn", out.ARN)
	d.Set("default_mail_domain", out.DefaultMailDomain)
	d.Set("directory_id", out.DirectoryId)
	d.Set("directory_type", out.DirectoryType)
	d.Set("state", out.State)

	return nil
}

func resourceOrganizationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).WorkMailClient()

	out, err := findOrganizationByID(ctx, conn, d.Id())

	if tfresource.NotFound(err) {
		log.Printf("[WARN] WorkMail Organization (%s) not found", d.Id())
		return nil
	}

	if err != nil {
		return create.DiagError(names.WorkMail, create.ErrActionDeleting, ResNameOrganization, d.Id(), err)
	}

	if aws.ToString(out.State) == statusDeleted {
		log.Printf("[WARN] WorkMail Organization (%s) is already deleted", d.Id())
		return nil
	}

	log.Printf("[INFO] Deleting WorkMail Organization %s", d.Id())

	_, err = conn.DeleteOrganization(ctx, &workmail.DeleteOrganizationInput{
		OrganizationId:  aws.String(d.Id()),
		DeleteDirectory: true,
		ClientToken:     aws.String(resource.UniqueId()),
	})

	if err != nil {
		return create.DiagError(names.WorkMail, create.ErrActionDeleting, ResNameOrganization, d.Id(), err)
	}

	if _, err := waitOrganizationDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.DiagError(names.WorkMail, create.ErrActionWaitingForDeletion, ResNameOrganization, d.Id(), err)
	}

	return nil
}
