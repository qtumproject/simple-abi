package parser

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/VoR0220/SimpleABI/definitions"
)

// components to denote where to put our strings when it comes time to assemble what we've parsed
const (
	NAME int = iota
	FUNCTION
)

// Parse opens up a file and returns a QInterfaceBuilder for building of templates
func Parse(filename string) (definitions.QInterfaceBuilder, error) {
	file, err := os.Open(filename)
	if err != nil {
		return definitions.QInterfaceBuilder{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var counter int
	var builtInterface definitions.QInterfaceBuilder
	for scanner.Scan() {
		isName, returned, err := parseLine(scanner.Text(), counter)
		if err != nil {
			return definitions.QInterfaceBuilder{}, err
		} else if isName {
			if builtInterface.ContractName != "" {
				return definitions.QInterfaceBuilder{}, fmt.Errorf("Attempted to declare multiple names for contract %v. Only one contract name allowed per instance.", builtInterface.ContractName)
			}
			builtInterface.ContractName = returned.(string)
		} else if !isName && returned == nil {
			continue
		} else {
			funcs := builtInterface.Functions
			builtInterface.Functions = append(funcs, returned.(definitions.QFunc))
		}
		counter++
	}
	return builtInterface, nil
}

// parseLine is a function that is used to create an interface builder from a line from a file
// the first output argument is a boolean to determine whether or not this is a name,
// the second output argument is an  interface that should be either a string or a qFunc
func parseLine(input string, number int) (bool, interface{}, error) {
	if strings.HasPrefix(input, "#") {
		// is a comment
		return false, nil, nil
	}
	// split on white space first
	firstGroup := strings.Split(input, " ")
	if len(firstGroup) > 1 {
		// it's a function or a comment
		daFunq, err := parseFunction(input, number)
		return false, daFunq, err
	}
	// it's a interface attribute
	name, err := parseAttribute(firstGroup[0], number)
	return name != "", name, err
}

// todo: maybe change this to something cleaner using regex in the future?
func parseAttribute(input string, number int) (string, error) {
	// ensure that it's using proper syntax
	secondGroup := strings.Split(input, ":")
	if len(secondGroup) == 1 {
		return "", fmt.Errorf("parser error: Expected \"%v\" at line %v", ":", number)
	}
	// ensure that it's using the "name" attribute, can add to this later but for now... technical debt!
	finalGroup := strings.Split(secondGroup[1], "=")
	if finalGroup[0] != "name" {
		return "", fmt.Errorf("parser error: No such token \"%v\" available, try \"name\" instead", finalGroup[0])
	}
	if len(finalGroup) != 2 {
		return "", fmt.Errorf("parser error: Invalid formatting, \"name\" should be in the following format: name=YourNameHere")
	}
	return finalGroup[1], nil

}

func parseFunction(input string, number int) (definitions.QFunc, error) {
	var inputs []definitions.QType
	var outputs []definitions.QType
	var name string

	left, right, err := validateAndSplitFunc(input)
	if err != nil {
		return definitions.QFunc{}, err
	}

	name, left, err = getNameFromFunc(left)
	if err != nil {
		return definitions.QFunc{}, err
	}

	inputs, err = gatherTypes(left)
	if err != nil {
		return definitions.QFunc{}, err
	}

	outputs, err = gatherTypes(right)
	if err != nil {
		return definitions.QFunc{}, err
	}

	return definitions.QFunc{FuncName: name, Inputs: inputs, Outputs: outputs}, nil
}

func getNameFromFunc(input string) (string, string, error) {
	var name string
	var nameFound bool
	var nameIndex int

	types := strings.Split(input, " ")
	for i, typ := range types {
		typeComponents := strings.Split(typ, ":")
		if typeComponents[1] == "fn" {
			if nameFound == true {
				return "", "", fmt.Errorf("Numerous fn declarations in one function signature")
			}
			nameFound = true
			name = typeComponents[0]
			nameIndex = i
		}
	}

	if name == "" {
		return "", "", fmt.Errorf("parser error: No function name defined in the function signature")
	}
	// return the name and cut out the name
	return name, strings.Join(append(types[:nameIndex], types[nameIndex+1:]...), " "), nil
}

func gatherTypes(input string) ([]definitions.QType, error) {
	var maTypez []definitions.QType
	for _, typ := range strings.Split(input, " ") {
		typeComponents := strings.Split(typ, ":")
		if len(typeComponents) > 2 {
			return nil, fmt.Errorf("parser error: Invalid formatting of output component \"%v\": needs to be formatted as name:type", typ)
		} else if isValidType(typeComponents[1]) {
			maTypez = append(maTypez, definitions.QType{TypeName: typeComponents[0], Type: typeComponents[1]})
		} else {
			return nil, fmt.Errorf("parser error: Invalid type requested, valid types include: uint8-64, int8-64, fn and uniaddress: recieved %v", typeComponents[1])
		}
	}
	return maTypez, nil
}

func validateAndSplitFunc(input string) (string, string, error) {
	functionGroup := strings.Split(input, "->")

	if len(functionGroup) > 2 {
		return "", "", fmt.Errorf("Unexpected multiple \"->\"s in function signature")
	}
	leftString := strings.TrimLeft(strings.TrimRight(functionGroup[0], " "), " ")
	rightString := strings.TrimRight(strings.TrimLeft(functionGroup[1], " "), " ")
	return leftString, rightString, nil
}

func isValidType(typ string) bool {
	switch typ {
	case "uint64", "uint32", "uint16", "uint8", "int64", "int32", "int16", "int8", "uniaddress":
		return true
	default:
		return false
	}
}
