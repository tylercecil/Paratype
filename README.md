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
