package cognitoidp

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceUserGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceUserGroupCreate,
		Read:   resourceUserGroupRead,
		Update: resourceUserGroupUpdate,
		Delete: resourceUserGroupDelete,

		Importer: &schema.ResourceImporter{
			State: resourceUserGroupImport,
		},

		// https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_CreateGroup.html
		Schema: map[string]*schema.Schema{
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 2048),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validUserGroupName,
			},
			"precedence": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"user_pool_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validUserPoolID,
			},
		},
	}
}

func resourceUserGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CognitoIDPConn

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

	log.Print("[DEBUG] Creating Cognito User Group")

	resp, err := conn.CreateGroup(params)
	if err != nil {
		return fmt.Errorf("Error creating Cognito User Group: %s", err)
	}

	d.SetId(fmt.Sprintf("%s/%s", *resp.Group.UserPoolId, *resp.Group.GroupName))

	return resourceUserGroupRead(d, meta)
}

func resourceUserGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CognitoIDPConn

	params := &cognitoidentityprovider.GetGroupInput{
		GroupName:  aws.String(d.Get("name").(string)),
		UserPoolId: aws.String(d.Get("user_pool_id").(string)),
	}

	log.Print("[DEBUG] Reading Cognito User Group")

	resp, err := conn.GetGroup(params)
	if !d.IsNewResource() && tfresource.NotFound(err) {
		names.LogNotFoundRemoveState(names.CognitoIDP, names.ErrActionReading, ResUserGroup, d.Get("name").(string))
		d.SetId("")
		return nil
	}

	if err != nil {
		return names.Error(names.CognitoIDP, names.ErrActionReading, ResUserGroup, d.Get("name").(string), err)
	}

	d.Set("description", resp.Group.Description)
	d.Set("precedence", resp.Group.Precedence)
	d.Set("role_arn", resp.Group.RoleArn)

	return nil
}

func resourceUserGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CognitoIDPConn

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

	log.Print("[DEBUG] Updating Cognito User Group")

	_, err := conn.UpdateGroup(params)
	if err != nil {
		return fmt.Errorf("Error updating Cognito User Group: %s", err)
	}

	return resourceUserGroupRead(d, meta)
}

func resourceUserGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CognitoIDPConn

	params := &cognitoidentityprovider.DeleteGroupInput{
		GroupName:  aws.String(d.Get("name").(string)),
		UserPoolId: aws.String(d.Get("user_pool_id").(string)),
	}

	log.Print("[DEBUG] Deleting Cognito User Group")

	_, err := conn.DeleteGroup(params)
	if err != nil {
		return fmt.Errorf("Error deleting Cognito User Group: %s", err)
	}

	return nil
}

func resourceUserGroupImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idSplit := strings.Split(d.Id(), "/")
	if len(idSplit) != 2 {
		return nil, errors.New("Error importing Cognito User Group. Must specify user_pool_id/group_name")
	}
	userPoolId := idSplit[0]
	name := idSplit[1]
	d.Set("user_pool_id", userPoolId)
	d.Set("name", name)
	return []*schema.ResourceData{d}, nil
}
