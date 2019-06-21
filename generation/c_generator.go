package generation

import (
	"fmt"
	"io"
	"text/template"

	"github.com/qtumproject/SimpleABI/definitions"
)

// cDecodingTemplateImpl is a template used for generation of a .c file
const cDecodingTemplateImpl = `{{ $contractName := .ContractName }}
#include <stdlib.h>
#include <qtum.h>

//Function IDs
{{range .Functions}}#define ID_{{$contractName}}_{{.FuncName}} {{.GenHashedFuncIdentifier $contractName}}
{{end}}
//prototypes 
{{range $i, $x := .Functions }}void {{.GenFuncSignatureC $contractName false}};
{{end}}
//dispatch code
void dispatch(){
    uint32_t fn;
    if(qtumPop(&fn, sizeof(fn) != sizeof(fn))){
        //fallback function/error
    }
    switch(fn){
    	{{range $i, $x := .Functions }}case ID_{{$contractName}}_{{- $x.FuncName}}:
    	{
		{{.GenDispatchCodeC $contractName}}
	}{{printf "\n\t"}}{{end -}}
	default:
		//fallback function / error
		break;
    }
}`

const cEncodingTemplateImpl = `{{ $contractName := .ContractName }}
#include <stdlib.h>
#include <qtum.h>

//Function IDs
{{range .Functions}}#define ID_{{$contractName}}_{{.FuncName}} {{.GenHashedFuncIdentifier $contractName}}
{{end}}
{{range .Functions}}QtumCallResult  {{.GenFuncSignatureC $contractName true}}{
{{.GenFuncCallQtum $contractName}}
}

{{end}}`

const headerEncodingTemplateImpl = `{{ $contractName := .ContractName }}
#ifndef {{$contractName}}ABI_H
#define {{$contractName}}ABI_H

//Function IDs
{{range .Functions}}#ifndef ID_{{$contractName}}_{{.FuncName}}
#define ID_{{$contractName}}_{{.FuncName}} {{.GenHashedFuncIdentifier $contractName}}
#endif
{{end}}

{{range .Functions}}QtumCallResult  {{.GenFuncSignatureC $contractName true}};

{{end}}
#endif`

const headerDecodingTemplateImpl = `{{ $contractName := .ContractName }}
#ifndef {{$contractName}}ABI_H
#define {{$contractName}}ABI_H

//Function IDs
{{range .Functions}}#ifndef ID_{{$contractName}}_{{.FuncName}}
#define ID_{{$contractName}}_{{.FuncName}} {{.GenHashedFuncIdentifier $contractName}}
#endif
{{end}}

void dispatch();

{{range $i, $x := .Functions }}void {{.GenFuncSignatureC $contractName false}};
{{end}}

#endif
`

// TemplateType is an enum used to tell what template the function GenerateTemplate should generate
type TemplateType int

//EncodeC generates a C encoding template, DecodeC generates a C decoding template and so on and so forth
const (
	EncodeC TemplateType = iota
	DecodeC
	EncodeH
	DecodeH
)

// GenerateTemplate takes in a QInterfaceBuilder, and defines a file for a decoding template to be used
// to generate a file from
func GenerateTemplate(builder definitions.QInterfaceBuilder, name string, output io.Writer, typ TemplateType) error {
	errMsg := "Error in decode template generation: %v"
	var toParse string

	switch typ {
	case EncodeC:
		toParse = cEncodingTemplateImpl
	case DecodeC:
		toParse = cDecodingTemplateImpl
	case EncodeH:
		toParse = headerEncodingTemplateImpl
	case DecodeH:
		toParse = headerDecodingTemplateImpl
	default:
		panic("invalid type selected")
	}

	templ, err := template.New(name).Parse(toParse)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}

	err = templ.Execute(output, builder)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}

	return nil
}
