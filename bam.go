package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

func ParseCmdLine() string {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "bambam needs exactly one golang source file to process.\n")
		os.Exit(1)
	}
	fn := os.Args[1]
	if !strings.HasSuffix(fn, ".go") {
		fmt.Fprintf(os.Stderr, "error: bambam input file '%s' did not end in '.go'.\n")
		os.Exit(1)
	}
	return fn
}

func main() {
	readMe := ParseCmdLine()
	ExtractStructs(readMe)
}

type Extractor struct {
	Out bytes.Buffer
}

func NewExtractor() *Extractor {
	return &Extractor{}
}

func ExtractStructs(readMeFilePath string) []byte {

	x := NewExtractor()
	x.Init()

	fset := token.NewFileSet() // positions are relative to fset

	f, err := parser.ParseFile(fset, readMeFilePath, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	//	fmt.Printf("parsed output f.Decls is:\n")
	//	goon.Dump(f.Decls)

	//fmt.Printf("len(f.Decls) = %d\n", len(f.Decls))

	for _, v := range f.Decls {
		switch v.(type) {
		case *ast.GenDecl:
			d := v.(*ast.GenDecl)
			//fmt.Printf("dump of d, an %#v = \n", ty)
			//goon.Dump(d)

			//fmt.Printf("\n\n\n  detail dump of d.Specs\n")
			for _, spe := range d.Specs {
				switch spe.(type) {
				case (*ast.TypeSpec):

					// go back and print the comments
					if d.Doc != nil && d.Doc.List != nil && len(d.Doc.List) > 0 {
						for _, com := range d.Doc.List {
							x.GenerateComment(com.Text)
						}
					}

					typeSpec := spe.(*ast.TypeSpec)
					//fmt.Printf("\n\n *ast.TypeSpec spe = \n")

					if typeSpec.Name.Obj.Kind == ast.Typ {

						switch typeSpec.Name.Obj.Decl.(type) {
						case (*ast.TypeSpec):
							ts2 := typeSpec.Name.Obj.Decl.(*ast.TypeSpec)
							switch ts2.Type.(type) {
							case (*ast.StructType):
								stru := ts2.Type.(*ast.StructType)

								x.StartStruct(typeSpec.Name.String())
								//fmt.Printf("stru = %#v\n", stru)

								if stru.Fields != nil {
									for _, fld := range stru.Fields.List {
										if fld != nil {
											//fmt.Printf("    fld.Names = %#v\n", fld.Names)
											//goon.Dump(fld.Names)
											for _, ident := range fld.Names {

												switch ident.Obj.Decl.(type) {
												case (*ast.Field):
													fld2 := ident.Obj.Decl.(*ast.Field)
													switch fld2.Type.(type) {
													case (*ast.Ident):
														ident2 := fld2.Type.(*ast.Ident)
														x.GenerateStructField(ident.Name, ident2.Name)
													}
												}
											}

										}
									}
								}

								//fmt.Printf("} // end of %s \n\n", typeSpec.Name) // prod
								fmt.Printf("} \n\n") // prod

								//goon.Dump(stru)
								//fmt.Printf("\n =========== end stru =======\n\n\n")
							}
						}

						//fmt.Printf("spe.Name.Obj.Kind = %s\n", typeSpec.Name.Obj.Kind)

						//fmt.Printf("spe.Name.Obj = %#v\n", typeSpec.Name.Obj)
						//goon.Dump(typeSpec.Name.Obj)

						//fmt.Printf("spe.Name.Obj.Decl = %#v\n", typeSpec.Name.Obj.Decl)
						//goon.Dump(typeSpec.Name.Obj.Decl)
					}

				}
			}
		}
	}

	return x.Out.Bytes()
}

func (x *Extractor) Init() {
	fmt.Printf("package main; func main(){println(`extractor output ran fine.`)}\n")
}
func (x *Extractor) StartStruct(name string) {
	fmt.Printf("type %s struct { ", name)
}
func (x *Extractor) EndStruct(name string) {
	fmt.Printf("}")
}

func (x *Extractor) GenerateComment(c string) {
	skipCommentsForNow := false
	if !skipCommentsForNow {
		fmt.Printf("%s\n", c) // prod
	}
}
func (x *Extractor) GenerateStructField(name string, typeName string) {
	fmt.Printf("%s %s; ", name, typeName) // prod
}

func (x *Extractor) GenerateEmbedded(typeName string) {
	//fmt.Printf("   %s \n",  typeName) // prod
}