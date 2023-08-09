package run

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/myyrakle/mongery/internal/config"
	"github.com/myyrakle/mongery/pkg/cast"
)

func changeFileSuffix(filePath, newSuffix string) string {
	dir := filepath.Dir(filePath)
	ext := filepath.Ext(filePath)
	base := filepath.Base(filePath)
	name := strings.TrimSuffix(base, ext)
	newName := name + newSuffix
	return filepath.Join(dir, newName)
}

func getPackageList(basePath string) map[string]*ast.Package {
	fset := token.NewFileSet()

	packages, err := parser.ParseDir(fset, basePath, nil, parser.ParseComments)

	if err != nil {
		panic(err)
	}

	return packages
}

func isEntityStruct(genDecl *ast.GenDecl) bool {
	if genDecl.Doc == nil {
		return false
	}

	if genDecl.Doc.List == nil {
		return false
	}

	for _, comment := range genDecl.Doc.List {
		if strings.Contains(comment.Text, "@Entity") {
			return true
		}
	}

	return false
}

func convertFieldToConstant(structName string, field *ast.Field) *string {
	if len(field.Names) == 0 {
		return nil
	}

	name := field.Names[0].Name

	if field.Tag == nil {
		return nil
	}

	tag := strings.ReplaceAll(field.Tag.Value, "`", "")

	bson := reflect.StructTag(tag).Get("bson")

	if bson == "" {
		return nil
	}

	bsonTokens := strings.Split(bson, ",")
	bsonName := bsonTokens[0]

	return cast.ToPointer(fmt.Sprintf("const %s_%s = \"%s\"\n", structName, name, bsonName))
}

func processFile(configFile config.ConfigFile, packageName string, filename string, file *ast.File) {
	bsonConstantList := make([]string, 0)

	for _, declare := range file.Decls {
		if genDecl, ok := declare.(*ast.GenDecl); ok {
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					structDecl, _ := typeSpec.Type.(*ast.StructType)

					if structDecl == nil {
						continue
					}

					if !isEntityStruct(genDecl) {
						continue
					}

					structName := typeSpec.Name.Name

					// 구조체 필드를 순회하면서 필요한 정보를 추출합니다.
					for _, field := range structDecl.Fields.List {
						constant := convertFieldToConstant(structName, field)

						if constant != nil {
							bsonConstantList = append(bsonConstantList, *constant)
						}
					}

					bsonConstantList = append(bsonConstantList, "\n")
				}
			}
		}
	}

	if len(bsonConstantList) > 0 {
		outputFilePath := changeFileSuffix(filename, configFile.OutputSuffix)

		output := ""
		output += `package ` + packageName
		output += "\n\n"

		for _, bsonConstant := range bsonConstantList {
			output += bsonConstant
		}

		os.WriteFile(outputFilePath, []byte(output), fs.FileMode(0644))

		fmt.Printf(">> generated [%s]\n", outputFilePath)
	}
}

func Generate() {
	configFile := config.Load()

	fmt.Println(">> scan files...")
	packages := getPackageList(configFile.Basedir)

	for packageName, asts := range packages {
		for filename, file := range asts.Files {
			fmt.Printf(">> scan [%s]...\n", filename)
			processFile(configFile, packageName, filename, file)
		}
	}

	fmt.Println(">>> done")
}
