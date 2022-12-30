package main

import (
	"fmt"
	"os"
	"testing"
)

var path = "test_files"

// Test the javascript file
func TestScanJS(t *testing.T) {
	filename := path + string(os.PathSeparator) + "javascript.js"
	test := File{path: filename, code: 16, lines: 27, comments: 9, blanks: 2}
	check_scan(t, filename, test)
}

// Test the PHP file
func TestScanPHP(t *testing.T) {
	filename := path + string(os.PathSeparator) + "php.php"
	test := File{path: filename, code: 19, lines: 27, comments: 5, blanks: 3}
	check_scan(t, filename, test)
}

// Check the scanner and compare against
// known values for the test
func check_scan(t *testing.T, filename string, test File) {
	stats, _ := os.Stat(filename)
	file := File{path: filename, info: stats}
	file.scan()

	if file.code != test.code {
		t.Error("Code wrong")
	}
	if file.lines != test.lines {
		t.Error("Lines wrong")
	}
	if file.comments != test.comments {
		t.Error("Comments wrong")
	}
	if file.blanks != test.blanks {
		t.Error("Blanks wrong")
	}

	if t.Failed() {
		printout(file, test)
	}
}

// Printout the scan results along with
// the known values for the test
func printout(file File, test File) {
	fmt.Printf("%-29s%10d%10d%10d%10d\n",
		file.info.Name(),
		file.blanks,
		file.comments,
		file.code,
		file.lines)
	fmt.Printf("%-29s%10d%10d%10d%10d\n",
		"Manual Count....",
		test.blanks,
		test.comments,
		test.code,
		test.lines)
}
