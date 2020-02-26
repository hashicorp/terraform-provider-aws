package a

import (
	"fmt"
	"regexp"
)

const resourceName = `aws_example_thing.test`

var testRegexp = regexp.MustCompile(`.*`)

func f() {
	/* Passing cases */

	fmt.Sprintf("%s.notamazonaws.com", "test")

	/* Comment ignored cases */

	//lintignore:AWSR001
	fmt.Sprintf("%s.amazonaws.com", "test")

	fmt.Sprintf("%s.amazonaws.com", "test") //lintignore:AWSR001

	/* Failing cases */

	fmt.Sprintf("%s.amazonaws.com", "test") // want "prefer \\(\\*AWSClient\\).PartitionHostname\\(\\) or \\(\\*AWSClient\\).RegionalHostname\\(\\)"
}
