package ds

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDirectoryShareAccepter() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDirectoryShareAccepterCreate,
		ReadContext:   resourceDirectoryShareAccepterRead,
		UpdateContext: resourceDirectoryShareAccepterUpdate,
		DeleteContext: resourceDirectoryShareAccepterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"shared_directory_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"alias": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"size": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"short_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"dns_ip_addresses": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Computed: true,
			},
			"sso_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"access_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_subnet_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"owner_availability_zones": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"owner_directory_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"share_notes": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDirectoryShareAccepterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	dirId := d.Get("shared_directory_id").(string)

	input := directoryservice.AcceptSharedDirectoryInput{
		SharedDirectoryId: aws.String(dirId),
	}

	log.Printf("[DEBUG] Accepting shared directory: %s", input)
	output, err := conn.AcceptSharedDirectoryWithContext(ctx, &input)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[DEBUG] Accepted shared directory: %s", output)

	d.SetId(dirId)

	_, err = waitDirectoryCreated(conn, dirId)

	if err != nil {
		return diag.Errorf("error waiting for Directory Service Directory (%s) to accept share: %s", dirId, err.Error())
	}

	_, err = conn.AddTagsToResourceWithContext(ctx, &directoryservice.AddTagsToResourceInput{
		ResourceId: aws.String(dirId),
		Tags:       Tags(tags),
	})

	if err != nil {
		return diag.Errorf("error tagging directory (%s): %s", dirId, err.Error())
	}

	return resourceDirectoryShareAccepterRead(ctx, d, meta)
}

func resourceDirectoryShareAccepterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	dir, err := findDirectoryByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		d.SetId("")
		return []diag.Diagnostic{{
			Severity: diag.Warning,
			Summary:  fmt.Sprintf("Directory Service Directory Share (%s) not found, removing from state", d.Id()),
		}}
	}

	if err != nil {
		return diag.Errorf("error reading Directory Service Directory (%s): %s", d.Id(), err.Error())
	}

	log.Printf("[DEBUG] Received DS directory: %s", dir)

	d.Set("access_url", dir.AccessUrl)
	d.Set("alias", dir.Alias)
	d.Set("description", dir.Description)
	d.Set("name", dir.Name)
	d.Set("share_notes", dir.ShareNotes)
	d.Set("short_name", dir.ShortName)
	d.Set("sso_enabled", dir.SsoEnabled)
	d.Set("type", dir.Type)
	d.Set("size", dir.Size)

	ownerDesc := dir.OwnerDirectoryDescription
	if ownerDesc != nil {
		return diag.Errorf("No owner directory description found for Directory Service Directory (%s)", d.Id())
	}
	d.Set("owner_directory_id", ownerDesc.DirectoryId)
	d.Set("owner_account_id", ownerDesc.AccountId)
	d.Set("dns_ip_addresses", flex.FlattenStringSet(ownerDesc.DnsIpAddrs))
	if ownerDesc.VpcSettings != nil {
		d.Set("owner_vpc_id", ownerDesc.VpcSettings.VpcId)
		d.Set("owner_subnet_ids", flex.FlattenStringSet(ownerDesc.VpcSettings.SubnetIds))
		d.Set("owner_availability_zones", flex.FlattenStringSet(ownerDesc.VpcSettings.AvailabilityZones))
	}

	tags, err := ListTags(conn, d.Id())

	if err != nil {
		return diag.Errorf("error listing tags for Directory Service Directory (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("error setting tags: %s", err.Error())
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("error setting tags_all: %s", err.Error())
	}

	return nil
}

func resourceDirectoryShareAccepterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DSConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return diag.Errorf("error update shared Directory Service Directory (%s) tags: %s", d.Id(), err.Error())
		}
	}

	return resourceDirectoryShareAccepterRead(ctx, d, meta)
}

func resourceDirectoryShareAccepterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DSConn

	input := directoryservice.RejectSharedDirectoryInput{
		SharedDirectoryId: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Rejecting Directory Service Directory Share: %s", input)
	output, err := conn.RejectSharedDirectoryWithContext(ctx, &input)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[DEBUG] Rejected Directory Service Share: %s", output)

	return nil
}
