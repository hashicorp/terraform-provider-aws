//go:build sweep
// +build sweep

package iam

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_iam_group", &resource.Sweeper{
		Name: "aws_iam_group",
		F:    sweepGroups,
		Dependencies: []string{
			"aws_iam_user",
		},
	})

	resource.AddTestSweepers("aws_iam_instance_profile", &resource.Sweeper{
		Name:         "aws_iam_instance_profile",
		F:            sweepInstanceProfile,
		Dependencies: []string{"aws_iam_role"},
	})

	resource.AddTestSweepers("aws_iam_openid_connect_provider", &resource.Sweeper{
		Name: "aws_iam_openid_connect_provider",
		F:    sweepOpenIDConnectProvider,
	})

	resource.AddTestSweepers("aws_iam_policy", &resource.Sweeper{
		Name: "aws_iam_policy",
		F:    sweepPolicies,
		Dependencies: []string{
			"aws_iam_group",
			"aws_iam_role",
			"aws_iam_user",
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
			"aws_datasync_location_s3",
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
			"aws_lambda_function",
			"aws_launch_configuration",
			"aws_redshift_cluster",
			"aws_redshift_scheduled_action",
			"aws_spot_fleet_request",
		},
		F: sweepRoles,
	})

	resource.AddTestSweepers("aws_iam_saml_provider", &resource.Sweeper{
		Name: "aws_iam_saml_provider",
		F:    sweepSAMLProvider,
	})

	resource.AddTestSweepers("aws_iam_service_specific_credential", &resource.Sweeper{
		Name: "aws_iam_service_specific_credential",
		F:    sweepServiceSpecificCredentials,
	})

	resource.AddTestSweepers("aws_iam_signing_certificate", &resource.Sweeper{
		Name: "aws_iam_signing_certificate",
		F:    sweepSigningCertificates,
	})

	resource.AddTestSweepers("aws_iam_server_certificate", &resource.Sweeper{
		Name: "aws_iam_server_certificate",
		F:    sweepServerCertificates,
	})

	resource.AddTestSweepers("aws_iam_service_linked_role", &resource.Sweeper{
		Name: "aws_iam_service_linked_role",
		F:    sweepServiceLinkedRoles,
	})

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

	resource.AddTestSweepers("aws_iam_virtual_mfa_device", &resource.Sweeper{
		Name: "aws_iam_virtual_mfa_device",
		F:    sweepVirtualMFADevice,
	})
}

