package pro_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

func main() {

	src := `package main 
			type Example struct { 
	Foo string` + " `json:\"foo\"` }"

	fmt.Println(src)

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "demo", src, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	fmt.Println(file.Name)

	ast.Inspect(file, func(x ast.Node) bool {
		s, ok := x.(*ast.StructType)
		if !ok {
			return true
		}

		for _, field := range s.Fields.List {
			fmt.Printf("Field: %s\n", field.Names[0].Name)
			fmt.Printf("Field: %s\n", field.Tag.Value)
		}
		return false
	})

}
