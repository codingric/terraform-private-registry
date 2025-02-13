package main

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

func compressFile(src, dest string) error {
	archive, _ := os.Create("1" + dest)

	bn := filepath.Base(src)

	// Create a new zip archive.
	w := zip.NewWriter(archive)
	zf, err := w.Create(bn)
	if err != nil {
		return err
	}

	f, err := os.Open(src)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	_, err = io.Copy(zf, f)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	r, _ := zip.OpenReader("1" + dest)
	defer r.Close()

	ff, _ := os.Create(dest)
	defer f.Close()

	w2 := zip.NewWriter(ff)
	defer w2.Close()

	for _, f := range r.File {
		f.SetMode(0777)
		w2.Copy(f)
	}
	os.Remove("1" + dest)
	return nil
}
