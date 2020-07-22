package a

import (
	"fmt"
)

func f() {
	/* Passing cases */
	fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role_policy_attachment" "test-AmazonEKSClusterPolicy" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.test.%s
}
`, "moln")

	/* Comment ignored cases */

	//lintignore:AWSAT008
	fmt.Sprintf(`policy_arn = "arn:aws:iam::aws:policy/%s"`, "AmazonEKSClusterPolicy")

	fmt.Sprintf(`policy_arn = "arn:aws:iam::aws:policy/%s"`, "AmazonEKSClusterPolicy") //lintignore:AWSAT008
	fmt.Sprintf(`policy_arn = "arn:aws:iam::aws:policy/%s"`, "AmazonEKSClusterPolicy") //lintignore:AWSAT008

	/* Failing cases */

	fmt.Sprintf(`policy_arn = "arn:aws:iam::aws:policy/%s"`, "AmazonEKSClusterPolicy") // want "avoid hardcoded ARN AWS partitions"
}
