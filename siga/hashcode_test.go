package siga

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Set to true to save outputs in testdata/ for validation with external tools.
const saveFromHashcode = false

func runFromHashcodeTest(t *testing.T, name string, expected error) {
	t.Helper()

	// given
	dst := ioutil.Discard
	if saveFromHashcode {
		fd, err := os.Create(fmt.Sprintf("testdata/%s_hashcode_output.asice", name))
		if err != nil {
			t.Fatal(err)
		}
		defer fd.Close()
		dst = fd
	}

	fd, err := os.Open(fmt.Sprintf("testdata/%s_hashcode.asice", name))
	if err != nil {
		t.Fatal(err)
	}
	defer fd.Close()
	src, size, err := toReaderAt(fd)
	if err != nil {
		t.Fatal(err)
	}

	glob, err := filepath.Glob(fmt.Sprintf("testdata/%s_datafile*", name))
	if err != nil {
		t.Fatal(err)
	}
	datafiles := make([]*DataFile, 0, len(glob))
	for _, path := range glob {
		df, err := ReadDataFile(path)
		if err != nil {
			t.Fatal(err)
		}
		datafiles = append(datafiles, df)
	}

	// when
	err = fromHashcode(dst, src, size, datafiles...)

	// then
	if expected == nil {
		if err != nil {
			t.Fatal("unexpected error:", err)
		}
		return
	}
	if err == nil {
		t.Fatal("unexpected success, expected:", expected)
	}
	if !strings.Contains(err.Error(), expected.Error()) {
		t.Fatalf("unexpected error:\n     got: %v\nexpected: %v", err, expected)
	}
}

func TestFromHashcode_EmptyContainer_Succeeds(t *testing.T) {
	runFromHashcodeTest(t, "empty", nil)
}

func TestFromHashcode_SingleDatafile_Succeeds(t *testing.T) {
	runFromHashcodeTest(t, "single", nil)
}

func TestFromHashcode_MultipleDatafiles_Succeeds(t *testing.T) {
	runFromHashcodeTest(t, "multiple", nil)
}

func TestFromHashcode_MissingDatafiles_Errors(t *testing.T) {
	runFromHashcodeTest(t, "missing", errors.New("missing missing_datafile.txt"))
}

func TestFromHashcode_UnknownDatafiles_Errors(t *testing.T) {
	runFromHashcodeTest(t, "unknown", errors.New("unknown unknown_datafile.txt"))
}

func TestFromHashcode_MismatchingDatafiles_Errors(t *testing.T) {
	runFromHashcodeTest(t, "mismatching", errors.New("mismatching mismatching_datafile.txt hash"))
}
