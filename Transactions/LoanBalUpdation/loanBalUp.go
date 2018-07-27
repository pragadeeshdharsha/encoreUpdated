package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type chainCode struct{}

func (c *chainCode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (c *chainCode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	if function == "updateLoanBal" {
		return updateLoanBal(stub, args)
	}
	return shim.Error("No function named " + function + " in loanBalUp")
}

func updateLoanBal(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	/*
			// From Disbursement
		*LoanID  -> args[0]
		*TxnID   -> args[1]
		*TxnDate -> args[2]
		*TxnType -> args[3]
		*DAmt    -> args[4]


		*OpenBal -> LoanBalance from Loan structure
		*CAmt

		*LoanBal -> OpenBal-DAmt+Camt
		*LoanStatus -> depends
	*/

	if len(args) == 1 {
		args = strings.Split(args[0], ",")
	}
	if len(args) != 5 {
		return shim.Error("Required 5 arguments in updateLoanBal from Disbursement")
	}

	//argsList := []string{"1", args[0], args[2], args[3], args[4], walletID, openBalString, args[1], args[5], cAmtString, dAmtString, txnBalString, args[8]}
	//argsListStr := strings.Join(argsList, ",")
	chaincodeArgs := util.ToChaincodeArgs("getLoanInfo")
	fmt.Println("calling the other chaincode")
	response := stub.InvokeChaincode("loancc", chaincodeArgs, "myc")
	if response.Status != shim.OK {
		return shim.Error(response.Message)
	}
	// spliting the arguments got from loan as response
	loanArgs := strings.Split(string(response.Payload), ",")
	openBal, err := strconv.ParseInt(loanArgs[0], 10, 64)
	if err != nil {
		return shim.Error("Error in parsing the openbalance in LoanBalance: " + err.Error())
	}
	CAmt := int64(0)
	DAmt, err := strconv.ParseInt(args[4], 10, 64)
	if err != nil {
		return shim.Error("Error in parsing the DAmt in LoanBalance: " + err.Error())
	}

	var status string
	status = loanArgs[1] // status of the current loan
	loanBal := openBal - DAmt + CAmt
	loanBalString := strconv.FormatInt(loanBal, 64)
	if status == "open" || status == "partly disbursed" {

		if openBal-loanBal == 0 {
			status = "disbursed"
		} else {
			status = "partly disbursed"
		}
	}
	fmt.Printf("Status:%s\n", status)
	chaincodeArgs = util.ToChaincodeArgs("updateLoanInfo", args[0], status, loanBalString)
	fmt.Println("calling the other chaincode")
	response = stub.InvokeChaincode("loancc", chaincodeArgs, "myc")
	if response.Status != shim.OK {
		return shim.Error(response.Message)
	}
	return shim.Success(nil)
}

func main() {

}
