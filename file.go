package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"regexp"
)

var (
	rName       = regexp.MustCompile(`^//\s*@name:\s*(.*)$`)
	rPrimaryKey = regexp.MustCompile(`^//\s*@pk:\s*(.*)$`)
	rStorage = regexp.MustCompile(`^//\s*@storage:\s*(.*)$`)
	rProtoInput = regexp.MustCompile(`(?msU)(.*var fileDescriptor\d+ = \[\]byte{.*}).*`)
)

type MessageData struct {
	Name       string
	PrimaryKey string
	Storage string
}

func fromComment(regex *regexp.Regexp, comment string) string {
	match := regex.FindStringSubmatch(comment)
	if len(match) == 2 {
		return match[1]
	}
	return ""
}

func inputFileFindOrigEOF(contents []byte) []byte {
	match := rProtoInput.FindStringSubmatch(string(contents))
	if len(match) == 2 {
		return []byte(match[1] + string('\n'))
	}
	return []byte("// Failed to Find EOF protoc-go-message-data")
}

func injectStaticStringFunction(contents []byte, name string, ret string, obj string) []byte {
	return append(contents, []byte(fmt.Sprintf("\nfunc (m *%s) %s() string { return \"%s\"}\n", obj, name, ret))...)
}

func parseFile(inputPath string) (results map[string]MessageData, err error) {
	log.Printf("parsing file %q", inputPath)
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, inputPath, nil, parser.ParseComments)
	if err != nil {
		return
	}

	results = make(map[string]MessageData)

	for _, decl := range f.Decls {

		// check if is generic declaration
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		var typeSpec *ast.TypeSpec
		for _, spec := range genDecl.Specs {
			if ts, tsOK := spec.(*ast.TypeSpec); tsOK {
				typeSpec = ts
				break
			}
		}

		// skip if can't get type spec
		if typeSpec == nil {
			continue
		}

		// not a struct, skip
		structDecl, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			continue
		}

		structName := typeSpec.Name.Name
		var pk, name, storage string

		for _, field := range structDecl.Fields.List {
			// skip if field has no doc
			if field.Doc == nil {
				continue
			}
			for _, comment := range field.Doc.List {
				p := fromComment(rPrimaryKey, comment.Text)
				if p != "" {
					pk = p
				}
				n := fromComment(rName, comment.Text)
				if n != "" {
					name = n
				}
				s := fromComment(rStorage, comment.Text)
				if s != "" {
					storage = s
				}
			}
		}
		results[structName] = MessageData{Name: name, PrimaryKey: pk, Storage: storage}
	}

	log.Printf("parsed file %q", inputPath)
	return
}

func writeFile(inputPath string, results map[string]MessageData) (err error) {
	f, err := os.Open(inputPath)
	if err != nil {
		return
	}

	contents, err := ioutil.ReadAll(f)
	if err != nil {
		return
	}

	if err = f.Close(); err != nil {
		return
	}

	contents = inputFileFindOrigEOF(contents)

	for key, data := range results {
		if data.PrimaryKey != "" {
			contents = injectStaticStringFunction(contents, "GetMetaMessagePrimaryKey", data.PrimaryKey, key)
		} else {
			contents = injectStaticStringFunction(contents, "GetMetaMessagePrimaryKey", "ID", key)
		}

		if data.Name != "" {
			contents = injectStaticStringFunction(contents, "GetMetaMessageName", data.Name, key)
		} else {
			contents = injectStaticStringFunction(contents, "GetMetaMessageName", key, key)
		}

		if data.Storage != "" {
			contents = injectStaticStringFunction(contents, "GetMetaMessageStorage", data.Storage, key)
		} else {
			contents = injectStaticStringFunction(contents, "GetMetaMessageStorage", "protobuf", key)
		}
	}

	if err = ioutil.WriteFile(inputPath, contents, 0644); err != nil {
		return
	}

	if len(results) > 0 {
		log.Printf("file %q was customized", inputPath)
	}
	return
}
