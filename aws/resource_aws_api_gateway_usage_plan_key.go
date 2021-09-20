package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceUsagePlanKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceUsagePlanKeyCreate,
		Read:   resourceUsagePlanKeyRead,
		Delete: resourceUsagePlanKeyDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected USAGE-PLAN-ID/USAGE-PLAN-KEY-ID", d.Id())
				}
				usagePlanId := idParts[0]
				usagePlanKeyId := idParts[1]
				d.Set("usage_plan_id", usagePlanId)
				d.Set("key_id", usagePlanKeyId)
				d.SetId(usagePlanKeyId)
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"key_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"key_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"usage_plan_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"value": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUsagePlanKeyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn
	log.Print("[DEBUG] Creating API Gateway Usage Plan Key")

	params := &apigateway.CreateUsagePlanKeyInput{
		KeyId:       aws.String(d.Get("key_id").(string)),
		KeyType:     aws.String(d.Get("key_type").(string)),
		UsagePlanId: aws.String(d.Get("usage_plan_id").(string)),
	}

	up, err := conn.CreateUsagePlanKey(params)

	if err != nil {
		return fmt.Errorf("error creating API Gateway Usage Plan Key: %w", err)
	}

	d.SetId(aws.StringValue(up.Id))

	return resourceUsagePlanKeyRead(d, meta)
}

func resourceUsagePlanKeyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn
	log.Printf("[DEBUG] Reading API Gateway Usage Plan Key: %s", d.Id())

	up, err := conn.GetUsagePlanKey(&apigateway.GetUsagePlanKeyInput{
		UsagePlanId: aws.String(d.Get("usage_plan_id").(string)),
		KeyId:       aws.String(d.Get("key_id").(string)),
	})
	if err != nil {
		if tfawserr.ErrMessageContains(err, apigateway.ErrCodeNotFoundException, "") {
			log.Printf("[WARN] API Gateway Usage Plan Key (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("name", up.Name)
	d.Set("value", up.Value)
	d.Set("key_type", up.Type)

	return nil
}

func resourceUsagePlanKeyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn

	log.Printf("[DEBUG] Deleting API Gateway Usage Plan Key: %s", d.Id())
	_, err := conn.DeleteUsagePlanKey(&apigateway.DeleteUsagePlanKeyInput{
		UsagePlanId: aws.String(d.Get("usage_plan_id").(string)),
		KeyId:       aws.String(d.Get("key_id").(string)),
	})
	if tfawserr.ErrMessageContains(err, apigateway.ErrCodeNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error deleting API Gateway usage plan key: %s", err)
	}

	return nil

}
