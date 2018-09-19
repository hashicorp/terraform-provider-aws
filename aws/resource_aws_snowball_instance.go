package aws

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"io/ioutil"
	"log"
	"time"
)

type EC2 struct {
	*client.Client
}

func getSnowballConnectionInfo(jobId string) map[string]string {
	log.Printf("Looking up Snowball IP address from JobID: %s", jobId)
	// Call Snowball Manager Here
	// Using static file for now
	snowballFile, fileErr := ioutil.ReadFile("/tmp/snowball_config.json")
	if fileErr != nil {
		panic(fileErr)
	}
	connectionInfo := map[string]string{}
	jsonErr := json.Unmarshal(snowballFile, &connectionInfo)
	if jsonErr != nil {
		panic(jsonErr)
	}
	return connectionInfo
}

func setupNewSnowballEC2Session(jobId string) *ec2.EC2 {
	connectionInfo := getSnowballConnectionInfo(jobId)
	newSession := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"),
		Endpoint:    aws.String(connectionInfo["endpoint"] + ":8008"),
		Credentials: credentials.NewStaticCredentials(connectionInfo["accessKeyId"], connectionInfo["secretKey"], ""),
	}))
	return ec2.New(newSession)
}

func resourceAwsSnowballInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSnowballInstanceCreate,
		Read:   resourceAwsSnowballInstanceRead,
		Update: resourceAwsSnowballInstanceUpdate,
		Delete: resourceAwsSnowballInstanceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		SchemaVersion: 1,
		MigrateState:  resourceAwsInstanceMigrateState,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"ami": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"associate_public_ip_address": {
				Type:     schema.TypeBool,
				ForceNew: true,
				Computed: true,
				Optional: true,
			},

			"availability_zone": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"instance_type": {
				Type:     schema.TypeString,
				Required: true,
			},

			"job_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"key_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"get_password_data": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"password_data": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"subnet_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"private_ip": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"source_dest_check": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// Suppress diff if network_interface is set
					_, ok := d.GetOk("network_interface")
					return ok
				},
			},

			"user_data": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"user_data_base64"},
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// Sometimes the EC2 API responds with the equivalent, empty SHA1 sum
					// echo -n "" | shasum
					if (old == "da39a3ee5e6b4b0d3255bfef95601890afd80709" && new == "") ||
						(old == "" && new == "da39a3ee5e6b4b0d3255bfef95601890afd80709") {
						return true
					}
					return false
				},
				StateFunc: func(v interface{}) string {
					switch v.(type) {
					case string:
						return userDataHashSum(v.(string))
					default:
						return ""
					}
				},
				ValidateFunc: validateMaxLength(16384),
			},

			"user_data_base64": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"user_data"},
				ValidateFunc: func(v interface{}, name string) (warns []string, errs []error) {
					s := v.(string)
					if !isBase64Encoded([]byte(s)) {
						errs = append(errs, fmt.Errorf(
							"%s: must be base64-encoded", name,
						))
					}
					return
				},
			},

			"public_dns": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"primary_network_interface_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"network_interface": {
				ConflictsWith: []string{"associate_public_ip_address", "subnet_id", "private_ip", "ipv6_addresses", "ipv6_address_count", "source_dest_check"},
				Type:          schema.TypeSet,
				Optional:      true,
				Computed:      true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"delete_on_termination": {
							Type:     schema.TypeBool,
							Default:  false,
							Optional: true,
							ForceNew: true,
						},
						"network_interface_id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"device_index": {
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},

			"public_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"instance_state": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"private_dns": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"disable_api_termination": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"instance_initiated_shutdown_behavior": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"monitoring": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"iam_instance_profile": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"ipv6_address_count": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"ipv6_addresses": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"tenancy": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"cpu_core_count": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"cpu_threads_per_core": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceAwsSnowballInstanceCreate(d *schema.ResourceData, meta interface{}) error {

	jobId := d.Get("job_id").(string)
	conn := setupNewSnowballEC2Session(jobId)

	instanceOpts, err := buildAwsSnowballInstanceOpts(d)
	if err != nil {
		return err
	}

	// Build the creation struct
	runOpts := &ec2.RunInstancesInput{
		Monitoring:                        instanceOpts.Monitoring,
		IamInstanceProfile:                instanceOpts.IAMInstanceProfile,
		ImageId:                           instanceOpts.ImageID,
		InstanceInitiatedShutdownBehavior: instanceOpts.InstanceInitiatedShutdownBehavior,
		InstanceType:                      instanceOpts.InstanceType,
		Ipv6AddressCount:                  instanceOpts.Ipv6AddressCount,
		Ipv6Addresses:                     instanceOpts.Ipv6Addresses,
		KeyName:                           instanceOpts.KeyName,
		MaxCount:                          aws.Int64(int64(1)),
		MinCount:                          aws.Int64(int64(1)),
		NetworkInterfaces:                 instanceOpts.NetworkInterfaces,
		PrivateIpAddress:                  instanceOpts.PrivateIPAddress,
		SubnetId:                          instanceOpts.SubnetID,
		UserData:                          instanceOpts.UserData64,
	}

	_, ipv6CountOk := d.GetOk("ipv6_address_count")
	_, ipv6AddressOk := d.GetOk("ipv6_addresses")

	if ipv6AddressOk && ipv6CountOk {
		return fmt.Errorf("Only 1 of `ipv6_address_count` or `ipv6_addresses` can be specified")
	}

	// Create the instance
	log.Printf("[DEBUG] Run configuration: %s", runOpts)

	var runResp *ec2.Reservation
	err = resource.Retry(30*time.Second, func() *resource.RetryError {
		var err error
		runResp, err = conn.RunInstances(runOpts)
		// IAM instance profiles can take ~10 seconds to propagate in AWS:
		// http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/iam-roles-for-amazon-ec2.html#launch-instance-with-role-console
		if isAWSErr(err, "InvalidParameterValue", "Invalid IAM Instance Profile") {
			log.Print("[DEBUG] Invalid IAM Instance Profile referenced, retrying...")
			return resource.RetryableError(err)
		}
		// IAM roles can also take time to propagate in AWS:
		if isAWSErr(err, "InvalidParameterValue", " has no associated IAM Roles") {
			log.Print("[DEBUG] IAM Instance Profile appears to have no IAM roles, retrying...")
			return resource.RetryableError(err)
		}
		return resource.NonRetryableError(err)
	})
	// Warn if the AWS Error involves group ids, to help identify situation
	// where a user uses group ids in security_groups for the Default VPC.
	//   See https://github.com/hashicorp/terraform/issues/3798
	if isAWSErr(err, "InvalidParameterValue", "groupId is invalid") {
		return fmt.Errorf("Error launching instance, possible mismatch of Security Group IDs and Names. See AWS Instance docs here: %s.\n\n\tAWS Error: %s", "https://terraform.io/docs/providers/aws/r/instance.html", err.(awserr.Error).Message())
	}
	if err != nil {
		return fmt.Errorf("Error launching source instance: %s", err)
	}
	if runResp == nil || len(runResp.Instances) == 0 {
		return errors.New("Error launching source instance: no instances returned in response")
	}

	instance := runResp.Instances[0]
	log.Printf("[INFO] Instance ID: %s", *instance.InstanceId)

	// Store the resulting ID so we can look this up later
	d.SetId(*instance.InstanceId)

	// Wait for the instance to become running so we can get some attributes
	// that aren't available until later.
	log.Printf(
		"[DEBUG] Waiting for instance (%s) to become running",
		*instance.InstanceId)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"pending"},
		Target:     []string{"running"},
		Refresh:    InstanceStateRefreshFunc(conn, *instance.InstanceId, []string{"terminated", "shutting-down"}),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	instanceRaw, err := stateConf.WaitForState()

	if err != nil {
		return fmt.Errorf(
			"Error waiting for instance (%s) to become ready: %s",
			*instance.InstanceId, err)
	}

	instance = instanceRaw.(*ec2.Instance)

	// Initialize the connection info
	if instance.PublicIpAddress != nil {
		d.SetConnInfo(map[string]string{
			"type": "ssh",
			"host": *instance.PublicIpAddress,
		})
	} else if instance.PrivateIpAddress != nil {
		d.SetConnInfo(map[string]string{
			"type": "ssh",
			"host": *instance.PrivateIpAddress,
		})
	}

	// Update if we need to
	return resourceAwsSnowballInstanceUpdate(d, meta)
}

