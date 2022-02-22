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
			"instance_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
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
	conn := meta.(*conns.AWSClient).SSOAdminConn
	instanceArn := d.Get("instance_arn").(string)
	attributes, err := expandAccessControlAttributes(d)
	if err != nil {
		return err
	}
	input := ssoadmin.CreateInstanceAccessControlAttributeConfigurationInput{
		InstanceArn: aws.String(instanceArn),
		InstanceAccessControlAttributeConfiguration: &ssoadmin.InstanceAccessControlAttributeConfiguration{
			AccessControlAttributes: attributes,
		},
	}

	_, err = conn.CreateInstanceAccessControlAttributeConfiguration(&input)
	if err != nil {
		return fmt.Errorf("error putting access control attributes for SSO Instance Arn (%s): %w", instanceArn, err)
	}
	d.SetId(instanceArn)
	return resourceAccessControlAttributesRead(d, meta)
}

func resourceAccessControlAttributesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSOAdminConn
	instanceArn := d.Id()
	input := ssoadmin.DescribeInstanceAccessControlAttributeConfigurationInput{
		InstanceArn: aws.String(instanceArn),
	}
	resp, err := conn.DescribeInstanceAccessControlAttributeConfiguration(&input)
	if err != nil {
		return fmt.Errorf("error reading access control attributes for SSO Instance Arn (%s): %w", instanceArn, err)
	}
	d.Set("instance_arn", instanceArn)
	if err := d.Set("attribute", flattenAccessControlAttributes(resp.InstanceAccessControlAttributeConfiguration.AccessControlAttributes)); err != nil {
		return fmt.Errorf("error setting attribute: %w", err)
	}
	d.Set("status", resp.Status)
	d.Set("status_reason", resp.StatusReason)
	return nil
}
func resourceAccessControlAttributesUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSOAdminConn
	instanceArn := d.Id()
	if d.HasChanges("attribute") {
		attributes, err := expandAccessControlAttributes(d)
		if err != nil {
			return fmt.Errorf("error updating access control attributes for SSO Instance Arn (%s): %w", instanceArn, err)
		}
		input := ssoadmin.UpdateInstanceAccessControlAttributeConfigurationInput{
			InstanceArn: aws.String(instanceArn),
			InstanceAccessControlAttributeConfiguration: &ssoadmin.InstanceAccessControlAttributeConfiguration{
				AccessControlAttributes: attributes,
			},
		}
		_, err = conn.UpdateInstanceAccessControlAttributeConfiguration(&input)
		if err != nil {
			return fmt.Errorf("error updating access control attributes for SSO Instance Arn (%s): %w", instanceArn, err)
		}
	}
	return resourceAccessControlAttributesRead(d, meta)
}
func resourceAccessControlAttributesDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSOAdminConn
	instanceArn := d.Id()
	input := ssoadmin.DeleteInstanceAccessControlAttributeConfigurationInput{
		InstanceArn: aws.String(instanceArn),
	}
	_, err := conn.DeleteInstanceAccessControlAttributeConfiguration(&input)
	if err != nil {
		return fmt.Errorf("error deleting access control attributes for SSO Instance Arn (%s): %w", instanceArn, err)
	}
	return nil
}

func expandAccessControlAttributes(d *schema.ResourceData) (attributes []*ssoadmin.AccessControlAttribute, err error) {
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
				Source: flex.ExpandStringList(v.List()),
			}
		}
		attributes = append(attributes, &attribute)
	}
	return attributes, nil
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
