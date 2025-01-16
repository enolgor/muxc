package merge

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

func MergeYaml(sourceFile string, cfgFile io.Reader, basedir string) (io.Reader, error) {
	yamlFile, err := resolveIncludes(path.Base(sourceFile), cfgFile, basedir, nil)
	if err != nil {
		return nil, err
	}
	return yamlFile.Resolve()
}

type YamlFile struct {
	FilePath string
	Data     []byte
	decoded  map[string]any
	Includes []*YamlFile
}

func (yf *YamlFile) Resolve() (io.Reader, error) {
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

func (yf *YamlFile) merge() map[string]any {
	for i := range yf.Includes {
		yf.decoded = mergeMaps(yf.decoded, yf.Includes[i].merge())
	}
	return yf.decoded
}

func (yf *YamlFile) decode() error {
	yf.decoded = map[string]any{}
	if err := yaml.Unmarshal(yf.Data, &yf.decoded); err != nil {
		return err
	}
	for i := range yf.Includes {
		if err := yf.Includes[i].decode(); err != nil {
			return err
		}
	}
	return nil
}

func (yf *YamlFile) String() string {
	buffer := &bytes.Buffer{}
	fmt.Fprintf(buffer, "-----------------------------\n")
	fmt.Fprintf(buffer, "File: %s\n", yf.FilePath)
	fmt.Fprintf(buffer, "Includes: ")
	for i := range yf.Includes {
		fmt.Fprintf(buffer, "%s ", yf.Includes[i].FilePath)
	}
	fmt.Fprintf(buffer, "\n-----------------------------\n")
	fmt.Fprintf(buffer, "%s\nEOF\n\n", string(yf.Data))
	for i := range yf.Includes {
		fmt.Fprint(buffer, yf.Includes[i])
	}
	return buffer.String()
}

var includeMatcher *regexp.Regexp = regexp.MustCompile(`^!include "*([^"]+)"*$`)

func resolveIncludes(filename string, file io.Reader, basedir string, alreadyIncluded []string) (*YamlFile, error) {
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
	yamlFile := &YamlFile{
		FilePath: filePath,
		Data:     buffer.Bytes(),
		Includes: []*YamlFile{},
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
		yamlFile.Includes = append(yamlFile.Includes, included)
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
