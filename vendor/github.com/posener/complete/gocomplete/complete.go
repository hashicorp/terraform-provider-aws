// Package main is complete tool for the go command line
package main

import "github.com/posener/complete"

var (
	ellipsis   = complete.PredictSet("./...")
	anyPackage = complete.PredictFunc(predictPackages)
	goFiles    = complete.PredictFiles("*.go")
	anyFile    = complete.PredictFiles("*")
	anyGo      = complete.PredictOr(goFiles, anyPackage, ellipsis)
)

func main() {
	build := complete.Command{
		Flags: complete.Flags{
			"-o": anyFile,
			"-i": complete.PredictNothing,

			"-a":             complete.PredictNothing,
			"-n":             complete.PredictNothing,
			"-p":             complete.PredictAnything,
			"-race":          complete.PredictNothing,
			"-msan":          complete.PredictNothing,
			"-v":             complete.PredictNothing,
			"-work":          complete.PredictNothing,
			"-x":             complete.PredictNothing,
			"-asmflags":      complete.PredictAnything,
			"-buildmode":     complete.PredictAnything,
			"-compiler":      complete.PredictAnything,
			"-gccgoflags":    complete.PredictSet("gccgo", "gc"),
			"-gcflags":       complete.PredictAnything,
			"-installsuffix": complete.PredictAnything,
			"-ldflags":       complete.PredictAnything,
			"-linkshared":    complete.PredictNothing,
			"-pkgdir":        anyPackage,
			"-tags":          complete.PredictAnything,
			"-toolexec":      complete.PredictAnything,
		},
		Args: anyGo,
	}

	run := complete.Command{
		Flags: complete.Flags{
			"-exec": complete.PredictAnything,
		},
		Args: goFiles,
	}

	test := complete.Command{
		Flags: complete.Flags{
			"-args": complete.PredictAnything,
			"-c":    complete.PredictNothing,
			"-exec": complete.PredictAnything,

			"-bench":     predictBenchmark,
			"-benchtime": complete.PredictAnything,
			"-count":     complete.PredictAnything,
			"-cover":     complete.PredictNothing,
			"-covermode": complete.PredictSet("set", "count", "atomic"),
			"-coverpkg":  complete.PredictDirs("*"),
			"-cpu":       complete.PredictAnything,
			"-run":       predictTest,
			"-short":     complete.PredictNothing,
			"-timeout":   complete.PredictAnything,

			"-benchmem":             complete.PredictNothing,
			"-blockprofile":         complete.PredictFiles("*.out"),
			"-blockprofilerate":     complete.PredictAnything,
			"-coverprofile":         complete.PredictFiles("*.out"),
			"-cpuprofile":           complete.PredictFiles("*.out"),
			"-memprofile":           complete.PredictFiles("*.out"),
			"-memprofilerate":       complete.PredictAnything,
			"-mutexprofile":         complete.PredictFiles("*.out"),
			"-mutexprofilefraction": complete.PredictAnything,
			"-outputdir":            complete.PredictDirs("*"),
			"-trace":                complete.PredictFiles("*.out"),
		},
		Args: anyGo,
	}

	fmt := complete.Command{
		Flags: complete.Flags{
			"-n": complete.PredictNothing,
			"-x": complete.PredictNothing,
		},
		Args: anyGo,
	}

	get := complete.Command{
		Flags: complete.Flags{
			"-d":        complete.PredictNothing,
			"-f":        complete.PredictNothing,
			"-fix":      complete.PredictNothing,
			"-insecure": complete.PredictNothing,
			"-t":        complete.PredictNothing,
			"-u":        complete.PredictNothing,
		},
		Args: anyGo,
	}

	generate := complete.Command{
		Flags: complete.Flags{
			"-n":   complete.PredictNothing,
			"-x":   complete.PredictNothing,
			"-v":   complete.PredictNothing,
			"-run": complete.PredictAnything,
		},
		Args: anyGo,
	}

	vet := complete.Command{
		Flags: complete.Flags{
			"-n": complete.PredictNothing,
			"-x": complete.PredictNothing,
		},
		Args: anyGo,
	}

	list := complete.Command{
		Flags: complete.Flags{
			"-e":    complete.PredictNothing,
			"-f":    complete.PredictAnything,
			"-json": complete.PredictNothing,
		},
		Args: complete.PredictOr(anyPackage, ellipsis),
	}

	doc := complete.Command{
		Flags: complete.Flags{
			"-c":   complete.PredictNothing,
			"-cmd": complete.PredictNothing,
			"-u":   complete.PredictNothing,
		},
		Args: anyPackage,
	}

	tool := complete.Command{
		Flags: complete.Flags{
			"-n": complete.PredictNothing,
		},
		Sub: complete.Commands{
			"addr2line": {
				Args: anyFile,
			},
			"asm": {
				Flags: complete.Flags{
					"-D":        complete.PredictAnything,
					"-I":        complete.PredictDirs("*"),
					"-S":        complete.PredictNothing,
					"-debug":    complete.PredictNothing,
					"-dynlink":  complete.PredictNothing,
					"-e":        complete.PredictNothing,
					"-o":        anyFile,
					"-shared":   complete.PredictNothing,
					"-trimpath": complete.PredictNothing,
				},
				Args: complete.PredictFiles("*.s"),
			},
			"cgo": {
				Flags: complete.Flags{
					"-debug-define":      complete.PredictNothing,
					"debug-gcc":          complete.PredictNothing,
					"dynimport":          anyFile,
					"dynlinker":          complete.PredictNothing,
					"dynout":             anyFile,
					"dynpackage":         anyPackage,
					"exportheader":       complete.PredictDirs("*"),
					"gccgo":              complete.PredictNothing,
					"gccgopkgpath":       complete.PredictDirs("*"),
					"gccgoprefix":        complete.PredictAnything,
					"godefs":             complete.PredictNothing,
					"import_runtime_cgo": complete.PredictNothing,
					"import_syscall":     complete.PredictNothing,
					"importpath":         complete.PredictDirs("*"),
					"objdir":             complete.PredictDirs("*"),
					"srcdir":             complete.PredictDirs("*"),
				},
				Args: goFiles,
			},
			"compile": {
				Flags: complete.Flags{
					"-%":              complete.PredictNothing,
					"-+":              complete.PredictNothing,
					"-B":              complete.PredictNothing,
					"-D":              complete.PredictDirs("*"),
					"-E":              complete.PredictNothing,
					"-I":              complete.PredictDirs("*"),
					"-K":              complete.PredictNothing,
					"-N":              complete.PredictNothing,
					"-S":              complete.PredictNothing,
					"-V":              complete.PredictNothing,
					"-W":              complete.PredictNothing,
					"-asmhdr":         anyFile,
					"-bench":          anyFile,
					"-buildid":        complete.PredictNothing,
					"-complete":       complete.PredictNothing,
					"-cpuprofile":     anyFile,
					"-d":              complete.PredictNothing,
					"-dynlink":        complete.PredictNothing,
					"-e":              complete.PredictNothing,
					"-f":              complete.PredictNothing,
					"-h":              complete.PredictNothing,
					"-i":              complete.PredictNothing,
					"-importmap":      complete.PredictAnything,
					"-installsuffix":  complete.PredictAnything,
					"-j":              complete.PredictNothing,
					"-l":              complete.PredictNothing,
					"-largemodel":     complete.PredictNothing,
					"-linkobj":        anyFile,
					"-live":           complete.PredictNothing,
					"-m":              complete.PredictNothing,
					"-memprofile":     complete.PredictNothing,
					"-memprofilerate": complete.PredictAnything,
					"-msan":           complete.PredictNothing,
					"-nolocalimports": complete.PredictNothing,
					"-o":              anyFile,
					"-p":              complete.PredictDirs("*"),
					"-pack":           complete.PredictNothing,
					"-r":              complete.PredictNothing,
					"-race":           complete.PredictNothing,
					"-s":              complete.PredictNothing,
					"-shared":         complete.PredictNothing,
					"-traceprofile":   anyFile,
					"-trimpath":       complete.PredictAnything,
					"-u":              complete.PredictNothing,
					"-v":              complete.PredictNothing,
					"-w":              complete.PredictNothing,
					"-wb":             complete.PredictNothing,
				},
				Args: goFiles,
			},
			"cover": {
				Flags: complete.Flags{
					"-func": complete.PredictAnything,
					"-html": complete.PredictAnything,
					"-mode": complete.PredictSet("set", "count", "atomic"),
					"-o":    anyFile,
					"-var":  complete.PredictAnything,
				},
				Args: anyFile,
			},
			"dist": {
				Sub: complete.Commands{
					"banner":    {Flags: complete.Flags{"-v": complete.PredictNothing}},
					"bootstrap": {Flags: complete.Flags{"-v": complete.PredictNothing}},
					"clean":     {Flags: complete.Flags{"-v": complete.PredictNothing}},
					"env":       {Flags: complete.Flags{"-v": complete.PredictNothing, "-p": complete.PredictNothing}},
					"install":   {Flags: complete.Flags{"-v": complete.PredictNothing}, Args: complete.PredictDirs("*")},
					"list":      {Flags: complete.Flags{"-v": complete.PredictNothing, "-json": complete.PredictNothing}},
					"test":      {Flags: complete.Flags{"-v": complete.PredictNothing, "-h": complete.PredictNothing}},
					"version":   {Flags: complete.Flags{"-v": complete.PredictNothing}},
				},
			},
			"doc": doc,
			"fix": {
				Flags: complete.Flags{
					"-diff":  complete.PredictNothing,
					"-force": complete.PredictAnything,
					"-r":     complete.PredictSet("context", "gotypes", "netipv6zone", "printerconfig"),
				},
				Args: anyGo,
			},
			"link": {},
			"nm": {
				Flags: complete.Flags{
					"-n":    complete.PredictNothing,
					"-size": complete.PredictNothing,
					"-sort": complete.PredictAnything,
					"-type": complete.PredictNothing,
				},
				Args: anyGo,
			},
			"objdump": {
				Flags: complete.Flags{
					"-s": complete.PredictAnything,
				},
				Args: anyFile,
			},
			"pack": {},
			"pprof": {
				Flags: complete.Flags{
					"-callgrind":     complete.PredictNothing,
					"-disasm":        complete.PredictAnything,
					"-dot":           complete.PredictNothing,
					"-eog":           complete.PredictNothing,
					"-evince":        complete.PredictNothing,
					"-gif":           complete.PredictNothing,
					"-gv":            complete.PredictNothing,
					"-list":          complete.PredictAnything,
					"-pdf":           complete.PredictNothing,
					"-peek":          complete.PredictAnything,
					"-png":           complete.PredictNothing,
					"-proto":         complete.PredictNothing,
					"-ps":            complete.PredictNothing,
					"-raw":           complete.PredictNothing,
					"-svg":           complete.PredictNothing,
					"-tags":          complete.PredictNothing,
					"-text":          complete.PredictNothing,
					"-top":           complete.PredictNothing,
					"-tree":          complete.PredictNothing,
					"-web":           complete.PredictNothing,
					"-weblist":       complete.PredictAnything,
					"-output":        anyFile,
					"-functions":     complete.PredictNothing,
					"-files":         complete.PredictNothing,
					"-lines":         complete.PredictNothing,
					"-addresses":     complete.PredictNothing,
					"-base":          complete.PredictAnything,
					"-drop_negative": complete.PredictNothing,
					"-cum":           complete.PredictNothing,
					"-seconds":       complete.PredictAnything,
					"-nodecount":     complete.PredictAnything,
					"-nodefraction":  complete.PredictAnything,
					"-edgefraction":  complete.PredictAnything,
					"-sample_index":  complete.PredictNothing,
					"-mean":          complete.PredictNothing,
					"-inuse_space":   complete.PredictNothing,
					"-inuse_objects": complete.PredictNothing,
					"-alloc_space":   complete.PredictNothing,
					"-alloc_objects": complete.PredictNothing,
					"-total_delay":   complete.PredictNothing,
					"-contentions":   complete.PredictNothing,
					"-mean_delay":    complete.PredictNothing,
					"-runtime":       complete.PredictNothing,
					"-focus":         complete.PredictAnything,
					"-ignore":        complete.PredictAnything,
					"-tagfocus":      complete.PredictAnything,
					"-tagignore":     complete.PredictAnything,
					"-call_tree":     complete.PredictNothing,
					"-unit":          complete.PredictAnything,
					"-divide_by":     complete.PredictAnything,
					"-buildid":       complete.PredictAnything,
					"-tools":         complete.PredictDirs("*"),
					"-help":          complete.PredictNothing,
				},
				Args: anyFile,
			},
			"tour": {
				Flags: complete.Flags{
					"-http":        complete.PredictAnything,
					"-openbrowser": complete.PredictNothing,
				},
			},
			"trace": {
				Flags: complete.Flags{
					"-http":  complete.PredictAnything,
					"-pprof": complete.PredictSet("net", "sync", "syscall", "sched"),
				},
				Args: anyFile,
			},
			"vet": {
				Flags: complete.Flags{
					"-all":                 complete.PredictNothing,
					"-asmdecl":             complete.PredictNothing,
					"-assign":              complete.PredictNothing,
					"-atomic":              complete.PredictNothing,
					"-bool":                complete.PredictNothing,
					"-buildtags":           complete.PredictNothing,
					"-cgocall":             complete.PredictNothing,
					"-composites":          complete.PredictNothing,
					"-compositewhitelist":  complete.PredictNothing,
					"-copylocks":           complete.PredictNothing,
					"-httpresponse":        complete.PredictNothing,
					"-lostcancel":          complete.PredictNothing,
					"-methods":             complete.PredictNothing,
					"-nilfunc":             complete.PredictNothing,
					"-printf":              complete.PredictNothing,
					"-printfuncs":          complete.PredictAnything,
					"-rangeloops":          complete.PredictNothing,
					"-shadow":              complete.PredictNothing,
					"-shadowstrict":        complete.PredictNothing,
					"-shift":               complete.PredictNothing,
					"-structtags":          complete.PredictNothing,
					"-tags":                complete.PredictAnything,
					"-tests":               complete.PredictNothing,
					"-unreachable":         complete.PredictNothing,
					"-unsafeptr":           complete.PredictNothing,
					"-unusedfuncs":         complete.PredictAnything,
					"-unusedresult":        complete.PredictNothing,
					"-unusedstringmethods": complete.PredictAnything,
					"-v": complete.PredictNothing,
				},
				Args: anyGo,
			},
		},
	}

	clean := complete.Command{
		Flags: complete.Flags{
			"-i": complete.PredictNothing,
			"-r": complete.PredictNothing,
			"-n": complete.PredictNothing,
			"-x": complete.PredictNothing,
		},
		Args: complete.PredictOr(anyPackage, ellipsis),
	}

	env := complete.Command{
		Args: complete.PredictAnything,
	}

	bug := complete.Command{}
	version := complete.Command{}

	fix := complete.Command{
		Args: anyGo,
	}

	// commands that also accepts the build flags
	for name, options := range build.Flags {
		test.Flags[name] = options
		run.Flags[name] = options
		list.Flags[name] = options
		vet.Flags[name] = options
		get.Flags[name] = options
	}

	gogo := complete.Command{
		Sub: complete.Commands{
			"build":    build,
			"install":  build, // install and build have the same flags
			"run":      run,
			"test":     test,
			"fmt":      fmt,
			"get":      get,
			"generate": generate,
			"vet":      vet,
			"list":     list,
			"doc":      doc,
			"tool":     tool,
			"clean":    clean,
			"env":      env,
			"bug":      bug,
			"fix":      fix,
			"version":  version,
		},
		GlobalFlags: complete.Flags{
			"-h": complete.PredictNothing,
		},
	}

	complete.New("go", gogo).Run()
}
