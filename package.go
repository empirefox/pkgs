// +build go1.5

package pkgs

import (
	"go/ast"
	"go/build"
	"go/doc"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/firewut/go-json-map"
	"github.com/mitchellh/mapstructure"
	"github.com/rolldever/go-json5"
)

var log = logrus.New()

type Package struct {
	Dir      string
	Name     string
	TypesPkg *types.Package

	Tools map[string]*JsonOptions
	Dst   string

	Doc         *doc.Package
	BasicTypes  map[string]*Basic
	StructTypes map[string]*Struct
	ArrayTypes  map[string]*Array
	MapTypes    map[string]*Map
}

func NewPackage(files ...string) *Package {

	// We accept either one directory or a list of files. Which do we have?
	if len(files) == 0 {
		// Default: process whole package in current directory.
		files = []string{"."}
	}

	// Parse the package once.
	p := &Package{
		BasicTypes:  make(map[string]*Basic),
		StructTypes: make(map[string]*Struct),
		ArrayTypes:  make(map[string]*Array),
		MapTypes:    make(map[string]*Map),
	}

	//	for _, t := range typeNames {
	//		if s := strings.TrimSpace(t); s != "" {
	//			p.InTypes = append(p.InTypes, s)
	//		}
	//	}

	if len(files) == 1 && isDirectory(files[0]) {
		p.Dst = files[0]
		p.parsePackageDir(files[0])
	} else {
		p.Dst = filepath.Dir(files[0])
		p.parsePackageFiles(files)
	}

	p.generateTypes()

	return p

}

// parsePackageDir parses the package residing in the directory.
func (p *Package) parsePackageDir(directory string) {
	pkg, err := build.Default.ImportDir(directory, 0)
	if err != nil {
		log.Fatalf("cannot process directory %s: %s", directory, err)
	}
	var names []string
	names = append(names, pkg.GoFiles...)
	names = append(names, pkg.CgoFiles...)
	// TODO: Need to think about constants in test files. Maybe write type_string_test.go
	// in a separate pass? For later.
	// names = append(names, pkg.TestGoFiles...) // These are also in the "foo" package.
	names = append(names, pkg.SFiles...)
	names = prefixDirectory(directory, names)
	p.parsePackage(directory, names, nil)
}

// parsePackageFiles parses the package occupying the named files.
func (p *Package) parsePackageFiles(names []string) {
	p.parsePackage(".", names, nil)
}

// prefixDirectory places the directory name on the beginning of each name in the list.
func prefixDirectory(directory string, names []string) []string {
	if directory == "." {
		return names
	}
	ret := make([]string, len(names))
	for i, name := range names {
		ret[i] = filepath.Join(directory, name)
	}
	return ret
}

// parsePackage analyzes the single package constructed from the named files.
// If text is non-nil, it is a string to be used instead of the content of the file,
// to be used for testing. parsePackage exits if there is an error.
func (p *Package) parsePackage(directory string, names []string, text interface{}) {
	fs := token.NewFileSet()
	astFiles, files := p.parseAstFiles(fs, names, text)
	p.Name = astFiles[0].Name.Name
	p.Dir = directory

	// Type check the package.
	p.check(fs, astFiles)

	fs = token.NewFileSet()
	astFiles, files = p.parseAstFiles(fs, names, text)
	astPkg, err := ast.NewPackage(fs, files, nil, nil)
	if err != nil {
		log.Errorf("new ast package: %s: %s", p.Name, err)
	}
	p.Doc = doc.New(astPkg, astFiles[0].Name.String(), doc.AllDecls)
}

