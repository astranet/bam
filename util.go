package bam

import (
	"go/ast"
	"io/ioutil"
	"os"
	"os/exec"
	"unicode"
)

var capnpKeywords map[string]bool = map[string]bool{
	"Void": true, "Bool": true, "Int8": true, "Int16": true, "Int32": true, "Int64": true, "UInt8": true, "UInt16": true, "UInt32": true, "UInt64": true, "Float32": true, "Float64": true, "Text": true, "Data": true, "List": true, "struct": true, "union": true, "group": true, "enum": true, "AnyPointer": true, "interface": true, "extends": true, "const": true, "using": true, "import": true, "annotation": true}

func isCapnpKeyword(w string) bool {
	return capnpKeywords[w] // not found will return false, the zero value for bool.
}

// recursively extract the go type as a string
func GetTypeAsString(ty ast.Expr, sofar string, goTypeSeq []string) (string, string, []string) {
	switch ty.(type) {

	case (*ast.StarExpr):
		return GetTypeAsString(ty.(*ast.StarExpr).X, sofar+"*", append(goTypeSeq, "*"))

	case (*ast.Ident):
		return sofar, ty.(*ast.Ident).Name, append(goTypeSeq, ty.(*ast.Ident).Name)

	case (*ast.ArrayType):
		// slice or array
		return GetTypeAsString(ty.(*ast.ArrayType).Elt, sofar+"[]", append(goTypeSeq, "[]"))
	}

	return sofar, "", goTypeSeq
}

func underToCamelCase(s string) string {
	ru := []rune(s)
	n := len(ru)
	last := n - 1
	for i := 0; i < n; i++ {
		if ru[i] == '_' && i < last {
			if unicode.IsLower(ru[i+1]) {
				ru[i+1] = unicode.ToUpper(ru[i+1])
			}
			copy(ru[i:], ru[i+1:])
			ru = ru[:last]

			last--
			n--
			i--
		}
	}
	return string(ru)
}

type TempDir struct {
	OrigDir string
	DirPath string
	Files   map[string]*os.File
}

func NewTempDir() *TempDir {
	dirname, err := ioutil.TempDir(".", "testdir_")
	if err != nil {
		panic(err)
	}
	origdir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// add files needed for capnpc -ogo compilation
	exec.Command("/bin/cp", "go.capnp", dirname).Run()

	return &TempDir{
		OrigDir: origdir,
		DirPath: dirname,
		Files:   make(map[string]*os.File),
	}
}

func (d *TempDir) MoveTo() {
	err := os.Chdir(d.DirPath)
	if err != nil {
		panic(err)
	}
}

func (d *TempDir) Close() {
	for _, f := range d.Files {
		f.Close()
	}
}

func (d *TempDir) Cleanup() {
	d.Close()
	err := os.RemoveAll(d.DirPath)
	if err != nil {
		panic(err)
	}
	err = os.Chdir(d.OrigDir)
	if err != nil {
		panic(err)
	}
}

func (d *TempDir) TempFile() *os.File {
	f, err := ioutil.TempFile(d.DirPath, "testfile.")
	if err != nil {
		panic(err)
	}
	d.Files[f.Name()] = f
	return f
}

type ByFinalOrder []*Field

func (s ByFinalOrder) Len() int {
	return len(s)
}

func (s ByFinalOrder) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByFinalOrder) Less(i, j int) bool {
	return s[i].finalOrder < s[j].finalOrder
}

type ByOrderOfAppearance []*Field

func (s ByOrderOfAppearance) Len() int {
	return len(s)
}

func (s ByOrderOfAppearance) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByOrderOfAppearance) Less(i, j int) bool {
	return s[i].orderOfAppearance < s[j].orderOfAppearance
}

type ByGoName []*Struct

func (s ByGoName) Len() int {
	return len(s)
}

func (s ByGoName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByGoName) Less(i, j int) bool {
	return s[i].goName < s[j].goName
}

type AlphaHelper struct {
	Name string
	Code []byte
}

type AlphaHelperSlice []AlphaHelper

func (s AlphaHelperSlice) Len() int {
	return len(s)
}

func (s AlphaHelperSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s AlphaHelperSlice) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}
