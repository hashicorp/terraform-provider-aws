package ds

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceRegion() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRegionCreate,
		ReadWithoutTimeout:   resourceRegionRead,
		DeleteWithoutTimeout: resourceRegionDelete,

		Schema: map[string]*schema.Schema{
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
	}
}

func resourceRegionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DSConn

	directoryID := d.Get("directory_id").(string)
	regionName := d.Get("region_name").(string)
	id := RegionCreateResourceID(directoryID, regionName)
	input := &directoryservice.AddRegionInput{
		DirectoryId: aws.String(directoryID),
		RegionName:  aws.String(regionName),
		VPCSettings: expandDirectoryVpcSettings(d.Get("vpc_settings").([]interface{})[0].(map[string]interface{})),
	}

	_, err := conn.AddRegionWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating Directory Service Region (%s): %s", id, err)
	}

	d.SetId(id)

	// TODO Waiter.

	return resourceRegionRead(ctx, d, meta)
}

func resourceRegionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DSConn

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

	d.Set("directory_id", region.DirectoryId)
	d.Set("region_name", region.RegionName)

	return nil
}

func resourceRegionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DSConn

	directoryID, _, err := RegionParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	_, err = conn.RemoveRegionWithContext(ctx, &directoryservice.RemoveRegionInput{
		DirectoryId: aws.String(directoryID),
	})

	if tfawserr.ErrCodeEquals(err, directoryservice.ErrCodeEntityDoesNotExistException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Directory Service Region (%s): %s", d.Id(), err)
	}

	// TODO Waiter.

	return nil
}

const regionIDSeparator = ","

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
