package opsworks

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func ResourceUserProfile() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUserProfileCreate,
		ReadWithoutTimeout:   resourceUserProfileRead,
		UpdateWithoutTimeout: resourceUserProfileUpdate,
		DeleteWithoutTimeout: resourceUserProfileDelete,

		Schema: map[string]*schema.Schema{
			"user_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"allow_self_management": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"ssh_username": {
				Type:     schema.TypeString,
				Required: true,
			},

			"ssh_public_key": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceUserProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*conns.AWSClient).OpsWorksConn()

	req := &opsworks.DescribeUserProfilesInput{
		IamUserArns: []*string{
			aws.String(d.Id()),
		},
	}

	log.Printf("[DEBUG] Reading OpsWorks user profile: %s", d.Id())

	resp, err := client.DescribeUserProfilesWithContext(ctx, req)
	if tfawserr.ErrCodeEquals(err, opsworks.ErrCodeResourceNotFoundException) {
		log.Printf("[DEBUG] OpsWorks user profile (%s) not found", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading OpsWorks User Profile (%s): %s", d.Id(), err)
	}

	for _, profile := range resp.UserProfiles {
		d.Set("allow_self_management", profile.AllowSelfManagement)
		d.Set("user_arn", profile.IamUserArn)
		d.Set("ssh_public_key", profile.SshPublicKey)
		d.Set("ssh_username", profile.SshUsername)
		break
	}

	return diags
}

func resourceUserProfileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*conns.AWSClient).OpsWorksConn()

	req := &opsworks.CreateUserProfileInput{
		AllowSelfManagement: aws.Bool(d.Get("allow_self_management").(bool)),
		IamUserArn:          aws.String(d.Get("user_arn").(string)),
		SshPublicKey:        aws.String(d.Get("ssh_public_key").(string)),
		SshUsername:         aws.String(d.Get("ssh_username").(string)),
	}

	resp, err := client.CreateUserProfileWithContext(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating OpsWorks User Profile (%s): %s", d.Id(), err)
	}

	d.SetId(aws.StringValue(resp.IamUserArn))

	return append(diags, resourceUserProfileUpdate(ctx, d, meta)...)
}

func resourceUserProfileUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*conns.AWSClient).OpsWorksConn()

	req := &opsworks.UpdateUserProfileInput{
		AllowSelfManagement: aws.Bool(d.Get("allow_self_management").(bool)),
		IamUserArn:          aws.String(d.Get("user_arn").(string)),
		SshPublicKey:        aws.String(d.Get("ssh_public_key").(string)),
		SshUsername:         aws.String(d.Get("ssh_username").(string)),
	}

	log.Printf("[DEBUG] Updating OpsWorks user profile: %s", req)

	_, err := client.UpdateUserProfileWithContext(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating OpsWorks User Profile (%s): %s", d.Id(), err)
	}

	return append(diags, resourceUserProfileRead(ctx, d, meta)...)
}

func resourceUserProfileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*conns.AWSClient).OpsWorksConn()

	req := &opsworks.DeleteUserProfileInput{
		IamUserArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting OpsWorks user profile: %s", d.Id())

	_, err := client.DeleteUserProfileWithContext(ctx, req)

	if tfawserr.ErrCodeEquals(err, opsworks.ErrCodeResourceNotFoundException) {
		log.Printf("[DEBUG] OpsWorks User Profile (%s) not found to delete; removed from state", d.Id())
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting OpsWorks User Profile (%s): %s", d.Id(), err)
	}

	return diags
}
