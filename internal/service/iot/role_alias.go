package iot

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func ResourceRoleAlias() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRoleAliasCreate,
		ReadWithoutTimeout:   resourceRoleAliasRead,
		UpdateWithoutTimeout: resourceRoleAliasUpdate,
		DeleteWithoutTimeout: resourceRoleAliasDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"alias": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"role_arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"credential_duration": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      3600,
				ValidateFunc: validation.IntBetween(900, 43200),
			},
		},
	}
}

func resourceRoleAliasCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn()

	roleAlias := d.Get("alias").(string)
	roleArn := d.Get("role_arn").(string)
	credentialDuration := d.Get("credential_duration").(int)

	_, err := conn.CreateRoleAliasWithContext(ctx, &iot.CreateRoleAliasInput{
		RoleAlias:                 aws.String(roleAlias),
		RoleArn:                   aws.String(roleArn),
		CredentialDurationSeconds: aws.Int64(int64(credentialDuration)),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating role alias %s for role %s: %s", roleAlias, roleArn, err)
	}

	d.SetId(roleAlias)
	return append(diags, resourceRoleAliasRead(ctx, d, meta)...)
}

func GetRoleAliasDescription(ctx context.Context, conn *iot.IoT, alias string) (*iot.RoleAliasDescription, error) {
	roleAliasDescriptionOutput, err := conn.DescribeRoleAliasWithContext(ctx, &iot.DescribeRoleAliasInput{
		RoleAlias: aws.String(alias),
	})

	if err != nil {
		return nil, err
	}

	if roleAliasDescriptionOutput == nil {
		return nil, nil
	}

	return roleAliasDescriptionOutput.RoleAliasDescription, nil
}

func resourceRoleAliasRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn()

	var roleAliasDescription *iot.RoleAliasDescription

	roleAliasDescription, err := GetRoleAliasDescription(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing role alias %s: %s", d.Id(), err)
	}

	if roleAliasDescription == nil {
		log.Printf("[WARN] Role alias (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	d.Set("arn", roleAliasDescription.RoleAliasArn)
	d.Set("alias", roleAliasDescription.RoleAlias)
	d.Set("role_arn", roleAliasDescription.RoleArn)
	d.Set("credential_duration", roleAliasDescription.CredentialDurationSeconds)

	return diags
}

func resourceRoleAliasDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn()

	alias := d.Get("alias").(string)

	_, err := conn.DeleteRoleAliasWithContext(ctx, &iot.DeleteRoleAliasInput{
		RoleAlias: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting role alias %s: %s", alias, err)
	}

	return diags
}

func resourceRoleAliasUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn()

	if d.HasChange("credential_duration") {
		roleAliasInput := &iot.UpdateRoleAliasInput{
			RoleAlias:                 aws.String(d.Id()),
			CredentialDurationSeconds: aws.Int64(int64(d.Get("credential_duration").(int))),
		}
		_, err := conn.UpdateRoleAliasWithContext(ctx, roleAliasInput)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating role alias %s: %s", d.Id(), err)
		}
	}

	if d.HasChange("role_arn") {
		roleAliasInput := &iot.UpdateRoleAliasInput{
			RoleAlias: aws.String(d.Id()),
			RoleArn:   aws.String(d.Get("role_arn").(string)),
		}
		_, err := conn.UpdateRoleAliasWithContext(ctx, roleAliasInput)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating role alias %s: %s", d.Id(), err)
		}
	}

	return append(diags, resourceRoleAliasRead(ctx, d, meta)...)
}
