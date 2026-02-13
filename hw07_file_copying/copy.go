package main

import (
	"bufio"
	"errors"
	"io"
	"os"
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
	ErrReadFileInfo          = errors.New("error read file info")
	ErrOpenFile              = errors.New("error open file")
	ErrSetFileOffset         = errors.New("file offset error")
	ErrReadFile              = errors.New("error read file")
	ErrWriteFile             = errors.New("error write file")
)

func Copy(fromPath, toPath string, offset, limit int64) error {
	// open from file
	fromFile, err := os.Open(fromPath)
	if err != nil {
		return ErrUnsupportedFile
	}
	defer fromFile.Close()

	// checl offset
	stat, err := fromFile.Stat()
	if err != nil {
		return ErrReadFileInfo
	}
	if stat.Size() < offset {
		return ErrOffsetExceedsFileSize
	}
	if limit == 0 {
		limit = stat.Size()
	}

	// open to file
	toFile, err := os.Create(toPath)
	if err != nil {
		return ErrOpenFile
	}
	defer toFile.Close()

	// set the offset
	_, err = fromFile.Seek(offset, 0)
	if err != nil {
		return ErrSetFileOffset
	}

	reader := bufio.NewReaderSize(fromFile, 10)
	writer := bufio.NewWriter(toFile)

	// read and write byte by byte
	for limit > 0 {
		b, err := reader.ReadByte()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return ErrReadFile
		}

		err = writer.WriteByte(b)
		if err != nil {
			return ErrWriteFile
		}

		limit--
	}

	writer.Flush()
	return nil
}
