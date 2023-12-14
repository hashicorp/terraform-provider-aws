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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv1"
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

	conn := client.IAMConn(ctx)
	input := &iam.ListGroupsInput{}
	var sweeperErrs *multierror.Error

	err = conn.ListGroupsPagesWithContext(ctx, input, func(page *iam.ListGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, group := range page.Groups {
			name := aws.StringValue(group.GroupName)

			if name == "Admin" || name == "TerraformAccTests" {
				continue
			}

			log.Printf("[INFO] Deleting IAM Group: %s", name)

			getGroupInput := &iam.GetGroupInput{
				GroupName: group.GroupName,
			}

			getGroupOutput, err := conn.GetGroupWithContext(ctx, getGroupInput)

			if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
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
					username := aws.StringValue(user.UserName)

					log.Printf("[INFO] Removing IAM User (%s) from Group: %s", username, name)

					input := &iam.RemoveUserFromGroupInput{
						UserName:  user.UserName,
						GroupName: group.GroupName,
					}

					_, err := conn.RemoveUserFromGroupWithContext(ctx, input)

					if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
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

			_, err = conn.DeleteGroupWithContext(ctx, input)

			if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
				continue
			}

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting IAM Group (%s): %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping IAM Group sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving IAM Groups: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepInstanceProfile(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.IAMConn(ctx)

	var sweepResources []sweep.Sweepable
	var sweeperErrs *multierror.Error

	err := conn.ListInstanceProfilesPagesWithContext(ctx, &iam.ListInstanceProfilesInput{}, func(page *iam.ListInstanceProfilesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, instanceProfile := range page.InstanceProfiles {
			name := aws.StringValue(instanceProfile.InstanceProfileName)

			if !roleNameFilter(name) {
				tflog.Warn(ctx, "Skipping resource", map[string]any{
					"skip_reason":           "no match on allow-list",
					"instance_profile_name": name,
				})
				continue
			}

			r := ResourceInstanceProfile()
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

		return !lastPage
	})

	return sweepResources, multierror.Append(err, sweeperErrs).ErrorOrNil()
}

func sweepOpenIDConnectProvider(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.IAMConn(ctx)

	var sweepResources []sweep.Sweepable

	out, err := conn.ListOpenIDConnectProvidersWithContext(ctx, &iam.ListOpenIDConnectProvidersInput{})
	if err != nil {
		return sweepResources, err
	}

	for _, oidcProvider := range out.OpenIDConnectProviderList {
		arn := aws.StringValue(oidcProvider.Arn)

		r := ResourceOpenIDConnectProvider()
		d := r.Data(nil)
		d.SetId(arn)

		sweepResources = append(sweepResources, sdk.NewSweepResource(r, d, client))
	}

	return sweepResources, err
}

func sweepServiceSpecificCredentials(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.IAMConn(ctx)

	prefixes := []string{
		"test-user",
		"test_user",
		"tf-acc",
		"tf_acc",
	}

	var users []*iam.User

	err := conn.ListUsersPagesWithContext(ctx, &iam.ListUsersInput{}, func(page *iam.ListUsersOutput, lastPage bool) bool {
		for _, user := range page.Users {
			for _, prefix := range prefixes {
				if strings.HasPrefix(aws.StringValue(user.UserName), prefix) {
					users = append(users, user)
					break
				}
			}
		}

		return !lastPage
	})
	if err != nil {
		return nil, err
	}

	var sweepResources []sweep.Sweepable

	for _, user := range users {
		out, err := conn.ListServiceSpecificCredentialsWithContext(ctx, &iam.ListServiceSpecificCredentialsInput{
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
			id := fmt.Sprintf("%s:%s:%s", aws.StringValue(cred.ServiceName), aws.StringValue(cred.UserName), aws.StringValue(cred.ServiceSpecificCredentialId))

			r := ResourceServiceSpecificCredential()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sdk.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, err
}

func sweepPolicies(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.IAMConn(ctx)
	input := &iam.ListPoliciesInput{
		Scope: aws.String(iam.PolicyScopeTypeLocal),
	}

	var sweepResources []sweep.Sweepable

	err = conn.ListPoliciesPagesWithContext(ctx, input, func(page *iam.ListPoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Policies {
			arn := aws.StringValue(v.Arn)

			if n := aws.Int64Value(v.AttachmentCount); n > 0 {
				log.Printf("[INFO] Skipping IAM Policy %s: AttachmentCount=%d", arn, n)
				continue
			}

			r := ResourcePolicy()
			d := r.Data(nil)
			d.SetId(arn)

			sweepResources = append(sweepResources, newPolicySweeper(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping IAM Policy sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("retrieving IAM Policies: %w", err)
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
	conn := client.IAMConn(ctx)

	roles := make([]string, 0)
	err = conn.ListRolesPagesWithContext(ctx, &iam.ListRolesInput{}, func(page *iam.ListRolesOutput, lastPage bool) bool {
		for _, role := range page.Roles {
			roleName := aws.StringValue(role.RoleName)
			if roleNameFilter(roleName) {
				roles = append(roles, roleName)
			} else {
				log.Printf("[INFO] Skipping IAM Role (%s): no match on allow-list", roleName)
			}
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping IAM Role sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error retrieving IAM Roles: %w", err)
	}

	if len(roles) == 0 {
		log.Print("[DEBUG] No IAM Roles to sweep")
		return nil
	}

	var sweeperErrs *multierror.Error

	for _, roleName := range roles {
		log.Printf("[DEBUG] Deleting IAM Role (%s)", roleName)

		err := DeleteRole(ctx, conn, roleName, true, true, true)

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
	conn := client.IAMConn(ctx)

	var sweepResources []sweep.Sweepable

	out, err := conn.ListSAMLProvidersWithContext(ctx, &iam.ListSAMLProvidersInput{})
	if err != nil {
		return sweepResources, err
	}

	for _, sampProvider := range out.SAMLProviderList {
		arn := aws.StringValue(sampProvider.Arn)

		r := ResourceSAMLProvider()
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
	conn := client.IAMConn(ctx)

	err = conn.ListServerCertificatesPagesWithContext(ctx, &iam.ListServerCertificatesInput{}, func(out *iam.ListServerCertificatesOutput, lastPage bool) bool {
		for _, sc := range out.ServerCertificateMetadataList {
			log.Printf("[INFO] Deleting IAM Server Certificate: %s", *sc.ServerCertificateName)

			_, err := conn.DeleteServerCertificateWithContext(ctx, &iam.DeleteServerCertificateInput{
				ServerCertificateName: sc.ServerCertificateName,
			})
			if err != nil {
				log.Printf("[ERROR] Failed to delete IAM Server Certificate %s: %s",
					*sc.ServerCertificateName, err)
				continue
			}
		}
		return !lastPage
	})
	if err != nil {
		if awsv1.SkipSweepError(err) {
			log.Printf("[WARN] Skipping IAM Server Certificate sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving IAM Server Certificates: %s", err)
	}

	return nil
}

func sweepServiceLinkedRoles(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.IAMConn(ctx)

	var sweepResources []sweep.Sweepable

	input := &iam.ListRolesInput{
		PathPrefix: aws.String("/aws-service-role/"),
	}

	// include generic service role names created by:
	// TestAccIAMServiceLinkedRole_basic
	// TestAccIAMServiceLinkedRole_CustomSuffix_diffSuppressFunc
	customSuffixRegex := regexache.MustCompile(`_?(tf-acc-test-\d+|ServiceRoleFor(ApplicationAutoScaling_CustomResource|ElasticBeanstalk))$`)
	err := conn.ListRolesPagesWithContext(ctx, input, func(page *iam.ListRolesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, role := range page.Roles {
			roleName := aws.StringValue(role.RoleName)

			if !customSuffixRegex.MatchString(roleName) {
				tflog.Warn(ctx, "Skipping resource", map[string]any{
					"skip_reason": "no match",
					"role_name":   roleName,
				})
				continue
			}

			r := ResourceServiceLinkedRole()
			d := r.Data(nil)
			d.SetId(aws.StringValue(role.Arn))

			sweepResources = append(sweepResources, sdk.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	return sweepResources, err
}

func sweepUsers(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.IAMConn(ctx)
	prefixes := []string{
		"test-user",
		"test_user",
		"tf-acc",
		"tf_acc",
	}
	users := make([]*iam.User, 0)

	err = conn.ListUsersPagesWithContext(ctx, &iam.ListUsersInput{}, func(page *iam.ListUsersOutput, lastPage bool) bool {
		for _, user := range page.Users {
			for _, prefix := range prefixes {
				if strings.HasPrefix(aws.StringValue(user.UserName), prefix) {
					users = append(users, user)
					break
				}
			}
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping IAM User sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error retrieving IAM Users: %s", err)
	}

	if len(users) == 0 {
		log.Print("[DEBUG] No IAM Users to sweep")
		return nil
	}

	var sweeperErrs *multierror.Error
	for _, user := range users {
		username := aws.StringValue(user.UserName)
		log.Printf("[DEBUG] Deleting IAM User: %s", username)

		listUserPoliciesInput := &iam.ListUserPoliciesInput{
			UserName: user.UserName,
		}
		listUserPoliciesOutput, err := conn.ListUserPoliciesWithContext(ctx, listUserPoliciesInput)

		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			continue
		}
		if err != nil {
			sweeperErr := fmt.Errorf("error listing IAM User (%s) inline policies: %s", username, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}

		for _, inlinePolicyName := range listUserPoliciesOutput.PolicyNames {
			log.Printf("[DEBUG] Deleting IAM User (%s) inline policy %q", username, *inlinePolicyName)

			input := &iam.DeleteUserPolicyInput{
				PolicyName: inlinePolicyName,
				UserName:   user.UserName,
			}

			if _, err := conn.DeleteUserPolicyWithContext(ctx, input); err != nil {
				if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
					continue
				}
				sweeperErr := fmt.Errorf("error deleting IAM User (%s) inline policy %q: %s", username, *inlinePolicyName, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		listAttachedUserPoliciesInput := &iam.ListAttachedUserPoliciesInput{
			UserName: user.UserName,
		}
		listAttachedUserPoliciesOutput, err := conn.ListAttachedUserPoliciesWithContext(ctx, listAttachedUserPoliciesInput)

		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			continue
		}
		if err != nil {
			sweeperErr := fmt.Errorf("error listing IAM User (%s) attached policies: %s", username, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}

		for _, attachedPolicy := range listAttachedUserPoliciesOutput.AttachedPolicies {
			policyARN := aws.StringValue(attachedPolicy.PolicyArn)

			log.Printf("[DEBUG] Detaching IAM User (%s) attached policy: %s", username, policyARN)

			if err := detachPolicyFromUser(ctx, conn, username, policyARN); err != nil {
				sweeperErr := fmt.Errorf("error detaching IAM User (%s) attached policy (%s): %s", username, policyARN, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		if err := DeleteUserGroupMemberships(ctx, conn, username); err != nil {
			sweeperErr := fmt.Errorf("error removing IAM User (%s) group memberships: %s", username, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}

		if err := DeleteUserAccessKeys(ctx, conn, username); err != nil {
			sweeperErr := fmt.Errorf("error removing IAM User (%s) access keys: %s", username, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}

		if err := DeleteUserSSHKeys(ctx, conn, username); err != nil {
			sweeperErr := fmt.Errorf("error removing IAM User (%s) SSH keys: %s", username, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}

		if err := DeleteUserVirtualMFADevices(ctx, conn, username); err != nil {
			sweeperErr := fmt.Errorf("error removing IAM User (%s) virtual MFA devices: %s", username, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}

		if err := DeactivateUserMFADevices(ctx, conn, username); err != nil {
			sweeperErr := fmt.Errorf("error removing IAM User (%s) MFA devices: %s", username, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}

		if err := DeleteUserLoginProfile(ctx, conn, username); err != nil {
			sweeperErr := fmt.Errorf("error removing IAM User (%s) login profile: %s", username, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}

		input := &iam.DeleteUserInput{
			UserName: aws.String(username),
		}

		_, err = conn.DeleteUserWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			continue
		}
		if err != nil {
			sweeperErr := fmt.Errorf("error deleting IAM User (%s): %s", username, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}
	}

	return sweeperErrs.ErrorOrNil()
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
	conn := client.IAMConn(ctx)
	var sweepResources []sweep.Sweepable

	input := &iam.ListVirtualMFADevicesInput{}

	err := conn.ListVirtualMFADevicesPagesWithContext(ctx, input, func(page *iam.ListVirtualMFADevicesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, device := range page.VirtualMFADevices {
			serialNum := aws.StringValue(device.SerialNumber)

			if strings.Contains(serialNum, "root-account-mfa-device") {
				tflog.Warn(ctx, "Skipping: IAM Root Virtual MFA Device", map[string]any{
					"serial_number": device,
				})
				continue
			}

			r := ResourceVirtualMFADevice()
			d := r.Data(nil)
			d.SetId(serialNum)

			sweepResources = append(sweepResources, sdk.NewSweepResource(r, d, client))
		}
		return !lastPage
	})

	return sweepResources, err
}

func sweepSigningCertificates(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.IAMConn(ctx)

	prefixes := []string{
		"test-user",
		"test_user",
		"tf-acc",
		"tf_acc",
	}

	var users []*iam.User

	err := conn.ListUsersPagesWithContext(ctx, &iam.ListUsersInput{}, func(page *iam.ListUsersOutput, lastPage bool) bool {
		for _, user := range page.Users {
			for _, prefix := range prefixes {
				if strings.HasPrefix(aws.StringValue(user.UserName), prefix) {
					users = append(users, user)
					break
				}
			}
		}

		return !lastPage
	})
	if err != nil {
		return nil, err
	}

	var sweepResources []sweep.Sweepable

	for _, user := range users {
		out, err := conn.ListSigningCertificatesWithContext(ctx, &iam.ListSigningCertificatesInput{
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
			id := fmt.Sprintf("%s:%s", aws.StringValue(cert.CertificateId), aws.StringValue(cert.UserName))

			r := ResourceSigningCertificate()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sdk.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, err
}
