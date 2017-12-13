package main

import (
	"go/build"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/posener/complete"
)

// predictPackages completes packages in the directory pointed by a.Last
// and packages that are one level below that package.
func predictPackages(a complete.Args) (prediction []string) {
	prediction = []string{a.Last}
	lastPrediction := ""
	for len(prediction) == 1 && (lastPrediction == "" || lastPrediction != prediction[0]) {
		// if only one prediction, predict files within this prediction,
		// for example, if the user entered 'pk' and we have a package named 'pkg',
		// which is the only package prefixed with 'pk', we will automatically go one
		// level deeper and give the user the 'pkg' and all the nested packages within
		// that package.
		lastPrediction = prediction[0]
		a.Last = prediction[0]
		prediction = predictLocalAndSystem(a)
	}
	return
}

func predictLocalAndSystem(a complete.Args) []string {
	localDirs := complete.PredictFilesSet(listPackages(a.Directory())).Predict(a)
	// System directories are not actual file names, for example: 'github.com/posener/complete' could
	// be the argument, but the actual filename is in $GOPATH/src/github.com/posener/complete'. this
	// is the reason to use the PredictSet and not the PredictDirs in this case.
	s := systemDirs(a.Last)
	sysDirs := complete.PredictSet(s...).Predict(a)
	return append(localDirs, sysDirs...)
}

// listPackages looks in current pointed dir and in all it's direct sub-packages
// and return a list of paths to go packages.
func listPackages(dir string) (directories []string) {
	// add subdirectories
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		complete.Log("failed reading directory %s: %s", dir, err)
		return
	}

	// build paths array
	paths := make([]string, 0, len(files)+1)
	for _, f := range files {
		if f.IsDir() {
			paths = append(paths, filepath.Join(dir, f.Name()))
		}
	}
	paths = append(paths, dir)

	// import packages according to given paths
	for _, p := range paths {
		pkg, err := build.ImportDir(p, 0)
		if err != nil {
			complete.Log("failed importing directory %s: %s", p, err)
			continue
		}
		directories = append(directories, pkg.Dir)
	}
	return
}

func systemDirs(dir string) (directories []string) {
	// get all paths from GOPATH environment variable and use their src directory
	paths := findGopath()
	for i := range paths {
		paths[i] = filepath.Join(paths[i], "src")
	}

	// normalize the directory to be an actual directory since it could be with an additional
	// characters after the last '/'.
	if !strings.HasSuffix(dir, "/") {
		dir = filepath.Dir(dir)
	}

	for _, basePath := range paths {
		path := filepath.Join(basePath, dir)
		files, err := ioutil.ReadDir(path)
		if err != nil {
			// path does not exists
			continue
		}
		// add the base path as one of the completion options
		switch dir {
		case "", ".", "/", "./":
		default:
			directories = append(directories, dir)
		}
		// add all nested directories of the base path
		// go supports only packages and not go files within the GOPATH
		for _, f := range files {
			if !f.IsDir() {
				continue
			}
			directories = append(directories, filepath.Join(dir, f.Name())+"/")
		}
	}
	return
}

func findGopath() []string {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		// By convention
		// See rationale at https://github.com/golang/go/issues/17262
		usr, err := user.Current()
		if err != nil {
			return nil
		}
		usrgo := filepath.Join(usr.HomeDir, "go")
		return []string{usrgo}
	}
	listsep := string([]byte{os.PathListSeparator})
	entries := strings.Split(gopath, listsep)
	return entries
}