func resourceAwsSnowballInstanceRead(d *schema.ResourceData, meta interface{}) error {
	jobId := d.Get("job_id").(string)
	conn := setupNewSnowballEC2Session(jobId)

	resp, err := conn.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(d.Id())},
	})
	if err != nil {
		// If the instance was not found, return nil so that we can show
		// that the instance is gone.
		if ec2err, ok := err.(awserr.Error); ok && ec2err.Code() == "InvalidInstanceID.NotFound" {
			d.SetId("")
			return nil
		}

		// Some other error, report it
		return err
	}

	// If nothing was found, then return no state
	if len(resp.Reservations) == 0 {
		d.SetId("")
		return nil
	}

	instance := resp.Reservations[0].Instances[0]

	if instance.State != nil {
		// If the instance is terminated, then it is gone
		if *instance.State.Name == "terminated" {
			d.SetId("")
			return nil
		}

		d.Set("instance_state", instance.State.Name)
	}

	if instance.CpuOptions != nil {
		d.Set("cpu_core_count", instance.CpuOptions.CoreCount)
		d.Set("cpu_threads_per_core", instance.CpuOptions.ThreadsPerCore)
	}

	d.Set("ami", instance.ImageId)
	d.Set("instance_type", instance.InstanceType)
	d.Set("key_name", instance.KeyName)
	d.Set("public_dns", instance.PublicDnsName)
	d.Set("public_ip", instance.PublicIpAddress)
	d.Set("private_dns", instance.PrivateDnsName)
	d.Set("private_ip", instance.PrivateIpAddress)
	d.Set("iam_instance_profile", iamInstanceProfileArnToName(instance.IamInstanceProfile))

	// Set configured Network Interface Device Index Slice
	// We only want to read, and populate state for the configured network_interface attachments. Otherwise, other
	// resources have the potential to attach network interfaces to the instance, and cause a perpetual create/destroy
	// diff. We should only read on changes configured for this specific resource because of this.
	var configuredDeviceIndexes []int
	if v, ok := d.GetOk("network_interface"); ok {
		vL := v.(*schema.Set).List()
		for _, vi := range vL {
			mVi := vi.(map[string]interface{})
			configuredDeviceIndexes = append(configuredDeviceIndexes, mVi["device_index"].(int))
		}
	}

	var ipv6Addresses []string
	if len(instance.NetworkInterfaces) > 0 {
		var primaryNetworkInterface ec2.InstanceNetworkInterface
		var networkInterfaces []map[string]interface{}
		for _, iNi := range instance.NetworkInterfaces {
			ni := make(map[string]interface{})
			if *iNi.Attachment.DeviceIndex == 0 {
				primaryNetworkInterface = *iNi
			}
			// If the attached network device is inside our configuration, refresh state with values found.
			// Otherwise, assume the network device was attached via an outside resource.
			for _, index := range configuredDeviceIndexes {
				if index == int(*iNi.Attachment.DeviceIndex) {
					ni["device_index"] = *iNi.Attachment.DeviceIndex
					ni["network_interface_id"] = *iNi.NetworkInterfaceId
					ni["delete_on_termination"] = *iNi.Attachment.DeleteOnTermination
				}
			}
			// Don't add empty network interfaces to schema
			if len(ni) == 0 {
				continue
			}
			networkInterfaces = append(networkInterfaces, ni)
		}
		if err := d.Set("network_interface", networkInterfaces); err != nil {
			return fmt.Errorf("Error setting network_interfaces: %v", err)
		}

		// Set primary network interface details
		// If an instance is shutting down, network interfaces are detached, and attributes may be nil,
		// need to protect against nil pointer dereferences
		if primaryNetworkInterface.SubnetId != nil {
			d.Set("subnet_id", primaryNetworkInterface.SubnetId)
		}
		if primaryNetworkInterface.NetworkInterfaceId != nil {
			d.Set("network_interface_id", primaryNetworkInterface.NetworkInterfaceId) // TODO: Deprecate me v0.10.0
			d.Set("primary_network_interface_id", primaryNetworkInterface.NetworkInterfaceId)
		}
		if primaryNetworkInterface.Ipv6Addresses != nil {
			d.Set("ipv6_address_count", len(primaryNetworkInterface.Ipv6Addresses))
		}
		if primaryNetworkInterface.SourceDestCheck != nil {
			d.Set("source_dest_check", primaryNetworkInterface.SourceDestCheck)
		}

		d.Set("associate_public_ip_address", primaryNetworkInterface.Association != nil)

		for _, address := range primaryNetworkInterface.Ipv6Addresses {
			ipv6Addresses = append(ipv6Addresses, *address.Ipv6Address)
		}

	} else {
		d.Set("subnet_id", instance.SubnetId)
		d.Set("network_interface_id", "") // TODO: Deprecate me v0.10.0
		d.Set("primary_network_interface_id", "")
	}

	if err := d.Set("ipv6_addresses", ipv6Addresses); err != nil {
		log.Printf("[WARN] Error setting ipv6_addresses for AWS Instance (%s): %s", d.Id(), err)
	}

	if instance.SubnetId != nil && *instance.SubnetId != "" {
		d.Set("source_dest_check", instance.SourceDestCheck)
	}

	if instance.Monitoring != nil && instance.Monitoring.State != nil {
		monitoringState := *instance.Monitoring.State
		d.Set("monitoring", monitoringState == "enabled" || monitoringState == "pending")
	}

	d.Set("tags", tagsToMap(instance.Tags))

	// ARN

	ec2Arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Service:   "ec2",
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("instance/%s", d.Id()),
	}
	d.Set("arn", ec2Arn.String())

	// Instance attributes

	{
		attr, err := conn.DescribeInstanceAttribute(&ec2.DescribeInstanceAttributeInput{
			Attribute:  aws.String(ec2.InstanceAttributeNameUserData),
			InstanceId: aws.String(d.Id()),
		})
		if err != nil {
			return err
		}
		if attr.UserData != nil && attr.UserData.Value != nil {
			// Since user_data and user_data_base64 conflict with each other,
			// we'll only set one or the other here to avoid a perma-diff.
			// Since user_data_base64 was added later, we'll prefer to set
			// user_data.
			_, b64 := d.GetOk("user_data_base64")
			if b64 {
				d.Set("user_data_base64", attr.UserData.Value)
			} else {
				d.Set("user_data", userDataHashSum(*attr.UserData.Value))
			}
		}
	}

	if d.Get("get_password_data").(bool) {
		passwordData, err := getAwsEc2InstancePasswordData(*instance.InstanceId, conn)
		if err != nil {
			return err
		}
		d.Set("password_data", passwordData)
	} else {
		d.Set("get_password_data", false)
		d.Set("password_data", nil)
	}

	return nil
}

func resourceAwsSnowballInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	jobId := d.Get("job_id").(string)
	conn := setupNewSnowballEC2Session(jobId)

	d.Partial(true)

	if d.HasChange("tags") {
		if !d.IsNewResource() {
			if err := setTags(conn, d); err != nil {
				return err
			} else {
				d.SetPartial("tags")
			}
		}
	}

	if d.HasChange("iam_instance_profile") && !d.IsNewResource() {
		request := &ec2.DescribeIamInstanceProfileAssociationsInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("instance-id"),
					Values: []*string{aws.String(d.Id())},
				},
			},
		}

		resp, err := conn.DescribeIamInstanceProfileAssociations(request)
		if err != nil {
			return err
		}

		// An Iam Instance Profile has been provided and is pending a change
		// This means it is an association or a replacement to an association
		if _, ok := d.GetOk("iam_instance_profile"); ok {
			// Does not have an Iam Instance Profile associated with it, need to associate
			if len(resp.IamInstanceProfileAssociations) == 0 {
				err := resource.Retry(1*time.Minute, func() *resource.RetryError {
					_, err := conn.AssociateIamInstanceProfile(&ec2.AssociateIamInstanceProfileInput{
						InstanceId: aws.String(d.Id()),
						IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
							Name: aws.String(d.Get("iam_instance_profile").(string)),
						},
					})
					if err != nil {
						if isAWSErr(err, "InvalidParameterValue", "Invalid IAM Instance Profile") {
							return resource.RetryableError(err)
						}
						return resource.NonRetryableError(err)
					}
					return nil
				})
				if err != nil {
					return err
				}

			} else {
				// Has an Iam Instance Profile associated with it, need to replace the association
				associationId := resp.IamInstanceProfileAssociations[0].AssociationId

				err := resource.Retry(1*time.Minute, func() *resource.RetryError {
					_, err := conn.ReplaceIamInstanceProfileAssociation(&ec2.ReplaceIamInstanceProfileAssociationInput{
						AssociationId: associationId,
						IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
							Name: aws.String(d.Get("iam_instance_profile").(string)),
						},
					})
					if err != nil {
						if isAWSErr(err, "InvalidParameterValue", "Invalid IAM Instance Profile") {
							return resource.RetryableError(err)
						}
						return resource.NonRetryableError(err)
					}
					return nil
				})
				if err != nil {
					return err
				}
			}
			// An Iam Instance Profile has _not_ been provided but is pending a change. This means there is a pending removal
		} else {
			if len(resp.IamInstanceProfileAssociations) > 0 {
				// Has an Iam Instance Profile associated with it, need to remove the association
				associationId := resp.IamInstanceProfileAssociations[0].AssociationId

				_, err := conn.DisassociateIamInstanceProfile(&ec2.DisassociateIamInstanceProfileInput{
					AssociationId: associationId,
				})
				if err != nil {
					return err
				}
			}
		}
	}

	// SourceDestCheck can only be modified on an instance without manually specified network interfaces.
	// SourceDestCheck, in that case, is configured at the network interface level
	if _, ok := d.GetOk("network_interface"); !ok {

		// If we have a new resource and source_dest_check is still true, don't modify
		sourceDestCheck := d.Get("source_dest_check").(bool)

		// Because we're calling Update prior to Read, and the default value of `source_dest_check` is `true`,
		// HasChange() thinks there is a diff between what is set on the instance and what is set in state. We need to ensure that
		// if a diff has occurred, it's not because it's a new instance.
		if d.HasChange("source_dest_check") && !d.IsNewResource() || d.IsNewResource() && !sourceDestCheck {
			// SourceDestCheck can only be set on VPC instances
			// AWS will return an error of InvalidParameterCombination if we attempt
			// to modify the source_dest_check of an instance in EC2 Classic
			log.Printf("[INFO] Modifying `source_dest_check` on Instance %s", d.Id())
			_, err := conn.ModifyInstanceAttribute(&ec2.ModifyInstanceAttributeInput{
				InstanceId: aws.String(d.Id()),
				SourceDestCheck: &ec2.AttributeBooleanValue{
					Value: aws.Bool(sourceDestCheck),
				},
			})
			if err != nil {
				if ec2err, ok := err.(awserr.Error); ok {
					// Tolerate InvalidParameterCombination error in Classic, otherwise
					// return the error
					if "InvalidParameterCombination" != ec2err.Code() {
						return err
					}
					log.Printf("[WARN] Attempted to modify SourceDestCheck on non VPC instance: %s", ec2err.Message())
				}
			}
		}
	}

	if d.HasChange("instance_type") && !d.IsNewResource() {
		log.Printf("[INFO] Stopping Instance %q for instance_type change", d.Id())
		_, err := conn.StopInstances(&ec2.StopInstancesInput{
			InstanceIds: []*string{aws.String(d.Id())},
		})
		if err != nil {
			return fmt.Errorf("error stopping instance (%s): %s", d.Id(), err)
		}

		stateConf := &resource.StateChangeConf{
			Pending:    []string{"pending", "running", "shutting-down", "stopped", "stopping"},
			Target:     []string{"stopped"},
			Refresh:    SnowballInstanceStateRefreshFunc(conn, d.Id(), []string{}),
			Timeout:    d.Timeout(schema.TimeoutUpdate),
			Delay:      10 * time.Second,
			MinTimeout: 3 * time.Second,
		}

		_, err = stateConf.WaitForState()
		if err != nil {
			return fmt.Errorf(
				"Error waiting for instance (%s) to stop: %s", d.Id(), err)
		}

		log.Printf("[INFO] Modifying instance type %s", d.Id())
		_, err = conn.ModifyInstanceAttribute(&ec2.ModifyInstanceAttributeInput{
			InstanceId: aws.String(d.Id()),
			InstanceType: &ec2.AttributeValue{
				Value: aws.String(d.Get("instance_type").(string)),
			},
		})
		if err != nil {
			return err
		}

		log.Printf("[INFO] Starting Instance %q after instance_type change", d.Id())
		_, err = conn.StartInstances(&ec2.StartInstancesInput{
			InstanceIds: []*string{aws.String(d.Id())},
		})
		if err != nil {
			return fmt.Errorf("error starting instance (%s): %s", d.Id(), err)
		}

		stateConf = &resource.StateChangeConf{
			Pending:    []string{"pending", "stopped"},
			Target:     []string{"running"},
			Refresh:    InstanceStateRefreshFunc(conn, d.Id(), []string{"terminated"}),
			Timeout:    d.Timeout(schema.TimeoutUpdate),
			Delay:      10 * time.Second,
			MinTimeout: 3 * time.Second,
		}

		_, err = stateConf.WaitForState()
		if err != nil {
			return fmt.Errorf(
				"Error waiting for instance (%s) to become ready: %s",
				d.Id(), err)
		}
	}

	if d.HasChange("instance_initiated_shutdown_behavior") {
		log.Printf("[INFO] Modifying instance %s", d.Id())
		_, err := conn.ModifyInstanceAttribute(&ec2.ModifyInstanceAttributeInput{
			InstanceId: aws.String(d.Id()),
			InstanceInitiatedShutdownBehavior: &ec2.AttributeValue{
				Value: aws.String(d.Get("instance_initiated_shutdown_behavior").(string)),
			},
		})
		if err != nil {
			return err
		}
	}

	if d.HasChange("monitoring") {
		var mErr error
		if d.Get("monitoring").(bool) {
			log.Printf("[DEBUG] Enabling monitoring for Instance (%s)", d.Id())
			_, mErr = conn.MonitorInstances(&ec2.MonitorInstancesInput{
				InstanceIds: []*string{aws.String(d.Id())},
			})
		} else {
			log.Printf("[DEBUG] Disabling monitoring for Instance (%s)", d.Id())
			_, mErr = conn.UnmonitorInstances(&ec2.UnmonitorInstancesInput{
				InstanceIds: []*string{aws.String(d.Id())},
			})
		}
		if mErr != nil {
			return fmt.Errorf("[WARN] Error updating Instance monitoring: %s", mErr)
		}
	}

	// TODO(mitchellh): wait for the attributes we modified to
	// persist the change...

	d.Partial(false)

	return resourceAwsSnowballInstanceRead(d, meta)
}

func resourceAwsSnowballInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	jobId := d.Get("job_id").(string)
	conn := setupNewSnowballEC2Session(jobId)

	if err := awsTerminateSnowballInstance(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return err
	}

	return nil
}

// SnowballInstanceStateRefreshFunc returns a resource.StateRefreshFunc that is used to watch
// an EC2 instance.
func SnowballInstanceStateRefreshFunc(conn *ec2.EC2, instanceID string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeInstances(&ec2.DescribeInstancesInput{
			InstanceIds: []*string{aws.String(instanceID)},
		})
		if err != nil {
			if ec2err, ok := err.(awserr.Error); ok && ec2err.Code() == "InvalidInstanceID.NotFound" {
				// Set this to nil as if we didn't find anything.
				resp = nil
			} else {
				log.Printf("Error on InstanceStateRefresh: %s", err)
				return nil, "", err
			}
		}

		if resp == nil || len(resp.Reservations) == 0 || len(resp.Reservations[0].Instances) == 0 {
			// Sometimes AWS just has consistency issues and doesn't see
			// our instance yet. Return an empty state.
			return nil, "", nil
		}

		i := resp.Reservations[0].Instances[0]
		state := *i.State.Name

		for _, failState := range failStates {
			if state == failState {
				return i, state, fmt.Errorf("Failed to reach target state. Reason: %s",
					stringifyStateReason(i.StateReason))
			}
		}

		return i, state, nil
	}
}

