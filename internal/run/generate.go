package run

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

	"github.com/jinzhu/inflection"
	"github.com/myyrakle/mongery/internal/annotation"
	"github.com/myyrakle/mongery/internal/config"
	"github.com/myyrakle/mongery/internal/utils/cast"
	"github.com/stoewer/go-strcase"
)

func RunGenerate() {
	configFile := config.Load()

	Generate(configFile)
}

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
func convertFieldToConstantCodes(field ProcessFileField, contexts []ProecssFileContext, keyWords []string, valueWords []string, depth uint) []string {
	if depth > 15 {
		return []string{}
	}

	constantCodes := make([]string, 0)

	keyWords = append(keyWords, field.fieldName)
	valueWords = append(valueWords, field.bsonName)

	constantKey := strings.Join(keyWords, "_")
	constantValue := strings.Join(valueWords, ".")

	constantCode := fmt.Sprintf("const %s = \"%s\"", constantKey, constantValue)

	if field.comment != nil && *field.comment != "" {
		constantCode += fmt.Sprintf(" // %s", *field.comment)
	} else {
		constantCode += "\n"
	}

	constantCodes = append(constantCodes, constantCode)

	for _, context := range contexts {
		if context.packageName == field.typePackageName && context.structName == field.typeName {
			for _, field := range context.fields {
				constantCodes = append(constantCodes, convertFieldToConstantCodes(field, contexts, keyWords, valueWords, depth+1)...)
			}
		}
	}

	return constantCodes
}

// 필드 정보를 받아서 ProcessFileField로 변환합니다.
func convertFieldToProcessFileField(structName string, packageName string, field *ast.Field) *ProcessFileField {
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

	if field.Comment != nil {
		comment := field.Comment.Text()
		processFileField.comment = cast.ToPointer(comment)
	}

	// 필드 타입이 포인터인 경우
	if starExpr, ok := field.Type.(*ast.StarExpr); ok {
		processFileField.isPointer = true

		// 패키지가 명시되어 있는 경우
		if selectorExpr, ok := starExpr.X.(*ast.SelectorExpr); ok {
			if xIdent, ok := selectorExpr.X.(*ast.Ident); ok {
				processFileField.typePackageName = xIdent.Name
				processFileField.typeName = selectorExpr.Sel.Name
			}
		} else /* 패키지가 명시되어있지 않은 경우 */ if ident, ok := starExpr.X.(*ast.Ident); ok {
			processFileField.typePackageName = packageName
			processFileField.typeName = ident.Name
		}
	} else /* 필드 타입이 non-pointer에 패키지 명시가 있는 경우 */ if selectorExpr, ok := field.Type.(*ast.SelectorExpr); ok {
		if xIdent, ok := selectorExpr.X.(*ast.Ident); ok {
			processFileField.typePackageName = xIdent.Name
			processFileField.typeName = selectorExpr.Sel.Name
		} else {
			panic("unexpected error")
		}
	} else /* 필드 타입이 non-pointer에 패키지 명시도 없는 경우 */ if ident, ok := field.Type.(*ast.Ident); ok {
		processFileField.typeName = ident.Name
		processFileField.typePackageName = packageName
	}

	return &processFileField
}

