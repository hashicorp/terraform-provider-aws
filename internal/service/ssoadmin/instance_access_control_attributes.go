package ssoadmin

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceAccessControlAttributes() *schema.Resource {
	return &schema.Resource{
		Create: resourceAccessControlAttributesCreate,
		Read:   resourceAccessControlAttributesRead,
		Update: resourceAccessControlAttributesUpdate,
		Delete: resourceAccessControlAttributesDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"attribute": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"source": {
										Type:     schema.TypeSet,
										Required: true,
										MinItems: 1,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
					},
				},
			},
			"instance_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status_reason": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAccessControlAttributesCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSOAdminConn()

	instanceARN := d.Get("instance_arn").(string)
	input := &ssoadmin.CreateInstanceAccessControlAttributeConfigurationInput{
		InstanceArn: aws.String(instanceARN),
		InstanceAccessControlAttributeConfiguration: &ssoadmin.InstanceAccessControlAttributeConfiguration{
			AccessControlAttributes: expandAccessControlAttributes(d),
		},
	}

	_, err := conn.CreateInstanceAccessControlAttributeConfiguration(input)

	if err != nil {
		return fmt.Errorf("creating SSO Instance Access Control Attributes (%s): %w", instanceARN, err)
	}

	d.SetId(instanceARN)

	return resourceAccessControlAttributesRead(d, meta)
}

func resourceAccessControlAttributesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSOAdminConn()

	input := &ssoadmin.DescribeInstanceAccessControlAttributeConfigurationInput{
		InstanceArn: aws.String(d.Id()),
	}
	resp, err := conn.DescribeInstanceAccessControlAttributeConfiguration(input)

	if err != nil {
		return fmt.Errorf("reading SSO Instance Access Control Attributes (%s): %w", d.Id(), err)
	}

	d.Set("instance_arn", d.Id())
	if err := d.Set("attribute", flattenAccessControlAttributes(resp.InstanceAccessControlAttributeConfiguration.AccessControlAttributes)); err != nil {
		return fmt.Errorf("setting attribute: %w", err)
	}
	d.Set("status", resp.Status)
	d.Set("status_reason", resp.StatusReason)

	return nil
}
func resourceAccessControlAttributesUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSOAdminConn()

	input := &ssoadmin.UpdateInstanceAccessControlAttributeConfigurationInput{
		InstanceArn: aws.String(d.Id()),
		InstanceAccessControlAttributeConfiguration: &ssoadmin.InstanceAccessControlAttributeConfiguration{
			AccessControlAttributes: expandAccessControlAttributes(d),
		},
	}

	_, err := conn.UpdateInstanceAccessControlAttributeConfiguration(input)

	if err != nil {
		return fmt.Errorf("updating SSO Instance Access Control Attributes (%s): %w", d.Id(), err)
	}

	return resourceAccessControlAttributesRead(d, meta)
}
func resourceAccessControlAttributesDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSOAdminConn()

	_, err := conn.DeleteInstanceAccessControlAttributeConfiguration(&ssoadmin.DeleteInstanceAccessControlAttributeConfigurationInput{
		InstanceArn: aws.String(d.Id()),
	})

	if err != nil {
		return fmt.Errorf("deleting SSO Instance Access Control Attributes (%s): %w", d.Id(), err)
	}

	return nil
}

func expandAccessControlAttributes(d *schema.ResourceData) (attributes []*ssoadmin.AccessControlAttribute) {
	attInterface := d.Get("attribute").(*schema.Set).List()
	for _, attrMap := range attInterface {
		attr := attrMap.(map[string]interface{})
		var attribute ssoadmin.AccessControlAttribute
		if key, ok := attr["key"].(string); ok {
			attribute.Key = aws.String(key)
		}
		val := attr["value"].(*schema.Set).List()[0].(map[string]interface{})
		if v, ok := val["source"].(*schema.Set); ok && len(v.List()) > 0 {
			attribute.Value = &ssoadmin.AccessControlAttributeValue{
				Source: flex.ExpandStringSet(v),
			}
		}
		attributes = append(attributes, &attribute)
	}
	return
}

func flattenAccessControlAttributes(attributes []*ssoadmin.AccessControlAttribute) []interface{} {
	var results []interface{}
	if len(attributes) == 0 {
		return []interface{}{}
	}
	for _, attr := range attributes {
		if attr == nil {
			continue
		}
		var val []interface{}
		val = append(val, map[string]interface{}{
			"source": flex.FlattenStringSet(attr.Value.Source),
		})
		results = append(results, map[string]interface{}{
			"key":   aws.StringValue(attr.Key),
			"value": val,
		})
	}
	return results
}
