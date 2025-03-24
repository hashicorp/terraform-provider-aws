// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package storagegateway

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"slices"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/storagegateway"
	awstypes "github.com/aws/aws-sdk-go-v2/service/storagegateway/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
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
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[awstypes.SMBSecurityStrategy](),
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
			customdiff.ForceNewIfChange("smb_active_directory_settings", func(_ context.Context, old, new, meta any) bool {
				return len(old.([]any)) == 1 && len(new.([]any)) == 0
			}),
		),
	}
}

func resourceGatewayCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayClient(ctx)

	region := meta.(*conns.AWSClient).Region(ctx)
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

		requestURL := fmt.Sprintf("http://%[1]s/?activationRegion=%[2]s", gatewayIPAddress, region)
		if v, ok := d.GetOk("gateway_vpc_endpoint"); ok {
			requestURL = fmt.Sprintf("%[1]s&vpcEndpoint=%[2]s", requestURL, v.(string))
		}

		request, err := http.NewRequest(http.MethodGet, requestURL, nil)

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		var response *http.Response
		err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
			response, err = client.Do(request)

			if err != nil {
				if errs.IsA[net.Error](err) {
					errMessage := fmt.Errorf("making HTTP request: %w", err)
					log.Printf("[DEBUG] retryable %s", errMessage)
					return retry.RetryableError(errMessage)
				}

				return retry.NonRetryableError(fmt.Errorf("making HTTP request: %w", err))
			}

			if slices.Contains([]int{504}, response.StatusCode) {
				errMessage := fmt.Errorf("status code in HTTP response: %d", response.StatusCode)
				log.Printf("[DEBUG] retryable %s", errMessage)
				return retry.RetryableError(errMessage)
			}

			return nil
		})

		if tfresource.TimedOut(err) {
			response, err = client.Do(request)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "retrieving activation key from IP Address (%s): %s", gatewayIPAddress, err)
		}

		if response.StatusCode != http.StatusFound {
			return sdkdiag.AppendErrorf(diags, "expected HTTP status code 302, received: %d", response.StatusCode)
		}

		redirectURL, err := response.Location()

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		activationKey = redirectURL.Query().Get("activationKey")
		if activationKey == "" {
			return sdkdiag.AppendErrorf(diags, "empty activationKey received from IP Address: %s", gatewayIPAddress)
		}
	}

	name := d.Get("gateway_name").(string)
	input := &storagegateway.ActivateGatewayInput{
		ActivationKey:   aws.String(activationKey),
		GatewayRegion:   aws.String(region),
		GatewayName:     aws.String(name),
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

	output, err := conn.ActivateGateway(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "activating Storage Gateway Gateway (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.GatewayARN))

	if _, err = waitGatewayConnected(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Storage Gateway Gateway (%s) connect: %s", d.Id(), err)
	}

	if v, ok := d.GetOk(names.AttrCloudWatchLogGroupARN); ok && v.(string) != "" {
		input := &storagegateway.UpdateGatewayInformationInput{
			CloudWatchLogGroupARN: aws.String(v.(string)),
			GatewayARN:            aws.String(d.Id()),
		}

		_, err := conn.UpdateGatewayInformation(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Storage Gateway Gateway (%s) CloudWatch log group: %s", d.Id(), err)
		}
	}

	if v, ok := d.GetOk("maintenance_start_time"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input := expandUpdateMaintenanceStartTimeInput(v.([]any)[0].(map[string]any))
		input.GatewayARN = aws.String(d.Id())

		_, err := conn.UpdateMaintenanceStartTime(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Storage Gateway Gateway (%s) maintenance start time: %s", d.Id(), err)
		}
	}

	if v, ok := d.GetOk("smb_active_directory_settings"); ok && len(v.([]any)) > 0 {
		input := expandJoinDomainInput(v.([]any), d.Id())

		_, err := conn.JoinDomain(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "joining Storage Gateway Gateway (%s) to Active Directory domain (%s): %s", d.Id(), aws.ToString(input.DomainName), err)
		}

		if _, err = waitGatewayJoinDomainJoined(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Storage Gateway Gateway (%s) domain join: %s", d.Id(), err)
		}
	}

	if v, ok := d.GetOk("smb_guest_password"); ok && v.(string) != "" {
		input := &storagegateway.SetSMBGuestPasswordInput{
			GatewayARN: aws.String(d.Id()),
			Password:   aws.String(v.(string)),
		}

		_, err := conn.SetSMBGuestPassword(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting Storage Gateway Gateway (%s) SMB guest password: %s", d.Id(), err)
		}
	}

	if v, ok := d.GetOk("smb_security_strategy"); ok {
		input := &storagegateway.UpdateSMBSecurityStrategyInput{
			GatewayARN:          aws.String(d.Id()),
			SMBSecurityStrategy: awstypes.SMBSecurityStrategy(v.(string)),
		}

		_, err := conn.UpdateSMBSecurityStrategy(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting Storage Gateway Gateway (%s) SMB security strategy: %s", d.Id(), err)
		}
	}

	if v, ok := d.GetOk("smb_file_share_visibility"); ok {
		input := &storagegateway.UpdateSMBFileShareVisibilityInput{
			FileSharesVisible: aws.Bool(v.(bool)),
			GatewayARN:        aws.String(d.Id()),
		}

		_, err := conn.UpdateSMBFileShareVisibility(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Storage Gateway Gateway (%s) SMB file share visibility: %s", d.Id(), err)
		}
	}

	switch d.Get("gateway_type").(string) {
	case gatewayTypeCached, gatewayTypeStored, gatewayTypeVTL, gatewayTypeVTLSnow:
		input := &storagegateway.UpdateBandwidthRateLimitInput{
			GatewayARN: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("average_download_rate_limit_in_bits_per_sec"); ok {
			input.AverageDownloadRateLimitInBitsPerSec = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("average_upload_rate_limit_in_bits_per_sec"); ok {
			input.AverageUploadRateLimitInBitsPerSec = aws.Int64(int64(v.(int)))
		}

		if input.AverageDownloadRateLimitInBitsPerSec != nil || input.AverageUploadRateLimitInBitsPerSec != nil {
			_, err := conn.UpdateBandwidthRateLimit(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating Storage Gateway Gateway (%s) bandwidth rate limits: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceGatewayRead(ctx, d, meta)...)
}

func resourceGatewayRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayClient(ctx)

	outputDGI, err := findGatewayByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Storage Gateway Gateway (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Storage Gateway Gateway (%s): %s", d.Id(), err)
	}

	outputDSS, err := findSMBSettingsByARN(ctx, conn, d.Id())

	switch {
	case errs.IsAErrorMessageContains[*awstypes.InvalidGatewayRequestException](err, "This operation is not valid for the specified gateway"):
	case err != nil:
		return sdkdiag.AppendErrorf(diags, "reading Storage Gateway Gateway (%s) SMB settings: %s", d.Id(), err)
	}

	d.Set("activation_key", d.Get("activation_key").(string))
	d.Set(names.AttrARN, outputDGI.GatewayARN)
	d.Set(names.AttrCloudWatchLogGroupARN, outputDGI.CloudWatchLogGroupARN)
	d.Set("ec2_instance_id", outputDGI.Ec2InstanceId)
	d.Set(names.AttrEndpointType, outputDGI.EndpointType)
	d.Set("gateway_id", outputDGI.GatewayId)
	d.Set("gateway_ip_address", d.Get("gateway_ip_address").(string))
	d.Set("gateway_name", outputDGI.GatewayName)
	if err := d.Set("gateway_network_interface", flattenNetworkInterfaces(outputDGI.GatewayNetworkInterfaces)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting gateway_network_interface: %s", err)
	}
	d.Set("gateway_timezone", outputDGI.GatewayTimezone)
	d.Set("gateway_type", outputDGI.GatewayType)
	d.Set("gateway_vpc_endpoint", outputDGI.VPCEndpoint)
	d.Set("host_environment", outputDGI.HostEnvironment)
	d.Set("medium_changer_type", d.Get("medium_changer_type").(string))
	if outputDSS == nil || aws.ToString(outputDSS.DomainName) == "" {
		if err := d.Set("smb_active_directory_settings", []any{}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting smb_active_directory_settings: %s", err)
		}
	} else {
		tfMap := map[string]any{
			"active_directory_status": outputDSS.ActiveDirectoryStatus,
			names.AttrDomainName:      aws.ToString(outputDSS.DomainName),
		}

		if v, ok := d.GetOk("smb_active_directory_settings"); ok && len(v.([]any)) > 0 {
			configM := v.([]any)[0].(map[string]any)
			tfMap[names.AttrPassword] = configM[names.AttrPassword]
			tfMap["timeout_in_seconds"] = configM["timeout_in_seconds"]
			tfMap[names.AttrUsername] = configM[names.AttrUsername]

			if v, ok := configM["domain_controllers"]; ok {
				tfMap["domain_controllers"] = v
			}

			if v, ok := configM["organizational_unit"]; ok {
				tfMap["organizational_unit"] = v
			}
		}

		if err := d.Set("smb_active_directory_settings", []map[string]any{tfMap}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting smb_active_directory_settings: %s", err)
		}
	}
	// The Storage Gateway API currently provides no way to read this value
	// We allow Terraform to _automatically_ passthrough the configuration value into the state here
	// as the API does clue us in whether or not its actually set at all,
	// which can be used to tell Terraform to show a difference in this case
	// as well as ensuring there is some sort of attribute value (unlike the others).
	if outputDSS == nil || !aws.ToBool(outputDSS.SMBGuestPasswordSet) {
		d.Set("smb_guest_password", "")
	}
	if outputDSS != nil {
		d.Set("smb_file_share_visibility", outputDSS.FileSharesVisible)
		d.Set("smb_security_strategy", outputDSS.SMBSecurityStrategy)
	}
	d.Set("tape_drive_type", d.Get("tape_drive_type").(string))

	setTagsOut(ctx, outputDGI.Tags)

	switch aws.ToString(outputDGI.GatewayType) {
	case gatewayTypeCached, gatewayTypeStored, gatewayTypeVTL, gatewayTypeVTLSnow:
		input := &storagegateway.DescribeBandwidthRateLimitInput{
			GatewayARN: aws.String(d.Id()),
		}

		outputDBRL, err := conn.DescribeBandwidthRateLimit(ctx, input)

		switch {
		case errs.IsAErrorMessageContains[*awstypes.InvalidGatewayRequestException](err, "not supported"):
		case errs.IsAErrorMessageContains[*awstypes.InvalidGatewayRequestException](err, "not valid"):
		case err != nil:
			return sdkdiag.AppendErrorf(diags, "reading Storage Gateway Gateway (%s) bandwidth rate limits: %s", d.Id(), err)
		default:
			d.Set("average_download_rate_limit_in_bits_per_sec", outputDBRL.AverageDownloadRateLimitInBitsPerSec)
			d.Set("average_upload_rate_limit_in_bits_per_sec", outputDBRL.AverageUploadRateLimitInBitsPerSec)
		}
	}

	input := &storagegateway.DescribeMaintenanceStartTimeInput{
		GatewayARN: aws.String(d.Id()),
	}

	outputDMST, err := conn.DescribeMaintenanceStartTime(ctx, input)

	switch {
	case errs.IsAErrorMessageContains[*awstypes.InvalidGatewayRequestException](err, "The specified operation is not supported"):
		fallthrough
	case errs.IsAErrorMessageContains[*awstypes.InvalidGatewayRequestException](err, "This operation is not valid for the specified gateway"):
		d.Set("maintenance_start_time", nil)
	case err != nil:
		return sdkdiag.AppendErrorf(diags, "reading Storage Gateway Gateway (%s) maintenance start time: %s", d.Id(), err)
	default:
		if err := d.Set("maintenance_start_time", []map[string]any{flattenDescribeMaintenanceStartTimeOutput(outputDMST)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting maintenance_start_time: %s", err)
		}
	}

	return diags
}

func resourceGatewayUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayClient(ctx)

	if d.HasChanges(names.AttrCloudWatchLogGroupARN, "gateway_name", "gateway_timezone") {
		input := &storagegateway.UpdateGatewayInformationInput{
			CloudWatchLogGroupARN: aws.String(d.Get(names.AttrCloudWatchLogGroupARN).(string)),
			GatewayARN:            aws.String(d.Id()),
			GatewayName:           aws.String(d.Get("gateway_name").(string)),
			GatewayTimezone:       aws.String(d.Get("gateway_timezone").(string)),
		}

		_, err := conn.UpdateGatewayInformation(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Storage Gateway Gateway (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("maintenance_start_time") {
		if v, ok := d.GetOk("maintenance_start_time"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			input := expandUpdateMaintenanceStartTimeInput(v.([]any)[0].(map[string]any))
			input.GatewayARN = aws.String(d.Id())

			_, err := conn.UpdateMaintenanceStartTime(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating Storage Gateway Gateway (%s) maintenance start time: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("smb_active_directory_settings") {
		input := expandJoinDomainInput(d.Get("smb_active_directory_settings").([]any), d.Id())

		_, err := conn.JoinDomain(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "joining Storage Gateway Gateway (%s) to Active Directory domain (%s): %s", d.Id(), aws.ToString(input.DomainName), err)
		}

		if _, err = waitGatewayJoinDomainJoined(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Storage Gateway Gateway (%s) domain join: %s", d.Id(), err)
		}
	}

	if d.HasChange("smb_guest_password") {
		input := &storagegateway.SetSMBGuestPasswordInput{
			GatewayARN: aws.String(d.Id()),
			Password:   aws.String(d.Get("smb_guest_password").(string)),
		}

		_, err := conn.SetSMBGuestPassword(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting Storage Gateway Gateway (%s) SMB guest password: %s", d.Id(), err)
		}
	}

	if d.HasChange("smb_security_strategy") {
		input := &storagegateway.UpdateSMBSecurityStrategyInput{
			GatewayARN:          aws.String(d.Id()),
			SMBSecurityStrategy: awstypes.SMBSecurityStrategy(d.Get("smb_security_strategy").(string)),
		}

		_, err := conn.UpdateSMBSecurityStrategy(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Storage Gateway Gateway (%s) SMB security strategy: %s", d.Id(), err)
		}
	}

	if d.HasChange("smb_file_share_visibility") {
		input := &storagegateway.UpdateSMBFileShareVisibilityInput{
			FileSharesVisible: aws.Bool(d.Get("smb_file_share_visibility").(bool)),
			GatewayARN:        aws.String(d.Id()),
		}

		_, err := conn.UpdateSMBFileShareVisibility(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Storage Gateway Gateway (%s) SMB file share visibility: %s", d.Id(), err)
		}
	}

	if d.HasChanges("average_download_rate_limit_in_bits_per_sec", "average_upload_rate_limit_in_bits_per_sec") {
		inputD := &storagegateway.DeleteBandwidthRateLimitInput{
			GatewayARN: aws.String(d.Id()),
		}
		needsDelete := false
		inputU := &storagegateway.UpdateBandwidthRateLimitInput{
			GatewayARN: aws.String(d.Id()),
		}
		needsUpdate := false

		if v, ok := d.GetOk("average_download_rate_limit_in_bits_per_sec"); ok {
			inputU.AverageDownloadRateLimitInBitsPerSec = aws.Int64(int64(v.(int)))
			needsUpdate = true
		} else if d.HasChange("average_download_rate_limit_in_bits_per_sec") {
			inputD.BandwidthType = aws.String(bandwidthTypeDownload)
			needsDelete = true
		}

		if v, ok := d.GetOk("average_upload_rate_limit_in_bits_per_sec"); ok {
			inputU.AverageUploadRateLimitInBitsPerSec = aws.Int64(int64(v.(int)))
			needsUpdate = true
		} else if d.HasChange("average_upload_rate_limit_in_bits_per_sec") {
			if needsDelete {
				inputD.BandwidthType = aws.String(bandwidthTypeAll)
			} else {
				inputD.BandwidthType = aws.String(bandwidthTypeUpload)
				needsDelete = true
			}
		}

		if needsUpdate {
			_, err := conn.UpdateBandwidthRateLimit(ctx, inputU)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating Storage Gateway Gateway (%s) bandwidth rate limits: %s", d.Id(), err)
			}
		}

		if needsDelete {
			_, err := conn.DeleteBandwidthRateLimit(ctx, inputD)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting Storage Gateway Gateway (%s) bandwidth rate limits: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceGatewayRead(ctx, d, meta)...)
}

func resourceGatewayDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayClient(ctx)

	log.Printf("[DEBUG] Deleting Storage Gateway Gateway: %s", d.Id())
	_, err := conn.DeleteGateway(ctx, &storagegateway.DeleteGatewayInput{
		GatewayARN: aws.String(d.Id()),
	})

	if isGatewayNotFoundErr(err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Storage Gateway Gateway (%s): %s", d.Id(), err)
	}

	return diags
}

func findGatewayByARN(ctx context.Context, conn *storagegateway.Client, arn string) (*storagegateway.DescribeGatewayInformationOutput, error) {
	input := &storagegateway.DescribeGatewayInformationInput{
		GatewayARN: aws.String(arn),
	}

	return findGateway(ctx, conn, input)
}

func findGateway(ctx context.Context, conn *storagegateway.Client, input *storagegateway.DescribeGatewayInformationInput) (*storagegateway.DescribeGatewayInformationOutput, error) {
	output, err := conn.DescribeGatewayInformation(ctx, input)

	if isGatewayNotFoundErr(err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func findSMBSettingsByARN(ctx context.Context, conn *storagegateway.Client, arn string) (*storagegateway.DescribeSMBSettingsOutput, error) {
	input := &storagegateway.DescribeSMBSettingsInput{
		GatewayARN: aws.String(arn),
	}

	return findSMBSettings(ctx, conn, input)
}

func findSMBSettings(ctx context.Context, conn *storagegateway.Client, input *storagegateway.DescribeSMBSettingsInput) (*storagegateway.DescribeSMBSettingsOutput, error) {
	output, err := conn.DescribeSMBSettings(ctx, input)

	if isGatewayNotFoundErr(err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

const (
	gatewayStatusConnected    = "GatewayConnected"
	gatewayStatusNotConnected = "GatewayNotConnected"
)

func statusGatewayConnected(ctx context.Context, conn *storagegateway.Client, gatewayARN string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findGatewayByARN(ctx, conn, gatewayARN)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if errs.IsAErrorMessageContains[*awstypes.InvalidGatewayRequestException](err, "The specified gateway is not connected") {
			return output, gatewayStatusNotConnected, nil
		}

		if err != nil {
			return output, "", err
		}

		return output, gatewayStatusConnected, nil
	}
}

func statusGatewayJoinDomain(ctx context.Context, conn *storagegateway.Client, gatewayARN string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findSMBSettingsByARN(ctx, conn, gatewayARN)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return output, "", err
		}

		return output, string(output.ActiveDirectoryStatus), nil
	}
}

func waitGatewayConnected(ctx context.Context, conn *storagegateway.Client, gatewayARN string, timeout time.Duration) (*storagegateway.DescribeGatewayInformationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{gatewayStatusNotConnected},
		Target:                    []string{gatewayStatusConnected},
		Refresh:                   statusGatewayConnected(ctx, conn, gatewayARN),
		Timeout:                   timeout,
		MinTimeout:                10 * time.Second,
		ContinuousTargetOccurence: 6, // Gateway activations can take a few seconds and can trigger a reboot of the Gateway.
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*storagegateway.DescribeGatewayInformationOutput); ok {
		return output, err
	}

	return nil, err
}

func waitGatewayJoinDomainJoined(ctx context.Context, conn *storagegateway.Client, gatewayARN string) (*storagegateway.DescribeSMBSettingsOutput, error) { //nolint:unparam
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ActiveDirectoryStatusJoining),
		Target:  enum.Slice(awstypes.ActiveDirectoryStatusJoined),
		Refresh: statusGatewayJoinDomain(ctx, conn, gatewayARN),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*storagegateway.DescribeSMBSettingsOutput); ok {
		return output, err
	}

	return nil, err
}

func expandJoinDomainInput(tfList []any, gatewayARN string) *storagegateway.JoinDomainInput {
	if tfList == nil || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &storagegateway.JoinDomainInput{
		DomainName:       aws.String(tfMap[names.AttrDomainName].(string)),
		GatewayARN:       aws.String(gatewayARN),
		Password:         aws.String(tfMap[names.AttrPassword].(string)),
		TimeoutInSeconds: aws.Int32(int32(tfMap["timeout_in_seconds"].(int))),
		UserName:         aws.String(tfMap[names.AttrUsername].(string)),
	}

	if v, ok := tfMap["domain_controllers"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.DomainControllers = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["organizational_unit"].(string); ok && v != "" {
		apiObject.OrganizationalUnit = aws.String(v)
	}

	return apiObject
}

func flattenNetworkInterfaces(apiObjects []awstypes.NetworkInterface) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			"ipv4_address": aws.ToString(apiObject.Ipv4Address),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func expandUpdateMaintenanceStartTimeInput(tfMap map[string]any) *storagegateway.UpdateMaintenanceStartTimeInput {
	if tfMap == nil {
		return nil
	}

	apiObject := &storagegateway.UpdateMaintenanceStartTimeInput{}

	if v, null, _ := nullable.Int(tfMap["day_of_month"].(string)).ValueInt32(); !null && v > 0 {
		apiObject.DayOfMonth = aws.Int32(v)
	}

	if v, null, _ := nullable.Int(tfMap["day_of_week"].(string)).ValueInt32(); !null {
		apiObject.DayOfWeek = aws.Int32(v)
	}

	if v, ok := tfMap["hour_of_day"].(int); ok {
		apiObject.HourOfDay = aws.Int32(int32(v))
	}

	if v, ok := tfMap["minute_of_hour"].(int); ok {
		apiObject.MinuteOfHour = aws.Int32(int32(v))
	}

	return apiObject
}

func flattenDescribeMaintenanceStartTimeOutput(apiObject *storagegateway.DescribeMaintenanceStartTimeOutput) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.DayOfMonth; v != nil {
		tfMap["day_of_month"] = flex.Int32ToStringValue(v)
	}

	if v := apiObject.DayOfWeek; v != nil {
		tfMap["day_of_week"] = flex.Int32ToStringValue(v)
	}

	if v := apiObject.HourOfDay; v != nil {
		tfMap["hour_of_day"] = aws.ToInt32(v)
	}

	if v := apiObject.MinuteOfHour; v != nil {
		tfMap["minute_of_hour"] = aws.ToInt32(v)
	}

	return tfMap
}
