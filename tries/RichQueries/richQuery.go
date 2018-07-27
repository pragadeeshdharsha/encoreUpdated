package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type rich struct{}

type sample struct {
	Name   string
	ID     string
	Colour string
}

func (r *rich) Init(stub shim.ChaincodeStubInterface) pb.Response {

	return shim.Success(nil)
}

func (r *rich) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()

	if function == "changeColour" {
		return r.changeColour(stub, args)
	} else if function == "new1" {
		return r.new1(stub, args)
	}
	return shim.Error("No function exists")
}

func (r *rich) new1(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	s1 := &sample{args[0], args[1], args[2]}
	s1Bytes, _ := json.Marshal(s1)
	stub.PutState(args[1], s1Bytes)
	fmt.Println(s1)

	fmt.Println("Initializing the values")
	indexName := "colour~id"
	colourIDkey, err := stub.CreateCompositeKey(indexName, []string{s1.Colour, s1.ID})
	if err != nil {
		fmt.Println("Composite key cannot be created")
		return shim.Error(err.Error())
	}
	value := []byte{0x00}
	stub.PutState(colourIDkey, value)
	return shim.Success([]byte("Successfully written into the ledger"))
}

func (r *rich) changeColour(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	fmt.Println("Entering change colour.")
	colour := args[0]
	//changeColour := args[1]

	colourIditerator, err := stub.GetStateByPartialCompositeKey("colour~id", []string{colour})
	if err != nil {
		return shim.Error("Iterator error:" + err.Error())
	}
	defer colourIditerator.Close()
	fmt.Println("Before for looop")
	fmt.Println("Before for looop")

	for i := 0; colourIditerator.HasNext(); i++ {
		x, err := colourIditerator.Next()
		if err != nil {
			return shim.Error("Unable to iterate:" + err.Error())
		}
		fmt.Println("xyzzzzzzzzzz")
		y, z, err := stub.SplitCompositeKey(x.Key)

		if err != nil {
			return shim.Error("error spliting the composite key:" + err.Error())
		}
		fmt.Println("i:", i)
		fmt.Println("x:", x)
		fmt.Println("y:", y)
		fmt.Println("z:", z)

	}
	return shim.Success([]byte("Successfully for loop executed"))
}

func main() {
	err := shim.Start(new(rich))
	if err != nil {
		fmt.Printf("Error starting the Rich chaincode:%s\n", err)
	}
}
