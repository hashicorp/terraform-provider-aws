package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsSsmParameter() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSsmParameterPut,
		Read:   resourceAwsSsmParameterRead,
		Update: resourceAwsSsmParameterPut,
		Delete: resourceAwsSsmParameterDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateSsmParameterType,
			},
			"value": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"key_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"overwrite": {
				Type:     schema.TypeBool,
				Optional: true,
				// The default should be set to true, terraform lifecycle should take care of not overriding the value if it is manually set by the user.
				// Otherwise, it is causing a breaking change because the first version did not allow overwrite parameter and overwrite was allowed.
				Default: true,
			},
			"allowed_pattern": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsSsmParameterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ssmconn

	log.Printf("[DEBUG] Reading SSM Parameter: %s", d.Id())

	if resp, err := conn.GetParameters(&ssm.GetParametersInput{
		Names:          []*string{aws.String(d.Get("name").(string))},
		WithDecryption: aws.Bool(true),
	}); err != nil {
		return errwrap.Wrapf("[ERROR] Error getting SSM parameter: {{err}}", err)
	} else {
		if len(resp.InvalidParameters) > 0 {
			log.Print("[INFO] The resource no longer exists, marking it for recreation:", d.Id())
			d.MarkNewResource()
			return nil
		}
		param := resp.Parameters[0]
		d.Set("name", param.Name)
		d.Set("type", param.Type)
		d.Set("value", param.Value)
	}

	if resp, err := conn.DescribeParameters(&ssm.DescribeParametersInput{
		Filters: []*ssm.ParametersFilter{
			&ssm.ParametersFilter{
				Key:    aws.String("Name"),
				Values: []*string{aws.String(d.Get("name").(string))},
			},
		},
	}); err != nil {
		return errwrap.Wrapf("[ERROR] Error describing SSM parameter: {{err}}", err)
	} else {
		param := resp.Parameters[0]
		d.Set("key_id", param.KeyId)
		d.Set("description", param.Description)
		d.Set("allowed_pattern", param.AllowedPattern)
	}

	if tagList, err := conn.ListTagsForResource(&ssm.ListTagsForResourceInput{
		ResourceId:   aws.String(d.Get("name").(string)),
		ResourceType: aws.String("Parameter"),
	}); err != nil {
		return fmt.Errorf("Failed to get SSM parameter tags for %s: %s", d.Get("name"), err)
	} else {
		d.Set("tags", tagsToMapSSM(tagList.TagList))
	}

	return nil
}

func resourceAwsSsmParameterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ssmconn

	log.Printf("[INFO] Deleting SSM Parameter: %s", d.Id())

	_, err := conn.DeleteParameter(&ssm.DeleteParameterInput{
		Name: aws.String(d.Get("name").(string)),
	})
	if err != nil {
		return err
	}

	d.SetId("")

	return nil
}

func resourceAwsSsmParameterPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ssmconn

	log.Printf("[INFO] Creating SSM Parameter: %s", d.Get("name").(string))

	paramInput := &ssm.PutParameterInput{
		Name:           aws.String(d.Get("name").(string)),
		Description:    aws.String(d.Get("description").(string)),
		Type:           aws.String(d.Get("type").(string)),
		Value:          aws.String(d.Get("value").(string)),
		Overwrite:      aws.Bool(d.Get("overwrite").(bool)),
		AllowedPattern: aws.String(d.Get("allowed_pattern").(string)),
	}

	if keyID, ok := d.GetOk("key_id"); ok {
		log.Printf("[DEBUG] Setting key_id for SSM Parameter %v: %s", d.Get("name"), keyID)
		paramInput.SetKeyId(keyID.(string))
	}

	log.Printf("[DEBUG] Waiting for SSM Parameter %v to be updated", d.Get("name"))
	if _, err := conn.PutParameter(paramInput); err != nil {
		return errwrap.Wrapf("[ERROR] Error creating SSM parameter: {{err}}", err)
	}

	if err := setTagsSSM(conn, d, d.Get("name").(string), "Parameter"); err != nil {
		return errwrap.Wrapf("[ERROR] Error creating SSM parameter tags: {{err}}", err)
	}

	d.SetId(d.Get("name").(string))

	return resourceAwsSsmParameterRead(d, meta)
}
