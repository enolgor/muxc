package main

import (
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

var file string
var watch bool

func init() {
	flag.StringVar(&file, "f", "muxc.yaml", "path to yaml configuration file")
	flag.BoolVar(&watch, "w", false, "watch and rebuild changes to configuration file")
	flag.Parse()
}

func main() {
	var run func() error
	if watch {
		run = watchAndRebuild
	} else {
		run = processFile
	}
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(-1)
	}
}

func processFile() error {
	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("unable to locate %s file: %s", file, err.Error())
	}
	defer f.Close()
	dir := filepath.Dir(file)
	yamlfile, err := NewMultiYamlFile(file, f, dir)
	if err != nil {
		return fmt.Errorf("error merging yaml files: %s", err.Error())
	}
	if err = Generate(yamlfile); err != nil {
		return fmt.Errorf("error generating muxc routes: %s", err.Error())
	}
	return nil
}

func isSameErr(err1, err2 error) bool {
	if err1 == nil && err2 == nil {
		return true
	}
	if err1 == nil || err2 == nil {
		return false
	}
	return err1.Error() == err2.Error()
}

func printErrIfNotSame(lastErr, err error, w io.Writer, format string, args ...any) error {
	if !isSameErr(lastErr, err) {
		fmt.Fprintf(w, format, args...)
	}
	return err
}

func watchAndRebuild() error {
	fmt.Println("watching for file changes...")
	var lastChecksum, checksum string
	var lastErr error
	for {
		f, err := os.Open(file)
		if err != nil {
			lastErr = printErrIfNotSame(lastErr, err, os.Stderr, "unable to locate %s file: %s\n", file, err.Error())
			continue
		}
		yamlfile, err := NewMultiYamlFile(file, f, filepath.Dir(file))
		if err != nil {
			lastErr = printErrIfNotSame(lastErr, err, os.Stderr, "error merging yaml files: %s\n", err.Error())
			continue
		}
		if checksum, err = getFileChecksums(yamlfile); err != nil {
			lastErr = printErrIfNotSame(lastErr, err, os.Stderr, "error calculating file checksums: %s\n", err.Error())
			continue
		}
		if checksum != lastChecksum {
			lastChecksum = checksum
			now := time.Now()
			if err = Generate(yamlfile); err != nil {
				lastErr = printErrIfNotSame(lastErr, err, os.Stderr, "error generating muxc routes: %s\n", err.Error())
				continue
			} else {
				fmt.Printf("Built changes in %s\n", time.Since(now))
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func getFileChecksums(yamlfile *MultiYamlFile) (string, error) {
	multihash := ""
	paths := yamlfile.GetAllFilePaths()
	for i := range paths {
		hash, err := getFileChecksum(paths[i])
		if err != nil {
			return "", err
		}
		multihash += hash
	}
	return multihash, nil
}

func getFileChecksum(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", fmt.Errorf("unable to locate %s file: %s", file, err.Error())
	}
	defer f.Close()
	hash := md5.New()
	if _, err := io.Copy(hash, f); err != nil {
		return "", fmt.Errorf("error calculating file hash: %s", err.Error())
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
