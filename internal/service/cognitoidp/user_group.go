package cognitoidp

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceUserGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUserGroupCreate,
		ReadWithoutTimeout:   resourceUserGroupRead,
		UpdateWithoutTimeout: resourceUserGroupUpdate,
		DeleteWithoutTimeout: resourceUserGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceUserGroupImport,
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

func resourceUserGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPConn()

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

	resp, err := conn.CreateGroupWithContext(ctx, params)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Error creating Cognito User Group: %s", err)
	}

	d.SetId(fmt.Sprintf("%s/%s", *resp.Group.UserPoolId, *resp.Group.GroupName))

	return append(diags, resourceUserGroupRead(ctx, d, meta)...)
}

func resourceUserGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPConn()

	params := &cognitoidentityprovider.GetGroupInput{
		GroupName:  aws.String(d.Get("name").(string)),
		UserPoolId: aws.String(d.Get("user_pool_id").(string)),
	}

	log.Print("[DEBUG] Reading Cognito User Group")

	resp, err := conn.GetGroupWithContext(ctx, params)
	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.CognitoIDP, create.ErrActionReading, ResNameUserGroup, d.Get("name").(string))
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.DiagError(names.CognitoIDP, create.ErrActionReading, ResNameUserGroup, d.Get("name").(string), err)
	}

	d.Set("description", resp.Group.Description)
	d.Set("precedence", resp.Group.Precedence)
	d.Set("role_arn", resp.Group.RoleArn)

	return diags
}

func resourceUserGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPConn()

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

	_, err := conn.UpdateGroupWithContext(ctx, params)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Error updating Cognito User Group: %s", err)
	}

	return append(diags, resourceUserGroupRead(ctx, d, meta)...)
}

func resourceUserGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPConn()

	params := &cognitoidentityprovider.DeleteGroupInput{
		GroupName:  aws.String(d.Get("name").(string)),
		UserPoolId: aws.String(d.Get("user_pool_id").(string)),
	}

	log.Print("[DEBUG] Deleting Cognito User Group")

	_, err := conn.DeleteGroupWithContext(ctx, params)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Error deleting Cognito User Group: %s", err)
	}

	return diags
}

func resourceUserGroupImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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
