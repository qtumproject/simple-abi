package parser

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

// make a function "Parse", which will parse a file's output after opening it

func TestParseName(t *testing.T) {
	const exampleFileInputLine = `:name=AirDropToken`

	isName, name, err := parseLine(exampleFileInputLine, 0)
	if err != nil {
		t.Fatalf("Unexpected error occurred: %v", err)
	}

	if isName == false {
		t.Errorf("Expected to receive a name but did not")
	}

	if name.(string) != "AirDropToken" {
		t.Errorf("Expected name %v but got %v", "AirDropToken", name)
	}
}

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
		{":version=0.2.0", "parser error: No such token \"version\" available, try \"name\" instead"}, // todo: take this one out eventually
		{":name:AirDropToken", "parser error: Invalid formatting, \"name\" should be in the following format: name=YourNameHere"},
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
			qFunc{Name: "myFunction", Inputs: []qType{qType{"somevar", "uint8"}, qType{"othervar", "int64"}}, Outputs: []qType{qType{"somereturn", "uint8"}, qType{"otherreturn", "int32"}}},
		},
		{
			"somevar:uint32 otherFunction:fn -> somereturn:uint32",
			qFunc{Name: "otherFunction", Inputs: []qType{qType{"somevar", "uint32"}}, Outputs: []qType{qType{"somereturn", "uint32"}}},
		},
		{
			"addressvar:uniaddress someFunction:fn -> addressreturn:uniaddress",
			qFunc{Name: "someFunction", Inputs: []qType{qType{"addressvar", "uniaddress"}}, Outputs: []qType{qType{"addressreturn", "uniaddress"}}},
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
			"Unexpected multiple \"->\"s in function signature",
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