func sweepGroups(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).IAMConn
	input := &iam.ListGroupsInput{}
	var sweeperErrs *multierror.Error

	err = conn.ListGroupsPages(input, func(page *iam.ListGroupsOutput, lastPage bool) bool {
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

			getGroupOutput, err := conn.GetGroup(getGroupInput)

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

					_, err := conn.RemoveUserFromGroup(input)

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

			if err := DeleteGroupPolicyAttachments(conn, name); err != nil {
				sweeperErr := fmt.Errorf("error deleting IAM Group (%s) policy attachments: %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}

			if err := DeleteGroupPolicies(conn, name); err != nil {
				sweeperErr := fmt.Errorf("error deleting IAM Group (%s) policies: %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}

			_, err = conn.DeleteGroup(input)

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

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping IAM Group sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving IAM Groups: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepInstanceProfile(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).IAMConn

	var sweeperErrs *multierror.Error

	err = conn.ListInstanceProfilesPages(&iam.ListInstanceProfilesInput{}, func(page *iam.ListInstanceProfilesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, instanceProfile := range page.InstanceProfiles {
			name := aws.StringValue(instanceProfile.InstanceProfileName)

			if !roleNameFilter(name) {
				log.Printf("[INFO] Skipping IAM Instance Profile (%s): no match on allow-list", name)
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

			log.Printf("[INFO] Sweeping IAM Instance Profile %q", name)
			err := sweep.DeleteResource(r, d, client)

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error deleting IAM Instance Profile (%s): %w", name, err))
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping IAM Instance Profile sweep for %q: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing IAM Instance Profiles: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepOpenIDConnectProvider(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).IAMConn

	var sweeperErrs *multierror.Error

	out, err := conn.ListOpenIDConnectProviders(&iam.ListOpenIDConnectProvidersInput{})

	for _, oidcProvider := range out.OpenIDConnectProviderList {
		arn := aws.StringValue(oidcProvider.Arn)

		r := ResourceOpenIDConnectProvider()
		d := r.Data(nil)
		d.SetId(arn)
		err := sweep.DeleteResource(r, d, client)

		if err != nil {
			sweeperErr := fmt.Errorf("error deleting IAM OIDC Provider (%s): %w", arn, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping IAM OIDC Provider sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error describing IAM OIDC Providers: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepServiceSpecificCredentials(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).IAMConn

	var sweeperErrs *multierror.Error

	prefixes := []string{
		"test-user",
		"test_user",
		"tf-acc",
		"tf_acc",
	}

	users := make([]*iam.User, 0)

	err = conn.ListUsersPages(&iam.ListUsersInput{}, func(page *iam.ListUsersOutput, lastPage bool) bool {
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

	for _, user := range users {
		out, err := conn.ListServiceSpecificCredentials(&iam.ListServiceSpecificCredentialsInput{
			UserName: user.UserName,
		})

		for _, cred := range out.ServiceSpecificCredentials {

			id := fmt.Sprintf("%s:%s:%s", aws.StringValue(cred.ServiceName), aws.StringValue(cred.UserName), aws.StringValue(cred.ServiceSpecificCredentialId))

			r := ResourceServiceSpecificCredential()
			d := r.Data(nil)
			d.SetId(id)
			err := sweep.DeleteResource(r, d, client)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting IAM Service Specific Credential (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping IAM Service Specific Credential sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error describing IAM Service Specific Credentials: %w", err))
		}
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepPolicies(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).IAMConn
	input := &iam.ListPoliciesInput{
		Scope: aws.String(iam.PolicyScopeTypeLocal),
	}
	var sweeperErrs *multierror.Error

	err = conn.ListPoliciesPages(input, func(page *iam.ListPoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, policy := range page.Policies {
			arn := aws.StringValue(policy.Arn)
			input := &iam.DeletePolicyInput{
				PolicyArn: policy.Arn,
			}

			log.Printf("[INFO] Deleting IAM Policy: %s", arn)
			if err := PolicyDeleteNondefaultVersions(arn, conn); err != nil {
				sweeperErr := fmt.Errorf("error deleting IAM Policy (%s) non-default versions: %w", arn, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}

			_, err := conn.DeletePolicy(input)

			// Treat this sweeper as best effort for now. There are a lot of edge cases
			// with lingering aws_iam_role resources in the HashiCorp testing accounts.
			if tfawserr.ErrCodeEquals(err, iam.ErrCodeDeleteConflictException) {
				log.Printf("[WARN] Ignoring IAM Policy (%s) deletion error: %s", arn, err)
				continue
			}

			if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
				continue
			}

			if tfawserr.ErrMessageContains(err, "AccessDenied", "with an explicit deny") {
				continue
			}

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting IAM Policy (%s): %w", arn, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping IAM Policy sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving IAM Policies: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepRoles(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).IAMConn

	roles := make([]string, 0)
	err = conn.ListRolesPages(&iam.ListRolesInput{}, func(page *iam.ListRolesOutput, lastPage bool) bool {
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

	if sweep.SkipSweepError(err) {
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

		err := DeleteRole(conn, roleName, true, true, true)
		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			continue
		}
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

func sweepSAMLProvider(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).IAMConn

	var sweeperErrs *multierror.Error

	out, err := conn.ListSAMLProviders(&iam.ListSAMLProvidersInput{})

	for _, sampProvider := range out.SAMLProviderList {
		arn := aws.StringValue(sampProvider.Arn)

		r := ResourceSAMLProvider()
		d := r.Data(nil)
		d.SetId(arn)
		err := sweep.DeleteResource(r, d, client)

		if err != nil {
			sweeperErr := fmt.Errorf("error deleting IAM SAML Provider (%s): %w", arn, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping IAM SAML Provider sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error describing IAM SAML Providers: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepServerCertificates(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).IAMConn

	err = conn.ListServerCertificatesPages(&iam.ListServerCertificatesInput{}, func(out *iam.ListServerCertificatesOutput, lastPage bool) bool {
		for _, sc := range out.ServerCertificateMetadataList {
			log.Printf("[INFO] Deleting IAM Server Certificate: %s", *sc.ServerCertificateName)

			_, err := conn.DeleteServerCertificate(&iam.DeleteServerCertificateInput{
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
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping IAM Server Certificate sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving IAM Server Certificates: %s", err)
	}

	return nil
}

func sweepServiceLinkedRoles(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).IAMConn
	var sweeperErrs *multierror.Error
	input := &iam.ListRolesInput{
		PathPrefix: aws.String("/aws-service-role/"),
	}

	// include generic service role names created by:
	// TestAccIAMServiceLinkedRole_basic
	// TestAccIAMServiceLinkedRole_CustomSuffix_diffSuppressFunc
	customSuffixRegex := regexp.MustCompile(`_?(tf-acc-test-\d+|ServiceRoleFor(ApplicationAutoScaling_CustomResource|ElasticBeanstalk))$`)
	err = conn.ListRolesPages(input, func(page *iam.ListRolesOutput, lastPage bool) bool {
		if len(page.Roles) == 0 {
			log.Printf("[INFO] No IAM Service Roles to sweep")
			return true
		}
		for _, role := range page.Roles {
			roleName := aws.StringValue(role.RoleName)

			if !customSuffixRegex.MatchString(roleName) {
				log.Printf("[INFO] Skipping IAM Service Role: %s", roleName)
				continue
			}

			r := ResourceServiceLinkedRole()
			d := r.Data(nil)
			d.SetId(aws.StringValue(role.Arn))
			err := sweep.DeleteResource(r, d, client)
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting IAM Service Linked Role (%s): %w", roleName, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}
		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping IAM Service Role sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error describing IAM Service Roles: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepUsers(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).IAMConn
	prefixes := []string{
		"test-user",
		"test_user",
		"tf-acc",
		"tf_acc",
	}
	users := make([]*iam.User, 0)

	err = conn.ListUsersPages(&iam.ListUsersInput{}, func(page *iam.ListUsersOutput, lastPage bool) bool {
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

	if sweep.SkipSweepError(err) {
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
		listUserPoliciesOutput, err := conn.ListUserPolicies(listUserPoliciesInput)

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

			if _, err := conn.DeleteUserPolicy(input); err != nil {
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
		listAttachedUserPoliciesOutput, err := conn.ListAttachedUserPolicies(listAttachedUserPoliciesInput)

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

			if err := DetachPolicyFromUser(conn, username, policyARN); err != nil {
				sweeperErr := fmt.Errorf("error detaching IAM User (%s) attached policy (%s): %s", username, policyARN, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		if err := DeleteUserGroupMemberships(conn, username); err != nil {
			sweeperErr := fmt.Errorf("error removing IAM User (%s) group memberships: %s", username, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}

		if err := DeleteUserAccessKeys(conn, username); err != nil {
			sweeperErr := fmt.Errorf("error removing IAM User (%s) access keys: %s", username, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}

		if err := DeleteUserSSHKeys(conn, username); err != nil {
			sweeperErr := fmt.Errorf("error removing IAM User (%s) SSH keys: %s", username, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}

		if err := DeleteUserVirtualMFADevices(conn, username); err != nil {
			sweeperErr := fmt.Errorf("error removing IAM User (%s) virtual MFA devices: %s", username, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}

		if err := DeactivateUserMFADevices(conn, username); err != nil {
			sweeperErr := fmt.Errorf("error removing IAM User (%s) MFA devices: %s", username, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}

		if err := DeleteUserLoginProfile(conn, username); err != nil {
			sweeperErr := fmt.Errorf("error removing IAM User (%s) login profile: %s", username, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}

		input := &iam.DeleteUserInput{
			UserName: aws.String(username),
		}

		_, err = conn.DeleteUser(input)

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
	standardNameRegexp := regexp.MustCompile(`tf[-_]acc[-_]test`)
	if standardNameRegexp.MatchString(name) {
		return true
	}

	// Some acceptance tests use sdkacctest.RandString(10) rather than sdkacctest.RandomWithPrefix()
	// Others use other lengths, e.g. sdkacctest.RandString(8), but this one is risky enough, so leave it as-is
	randString10 := regexp.MustCompile(`^[a-zA-Z0-9]{10}$`)
	if randString10.MatchString(name) {
		return true
	}

	// We have a lot of role name prefixes for role names that don't match the standard pattern. This is not an
	// exhaustive list.
	prefixes := []string{
		"another_rds",
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
		"rds",
		"role",
		"sns-delivery-status",
		"ssm_role",
		"ssm-role",
		"test",
		"tf-acc",
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}

	return false
}

func sweepVirtualMFADevice(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).IAMConn
	var sweeperErrs *multierror.Error
	input := &iam.ListVirtualMFADevicesInput{}

	err = conn.ListVirtualMFADevicesPages(input, func(page *iam.ListVirtualMFADevicesOutput, lastPage bool) bool {
		if len(page.VirtualMFADevices) == 0 {
			log.Printf("[INFO] No IAM Virtual MFA Devices to sweep")
			return true
		}
		for _, device := range page.VirtualMFADevices {
			serialNum := aws.StringValue(device.SerialNumber)

			if strings.Contains(serialNum, "root-account-mfa-device") {
				log.Printf("[INFO] Skipping IAM Root Virtual MFA Device: %s", device)
				continue
			}

			r := ResourceVirtualMFADevice()
			d := r.Data(nil)
			d.SetId(serialNum)
			err := sweep.DeleteResource(r, d, client)
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting IAM Virtual MFA Device (%s): %w", device, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}
		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping IAM Virtual MFA Device sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error describing IAM Virtual MFA Devices: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepSigningCertificates(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).IAMConn

	var sweeperErrs *multierror.Error

	prefixes := []string{
		"test-user",
		"test_user",
		"tf-acc",
		"tf_acc",
	}

	users := make([]*iam.User, 0)

	err = conn.ListUsersPages(&iam.ListUsersInput{}, func(page *iam.ListUsersOutput, lastPage bool) bool {
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

	for _, user := range users {
		out, err := conn.ListSigningCertificates(&iam.ListSigningCertificatesInput{
			UserName: user.UserName,
		})

		for _, cert := range out.Certificates {

			id := fmt.Sprintf("%s:%s", aws.StringValue(cert.CertificateId), aws.StringValue(cert.UserName))

			r := ResourceSigningCertificate()
			d := r.Data(nil)
			d.SetId(id)
			err := sweep.DeleteResource(r, d, client)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting IAM Signing Certificate (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping IAM Signing Certificate sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error describing IAM Signing Certificates: %w", err))
		}
	}

	return sweeperErrs.ErrorOrNil()
}
