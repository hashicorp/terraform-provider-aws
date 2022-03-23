package cognitoidp

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceUserInGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceUserInGroupCreate,
		Read:   resourceUserInGroupRead,
		Delete: resourceUserInGroupDelete,
		Schema: map[string]*schema.Schema{
			"group_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validUserGroupName,
			},
			"user_pool_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validUserPoolID,
			},
			"username": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
		},
	}
}

func resourceUserInGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CognitoIDPConn

	input := &cognitoidentityprovider.AdminAddUserToGroupInput{}

	if v, ok := d.GetOk("group_name"); ok {
		input.GroupName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("user_pool_id"); ok {
		input.UserPoolId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("username"); ok {
		input.Username = aws.String(v.(string))
	}

	_, err := conn.AdminAddUserToGroup(input)

	if err != nil {
		return fmt.Errorf("error adding user to group: %w", err)
	}

	//lintignore:R015 // Allow legacy unstable ID usage in managed resource
	d.SetId(resource.UniqueId())

	return resourceUserInGroupRead(d, meta)
}

func resourceUserInGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CognitoIDPConn

	groupName := d.Get("group_name").(string)
	userPoolId := d.Get("user_pool_id").(string)
	username := d.Get("username").(string)

	found, err := FindCognitoUserInGroup(conn, groupName, userPoolId, username)

	if err != nil {
		return err
	}

	if !found {
		d.SetId("")
	}

	return nil
}

func resourceUserInGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CognitoIDPConn

	groupName := d.Get("group_name").(string)
	userPoolID := d.Get("user_pool_id").(string)
	username := d.Get("username").(string)

	input := &cognitoidentityprovider.AdminRemoveUserFromGroupInput{
		GroupName:  aws.String(groupName),
		UserPoolId: aws.String(userPoolID),
		Username:   aws.String(username),
	}

	_, err := conn.AdminRemoveUserFromGroup(input)

	if err != nil {
		return fmt.Errorf("error removing user from group: %w", err)
	}

	return nil
}
