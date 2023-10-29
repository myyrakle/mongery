package pkg

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/myyrakle/mongery/internal/annotation"
	"github.com/myyrakle/mongery/pkg/cast"
	"github.com/stoewer/go-strcase"
)

// suffix를 붙여서 새로운 파일명을 만듭니다.
func changeFileSuffix(filePath, newSuffix string) string {
	dir := filepath.Dir(filePath)
	ext := filepath.Ext(filePath)
	base := filepath.Base(filePath)
	name := strings.TrimSuffix(base, ext)
	newName := name + newSuffix
	return filepath.Join(dir, newName)
}

// 패키지 목록을 가져옵니다.
func getPackageList(basePath string) map[string]*ast.Package {
	fset := token.NewFileSet()

	packages, err := parser.ParseDir(fset, basePath, nil, parser.ParseComments)

	if err != nil {
		panic(err)
	}

	return packages
}

// 주석을 읽어와서 @Entity 구조체인지 검증합니다.
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

// @Entity의 파라미터를 가져옵니다.
func getEntityParam(genDecl *ast.GenDecl) *string {
	if genDecl.Doc == nil {
		return nil
	}

	if genDecl.Doc.List == nil {
		return nil
	}

	for _, comment := range genDecl.Doc.List {
		if strings.Contains(comment.Text, "@Entity") {
			params := annotation.ParseParameters(comment.Text)

			if len(params) > 0 {
				return cast.ToPointer(params[0])
			}
		}
	}

	return nil
}

// 필드 정보를 받아서 내보낼 상수 정의 코드로 변환합니다.
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

	if bson == "-" {
		return nil
	}

	bsonTokens := strings.Split(bson, ",")
	bsonName := bsonTokens[0]

	return cast.ToPointer(fmt.Sprintf("const %s_%s = \"%s\"\n", structName, name, bsonName))
}

// 필드 정보를 받아서 ProcessFileField로 변환합니다.
func convertFieldToProcessFileField(structName string, field *ast.Field) *ProcessFileField {
	processFileField := ProcessFileField{}

	if len(field.Names) == 0 {
		return nil
	}

	name := field.Names[0].Name
	processFileField.fieldName = name

	if field.Tag == nil {
		return nil
	}

	tag := strings.ReplaceAll(field.Tag.Value, "`", "")

	bson := reflect.StructTag(tag).Get("bson")

	if bson == "" {
		return nil
	}

	if bson == "-" {
		return nil
	}

	bsonTokens := strings.Split(bson, ",")
	bsonName := bsonTokens[0]

	processFileField.bsonName = bsonName

	if field.Type == nil {
		return nil
	}

	// 필드 타입이 포인터인 경우
	if starExpr, ok := field.Type.(*ast.StarExpr); ok {
		processFileField.isPointer = true

		if selectorExpr, ok := starExpr.X.(*ast.SelectorExpr); ok {
			if xIdent, ok := selectorExpr.X.(*ast.Ident); ok {
				processFileField.typePackageName = cast.ToPointer(xIdent.Name)
				processFileField.typeName = selectorExpr.Sel.Name
			}
		}
	} else /* 필드 타입이 non-pointer에 패키지 명시가 있는 경우 */ if selectorExpr, ok := field.Type.(*ast.SelectorExpr); ok {
		if xIdent, ok := selectorExpr.X.(*ast.Ident); ok {
			processFileField.typePackageName = cast.ToPointer(xIdent.Name)
			processFileField.typeName = selectorExpr.Sel.Name
		}
	} else /* 필드 타입이 non-pointer에 패키지 명시도 없는 경우 */ if ident, ok := field.Type.(*ast.Ident); ok {
		processFileField.typeName = ident.Name
	}

	return &processFileField
}

type ProcessFileField struct {
	fieldName       string
	bsonName        string
	isPointer       bool
	typePackageName *string
	typeName        string
}

type ProecssFileContext struct {
	packageName string
	file        *ast.File
	filename    string
	structName  string
	entityParam *string
	fields      []ProcessFileField
}

// 단일 파일을 읽어서 형식화하는 단위 함수입니다.
func readFile(configFile ConfigFile, packageName string, filename string, file *ast.File) []ProecssFileContext {
	contexts := make([]ProecssFileContext, 0)

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

					entityParam := getEntityParam(genDecl)

					structName := typeSpec.Name.Name

					processFileContext := ProecssFileContext{
						packageName: packageName,
						file:        file,
						filename:    filename,
						structName:  structName,
						entityParam: entityParam,
					}

					// 구조체 필드를 순회하면서 필요한 정보를 추출합니다.
					for _, field := range structDecl.Fields.List {
						processFileField := convertFieldToProcessFileField(structName, field)

						if processFileField != nil {
							processFileContext.fields = append(processFileContext.fields, *processFileField)
						}
					}

					contexts = append(contexts, processFileContext)
				}
			}
		}
	}

	return contexts
}

// 단일 파일을 처리하는 단위 함수입니다.
func processFile(configFile ConfigFile, packageName string, filename string, file *ast.File) {
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

					entityParam := getEntityParam(genDecl)

					structName := typeSpec.Name.Name
					collectionConstKey := structName + "Collection"
					collectionConstValue := strcase.SnakeCase(structName)

					if entityParam != nil {
						collectionConstValue = *entityParam
					}

					collectionConst := fmt.Sprintf("const %s = \"%s\"\n", collectionConstKey, collectionConstValue)
					bsonConstantList = append(bsonConstantList, collectionConst)

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
		output += "// Code generated by mongery. DO NOT EDIT.\n"
		output += `package ` + packageName
		output += "\n\n"

		for _, bsonConstant := range bsonConstantList {
			output += bsonConstant
		}

		os.WriteFile(outputFilePath, []byte(output), fs.FileMode(0644))

		fmt.Printf(">>>> generated [%s]\n", outputFilePath)
	} else {
		fmt.Printf(">>>> no entity struct found in [%s]\n", filename)
	}
}

func getDirList(basePath string) []string {
	dirs, err := os.ReadDir(basePath)
	if err != nil {
		log.Fatal(err)
	}

	var dirList []string
	for _, dir := range dirs {
		if dir.IsDir() {
			dirList = append(dirList, dir.Name())
		}
	}

	return dirList
}

func readFileRecursive(basedir string, configFile ConfigFile) {
	packages := getPackageList(basedir)

	for packageName, asts := range packages {
		for filename, file := range asts.Files {
			if strings.HasSuffix(filename, "_test.go") {
				continue
			}

			fmt.Printf(">> scan [%s]...\n", filename)
			readFile(configFile, packageName, filename, file)
		}
	}

	dirList := getDirList(basedir)

	for _, dir := range dirList {
		readFileRecursive(path.Join(basedir, dir), configFile)
	}
}

func Generate(configFile ConfigFile) {
	fmt.Println(">>> scan files...")

	readFileRecursive(configFile.Basedir, configFile)

	fmt.Println(">>> done")
}
