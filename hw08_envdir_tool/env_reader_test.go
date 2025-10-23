package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testdataDir = "./testdata/env"
)

func TestReadDir(t *testing.T) {
	t.Run("not_a_directory", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "testFile")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { os.Remove(tempFile.Name()) })
		tempFile.Close()

		_, err = ReadDir(tempFile.Name())
		require.ErrorContains(t, err, "not a directory")
	})

	t.Run("filename_contains_equal", func(t *testing.T) {
		filePath := filepath.Join(testdataDir, "F=OO")
		file, err := os.Create(filePath)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { os.Remove(filePath) })
		file.Close()

		_, err = ReadDir(testdataDir)
		if !errors.Is(err, ErrUnsupportedFilename) {
			t.Errorf("expected error ErrUnsupportedFilename but got %v", err)
		}
	})

	t.Run("target_is_a_directory", func(t *testing.T) {
		testdataDirInfo, err := os.Lstat(testdataDir)
		if err != nil {
			t.Fatal(err)
		}

		filePath := filepath.Join(testdataDir, "test_directory")
		if err := os.Mkdir(filePath, testdataDirInfo.Mode().Perm()); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { os.Remove(filePath) })

		_, err = ReadDir(testdataDir)
		if !errors.Is(err, ErrNotAFile) {
			t.Errorf("expected error ErrNotAFile but got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		refMap := Environment{
			"BAR":   {Value: "bar", NeedRemove: false},
			"EMPTY": {Value: "", NeedRemove: false},
			"FOO":   {Value: "   foo\nwith new line", NeedRemove: false},
			"HELLO": {Value: "\"hello\"", NeedRemove: false},
			"UNSET": {Value: "", NeedRemove: true},
		}

		env, err := ReadDir(testdataDir)
		if err != nil {
			t.Fatal(err)
		}

		require.EqualValuesf(t, refMap, env, "environment variables mismatch")
	})

	t.Run("trailing_null_bytes", func(t *testing.T) {
		refMap := Environment{
			"BAR":   {Value: "bar", NeedRemove: false},
			"EMPTY": {Value: "", NeedRemove: false},
			"FOO":   {Value: "   foo\nwith new line", NeedRemove: false},
			"HELLO": {Value: "\"hello\"", NeedRemove: false},
			"UNSET": {Value: "", NeedRemove: true},
			"TEMP":  {Value: "temp\nnewline", NeedRemove: false},
		}

		filePath := filepath.Join(testdataDir, "TEMP")
		file, err := os.Create(filePath)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { os.Remove(filePath) })
		defer file.Close()

		nulls := []byte("temp\x00newline")
		if _, err = file.Write(nulls); err != nil {
			t.Fatal(err)
		}

		env, err := ReadDir(testdataDir)
		if err != nil {
			t.Fatal(err)
		}

		require.EqualValuesf(t, refMap, env, "environment variables mismatch")
	})
}
