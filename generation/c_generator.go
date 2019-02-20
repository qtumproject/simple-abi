package generation

import (
	"fmt"
	"io"
	"text/template"

	"github.com/VoR0220/SimpleABI/definitions"
)

// cDecodingTemplateImpl is a template used for generation of a .c file
const cDecodingTemplateImpl = `{{ $contractName := .ContractName }}
#include <stdlib.h>
#include <qtum.h>

//Function IDs
{{range .Functions}}#define ID_{{$contractName}}_{{.FuncName}} {{.GenHashedFuncIdentifier $contractName}}
{{end}}
//prototypes 
{{range $i, $x := .Functions }}void {{.GenDecodeFuncSignatureC $contractName true}};
{{end}}
//dispatch code
void dispatch(){
    uint32_t fn;
    if(qtumPop(&fn, sizeof(fn) != sizeof(fn)){
        //fallback function/error
    }
    switch(fn){
    	{{range $i, $x := .Functions }}case ID_{{$contractName}}_{{- $x.FuncName}}:
    	{
		{{.GenDispatchCodeC $contractName}}
	}{{printf "\n\t"}}{{end -}}
	default:
		//fallback function / error
    }
}`

const cEncodingTemplateImpl = `{{ $contractName := .ContractName }}
#include <stdlib.h>
#include <qtum.h>

//Function IDs
{{range .Functions}}#define ID_{{$contractName}}_{{.FuncName}} {{.GenHashedFuncIdentifier $contractName}}
{{end}}
{{range .Functions}}QtumCallResult  {{.GenEncodeFuncSignatureC $contractName}}{
{{.GenFuncCallQtum $contractName}}
}

{{end}}`

// GenerateTemplate takes in a QInterfaceBuilder, and defines a file for a decoding template to be used
// to generate a file from
func GenerateTemplate(builder definitions.QInterfaceBuilder, name string, output io.Writer, encode bool) error {
	errMsg := "Error in decode template generation: %v"
	var toParse string
	if encode {
		toParse = cEncodingTemplateImpl
	} else {
		toParse = cDecodingTemplateImpl
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