type ProcessFileField struct {
	fieldName       string
	bsonName        string
	isPointer       bool
	typePackageName string
	typeName        string
	comment         *string
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
func readFile(configFile config.ConfigFile, packageName string, filename string, file *ast.File) []ProecssFileContext {
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
						processFileField := convertFieldToProcessFileField(structName, packageName, field)

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

var fileWriteMap map[string]struct{} = make(map[string]struct{})

// 단일 파일을 처리하는 단위 함수입니다.
func writeFile(configFile config.ConfigFile, contexts []ProecssFileContext, index int) {
	processFileContext := contexts[index]

	bsonConstantList := make([]string, 0)

	packageName := processFileContext.packageName
	filename := processFileContext.filename
	structName := processFileContext.structName
	collectionConstKey := structName + "Collection"
	collectionConstValue := strcase.SnakeCase(structName)

	if processFileContext.entityParam != nil {
		collectionConstValue = *processFileContext.entityParam
	}

	collectionConst := fmt.Sprintf("const %s = \"%s\"\n", collectionConstKey, collectionConstValue)
	bsonConstantList = append(bsonConstantList, collectionConst)

	// 구조체 필드를 순회하면서 필요한 정보를 추출합니다.
	for _, field := range processFileContext.fields {
		constantCodes := convertFieldToConstantCodes(field, contexts, []string{structName}, []string{}, 0)
		bsonConstantList = append(bsonConstantList, constantCodes...)
	}

	bsonConstantList = append(bsonConstantList, "\n")

	if len(bsonConstantList) > 0 {
		outputFilePath := changeFileSuffix(filename, configFile.OutputSuffix)

		if _, ok := fileWriteMap[outputFilePath]; ok {
			output := ""

			for _, bsonConstant := range bsonConstantList {
				output += bsonConstant
			}

			// 슬라이브 보일러플레이트 생성
			if configFile.Features.Contains(config.FeatureSlice) {
				// Slice named type 생성
				sliceTypeName := inflection.Plural(structName)
				output += fmt.Sprintf("type %s []%s\n\n", sliceTypeName, structName)

				// Len 메서드 생성
				lenMethod := fmt.Sprintf("func (t %s) Len() int {\n", sliceTypeName)
				lenMethod += "\treturn len(t)\n"
				lenMethod += "}\n\n"

				output += lenMethod

				// Append 메서드 생성
				appendMethod := fmt.Sprintf("func (t %s) Append(v %s) %s {\n", sliceTypeName, structName, sliceTypeName)
				appendMethod += "\tt = append(t, v)\n"
				appendMethod += "\treturn t\n"
				appendMethod += "}\n\n"

				output += appendMethod

				// Empty 메서드 생성
				emptyMethod := fmt.Sprintf("func (t %s) Empty() bool {\n", sliceTypeName)
				emptyMethod += "\treturn len(t) == 0\n"
				emptyMethod += "}\n\n"

				output += emptyMethod

				// First 메서드 생성
				firstMethod := fmt.Sprintf("func (t %s) First() %s {\n", sliceTypeName, structName)
				firstMethod += "\tif len(t) == 0 {\n"
				firstMethod += fmt.Sprintf("\t\treturn %s{}\n", structName)
				firstMethod += "\t}\n"
				firstMethod += "\treturn t[0]\n"
				firstMethod += "}\n\n"

				output += firstMethod

				// Last 메서드 생성
				lastMethod := fmt.Sprintf("func (t %s) Last() %s {\n", sliceTypeName, structName)
				lastMethod += "\tif len(t) == 0 {\n"
				lastMethod += fmt.Sprintf("\t\treturn %s{}\n", structName)
				lastMethod += "\t}\n"
				lastMethod += "\treturn t[len(s)-1]\n"
				lastMethod += "}\n\n"

				output += lastMethod
			}

			file, err := os.OpenFile(outputFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
			if err != nil {
				panic(err)
			}
			file.WriteString(output)
			if err := file.Close(); err != nil {
				panic(err)
			}

			fmt.Printf(">>>> write [%s]\n", outputFilePath)
		} else {
			fileWriteMap[outputFilePath] = struct{}{}

			output := ""
			output += "// Code generated by mongery. DO NOT EDIT.\n"
			output += `package ` + packageName
			output += "\n\n"

			for _, bsonConstant := range bsonConstantList {
				output += bsonConstant
			}

			os.WriteFile(outputFilePath, []byte(output), fs.FileMode(0644))

			fmt.Printf(">>>> generated [%s]\n", outputFilePath)
		}

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

func readFileRecursive(basedir string, configFile config.ConfigFile) []ProecssFileContext {
	contexts := make([]ProecssFileContext, 0)

	packages := getPackageList(basedir)

	for packageName, asts := range packages {
		for filename, file := range asts.Files {
			if strings.HasSuffix(filename, "_test.go") {
				continue
			}

			fmt.Printf(">> scan [%s]...\n", filename)
			eachContexts := readFile(configFile, packageName, filename, file)
			contexts = append(contexts, eachContexts...)
		}
	}

	dirList := getDirList(basedir)

	for _, dir := range dirList {
		eachContexts := readFileRecursive(path.Join(basedir, dir), configFile)
		contexts = append(contexts, eachContexts...)
	}

	return contexts
}

func Generate(configFile config.ConfigFile) {
	fmt.Println(">>> scan files...")
	processFileContexts := readFileRecursive(configFile.Basedir, configFile)
	fmt.Println(">>> scan files done")

	fmt.Println(">>> process files...")

	for i := range processFileContexts {
		writeFile(configFile, processFileContexts, i)
	}

	fmt.Println(">>> process files done")
}
