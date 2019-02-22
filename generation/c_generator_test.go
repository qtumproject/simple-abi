package generation

import (
	"bytes"
	"testing"

	def "github.com/VoR0220/SimpleABI/definitions"
)

const decodeTest1 = `
#include <stdlib.h>
#include <qtum.h>

//Function IDs
#define ID_MyContract_myFunction 0x996c38c3
#define ID_MyContract_otherFunction 0x01c66199

//prototypes
void MyContract_myFunction(uint8_t somevar, int64_t othervar, uint8_t* somereturn, int32_t* otherreturn);
void MyContract_otherFunction(uint8_t somevar, uint32_t* somereturn);

//dispatch code
void dispatch(){
	uint32_t fn;
	if(qtumPop(&fn, sizeof(fn) != sizeof(fn)){
		//fallback function/error
	}
	switch(fn){
	    case ID_MyContract_myFunction:
	    {
		    uint8_t somevar = qtumPop8();
		    int64_t othervar = qtumPop64();
		    uint8_t somereturn = 0;
		    int32_t otherreturn = 0;
		    MyContract_myFunction(uint8_t somevar, int64_t othervar, &somereturn, &otherreturn)
		    qtumPop8()(somereturn);
		    qtumPop32()(otherreturn);
		    break;
	    }
	    case ID_MyContract_otherFunction:
	    {
		    uint8_t somevar = qtumPop8();
		    uint32_t somereturn = 0;
		    MyContract_otherFunction(uint8_t somevar, &somereturn)
		    qtumPop32()(somereturn);
		    break;
	    }
	    default:
		    //fallback function / error
	}
}`

const decodeTest2 = `
#include <stdlib.h>
#include <qtum.h>

//Function IDs
#define ID_MyContract_myFunction 0x6985f0c9

//prototypes
void MyContract_myFunction(UniversalAddressABI* addressvar, UniversalAddressABI** addressreturn);

//dispatch code
void dispatch(){
	uint32_t fn;
	if(qtumPop(&fn, sizeof(fn) != sizeof(fn)){
		//fallback function/error
	}
	switch(fn){
		case ID_MyContract_myFunction:
		{
			UniversalAddressABI* addressvar = malloc(sizeof(UniversalAddressABI));
			qtumPopExact(addressvar, sizeof(UniversalAddressABI));
			UniversalAddressABI* addressreturn = NULL;
			MyContract_myFunction(UniversalAddressABI* addressvar, &addressreturn)
			qtumPopExact(addressreturn, sizeof(UniversalAddressABI));
			break;
		}
		default:
			//fallback function / error
	}
}`

const encodeTest1 = `
#include <stdlib.h>
#include <qtum.h>

//Function IDs
#define ID_MyContract_myFunction 0x996c38c3
#define ID_MyContract_otherFunction 0x01c66199

QtumCallResult  MyContract_myFunction(UniversalAddress __address, QtumCallOptions* __options, uint8_t somevar, int64_t othervar, uint8_t* somereturn, int32_t* otherreturn){
	qtumPush8(somevar);
	qtumPush64(othervar);
	qtumPush32(ID_MyContract_myFunction);
	QtumCallResult r = qtumCall(__address, __options);
	if(r.error == QTUM_CALL_SUCCESS){
		*somereturn = qtumPop8();
		*otherreturn = qtumPop32();
	}
	return r;
}

QtumCallResult  MyContract_otherFunction(UniversalAddress __address, QtumCallOptions* __options, uint8_t somevar, uint32_t* somereturn){
	qtumPush8(somevar);
	qtumPush32(ID_MyContract_otherFunction);
	QtumCallResult r = qtumCall(__address, __options);
	if(r.error == QTUM_CALL_SUCCESS){
		*somereturn = qtumPop32();
	}
	return r;
}

`

const encodeTest2 = `
#include <stdlib.h>
#include <qtum.h>

//Function IDs
#define ID_MyContract_myFunction 0x6985f0c9

QtumCallResult  MyContract_myFunction(UniversalAddress __address, QtumCallOptions* __options, UniversalAddressABI* addressvar, UniversalAddressABI** addressreturn){
	qtumPush(addressvar);
	qtumPush32(ID_MyContract_myFunction);
	QtumCallResult r = qtumCall(__address, __options);
	if(r.error == QTUM_CALL_SUCCESS){
		if(addressreturn == NULL){
			addressreturn = malloc(sizeof(UniversalAddressABI));
		}
		if(addressreturn == NULL){
			qtumErase();
		}else{
			qtumPop(addressreturn, sizeof(UniversalAddressABI));
		}
	}
	return r;
}

`

