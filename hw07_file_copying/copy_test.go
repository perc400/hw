package main

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	tempFile   = "/tmp/test"
	sourceFile = "testdata/input.txt"
)

func TestFullCopy(t *testing.T) {
	t.Run("copyFullFile", func(t *testing.T) {
		defer os.Remove(tempFile)

		err := Copy(sourceFile, tempFile, 0, 0)
		if err != nil {
			t.Fatal(err)
		}

		sourceFileInfo, err := os.Lstat(sourceFile)
		if err != nil {
			return
		}
		tempFileInfo, err := os.Lstat(tempFile)
		if err != nil {
			return
		}

		require.Equal(t, sourceFileInfo.Size(), tempFileInfo.Size())
		require.Equal(t, sourceFileInfo.Mode().Perm(), tempFileInfo.Mode().Perm())
	})
}

func TestOffset0Limit10(t *testing.T) {
	var limit int64 = 10
	var offset int64

	t.Run("out_offset0_limit10", func(t *testing.T) {
		defer os.Remove(tempFile)

		err := Copy(sourceFile, tempFile, offset, limit)
		if err != nil {
			t.Fatal(err)
		}

		tempFileInfo, err := os.Lstat(tempFile)
		if err != nil {
			return
		}

		require.Equal(t, tempFileInfo.Size(), limit)
	})
}

func TestOffsetEqualFileSize(t *testing.T) {
	var limit int64
	var offset int64

	t.Run("out_offset_equal_filesize", func(t *testing.T) {
		defer os.Remove(tempFile)

		sourceFileInfo, err := os.Lstat(sourceFile)
		if err != nil {
			return
		}
		offset = sourceFileInfo.Size()

		err = Copy(sourceFile, tempFile, offset, limit)
		if err != nil {
			t.Fatal(err)
		}

		tempFileInfo, err := os.Lstat(tempFile)
		if err != nil {
			return
		}

		require.Equal(t, tempFileInfo.Size(), int64(0))
	})
}

func TestOffsetBiggerThanFileSize(t *testing.T) {
	var limit int64 = 100
	var offset int64

	t.Run("offset_bigger_than_filesize", func(t *testing.T) {
		defer os.Remove(tempFile)

		sourceFileInfo, err := os.Lstat(sourceFile)
		if err != nil {
			return
		}
		offset = sourceFileInfo.Size() + 1000

		err = Copy(sourceFile, tempFile, offset, limit)
		if !errors.Is(err, ErrOffsetExceedsFileSize) {
			t.Errorf("expected error ErrOffsetExceedsFileSize but got %v", err)
		}
	})
}

func TestCopyNonRegularFile(t *testing.T) {
	deviceFile := "/dev/urandom"
	var limit int64
	var offset int64

	t.Run("skip_device", func(t *testing.T) {
		defer os.Remove(tempFile)

		err := Copy(deviceFile, tempFile, offset, limit)
		if !errors.Is(err, ErrUnsupportedFile) {
			t.Errorf("expected error ErrUnsupportedFile but got %v", err)
		}
	})
}
