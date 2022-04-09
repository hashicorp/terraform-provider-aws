package ds

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceDirectoryShare() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDirectoryShareCreate,
		ReadContext:   resourceDirectoryShareRead,
		DeleteContext: resourceDirectoryShareDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDirectoryShareImport,
		},

		Schema: map[string]*schema.Schema{
			"directory_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"share_method": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      directoryservice.ShareMethodHandshake,
				ValidateFunc: validation.StringInSlice(directoryservice.ShareMethod_Values(), false),
			},
			"share_notes": {
				Type:      schema.TypeString,
				Optional:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			"target": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"type": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      directoryservice.TargetTypeAccount,
							ValidateFunc: validation.StringInSlice(directoryservice.TargetType_Values(), false),
						},
					},
				},
			},
			"shared_directory_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDirectoryShareCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DSConn

	dirId := d.Get("directory_id").(string)
	input := directoryservice.ShareDirectoryInput{
		DirectoryId: aws.String(dirId),
		ShareMethod: aws.String(d.Get("share_method").(string)),
		ShareTarget: &directoryservice.ShareTarget{
			Id:   aws.String(d.Get("target.id").(string)),
			Type: aws.String(d.Get("target.type").(string)),
		},
	}
	if v, ok := d.GetOk("share_notes"); ok {
		input.ShareNotes = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Shared Directory: %s", input)
	out, err := conn.ShareDirectoryWithContext(ctx, &input)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[DEBUG] Shared Directory created: %s", out)
	d.SetId(fmt.Sprintf("%s/%s", dirId, aws.StringValue(out.SharedDirectoryId)))

	return nil
}

func resourceDirectoryShareRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DSConn

	dirId := d.Get("directory_id").(string)
	sharedId := d.Get("shared_directory_id").(string)

	dir, err := findDirectoryShareByIDs(ctx, conn, dirId, sharedId)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Directory Service Shared Directory (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Received DS shared directory: %s", dir)

	d.Set("method", aws.StringValue(dir.ShareMethod))
	d.Set("notes", aws.StringValue(dir.ShareNotes))
	d.Set("target.id", aws.StringValue(dir.SharedDirectoryId))

	return nil
}

func resourceDirectoryShareDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DSConn

	dirId := d.Get("directory_id").(string)
	sharedId := d.Get("shared_directory_id").(string)

	input := directoryservice.UnshareDirectoryInput{
		DirectoryId: aws.String(dirId),
		UnshareTarget: &directoryservice.UnshareTarget{
			Id:   aws.String(d.Get("target.id").(string)),
			Type: aws.String(d.Get("target.type").(string)),
		},
	}

	log.Printf("[DEBUG] Unsharing Directory Service Directory: %s", input)
	output, err := conn.UnshareDirectoryWithContext(ctx, &input)
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = waitDirectoryShareDeleted(ctx, conn, dirId, sharedId)

	if err != nil {
		return diag.Errorf("error waiting for Directory Service Share (%s) to delete: %s", d.Id(), err.Error())
	}
	log.Printf("[DEBUG] Unshared Directory Service Directory: %s", output)

	return nil
}

func resourceDirectoryShareImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.SplitN(d.Id(), "/", 2)
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("unexpected format of ID (%q), expected <owner-directory-id>/<shared-directory-id>", d.Id())
	}

	ownerDirId := idParts[0]
	sharedDirId := idParts[1]

	d.Set("directory_id", ownerDirId)
	d.Set("shared_directory_id", sharedDirId)
	return []*schema.ResourceData{d}, nil
}