func (p *Package) parseAstFiles(fs *token.FileSet, names []string, text interface{}) (astFiles []*ast.File, fileMap map[string]*ast.File) {
	fileMap = make(map[string]*ast.File)

	p.parseJsonFile(filepath.Dir(names[0]))

	for _, name := range names {
		if !strings.HasSuffix(name, ".go") {
			continue
		}
		parsedFile, err := parser.ParseFile(fs, name, text, parser.ParseComments)
		if err != nil {
			log.Fatalf("parsing package: %s: %s", name, err)
		}
		astFiles = append(astFiles, parsedFile)
		fileMap[name] = parsedFile
	}
	if len(astFiles) == 0 {
		log.Fatalf("%s: no buildable Go files", strings.Join(names, "\n\t"))
	}
	return
}

// check type-checks the package. The package must be OK to proceed.
func (p *Package) check(fs *token.FileSet, astFiles []*ast.File) {
	config := types.Config{Importer: importer.Default(), FakeImportC: true}
	TypesPkg, err := config.Check(p.Dir, fs, astFiles, nil)
	if err != nil {
		log.Fatalf("checking package: %s", err)
	}
	p.TypesPkg = TypesPkg
}

// generateTypes produces the String method for the named type.
func (p *Package) generateTypes() {
	// Find all named types at package level.
	scope := p.TypesPkg.Scope()
	for _, name := range scope.Names() {
		if obj, ok := scope.Lookup(name).(*types.TypeName); ok {
			switch t := obj.Type().(*types.Named).Underlying().(type) {
			case *types.Basic:
				if typ := NewBasic(name, t); typ != nil {
					p.BasicTypes[name] = typ
				}
			case *types.Struct:
				if typ := NewStruct(name, t, p); typ != nil {
					p.StructTypes[name] = typ
				}
			case *types.Array:
				if typ := NewArray(name, t.Elem(), scope); typ != nil {
					p.ArrayTypes[name] = typ
				}
			case *types.Slice:
				if typ := NewArray(name, t.Elem(), scope); typ != nil {
					p.ArrayTypes[name] = typ
				}
			case *types.Map:
				if typ := NewMap(name, t.Key(), t.Elem(), scope); typ != nil {
					p.MapTypes[name] = typ
				}
			default:
				log.WithField(name, t.String()).Infoln("ignore other type")
			}
		}
	}

	// Load doc for all Types
	for _, t := range p.Doc.Types {
		if typ, ok := p.BasicTypes[t.Name]; ok {
			typ.Doc = t.Doc
		} else if typ, ok := p.StructTypes[t.Name]; ok {
			typ.Doc = t.Doc
		} else if typ, ok := p.ArrayTypes[t.Name]; ok {
			typ.Doc = t.Doc
		} else if typ, ok := p.MapTypes[t.Name]; ok {
			typ.Doc = t.Doc
		}
	}

	// process json options
	for _, opt := range p.Tools {
		opt.process(p)
	}
}

func (p *Package) parseJsonFile(directory string) {
	js, err := p.readJson(path.Join(directory, "autogens.json5"))
	if js == nil {
		js, err = p.readJson(path.Join(directory, "autogens.json"))
	}
	if err != nil {
		log.WithField("directory", directory).Fatal(err)
	}
	p.Tools = js
}

func (p *Package) readJson(jsonPath string) (map[string]*JsonOptions, error) {
	jsraw := make(map[string]interface{})

	bDoc, err := ioutil.ReadFile(jsonPath)
	if err != nil {
		return nil, err
	}
	if err = json5.Unmarshal(bDoc, &jsraw); err != nil {
		return nil, err
	}

	js := make(map[string]*JsonOptions)
	for tool, optraw := range jsraw {
		var opt JsonOptions

		if err := mapstructure.Decode(optraw, &opt); err != nil {
			log.WithFields(logrus.Fields{
				"tool":  tool,
				"error": err,
			}).Fatal("parse json option failed")
		}

		js[tool] = &opt
	}

	return js, nil
}

func (p *Package) supported(name string) bool {
	if _, ok := p.BasicTypes[name]; ok {
		return true
	}
	if _, ok := p.StructTypes[name]; ok {
		return true
	}
	if _, ok := p.ArrayTypes[name]; ok {
		return true
	}
	if _, ok := p.MapTypes[name]; ok {
		return true
	}
	return false
}

