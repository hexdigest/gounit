package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/hexdigest/gounit"
	"github.com/shibukawa/configdir"
)

const defaultTemplateName = "default"

var conf = configdir.New("gounit", "gounit").QueryFolders(configdir.Global)[0]

type Config struct {
	DefaultTemplate string
}

//TemplateCommand implements Command interface
type TemplateCommand struct {
	fs *flag.FlagSet

	templateName     string
	templateFileName string
}

//Description implements Command interface
func (tc *TemplateCommand) Description() string {
	return "manage templates"
}

func (tc *TemplateCommand) Usage() string {
	return `usage: gounit template subcommand [args]

Subcommands usage examples:

	gounit template add <file>
		install a template, file name is used as a template name

	gounit template list
		show all installed templates

	gounit template use <template>
		use selected template by default

	gounit template remove <template>
		remove a template
`
}

func (tc *TemplateCommand) FlagSet() *flag.FlagSet {
	return nil
}

func (tc *TemplateCommand) Run(args []string, stdout, stderr io.Writer) error {
	if len(args) < 1 {
		return gounit.CommandLineError("invalid number of arguments")
	}

	switch args[0] {
	case "add":
		if len(args) < 2 {
			return gounit.CommandLineError("missing file name")
		}
		return installTemplate(args[1])
	case "list":
		return listTemplates()
	case "use":
		if len(args) < 2 {
			return gounit.CommandLineError("missing template name")
		}
		return useTemplate(args[1])
	case "remove":
		if len(args) < 2 {
			return gounit.CommandLineError("missing template name")
		}
		return removeTemplate(args[1])
	}

	return gounit.CommandLineError(fmt.Sprintf("invalid subcommand %q", args[0]))
}

func installTemplate(filename string) error {
	if filename == "" {
		return gounit.CommandLineError("missing file name")
	}

	_, templateName := filepath.Split(filename)
	if templateName == defaultTemplateName {
		return gounit.CommandLineError("can't rewrite default template")
	}

	if err := checkTemplate(filename); err != nil {
		return err
	}

	from, err := os.Open(filename)
	if err != nil {
		return err
	}

	to, err := conf.Create(filepath.Join("templates", templateName))
	if err != nil {
		return err
	}

	defer to.Close()

	if _, err := io.Copy(to, from); err != nil {
		return fmt.Errorf("failed to copy template: %v", err)
	}

	return nil
}

func listTemplates() error {
	names, err := getTemplatesNames()
	if err != nil {
		return err
	}

	templateName, err := getDefaultTemplateName()
	if err != nil {
		return err
	}

	fmt.Printf("\ngounit templates installed\n\n")

	for _, name := range names {
		format := "      %s\n"
		if name == templateName {
			format = "    * %s\n"
		}
		fmt.Printf(format, name)
	}

	fmt.Println()

	return nil
}

