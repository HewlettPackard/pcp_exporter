package sources

import (
	"testing"
)

func TestGetRequest(t *testing.T) {
	testString := "http://httpbin.org/get"
	response := getRequest(testString)
	if response == nil {
		t.Fatal("retrieved an unexpected result")
	}
}

func TestUnmarshal(t *testing.T) {
	testBody := getRequest(http:/httpbin.org/get)
	testMap := &testMapHeader{}
	response := unmarshal([]byte(testBody),testMapHeader)	
	if response == nil {
		t.Fatal("retrieved an unexpected result")
	}	
		
}

func TestLabelTypes(t *testing.T) {
	testStringOne := ""
	testStringTwo := "" 	
	testStringThree := "kal"
	response := typeLabel(testStringOne, testStringTwo, testStringThree)
	if response == "" {
		t.Fatal("retrieved an unexpected result")
	}
}

//func TestNewPcpSource(t *testing.T) {
//	
//	
//	
//}