func TestDecodeTemplate1(t *testing.T) {
	t.Skip()
	var b bytes.Buffer
	builder := def.QInterfaceBuilder{
		ContractName: "MyContract",
		Functions: []def.QFunc{
			def.QFunc{
				FuncName: "myFunction",
				Inputs: []def.QType{
					def.QType{Type: "uint8", TypeName: "somevar"},
					def.QType{Type: "int64", TypeName: "othervar"},
				},
				Outputs: []def.QType{
					def.QType{Type: "uint8", TypeName: "somereturn"},
					def.QType{Type: "int32", TypeName: "otherreturn"},
				},
			},
			def.QFunc{
				FuncName: "otherFunction",
				Inputs: []def.QType{
					def.QType{Type: "uint8", TypeName: "somevar"},
				},
				Outputs: []def.QType{
					def.QType{Type: "uint32", TypeName: "somereturn"},
				},
			},
		},
	}
	err := GenerateTemplate(builder, "decodeTest1", &b, false)
	if err != nil {
		t.Errorf("Unexpected error in template generation of decodeTest1: %v", err)
	}
	got := b.String()
	want := decodeTest1
	if got != want {
		t.Errorf("DecodeTest1 got %v, want %v", got, want)
	}
}

func TestDecodeTemplate2(t *testing.T) {
	t.Skip()
	var b bytes.Buffer
	builder := def.QInterfaceBuilder{
		ContractName: "MyContract",
		Functions: []def.QFunc{
			def.QFunc{
				FuncName: "myFunction",
				Inputs: []def.QType{
					def.QType{Type: "uniaddress", TypeName: "addressvar"},
				},
				Outputs: []def.QType{
					def.QType{Type: "uniaddress", TypeName: "addressreturn"},
				},
			},
		},
	}
	err := GenerateTemplate(builder, "decodeTest2", &b, false)
	if err != nil {
		t.Errorf("Unexpected error in template generation of decodeTest2: %v", err)
	}
	got := b.String()
	want := decodeTest2
	if got != want {
		t.Errorf("DecodeTest2 got %v, want %v", got, want)
	}
}

func TestEncodeTemplate1(t *testing.T) {
	var b bytes.Buffer
	builder := def.QInterfaceBuilder{
		ContractName: "MyContract",
		Functions: []def.QFunc{
			def.QFunc{
				FuncName: "myFunction",
				Inputs: []def.QType{
					def.QType{Type: "uint8", TypeName: "somevar"},
					def.QType{Type: "int64", TypeName: "othervar"},
				},
				Outputs: []def.QType{
					def.QType{Type: "uint8", TypeName: "somereturn"},
					def.QType{Type: "int32", TypeName: "otherreturn"},
				},
			},
			def.QFunc{
				FuncName: "otherFunction",
				Inputs: []def.QType{
					def.QType{Type: "uint8", TypeName: "somevar"},
				},
				Outputs: []def.QType{
					def.QType{Type: "uint32", TypeName: "somereturn"},
				},
			},
		},
	}
	err := GenerateTemplate(builder, "encodeTest1", &b, true)
	if err != nil {
		t.Errorf("Unexpected error in template generation of encodeTest1: %v", err)
	}
	got := b.String()
	want := encodeTest1
	if got != want {
		t.Errorf("EncodeTest1 got %v, want %v", got, want)
	}
}

func TestEncodeTemplate2(t *testing.T) {
	var b bytes.Buffer
	builder := def.QInterfaceBuilder{
		ContractName: "MyContract",
		Functions: []def.QFunc{
			def.QFunc{
				FuncName: "myFunction",
				Inputs: []def.QType{
					def.QType{Type: "uniaddress", TypeName: "addressvar"},
				},
				Outputs: []def.QType{
					def.QType{Type: "uniaddress", TypeName: "addressreturn"},
				},
			},
		},
	}
	err := GenerateTemplate(builder, "encodeTest2", &b, true)
	if err != nil {
		t.Errorf("Unexpected error in template generation of encodeTest2: %v", err)
	}
	got := b.String()
	want := encodeTest2
	if got != want {
		t.Errorf("EncodeTest2 got %v, want %v", got, want)
	}
}
