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
					rds.EngineFamilyPostgresql,
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
				ForceNew: true,
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
	conn := meta.(*AWSClient).rdsconn
	tags := keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().RdsTags()

	params := rds.CreateDBProxyInput{
		Auth:         expandDbProxyAuth(d.Get("auth").(*schema.Set).List()),
		DBProxyName:  aws.String(d.Get("name").(string)),
		DebugLogging: aws.Bool(d.Get("debug_logging").(bool)),
		EngineFamily: aws.String(d.Get("engine_family").(string)),
		RequireTLS:   aws.Bool(d.Get("require_tls").(bool)),
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
	resp, err := conn.CreateDBProxy(&params)
	if err != nil {
		return fmt.Errorf("Error creating DB Proxy: %s", err)
	}

	d.SetId(aws.StringValue(resp.DBProxy.DBProxyName))
	d.Set("arn", resp.DBProxy.DBProxyArn)
	log.Printf("[INFO] DB Proxy ID: %s", d.Id())

	stateChangeConf := &resource.StateChangeConf{
		Pending: []string{rds.DBProxyStatusCreating},
		Target:  []string{rds.DBProxyStatusAvailable},
		Refresh: resourceAwsDbProxyRefreshFunc(conn, d.Id()),
		Timeout: d.Timeout(schema.TimeoutCreate),
	}

	_, err = stateChangeConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for DB Proxy creation: %s", err)
	}

	return resourceAwsDbProxyRead(d, meta)
}

func resourceAwsDbProxyRefreshFunc(conn *rds.RDS, proxyName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeDBProxies(&rds.DescribeDBProxiesInput{
			DBProxyName: aws.String(proxyName),
		})

		if err != nil {
			if isAWSErr(err, rds.ErrCodeDBProxyNotFoundFault, "") {
				return 42, "", nil
			}
			return 42, "", err
		}

		dbProxy := resp.DBProxies[0]
		return dbProxy, *dbProxy.Status, nil
	}
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

func flattenDbProxyAuth(userAuthConfig *rds.UserAuthConfigInfo) map[string]interface{} {
	m := make(map[string]interface{})

	m["auth_scheme"] = aws.StringValue(userAuthConfig.AuthScheme)
	m["description"] = aws.StringValue(userAuthConfig.Description)
	m["iam_auth"] = aws.StringValue(userAuthConfig.IAMAuth)
	m["secret_arn"] = aws.StringValue(userAuthConfig.SecretArn)
	m["username"] = aws.StringValue(userAuthConfig.UserName)

	return m
}

func flattenDbProxyAuths(userAuthConfigs []*rds.UserAuthConfigInfo) *schema.Set {
	s := []interface{}{}
	for _, v := range userAuthConfigs {
		s = append(s, flattenDbProxyAuth(v))
	}
	return schema.NewSet(resourceAwsDbProxyAuthHash, s)
}

func resourceAwsDbProxyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn

	params := rds.DescribeDBProxiesInput{
		DBProxyName: aws.String(d.Id()),
	}

	resp, err := conn.DescribeDBProxies(&params)
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

	dbProxy := resp.DBProxies[0]

	d.Set("arn", aws.StringValue(dbProxy.DBProxyArn))
	d.Set("auth", flattenDbProxyAuths(dbProxy.Auth))
	d.Set("name", dbProxy.DBProxyName)
	d.Set("debug_logging", dbProxy.DebugLogging)
	d.Set("engine_family", dbProxy.EngineFamily)
	d.Set("idle_client_timeout", dbProxy.IdleClientTimeout)
	d.Set("require_tls", dbProxy.RequireTLS)
	d.Set("role_arn", dbProxy.RoleArn)
	d.Set("vpc_subnet_ids", flattenStringSet(dbProxy.VpcSubnetIds))
	d.Set("security_group_ids", flattenStringSet(dbProxy.VpcSecurityGroupIds))

	tags, err := keyvaluetags.RdsListTags(conn, d.Get("arn").(string))

	if err != nil {
		return fmt.Errorf("Error listing tags for RDS DB Proxy (%s): %s", d.Get("arn").(string), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().Map()); err != nil {
		return fmt.Errorf("Error setting tags: %s", err)
	}

	return nil
}

func resourceAwsDbProxyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn

	oName, nName := d.GetChange("name")

	params := rds.ModifyDBProxyInput{
		Auth:           expandDbProxyAuth(d.Get("auth").(*schema.Set).List()),
		DBProxyName:    aws.String(oName.(string)),
		NewDBProxyName: aws.String(nName.(string)),
		DebugLogging:   aws.Bool(d.Get("debug_logging").(bool)),
		RequireTLS:     aws.Bool(d.Get("require_tls").(bool)),
		RoleArn:        aws.String(d.Get("role_arn").(string)),
	}

	if v, ok := d.GetOk("idle_client_timeout"); ok {
		params.IdleClientTimeout = aws.Int64(int64(v.(int)))
	}

	if v := d.Get("vpc_security_group_ids").(*schema.Set); v.Len() > 0 {
		params.SecurityGroups = expandStringSet(v)
	}

	log.Printf("[DEBUG] Update DB Proxy: %#v", params)
	_, err := conn.ModifyDBProxy(&params)
	if err != nil {
		return fmt.Errorf("Error updating DB Proxy: %s", err)
	}

	stateChangeConf := &resource.StateChangeConf{
		Pending: []string{rds.DBProxyStatusModifying},
		Target:  []string{rds.DBProxyStatusAvailable},
		Refresh: resourceAwsDbProxyRefreshFunc(conn, d.Id()),
		Timeout: d.Timeout(schema.TimeoutCreate),
	}

	_, err = stateChangeConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for DB Proxy update: %s", err)
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.RdsUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
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
	_, err := conn.DeleteDBProxy(&params)
	if err != nil {
		return fmt.Errorf("Error deleting DB Proxy: %s", err)
	}

	stateChangeConf := &resource.StateChangeConf{
		Pending: []string{rds.DBProxyStatusDeleting},
		Target:  []string{""},
		Refresh: resourceAwsDbProxyRefreshFunc(conn, d.Id()),
		Timeout: d.Timeout(schema.TimeoutDelete),
	}

	_, err = stateChangeConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for DB Proxy deletion: %s", err)
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
