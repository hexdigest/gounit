package gounit

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/tools/imports"
)

var (
	ErrGenerateHeader        = GenericError("failed to write header: %v")
	ErrGenerateTest          = GenericError("failed to write test: %v")
	ErrFuncNotFound          = GenericError("unable to find a function declaration")
	ErrFailedToParseInFile   = GenericError("failed to parse input file: %v")
	ErrFailedToParseOutFile  = GenericError("failed to parse output file: %v")
	ErrFailedToOpenInFile    = GenericError("failed to open input file: %v")
	ErrFailedToOpenOutFile   = GenericError("failed to open output file: %v")
	ErrFailedToCreateOutFile = GenericError("failed to create output file: %v")
	ErrInputFileDoesNotExist = GenericError("input file does not exist")
	ErrSeekFailed            = GenericError("failed to seek: %v")
	ErrFixImports            = GenericError("failed to fix imports: %v")
	ErrWriteTest             = GenericError("failed to write generated test: %v")
	ErrInvalidTestTemplate   = GenericError("invalid test template: %v")
)

type Options struct {
	Lines      []int
	Functions  []string
	InputFile  string
	OutputFile string
	Comment    string
	Template   string
	All        bool
	UseJSON    bool
	UseStdin   bool
	UseStdout  bool
}

//Generator is used to generate a test stub for function Func
type Generator struct {
	fs             *token.FileSet
	funcs          []*Func
	imports        []*ast.ImportSpec
	pkg            string
	opt            Options
	buf            *bytes.Buffer
	headerTemplate *template.Template
	testTemplate   *template.Template
}

//NewGenerator returns a pointer to Generator
func NewGenerator(opt Options, src, testSrc io.Reader) (*Generator, error) {
	fs := token.NewFileSet()
	file, err := parser.ParseFile(fs, opt.InputFile, src, 0)

	srcPackageName := file.Name.String()
	if srcPackageName == "" {
		return nil, ErrFailedToParseInFile.Format(err)
	}

	funcs := findFunctions(file.Decls, func(fd *ast.FuncDecl) bool {
		if opt.All {
			return true
		}

		for _, l := range opt.Lines {
			if l == fs.Position(fd.Pos()).Line {
				return true
			}
		}

		for _, f := range opt.Functions {
			if f == fd.Name.Name {
				return true
			}
		}

		return false
	})

	if len(funcs) == 0 {
		return nil, ErrFuncNotFound
	}

	var (
		buf            = bytes.NewBuffer([]byte{})
		dstPackageName = srcPackageName
	)

	if testSrc != nil {
		tr := io.TeeReader(testSrc, buf)

		//parsing source buffer as it can differ from the actual file in the package
		file, err := parser.ParseFile(fs, opt.OutputFile, tr, 0)
		if err != nil {
			return nil, ErrFailedToParseOutFile.Format(err)
		}

		//using package name from the destination file since it can be a *_test package
		dstPackageName = file.Name.String()
		funcs = findMissingTests(file, funcs)
	}

	//this filter leaves only test files so we can ignore syntax errors in the tested code
	//but we still have to fail when test files contain syntax errors because it's not possible
	//to identify missing tests in such case
	filter := func(fi os.FileInfo) bool {
		if fi.IsDir() {
			return false
		}

		if strings.HasSuffix(fi.Name(), "_test.go") {
			return true
		}

		f, err := os.Open(fi.Name())
		if err != nil {
			return false
		}
		defer f.Close()

		astFile, _ := parser.ParseFile(fs, fi.Name(), f, parser.PackageClauseOnly)

		return astFile.Name.String() == srcPackageName+"_test"
	}

	packages, err := parser.ParseDir(fs, filepath.Dir(opt.OutputFile), filter, 0)
	if err != nil {
		return nil, ErrFailedToParseOutFile.Format(err)
	}

	for _, pkg := range packages {
		for _, file := range pkg.Files {
			funcs = findMissingTests(file, funcs)
		}
	}

	testTemplate, err := template.New("test").Funcs(templateHelpers(fs)).Parse(opt.Template)
	if err != nil {
		return nil, ErrInvalidTestTemplate.Format(err)
	}

	return &Generator{
		buf:            buf,
		opt:            opt,
		fs:             fs,
		funcs:          funcs,
		imports:        file.Imports,
		pkg:            dstPackageName,
		headerTemplate: template.Must(template.New("header").Funcs(templateHelpers(fs)).Parse(headerTemplate)),
		testTemplate:   testTemplate,
	}, nil
}

func (g *Generator) Write(w io.Writer) error {
	if len(g.funcs) == 0 {
		return nil
	}

	if g.buf.Len() == 0 {
		if err := g.WriteHeader(g.buf); err != nil {
			return ErrGenerateHeader.Format(err)
		}
	}

	if err := g.WriteTests(g.buf); err != nil {
		return ErrGenerateTest.Format(err)
	}

	formattedSource, err := imports.Process(g.opt.OutputFile, g.buf.Bytes(), nil)
	if err != nil {
		return ErrFixImports.Format(err)
	}

	if _, err = w.Write(formattedSource); err != nil {
		return ErrWriteTest.Format(err)
	}

	return nil
}

func (g *Generator) Source() string {
	return g.buf.String()
}

//WriteHeader writes a package name and import specs
func (g *Generator) WriteHeader(w io.Writer) error {
	return g.headerTemplate.Execute(w, struct {
		Imports []*ast.ImportSpec
		Package string
	}{
		Imports: g.imports,
		Package: g.pkg,
	})
}

//WriteTests writes test stubs for every function that don't have test yet
func (g *Generator) WriteTests(w io.Writer) error {
	for _, f := range g.funcs {
		err := g.testTemplate.Execute(w, struct {
			Func    *Func
			Comment string
		}{
			Func:    f,
			Comment: g.opt.Comment,
		})

		if err != nil {
			return fmt.Errorf("failed to write test: %v", err)
		}
	}

	return nil
}

var headerTemplate = `package {{.Package}}

import(
	"testing"
	"reflect"{{range $import := .Imports}}
	{{ast $import}}{{end}}
)`
