package main

import (
	"errors"
	"io"
	"os"

	"github.com/cheggaaa/pb/v3" //nolint:depguard
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

func Copy(fromPath, toPath string, offset, limit int64) error {
	sourceFileInfo, err := os.Lstat(fromPath)
	if err != nil {
		return err
	}

	if !sourceFileInfo.Mode().IsRegular() {
		return ErrUnsupportedFile
	}

	sourceFileSize := sourceFileInfo.Size()
	if offset > sourceFileSize {
		return ErrOffsetExceedsFileSize
	}

	sourceFile, err := os.OpenFile(fromPath, os.O_RDONLY, sourceFileInfo.Mode().Perm())
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	var bytesToCopy int64
	switch {
	case limit == 0:
		bytesToCopy = sourceFileSize - offset
	case limit+offset > sourceFileSize:
		bytesToCopy = sourceFileSize - offset
	case limit <= sourceFileSize-offset:
		bytesToCopy = limit
	}

	destFile, err := os.Create(toPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	err = os.Chmod(toPath, sourceFileInfo.Mode().Perm())
	if err != nil {
		return err
	}

	_, err = sourceFile.Seek(offset, io.SeekStart)
	if err != nil {
		return err
	}

	bar := pb.Full.Start64(bytesToCopy)
	proxyReader := bar.NewProxyReader(sourceFile)

	_, err = io.CopyN(destFile, proxyReader, bytesToCopy)
	if err != nil {
		if !errors.Is(err, io.EOF) {
			return err
		}
	}
	bar.Finish()

	return nil
}
