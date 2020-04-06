package aws

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsDbProxy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDbProxyCreate,
		Read:   resourceAwsDbProxyRead,
		Update: resourceAwsDbProxyUpdate,
		Delete: resourceAwsDbProxyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateRdsIdentifier,
			},
			"debug_logging": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"engine_family": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					rds.EngineFamilyMysql,
				}, false),
			},
			"idle_client_timeout": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"require_tls": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
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
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"auth": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auth_scheme": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								rds.AuthSchemeSecrets,
							}, false),
						},
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"iam_auth": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								rds.IAMAuthModeDisabled,
								rds.IAMAuthModeRequired,
							}, false),
						},
						"secret_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateArn,
						},
						"username": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
				Set: resourceAwsDbProxyAuthHash,
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceAwsDbProxyCreate(d *schema.ResourceData, meta interface{}) error {
	rdsconn := meta.(*AWSClient).rdsconn
	tags := keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().RdsTags()

	params := rds.CreateDBProxyInput{
		Auth:         expandDbProxyAuth(d.Get("auth").(*schema.Set).List()),
		DBProxyName:  aws.String(d.Get("name").(string)),
		EngineFamily: aws.String(d.Get("engine_family").(string)),
		RoleArn:      aws.String(d.Get("role_arn").(string)),
		Tags:         tags,
		VpcSubnetIds: expandStringSet(d.Get("vpc_subnet_ids").(*schema.Set)),
	}

	if v, ok := d.GetOk("debug_logging"); ok {
		params.DebugLogging = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("idle_client_timeout"); ok {
		params.IdleClientTimeout = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("require_tls"); ok {
		params.RequireTLS = aws.Bool(v.(bool))
	}

	if v := d.Get("vpc_security_group_ids").(*schema.Set); v.Len() > 0 {
		params.VpcSecurityGroupIds = expandStringSet(v)
	}

	log.Printf("[DEBUG] Create DB Proxy: %#v", params)
	resp, err := rdsconn.CreateDBProxy(&params)
	if err != nil {
		return fmt.Errorf("Error creating DB Proxy: %s", err)
	}

	d.SetId(aws.StringValue(resp.DBProxy.DBProxyName))
	d.Set("arn", resp.DBProxy.DBProxyArn)
	log.Printf("[INFO] DB Proxy ID: %s", d.Id())

	return resourceAwsDbProxyRead(d, meta)
}

func expandDbProxyAuth(l []interface{}) []*rds.UserAuthConfig {
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

func resourceAwsDbProxyRead(d *schema.ResourceData, meta interface{}) error {
	rdsconn := meta.(*AWSClient).rdsconn

	params := rds.DescribeDBProxiesInput{
		DBProxyName: aws.String(d.Id()),
	}

	resp, err := rdsconn.DescribeDBProxies(&params)
	if err != nil {
		if isAWSErr(err, rds.ErrCodeDBProxyNotFoundFault, "") {
			log.Printf("[WARN] DB Proxy (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	if len(resp.DBProxies) != 1 ||
		*resp.DBProxies[0].DBProxyName != d.Id() {
		return fmt.Errorf("Unable to find DB Proxy: %#v", resp.DBProxies)
	}

	v := resp.DBProxies[0]

	d.Set("arn", aws.StringValue(v.DBProxyArn))
	d.Set("name", v.DBProxyName)
	d.Set("debug_logging", v.DebugLogging)
	d.Set("engine_family", v.EngineFamily)
	d.Set("idle_client_timeout", v.IdleClientTimeout)
	d.Set("require_tls", v.RequireTLS)
	d.Set("role_arn", v.RoleArn)
	d.Set("vpc_subnet_ids", flattenStringSet(v.VpcSubnetIds))
	d.Set("security_group_ids", flattenStringSet(v.VpcSecurityGroupIds))

	tags, err := keyvaluetags.RdsListTags(rdsconn, d.Get("arn").(string))

	if err != nil {
		return fmt.Errorf("Error listing tags for RDS DB Proxy (%s): %s", d.Get("arn").(string), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().Map()); err != nil {
		return fmt.Errorf("Error setting tags: %s", err)
	}

	return nil
}

func resourceAwsDbProxyUpdate(d *schema.ResourceData, meta interface{}) error {
	rdsconn := meta.(*AWSClient).rdsconn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.RdsUpdateTags(rdsconn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("Error updating RDS DB Proxy (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	return resourceAwsDbProxyRead(d, meta)
}

func resourceAwsDbProxyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn
	params := rds.DeleteDBProxyInput{
		DBProxyName: aws.String(d.Id()),
	}
	err := resource.Retry(3*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteDBProxy(&params)
		if err != nil {
			if isAWSErr(err, rds.ErrCodeDBProxyNotFoundFault, "") || isAWSErr(err, rds.ErrCodeInvalidDBProxyStateFault, "") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		_, err = conn.DeleteDBProxy(&params)
	}
	if err != nil {
		return fmt.Errorf("Error deleting DB Proxy: %s", err)
	}
	return nil
}

func resourceAwsDbProxyAuthHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	if v, ok := m["auth_scheme"].(string); ok {
		buf.WriteString(fmt.Sprintf("%s-", v))
	}
	if v, ok := m["description"].(string); ok {
		buf.WriteString(fmt.Sprintf("%s-", v))
	}
	if v, ok := m["iam_auth"].(string); ok {
		buf.WriteString(fmt.Sprintf("%s-", v))
	}
	if v, ok := m["secret_arn"].(string); ok {
		buf.WriteString(fmt.Sprintf("%s-", v))
	}
	if v, ok := m["username"].(string); ok {
		buf.WriteString(fmt.Sprintf("%s-", v))
	}
	return hashcode.String(buf.String())
}
