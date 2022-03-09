<!--- See what makes a good Pull Request at : https://github.com/hashicorp/terraform-provider-aws/blob/main/docs/contributing --->

<!--- If your PR fully resolves and should automatically close the linked issue, use Closes. Otherwise, use Relates --->
Relates OR Closes #0000

Output from acceptance testing:

<!--
Replace TestAccXXX with a pattern that matches the tests affected by this PR.

Replace ec2 with the service package corresponding to your tests.

For more information on the `-run` flag, see the `go test` documentation at https://tip.golang.org/cmd/go/#hdr-Testing_flags.
-->
```
$ make testacc TESTS=TestAccXXX PKG=ec2

...
```
