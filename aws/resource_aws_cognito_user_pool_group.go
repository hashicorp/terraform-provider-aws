package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsCognitoUserPoolGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCognitoUserPoolGroupCreate,
		Read:   resourceAwsCognitoUserPoolGroupRead,
		Update: resourceAwsCognitoUserPoolGroupUpdate,
		Delete: resourceAwsCognitoUserPoolGroupDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		// https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_CreateGroup.html
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateMaxLength(128),
			},

			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateMaxLength(2048),
			},

			"precedence": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validateIntegerInRange(0, 1000),
			},

			"role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateArn,
			},

			"user_pool_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateMaxLength(55),
			},
		},
	}
}

func resourceAwsCognitoUserPoolGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	params := &cognitoidentityprovider.CreateGroupInput{
		GroupName:  aws.String(d.Get("name").(string)),
		UserPoolId: aws.String(d.Get("user_pool_id").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		params.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("precedence"); ok {
		params.Precedence = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("role_arn"); ok {
		params.RoleArn = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Cognito User Pool Group: %s", params)

	resp, err := conn.CreateGroup(params)

	if err != nil {
		return errwrap.Wrapf("Error creating Cognito User Pool Group: {{err}}", err)
	}

	d.SetId(fmt.Sprintf("%s/%s", *resp.Group.UserPoolId, *resp.Group.GroupName))

	return resourceAwsCognitoUserPoolGroupRead(d, meta)
}

func resourceAwsCognitoUserPoolGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	params := &cognitoidentityprovider.GetGroupInput{
		GroupName:  aws.String(d.Get("name").(string)),
		UserPoolId: aws.String(d.Get("user_pool_id").(string)),
	}

	log.Printf("[DEBUG] Reading Cognito User Pool Group: %s", params)

	resp, err := conn.GetGroup(params)

	if err != nil {
		if isAWSErr(err, "ResourceNotFoundException", "") {
			log.Printf("[WARN] Cognito User Pool Group %s is already gone", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	d.SetId(fmt.Sprintf("%s/%s", *resp.Group.UserPoolId, *resp.Group.GroupName))
	d.Set("user_pool_id", resp.Group.UserPoolId)
	d.Set("name", resp.Group.GroupName)
	d.Set("description", resp.Group.Description)
	d.Set("precedence", resp.Group.Precedence)
	d.Set("role_arn", resp.Group.RoleArn)

	return nil
}

func resourceAwsCognitoUserPoolGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	params := &cognitoidentityprovider.UpdateGroupInput{
		GroupName:  aws.String(d.Get("name").(string)),
		UserPoolId: aws.String(d.Get("user_pool_id").(string)),
	}

	if d.HasChange("description") {
		params.Description = aws.String(d.Get("description").(string))
	}

	if d.HasChange("precedence") {
		params.Precedence = aws.Int64(int64(d.Get("precedence").(int)))
	}

	if d.HasChange("role_arn") {
		params.RoleArn = aws.String(d.Get("role_arn").(string))
	}

	log.Printf("[DEBUG] Updating Cognito User Pool Group: %s", params)

	_, err := conn.UpdateGroup(params)
	if err != nil {
		return errwrap.Wrapf("Error updating Cognito User Pool Group: {{err}}", err)
	}

	return resourceAwsCognitoUserPoolGroupRead(d, meta)
}

func resourceAwsCognitoUserPoolGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	params := &cognitoidentityprovider.DeleteGroupInput{
		GroupName:  aws.String(d.Get("name").(string)),
		UserPoolId: aws.String(d.Get("user_pool_id").(string)),
	}

	log.Printf("[DEBUG] Deleting Cognito User Pool Group: %s", params)

	_, err := conn.DeleteGroup(params)

	if err != nil {
		return errwrap.Wrapf("Error deleting Cognito User Pool Group: {{err}}", err)
	}

	return nil
}
