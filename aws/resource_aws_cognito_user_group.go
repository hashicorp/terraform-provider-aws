package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsCognitoUserGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCognitoUserGroupCreate,
		Read:   resourceAwsCognitoUserGroupRead,
		Update: resourceAwsCognitoUserGroupUpdate,
		Delete: resourceAwsCognitoUserGroupDelete,

		Schema: map[string]*schema.Schema{
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateMaxLength(2048),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateCognitoUserGroupName,
			},
			"precedence": {
				Type:     schema.TypeInt,
				Optional: true,
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
				ValidateFunc: validateCognitoUserPoolId,
			},
		},
	}
}

func resourceAwsCognitoUserGroupCreate(d *schema.ResourceData, meta interface{}) error {
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

	resp, err := conn.CreateGroup(params)
	if err != nil {
		return errwrap.Wrapf("Error creating Cognito User Group: {{err}}", err)
	}

	d.SetId(*resp.Group.GroupName)

	return resourceAwsCognitoUserGroupRead(d, meta)
}

func resourceAwsCognitoUserGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	params := &cognitoidentityprovider.GetGroupInput{
		GroupName:  aws.String(d.Get("name").(string)),
		UserPoolId: aws.String(d.Get("user_pool_id").(string)),
	}

	resp, err := conn.GetGroup(params)
	if err != nil {
		return errwrap.Wrapf("Error reading Cognito User Group: {{err}}", err)
	}

	if resp.Group.Description != nil {
		d.Set("description", *resp.Group.Description)
	}

	if resp.Group.Precedence != nil {
		d.Set("precedence", *resp.Group.Precedence)
	}

	if resp.Group.RoleArn != nil {
		d.Set("role_arn", *resp.Group.RoleArn)
	}

	return nil
}

func resourceAwsCognitoUserGroupUpdate(d *schema.ResourceData, meta interface{}) error {
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
		params.RoleArn = aws.String(d.Get("description").(string))
	}

	_, err := conn.UpdateGroup(params)
	if err != nil {
		return errwrap.Wrapf("Error updating Cognito User Group: {{err}}", err)
	}

	return resourceAwsCognitoUserGroupRead(d, meta)
}

func resourceAwsCognitoUserGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	params := &cognitoidentityprovider.DeleteGroupInput{
		GroupName:  aws.String(d.Get("name").(string)),
		UserPoolId: aws.String(d.Get("user_pool_id").(string)),
	}

	_, err := conn.DeleteGroup(params)

	if err != nil {
		return errwrap.Wrapf("Error deleting Cognito User Group: {{err}}", err)
	}

	return nil
}
