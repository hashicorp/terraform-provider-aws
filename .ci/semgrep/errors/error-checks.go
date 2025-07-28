package main

import (
	"context"
	"errors"
	"time"

	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func test1() {
	_, err := call()

	// ruleid: notfound-without-err-checks
	if tfresource.NotFound(err) {
		return
	}

	return
}

func test2() {
	_, err := call()

	// ok: notfound-without-err-checks
	if tfresource.NotFound(err) {
		return
	}

	if err != nil {
		return
	}

	return
}

func test3() {
	_, err := call()

	if err != nil {
		// ok: notfound-without-err-checks
		if tfresource.NotFound(err) {
			return
		}
		return
	}

	return
}

func test4() {
	_, err := call()

	if err == nil {
		return
		// ok: notfound-without-err-checks
	} else if tfresource.NotFound(err) {
		return
	} else {
		return
	}
}

func test5() {
	_, err := call()

	// ok: notfound-without-err-checks
	if tfresource.NotFound(err) {
		return
	} else if err != nil {
		return
	} else {
		return
	}
}

func test6() error {
	_, err := call()

	// ok: notfound-without-err-checks
	if tfresource.NotFound(err) {
		return
	}

	return err
}

func test7() {
	for {
		_, err := call()

		// ok: notfound-without-err-checks
		if tfresource.NotFound(err) {
			continue
		}
	}

	return err
}

func test8() {
	_, err := call()

	// ok: notfound-without-err-checks
	if tfresource.NotFound(err) {
		return
	}

	if tfawserr.ErrCodeEquals(err, "SomeError") {
		return
	}

	if err != nil {
		return
	}

	return
}

func test9() {
	_, err := call()

	if err != nil {
		// ok: notfound-without-err-checks
		if tfresource.NotFound(err) {
			return
		} else {
			return
		}
	}

	return
}

func test10() {
	_, err := call()

	// ok: notfound-without-err-checks
	if tfresource.NotFound(err) {
		return
	} else if err != nil {
		return
	}

	return
}

func test11() {
	ctx := context.Background()

	tfresource.RetryWhen(ctx, 1*time.Second, nil, func(err error) (bool error) {
		// ok: notfound-without-err-checks
		if tfresource.NotFound(err) {
			return true, err
		}

		return false, err
	})
}

func test12() {
	_, err := call()

	// ok: notfound-without-err-checks
	if tfresource.NotFound(err) {
		return
	}

	if PreCheckSkipError(err) {
		return
	}

	if err != nil {
		return
	}

	return
}

func call() (any, error) {
	return nil, errors.New("error")
}