func (p *Package) shouldIgnore(name string, ignores []string) bool {
	for _, ign := range ignores {
		yes := false
		switch ign {
		case "basic":
			_, yes = p.BasicTypes[name]
		case "struct":
			_, yes = p.StructTypes[name]
		case "array":
			_, yes = p.ArrayTypes[name]
		case "array*":
			typ, ok := p.ArrayTypes[name]
			yes = ok && typ.IsPtr
		case "array&":
			typ, ok := p.ArrayTypes[name]
			yes = ok && !typ.IsPtr
		case "struct_array":
			typ, ok := p.ArrayTypes[name]
			yes = ok && typ.IsStruct
		case "struct_array*":
			typ, ok := p.ArrayTypes[name]
			yes = ok && typ.IsStruct && typ.IsPtr
		case "struct_array&":
			typ, ok := p.ArrayTypes[name]
			yes = ok && typ.IsStruct && !typ.IsPtr
		case "basic_array":
			typ, ok := p.ArrayTypes[name]
			yes = ok && !typ.IsStruct
		case "basic_array*":
			typ, ok := p.ArrayTypes[name]
			yes = ok && !typ.IsStruct && typ.IsPtr
		case "basic_array&":
			typ, ok := p.ArrayTypes[name]
			yes = ok && !typ.IsStruct && !typ.IsPtr
		case "map":
			_, yes = p.MapTypes[name]
		case "map*":
			typ, ok := p.MapTypes[name]
			yes = ok && typ.IsPtr
		case "map&":
			typ, ok := p.MapTypes[name]
			yes = ok && !typ.IsPtr
		case "struct_map":
			typ, ok := p.MapTypes[name]
			yes = ok && typ.IsStruct
		case "struct_map*":
			typ, ok := p.MapTypes[name]
			yes = ok && typ.IsStruct && typ.IsPtr
		case "struct_map&":
			typ, ok := p.MapTypes[name]
			yes = ok && typ.IsStruct && !typ.IsPtr
		case "basic_map":
			typ, ok := p.MapTypes[name]
			yes = ok && !typ.IsStruct
		case "basic_map*":
			typ, ok := p.MapTypes[name]
			yes = ok && !typ.IsStruct && typ.IsPtr
		case "basic_map&":
			typ, ok := p.MapTypes[name]
			yes = ok && !typ.IsStruct && !typ.IsPtr
		}
		if yes {
			return true
		}
	}

	return false
}

type JsonOptions struct {
	Command string
	Presets map[string]interface{}
	Ignore  []string
	Ignored map[string]interface{}
	Types   map[string]interface{}
	Data    interface{}
}

// process check supported type and load preset options
func (opt *JsonOptions) process(p *Package) {
	ignores := strings.Split(strings.Join(opt.Ignore, ","), ",")
	for i := range ignores {
		ignores[i] = strings.TrimSpace(ignores[i])
	}
	opt.Ignored = make(map[string]interface{})
	for typ, typOpt := range opt.Types {
		if !p.supported(typ) {
			log.WithField("type", typ).Fatal("Not a supported type")
		}

		if quote, ok := typOpt.(string); ok && strings.HasPrefix(quote, "&") {
			preset, err := gjm.GetProperty(opt.Presets, quote[1:])
			if err != nil {
				log.WithFields(logrus.Fields{
					"type":   typ,
					"preset": quote,
					"error":  err,
				}).Fatal("Preset not found")
			}
			opt.Types[typ] = preset
		}

		if p.shouldIgnore(typ, ignores) {
			opt.Ignored[typ] = opt.Types[typ]
			delete(opt.Types, typ)
		}
	}
}

// isDirectory reports whether the named file is a directory.
func isDirectory(name string) bool {
	info, err := os.Stat(name)
	if err != nil {
		logrus.Fatal(err)
	}
	return info.IsDir()
}
