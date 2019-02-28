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
}

// QType is a helper type for better code generation of inputs and outputs.
// It contains a type string and a name.
type QType struct {
	TypeName string
	Type     string
}

// GenDecodeFuncSignatureC generates a function signature for a C translation of the contract function
func (q QFunc) GenDecodeFuncSignatureC(contractName string, usePointers bool) string {
	var inputs []string
	var outputs []string
	for _, input := range q.Inputs {

		if input.Type == "uniaddress" {
			inputs = append(inputs, "UniversalAddressABI* "+input.TypeName)
		} else {
			inputs = append(inputs, input.Type+"_t "+input.TypeName)
		}
	}
	for _, output := range q.Outputs {
		if output.Type == "uniaddress" && usePointers {
			outputs = append(outputs, "UniversalAddressABI** "+output.TypeName)
		} else if usePointers {
			outputs = append(outputs, output.Type+"_t* "+output.TypeName)
		} else {
			outputs = append(outputs, "&"+output.TypeName)
		}

	}
	full := append(inputs, outputs...)
	return contractName + "_" + q.FuncName + "(" + strings.Join(full, ", ") + ")"
}

func (q QFunc) generateFuncCallSignatureC(contractName string) string {
	var sig []string
	for _, input := range q.Inputs {
		sig = append(sig, input.TypeName)
	}
	for _, output := range q.Outputs {
		sig = append(sig, "&"+output.TypeName)
	}
	return contractName + "_" + q.FuncName + "(" + strings.Join(sig, ", ") + ");"
}

// GenEncodeFuncSignatureC could probably be melded together with GenDecodeFuncSignatureC if we make some small changes
func (q QFunc) GenEncodeFuncSignatureC(contractName string) string {
	sigInParens := []string{"UniversalAddress __address", "QtumCallOptions* __options"}
	for _, input := range q.Inputs {
		if input.Type == "uniaddress" {
			sigInParens = append(sigInParens, "UniversalAddressABI* "+input.TypeName)
		} else {
			sigInParens = append(sigInParens, input.Type+"_t "+input.TypeName)
		}
	}

	for _, output := range q.Outputs {
		if output.Type == "uniaddress" {
			sigInParens = append(sigInParens, "UniversalAddressABI** "+output.TypeName)
		} else {
			sigInParens = append(sigInParens, output.Type+"_t* "+output.TypeName)
		}
	}

	return contractName + "_" + q.FuncName + "(" + strings.Join(sigInParens, ", ") + ")"
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
	// push inputs onto stack
	for i, input := range q.Inputs {
		if i == 0 {
			statement = append(statement, "\t"+getQtumPushStatement(input.Type)+"("+input.TypeName+");")
		} else {
			statement = append(statement, getQtumPushStatement(input.Type)+"("+input.TypeName+");")
		}
	}
	statement = append(statement, getQtumPushStatement("int32")+"(ID_"+contractName+"_"+q.FuncName+");")
	statement = append(statement, "QtumCallResult r = qtumCall(__address, __options);")
	statement = append(statement, "if(r.error == QTUM_CALL_SUCCESS){")

	for _, output := range q.Outputs {
		if output.Type == "uniaddress" {
			statement = append(statement, "\tif("+output.TypeName+" == NULL){")
			statement = append(statement, "\t\t"+output.TypeName+" = malloc(sizeof(UniversalAddressABI));")
			statement = append(statement, "\t}")
			statement = append(statement, "\tif("+output.TypeName+" == NULL){")
			statement = append(statement, "\t\tqtumErase();")
			statement = append(statement, "\t}else{")
			statement = append(statement, "\t\tqtumPop("+output.TypeName+", sizeof(UniversalAddressABI));")
			statement = append(statement, "\t}")
		} else {
			statement = append(statement, "\t*"+output.TypeName+" = "+getQtumPopStatement(output.Type)+";")
		}
	}
	statement = append(statement, "}")
	statement = append(statement, "return r;")
	return strings.Join(statement, "\n\t")
}

//GenDispatchCodeC generates the dispatch code for the template in C
func (q QFunc) GenDispatchCodeC(contractName string) string {
	var statement []string
	// Pop off inputs
	for _, input := range q.Inputs {
		popStatement := getQtumPopStatement(input.Type)
		if input.Type == "uniaddress" {
			statement = append(statement, "UniversalAddressABI* "+input.TypeName+" = malloc(sizeof(UniversalAddressABI));")
			statement = append(statement, popStatement+"("+input.TypeName+", sizeof(UniversalAddressABI));")
		} else {
			statement = append(statement, input.Type+"_t "+input.TypeName+" = "+popStatement+";")
		}
	}
	// Declare types with assigned null values
	for _, output := range q.Outputs {
		if output.Type == "uniaddress" {
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
		if output.Type == "uniaddress" {
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
		panic("invalid type string requested")
	}
}
