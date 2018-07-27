package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type rich1 struct{}

func (r *rich1) Init(stub shim.ChaincodeStubInterface) pb.Response {

	return shim.Success(nil)
}

func (r *rich1) Invoke(stub shim.ChaincodeStubInterface) pb.Response {

	return shim.Error("No function exists")
}

/*func main() {
	err := shim.Start(new(rich1))
	if err != nil {
		fmt.Printf("Error starting the rich1 chaincode:%s\n", err)
	}
}*/
