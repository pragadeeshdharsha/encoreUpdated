package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type chainCode struct {
}

type transactionInfo struct {
	TxnType string    //args[1]
	TxnDate time.Time //args[2]
	LoanID  string    //args[3]
	InsID   string    //args[4]
	Amt     int64     //args[5]
	FromID  string    //args[6]
	ToID    string    //args[7]
	By      string    //args[8]
	PprID   string    //args[9]
}

func (c *chainCode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (c *chainCode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()

	if function == "newTxnInfo" {
		return newTxnInfo(stub, args)
	} else if function == "getTxnInfo" {
		return getTxnInfo(stub, args)
	}
	return shim.Success(nil)
}

func newTxnInfo(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 10 {
		xLenStr := strconv.Itoa(len(args))
		return shim.Error("Invalid number of arguments in newTxnInfo(transactions) (required:10) given: " + xLenStr)
	}

	tTypeValues := map[string]bool{
		"disbursement": true,
		"repayment":    true,
		"collection":   true,
		"refund":       true,
	}

	//Converting into lower case for comparison
	tTypeLower := strings.ToLower(args[1])
	if !tTypeValues[tTypeLower] {
		return shim.Error("Invalid transaction type " + args[1])
	}

	//TxnDate -> tDate
	tDate, err := time.Parse("02/01/2006", args[2])
	if err != nil {
		return shim.Error(err.Error())
	}

	amt, err := strconv.ParseInt(args[5], 10, 64)
	if err != nil {
		return shim.Error(err.Error())
	}

	//TODO: put it at last for redability

	switch tTypeLower {

	case "disbursement":
		argsStr := strings.Join(args, ",")
		chaincodeArgs := toChaincodeArgs("newDisbInfo", argsStr)
		fmt.Println("calling the disbursement chaincode")
		response := stub.InvokeChaincode("disbursementcc", chaincodeArgs, "myc")
		if response.Status != shim.OK {
			return shim.Error(response.Message)
		}
		transaction := transactionInfo{tTypeLower, tDate, args[3], args[4], amt, args[6], args[7], args[8], args[9]}
		fmt.Println(transaction)

		txnBytes, err := json.Marshal(transaction)
		err = stub.PutState(args[0], txnBytes)
		if err != nil {
			return shim.Error("Cannot write into ledger the transactino details")
		} else {
			fmt.Println("Successfully inserted disbursement transaction into the ledger")
		}
		//chaincodeArgs = toChaincodeArgs("updateLoanBal",)

	case "repayment":
		argsStr := strings.Join(args, ",")
		chaincodeArgs := toChaincodeArgs("newRepayInfo", argsStr)
		fmt.Println("calling the repayment chaincode")
		response := stub.InvokeChaincode("repaycc", chaincodeArgs, "myc")
		if response.Status != shim.OK {
			return shim.Error(response.Message)
		}
		transaction := transactionInfo{tTypeLower, tDate, args[3], args[4], amt, args[6], args[7], args[8], args[9]}
		fmt.Println(transaction)
		txnBytes, err := json.Marshal(transaction)
		err = stub.PutState(args[0], txnBytes)
		if err != nil {
			return shim.Error("Cannot write into ledger the transaction details")
		} else {
			fmt.Println("Successfully inserted repayment transaction into the ledger")
		}

	default:
		fmt.Println("incorrect txnType")
		return shim.Error("incorrect txnType from txncc")
	}

	return shim.Success(nil)
}

func getTxnInfo(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		xLenStr := strconv.Itoa(len(args))
		return shim.Error("Invalid number of arguments in getTxnInfo (required:1) given: " + xLenStr)
	}

	txnBytes, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error(err.Error())
	} else if txnBytes == nil {
		return shim.Error("No data exists on this txnID: " + args[0])
	}

	transaction := transactionInfo{}
	err = json.Unmarshal(txnBytes, &transaction)
	if err != nil {
		return shim.Error("error while unmarshaling:" + err.Error())
	}

	tString := fmt.Sprintf("%+v", transaction)
	return shim.Success([]byte(tString))

}

func toChaincodeArgs(args ...string) [][]byte {
	bargs := make([][]byte, len(args))
	for i, arg := range args {
		bargs[i] = []byte(arg)
	}
	return bargs
}

func main() {
	err := shim.Start(new(chainCode))
	if err != nil {
		fmt.Printf("Error starting Transaction chaincode: %s\n", err)
	}
}
