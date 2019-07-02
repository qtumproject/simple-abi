package definitions

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

// QInterfaceBuilder is a struct that is created by the parser.
// It is to be used in the template building stage where all the variables are assembled into a solid interface for a contract.
type QInterfaceBuilder struct {
	ContractName string
	Functions    []QFunc
}

// QFunc is a function as defined in the SimpleABI protocol
// It contains a name, inputs, and outputs.
type QFunc struct {
	FuncName string
	Inputs   []QType
	Outputs  []QType
	Payable  bool
}

// QType is a helper type for better code generation of inputs and outputs.
// It contains a type string and a name.
type QType struct {
	TypeName string
	Type     string
}

// GenFuncSignatureC generates a function signature to be used in templating. Takes a contract name to complete the function signature
// and an addCallOpts to prefix the various inputs and outputs with a UniversalAddress __address and QtumCallOptions* __options
func (q QFunc) GenFuncSignatureC(contractName string, addCallOpts bool) string {
	var sigInParens []string

	if addCallOpts {
		sigInParens = append(sigInParens, []string{"UniversalAddress *__address", "QtumCallOptions* __options"}...)
	}
	for _, input := range q.Inputs {
		if isArray(input.Type) {
			sigInParens = append(sigInParens, getBaseType(input.Type)+"_t* "+input.TypeName)
			sigInParens = append(sigInParens, "size_t "+input.TypeName+"_sz")
		} else if input.Type == "uniaddress" {
			sigInParens = append(sigInParens, "UniversalAddressABI* "+input.TypeName)
		} else {
			sigInParens = append(sigInParens, input.Type+"_t "+input.TypeName)
		}
	}

	for _, output := range q.Outputs {
		if isArray(output.Type) {
			sigInParens = append(sigInParens, getBaseType(output.Type)+"_t** "+output.TypeName)
			sigInParens = append(sigInParens, "size_t* "+output.TypeName+"_sz")
		} else if output.Type == "uniaddress" {
			sigInParens = append(sigInParens, "UniversalAddressABI** "+output.TypeName)
		} else {
			sigInParens = append(sigInParens, output.Type+"_t* "+output.TypeName)
		}
	}

	return contractName + "_" + q.FuncName + "(" + strings.Join(sigInParens, ", ") + ")"
}

func (q QFunc) generateFuncCallSignatureC(contractName string) string {
	var sig []string
	for _, input := range q.Inputs {
		sig = append(sig, input.TypeName)
		if(isArray(input.Type)){
			sig = append(sig, input.TypeName + "_sz");
		}
	}
	for _, output := range q.Outputs {
		sig = append(sig, "&" + output.TypeName)
		if(isArray(output.Type)){
			sig = append(sig, "&" + output.TypeName + "_sz");
		}
	}
	return contractName + "_" + q.FuncName + "(" + strings.Join(sig, ", ") + ");"
}

// GenHashedFuncIdentifier generates a hashed function identifier from a function signature
func (q QFunc) GenHashedFuncIdentifier(contractName string) string {
	var toHashArr []string
	for _, input := range q.Inputs {
		toHashArr = append(toHashArr, input.Type)
	}
	toHashArr = append(toHashArr, contractName+"_"+q.FuncName)
	toHashArr = append(toHashArr, "->")
	for _, output := range q.Outputs {
		toHashArr = append(toHashArr, output.Type)
	}
	toHash := []byte(strings.Join(toHashArr, " "))
	h := sha256.New()
	h.Write(toHash)
	funcHash := h.Sum(nil)
	return fmt.Sprintf("0x%x", funcHash[:4])
}

// GenFuncCallQtum creates a function body for a Qtum Function Call in C
func (q QFunc) GenFuncCallQtum(contractName string) string {
	var statement []string
	if !q.Payable {
		statement = append(statement, "if(__options->value > 0) {")
		statement = append(statement, "\tqtumError(\"nonpayable function\");")
		statement = append(statement, "}")
	}
	// push inputs onto stack
	for i, input := range q.Inputs {
		var pushStatement string
		if isArray(input.Type) {
			pushStatement = getQtumPushStatement(input.Type) + "(" + input.TypeName + ", " + input.TypeName + "_sz);"
		} else {
			pushStatement = getQtumPushStatement(input.Type) + "(" + input.TypeName + ");"
		}

		if i == 0 {
			statement = append(statement, "\t"+pushStatement)
		} else {
			statement = append(statement, pushStatement)
		}
	}
	statement = append(statement, getQtumPushStatement("int32")+"(ID_"+contractName+"_"+q.FuncName+");")
	statement = append(statement, "QtumCallResult r = qtumCall(__address, __options);")
	statement = append(statement, "if(r.error == QTUM_CALL_SUCCESS){")

	for _, output := range q.Outputs {
		statement = append(statement, output.generateFuncCallBody()...)
	}
	statement = append(statement, "}")
	statement = append(statement, "return r;")
	return strings.Join(statement, "\n\t")
}

