package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codestarconnections"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsCodeStarConnectionsHost() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCodeStarConnectionsHostCreate,
		Read:   resourceAwsCodeStarConnectionsHostRead,
		Update: resourceAwsCodeStarConnectionsHostUpdate,
		Delete: resourceAwsCodeStarConnectionsHostDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"provider_endpoint": {
				Type:     schema.TypeString,
				Required: true,
			},
			"provider_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					codestarconnections.ProviderTypeGitHubEnterpriseServer,
				}, false),
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"security_group_ids": {
							Type:     schema.TypeSet,
							Required: true,
							MinItems: 1,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							Required: true,
							MinItems: 1,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"tls_certificate": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"vpc_id": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceAwsCodeStarConnectionsHostCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codestarconnectionsconn

	params := &codestarconnections.CreateHostInput{
		Name:             aws.String(d.Get("name").(string)),
		ProviderEndpoint: aws.String(d.Get("provider_endpoint").(string)),
		ProviderType:     aws.String(d.Get("provider_type").(string)),
		VpcConfiguration: expandCodeStarConnectionsHostVpcConfiguration(d.Get("vpc_configuration").([]interface{})),
	}

	resp, err := conn.CreateHost(params)
	if err != nil {
		return fmt.Errorf("error creating CodeStar host: %w", err)
	}

	d.SetId(aws.StringValue(resp.HostArn))
	d.Set("arn", aws.StringValue(resp.HostArn))

	if err := waitForCodeStarConnectionsHost(conn, d.Id()); err != nil {
		return err
	}

	return resourceAwsCodeStarConnectionsHostRead(d, meta)
}

func resourceAwsCodeStarConnectionsHostRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codestarconnectionsconn

	resp, err := conn.GetHost(&codestarconnections.GetHostInput{
		HostArn: aws.String(d.Id()),
	})
	if tfawserr.ErrCodeEquals(err, codestarconnections.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] CodeStar host (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading CodeStar host: %w", err)
	}
	if resp == nil {
		return fmt.Errorf("error reading CodeStar host (%s): empty response", d.Id())
	}

	d.Set("arn", d.Id())
	d.Set("name", resp.Name)
	d.Set("provider_endpoint", resp.ProviderEndpoint)
	d.Set("provider_type", resp.ProviderType)
	d.Set("status", resp.Status)
	d.Set("vpc_configuration", flattenCodeStarConnectionsHostVpcConfiguration(resp.VpcConfiguration))

	return nil
}

func resourceAwsCodeStarConnectionsHostUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codestarconnectionsconn

	if d.HasChanges("provider_endpoint", "vpc_configuration") {
		input := codestarconnections.UpdateHostInput{
			HostArn:          aws.String(d.Get("name").(string)),
			ProviderEndpoint: aws.String(d.Get("provider_endpoint").(string)),
			VpcConfiguration: expandCodeStarConnectionsHostVpcConfiguration(d.Get("vpc_configuration").([]interface{})),
		}

		_, err := conn.UpdateHost(&input)
		if err != nil {
			return fmt.Errorf("error updating CodeStar host: %w", err)
		}

		if err := waitForCodeStarConnectionsHost(conn, d.Id()); err != nil {
			return err
		}
	}

	return resourceAwsCodeStarConnectionsHostRead(d, meta)
}

func resourceAwsCodeStarConnectionsHostDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codestarconnectionsconn

	_, err := conn.DeleteHost(&codestarconnections.DeleteHostInput{
		HostArn: aws.String(d.Id()),
	})
	if tfawserr.ErrCodeEquals(err, codestarconnections.ErrCodeResourceNotFoundException) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting CodeStar host: %w", err)
	}

	return nil
}

func expandCodeStarConnectionsHostVpcConfiguration(l []interface{}) *codestarconnections.VpcConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	vc := &codestarconnections.VpcConfiguration{
		SecurityGroupIds: expandStringSet(m["security_group_ids"].(*schema.Set)),
		SubnetIds:        expandStringSet(m["subnet_ids"].(*schema.Set)),
		VpcId:            aws.String(m["vpc_id"].(string)),
	}

	if v := m["tls_certificate"].(string); len(v) != 0 {
		vc.TlsCertificate = aws.String(v)
	}

	return vc
}

func flattenCodeStarConnectionsHostVpcConfiguration(vpcConfig *codestarconnections.VpcConfiguration) []interface{} {
	if vpcConfig == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"security_group_ids": flattenStringSet(vpcConfig.SecurityGroupIds),
		"subnet_ids":         flattenStringSet(vpcConfig.SubnetIds),
		"vpc_id":             aws.StringValue(vpcConfig.VpcId),
	}

	if vpcConfig.TlsCertificate != nil {
		m["tls_certificate"] = aws.StringValue(vpcConfig.TlsCertificate)
	}

	return []interface{}{m}
}

func waitForCodeStarConnectionsHost(conn *codestarconnections.CodeStarConnections, arn string) error {
	stateConf := resource.StateChangeConf{
		Pending: []string{
			"VPC_CONFIG_INITIALIZING",
			"VPC_CONFIG_DELETING",
		},
		Target: []string{
			"VPC_CONFIG_FAILED_INITIALIZATION",
			"PENDING",
			"AVAILABLE",
		},
		Timeout: 30 * time.Minute,
		Delay:   1 * time.Minute,
		Refresh: func() (interface{}, string, error) {
			out, err := conn.GetHost(&codestarconnections.GetHostInput{
				HostArn: aws.String(arn),
			})

			if err != nil {
				return nil, "", err
			}

			return out, *out.Status, nil
		},
	}
	_, err := stateConf.WaitForState()
	return err
}
