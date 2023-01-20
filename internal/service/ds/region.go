package ds

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceRegion() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRegionCreate,
		ReadWithoutTimeout:   resourceRegionRead,
		UpdateWithoutTimeout: resourceRegionUpdate,
		DeleteWithoutTimeout: resourceRegionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(180 * time.Minute),
			Update: schema.DefaultTimeout(90 * time.Minute),
			Delete: schema.DefaultTimeout(90 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"desired_number_of_domain_controllers": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(2),
			},
			"directory_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"region_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidRegionName,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"vpc_settings": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet_ids": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"vpc_id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceRegionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DSConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	directoryID := d.Get("directory_id").(string)
	regionName := d.Get("region_name").(string)
	id := RegionCreateResourceID(directoryID, regionName)
	input := &directoryservice.AddRegionInput{
		DirectoryId: aws.String(directoryID),
		RegionName:  aws.String(regionName),
	}

	if v, ok := d.GetOk("vpc_settings"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.VPCSettings = expandDirectoryVpcSettings(v.([]interface{})[0].(map[string]interface{}))
	}

	_, err := conn.AddRegionWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating Directory Service Region (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitRegionCreated(ctx, conn, directoryID, regionName, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("waiting for Directory Service Region (%s) create: %s", d.Id(), err)
	}

	if len(tags) > 0 {
		regionConn, err := regionalConn(meta.(*conns.AWSClient), regionName)

		if err != nil {
			return diag.FromErr(err)
		}

		if err := UpdateTags(ctx, regionConn, directoryID, nil, tags); err != nil {
			return diag.Errorf("adding Directory Service Directory (%s) tags: %s", directoryID, err)
		}

		if v, ok := d.GetOk("desired_number_of_domain_controllers"); ok {
			if err := updateNumberOfDomainControllers(ctx, regionConn, directoryID, v.(int), d.Timeout(schema.TimeoutCreate)); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return resourceRegionRead(ctx, d, meta)
}

func resourceRegionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DSConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	directoryID, regionName, err := RegionParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	region, err := FindRegion(ctx, conn, directoryID, regionName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Directory Service Region (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Directory Service Region (%s): %s", d.Id(), err)
	}

	d.Set("desired_number_of_domain_controllers", region.DesiredNumberOfDomainControllers)
	d.Set("directory_id", region.DirectoryId)
	d.Set("region_name", region.RegionName)
	if region.VpcSettings != nil {
		if err := d.Set("vpc_settings", []interface{}{flattenDirectoryVpcSettings(region.VpcSettings)}); err != nil {
			return diag.Errorf("setting vpc_settings: %s", err)
		}
	} else {
		d.Set("vpc_settings", nil)
	}

	regionConn, err := regionalConn(meta.(*conns.AWSClient), regionName)

	if err != nil {
		return diag.FromErr(err)
	}

	tags, err := ListTags(ctx, regionConn, directoryID)

	if err != nil {
		return diag.Errorf("listing tags for Directory Service Directory (%s): %s", directoryID, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return nil
}

func resourceRegionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	directoryID, regionName, err := RegionParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	conn, err := regionalConn(meta.(*conns.AWSClient), regionName)

	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("desired_number_of_domain_controllers") {
		if err := updateNumberOfDomainControllers(ctx, conn, directoryID, d.Get("desired_number_of_domain_controllers").(int), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, directoryID, o, n); err != nil {
			return diag.Errorf("updating Directory Service Directory (%s) tags: %s", directoryID, err)
		}
	}

	return resourceRegionRead(ctx, d, meta)
}

func resourceRegionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	directoryID, regionName, err := RegionParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	// The Region must be removed using a client in the region.
	conn, err := regionalConn(meta.(*conns.AWSClient), regionName)

	if err != nil {
		return diag.FromErr(err)
	}

	_, err = conn.RemoveRegionWithContext(ctx, &directoryservice.RemoveRegionInput{
		DirectoryId: aws.String(directoryID),
	})

	if tfawserr.ErrCodeEquals(err, directoryservice.ErrCodeDirectoryDoesNotExistException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Directory Service Region (%s): %s", d.Id(), err)
	}

	if _, err := waitRegionDeleted(ctx, conn, directoryID, regionName, d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("waiting for Directory Service Region (%s) delete: %s", d.Id(), err)
	}

	return nil
}

func regionalConn(client *conns.AWSClient, regionName string) (*directoryservice.DirectoryService, error) {
	sess, err := conns.NewSessionForRegion(&client.DSConn().Config, regionName, client.TerraformVersion)

	if err != nil {
		return nil, fmt.Errorf("creating AWS session (%s): %w", regionName, err)
	}

	return directoryservice.New(sess), nil
}

const regionIDSeparator = "," // nosemgrep:ci.ds-in-const-name,ci.ds-in-var-name

func RegionCreateResourceID(directoryID, regionName string) string {
	parts := []string{directoryID, regionName}
	id := strings.Join(parts, regionIDSeparator)

	return id
}

func RegionParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, regionIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected DirectoryID%[2]sRegionName", id, regionIDSeparator)
}
