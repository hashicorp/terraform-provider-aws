// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/sdk"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_iam_group", &resource.Sweeper{
		Name: "aws_iam_group",
		F:    sweepGroups,
		Dependencies: []string{
			"aws_iam_user",
		},
	})

	sweep.Register("aws_iam_instance_profile", sweepInstanceProfile,
		"aws_iam_role",
	)

	sweep.Register("aws_iam_openid_connect_provider", sweepOpenIDConnectProvider)

	resource.AddTestSweepers("aws_iam_policy", &resource.Sweeper{
		Name: "aws_iam_policy",
		F:    sweepPolicies,
		Dependencies: []string{
			"aws_iam_group",
			"aws_iam_role",
			"aws_iam_user",
			"aws_quicksight_group",
			"aws_quicksight_user",
		},
	})

	resource.AddTestSweepers("aws_iam_role", &resource.Sweeper{
		Name: "aws_iam_role",
		Dependencies: []string{
			"aws_batch_compute_environment",
			"aws_cloudformation_stack_set_instance",
			"aws_cognito_user_pool",
			"aws_config_configuration_aggregator",
			"aws_config_configuration_recorder",
			"aws_datasync_location",
			"aws_dax_cluster",
			"aws_db_instance",
			"aws_db_option_group",
			"aws_eks_cluster",
			"aws_elastic_beanstalk_application",
			"aws_elastic_beanstalk_environment",
			"aws_elasticsearch_domain",
			"aws_glue_crawler",
			"aws_glue_job",
			"aws_instance",
			"aws_iot_topic_rule_destination",
			"aws_lambda_function",
			"aws_launch_configuration",
			"aws_opensearch_domain",
			"aws_redshift_cluster",
			"aws_redshift_scheduled_action",
			"aws_spot_fleet_request",
			"aws_vpc",
		},
		F: sweepRoles,
	})

	sweep.Register("aws_iam_saml_provider", sweepSAMLProvider)

	sweep.Register("aws_iam_service_specific_credential", sweepServiceSpecificCredentials)

	sweep.Register("aws_iam_signing_certificate", sweepSigningCertificates)

	resource.AddTestSweepers("aws_iam_server_certificate", &resource.Sweeper{
		Name: "aws_iam_server_certificate",
		F:    sweepServerCertificates,
	})

	sweep.Register("aws_iam_service_linked_role", sweepServiceLinkedRoles)

	resource.AddTestSweepers("aws_iam_user", &resource.Sweeper{
		Name: "aws_iam_user",
		F:    sweepUsers,
		Dependencies: []string{
			"aws_iam_service_specific_credential",
			"aws_iam_virtual_mfa_device",
			"aws_iam_signing_certificate",
			"aws_opsworks_user_profile",
		},
	})

	sweep.Register("aws_iam_virtual_mfa_device", sweepVirtualMFADevice)
}

func sweepGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.IAMClient(ctx)
	input := &iam.ListGroupsInput{}
	var sweeperErrs *multierror.Error

	pages := iam.NewListGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping IAM Group sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil()
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving IAM Groups: %w", err))
		}

		for _, group := range page.Groups {
			name := aws.ToString(group.GroupName)

			if name == "Admin" || name == "TerraformAccTests" {
				continue
			}

			log.Printf("[INFO] Deleting IAM Group: %s", name)

			getGroupInput := &iam.GetGroupInput{
				GroupName: group.GroupName,
			}

			getGroupOutput, err := conn.GetGroup(ctx, getGroupInput)

			if errs.IsA[*awstypes.NoSuchEntityException](err) {
				continue
			}

			if err != nil {
				sweeperErr := fmt.Errorf("error reading IAM Group (%s): %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}

			if getGroupOutput != nil {
				for _, user := range getGroupOutput.Users {
					username := aws.ToString(user.UserName)

					log.Printf("[INFO] Removing IAM User (%s) from Group: %s", username, name)

					input := &iam.RemoveUserFromGroupInput{
						UserName:  user.UserName,
						GroupName: group.GroupName,
					}

					_, err := conn.RemoveUserFromGroup(ctx, input)

					if errs.IsA[*awstypes.NoSuchEntityException](err) {
						continue
					}

					if err != nil {
						sweeperErr := fmt.Errorf("error removing IAM User (%s) from IAM Group (%s): %w", username, name, err)
						log.Printf("[ERROR] %s", sweeperErr)
						sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
						continue
					}
				}
			}

			input := &iam.DeleteGroupInput{
				GroupName: group.GroupName,
			}

			if err := DeleteGroupPolicyAttachments(ctx, conn, name); err != nil {
				sweeperErr := fmt.Errorf("error deleting IAM Group (%s) policy attachments: %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}

			if err := DeleteGroupPolicies(ctx, conn, name); err != nil {
				sweeperErr := fmt.Errorf("error deleting IAM Group (%s) policies: %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}

			_, err = conn.DeleteGroup(ctx, input)

			if errs.IsA[*awstypes.NoSuchEntityException](err) {
				continue
			}

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting IAM Group (%s): %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepInstanceProfile(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.IAMClient(ctx)

	var sweepResources []sweep.Sweepable
	var sweeperErrs *multierror.Error

	pages := iam.NewListInstanceProfilesPaginator(conn, &iam.ListInstanceProfilesInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping IAM Instance Profile sweep: %s", err)
			return sweepResources, nil
		}

		for _, instanceProfile := range page.InstanceProfiles {
			name := aws.ToString(instanceProfile.InstanceProfileName)

			if !roleNameFilter(name) {
				tflog.Warn(ctx, "Skipping resource", map[string]any{
					"skip_reason":           "no match on allow-list",
					"instance_profile_name": name,
				})
				continue
			}

			r := resourceInstanceProfile()
			d := r.Data(nil)
			d.SetId(name)

			roles := instanceProfile.Roles
			if r := len(roles); r > 1 {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("unexpected number of roles for IAM Instance Profile (%s): %d", name, r))
			} else if r == 1 {
				d.Set("role", roles[0].RoleName)
			}

			sweepResources = append(sweepResources, sdk.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepOpenIDConnectProvider(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.IAMClient(ctx)

	var sweepResources []sweep.Sweepable

	out, err := conn.ListOpenIDConnectProviders(ctx, &iam.ListOpenIDConnectProvidersInput{})
	if err != nil {
		return sweepResources, err
	}

	for _, oidcProvider := range out.OpenIDConnectProviderList {
		arn := aws.ToString(oidcProvider.Arn)

		r := resourceOpenIDConnectProvider()
		d := r.Data(nil)
		d.SetId(arn)

		sweepResources = append(sweepResources, sdk.NewSweepResource(r, d, client))
	}

	return sweepResources, err
}

func sweepServiceSpecificCredentials(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.IAMClient(ctx)

	prefixes := []string{
		"test-user",
		"test_user",
		"tf-acc",
		"tf_acc",
	}

	var users []awstypes.User

	pages := iam.NewListUsersPaginator(conn, &iam.ListUsersInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, user := range page.Users {
			for _, prefix := range prefixes {
				if strings.HasPrefix(aws.ToString(user.UserName), prefix) {
					users = append(users, user)
					break
				}
			}
		}
	}

	var sweepResources []sweep.Sweepable

	for _, user := range users {
		out, err := conn.ListServiceSpecificCredentials(ctx, &iam.ListServiceSpecificCredentialsInput{
			UserName: user.UserName,
		})
		if err != nil {
			tflog.Warn(ctx, "Skipping resource", map[string]any{
				"error":     err.Error(),
				"user_name": user.UserName,
			})
			continue
		}

		for _, cred := range out.ServiceSpecificCredentials {
			id := fmt.Sprintf("%s:%s:%s", aws.ToString(cred.ServiceName), aws.ToString(cred.UserName), aws.ToString(cred.ServiceSpecificCredentialId))

			r := resourceServiceSpecificCredential()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sdk.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepPolicies(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.IAMClient(ctx)
	input := &iam.ListPoliciesInput{
		Scope: awstypes.PolicyScopeTypeLocal,
	}

	var sweepResources []sweep.Sweepable

	pages := iam.NewListPoliciesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping IAM Policy sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("retrieving IAM Policies: %w", err)
		}

		for _, policy := range page.Policies {
			arn := aws.ToString(policy.Arn)

			if n := aws.ToInt32(policy.AttachmentCount); n > 0 {
				log.Printf("[INFO] Skipping IAM Policy %s: AttachmentCount=%d", arn, n)
				continue
			}

			r := resourcePolicy()
			d := r.Data(nil)
			d.SetId(arn)

			sweepResources = append(sweepResources, newPolicySweeper(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)
	if err != nil {
		return fmt.Errorf("sweeping IAM Policies (%s): %w", region, err)
	}

	return nil
}

type policySweeper struct {
	d         *schema.ResourceData
	sweepable sweep.Sweepable
}

func newPolicySweeper(resource *schema.Resource, d *schema.ResourceData, client *conns.AWSClient) *policySweeper {
	return &policySweeper{
		d:         d,
		sweepable: sdk.NewSweepResource(resource, d, client),
	}
}

func (ps policySweeper) Delete(ctx context.Context, timeout time.Duration, optFns ...tfresource.OptionsFunc) error {
	if err := ps.sweepable.Delete(ctx, timeout, optFns...); err != nil {
		accessDenied := regexache.MustCompile(`AccessDenied: .+ with an explicit deny`)
		if accessDenied.MatchString(err.Error()) {
			log.Printf("[DEBUG] Skipping IAM Policy (%s): %s", ps.d.Id(), err)
			return nil
		}
		return err
	}
	return nil
}

func sweepRoles(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.IAMClient(ctx)

	roles := make([]string, 0)
	pages := iam.NewListRolesPaginator(conn, &iam.ListRolesInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping IAM Role sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("retrieving IAM Roles: %w", err)
		}

		for _, role := range page.Roles {
			roleName := aws.ToString(role.RoleName)
			if roleNameFilter(roleName) {
				roles = append(roles, roleName)
			} else {
				log.Printf("[INFO] Skipping IAM Role (%s): no match on allow-list", roleName)
			}
		}
	}

	if len(roles) == 0 {
		log.Print("[DEBUG] No IAM Roles to sweep")
		return nil
	}

	var sweeperErrs *multierror.Error

	for _, roleName := range roles {
		log.Printf("[DEBUG] Deleting IAM Role (%s)", roleName)

		err := deleteRole(ctx, conn, roleName, true, true, true)

		if tfawserr.ErrCodeContains(err, "AccessDenied") {
			log.Printf("[WARN] Skipping IAM Role (%s): %s", roleName, err)
			continue
		}
		if err != nil {
			sweeperErr := fmt.Errorf("error deleting IAM Role (%s): %w", roleName, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepSAMLProvider(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.IAMClient(ctx)

	var sweepResources []sweep.Sweepable

	out, err := conn.ListSAMLProviders(ctx, &iam.ListSAMLProvidersInput{})
	if err != nil {
		return sweepResources, err
	}

	for _, sampProvider := range out.SAMLProviderList {
		arn := aws.ToString(sampProvider.Arn)

		r := resourceSAMLProvider()
		d := r.Data(nil)
		d.SetId(arn)

		sweepResources = append(sweepResources, sdk.NewSweepResource(r, d, client))
	}

	return sweepResources, err
}

func sweepServerCertificates(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.IAMClient(ctx)

	pages := iam.NewListServerCertificatesPaginator(conn, &iam.ListServerCertificatesInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping IAM Server Certificate sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving IAM Server Certificates: %s", err)
		}

		for _, sc := range page.ServerCertificateMetadataList {
			log.Printf("[INFO] Deleting IAM Server Certificate: %s", aws.ToString(sc.ServerCertificateName))

			_, err := conn.DeleteServerCertificate(ctx, &iam.DeleteServerCertificateInput{
				ServerCertificateName: sc.ServerCertificateName,
			})
			if err != nil {
				log.Printf("[ERROR] Failed to delete IAM Server Certificate %s: %s",
					aws.ToString(sc.ServerCertificateName), err)
				continue
			}
		}
	}

	return nil
}

func sweepServiceLinkedRoles(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.IAMClient(ctx)

	var sweepResources []sweep.Sweepable

	input := &iam.ListRolesInput{
		PathPrefix: aws.String("/aws-service-role/"),
	}

	// include generic service role names created by:
	// TestAccIAMServiceLinkedRole_basic
	// TestAccIAMServiceLinkedRole_CustomSuffix_diffSuppressFunc
	customSuffixRegex := regexache.MustCompile(`_?(tf-acc-test-\d+|ServiceRoleForApplicationAutoScaling_CustomResource)$`)
	pages := iam.NewListRolesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping IAM Service Linked Role sweep: %s", err)
			return sweepResources, nil
		}

		if err != nil {
			return sweepResources, err
		}

		for _, role := range page.Roles {
			roleName := aws.ToString(role.RoleName)

			if !customSuffixRegex.MatchString(roleName) {
				tflog.Warn(ctx, "Skipping resource", map[string]any{
					"skip_reason": "no match",
					"role_name":   roleName,
				})
				continue
			}

			r := resourceServiceLinkedRole()
			d := r.Data(nil)
			d.SetId(aws.ToString(role.Arn))

			sweepResources = append(sweepResources, sdk.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepUsers(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.IAMClient(ctx)
	prefixes := []string{
		"test-user",
		"test_user",
		"tf-acc",
		"tf_acc",
	}

	var sweepResources []sweep.Sweepable

	pages := iam.NewListUsersPaginator(conn, &iam.ListUsersInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping IAM User sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("retrieving IAM Users: %s", err)
		}

		for _, user := range page.Users {
			for _, prefix := range prefixes {
				if strings.HasPrefix(aws.ToString(user.UserName), prefix) {
					r := resourceUser()
					d := r.Data(nil)
					d.SetId(aws.ToString(user.UserName))
					d.Set("force_destroy", true)

					// In general, sweeping should use the resource's Delete function. If Delete
					// is missing something that affects sweeping, fix Delete. Most of the time,
					// if something in Delete is causing sweep problems, it's also affecting
					// some users when they destroy.
					sweepResources = append(sweepResources, sdk.NewSweepResource(r, d, client))
					break
				}
			}
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)
	if err != nil {
		return fmt.Errorf("sweeping IAM Users (%s): %w", region, err)
	}

	return nil
}

func roleNameFilter(name string) bool {
	// The standard naming pattern for resources is generated by sdkacctest.RandomWithPrefix(acctest.ResourcePrefix).
	// Some roles automatically generated by AWS will add a prefix to the associated resource name, so look for
	// this pattern anywhere in the name, not just as a prefix.
	// Some names use "tf_acc_test" instead, so catch those, too.
	standardNameRegexp := regexache.MustCompile(`tf[-_]acc[-_]test`)
	if standardNameRegexp.MatchString(name) {
		return true
	}

	// Some acceptance tests use sdkacctest.RandString(10) rather than sdkacctest.RandomWithPrefix()
	// Others use other lengths, e.g. sdkacctest.RandString(8), but this one is risky enough, so leave it as-is
	randString10 := regexache.MustCompile(`^[0-9A-Za-z]{10}$`)
	if randString10.MatchString(name) {
		return true
	}

	randTF := regexache.MustCompile(`^tf-[0-9]{16}`)
	if randTF.MatchString(name) {
		return true
	}

	// We have a lot of role name prefixes for role names that don't match the standard pattern. This is not an
	// exhaustive list.
	prefixes := []string{
		"another_rds",
		"AmazonComprehendServiceRole-",
		"aws_batch_service_role",
		"aws_elastictranscoder_pipeline_tf_test",
		"batch_tf_batch_target-",
		"codebuild-",
		"codepipeline-",
		"cognito_authenticated_",
		"cognito_unauthenticated_",
		"CWLtoKinesisRole_",
		"EMR_AutoScaling_DefaultRole_",
		"es-domain-role-",
		"event_",
		"foobar",
		"iam_emr",
		"iam_for_sfn",
		"KinesisFirehoseServiceRole-test",
		"rds",
		"resource-test-terraform-",
		"role",
		"sns-delivery-status",
		"ssm_role",
		"ssm-role",
		"terraform-2021",
		"terraform-2022",
		"test",
		"tf_ecs_target",
		"tf_ecs",
		"tf_test",
		"tf-acc",
		"tf-iam-role-replication",
		"tf-opsworks-acc",
		"tf-test-iam",
		"tf-test",
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}

	return false
}

func sweepVirtualMFADevice(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.IAMClient(ctx)
	var sweepResources []sweep.Sweepable

	input := &iam.ListVirtualMFADevicesInput{}

	pages := iam.NewListVirtualMFADevicesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping IAM Virtual MFA Device sweep: %s", err)
			return sweepResources, nil
		}

		if err != nil {
			return nil, err
		}

		for _, device := range page.VirtualMFADevices {
			serialNum := aws.ToString(device.SerialNumber)

			if strings.Contains(serialNum, "root-account-mfa-device") {
				log.Printf("[INFO] Skipping IAM Root Virtual MFA Device: %s", serialNum)
				continue
			}

			r := resourceVirtualMFADevice()
			d := r.Data(nil)
			d.SetId(serialNum)

			sweepResources = append(sweepResources, sdk.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepSigningCertificates(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.IAMClient(ctx)

	prefixes := []string{
		"test-user",
		"test_user",
		"tf-acc",
		"tf_acc",
	}

	var users []awstypes.User

	pages := iam.NewListUsersPaginator(conn, &iam.ListUsersInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, user := range page.Users {
			for _, prefix := range prefixes {
				if strings.HasPrefix(aws.ToString(user.UserName), prefix) {
					users = append(users, user)
					break
				}
			}
		}
	}

	var sweepResources []sweep.Sweepable

	for _, user := range users {
		out, err := conn.ListSigningCertificates(ctx, &iam.ListSigningCertificatesInput{
			UserName: user.UserName,
		})
		if err != nil {
			tflog.Warn(ctx, "Skipping resource", map[string]any{
				"error":     err.Error(),
				"user_name": user.UserName,
			})
			continue
		}

		for _, cert := range out.Certificates {
			id := fmt.Sprintf("%s:%s", aws.ToString(cert.CertificateId), aws.ToString(cert.UserName))

			r := resourceSigningCertificate()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sdk.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}
