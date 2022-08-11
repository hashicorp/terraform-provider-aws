package worklink

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/worklink"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceFleet() *schema.Resource {
	return &schema.Resource{
		Create: resourceFleetCreate,
		Read:   resourceFleetRead,
		Update: resourceFleetUpdate,
		Delete: resourceFleetDelete,
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
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringMatch(regexp.MustCompile(`^[a-z0-9](?:[a-z0-9\-]{0,46}[a-z0-9])?$`), "must contain only alphanumeric characters"),
					validation.StringLenBetween(1, 48),
				),
			},
			"display_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 100),
			},
			"audit_stream_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"network": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vpc_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"security_group_ids": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
					},
				},
			},
			"device_ca_certificate": {
				Type:     schema.TypeString,
				Optional: true,
				StateFunc: func(v interface{}) string {
					s, ok := v.(string)
					if !ok {
						return ""
					}
					return strings.TrimSpace(s)
				},
			},
			"identity_provider": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"saml_metadata": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 204800),
						},
					},
				},
			},
			"company_code": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_updated_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"optimize_for_end_user_location": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func resourceFleetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WorkLinkConn

	input := &worklink.CreateFleetInput{
		FleetName:                  aws.String(d.Get("name").(string)),
		OptimizeForEndUserLocation: aws.Bool(d.Get("optimize_for_end_user_location").(bool)),
	}

	if v, ok := d.GetOk("display_name"); ok {
		input.DisplayName = aws.String(v.(string))
	}

	resp, err := conn.CreateFleet(input)
	if err != nil {
		return fmt.Errorf("Error creating WorkLink Fleet: %w", err)
	}

	d.SetId(aws.StringValue(resp.FleetArn))

	if err := updateAuditStreamConfiguration(conn, d); err != nil {
		return err
	}

	if err := updateCompanyNetworkConfiguration(conn, d); err != nil {
		return err
	}

	if err := updateDevicePolicyConfiguration(conn, d); err != nil {
		return err
	}

	if err := updateIdentityProviderConfiguration(conn, d); err != nil {
		return err
	}

	return resourceFleetRead(d, meta)
}

func resourceFleetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WorkLinkConn

	resp, err := conn.DescribeFleetMetadata(&worklink.DescribeFleetMetadataInput{
		FleetArn: aws.String(d.Id()),
	})
	if err != nil {
		if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, worklink.ErrCodeResourceNotFoundException) {
			log.Printf("[WARN] WorkLink Fleet (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error describing WorkLink Fleet (%s): %w", d.Id(), err)
	}

	d.Set("arn", d.Id())
	d.Set("name", resp.FleetName)
	d.Set("display_name", resp.DisplayName)
	d.Set("optimize_for_end_user_location", resp.OptimizeForEndUserLocation)
	d.Set("company_code", resp.CompanyCode)
	d.Set("created_time", resp.CreatedTime.Format(time.RFC3339))
	if resp.LastUpdatedTime != nil {
		d.Set("last_updated_time", resp.LastUpdatedTime.Format(time.RFC3339))
	}
	auditStreamConfigurationResp, err := conn.DescribeAuditStreamConfiguration(&worklink.DescribeAuditStreamConfigurationInput{
		FleetArn: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("error describing WorkLink Fleet (%s) audit stream configuration: %w", d.Id(), err)
	}
	d.Set("audit_stream_arn", auditStreamConfigurationResp.AuditStreamArn)

	companyNetworkConfigurationResp, err := conn.DescribeCompanyNetworkConfiguration(&worklink.DescribeCompanyNetworkConfigurationInput{
		FleetArn: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("error describing WorkLink Fleet (%s) company network configuration: %w", d.Id(), err)
	}
	if err := d.Set("network", flattenNetworkConfigResponse(companyNetworkConfigurationResp)); err != nil {
		return fmt.Errorf("error setting network: %w", err)
	}

	identityProviderConfigurationResp, err := conn.DescribeIdentityProviderConfiguration(&worklink.DescribeIdentityProviderConfigurationInput{
		FleetArn: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("error describing WorkLink Fleet (%s) identity provider configuration: %w", d.Id(), err)
	}
	if err := d.Set("identity_provider", flattenIdentityProviderConfigResponse(identityProviderConfigurationResp)); err != nil {
		return fmt.Errorf("error setting identity_provider: %w", err)
	}

	devicePolicyConfigurationResp, err := conn.DescribeDevicePolicyConfiguration(&worklink.DescribeDevicePolicyConfigurationInput{
		FleetArn: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("error describing WorkLink Fleet (%s) device policy configuration: %w", d.Id(), err)
	}
	d.Set("device_ca_certificate", strings.TrimSpace(aws.StringValue(devicePolicyConfigurationResp.DeviceCaCertificate)))

	return nil
}

func resourceFleetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WorkLinkConn

	input := &worklink.UpdateFleetMetadataInput{
		FleetArn:                   aws.String(d.Id()),
		OptimizeForEndUserLocation: aws.Bool(d.Get("optimize_for_end_user_location").(bool)),
	}

	if v, ok := d.GetOk("display_name"); ok {
		input.DisplayName = aws.String(v.(string))
	}

	if d.HasChanges("display_name", "optimize_for_end_user_location") {
		_, err := conn.UpdateFleetMetadata(input)
		if err != nil {
			return fmt.Errorf("error updating WorkLink Fleet (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("audit_stream_arn") {
		if err := updateAuditStreamConfiguration(conn, d); err != nil {
			return err
		}
	}

	if d.HasChange("network") {
		if err := updateCompanyNetworkConfiguration(conn, d); err != nil {
			return err
		}
	}

	if d.HasChange("device_ca_certificate") {
		if err := updateDevicePolicyConfiguration(conn, d); err != nil {
			return err
		}
	}

	if d.HasChange("identity_provider") {
		if err := updateIdentityProviderConfiguration(conn, d); err != nil {
			return err
		}
	}

	return resourceFleetRead(d, meta)
}

func resourceFleetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WorkLinkConn

	input := &worklink.DeleteFleetInput{
		FleetArn: aws.String(d.Id()),
	}

	if _, err := conn.DeleteFleet(input); err != nil {
		if tfawserr.ErrCodeEquals(err, worklink.ErrCodeResourceNotFoundException) {
			return nil
		}
		return fmt.Errorf("error deleting WorkLink Fleet resource share (%s): %w", d.Id(), err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"DELETING"},
		Target:     []string{"DELETED"},
		Refresh:    FleetStateRefresh(conn, d.Id()),
		Timeout:    15 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for WorkLink Fleet (%s) to become deleted: %w", d.Id(), err)
	}

	return nil
}

func FleetStateRefresh(conn *worklink.WorkLink, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		emptyResp := &worklink.DescribeFleetMetadataOutput{}

		resp, err := conn.DescribeFleetMetadata(&worklink.DescribeFleetMetadataInput{
			FleetArn: aws.String(arn),
		})
		if err != nil {
			if tfawserr.ErrCodeEquals(err, worklink.ErrCodeResourceNotFoundException) {
				return emptyResp, "DELETED", nil
			}
		}

		return resp, *resp.FleetStatus, nil
	}
}

func updateAuditStreamConfiguration(conn *worklink.WorkLink, d *schema.ResourceData) error {
	input := &worklink.UpdateAuditStreamConfigurationInput{
		FleetArn: aws.String(d.Id()),
	}
	if v, ok := d.GetOk("audit_stream_arn"); ok {
		input.AuditStreamArn = aws.String(v.(string))
	} else if d.IsNewResource() {
		return nil
	}

	if _, err := conn.UpdateAuditStreamConfiguration(input); err != nil {
		return fmt.Errorf("error updating WorkLink Fleet (%s) Audit Stream Configuration: %w", d.Id(), err)
	}

	return nil
}

func updateCompanyNetworkConfiguration(conn *worklink.WorkLink, d *schema.ResourceData) error {
	oldNetwork, newNetwork := d.GetChange("network")
	if len(oldNetwork.([]interface{})) > 0 && len(newNetwork.([]interface{})) == 0 {
		return fmt.Errorf("Company Network Configuration cannot be removed from WorkLink Fleet(%s),"+
			" use 'terraform taint' to recreate the resource if you wish.", d.Id())
	}

	if v, ok := d.GetOk("network"); ok && len(v.([]interface{})) > 0 {
		config := v.([]interface{})[0].(map[string]interface{})
		input := &worklink.UpdateCompanyNetworkConfigurationInput{
			FleetArn:         aws.String(d.Id()),
			SecurityGroupIds: flex.ExpandStringSet(config["security_group_ids"].(*schema.Set)),
			SubnetIds:        flex.ExpandStringSet(config["subnet_ids"].(*schema.Set)),
			VpcId:            aws.String(config["vpc_id"].(string)),
		}
		if _, err := conn.UpdateCompanyNetworkConfiguration(input); err != nil {
			return fmt.Errorf("error updating WorkLink Fleet (%s) Company Network Configuration: %w", d.Id(), err)
		}
	}
	return nil
}

func updateDevicePolicyConfiguration(conn *worklink.WorkLink, d *schema.ResourceData) error {
	input := &worklink.UpdateDevicePolicyConfigurationInput{
		FleetArn: aws.String(d.Id()),
	}
	if v, ok := d.GetOk("device_ca_certificate"); ok {
		input.DeviceCaCertificate = aws.String(v.(string))
	} else if d.IsNewResource() {
		return nil
	}

	if _, err := conn.UpdateDevicePolicyConfiguration(input); err != nil {
		return fmt.Errorf("error updating WorkLink Fleet (%s) Device Policy Configuration: %w", d.Id(), err)
	}
	return nil
}

func updateIdentityProviderConfiguration(conn *worklink.WorkLink, d *schema.ResourceData) error {
	oldIdentityProvider, newIdentityProvider := d.GetChange("identity_provider")

	if len(oldIdentityProvider.([]interface{})) > 0 && len(newIdentityProvider.([]interface{})) == 0 {
		return fmt.Errorf("Identity Provider Configuration cannot be removed from WorkLink Fleet(%s),"+
			" use 'terraform taint' to recreate the resource if you wish.", d.Id())
	}

	if v, ok := d.GetOk("identity_provider"); ok && len(v.([]interface{})) > 0 {
		config := v.([]interface{})[0].(map[string]interface{})
		input := &worklink.UpdateIdentityProviderConfigurationInput{
			FleetArn:                     aws.String(d.Id()),
			IdentityProviderType:         aws.String(config["type"].(string)),
			IdentityProviderSamlMetadata: aws.String(config["saml_metadata"].(string)),
		}
		if _, err := conn.UpdateIdentityProviderConfiguration(input); err != nil {
			return fmt.Errorf("error updating WorkLink Fleet (%s) Identity Provider Configuration: %w", d.Id(), err)
		}
	}

	return nil
}
