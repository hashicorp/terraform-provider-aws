// Package self
// a program that complete itself
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/posener/complete"
)

func main() {

	// add a variable to the program
	var name string
	flag.StringVar(&name, "name", "", "Give your name")

	// create the complete command
	cmp := complete.New(
		"self",
		complete.Command{Flags: complete.Flags{"-name": complete.PredictAnything}},
	)

	// AddFlags adds the completion flags to the program flags,
	// in case of using non-default flag set, it is possible to pass
	// it as an argument.
	// it is possible to set custom flags name
	// so when one will type 'self -h', he will see '-complete' to install the
	// completion and -uncomplete to uninstall it.
	cmp.CLI.InstallName = "complete"
	cmp.CLI.UninstallName = "uncomplete"
	cmp.AddFlags(nil)

	// parse the flags - both the program's flags and the completion flags
	flag.Parse()

	// run the completion, in case that the completion was invoked
	// and ran as a completion script or handled a flag that passed
	// as argument, the Run method will return true,
	// in that case, our program have nothing to do and should return.
	if cmp.Complete() {
		return
	}

	// if the completion did not do anything, we can run our program logic here.
	if name == "" {
		fmt.Println("Your name is missing")
		os.Exit(1)
	}

	fmt.Println("Hi,", name)
}
