package storagegateway

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/experimental/nullable"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceGateway() *schema.Resource {
	return &schema.Resource{
		Create: resourceGatewayCreate,
		Read:   resourceGatewayRead,
		Update: resourceGatewayUpdate,
		Delete: resourceGatewayDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"activation_key": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"activation_key", "gateway_ip_address"},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"average_download_rate_limit_in_bits_per_sec": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(102400),
			},
			"average_upload_rate_limit_in_bits_per_sec": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(51200),
			},
			"cloudwatch_log_group_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"ec2_instance_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"gateway_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"gateway_ip_address": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsIPv4Address,
				ExactlyOneOf: []string{"activation_key", "gateway_ip_address"},
			},
			"gateway_name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringMatch(regexp.MustCompile(`^[ -\.0-\[\]-~]*[!-\.0-\[\]-~][ -\.0-\[\]-~]*$`), ""),
					validation.StringLenBetween(2, 255),
				),
			},
			"gateway_network_interface": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ipv4_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"gateway_timezone": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.Any(
					validation.StringMatch(regexp.MustCompile(`^GMT[+-][0-9]{1,2}:[0-9]{2}$`), ""),
					validation.StringMatch(regexp.MustCompile(`^GMT$`), ""),
				),
			},
			"gateway_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      gatewayTypeStored,
				ValidateFunc: validation.StringInSlice(gatewayType_Values(), false),
			},
			"gateway_vpc_endpoint": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"host_environment": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"maintenance_start_time": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"day_of_week": {
							Type:         nullable.TypeNullableInt,
							Optional:     true,
							ValidateFunc: nullable.ValidateTypeStringNullableIntBetween(0, 6),
						},
						"day_of_month": {
							Type:         nullable.TypeNullableInt,
							Optional:     true,
							ValidateFunc: nullable.ValidateTypeStringNullableIntBetween(1, 28),
						},
						"hour_of_day": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 23),
						},
						"minute_of_hour": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 59),
						},
					},
				},
			},
			"medium_changer_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(mediumChangerType_Values(), false),
			},
			"smb_active_directory_settings": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"active_directory_status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"domain_controllers": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateFunc: validation.All(
									validation.StringMatch(regexp.MustCompile(`^(([a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9\-]*[A-Za-z0-9])(:(\d+))?$`), ""),
									validation.StringLenBetween(6, 1024),
								),
							},
						},
						"domain_name": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringMatch(regexp.MustCompile(`^([a-z0-9]+(-[a-z0-9]+)*\.)+[a-z]{2,}$`), ""),
								validation.StringLenBetween(1, 1024),
							),
						},
						"organizational_unit": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 1024),
						},
						"password": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
							ValidateFunc: validation.All(
								validation.StringMatch(regexp.MustCompile(`^[ -~]+$`), ""),
								validation.StringLenBetween(1, 1024),
							),
						},
						"timeout_in_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      20,
							ValidateFunc: validation.IntBetween(0, 3600),
						},
						"username": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringMatch(regexp.MustCompile(`^\w[\w\.\- ]*$`), ""),
								validation.StringLenBetween(1, 1024),
							),
						},
					},
				},
			},
			"smb_file_share_visibility": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"smb_guest_password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
				ValidateFunc: validation.All(
					validation.StringMatch(regexp.MustCompile(`^[ -~]+$`), ""),
					validation.StringLenBetween(6, 512),
				),
			},
			"smb_security_strategy": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(storagegateway.SMBSecurityStrategy_Values(), false),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"tape_drive_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(tapeDriveType_Values(), false),
			},
		},

		CustomizeDiff: customdiff.Sequence(
			customdiff.ForceNewIfChange("smb_active_directory_settings", func(_ context.Context, old, new, meta interface{}) bool {
				return len(old.([]interface{})) == 1 && len(new.([]interface{})) == 0
			}),
			verify.SetTagsDiff,
		),
	}
}

func resourceGatewayCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).StorageGatewayConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
	region := meta.(*conns.AWSClient).Region

	activationKey := d.Get("activation_key").(string)

	// Perform one time fetch of activation key from gateway IP address.
	if v, ok := d.GetOk("gateway_ip_address"); ok {
		gatewayIPAddress := v.(string)

		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Timeout: time.Second * 10,
		}

		requestURL := fmt.Sprintf("http://%s/?activationRegion=%s", gatewayIPAddress, region)
		if v, ok := d.GetOk("gateway_vpc_endpoint"); ok {
			requestURL = fmt.Sprintf("%s&vpcEndpoint=%s", requestURL, v.(string))
		}

		log.Printf("[DEBUG] Creating HTTP request: %s", requestURL)
		request, err := http.NewRequest("GET", requestURL, nil)
		if err != nil {
			return fmt.Errorf("error creating HTTP request: %w", err)
		}

		var response *http.Response
		err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
			log.Printf("[DEBUG] Making HTTP request: %s", request.URL.String())
			response, err = client.Do(request)

			if err != nil {
				if err, ok := err.(net.Error); ok {
					errMessage := fmt.Errorf("error making HTTP request: %s", err)
					log.Printf("[DEBUG] retryable %s", errMessage)
					return resource.RetryableError(errMessage)
				}

				return resource.NonRetryableError(fmt.Errorf("error making HTTP request: %w", err))
			}

			for _, retryableStatusCode := range []int{504} {
				if response.StatusCode == retryableStatusCode {
					errMessage := fmt.Errorf("status code in HTTP response: %d", response.StatusCode)
					log.Printf("[DEBUG] retryable %s", errMessage)
					return resource.RetryableError(errMessage)
				}
			}

			return nil
		})
		if tfresource.TimedOut(err) {
			response, err = client.Do(request)
		}
		if err != nil {
			return fmt.Errorf("error retrieving activation key from IP Address (%s): %w", gatewayIPAddress, err)
		}

		log.Printf("[DEBUG] Received HTTP response: %#v", response)
		if response.StatusCode != 302 {
			return fmt.Errorf("expected HTTP status code 302, received: %d", response.StatusCode)
		}

		redirectURL, err := response.Location()
		if err != nil {
			return fmt.Errorf("error extracting HTTP Location header: %w", err)
		}

		activationKey = redirectURL.Query().Get("activationKey")
		if activationKey == "" {
			return fmt.Errorf("empty activationKey received from IP Address: %s", gatewayIPAddress)
		}
	}

	input := &storagegateway.ActivateGatewayInput{
		ActivationKey:   aws.String(activationKey),
		GatewayRegion:   aws.String(region),
		GatewayName:     aws.String(d.Get("gateway_name").(string)),
		GatewayTimezone: aws.String(d.Get("gateway_timezone").(string)),
		GatewayType:     aws.String(d.Get("gateway_type").(string)),
		Tags:            Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("medium_changer_type"); ok {
		input.MediumChangerType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tape_drive_type"); ok {
		input.TapeDriveType = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Activating Storage Gateway Gateway: %s", input)
	output, err := conn.ActivateGateway(input)
	if err != nil {
		return fmt.Errorf("error activating Storage Gateway Gateway: %w", err)
	}

	d.SetId(aws.StringValue(output.GatewayARN))
	log.Printf("[INFO] Storage Gateway Gateway ID: %s", d.Id())

	log.Printf("[DEBUG] Waiting for Storage Gateway Gateway (%s) to be connected", d.Id())

	if _, err = waitGatewayConnected(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for Storage Gateway Gateway (%q) to be Connected: %w", d.Id(), err)
	}

	if v, ok := d.GetOk("cloudwatch_log_group_arn"); ok && v.(string) != "" {
		input := &storagegateway.UpdateGatewayInformationInput{
			GatewayARN:            aws.String(d.Id()),
			CloudWatchLogGroupARN: aws.String(v.(string)),
		}

		log.Printf("[DEBUG] Storage Gateway Gateway %q setting CloudWatch Log Group", input)
		_, err := conn.UpdateGatewayInformation(input)
		if err != nil {
			return fmt.Errorf("error setting CloudWatch Log Group: %w", err)
		}
	}

	if v, ok := d.GetOk("maintenance_start_time"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input := expandUpdateMaintenanceStartTimeInput(v.([]interface{})[0].(map[string]interface{}))
		input.GatewayARN = aws.String(d.Id())

		log.Printf("[DEBUG] Storage Gateway Gateway %q updating maintenance start time", d.Id())
		_, err := conn.UpdateMaintenanceStartTime(input)

		if err != nil {
			return fmt.Errorf("error updating maintenance start time: %w", err)
		}
	}

	if v, ok := d.GetOk("smb_active_directory_settings"); ok && len(v.([]interface{})) > 0 {
		input := expandGatewayDomain(v.([]interface{}), d.Id())
		log.Printf("[DEBUG] Storage Gateway Gateway %q joining Active Directory domain: %s", d.Id(), aws.StringValue(input.DomainName))
		_, err := conn.JoinDomain(input)
		if err != nil {
			return fmt.Errorf("error joining Active Directory domain: %w", err)
		}
		log.Printf("[DEBUG] Waiting for Storage Gateway Gateway (%s) to be connected", d.Id())
		if _, err = waitGatewayJoinDomainJoined(conn, d.Id()); err != nil {
			return fmt.Errorf("error waiting for Storage Gateway Gateway (%q) to join domain (%s): %w", d.Id(), aws.StringValue(input.DomainName), err)
		}
	}

	if v, ok := d.GetOk("smb_guest_password"); ok && v.(string) != "" {
		input := &storagegateway.SetSMBGuestPasswordInput{
			GatewayARN: aws.String(d.Id()),
			Password:   aws.String(v.(string)),
		}

		log.Printf("[DEBUG] Storage Gateway Gateway %q setting SMB guest password", d.Id())
		_, err := conn.SetSMBGuestPassword(input)
		if err != nil {
			return fmt.Errorf("error setting SMB guest password: %w", err)
		}
	}

	if v, ok := d.GetOk("smb_security_strategy"); ok {
		input := &storagegateway.UpdateSMBSecurityStrategyInput{
			GatewayARN:          aws.String(d.Id()),
			SMBSecurityStrategy: aws.String(v.(string)),
		}

		log.Printf("[DEBUG] Storage Gateway Gateway %q setting SMB Security Strategy", input)
		_, err := conn.UpdateSMBSecurityStrategy(input)
		if err != nil {
			return fmt.Errorf("error setting SMB Security Strategy: %w", err)
		}
	}

	if v, ok := d.GetOk("smb_file_share_visibility"); ok {
		input := &storagegateway.UpdateSMBFileShareVisibilityInput{
			GatewayARN:        aws.String(d.Id()),
			FileSharesVisible: aws.Bool(v.(bool)),
		}

		log.Printf("[DEBUG] Storage Gateway Gateway %q setting SMB File Share Visibility", input)
		_, err := conn.UpdateSMBFileShareVisibility(input)
		if err != nil {
			return fmt.Errorf("error updating Storage Gateway Gateway (%s) SMB file share visibility: %w", d.Id(), err)
		}
	}

	bandwidthInput := &storagegateway.UpdateBandwidthRateLimitInput{
		GatewayARN: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("average_download_rate_limit_in_bits_per_sec"); ok {
		bandwidthInput.AverageDownloadRateLimitInBitsPerSec = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("average_upload_rate_limit_in_bits_per_sec"); ok {
		bandwidthInput.AverageUploadRateLimitInBitsPerSec = aws.Int64(int64(v.(int)))
	}

	if bandwidthInput.AverageDownloadRateLimitInBitsPerSec != nil || bandwidthInput.AverageUploadRateLimitInBitsPerSec != nil {
		log.Printf("[DEBUG] Storage Gateway Gateway %q setting Bandwidth Rate Limit: %#v", d.Id(), bandwidthInput)
		_, err := conn.UpdateBandwidthRateLimit(bandwidthInput)
		if err != nil {
			return fmt.Errorf("error setting Bandwidth Rate Limit: %s", err)
		}
	}

	return resourceGatewayRead(d, meta)
}

func resourceGatewayRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).StorageGatewayConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	output, err := FindGatewayByARN(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Storage Gateway Gateway (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Storage Gateway Gateway (%s): %w", d.Id(), err)
	}

	tags := KeyValueTags(output.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	smbSettingsInput := &storagegateway.DescribeSMBSettingsInput{
		GatewayARN: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading Storage Gateway SMB Settings: %s", smbSettingsInput)
	smbSettingsOutput, err := conn.DescribeSMBSettings(smbSettingsInput)
	if err != nil && !tfawserr.ErrMessageContains(err, storagegateway.ErrCodeInvalidGatewayRequestException, "This operation is not valid for the specified gateway") {
		if IsErrGatewayNotFound(err) {
			log.Printf("[WARN] Storage Gateway Gateway %q not found - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading Storage Gateway SMB Settings: %w", err)
	}

	// The Storage Gateway API currently provides no way to read this value
	// We allow Terraform to passthrough the configuration value into the state
	d.Set("activation_key", d.Get("activation_key").(string))

	d.Set("arn", output.GatewayARN)
	d.Set("gateway_id", output.GatewayId)

	// The Storage Gateway API currently provides no way to read this value
	// We allow Terraform to passthrough the configuration value into the state
	d.Set("gateway_ip_address", d.Get("gateway_ip_address").(string))

	d.Set("gateway_name", output.GatewayName)
	d.Set("gateway_timezone", output.GatewayTimezone)
	d.Set("gateway_type", output.GatewayType)
	d.Set("gateway_vpc_endpoint", output.VPCEndpoint)

	// The Storage Gateway API currently provides no way to read this value
	// We allow Terraform to passthrough the configuration value into the state
	d.Set("medium_changer_type", d.Get("medium_changer_type").(string))

	// Treat the entire nested argument as a whole, based on domain name
	// to simplify schema and difference logic
	if smbSettingsOutput == nil || aws.StringValue(smbSettingsOutput.DomainName) == "" {
		if err := d.Set("smb_active_directory_settings", []interface{}{}); err != nil {
			return fmt.Errorf("error setting smb_active_directory_settings: %w", err)
		}
	} else {
		m := map[string]interface{}{
			"domain_name":             aws.StringValue(smbSettingsOutput.DomainName),
			"active_directory_status": aws.StringValue(smbSettingsOutput.ActiveDirectoryStatus),
			// The Storage Gateway API currently provides no way to read these values
			// "password": ...,
			// "username": ...,
		}
		// We must assemble these into the map from configuration or Terraform will enter ""
		// into state and constantly show a difference (also breaking downstream references)
		//  UPDATE: aws_storagegateway_gateway.test
		//    smb_active_directory_settings.0.password: "<sensitive>" => "<sensitive>" (attribute changed)
		//    smb_active_directory_settings.0.username: "" => "Administrator"
		if v, ok := d.GetOk("smb_active_directory_settings"); ok && len(v.([]interface{})) > 0 {
			configM := v.([]interface{})[0].(map[string]interface{})
			m["password"] = configM["password"]
			m["username"] = configM["username"]
			m["timeout_in_seconds"] = configM["timeout_in_seconds"]

			if v, ok := configM["organizational_unit"]; ok {
				m["organizational_unit"] = v
			}

			if v, ok := configM["domain_controllers"]; ok {
				m["domain_controllers"] = v
			}
		}
		if err := d.Set("smb_active_directory_settings", []map[string]interface{}{m}); err != nil {
			return fmt.Errorf("error setting smb_active_directory_settings: %w", err)
		}
	}

	// The Storage Gateway API currently provides no way to read this value
	// We allow Terraform to _automatically_ passthrough the configuration value into the state here
	// as the API does clue us in whether or not its actually set at all,
	// which can be used to tell Terraform to show a difference in this case
	// as well as ensuring there is some sort of attribute value (unlike the others)
	if smbSettingsOutput == nil || !aws.BoolValue(smbSettingsOutput.SMBGuestPasswordSet) {
		d.Set("smb_guest_password", "")
	}

	// The Storage Gateway API currently provides no way to read this value
	// We allow Terraform to passthrough the configuration value into the state
	d.Set("tape_drive_type", d.Get("tape_drive_type").(string))
	d.Set("cloudwatch_log_group_arn", output.CloudWatchLogGroupARN)
	d.Set("smb_security_strategy", smbSettingsOutput.SMBSecurityStrategy)
	d.Set("smb_file_share_visibility", smbSettingsOutput.FileSharesVisible)
	d.Set("ec2_instance_id", output.Ec2InstanceId)
	d.Set("endpoint_type", output.EndpointType)
	d.Set("host_environment", output.HostEnvironment)

	if err := d.Set("gateway_network_interface", flattenGatewayNetworkInterfaces(output.GatewayNetworkInterfaces)); err != nil {
		return fmt.Errorf("error setting gateway_network_interface: %w", err)
	}

	bandwidthOutput, err := conn.DescribeBandwidthRateLimit(&storagegateway.DescribeBandwidthRateLimitInput{
		GatewayARN: aws.String(d.Id()),
	})

	if tfawserr.ErrMessageContains(err, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified operation is not supported") ||
		tfawserr.ErrMessageContains(err, storagegateway.ErrCodeInvalidGatewayRequestException, "This operation is not valid for the specified gateway") {
		err = nil
	}

	if err != nil {
		return fmt.Errorf("error reading Storage Gateway Bandwidth rate limit: %w", err)
	}

	if bandwidthOutput != nil {
		d.Set("average_download_rate_limit_in_bits_per_sec", bandwidthOutput.AverageDownloadRateLimitInBitsPerSec)
		d.Set("average_upload_rate_limit_in_bits_per_sec", bandwidthOutput.AverageUploadRateLimitInBitsPerSec)
	}

	maintenanceStartTimeOutput, err := conn.DescribeMaintenanceStartTime(&storagegateway.DescribeMaintenanceStartTimeInput{
		GatewayARN: aws.String(d.Id()),
	})

	if tfawserr.ErrMessageContains(err, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified operation is not supported") ||
		tfawserr.ErrMessageContains(err, storagegateway.ErrCodeInvalidGatewayRequestException, "This operation is not valid for the specified gateway") {
		err = nil
	}

	if err != nil {
		return fmt.Errorf("error reading Storage Gateway maintenance start time: %w", err)
	}

	if maintenanceStartTimeOutput != nil {
		if err := d.Set("maintenance_start_time", []map[string]interface{}{flattenDescribeMaintenanceStartTimeOutput(maintenanceStartTimeOutput)}); err != nil {
			return fmt.Errorf("error setting maintenance_start_time: %w", err)
		}
	} else {
		d.Set("maintenance_start_time", nil)
	}

	return nil
}

func resourceGatewayUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).StorageGatewayConn

	if d.HasChanges("gateway_name", "gateway_timezone", "cloudwatch_log_group_arn") {
		input := &storagegateway.UpdateGatewayInformationInput{
			CloudWatchLogGroupARN: aws.String(d.Get("cloudwatch_log_group_arn").(string)),
			GatewayARN:            aws.String(d.Id()),
			GatewayName:           aws.String(d.Get("gateway_name").(string)),
			GatewayTimezone:       aws.String(d.Get("gateway_timezone").(string)),
		}

		log.Printf("[DEBUG] Updating Storage Gateway Gateway: %s", input)
		_, err := conn.UpdateGatewayInformation(input)

		if err != nil {
			return fmt.Errorf("error updating Storage Gateway Gateway (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("maintenance_start_time") {
		if v, ok := d.GetOk("maintenance_start_time"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input := expandUpdateMaintenanceStartTimeInput(v.([]interface{})[0].(map[string]interface{}))
			input.GatewayARN = aws.String(d.Id())

			log.Printf("[DEBUG] Updating Storage Gateway maintenance start time: %s", input)
			_, err := conn.UpdateMaintenanceStartTime(input)

			if err != nil {
				return fmt.Errorf("error updating Storage Gateway Gateway (%s) maintenance start time: %w", d.Id(), err)
			}
		}
	}

	if d.HasChange("smb_active_directory_settings") {
		input := expandGatewayDomain(d.Get("smb_active_directory_settings").([]interface{}), d.Id())
		domainName := aws.StringValue(input.DomainName)

		log.Printf("[DEBUG] Joining Storage Gateway to Active Directory domain: %s", input)
		_, err := conn.JoinDomain(input)

		if err != nil {
			return fmt.Errorf("error joining Storage Gateway Gateway (%s) to Active Directory domain (%s): %w", d.Id(), domainName, err)
		}

		if _, err = waitGatewayJoinDomainJoined(conn, d.Id()); err != nil {
			return fmt.Errorf("error waiting for Storage Gateway Gateway (%s) to join Active Directory domain (%s): %w", d.Id(), domainName, err)
		}
	}

	if d.HasChange("smb_guest_password") {
		input := &storagegateway.SetSMBGuestPasswordInput{
			GatewayARN: aws.String(d.Id()),
			Password:   aws.String(d.Get("smb_guest_password").(string)),
		}

		log.Printf("[DEBUG] Setting Storage Gateway SMB guest password: %s", input)
		_, err := conn.SetSMBGuestPassword(input)

		if err != nil {
			return fmt.Errorf("error updating Storage Gateway Gateway (%s) SMB guest password: %w", d.Id(), err)
		}
	}

	if d.HasChange("smb_security_strategy") {
		input := &storagegateway.UpdateSMBSecurityStrategyInput{
			GatewayARN:          aws.String(d.Id()),
			SMBSecurityStrategy: aws.String(d.Get("smb_security_strategy").(string)),
		}

		log.Printf("[DEBUG] Updating Storage Gateway SMB security strategy: %s", input)
		_, err := conn.UpdateSMBSecurityStrategy(input)

		if err != nil {
			return fmt.Errorf("error updating Storage Gateway Gateway (%s) SMB security strategy: %w", d.Id(), err)
		}
	}

	if d.HasChange("smb_file_share_visibility") {
		input := &storagegateway.UpdateSMBFileShareVisibilityInput{
			FileSharesVisible: aws.Bool(d.Get("smb_file_share_visibility").(bool)),
			GatewayARN:        aws.String(d.Id()),
		}

		log.Printf("[DEBUG] Updating Storage Gateway SMB file share visibility: %s", input)
		_, err := conn.UpdateSMBFileShareVisibility(input)

		if err != nil {
			return fmt.Errorf("error updating Storage Gateway Gateway (%s) SMB file share visibility: %w", d.Id(), err)
		}
	}

	if d.HasChanges("average_download_rate_limit_in_bits_per_sec", "average_upload_rate_limit_in_bits_per_sec") {
		deleteInput := &storagegateway.DeleteBandwidthRateLimitInput{
			GatewayARN: aws.String(d.Id()),
		}
		updateInput := &storagegateway.UpdateBandwidthRateLimitInput{
			GatewayARN: aws.String(d.Id()),
		}
		needsDelete := false
		needsUpdate := false

		if v, ok := d.GetOk("average_download_rate_limit_in_bits_per_sec"); ok {
			updateInput.AverageDownloadRateLimitInBitsPerSec = aws.Int64(int64(v.(int)))
			needsUpdate = true
		} else if d.HasChange("average_download_rate_limit_in_bits_per_sec") {
			deleteInput.BandwidthType = aws.String(bandwidthTypeDownload)
			needsDelete = true
		}

		if v, ok := d.GetOk("average_upload_rate_limit_in_bits_per_sec"); ok {
			updateInput.AverageUploadRateLimitInBitsPerSec = aws.Int64(int64(v.(int)))
			needsUpdate = true
		} else if d.HasChange("average_upload_rate_limit_in_bits_per_sec") {
			if needsDelete {
				deleteInput.BandwidthType = aws.String(bandwidthTypeAll)
			} else {
				deleteInput.BandwidthType = aws.String(bandwidthTypeUpload)
				needsDelete = true
			}
		}

		if needsUpdate {
			log.Printf("[DEBUG] Updating Storage Gateway bandwidth rate limit: %s", updateInput)
			_, err := conn.UpdateBandwidthRateLimit(updateInput)

			if err != nil {
				return fmt.Errorf("error updating Storage Gateway Gateway (%s) bandwidth rate limit: %w", d.Id(), err)
			}
		}

		if needsDelete {
			log.Printf("[DEBUG] Deleting Storage Gateway bandwidth rate limit: %s", deleteInput)
			_, err := conn.DeleteBandwidthRateLimit(deleteInput)

			if err != nil {
				return fmt.Errorf("error deleting Storage Gateway Gateway (%s) bandwidth rate limit: %w", d.Id(), err)
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	return resourceGatewayRead(d, meta)
}

func resourceGatewayDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).StorageGatewayConn

	log.Printf("[DEBUG] Deleting Storage Gateway Gateway: %s", d.Id())
	_, err := conn.DeleteGateway(&storagegateway.DeleteGatewayInput{
		GatewayARN: aws.String(d.Id()),
	})

	if operationErrorCode(err) == operationErrCodeGatewayNotFound || tfawserr.ErrCodeEquals(err, storagegateway.ErrorCodeGatewayNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Storage Gateway Gateway (%s): %w", d.Id(), err)
	}

	return nil
}

func expandGatewayDomain(l []interface{}, gatewayArn string) *storagegateway.JoinDomainInput {
	if l == nil || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	domain := &storagegateway.JoinDomainInput{
		DomainName:       aws.String(tfMap["domain_name"].(string)),
		GatewayARN:       aws.String(gatewayArn),
		Password:         aws.String(tfMap["password"].(string)),
		UserName:         aws.String(tfMap["username"].(string)),
		TimeoutInSeconds: aws.Int64(int64(tfMap["timeout_in_seconds"].(int))),
	}

	if v, ok := tfMap["organizational_unit"].(string); ok && v != "" {
		domain.OrganizationalUnit = aws.String(v)
	}

	if v, ok := tfMap["domain_controllers"].(*schema.Set); ok && v.Len() > 0 {
		domain.DomainControllers = flex.ExpandStringSet(v)
	}

	return domain
}

func flattenGatewayNetworkInterfaces(nis []*storagegateway.NetworkInterface) []interface{} {
	if len(nis) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, ni := range nis {
		if ni == nil {
			continue
		}

		tfMap := map[string]interface{}{
			"ipv4_address": aws.StringValue(ni.Ipv4Address),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func expandUpdateMaintenanceStartTimeInput(tfMap map[string]interface{}) *storagegateway.UpdateMaintenanceStartTimeInput {
	if tfMap == nil {
		return nil
	}

	apiObject := &storagegateway.UpdateMaintenanceStartTimeInput{}

	if v, null, _ := nullable.Int(tfMap["day_of_month"].(string)).Value(); !null && v > 0 {
		apiObject.DayOfMonth = aws.Int64(v)
	}

	if v, null, _ := nullable.Int(tfMap["day_of_week"].(string)).Value(); !null && v > 0 {
		apiObject.DayOfWeek = aws.Int64(v)
	}

	if v, ok := tfMap["hour_of_day"].(int); ok {
		apiObject.HourOfDay = aws.Int64(int64(v))
	}

	if v, ok := tfMap["minute_of_hour"].(int); ok {
		apiObject.MinuteOfHour = aws.Int64(int64(v))
	}

	return apiObject
}

func flattenDescribeMaintenanceStartTimeOutput(apiObject *storagegateway.DescribeMaintenanceStartTimeOutput) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.DayOfMonth; v != nil {
		tfMap["day_of_month"] = strconv.FormatInt(aws.Int64Value(v), 10)
	}

	if v := apiObject.DayOfWeek; v != nil {
		tfMap["day_of_week"] = strconv.FormatInt(aws.Int64Value(v), 10)
	}

	if v := apiObject.HourOfDay; v != nil {
		tfMap["hour_of_day"] = aws.Int64Value(v)
	}

	if v := apiObject.MinuteOfHour; v != nil {
		tfMap["minute_of_hour"] = aws.Int64Value(v)
	}

	return tfMap
}

// The API returns multiple responses for a missing gateway
func IsErrGatewayNotFound(err error) bool {
	if tfawserr.ErrMessageContains(err, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified gateway was not found.") {
		return true
	}
	if tfawserr.ErrCodeEquals(err, storagegateway.ErrorCodeGatewayNotFound) {
		return true
	}
	return false
}
