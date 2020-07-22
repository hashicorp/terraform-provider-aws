# AWSAT008

The AWSAT008 analyzer reports hardcoded AWS partitions in ARNs. For tests
to work across AWS partitions, the partition should not be hardcoded.

## Flagged Code

```go
func testAccAWSMisericordiamHumilitatemPulchritudo(name string) string {
    return fmt.Sprintf(`
resource "aws_emr_cluster" "tf-test-cluster" {
  ec2_attributes {
    instance_profile = "arn:aws:iam::%s:instance-profile/EMR_EC2_DefaultRole"
  }

  service_role = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/EMR_DefaultRole"
}
`, r)
}
```

## Passing Code

```go
func testAccAWSMisericordiamHumilitatemPulchritudo(name string) string {
    return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role_policy_attachment" "test-AmazonEKSClusterPolicy" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.test.%s
}
`, name)
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:AWSAT008` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
policy_arn = "arn:aws:iam::2342342342324:role/EMR_DefaultRole" //lintignore:AWSAT008
```