func (typ QType) generateFuncCallBody() []string {
	switch {
	case isArray(typ.Type):
		return []string{
			fmt.Sprintf("\t*%v_sz = qtumPeekSize();", typ.TypeName),
			fmt.Sprintf("\t*%v = malloc(*%v_sz * sizeof(**%v));", typ.TypeName, typ.TypeName, typ.TypeName),
			fmt.Sprintf("\t%v(*%v, *%v_sz * sizeof(**%v));", getQtumPopStatement(typ.Type), typ.TypeName, typ.TypeName, typ.TypeName),
			fmt.Sprintf("\t*%v_sz /= sizeof(**%v);", typ.TypeName, typ.TypeName),
		}
	case typ.Type == "uniaddress":
		return []string{
			fmt.Sprintf("\tif(%v == NULL){", typ.TypeName),
			fmt.Sprintf("\t\t%v = malloc(sizeof(UniversalAddressABI));", typ.TypeName),
			"\t}",
			fmt.Sprintf("\tif(%v == NULL){", typ.TypeName),
			"\t\tqtumErase();",
			"\t}else{",
			fmt.Sprintf("\t\tqtumPop(%v, sizeof(UniversalAddressABI));", typ.TypeName),
			"\t}",
		}
	default:
		return []string{fmt.Sprintf("\t*%v = %v;", typ.TypeName, getQtumPopStatement(typ.Type))}
	}
}

//GenDispatchCodeC generates the dispatch code for the template in C
func (q QFunc) GenDispatchCodeC(contractName string) string {
	var statement []string
	if !q.Payable {
		statement = append(statement, "if(qtumExec->valueSent > 0) {")
		statement = append(statement, "\tqtumError(\"nonpayable function\");")
		statement = append(statement, "}")
	}
	// Pop off inputs
	for _, input := range q.Inputs {
		popStatement := getQtumPopStatement(input.Type)
		if isArray(input.Type) {
			statement = append(statement, getBaseType(input.Type)+"_t* "+input.TypeName+";")
			statement = append(statement, "size_t "+input.TypeName+"_sz = qtumPeekSize();")
			statement = append(statement, input.TypeName+" = malloc("+input.TypeName+"_sz);")
			statement = append(statement, popStatement+"("+input.TypeName+", "+input.TypeName+"_sz);")
		} else if input.Type == "uniaddress" {
			statement = append(statement, "UniversalAddressABI* "+input.TypeName+" = malloc(sizeof(UniversalAddressABI));")
			statement = append(statement, popStatement+"("+input.TypeName+", sizeof(UniversalAddressABI));")
		} else {
			statement = append(statement, input.Type+"_t "+input.TypeName+" = "+popStatement+";")
		}
	}
	// Declare types with assigned null values
	for _, output := range q.Outputs {
		if isArray(output.Type) {
			statement = append(statement, output.Type+"_t* "+output.TypeName+" = NULL;")
			statement = append(statement, "size_t "+output.TypeName+"_sz;")
		} else if output.Type == "uniaddress" {
			statement = append(statement, "UniversalAddressABI* "+output.TypeName+" = NULL;")
		} else {
			statement = append(statement, output.Type+"_t "+output.TypeName+" = 0;")
		}
	}
	// append function call
	statement = append(statement, q.generateFuncCallSignatureC(contractName))
	// append push statements for outputs
	for _, output := range q.Outputs {
		pushStatement := getQtumPushStatement(output.Type)
		if isArray(output.Type) {
			statement = append(statement, pushStatement+"("+output.TypeName+", "+output.TypeName+"_sz * sizeof(*"+output.TypeName+"));")
		} else if output.Type == "uniaddress" {
			statement = append(statement, pushStatement+"("+output.TypeName+", sizeof(UniversalAddressABI));")
		} else {
			statement = append(statement, pushStatement+"("+output.TypeName+");")
		}
	}
	statement = append(statement, "break;")
	return strings.Join(statement, "\n\t\t")
}

func getQtumPushStatement(typ string) string {
	switch typ {
	case "uint8", "int8":
		return "qtumPush8"
	case "uint16", "int16":
		return "qtumPush16"
	case "uint32", "int32":
		return "qtumPush32"
	case "uint64", "int64":
		return "qtumPush64"
	default:
		return "qtumPush"
	}
}

func getQtumPopStatement(typ string) string {
	switch typ {
	case "uint8", "int8":
		return "qtumPop8()"
	case "uint16", "int16":
		return "qtumPop16()"
	case "uint32", "int32":
		return "qtumPop32()"
	case "uint64", "int64":
		return "qtumPop64()"
	case "uniaddress":
		return "qtumPopExact"
	default:
		return "qtumPop"
	}
}

func isArray(typ string) bool {
	return strings.HasSuffix(typ, "[]")
}

func getBaseType(typ string) string {
	return strings.TrimSuffix(typ, "[]")
}
