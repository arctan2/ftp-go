package common

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func DirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

func ZipSource(source []string, target string, gh *GobHandler) error {
	var (
		zf  *os.File
		err error
	)

	if PathExists(target) {
		zf, err = os.Open(target)
	} else {
		zf, err = os.Create(target)
	}

	if err != nil {
		return err
	}
	defer zf.Close()

	writer := zip.NewWriter(zf)
	defer writer.Close()

	var totalSize, doneSize int64
	if gh != nil {
		for _, path := range source {
			s, _ := DirSize(path)
			totalSize += s
		}
	}

	for _, sourcePath := range source {
		if err = filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Println(err)
				return err
			}

			if info.IsDir() {
				return nil
			}

			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}

			header.Method = zip.Deflate

			// path = filepath.ToSlash(path)

			header.Name, err = filepath.Rel(filepath.Dir(sourcePath), path)
			if err != nil {
				return err
			}

			headerWriter, err := writer.CreateHeader(header)
			if err != nil {
				return err
			}

			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()

			n, err := io.Copy(headerWriter, f)
			if gh != nil {
				doneSize += n
				go gh.Encode(ZipProgress{Max: totalSize, Current: doneSize, IsDone: totalSize == doneSize})
			}
			return err
		}); err != nil {
			return err
		}
	}
	return err
}

func UnzipSource(source, destination string) error {
	reader, err := zip.OpenReader(source)
	if err != nil {
		return err
	}
	defer reader.Close()

	destination, err = filepath.Abs(destination)
	if err != nil {
		return err
	}

	for _, f := range reader.File {
		err := unzipFile(f, destination)
		if err != nil {
			return err
		}
	}

	return nil
}

func unzipFile(f *zip.File, destination string) error {
	filePath := filepath.Join(destination, f.Name)
	if !strings.HasPrefix(filePath, filepath.Clean(destination)+string(os.PathSeparator)) {
		return fmt.Errorf("invalid file path: %s", filePath)
	}

	filePath = strings.ReplaceAll(filePath, "\\", "/")
	if f.FileInfo().IsDir() {
		if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
			return err
		}
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return err
	}

	destinationFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	zippedFile, err := f.Open()
	if err != nil {
		return err
	}
	defer zippedFile.Close()

	if _, err := io.Copy(destinationFile, zippedFile); err != nil {
		return err
	}
	return nil
}
