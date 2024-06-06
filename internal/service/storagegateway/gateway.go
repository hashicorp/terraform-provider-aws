// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package storagegateway

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2/types/nullable"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_storagegateway_gateway", name="Gateway")
// @Tags(identifierAttribute="arn")
func resourceGateway() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGatewayCreate,
		ReadWithoutTimeout:   resourceGatewayRead,
		UpdateWithoutTimeout: resourceGatewayUpdate,
		DeleteWithoutTimeout: resourceGatewayDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
			names.AttrARN: {
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
			names.AttrCloudWatchLogGroupARN: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"ec2_instance_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEndpointType: {
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
					validation.StringMatch(regexache.MustCompile(`^[ -\.0-\[\]-~]*[!-\.0-\[\]-~][ -\.0-\[\]-~]*$`), ""),
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
					validation.StringMatch(regexache.MustCompile(`^GMT[+-][0-9]{1,2}:[0-9]{2}$`), ""),
					validation.StringMatch(regexache.MustCompile(`^GMT$`), ""),
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
									validation.StringMatch(regexache.MustCompile(`^(([0-9A-Za-z-]*[0-9A-Za-z])\.)*([0-9A-Za-z-]*[0-9A-Za-z])(:(\d+))?$`), ""),
									validation.StringLenBetween(6, 1024),
								),
							},
						},
						names.AttrDomainName: {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringMatch(regexache.MustCompile(`^([0-9a-z]+(-[0-9a-z]+)*\.)+[a-z]{2,}$`), ""),
								validation.StringLenBetween(1, 1024),
							),
						},
						"organizational_unit": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 1024),
						},
						names.AttrPassword: {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
							ValidateFunc: validation.All(
								validation.StringMatch(regexache.MustCompile(`^[ -~]+$`), ""),
								validation.StringLenBetween(1, 1024),
							),
						},
						"timeout_in_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      20,
							ValidateFunc: validation.IntBetween(0, 3600),
						},
						names.AttrUsername: {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringMatch(regexache.MustCompile(`^\w[\w\.\- ]*$`), ""),
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
					validation.StringMatch(regexache.MustCompile(`^[ -~]+$`), ""),
					validation.StringLenBetween(6, 512),
				),
			},
			"smb_security_strategy": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(storagegateway.SMBSecurityStrategy_Values(), false),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
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

func resourceGatewayCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayConn(ctx)

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
		request, err := http.NewRequest(http.MethodGet, requestURL, nil)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating HTTP request: %s", err)
		}

		var response *http.Response
		err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
			response, err = client.Do(request)

			if err != nil {
				if errs.IsA[net.Error](err) {
					errMessage := fmt.Errorf("making HTTP request: %s", err)
					log.Printf("[DEBUG] retryable %s", errMessage)
					return retry.RetryableError(errMessage)
				}

				return retry.NonRetryableError(fmt.Errorf("making HTTP request: %w", err))
			}

			for _, retryableStatusCode := range []int{504} {
				if response.StatusCode == retryableStatusCode {
					errMessage := fmt.Errorf("status code in HTTP response: %d", response.StatusCode)
					log.Printf("[DEBUG] retryable %s", errMessage)
					return retry.RetryableError(errMessage)
				}
			}

			return nil
		})
		if tfresource.TimedOut(err) {
			response, err = client.Do(request)
		}
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "retrieving activation key from IP Address (%s): %s", gatewayIPAddress, err)
		}

		log.Printf("[DEBUG] Received HTTP response: %#v", response)
		if response.StatusCode != http.StatusFound {
			return sdkdiag.AppendErrorf(diags, "expected HTTP status code 302, received: %d", response.StatusCode)
		}

		redirectURL, err := response.Location()
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "extracting HTTP Location header: %s", err)
		}

		activationKey = redirectURL.Query().Get("activationKey")
		if activationKey == "" {
			return sdkdiag.AppendErrorf(diags, "empty activationKey received from IP Address: %s", gatewayIPAddress)
		}
	}

	input := &storagegateway.ActivateGatewayInput{
		ActivationKey:   aws.String(activationKey),
		GatewayRegion:   aws.String(region),
		GatewayName:     aws.String(d.Get("gateway_name").(string)),
		GatewayTimezone: aws.String(d.Get("gateway_timezone").(string)),
		GatewayType:     aws.String(d.Get("gateway_type").(string)),
		Tags:            getTagsIn(ctx),
	}

	if v, ok := d.GetOk("medium_changer_type"); ok {
		input.MediumChangerType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tape_drive_type"); ok {
		input.TapeDriveType = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Activating Storage Gateway Gateway: %s", input)
	output, err := conn.ActivateGatewayWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "activating Storage Gateway Gateway: %s", err)
	}

	d.SetId(aws.StringValue(output.GatewayARN))
	log.Printf("[INFO] Storage Gateway Gateway ID: %s", d.Id())

	log.Printf("[DEBUG] Waiting for Storage Gateway Gateway (%s) to be connected", d.Id())

	if _, err = waitGatewayConnected(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Storage Gateway Gateway (%q) to be Connected: %s", d.Id(), err)
	}

	if v, ok := d.GetOk(names.AttrCloudWatchLogGroupARN); ok && v.(string) != "" {
		input := &storagegateway.UpdateGatewayInformationInput{
			GatewayARN:            aws.String(d.Id()),
			CloudWatchLogGroupARN: aws.String(v.(string)),
		}

		log.Printf("[DEBUG] Storage Gateway Gateway %q setting CloudWatch Log Group", input)
		_, err := conn.UpdateGatewayInformationWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting CloudWatch Log Group: %s", err)
		}
	}

	if v, ok := d.GetOk("maintenance_start_time"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input := expandUpdateMaintenanceStartTimeInput(v.([]interface{})[0].(map[string]interface{}))
		input.GatewayARN = aws.String(d.Id())

		log.Printf("[DEBUG] Storage Gateway Gateway %q updating maintenance start time", d.Id())
		_, err := conn.UpdateMaintenanceStartTimeWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating maintenance start time: %s", err)
		}
	}

	if v, ok := d.GetOk("smb_active_directory_settings"); ok && len(v.([]interface{})) > 0 {
		input := expandGatewayDomain(v.([]interface{}), d.Id())
		log.Printf("[DEBUG] Storage Gateway Gateway %q joining Active Directory domain: %s", d.Id(), aws.StringValue(input.DomainName))
		_, err := conn.JoinDomainWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "joining Active Directory domain: %s", err)
		}
		log.Printf("[DEBUG] Waiting for Storage Gateway Gateway (%s) to be connected", d.Id())
		if _, err = waitGatewayJoinDomainJoined(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Storage Gateway Gateway (%q) to join domain (%s): %s", d.Id(), aws.StringValue(input.DomainName), err)
		}
	}

	if v, ok := d.GetOk("smb_guest_password"); ok && v.(string) != "" {
		input := &storagegateway.SetSMBGuestPasswordInput{
			GatewayARN: aws.String(d.Id()),
			Password:   aws.String(v.(string)),
		}

		log.Printf("[DEBUG] Storage Gateway Gateway %q setting SMB guest password", d.Id())
		_, err := conn.SetSMBGuestPasswordWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting SMB guest password: %s", err)
		}
	}

	if v, ok := d.GetOk("smb_security_strategy"); ok {
		input := &storagegateway.UpdateSMBSecurityStrategyInput{
			GatewayARN:          aws.String(d.Id()),
			SMBSecurityStrategy: aws.String(v.(string)),
		}

		log.Printf("[DEBUG] Storage Gateway Gateway %q setting SMB Security Strategy", input)
		_, err := conn.UpdateSMBSecurityStrategyWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting SMB Security Strategy: %s", err)
		}
	}

	if v, ok := d.GetOk("smb_file_share_visibility"); ok {
		input := &storagegateway.UpdateSMBFileShareVisibilityInput{
			GatewayARN:        aws.String(d.Id()),
			FileSharesVisible: aws.Bool(v.(bool)),
		}

		log.Printf("[DEBUG] Storage Gateway Gateway %q setting SMB File Share Visibility", input)
		_, err := conn.UpdateSMBFileShareVisibilityWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Storage Gateway Gateway (%s) SMB file share visibility: %s", d.Id(), err)
		}
	}

	switch d.Get("gateway_type").(string) {
	case gatewayTypeCached, gatewayTypeStored, gatewayTypeVTL, gatewayTypeVTLSnow:
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
			_, err := conn.UpdateBandwidthRateLimitWithContext(ctx, bandwidthInput)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "setting Bandwidth Rate Limit: %s", err)
			}
		}
	}

	return append(diags, resourceGatewayRead(ctx, d, meta)...)
}

func resourceGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayConn(ctx)

	output, err := FindGatewayByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Storage Gateway Gateway (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Storage Gateway Gateway (%s): %s", d.Id(), err)
	}

	setTagsOut(ctx, output.Tags)

	smbSettingsInput := &storagegateway.DescribeSMBSettingsInput{
		GatewayARN: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading Storage Gateway SMB Settings: %s", smbSettingsInput)
	smbSettingsOutput, err := conn.DescribeSMBSettingsWithContext(ctx, smbSettingsInput)
	if err != nil && !tfawserr.ErrMessageContains(err, storagegateway.ErrCodeInvalidGatewayRequestException, "This operation is not valid for the specified gateway") {
		if IsErrGatewayNotFound(err) {
			log.Printf("[WARN] Storage Gateway Gateway %q not found - removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading Storage Gateway SMB Settings: %s", err)
	}

	// The Storage Gateway API currently provides no way to read this value
	// We allow Terraform to passthrough the configuration value into the state
	d.Set("activation_key", d.Get("activation_key").(string))

	d.Set(names.AttrARN, output.GatewayARN)
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
			return sdkdiag.AppendErrorf(diags, "setting smb_active_directory_settings: %s", err)
		}
	} else {
		m := map[string]interface{}{
			names.AttrDomainName:      aws.StringValue(smbSettingsOutput.DomainName),
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
			m[names.AttrPassword] = configM[names.AttrPassword]
			m[names.AttrUsername] = configM[names.AttrUsername]
			m["timeout_in_seconds"] = configM["timeout_in_seconds"]

			if v, ok := configM["organizational_unit"]; ok {
				m["organizational_unit"] = v
			}

			if v, ok := configM["domain_controllers"]; ok {
				m["domain_controllers"] = v
			}
		}
		if err := d.Set("smb_active_directory_settings", []map[string]interface{}{m}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting smb_active_directory_settings: %s", err)
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
	d.Set(names.AttrCloudWatchLogGroupARN, output.CloudWatchLogGroupARN)
	d.Set("smb_security_strategy", smbSettingsOutput.SMBSecurityStrategy)
	d.Set("smb_file_share_visibility", smbSettingsOutput.FileSharesVisible)
	d.Set("ec2_instance_id", output.Ec2InstanceId)
	d.Set(names.AttrEndpointType, output.EndpointType)
	d.Set("host_environment", output.HostEnvironment)

	if err := d.Set("gateway_network_interface", flattenGatewayNetworkInterfaces(output.GatewayNetworkInterfaces)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting gateway_network_interface: %s", err)
	}

	switch aws.StringValue(output.GatewayType) {
	case gatewayTypeCached, gatewayTypeStored, gatewayTypeVTL, gatewayTypeVTLSnow:
		bandwidthOutput, err := conn.DescribeBandwidthRateLimitWithContext(ctx, &storagegateway.DescribeBandwidthRateLimitInput{
			GatewayARN: aws.String(d.Id()),
		})

		if tfawserr.ErrMessageContains(err, storagegateway.ErrCodeInvalidGatewayRequestException, "not supported") ||
			tfawserr.ErrMessageContains(err, storagegateway.ErrCodeInvalidGatewayRequestException, "not valid") {
			err = nil
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Storage Gateway Bandwidth rate limit: %s", err)
		}

		if bandwidthOutput != nil {
			d.Set("average_download_rate_limit_in_bits_per_sec", bandwidthOutput.AverageDownloadRateLimitInBitsPerSec)
			d.Set("average_upload_rate_limit_in_bits_per_sec", bandwidthOutput.AverageUploadRateLimitInBitsPerSec)
		}
	}

	maintenanceStartTimeOutput, err := conn.DescribeMaintenanceStartTimeWithContext(ctx, &storagegateway.DescribeMaintenanceStartTimeInput{
		GatewayARN: aws.String(d.Id()),
	})

	if tfawserr.ErrMessageContains(err, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified operation is not supported") ||
		tfawserr.ErrMessageContains(err, storagegateway.ErrCodeInvalidGatewayRequestException, "This operation is not valid for the specified gateway") {
		err = nil
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Storage Gateway maintenance start time: %s", err)
	}

	if maintenanceStartTimeOutput != nil {
		if err := d.Set("maintenance_start_time", []map[string]interface{}{flattenDescribeMaintenanceStartTimeOutput(maintenanceStartTimeOutput)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting maintenance_start_time: %s", err)
		}
	} else {
		d.Set("maintenance_start_time", nil)
	}

	return diags
}

func resourceGatewayUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayConn(ctx)

	if d.HasChanges("gateway_name", "gateway_timezone", names.AttrCloudWatchLogGroupARN) {
		input := &storagegateway.UpdateGatewayInformationInput{
			CloudWatchLogGroupARN: aws.String(d.Get(names.AttrCloudWatchLogGroupARN).(string)),
			GatewayARN:            aws.String(d.Id()),
			GatewayName:           aws.String(d.Get("gateway_name").(string)),
			GatewayTimezone:       aws.String(d.Get("gateway_timezone").(string)),
		}

		log.Printf("[DEBUG] Updating Storage Gateway Gateway: %s", input)
		_, err := conn.UpdateGatewayInformationWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Storage Gateway Gateway (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("maintenance_start_time") {
		if v, ok := d.GetOk("maintenance_start_time"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input := expandUpdateMaintenanceStartTimeInput(v.([]interface{})[0].(map[string]interface{}))
			input.GatewayARN = aws.String(d.Id())

			log.Printf("[DEBUG] Updating Storage Gateway maintenance start time: %s", input)
			_, err := conn.UpdateMaintenanceStartTimeWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating Storage Gateway Gateway (%s) maintenance start time: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("smb_active_directory_settings") {
		input := expandGatewayDomain(d.Get("smb_active_directory_settings").([]interface{}), d.Id())
		domainName := aws.StringValue(input.DomainName)

		_, err := conn.JoinDomainWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "joining Storage Gateway Gateway (%s) to Active Directory domain (%s): %s", d.Id(), domainName, err)
		}

		if _, err = waitGatewayJoinDomainJoined(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Storage Gateway Gateway (%s) to join Active Directory domain (%s): %s", d.Id(), domainName, err)
		}
	}

	if d.HasChange("smb_guest_password") {
		input := &storagegateway.SetSMBGuestPasswordInput{
			GatewayARN: aws.String(d.Id()),
			Password:   aws.String(d.Get("smb_guest_password").(string)),
		}

		_, err := conn.SetSMBGuestPasswordWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Storage Gateway Gateway (%s) SMB guest password: %s", d.Id(), err)
		}
	}

	if d.HasChange("smb_security_strategy") {
		input := &storagegateway.UpdateSMBSecurityStrategyInput{
			GatewayARN:          aws.String(d.Id()),
			SMBSecurityStrategy: aws.String(d.Get("smb_security_strategy").(string)),
		}

		log.Printf("[DEBUG] Updating Storage Gateway SMB security strategy: %s", input)
		_, err := conn.UpdateSMBSecurityStrategyWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Storage Gateway Gateway (%s) SMB security strategy: %s", d.Id(), err)
		}
	}

	if d.HasChange("smb_file_share_visibility") {
		input := &storagegateway.UpdateSMBFileShareVisibilityInput{
			FileSharesVisible: aws.Bool(d.Get("smb_file_share_visibility").(bool)),
			GatewayARN:        aws.String(d.Id()),
		}

		log.Printf("[DEBUG] Updating Storage Gateway SMB file share visibility: %s", input)
		_, err := conn.UpdateSMBFileShareVisibilityWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Storage Gateway Gateway (%s) SMB file share visibility: %s", d.Id(), err)
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
			_, err := conn.UpdateBandwidthRateLimitWithContext(ctx, updateInput)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating Storage Gateway Gateway (%s) bandwidth rate limit: %s", d.Id(), err)
			}
		}

		if needsDelete {
			log.Printf("[DEBUG] Deleting Storage Gateway bandwidth rate limit: %s", deleteInput)
			_, err := conn.DeleteBandwidthRateLimitWithContext(ctx, deleteInput)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting Storage Gateway Gateway (%s) bandwidth rate limit: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceGatewayRead(ctx, d, meta)...)
}

func resourceGatewayDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayConn(ctx)

	log.Printf("[DEBUG] Deleting Storage Gateway Gateway: %s", d.Id())
	_, err := conn.DeleteGatewayWithContext(ctx, &storagegateway.DeleteGatewayInput{
		GatewayARN: aws.String(d.Id()),
	})

	if operationErrorCode(err) == operationErrCodeGatewayNotFound || tfawserr.ErrCodeEquals(err, storagegateway.ErrorCodeGatewayNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Storage Gateway Gateway (%s): %s", d.Id(), err)
	}

	return diags
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
		DomainName:       aws.String(tfMap[names.AttrDomainName].(string)),
		GatewayARN:       aws.String(gatewayArn),
		Password:         aws.String(tfMap[names.AttrPassword].(string)),
		UserName:         aws.String(tfMap[names.AttrUsername].(string)),
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

	if v, null, _ := nullable.Int(tfMap["day_of_month"].(string)).ValueInt64(); !null && v > 0 {
		apiObject.DayOfMonth = aws.Int64(v)
	}

	if v, null, _ := nullable.Int(tfMap["day_of_week"].(string)).ValueInt64(); !null {
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
