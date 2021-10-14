package autoscaling

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceGroupTag() *schema.Resource {
	return &schema.Resource{
		Create: resourceGroupTagCreate,
		Read:   resourceGroupTagRead,
		Update: resourceGroupTagUpdate,
		Delete: resourceGroupTagDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"autoscaling_group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tag": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
						"propagate_at_launch": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceGroupTagCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AutoScalingConn

	identifier := d.Get("autoscaling_group_name").(string)
	tags := d.Get("tag").([]interface{})
	key := tags[0].(map[string]interface{})["key"].(string)

	if err := tftags.AutoscalingUpdateTags(conn, identifier, TagResourceTypeGroup, nil, tags); err != nil {
		return fmt.Errorf("error creating AutoScaling Group (%s) tag (%s): %w", identifier, key, err)
	}

	d.SetId(tftags.SetResourceID(identifier, key))

	return resourceGroupTagRead(d, meta)
}

func resourceGroupTagRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AutoScalingConn
	identifier, key, err := tftags.GetResourceID(d.Id())

	if err != nil {
		return err
	}

	value, err := tftags.AutoscalingGetTag(conn, identifier, TagResourceTypeGroup, key)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AutoScaling Group (%s) tag (%s), removing from state", identifier, key)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading AutoScaling Group (%s) tag (%s): %w", identifier, key, err)
	}

	d.Set("autoscaling_group_name", identifier)

	if err := d.Set("tag", []map[string]interface{}{{
		"key":                 key,
		"value":               value.Value,
		"propagate_at_launch": value.AdditionalBoolFields["PropagateAtLaunch"],
	}}); err != nil {
		return fmt.Errorf("error setting tag: %w", err)
	}

	return nil
}

func resourceGroupTagUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AutoScalingConn
	identifier, key, err := tftags.GetResourceID(d.Id())

	if err != nil {
		return err
	}

	if err := tftags.AutoscalingUpdateTags(conn, identifier, TagResourceTypeGroup, nil, d.Get("tag")); err != nil {
		return fmt.Errorf("error updating AutoScaling Group (%s) tag (%s): %w", identifier, key, err)
	}

	return resourceGroupTagRead(d, meta)
}

func resourceGroupTagDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AutoScalingConn
	identifier, key, err := tftags.GetResourceID(d.Id())

	if err != nil {
		return err
	}

	if err := tftags.AutoscalingUpdateTags(conn, identifier, TagResourceTypeGroup, d.Get("tag"), nil); err != nil {
		return fmt.Errorf("error deleting AutoScaling Group (%s) tag (%s): %w", identifier, key, err)
	}

	return nil
}