func getTemplate(name string) (string, error) {
	var err error

	if name == "" {
		if name, err = getDefaultTemplateName(); err != nil {
			return "", err
		}
	}

	if name == defaultTemplateName {
		return testTemplate, nil
	}

	b, err := ioutil.ReadFile(filepath.Join(conf.Path, "templates", name))
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func useTemplate(name string) error {
	if err := templateExists(name); err != nil {
		return err
	}

	c, err := readConfig()
	if err != nil {
		return err
	}

	c.DefaultTemplate = name
	return writeConfig(*c)
}

func removeTemplate(name string) error {
	if name == defaultTemplateName {
		return gounit.CommandLineError("can't remove default template")
	}

	if err := templateExists(name); err != nil {
		return err
	}

	if err := os.Remove(filepath.Join(conf.Path, "templates", name)); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func templateExists(name string) error {
	names, err := getTemplatesNames()
	if err != nil {
		return err
	}

	for _, n := range names {
		if n == name {
			return nil
		}
	}

	return gounit.CommandLineError("template does not exist: " + name)
}

func getTemplatesNames() ([]string, error) {
	files, err := ioutil.ReadDir(filepath.Join(conf.Path, "templates"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, err
	}

	templates := []string{defaultTemplateName}
	for _, f := range files {
		templates = append(templates, f.Name())
	}

	return templates, nil
}

func getDefaultTemplateName() (string, error) {
	c, err := readConfig()
	if err != nil {
		return "", err
	}

	names, err := getTemplatesNames()
	if err != nil {
		return "", err
	}

	for _, name := range names {
		if name == c.DefaultTemplate {
			return name, nil
		}
	}

	return defaultTemplateName, nil
}

func readConfig() (*Config, error) {
	f, err := conf.Open("config.json")
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}

	decoder := json.NewDecoder(f)
	c := Config{}
	if err := decoder.Decode(&c); err != nil {
		return nil, fmt.Errorf("failed to read configuration: %v", err)
	}

	return &c, nil
}

func writeConfig(c Config) error {
	f, err := conf.Create("config.json")
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(f)
	if err := encoder.Encode(c); err != nil {
		return fmt.Errorf("failed to write configuration: %v", err)
	}

	return nil
}

//checkTemplate executes template against a simple func and
//verifies that produced result is a syntactically correct .go code
func checkTemplate(filename string) error {
	src := strings.NewReader(`
		package funcs

		func function() int {
			return 0
		}`)

	f, err := os.Open(filename)
	if err != nil {
		return err
	}

	contents, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	opt := gounit.Options{
		Template: string(contents),
		All:      true,
	}

	g, err := gounit.NewGenerator(opt, src, nil)
	if err != nil {
		return err
	}

	testSrc := bytes.NewBuffer([]byte{})
	if err := g.Write(testSrc); err != nil {
		return fmt.Errorf("template produces invalid .go file\n%v\n%s", err, g.Source())
	}

	return nil
}

var testTemplate = `{{$func := .Func}}

func {{ $func.TestName }}(t *testing.T) {
	{{- if (gt $func.NumParams 0) }}
		type args struct {
			{{ range $param := params $func }}
				{{- $param}}
			{{ end }}
		}
	{{ end -}}
	tests := []struct {
		name string
		{{- if $func.IsMethod }}
			init func(t *testing.T) {{ ast $func.ReceiverType }}
			inspect func(r {{ ast $func.ReceiverType }}, t *testing.T) //inspects receiver after test run
		{{ end }}
		{{- if (gt $func.NumParams 0) }}
			args func(t *testing.T) args
		{{ end }}
		{{ range $result := results $func}}
			{{ want $result -}}
		{{ end }}
		{{- if $func.ReturnsError }}
			wantErr bool
			inspectErr func (err error, t *testing.T) //use for more precise error evaluation after test
		{{ end -}}
	}{
		{{- if eq .Comment "" }}
			//TODO: Add test cases
		{{else}}
			//{{ .Comment }}
		{{end -}}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			{{- if (gt $func.NumParams 0) }}
				tArgs := tt.args(t)
			{{ end -}}
			{{ if $func.IsMethod }}
				receiver := tt.init(t)
				{{ if (gt $func.NumResults 0) }}{{ join $func.ResultsNames ", " }} := {{end}}receiver.{{$func.Name}}(
					{{- range $i, $pn := $func.ParamsNames }}
						{{- if not (eq $i 0)}},{{end}}tArgs.{{ $pn }}{{ end }})

				if tt.inspect != nil {
					tt.inspect(receiver, t)
				}
			{{ else }}
				{{ if (gt $func.NumResults 0) }}{{ join $func.ResultsNames ", " }} := {{end}}{{$func.Name}}(
					{{- range $i, $pn := $func.ParamsNames }}
						{{- if not (eq $i 0)}},{{end}}tArgs.{{ $pn }}{{ end }})
			{{end}}
			{{ range $result := $func.ResultsNames }}
				{{ if (eq $result "err") }}
					if (err != nil) != tt.wantErr {
						t.Fatalf("{{ receiver $func }}{{ $func.Name }} error = %v, wantErr: %t", err, tt.wantErr)
					}

					if tt.inspectErr!= nil {
						tt.inspectErr(err, t)
					}
				{{ else }}
					if !reflect.DeepEqual({{ $result }}, tt.{{ want $result }}) {
						t.Errorf("{{ receiver $func }}{{ $func.Name }} {{ $result }} = %v, {{ want $result }}: %v", {{ $result }}, tt.{{ want $result }})
					}
				{{end -}}
			{{end -}}
		})
	}
}`
