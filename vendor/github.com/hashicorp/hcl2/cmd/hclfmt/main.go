package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hclparse"
	"github.com/hashicorp/hcl2/hclwrite"
	"golang.org/x/crypto/ssh/terminal"
)

const versionStr = "0.0.1-dev"

var (
	check       = flag.Bool("check", false, "perform a syntax check on the given files and produce diagnostics")
	reqNoChange = flag.Bool("require-no-change", false, "return a non-zero status if any files are changed during formatting")
	overwrite   = flag.Bool("w", false, "overwrite source files instead of writing to stdout")
	showVersion = flag.Bool("version", false, "show the version number and immediately exit")
)

var parser = hclparse.NewParser()
var diagWr hcl.DiagnosticWriter // initialized in init
var checkErrs = false
var changed []string

func init() {
	color := terminal.IsTerminal(int(os.Stderr.Fd()))
	w, _, err := terminal.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		w = 80
	}
	diagWr = hcl.NewDiagnosticTextWriter(os.Stderr, parser.Files(), uint(w), color)
}

func main() {
	err := realmain()

	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func realmain() error {
	flag.Usage = usage
	flag.Parse()

	if *showVersion {
		fmt.Println(versionStr)
		return nil
	}

	err := processFiles()
	if err != nil {
		return err
	}

	if checkErrs {
		return errors.New("one or more files contained errors")
	}

	if *reqNoChange {
		if len(changed) != 0 {
			return fmt.Errorf("file(s) were changed: %s", strings.Join(changed, ", "))
		}
	}

	return nil
}

func processFiles() error {
	if flag.NArg() == 0 {
		if *overwrite {
			return errors.New("error: cannot use -w without source filenames")
		}

		return processFile("<stdin>", os.Stdin)
	}

	for i := 0; i < flag.NArg(); i++ {
		path := flag.Arg(i)
		switch dir, err := os.Stat(path); {
		case err != nil:
			return err
		case dir.IsDir():
			// This tool can't walk a whole directory because it doesn't
			// know what file naming schemes will be used by different
			// zcl-embedding applications, so it'll leave that sort of
			// functionality for apps themselves to implement.
			return fmt.Errorf("can't format directory %s", path)
		default:
			if err := processFile(path, nil); err != nil {
				return err
			}
		}
	}

	return nil
}

func processFile(fn string, in *os.File) error {
	var err error
	if in == nil {
		in, err = os.Open(fn)
		if err != nil {
			return fmt.Errorf("failed to open %s: %s", fn, err)
		}
	}

	inSrc, err := ioutil.ReadAll(in)
	if err != nil {
		return fmt.Errorf("failed to read %s: %s", fn, err)
	}

	if *check {
		_, diags := parser.ParseHCL(inSrc, fn)
		diagWr.WriteDiagnostics(diags)
		if diags.HasErrors() {
			checkErrs = true
			return nil
		}
	}

	outSrc := hclwrite.Format(inSrc)

	if !bytes.Equal(inSrc, outSrc) {
		changed = append(changed, fn)
	}

	if *overwrite {
		return ioutil.WriteFile(fn, outSrc, 0644)
	}

	_, err = os.Stdout.Write(outSrc)
	return err
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: zclfmt [flags] [path ...]\n")
	flag.PrintDefaults()
	os.Exit(2)
}