func buildSnowballNetworkInterfaceOpts(d *schema.ResourceData, groups []*string, nInterfaces interface{}) []*ec2.InstanceNetworkInterfaceSpecification {
	networkInterfaces := []*ec2.InstanceNetworkInterfaceSpecification{}
	// Get necessary items
	subnet, hasSubnet := d.GetOk("subnet_id")

	if hasSubnet {
		// If we have a non-default VPC / Subnet specified, we can flag
		// AssociatePublicIpAddress to get a Public IP assigned. By default these are not provided.
		// You cannot specify both SubnetId and the NetworkInterface.0.* parameters though, otherwise
		// you get: Network interfaces and an instance-level subnet ID may not be specified on the same request
		// You also need to attach Security Groups to the NetworkInterface instead of the instance,
		// to avoid: Network interfaces and an instance-level security groups may not be specified on
		// the same request
		ni := &ec2.InstanceNetworkInterfaceSpecification{
			DeviceIndex: aws.Int64(int64(0)),
			SubnetId:    aws.String(subnet.(string)),
			Groups:      groups,
		}

		if v, ok := d.GetOkExists("associate_public_ip_address"); ok {
			ni.AssociatePublicIpAddress = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk("private_ip"); ok {
			ni.PrivateIpAddress = aws.String(v.(string))
		}

		if v, ok := d.GetOk("ipv6_address_count"); ok {
			ni.Ipv6AddressCount = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("ipv6_addresses"); ok {
			ipv6Addresses := make([]*ec2.InstanceIpv6Address, len(v.([]interface{})))
			for _, address := range v.([]interface{}) {
				ipv6Address := &ec2.InstanceIpv6Address{
					Ipv6Address: aws.String(address.(string)),
				}

				ipv6Addresses = append(ipv6Addresses, ipv6Address)
			}

			ni.Ipv6Addresses = ipv6Addresses
		}

		networkInterfaces = append(networkInterfaces, ni)
	} else {
		// If we have manually specified network interfaces, build and attach those here.
		vL := nInterfaces.(*schema.Set).List()
		for _, v := range vL {
			ini := v.(map[string]interface{})
			ni := &ec2.InstanceNetworkInterfaceSpecification{
				DeviceIndex:         aws.Int64(int64(ini["device_index"].(int))),
				NetworkInterfaceId:  aws.String(ini["network_interface_id"].(string)),
				DeleteOnTermination: aws.Bool(ini["delete_on_termination"].(bool)),
			}
			networkInterfaces = append(networkInterfaces, ni)
		}
	}

	return networkInterfaces
}

type awsSnowballInstanceOpts struct {
	DisableAPITermination             *bool
	Monitoring                        *ec2.RunInstancesMonitoringEnabled
	IAMInstanceProfile                *ec2.IamInstanceProfileSpecification
	ImageID                           *string
	InstanceInitiatedShutdownBehavior *string
	InstanceType                      *string
	Ipv6AddressCount                  *int64
	Ipv6Addresses                     []*ec2.InstanceIpv6Address
	KeyName                           *string
	NetworkInterfaces                 []*ec2.InstanceNetworkInterfaceSpecification
	PrivateIPAddress                  *string
	SubnetID                          *string
	UserData64                        *string
	CpuOptions                        *ec2.CpuOptionsRequest
}

func buildAwsSnowballInstanceOpts(
	d *schema.ResourceData) (*awsSnowballInstanceOpts, error) {

	instanceType := d.Get("instance_type").(string)
	opts := &awsSnowballInstanceOpts{
		ImageID:      aws.String(d.Get("ami").(string)),
		InstanceType: aws.String(instanceType),
	}

	if v := d.Get("instance_initiated_shutdown_behavior").(string); v != "" {
		opts.InstanceInitiatedShutdownBehavior = aws.String(v)
	}

	opts.Monitoring = &ec2.RunInstancesMonitoringEnabled{
		Enabled: aws.Bool(d.Get("monitoring").(bool)),
	}

	opts.IAMInstanceProfile = &ec2.IamInstanceProfileSpecification{
		Name: aws.String(d.Get("iam_instance_profile").(string)),
	}

	userData := d.Get("user_data").(string)
	userDataBase64 := d.Get("user_data_base64").(string)

	if userData != "" {
		opts.UserData64 = aws.String(base64Encode([]byte(userData)))
	} else if userDataBase64 != "" {
		opts.UserData64 = aws.String(userDataBase64)
	}

	// check for non-default Subnet, and cast it to a String
	subnet, hasSubnet := d.GetOk("subnet_id")
	subnetID := subnet.(string)

	if v := d.Get("cpu_core_count").(int); v > 0 {
		tc := d.Get("cpu_threads_per_core").(int)
		if tc < 0 {
			tc = 2
		}
		opts.CpuOptions = &ec2.CpuOptionsRequest{
			CoreCount:      aws.Int64(int64(v)),
			ThreadsPerCore: aws.Int64(int64(tc)),
		}
	}

	var groups []*string

	networkInterfaces, interfacesOk := d.GetOk("network_interface")

	// If setting subnet and public address, OR manual network interfaces, populate those now.
	if hasSubnet || interfacesOk {
		// Otherwise we're attaching (a) network interface(s)
		opts.NetworkInterfaces = buildSnowballNetworkInterfaceOpts(d, groups, networkInterfaces)
	} else {
		// If simply specifying a subnetID, privateIP, Security Groups, or VPC Security Groups, build these now
		if subnetID != "" {
			opts.SubnetID = aws.String(subnetID)
		}

		if v, ok := d.GetOk("private_ip"); ok {
			opts.PrivateIPAddress = aws.String(v.(string))
		}

		if v, ok := d.GetOk("ipv6_address_count"); ok {
			opts.Ipv6AddressCount = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("ipv6_addresses"); ok {
			ipv6Addresses := make([]*ec2.InstanceIpv6Address, len(v.([]interface{})))
			for _, address := range v.([]interface{}) {
				ipv6Address := &ec2.InstanceIpv6Address{
					Ipv6Address: aws.String(address.(string)),
				}

				ipv6Addresses = append(ipv6Addresses, ipv6Address)
			}

			opts.Ipv6Addresses = ipv6Addresses
		}
	}

	if v, ok := d.GetOk("key_name"); ok {
		opts.KeyName = aws.String(v.(string))
	}

	return opts, nil
}

func awsTerminateSnowballInstance(conn *ec2.EC2, id string, timeout time.Duration) error {
	log.Printf("[INFO] Terminating instance: %s", id)
	req := &ec2.TerminateInstancesInput{
		InstanceIds: []*string{aws.String(id)},
	}
	if _, err := conn.TerminateInstances(req); err != nil {
		return fmt.Errorf("Error terminating instance: %s", err)
	}

	log.Printf("[DEBUG] Waiting for instance (%s) to become terminated", id)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"pending", "running", "shutting-down", "stopped", "stopping"},
		Target:     []string{"terminated"},
		Refresh:    InstanceStateRefreshFunc(conn, id, []string{}),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf(
			"Error waiting for instance (%s) to terminate: %s", id, err)
	}

	return nil
}
