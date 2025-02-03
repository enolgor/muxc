package main

import (
	"embed"
	"fmt"
	"io"
	"os"
	"path"
	"slices"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

const version string = "v1.0.0"

type Conf struct {
	Package     string            `yaml:"package"`
	Out         string            `yaml:"out"`
	Imports     []string          `yaml:"imports"`
	Args        map[string]string `yaml:"args"`
	Routes      []Routes          `yaml:"routes"`
	Vars        map[string]string `yaml:"vars"`
	PackageName string
	MuxcVersion string
	SourceFile  string
}

type RoutePath string

type ParsedPath struct {
	Method      string
	Pattern     string
	Handler     string
	Middlewares []string
}

var isEmpty func(s string) bool = func(s string) bool { return s == "" }

func (rp RoutePath) Parse() (parsed ParsedPath, err error) {
	parts := strings.Split(string(rp), ";")
	if len(parts) < 2 {
		err = fmt.Errorf("invalid path '%s', it should contain at least pattern and handler parts", string(rp))
		return
	}
	if len(parts) > 3 {
		err = fmt.Errorf("invalid path '%s', middlewares should be comma separated", string(rp))
		return
	}
	parsed.Pattern = strings.TrimSpace(parts[0])
	pattern_parts := strings.Split(parsed.Pattern, " ")
	pattern_parts = slices.DeleteFunc(pattern_parts, isEmpty)
	if len(pattern_parts) > 2 {
		err = fmt.Errorf("invalid path '%s', pattern has more than 2 parts", string(rp))
		return
	}
	if len(pattern_parts) == 2 {
		parsed.Method = pattern_parts[0]
		parsed.Pattern = pattern_parts[1]
	}
	parsed.Handler = strings.TrimSpace(parts[1])
	if len(parts) == 3 {
		mwparts := strings.Split(parts[2], ",")
		parsed.Middlewares = make([]string, len(mwparts))
		for i := range mwparts {
			parsed.Middlewares[i] = strings.TrimSpace(mwparts[i])
		}
	}
	return
}

type Routes struct {
	Use         []string    `yaml:"use"`
	Base        string      `yaml:"base"`
	Paths       []RoutePath `yaml:"paths"`
	ParsedPaths []ParsedPath
}

var (
	//go:embed templates
	tmplFS    embed.FS
	templates *template.Template
)

func init() {
	templates = template.Must(
		template.New("muxc").
			Funcs(template.FuncMap{
				"Join": strings.Join,
				"Slice": func(s string) []string {
					return []string{s}
				},
				"Append": func(slice1 []string, slice2 []string) []string {
					return append(slice1, slice2...)
				},
				"Reverse": func(s []string) []string {
					copy := make([]string, len(s))
					for i := range s {
						copy[len(s)-1-i] = s[i]
					}
					return copy
				},
				"Contains": func(slice []string, s string) bool {
					for i := range slice {
						if slice[i] == s {
							return true
						}
					}
					return false
				},
			}).
			ParseFS(tmplFS, "templates/*.tmpl"),
	)
}

func Generate(yamlFile *MultiYamlFile) error {
	cfgFile, err := yamlFile.Resolve()
	if err != nil {
		return err
	}
	var cfg *Conf
	if cfg, err = parseConf(cfgFile); err != nil {
		return err
	}
	cfg.SourceFile = path.Base(yamlFile.SourceFile)
	cfg.MuxcVersion = version
	if err = createRoutesFile(cfg, yamlFile.BaseDir); err != nil {
		return err
	}
	return nil
}

func parseConf(cfgFile io.Reader) (*Conf, error) {
	dec := yaml.NewDecoder(cfgFile)
	cfg := &Conf{}
	var err error
	if err = dec.Decode(cfg); err != nil {
		return nil, fmt.Errorf("error decoding yaml file: %w", err)
	}
	for i := range cfg.Routes {
		cfg.Routes[i].ParsedPaths = make([]ParsedPath, len(cfg.Routes[i].Paths))
		for j := range cfg.Routes[i].Paths {
			if cfg.Routes[i].ParsedPaths[j], err = cfg.Routes[i].Paths[j].Parse(); err != nil {
				return nil, fmt.Errorf("error parsing route path '%s': %w", cfg.Routes[i].Paths[j], err)
			}
		}
	}
	return cfg, nil
}

func createRoutesFile(cfg *Conf, basedir string) error {
	if err := os.MkdirAll(path.Join(basedir, cfg.Out), os.ModePerm); err != nil {
		return fmt.Errorf("error creating routes directory: %w", err)
	}
	routesFile, err := os.OpenFile(path.Join(basedir, cfg.Out, "routes.go"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating routes.go file: %w", err)
	}
	defer routesFile.Close()
	if err := templates.ExecuteTemplate(routesFile, "routes.go.tmpl", cfg); err != nil {
		return fmt.Errorf("error generating routes.go file: %w", err)
	}
	return nil
}
