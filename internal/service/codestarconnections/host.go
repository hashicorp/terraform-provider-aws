package codestarconnections

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codestarconnections"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceHost() *schema.Resource {
	return &schema.Resource{
		Create: resourceHostCreate,
		Read:   resourceHostRead,
		Update: resourceHostUpdate,
		Delete: resourceHostDelete,

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
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"provider_endpoint": {
				Type:     schema.TypeString,
				Required: true,
			},
			"provider_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(codestarconnections.ProviderType_Values(), false),
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

func resourceHostCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodeStarConnectionsConn

	name := d.Get("name").(string)
	input := &codestarconnections.CreateHostInput{
		Name:             aws.String(name),
		ProviderEndpoint: aws.String(d.Get("provider_endpoint").(string)),
		ProviderType:     aws.String(d.Get("provider_type").(string)),
		VpcConfiguration: expandHostVPCConfiguration(d.Get("vpc_configuration").([]interface{})),
	}

	log.Printf("[DEBUG] Creating CodeStar Connections Host: %s", input)
	output, err := conn.CreateHost(input)

	if err != nil {
		return fmt.Errorf("creating CodeStar Connections Host (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(output.HostArn))

	if _, err := waitHostPendingOrAvailable(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("waiting for CodeStar Connections Host (%s) create: %w", d.Id(), err)
	}

	return resourceHostRead(d, meta)
}

func resourceHostRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodeStarConnectionsConn

	output, err := FindHostByARN(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CodeStar Connections Host (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading CodeStar Connections Host (%s): %w", d.Id(), err)
	}

	d.Set("arn", d.Id())
	d.Set("name", output.Name)
	d.Set("provider_endpoint", output.ProviderEndpoint)
	d.Set("provider_type", output.ProviderType)
	d.Set("status", output.Status)
	d.Set("vpc_configuration", flattenHostVPCConfiguration(output.VpcConfiguration))

	return nil
}

func resourceHostUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodeStarConnectionsConn

	if d.HasChanges("provider_endpoint", "vpc_configuration") {
		input := &codestarconnections.UpdateHostInput{
			HostArn:          aws.String(d.Id()),
			ProviderEndpoint: aws.String(d.Get("provider_endpoint").(string)),
			VpcConfiguration: expandHostVPCConfiguration(d.Get("vpc_configuration").([]interface{})),
		}

		_, err := conn.UpdateHost(input)

		if err != nil {
			return fmt.Errorf("updating CodeStar Connections Host (%s): %w", d.Id(), err)
		}

		if _, err := waitHostPendingOrAvailable(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("waiting for CodeStar Connections Host (%s) update: %w", d.Id(), err)
		}
	}

	return resourceHostRead(d, meta)
}

func resourceHostDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodeStarConnectionsConn

	log.Printf("[DEBUG] Deleting CodeStar Connections Host: %s", d.Id())
	_, err := conn.DeleteHost(&codestarconnections.DeleteHostInput{
		HostArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, codestarconnections.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting CodeStar Connections Host (%s): %w", d.Id(), err)
	}

	return nil
}

func expandHostVPCConfiguration(l []interface{}) *codestarconnections.VpcConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	vc := &codestarconnections.VpcConfiguration{
		SecurityGroupIds: flex.ExpandStringSet(m["security_group_ids"].(*schema.Set)),
		SubnetIds:        flex.ExpandStringSet(m["subnet_ids"].(*schema.Set)),
		VpcId:            aws.String(m["vpc_id"].(string)),
	}

	if v, ok := m["tls_certificate"].(string); ok && v != "" {
		vc.TlsCertificate = aws.String(v)
	}

	return vc
}

func flattenHostVPCConfiguration(vpcConfig *codestarconnections.VpcConfiguration) []interface{} {
	if vpcConfig == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"security_group_ids": flex.FlattenStringSet(vpcConfig.SecurityGroupIds),
		"subnet_ids":         flex.FlattenStringSet(vpcConfig.SubnetIds),
		"vpc_id":             aws.StringValue(vpcConfig.VpcId),
	}

	if vpcConfig.TlsCertificate != nil {
		m["tls_certificate"] = aws.StringValue(vpcConfig.TlsCertificate)
	}

	return []interface{}{m}
}

func statusHost(conn *codestarconnections.CodeStarConnections, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindHostByARN(conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

const (
	hostStatusAvailable = "AVAILABLE"
	hostStatusPending   = "PENDING"
	// hostStatusVPCConfigDeleting             = "VPC_CONFIG_DELETING"
	// hostStatusVPCConfigFailedInitialization = "VPC_CONFIG_FAILED_INITIALIZATION"
	hostStatusVPCConfigInitializing = "VPC_CONFIG_INITIALIZING"
)

func waitHostPendingOrAvailable(conn *codestarconnections.CodeStarConnections, arn string, timeout time.Duration) (*codestarconnections.Host, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending: []string{hostStatusVPCConfigInitializing},
		Target:  []string{hostStatusAvailable, hostStatusPending},
		Refresh: statusHost(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*codestarconnections.Host); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))

		return output, err
	}

	return nil, err
}
