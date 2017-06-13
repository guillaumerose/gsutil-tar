package main

import (
	"archive/tar"
	"context"
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
)

func main() {
	var id, bucket, mode, directory string
	flag.StringVar(&mode, "mode", "", "mode: push or pull")
	flag.StringVar(&id, "id", "", "id")
	flag.StringVar(&bucket, "bucket", "pipeline-contexts", "bucket")
	flag.StringVar(&directory, "directory", "", "directory")
	flag.Parse()

	object := id + ".tar"

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	if mode == "push" {
		writer := client.Bucket(bucket).Object(object).NewWriter(ctx)
		err = tarit(directory, writer)
		if err != nil {
			log.Fatal(err)
		}
		err = writer.Close()
		if err != nil {
			log.Fatal(err)
		}
	} else if mode == "pull" {
		reader, err := client.Bucket(bucket).Object(object).NewReader(ctx)
		if err != nil {
			log.Fatal(err)
		}
		defer reader.Close()
		err = untar(reader, directory)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		panic("Unrecognized mode")
	}
}

func tarit(source string, target io.Writer) error {
	tarball := tar.NewWriter(target)
	defer tarball.Close()

	info, err := os.Stat(source)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	return filepath.Walk(source,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			header, err := tar.FileInfoHeader(info, info.Name())
			if err != nil {
				return err
			}

			if baseDir != "" {
				header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
			}

			if err := tarball.WriteHeader(header); err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(tarball, file)
			return err
		})
}

func untar(tarball io.Reader, target string) error {
	tarReader := tar.NewReader(tarball)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		path := filepath.Join(target, header.Name)
		info := header.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(file, tarReader)
		if err != nil {
			return err
		}
	}
	return nil
}
