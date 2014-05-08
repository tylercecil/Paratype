Paratype
========

A parallel type analysis/inference system. New Mexico Tech CSE451 Project.

Testing
-------

### Merge Unit Tests
To be found in merge_test.go, run with

	go test -v -run Name merge_test.go

* Type by parent (works, see unit test named Down)
* Type by child (works, see unit test named Up[0-3] where 0-2 with type errors)
* Two parents with type (see unit test named Two)
* Our favorite example (see unit test named Flow)

To run our favorite example:

	go test -v -run Flow merge_test.go

Compilation
-----------

To set up Go, please visit http://golang.org/doc/install.

To compile Paratype:

	go get https://github.com/skelterjohn/gopp
	go build paratype.go

For your convenience (so there is no need to install Google Go and gopp), we
have included a binary called "paratype" that was compiled from the latest
version on a 64-bit machine and should run on the CS department
login.cs.nmt.edu machine.

To list command line options for Paratype:

	./paratype -h

The command line options are:

	-infile=
        default: ""
        This command line option is necessary and an error will be thrown if it 
		is not present. 

	-outfile=
		default: "" 
		This is necessary if the print flag is used. This is the file to print 
		the generated implementations to.

	-print=
        default: false
        This determines whether to print the implementations to the given file
		or not.

	-procs=
        default: 4
        This determines the number for GOMAXPROCS

	-time=
		default: false
        This determines whether the time in nanoseconds should be gathered. 

You may find files to test in the testfiles directory. Please see known bugs in
the report.
