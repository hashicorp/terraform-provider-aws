// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_db_proxy", name="DB Proxy")
// @Tags(identifierAttribute="arn")
func ResourceProxy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceProxyCreate,
		ReadWithoutTimeout:   resourceProxyRead,
		UpdateWithoutTimeout: resourceProxyUpdate,
		DeleteWithoutTimeout: resourceProxyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auth": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auth_scheme": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(rds.AuthScheme_Values(), false),
						},
						"client_password_auth_type": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice(rds.ClientPasswordAuthType_Values(), false),
						},
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"iam_auth": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(rds.IAMAuthMode_Values(), false),
						},
						"secret_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
						"username": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"debug_logging": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine_family": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(rds.EngineFamily_Values(), false),
			},
			"idle_client_timeout": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validIdentifier,
			},
			"require_tls": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"vpc_security_group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"vpc_subnet_ids": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceProxyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	input := rds.CreateDBProxyInput{
		Auth:         expandProxyAuth(d.Get("auth").([]interface{})),
		DBProxyName:  aws.String(d.Get("name").(string)),
		EngineFamily: aws.String(d.Get("engine_family").(string)),
		RoleArn:      aws.String(d.Get("role_arn").(string)),
		Tags:         getTagsIn(ctx),
		VpcSubnetIds: flex.ExpandStringSet(d.Get("vpc_subnet_ids").(*schema.Set)),
	}

	if v, ok := d.GetOk("debug_logging"); ok {
		input.DebugLogging = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("idle_client_timeout"); ok {
		input.IdleClientTimeout = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("require_tls"); ok {
		input.RequireTLS = aws.Bool(v.(bool))
	}

	if v := d.Get("vpc_security_group_ids").(*schema.Set); v.Len() > 0 {
		input.VpcSecurityGroupIds = flex.ExpandStringSet(v)
	}

	output, err := conn.CreateDBProxyWithContext(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating RDS DB Proxy: %s", err)
	}

	d.SetId(aws.StringValue(output.DBProxy.DBProxyName))

	if _, err := waitDBProxyCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS DB Proxy (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceProxyRead(ctx, d, meta)...)
}

func resourceProxyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	dbProxy, err := FindDBProxyByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RDS DB Proxy %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Proxy (%s): %s", d.Id(), err)
	}

	d.Set("arn", dbProxy.DBProxyArn)
	d.Set("auth", flattenProxyAuths(dbProxy.Auth))
	d.Set("name", dbProxy.DBProxyName)
	d.Set("debug_logging", dbProxy.DebugLogging)
	d.Set("engine_family", dbProxy.EngineFamily)
	d.Set("idle_client_timeout", dbProxy.IdleClientTimeout)
	d.Set("require_tls", dbProxy.RequireTLS)
	d.Set("role_arn", dbProxy.RoleArn)
	d.Set("vpc_subnet_ids", flex.FlattenStringSet(dbProxy.VpcSubnetIds))
	d.Set("vpc_security_group_ids", flex.FlattenStringSet(dbProxy.VpcSecurityGroupIds))
	d.Set("endpoint", dbProxy.Endpoint)

	return diags
}

func resourceProxyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		oName, nName := d.GetChange("name")
		input := &rds.ModifyDBProxyInput{
			Auth:           expandProxyAuth(d.Get("auth").([]interface{})),
			DBProxyName:    aws.String(oName.(string)),
			DebugLogging:   aws.Bool(d.Get("debug_logging").(bool)),
			NewDBProxyName: aws.String(nName.(string)),
			RequireTLS:     aws.Bool(d.Get("require_tls").(bool)),
			RoleArn:        aws.String(d.Get("role_arn").(string)),
		}

		if v, ok := d.GetOk("idle_client_timeout"); ok {
			input.IdleClientTimeout = aws.Int64(int64(v.(int)))
		}

		if v := d.Get("vpc_security_group_ids").(*schema.Set); v.Len() > 0 {
			input.SecurityGroups = flex.ExpandStringSet(v)
		}

		_, err := conn.ModifyDBProxyWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating RDS DB Proxy (%s): %s", d.Id(), err)
		}

		// DB Proxy Name is used as an ID as the API doesn't provide a way to read/
		// update/delete DB proxies using the ARN
		d.SetId(nName.(string))

		if _, err := waitDBProxyUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for RDS DB Proxy (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceProxyRead(ctx, d, meta)...)
}

func resourceProxyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	log.Printf("[DEBUG] Deleting RDS DB Proxy: %s", d.Id())
	_, err := conn.DeleteDBProxyWithContext(ctx, &rds.DeleteDBProxyInput{
		DBProxyName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBProxyNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RDS DB Proxy (%s): %s", d.Id(), err)
	}

	if _, err := waitDBProxyDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS DB Proxy (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func expandProxyAuth(l []interface{}) []*rds.UserAuthConfig {
	if len(l) == 0 {
		return nil
	}

	userAuthConfigs := make([]*rds.UserAuthConfig, 0, len(l))

	for _, mRaw := range l {
		m, ok := mRaw.(map[string]interface{})

		if !ok {
			continue
		}

		userAuthConfig := &rds.UserAuthConfig{}

		if v, ok := m["auth_scheme"].(string); ok && v != "" {
			userAuthConfig.AuthScheme = aws.String(v)
		}

		if v, ok := m["client_password_auth_type"].(string); ok && v != "" {
			userAuthConfig.ClientPasswordAuthType = aws.String(v)
		}

		if v, ok := m["description"].(string); ok && v != "" {
			userAuthConfig.Description = aws.String(v)
		}

		if v, ok := m["iam_auth"].(string); ok && v != "" {
			userAuthConfig.IAMAuth = aws.String(v)
		}

		if v, ok := m["secret_arn"].(string); ok && v != "" {
			userAuthConfig.SecretArn = aws.String(v)
		}

		if v, ok := m["username"].(string); ok && v != "" {
			userAuthConfig.UserName = aws.String(v)
		}

		userAuthConfigs = append(userAuthConfigs, userAuthConfig)
	}

	return userAuthConfigs
}

func flattenProxyAuth(userAuthConfig *rds.UserAuthConfigInfo) map[string]interface{} {
	m := make(map[string]interface{})

	m["auth_scheme"] = aws.StringValue(userAuthConfig.AuthScheme)
	m["client_password_auth_type"] = aws.StringValue(userAuthConfig.ClientPasswordAuthType)
	m["description"] = aws.StringValue(userAuthConfig.Description)
	m["iam_auth"] = aws.StringValue(userAuthConfig.IAMAuth)
	m["secret_arn"] = aws.StringValue(userAuthConfig.SecretArn)
	m["username"] = aws.StringValue(userAuthConfig.UserName)

	return m
}

func flattenProxyAuths(userAuthConfigs []*rds.UserAuthConfigInfo) []interface{} {
	s := []interface{}{}
	for _, v := range userAuthConfigs {
		s = append(s, flattenProxyAuth(v))
	}
	return s
}
