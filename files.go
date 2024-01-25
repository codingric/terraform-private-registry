package main

import (
	"archive/zip"
	"bytes"
)

func compressFile(contents []byte, name string) ([]byte, error) {
	buf := new(bytes.Buffer)

	// Create a new zip archive.
	w := zip.NewWriter(buf)
	f, err := w.Create(name)
	if err != nil {
		return nil, err
	}
	_, err = f.Write(contents)
	if err != nil {
		return nil, err
	}
	err = w.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
