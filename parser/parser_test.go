package parser

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	def "github.com/qtumproject/simple-abi/definitions"
)

// make a function "Parse", which will parse a file's output after opening it

func TestParseName(t *testing.T) {
	const exampleFileInputLine = `:name=AirDropToken`

	isName, name, err := parseLine(exampleFileInputLine, 0)
	if err != nil {
		t.Fatalf("Unexpected error occurred: %v", err)
	}

	if isName != nameComponent {
		t.Errorf("Expected to receive a name but did not")
	}

	if name.(string) != "AirDropToken" {
		t.Errorf("Expected name %v but got %v", "AirDropToken", name)
	}
}

/*func TestParseImplementsInterface(t *testing.T) {
	const exampleFileInputLine = `:implements=ERC20, ERC721`

	_, interfac, err := parseLine(exampleFileInputLine, 0)
	if err != nil {
		t.Fatalf("Unexpected error occurred: %v", err)
	}

	if interfac.(string) != "ERC20, ERC721" {
		t.Errorf("Expected interface %v but got %v", "ERC20, ERC721", interfac)
	}
}*/

func TestParseComment(t *testing.T) {
	const exampleInput = `# this is a comment`
	_, output, err := parseLine(exampleInput, 0)
	if err != nil {
		t.Errorf("Expected to recieve nil error got %v", err)
	}
	if output != nil {
		t.Errorf("Expected to recieve nil value got %v", output)
	}
}

func TestParseNameFailures(t *testing.T) {
	var parserNameFailures = []struct {
		input  string
		output string
	}{
		{"name=AirDropToken", "parser error: Expected \":\" at line 0"},
		{":version=0.2.0", "parser error: No such token \"version\" available, try \"name\" or \"implements\" instead"}, // todo: take this one out eventually
		{":name:AirDropToken", "parser error: Invalid formatting, \"name\" or \"implements\" should be in the following format: name=YourNameHere, implements=YourImplementationHere"},
	}

	for _, test := range parserNameFailures {
		_, _, err := parseLine(test.input, 0)
		if err == nil || err.Error() != test.output {
			t.Errorf("Expected error %v, got %v", test.output, err)
		}
	}
}

func TestParseFunction(t *testing.T) {
	var functionInputs = []struct {
		input  string
		output interface{}
	}{
		{
			"somevar:uint8 othervar:int64 myFunction:fn -> somereturn:uint8 otherreturn:int32",
			def.QFunc{FuncName: "myFunction", Inputs: []def.QType{def.QType{TypeName: "somevar", Type: "uint8"}, def.QType{TypeName: "othervar", Type: "int64"}}, Outputs: []def.QType{def.QType{TypeName: "somereturn", Type: "uint8"}, def.QType{TypeName: "otherreturn", Type: "int32"}}},
		},
		{
			"somevar:uint32 otherFunction:fn -> somereturn:uint32",
			def.QFunc{FuncName: "otherFunction", Inputs: []def.QType{def.QType{TypeName: "somevar", Type: "uint32"}}, Outputs: []def.QType{def.QType{TypeName: "somereturn", Type: "uint32"}}},
		},
		{
			"addressvar:uniaddress someFunction:fn -> addressreturn:uniaddress",
			def.QFunc{FuncName: "someFunction", Inputs: []def.QType{def.QType{TypeName: "addressvar", Type: "uniaddress"}}, Outputs: []def.QType{def.QType{TypeName: "addressreturn", Type: "uniaddress"}}},
		},
		{
			"addressarray:uniaddress[] intarray:int64[] arrFunction:fn -> uintarray:uint32[]",
			def.QFunc{FuncName: "arrFunction", Inputs: []def.QType{def.QType{TypeName: "addressarray", Type: "uniaddress[]"}, def.QType{TypeName: "intarray", Type: "int64[]"}}, Outputs: []def.QType{def.QType{TypeName: "uintarray", Type: "uint32[]"}}},
		},
		{
			"void voidFunction:fn -> a:uint32",
			def.QFunc{FuncName: "voidFunction", Inputs: nil, Outputs: []def.QType{def.QType{TypeName: "a", Type: "uint32"}}},
		},
		{
			"a:uint8 payableVoidFunc:fn:payable -> void",
			def.QFunc{FuncName: "payableVoidFunc", Inputs: []def.QType{def.QType{TypeName: "a", Type: "uint8"}}, Outputs: nil, Payable: true},
		},
	}

	for _, test := range functionInputs {
		_, daFunq, err := parseLine(test.input, 0)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if !cmp.Equal(daFunq, test.output) {
			t.Errorf("Expected output to equal %v: got output %v", test.output, daFunq)
		}
	}
}

func TestParseFunctionErrors(t *testing.T) {
	var functionInputs = []struct {
		input string
		err   string
	}{
		{
			"somevar:uint8 othervar:int64 -> somereturn:uint8 otherreturn:int32",
			"parser error: No function name defined in the function signature",
		},
		{
			":somevar:uint32:someothervar otherFunction:fn -> somereturn:uint32",
			"parser error: Invalid formatting of output component \":somevar:uint32:someothervar\": needs to be formatted as name:type",
		},
		{
			"somevar:uint18 otherFunction:fn -> somereturn:uint32",
			"parser error: Invalid type requested, valid types include: uint8-64, int8-64, fn and uniaddress: recieved uint18",
		},
		{
			"somevar:uint32 -> otherFunction:fn -> somereturn:uin32",
			"parser error: unexpected multiple \"->\"s in function signature",
		},
	}

	for _, test := range functionInputs {
		_, _, err := parseLine(test.input, 0)
		if err == nil {
			t.Errorf("Expected error, got none")
		}

		if test.err != err.Error() {
			t.Errorf("Expected error %v: got %v", test.err, err)
		}
	}
}
