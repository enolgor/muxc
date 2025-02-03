package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"slices"

	"gopkg.in/yaml.v3"
)

type MultiYamlFile struct {
	BaseDir    string
	SourceFile string
	FilePath   string
	data       []byte
	decoded    map[string]any
	includes   []*MultiYamlFile
}

func NewMultiYamlFile(sourceFile string, cfgFile io.Reader, basedir string) (*MultiYamlFile, error) {
	file, err := resolveIncludes(path.Base(sourceFile), cfgFile, basedir, nil)
	if err != nil {
		return nil, err
	}
	file.SourceFile = sourceFile
	file.BaseDir = basedir
	return file, err
}

func (yf *MultiYamlFile) GetAllFilePaths() []string {
	paths := []string{yf.FilePath}
	for i := range yf.includes {
		paths = append(paths, yf.includes[i].GetAllFilePaths()...)
	}
	return paths
}

func (yf *MultiYamlFile) Resolve() (io.Reader, error) {
	if err := yf.decode(); err != nil {
		return nil, err
	}
	data := yf.merge()
	buffer := &bytes.Buffer{}
	enc := yaml.NewEncoder(buffer)
	if err := enc.Encode(data); err != nil {
		return nil, err
	}
	return buffer, nil
}

func (yf *MultiYamlFile) merge() map[string]any {
	for i := range yf.includes {
		yf.decoded = mergeMaps(yf.decoded, yf.includes[i].merge())
	}
	return yf.decoded
}

func (yf *MultiYamlFile) decode() error {
	yf.decoded = map[string]any{}
	if err := yaml.Unmarshal(yf.data, &yf.decoded); err != nil {
		return err
	}
	for i := range yf.includes {
		if err := yf.includes[i].decode(); err != nil {
			return err
		}
	}
	return nil
}

var includeMatcher *regexp.Regexp = regexp.MustCompile(`^!include "*([^"]+)"*$`)

func resolveIncludes(filename string, file io.Reader, basedir string, alreadyIncluded []string) (*MultiYamlFile, error) {
	if alreadyIncluded == nil {
		alreadyIncluded = []string{}
	}
	buffer := &bytes.Buffer{}
	filePath := path.Join(basedir, filename)
	if slices.Contains(alreadyIncluded, filePath) {
		return nil, fmt.Errorf("include loop detected, '%s' is already included", filePath)
	}
	alreadyIncluded = append(alreadyIncluded, filePath)
	scanner := bufio.NewScanner(file)
	includes := []string{}
	for scanner.Scan() {
		line := scanner.Text()
		matches := includeMatcher.FindStringSubmatch(line)
		if len(matches) == 2 {
			includes = append(includes, matches[1])
		} else {
			buffer.WriteString(line + "\n")
		}
	}
	yamlFile := &MultiYamlFile{
		FilePath: filePath,
		data:     buffer.Bytes(),
		includes: []*MultiYamlFile{},
	}
	for i := range includes {
		f, err := os.Open(path.Join(basedir, includes[i]))
		if err != nil {
			return nil, err
		}
		defer f.Close()
		included, err := resolveIncludes(includes[i], f, basedir, alreadyIncluded)
		if err != nil {
			return nil, err
		}
		yamlFile.includes = append(yamlFile.includes, included)
	}
	return yamlFile, nil
}

func mergeMaps(data1, data2 map[string]any) map[string]any {
	for key, value2 := range data2 {
		if value1, exists := data1[key]; exists {
			map1, ok1 := value1.(map[string]any)
			map2, ok2 := value2.(map[string]any)
			if ok1 && ok2 {
				data1[key] = mergeMaps(map1, map2)
			} else {
				slice1, ok1 := value1.([]any)
				slice2, ok2 := value2.([]any)
				if ok1 && ok2 {
					data1[key] = append(slice1, slice2...)
				} else {
					data1[key] = value2
				}
			}
		} else {
			data1[key] = value2
		}
	}
	return data1
}
