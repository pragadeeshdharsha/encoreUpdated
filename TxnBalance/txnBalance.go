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

type txnBalanceInfo struct {
	TxnID      string
	TxnDate    time.Time
	LoanID     string
	InsID      string
	WalletID   string
	OpeningBal int64
	TxnType    string
	Amt        int64
	CAmt       int64
	DAmt       int64
	TxnBal     int64
	By         string
}

func (c *chainCode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (c *chainCode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	if function == "putTxnInfo" { //Inserting a New Business information
		return putTxnInfo(stub, args)
	} else if function == "getTxnBalInfo" { // To view a Transaction Balance
		return getTxnBalInfo(stub, args)
	}
	return shim.Error("No function named " + function + " in TxnBalance")
}

func putTxnInfo(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) == 1 {
		args = strings.Split(args[0], ",")
	}
	if len(args) != 13 {
		xLenStr := strconv.Itoa(len(args))
		return shim.Error("Invalid number of arguments in putTxnInfo (required:13) given:" + xLenStr)
	}
	fmt.Println("Printing args")
	fmt.Println(args)
	//TxnDate ->txnDate
	txnDate, err := time.Parse("02/01/2006", args[2])
	if err != nil {
		return shim.Error("err in txndate " + err.Error())
	}

	openBal, err := strconv.ParseInt(args[6], 10, 64)
	if err != nil {
		return shim.Error("err in openbal " + err.Error())
	}

	txnTypeValues := map[string]bool{
		"loan Sanction":       true,
		"disbursement":        true,
		"charges":             true,
		"repayment":           true,
		"collection":          true,
		"margin refund":       true,
		"interest refund":     true,
		"tds":                 true,
		"penal charges":       true,
		"cersai carges":       true,
		"factor regn charges": true,
	}

	txnTypeLower := strings.ToLower(args[7])
	if !txnTypeValues[txnTypeLower] {
		return shim.Error("Invalid Transaction type (TxnBalance):" + txnTypeLower)
	}

	amt, err := strconv.ParseInt(args[8], 10, 64)
	if err != nil {
		return shim.Error("err in amt(TxnBlance)" + err.Error())
	}

	cAmt, err := strconv.ParseInt(string(args[9]), 10, 64)
	if err != nil {

		return shim.Error("err in camt(TxnBlance)" + err.Error())
	}

	dAmt, err := strconv.ParseInt(args[10], 10, 64)
	if err != nil {
		return shim.Error("err in damt(TxnBlance) " + err.Error())
	}

	txnBal, err := strconv.ParseInt(args[11], 10, 64)
	if err != nil {
		return shim.Error("err in txnbal (TxnBlance)" + err.Error())
	}

	ifExists, err := stub.GetState(args[0])
	if ifExists != nil {
		return shim.Error("TxnBalanceId " + args[0] + " exits. Cannot create new ID")
	}

	txnBalance := txnBalanceInfo{args[1], txnDate, args[3], args[4], args[5], openBal, txnTypeLower, amt, cAmt, dAmt, txnBal, args[12]}
	txnBalanceBytes, err := json.Marshal(txnBalance)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(args[0], txnBalanceBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	//fmt.Println("Transaction :", txnBalance)
	fmt.Printf("Succefully wrote txnID %s into the ledger\n", args[0])

	return shim.Success([]byte("Successful"))

}

func getTxnBalInfo(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		xLenStr := strconv.Itoa(len(args))
		return shim.Error("Invalid number of argumentrs in getTxnBalInfo (required:1) given:" + xLenStr)
	}

	//fmt.Println("Inside TxnBalance function")

	txnBalance := txnBalanceInfo{}
	txnBalanceBytes, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error("Failed to get the Transaction information: " + err.Error())
	} else if txnBalanceBytes == nil {
		return shim.Error("No information is avalilable on this TxnBalID " + args[0])
	}
	//fmt.Println("Got TxnBalance")

	err = json.Unmarshal(txnBalanceBytes, &txnBalance)
	if err != nil {
		return shim.Error("Unable to parse TxnBalance into the structure " + err.Error())
	}
	//fmt.Println("Unmarshled TxnBalance function")
	jsonString := fmt.Sprintf("%+v", txnBalance)
	fmt.Printf("Transaction info %s : %s\n", args[0], jsonString)
	return shim.Success(nil)
}

func main() {
	err := shim.Start(new(chainCode))
	if err != nil {
		fmt.Printf("Error starting TxnBalance chaincode: %s\n", err)
	}

}
