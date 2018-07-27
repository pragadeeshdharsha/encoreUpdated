package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type chainCode struct {
}

type loanInfo struct {
	InstNum            string //Instrument Number
	ExposureBusinessID string
	ProgramID          string
	SanctionAmt        int64
	SanctionDate       time.Time //with time
	SanctionAuthority  string
	ROI                float64
	DueDate            time.Time
	ValueDate          time.Time //with time
	LoanStatus         string
	LoanBalance        int64
}

func (c *chainCode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (c *chainCode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()

	if function == "newLoanInfo" {
		return newLoanInfo(stub, args)
	} else if function == "getLoanInfo" {
		return getLoanInfo(stub, args)
	} else if function == "updateLoanInfo" {
		return updateLoanInfo(stub, args)
	}
	return shim.Error("No function named " + function + " in Loan")
}

func newLoanInfo(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 12 {
		return shim.Error("Invalid number of arguments")
	}

	//Checking if the instrumentID exist or not
	/*chk, err := stub.GetState(args[1])
	if chk == nil {
		return shim.Error("There is no instrument ID:" + args[1])
	}*/

	//SanctionAmt -> sAmt
	sAmt, err := strconv.ParseInt(args[4], 10, 64)
	if err != nil {
		return shim.Error(err.Error())
	}

	//Converting the incoming date from Dd/mm/yy:hh:mm:ss to Dd/mm/yyThh:mm:ss for parsing
	sDateStr := args[5][:10]
	sTime := args[5][11:]
	sStr := sDateStr + "T" + sTime

	//SanctionDate ->sDate
	sDate, err := time.Parse("02/01/2006T15:04:05", sStr)
	if err != nil {
		return shim.Error(err.Error())
	}

	roi, err := strconv.ParseFloat(args[7], 32)
	if err != nil {
		return shim.Error(err.Error())
	}

	//Parsing into date for storage but hh:mm:ss will also be stored as
	//00:00:00 .000Z with the date
	//DueDate -> dDate
	dDate, err := time.Parse("02/01/2006", args[8])
	if err != nil {
		return shim.Error(err.Error())
	}

	//Converting the incoming date from Dd/mm/yy:hh:mm:ss to Dd/mm/yyThh:mm:ss for parsing
	vDateStr := args[5][:10]
	vTime := args[5][11:]
	vStr := vDateStr + "T" + vTime

	//ValueDate ->vDate
	vDate, err := time.Parse("02/01/2006T15:04:05", vStr)
	if err != nil {
		return shim.Error(err.Error())
	}

	/*	loanStatusValues := map[string]bool{
			"open":              true,
			"sanctioned":        true,
			"part disbursed":    true,
			"disbursed":         true,
			"part collected":    true,
			"collected/settled": true,
			"overdue":           true,
		}

		loanStatusValuesLower := strings.ToLower(args[10])
		if !loanStatusValues[loanStatusValuesLower] {
			return shim.Error("Invalid Instrument Status " + args[10])
		}*/

	loanBalanceString, err := strconv.ParseInt(args[11], 10, 64)
	if err != nil {
		return shim.Error("Error in parsing int in newLoanInfo:" + err.Error())
	}

	ifExists, err := stub.GetState(args[0])
	if ifExists != nil {
		fmt.Println(ifExists)
		return shim.Error("LoanId " + args[0] + " exits. Cannot create new ID")
	}

	loan := loanInfo{args[1], args[2], args[3], sAmt, sDate, args[6], roi, dDate, vDate, "open", loanBalanceString}
	loanBytes, err := json.Marshal(loan)
	if err != nil {
		return shim.Error(err.Error())
	}
	stub.PutState(args[0], loanBytes)
	return shim.Success(nil)
}

func getLoanInfo(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		xLenStr := strconv.Itoa(len(args))
		return shim.Error("Invalid number of arguments in getLoanInfo (required:1) given:" + xLenStr)

	}

	loanBytes, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error(err.Error())
	} else if loanBytes == nil {
		return shim.Error("No data exists on this loanID: " + args[0])
	}

	loan := loanInfo{}
	err = json.Unmarshal(loanBytes, &loan)
	if err != nil {
		return shim.Error(err.Error())
	}
	loanString := fmt.Sprintf("%+v", loan)
	fmt.Printf("Loan Info:%s\n ", loanString)
	loanBalString := strconv.FormatInt(loan.LoanBalance, 10)
	loanStatus := loan.LoanStatus
	loanSanctionString := strconv.FormatInt(loan.SanctionAmt, 10)
	var xString bytes.Buffer

	xString.WriteString(loanBalString)
	xString.WriteString(",")
	xString.WriteString(loanStatus)
	xString.WriteString(",")
	xString.WriteString(loanSanctionString)
	fmt.Println("args:", xString.String())
	fmt.Println("Type:", reflect.TypeOf(xString.String()))
	fmt.Println("Returning the values")
	//xString = sanctionString + "," + loanStatus
	return shim.Success([]byte(xString.String()))
}

func updateLoanInfo(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	/*
		Updating the variables for loan structure
	*/
	loanBytes, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error(err.Error())
	} else if loanBytes == nil {
		return shim.Error("No data exists on this loanID: " + args[0])
	}

	loan := loanInfo{}
	err = json.Unmarshal(loanBytes, &loan)
	if err != nil {
		return shim.Error("error in unmarshiling loan: in updateLoanInfo" + err.Error())
	}

	fmt.Printf("args[1]:%s\n", args[1])
	// To change the LoanStatus from "open" to "sanction"
	if len(args) == 2 && args[1] == "sanctioned" {

		loan.LoanStatus = strings.ToLower(args[1])
		sanctionString := strconv.FormatInt(loan.SanctionAmt, 10)
		argsToLoanBal := []string{"1loanbal", args[0], "0", "02/01/2006", "0", sanctionString, "0", "0", sanctionString, "sanctioned"}
		argsString := strings.Join(argsToLoanBal, ",")
		chaincodeArgs := toChaincodeArgs("putLoanBalInfo", argsString)
		response := stub.InvokeChaincode("loanbalcc", chaincodeArgs, "myc")
		if response.Status != shim.OK {
			return shim.Error("Unable to create a loanBal entry from loan:" + response.Message)
		}
		loanBytes, _ := json.Marshal(loan)
		err = stub.PutState(args[0], loanBytes)
		if err != nil {
			return shim.Error("Error in loan updation " + err.Error())
		}
		return shim.Success([]byte("sanction updated succesfully"))

	} else if len(args) == 3 { // used when called from loanBal
		//xLenStr := strconv.Itoa(len(args))
		//return shim.Error("Invalid number of arguments in updateLoanInfo (required:3) given:" + xLenStr)

		// This "if" condition will be executed when loanBalance chaincode call for updation

		loan.LoanStatus = args[1]
		loan.LoanBalance, err = strconv.ParseInt(args[2], 10, 64)
		if err != nil {
			return shim.Error("Unable to parse int in updateLoanInfo:" + err.Error())
		}

		loanBytes, _ = json.Marshal(loan)
		err = stub.PutState(args[0], loanBytes)
		if err != nil {
			return shim.Error("Error in loan updation " + err.Error())
		}
		return shim.Success([]byte("Successfully updated loan with data from loanbal"))
	}
	return shim.Error("Invalid info for update loan")
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
		fmt.Printf("Error starting Loan chaincode: %s\n", err)
	}
}
