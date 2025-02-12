package compute

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/text"
)

// TestGetViceroy validates that Viceroy is installed to the appropriate
// directory.
//
// There isn't an executable binary that exists in the test environment, so we
// expect the spawning of a subprocess (to call `<binary> --version`) to error
// and subsequently the `installViceroy()` function to be called.
//
// The `installViceroy()` function will then think it has downloaded the latest
// release as we have instructed the mock to provide that behaviour.
//
// Subsequently the `os.Rename()` will move the downloaded Viceroy binary,
// which is just a dummy file created by `makeEnvironment()`, into the intended
// destination directory.
func TestGetViceroy(t *testing.T) {
	binary := "foo"
	downloadDir, installDir, downloadedFile := makeEnvironment(binary, t)

	defer os.RemoveAll(downloadDir) // clean up

	InstallDir = installDir

	var out bytes.Buffer

	progress := text.NewQuietProgress(&out)
	versioner := mock.Versioner{
		Version:        "v1.2.3",
		BinaryFilename: binary,
		DownloadOK:     true,
		DownloadedFile: downloadedFile,
	}

	_, err := getViceroy(progress, &out, versioner)
	if err != nil {
		t.Fatal(err)
	}

	// NOTE: We have to call progress.Done() here to prevent a data race because
	// the getViceroy() function itself doesn't call it and so if we try to read
	// from the shared bytes.Buffer (as we do below to validate its content)
	// before calling Done(), then we'll get a race condition (this only shows up
	// when running the complete test suite or this specific test with a -count
	// value of 20 or above).
	progress.Done()

	if !strings.Contains(out.String(), "Fetching latest Viceroy release") {
		t.Fatalf("expected file to be downloaded successfully")
	}

	movedPath := filepath.Join(installDir, binary)

	if _, err := os.Stat(movedPath); err != nil {
		t.Fatalf("binary was not moved to the install directory: %s", err)
	}
}

// makeEnvironment creates a temporary directory for the test suite to utilise
// when validating Viceroy installation behaviours.
//
// It will create a file within the temporary directory to represent the call
// to `versioner.Download()` being successful.
//
// It also creates a nested directory within the temp directory to represent
// where the downloaded binary should be moved into.
//
// TODO: refactor testutil.NewEnv() to support directory creation.
func makeEnvironment(downloadedFilename string, t *testing.T) (string, string, string) {
	t.Helper()

	downloadDir, err := os.MkdirTemp("", "fastly-serve-*")
	if err != nil {
		t.Fatal(err)
	}

	fpath := filepath.Join(downloadDir, downloadedFilename)
	if err := os.WriteFile(fpath, []byte("..."), 0777); err != nil {
		t.Fatal(err)
	}

	installDir, err := os.MkdirTemp(downloadDir, "install")
	if err != nil {
		t.Fatal(err)
	}

	return downloadDir, installDir, fpath
}

// TODO: Write tests for the other functions in serve.go
