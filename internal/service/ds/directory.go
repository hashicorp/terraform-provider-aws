// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ds

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	directoryApplicationDeauthorizedPropagationTimeout = 2 * time.Minute
)

// @SDKResource("aws_directory_service_directory", name="Directory")
// @Tags(identifierAttribute="id")
func ResourceDirectory() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDirectoryCreate,
		ReadWithoutTimeout:   resourceDirectoryRead,
		UpdateWithoutTimeout: resourceDirectoryUpdate,
		DeleteWithoutTimeout: resourceDirectoryDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"access_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrAlias: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"connect_settings": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrAvailabilityZones: {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"connect_ips": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"customer_dns_ips": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.IsIPAddress,
							},
						},
						"customer_username": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						names.AttrSubnetIDs: {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrVPCID: {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"desired_number_of_domain_controllers": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(2),
			},
			"dns_ip_addresses": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"edition": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(directoryservice.DirectoryEdition_Values(), false),
			},
			"enable_sso": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: domainValidator,
			},
			names.AttrPassword: {
				Type:      schema.TypeString,
				Required:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			"security_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"short_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			names.AttrSize: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(directoryservice.DirectorySize_Values(), false),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrType: {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      directoryservice.DirectoryTypeSimpleAd,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(directoryservice.DirectoryType_Values(), false),
			},
			"vpc_settings": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrAvailabilityZones: {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrSubnetIDs: {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrVPCID: {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDirectoryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DSConn(ctx)

	name := d.Get(names.AttrName).(string)
	var creator directoryCreator
	switch directoryType := d.Get(names.AttrType).(string); directoryType {
	case directoryservice.DirectoryTypeAdconnector:
		creator = adConnectorCreator{}

	case directoryservice.DirectoryTypeMicrosoftAd:
		creator = microsoftADCreator{}

	case directoryservice.DirectoryTypeSimpleAd:
		creator = simpleADCreator{}
	}

	// Sometimes creating a directory will return `Failed`, especially when multiple directories are being
	// created concurrently. Retry creation in that case.
	// When it fails, it will typically be within the first few minutes of creation, so there is no need
	// to wait for deletion.
	err := tfresource.Retry(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
		if err := creator.Create(ctx, conn, name, d); err != nil {
			return retry.NonRetryableError(err)
		}

		if _, err := waitDirectoryCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
			if use, ok := errs.As[*retry.UnexpectedStateError](err); ok {
				if use.State == directoryservice.DirectoryStageFailed {
					tflog.Info(ctx, "retrying failed Directory creation", map[string]any{
						"directory_id":       d.Id(),
						names.AttrDomainName: name,
					})
					_, deleteErr := conn.DeleteDirectoryWithContext(ctx, &directoryservice.DeleteDirectoryInput{
						DirectoryId: aws.String(d.Id()),
					})
					if deleteErr != nil {
						diags = append(diags, errs.NewWarningDiagnostic(
							"Unable to Delete Failed Directory",
							fmt.Sprintf("While creating the Directory Service Directory %q, an attempt failed. Deleting the failed Directory failed: %s", name, deleteErr),
						))
					}
					return retry.RetryableError(err)
				}
			}
			return retry.NonRetryableError(err)
		}

		return nil
	}, tfresource.WithPollInterval(1*time.Minute))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, fmt.Errorf("creating Directory Service %s Directory (%s): %w", creator.TypeName(), name, err))
	}

	if v, ok := d.GetOk(names.AttrAlias); ok {
		if err := createAlias(ctx, conn, d.Id(), v.(string)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if v, ok := d.GetOk("desired_number_of_domain_controllers"); ok {
		if err := updateNumberOfDomainControllers(ctx, conn, d.Id(), v.(int), d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if _, ok := d.GetOk("enable_sso"); ok {
		if err := enableSSO(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceDirectoryRead(ctx, d, meta)...)
}

func resourceDirectoryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DSConn(ctx)

	dir, err := FindDirectoryByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Directory Service Directory (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Directory Service Directory (%s): %s", d.Id(), err)
	}

	d.Set("access_url", dir.AccessUrl)
	d.Set(names.AttrAlias, dir.Alias)
	if dir.ConnectSettings != nil {
		if err := d.Set("connect_settings", []interface{}{flattenDirectoryConnectSettingsDescription(dir.ConnectSettings, dir.DnsIpAddrs)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting connect_settings: %s", err)
		}
	} else {
		d.Set("connect_settings", nil)
	}
	d.Set(names.AttrDescription, dir.Description)
	d.Set("desired_number_of_domain_controllers", dir.DesiredNumberOfDomainControllers)
	if aws.StringValue(dir.Type) == directoryservice.DirectoryTypeAdconnector {
		d.Set("dns_ip_addresses", aws.StringValueSlice(dir.ConnectSettings.ConnectIps))
	} else {
		d.Set("dns_ip_addresses", aws.StringValueSlice(dir.DnsIpAddrs))
	}
	d.Set("edition", dir.Edition)
	d.Set("enable_sso", dir.SsoEnabled)
	d.Set(names.AttrName, dir.Name)
	if aws.StringValue(dir.Type) == directoryservice.DirectoryTypeAdconnector {
		d.Set("security_group_id", dir.ConnectSettings.SecurityGroupId)
	} else {
		d.Set("security_group_id", dir.VpcSettings.SecurityGroupId)
	}
	d.Set("short_name", dir.ShortName)
	d.Set(names.AttrSize, dir.Size)
	d.Set(names.AttrType, dir.Type)
	if dir.VpcSettings != nil {
		if err := d.Set("vpc_settings", []interface{}{flattenDirectoryVpcSettingsDescription(dir.VpcSettings)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting vpc_settings: %s", err)
		}
	} else {
		d.Set("vpc_settings", nil)
	}

	return diags
}

func resourceDirectoryUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DSConn(ctx)

	if d.HasChange("desired_number_of_domain_controllers") {
		if err := updateNumberOfDomainControllers(ctx, conn, d.Id(), d.Get("desired_number_of_domain_controllers").(int), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChange("enable_sso") {
		if _, ok := d.GetOk("enable_sso"); ok {
			if err := enableSSO(ctx, conn, d.Id()); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		} else {
			if err := disableSSO(ctx, conn, d.Id()); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	return append(diags, resourceDirectoryRead(ctx, d, meta)...)
}

func resourceDirectoryDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DSConn(ctx)

	log.Printf("[DEBUG] Deleting Directory Service Directory: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, directoryApplicationDeauthorizedPropagationTimeout, func() (interface{}, error) {
		return conn.DeleteDirectoryWithContext(ctx, &directoryservice.DeleteDirectoryInput{
			DirectoryId: aws.String(d.Id()),
		})
	}, directoryservice.ErrCodeClientException, "authorized applications")

	if tfawserr.ErrCodeEquals(err, directoryservice.ErrCodeEntityDoesNotExistException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Directory Service Directory (%s): %s", d.Id(), err)
	}

	if _, err := waitDirectoryDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Directory Service Directory (%s) delete: %s", d.Id(), err)
	}

	return diags
}

type directoryCreator interface {
	TypeName() string
	Create(ctx context.Context, conn *directoryservice.DirectoryService, name string, d *schema.ResourceData) error
}

type adConnectorCreator struct{}

func (c adConnectorCreator) TypeName() string {
	return "AD Connector"
}

func (c adConnectorCreator) Create(ctx context.Context, conn *directoryservice.DirectoryService, name string, d *schema.ResourceData) error {
	input := &directoryservice.ConnectDirectoryInput{
		Name:     aws.String(name),
		Password: aws.String(d.Get(names.AttrPassword).(string)),
		Tags:     getTagsIn(ctx),
	}

	if v, ok := d.GetOk("connect_settings"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ConnectSettings = expandDirectoryConnectSettings(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrSize); ok {
		input.Size = aws.String(v.(string))
	} else {
		// Matching previous behavior of Default: "Large" for Size attribute.
		input.Size = aws.String(directoryservice.DirectorySizeLarge)
	}

	if v, ok := d.GetOk("short_name"); ok {
		input.ShortName = aws.String(v.(string))
	}

	output, err := conn.ConnectDirectoryWithContext(ctx, input)

	if err != nil {
		return err
	}

	d.SetId(aws.StringValue(output.DirectoryId))

	return nil
}

type microsoftADCreator struct{}

func (c microsoftADCreator) TypeName() string {
	return "Microsoft AD"
}

func (c microsoftADCreator) Create(ctx context.Context, conn *directoryservice.DirectoryService, name string, d *schema.ResourceData) error {
	input := &directoryservice.CreateMicrosoftADInput{
		Name:     aws.String(name),
		Password: aws.String(d.Get(names.AttrPassword).(string)),
		Tags:     getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("edition"); ok {
		input.Edition = aws.String(v.(string))
	}

	if v, ok := d.GetOk("short_name"); ok {
		input.ShortName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("vpc_settings"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.VpcSettings = expandDirectoryVpcSettings(v.([]interface{})[0].(map[string]interface{}))
	}

	output, err := conn.CreateMicrosoftADWithContext(ctx, input)

	if err != nil {
		return err
	}

	d.SetId(aws.StringValue(output.DirectoryId))

	return nil
}

type simpleADCreator struct{}

func (c simpleADCreator) TypeName() string {
	return "Simple AD"
}

func (c simpleADCreator) Create(ctx context.Context, conn *directoryservice.DirectoryService, name string, d *schema.ResourceData) error {
	input := &directoryservice.CreateDirectoryInput{
		Name:     aws.String(name),
		Password: aws.String(d.Get(names.AttrPassword).(string)),
		Tags:     getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrSize); ok {
		input.Size = aws.String(v.(string))
	} else {
		// Matching previous behavior of Default: "Large" for Size attribute.
		input.Size = aws.String(directoryservice.DirectorySizeLarge)
	}

	if v, ok := d.GetOk("short_name"); ok {
		input.ShortName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("vpc_settings"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.VpcSettings = expandDirectoryVpcSettings(v.([]interface{})[0].(map[string]interface{}))
	}

	output, err := conn.CreateDirectoryWithContext(ctx, input)

	if err != nil {
		return err
	}

	d.SetId(aws.StringValue(output.DirectoryId))

	return nil
}

func createAlias(ctx context.Context, conn *directoryservice.DirectoryService, directoryID, alias string) error {
	input := &directoryservice.CreateAliasInput{
		Alias:       aws.String(alias),
		DirectoryId: aws.String(directoryID),
	}

	_, err := conn.CreateAliasWithContext(ctx, input)

	if err != nil {
		return fmt.Errorf("creating Directory Service Directory (%s) alias (%s): %w", directoryID, alias, err)
	}

	return nil
}

func disableSSO(ctx context.Context, conn *directoryservice.DirectoryService, directoryID string) error {
	input := &directoryservice.DisableSsoInput{
		DirectoryId: aws.String(directoryID),
	}

	_, err := conn.DisableSsoWithContext(ctx, input)

	if err != nil {
		return fmt.Errorf("disabling Directory Service Directory (%s) SSO: %w", directoryID, err)
	}

	return nil
}

func enableSSO(ctx context.Context, conn *directoryservice.DirectoryService, directoryID string) error {
	input := &directoryservice.EnableSsoInput{
		DirectoryId: aws.String(directoryID),
	}

	_, err := conn.EnableSsoWithContext(ctx, input)

	if err != nil {
		return fmt.Errorf("enabling Directory Service Directory (%s) SSO: %w", directoryID, err)
	}

	return nil
}

func updateNumberOfDomainControllers(ctx context.Context, conn *directoryservice.DirectoryService, directoryID string, desiredNumber int, timeout time.Duration) error {
	oldDomainControllers, err := FindDomainControllers(ctx, conn, &directoryservice.DescribeDomainControllersInput{
		DirectoryId: aws.String(directoryID),
	})

	if err != nil {
		return fmt.Errorf("reading Directory Service Directory (%s) domain controllers: %w", directoryID, err)
	}

	input := &directoryservice.UpdateNumberOfDomainControllersInput{
		DesiredNumber: aws.Int64(int64(desiredNumber)),
		DirectoryId:   aws.String(directoryID),
	}

	_, err = conn.UpdateNumberOfDomainControllersWithContext(ctx, input)

	if err != nil {
		return fmt.Errorf("updating Directory Service Directory (%s) number of domain controllers (%d): %w", directoryID, desiredNumber, err)
	}

	newDomainControllers, err := FindDomainControllers(ctx, conn, &directoryservice.DescribeDomainControllersInput{
		DirectoryId: aws.String(directoryID),
	})

	if err != nil {
		return fmt.Errorf("reading Directory Service Directory (%s) domain controllers: %w", directoryID, err)
	}

	var wait []string

	for _, v := range newDomainControllers {
		domainControllerID := aws.StringValue(v.DomainControllerId)
		isNew := true

		for _, v := range oldDomainControllers {
			if aws.StringValue(v.DomainControllerId) == domainControllerID {
				isNew = false

				if aws.StringValue(v.Status) != directoryservice.DomainControllerStatusActive {
					wait = append(wait, domainControllerID)
				}
			}
		}

		if isNew {
			wait = append(wait, domainControllerID)
		}
	}

	for _, v := range wait {
		if len(newDomainControllers) > len(oldDomainControllers) {
			if _, err = waitDomainControllerCreated(ctx, conn, directoryID, v, timeout); err != nil {
				return fmt.Errorf("waiting for Directory Service Directory (%s) Domain Controller (%s) create: %w", directoryID, v, err)
			}
		} else {
			if _, err := waitDomainControllerDeleted(ctx, conn, directoryID, v, timeout); err != nil {
				return fmt.Errorf("waiting for Directory Service Directory (%s) Domain Controller (%s) delete: %w", directoryID, v, err)
			}
		}
	}

	return nil
}

func expandDirectoryConnectSettings(tfMap map[string]interface{}) *directoryservice.DirectoryConnectSettings {
	if tfMap == nil {
		return nil
	}

	apiObject := &directoryservice.DirectoryConnectSettings{}

	if v, ok := tfMap["customer_dns_ips"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.CustomerDnsIps = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["customer_username"].(string); ok && v != "" {
		apiObject.CustomerUserName = aws.String(v)
	}

	if v, ok := tfMap[names.AttrSubnetIDs].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SubnetIds = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap[names.AttrVPCID].(string); ok && v != "" {
		apiObject.VpcId = aws.String(v)
	}

	return apiObject
}

func flattenDirectoryConnectSettingsDescription(apiObject *directoryservice.DirectoryConnectSettingsDescription, dnsIpAddrs []*string) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AvailabilityZones; v != nil {
		tfMap[names.AttrAvailabilityZones] = aws.StringValueSlice(v)
	}

	if v := apiObject.ConnectIps; v != nil {
		tfMap["connect_ips"] = aws.StringValueSlice(v)
	}

	if dnsIpAddrs != nil {
		tfMap["customer_dns_ips"] = aws.StringValueSlice(dnsIpAddrs)
	}

	if v := apiObject.CustomerUserName; v != nil {
		tfMap["customer_username"] = aws.StringValue(v)
	}

	if v := apiObject.SubnetIds; v != nil {
		tfMap[names.AttrSubnetIDs] = aws.StringValueSlice(v)
	}

	if v := apiObject.VpcId; v != nil {
		tfMap[names.AttrVPCID] = aws.StringValue(v)
	}

	return tfMap
}

func expandDirectoryVpcSettings(tfMap map[string]interface{}) *directoryservice.DirectoryVpcSettings { // nosemgrep:ci.caps5-in-func-name
	if tfMap == nil {
		return nil
	}

	apiObject := &directoryservice.DirectoryVpcSettings{}

	if v, ok := tfMap[names.AttrSubnetIDs].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SubnetIds = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap[names.AttrVPCID].(string); ok && v != "" {
		apiObject.VpcId = aws.String(v)
	}

	return apiObject
}

func flattenDirectoryVpcSettings(apiObject *directoryservice.DirectoryVpcSettings) map[string]interface{} { // nosemgrep:ci.caps5-in-func-name
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.SubnetIds; v != nil {
		tfMap[names.AttrSubnetIDs] = aws.StringValueSlice(v)
	}

	if v := apiObject.VpcId; v != nil {
		tfMap[names.AttrVPCID] = aws.StringValue(v)
	}

	return tfMap
}

func flattenDirectoryVpcSettingsDescription(apiObject *directoryservice.DirectoryVpcSettingsDescription) map[string]interface{} { // nosemgrep:ci.caps5-in-func-name
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AvailabilityZones; v != nil {
		tfMap[names.AttrAvailabilityZones] = aws.StringValueSlice(v)
	}

	if v := apiObject.SubnetIds; v != nil {
		tfMap[names.AttrSubnetIDs] = aws.StringValueSlice(v)
	}

	if v := apiObject.VpcId; v != nil {
		tfMap[names.AttrVPCID] = aws.StringValue(v)
	}

	return tfMap
}

func waitDirectoryCreated(ctx context.Context, conn *directoryservice.DirectoryService, id string, timeout time.Duration) (*directoryservice.DirectoryDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{directoryservice.DirectoryStageRequested, directoryservice.DirectoryStageCreating, directoryservice.DirectoryStageCreated},
		Target:  []string{directoryservice.DirectoryStageActive},
		Refresh: statusDirectoryStage(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	// Wrap any error returned with waiting message
	defer func() {
		if err != nil {
			err = fmt.Errorf("waiting for completion: %w", err)
		}
	}()

	if output, ok := outputRaw.(*directoryservice.DirectoryDescription); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StageReason)))

		return output, err
	}

	return nil, err
}

func waitDirectoryDeleted(ctx context.Context, conn *directoryservice.DirectoryService, id string, timeout time.Duration) (*directoryservice.DirectoryDescription, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{directoryservice.DirectoryStageActive, directoryservice.DirectoryStageDeleting},
		Target:  []string{},
		Refresh: statusDirectoryStage(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	// Wrap any error returned with waiting message
	defer func() {
		if err != nil {
			err = fmt.Errorf("waiting for completion: %w", err)
		}
	}()

	if output, ok := outputRaw.(*directoryservice.DirectoryDescription); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StageReason)))

		return output, err
	}

	return nil, err
}
