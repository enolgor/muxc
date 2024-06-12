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

	"github.com/enolgor/muxc/muxc"
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
		fmt.Fprintf(os.Stderr, err.Error()+"\n")
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
	if err = muxc.Generate(file, f, dir); err != nil {
		return fmt.Errorf("error generating muxc routes: %s", err.Error())
	}
	return nil
}

func watchAndRebuild() error {
	fmt.Println("watching for file changes...")
	var lastChecksum, checksum string
	var err error
	for {
		if checksum, err = getFileChecksum(); err != nil {
			return err
		}
		if checksum != lastChecksum {
			lastChecksum = checksum
			now := time.Now()
			if err = processFile(); err != nil {
				fmt.Fprintf(os.Stderr, err.Error()+"\n")
			} else {
				fmt.Printf("Built changes in %s\n", time.Since(now))
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func getFileChecksum() (string, error) {
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
