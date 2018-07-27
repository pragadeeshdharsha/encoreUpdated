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

type loanBalanceInfo struct {
	LoanID     string
	TxnID      string
	TxnDate    time.Time
	TxnType    string
	OpenBal    int64
	CAmt       int64
	DAmt       int64
	LoanBal    int64
	LoanStatus string
}

func (c *chainCode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (c *chainCode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	if function == "putLoanBalInfo" {
		return putLoanBalInfo(stub, args)
	} else if function == "getLoanBalInfo" {
		return getLoanBalInfo(stub, args)
	} else if function == "updateLoanBal" {
		return updateLoanBal(stub, args)
	}
	return shim.Error("No function named " + function + " in loanBalance")
}

func putLoanBalInfo(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) == 1 {
		args = strings.Split(args[0], ",")
	}
	if len(args) != 10 {
		xLenStr := strconv.Itoa(len(args))
		return shim.Error("Invalid number of arguments in putLoanBalInfo (required:10) given:" + xLenStr)
	}

	//TxnDate -> transDate
	transDate, err := time.Parse("02/01/2006", args[3])
	if err != nil {
		return shim.Error(err.Error())
	}

	/*txnTypeValues := map[string]bool{

		"disbursement":  true,
		"charges":       true,
		"payment":       true,
		"other changes": true,
	}

	txnTypeLower := strings.ToLower(args[4])
	if !txnTypeValues[txnTypeLower] {
		return shim.Error("Invalid Transaction type " + txnTypeLower)
	}*/

	openBal, err := strconv.ParseInt(args[5], 10, 64)
	if err != nil {
		return shim.Error(err.Error())
	}

	cAmt, err := strconv.ParseInt(args[6], 10, 64)
	if err != nil {
		return shim.Error(err.Error())
	}

	dAmt, err := strconv.ParseInt(args[7], 10, 64)
	if err != nil {
		return shim.Error(err.Error())
	}

	loanBal, err := strconv.ParseInt(args[8], 10, 64)
	if err != nil {
		return shim.Error(err.Error())
	}

	loanStatusValues := map[string]bool{
		"open":           true,
		"sanctioned":     true,
		"part disbursed": true,
		"disbursed":      true,
		"part collected": true,
		"collected":      true,
		"overdue":        true,
	}
	loanStatusLower := strings.ToLower(args[9])
	if !loanStatusValues[loanStatusLower] {
		return shim.Error("Invalid Loan Status type " + loanStatusLower)
	}

	loanBalance := loanBalanceInfo{args[1], args[2], transDate, args[4], openBal, cAmt, dAmt, loanBal, loanStatusLower}
	loanBalanceBytes, err := json.Marshal(loanBalance)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(args[0], loanBalanceBytes)

	return shim.Success(nil)

}

func getLoanBalInfo(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		xLenStr := strconv.Itoa(len(args))
		return shim.Error("Required only one argument in getLoanBalInfo, given:" + xLenStr)
	}

	loanBalance := loanBalanceInfo{}
	loanBalanceBytes, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error("Failed to get the business information: " + err.Error())
	} else if loanBalanceBytes == nil {
		return shim.Error("No information is avalilable on this businessID " + args[0])
	}

	err = json.Unmarshal(loanBalanceBytes, &loanBalance)
	if err != nil {
		return shim.Error("Unable to parse into the structure " + err.Error())
	}
	jsonString := fmt.Sprintf("%+v", loanBalance)
	return shim.Success([]byte(jsonString))
}

func updateLoanBal(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	/*
			// From Disbursement
		/*
		*LoanBalId -> args[0]
		*LoanID  -> args[1]
		*TxnID   -> args[2]
		*TxnDate -> args[3]
		*TxnType -> args[4]
		*DAmt    -> args[5]


		*OpenBal -> LoanBalance from Loan structure
		*CAmt

		*LoanBal -> OpenBal-DAmt+Camt
		*LoanStatus -> depends
	*/

	if len(args) == 1 {
		args = strings.Split(args[0], ",")
	}

	loanBalance := loanBalanceInfo{}
	loanBalanceBytes, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error("Failed to get the loan balance information: " + err.Error())
	} else if loanBalanceBytes == nil {
		return shim.Error("No information is avalilable on this loan balance " + args[0])
	}

	err = json.Unmarshal(loanBalanceBytes, &loanBalance)
	if err != nil {
		return shim.Error("Unable to parse loan balance into the structure " + err.Error())
	}
	chaincodeArgs := toChaincodeArgs("getLoanInfo", args[1])
	fmt.Println("calling the other chaincode")
	response := stub.InvokeChaincode("loancc", chaincodeArgs, "myc")
	if response.Status != shim.OK {
		return shim.Error(response.Message)
	}
	//spliting the arguments got from loan as response (loanBalance -> [0] and status -> [1] and SanctionAmt -> args[2])
	loanArgs := strings.Split(string(response.Payload), ",")
	timeType, err := time.Parse("02/01/2006", args[3])
	if err != nil {
		return shim.Error("timeType cant be converted," + err.Error())
	}

	loanBalance.TxnDate = timeType

	if args[7] == "disb" {
		if len(args) != 8 {
			xLenStr := strconv.Itoa(len(args))
			return shim.Error("Invalid number of arguments in updateLoanBal:disb (required:8) given:" + xLenStr)

		}

		timeType, err := time.Parse("02/01/2006", args[3])
		if err != nil {
			return shim.Error("timeType cant be converted," + err.Error())
		}

		//loanArgs := string(response.Payload)
		/*fmt.Printf("disbursement payload:%s\n", loanArgs)
		fmt.Println("loanArgs[0]", loanArgs[0])
		fmt.Println("[0] type:", reflect.TypeOf(loanArgs[0]))
		fmt.Println("loanArgs[1]", loanArgs[1])
		fmt.Println("[1] type:", reflect.TypeOf(loanArgs[1]))*/

		openBal, err := strconv.ParseInt(loanArgs[0], 10, 64)
		if err != nil {
			return shim.Error("Error in parsing the openbalance in LoanBalance: " + err.Error())
		}
		fmt.Println("Strconv is done")
		CAmt, err := strconv.ParseInt(args[5], 10, 64)
		if err != nil {
			return shim.Error("Error in parsing the CAmt in LoanBalance: " + err.Error())
		}
		DAmt, err := strconv.ParseInt(args[6], 10, 64)
		if err != nil {
			return shim.Error("Error in parsing the DAmt in LoanBalance: " + err.Error())
		}

		var status string
		status = loanArgs[1] // status of the current loan
		fmt.Println("status after received :", status)
		loanBal := openBal - DAmt + CAmt
		/*fmt.Println("openBal:", openBal)
		fmt.Println("DAmt:", DAmt)
		fmt.Println("CAmt:", CAmt)
		fmt.Println("loanBal:", loanBal)
		fmt.Println("above loanBalstring")*/
		loanBalString := strconv.FormatInt(loanBal, 10)
		if status == "sanctioned" || status == "partly disbursed" {

			if loanBal == 0 {
				status = "disbursed"
			} else {
				status = "partly disbursed"
			}
		}
		//fmt.Println("after the if statement")

		//Updating loanBalance ledger

		/*LoanID  -> args[1]
		*TxnID   -> args[2]
		*TxnDate -> args[3]
		*TxnType -> args[4]
		*DAmt    -> args[5]
		 */

		loanBalance.LoanID = args[1]
		loanBalance.TxnID = args[2]
		loanBalance.TxnDate = timeType
		loanBalance.TxnType = args[4]
		loanBalance.LoanStatus = status
		loanBalance.CAmt = CAmt
		loanBalance.DAmt = DAmt
		loanBalance.LoanBal = loanBal
		loanBalance.OpenBal = openBal

		loanBalanceBytes, _ = json.Marshal(loanBalance)
		stub.PutState(args[0], loanBalanceBytes)
		fmt.Println("written into loan balance ledger")

		fmt.Printf("Status:%s\n", status)
		chaincodeArgs = toChaincodeArgs("updateLoanInfo", args[1], status, loanBalString)
		fmt.Println("calling the other chaincode in if condition")
		response = stub.InvokeChaincode("loancc", chaincodeArgs, "myc")
		if response.Status != shim.OK {
			return shim.Error(response.Message)
		}
	}
	if args[7] == "inst" {
		if len(args) != 8 {
			xLenStr := strconv.Itoa(len(args))
			return shim.Error("Invalid number of arguments in updateLoanBal:inst (required:8) given:" + xLenStr)
		}
		/*
				// From Repayment
			*LoanBalId -> args[0]
			*LoanID  -> args[1]
			*TxnID   -> args[2]
			*TxnDate -> args[3]
			*TxnType -> args[4]
			*Amt    -> args[5] // Repayed Amt
			*insId  -> args[4]


			*OpenBal -> LoanBalance from Loan structure
			*CAmt

			*LoanBal -> OpenBal-DAmt+Camt
			*LoanStatus -> depends
		*/

		invoiceAmt := int64(1000) //Have to get from instrument
		sanctionedAmt, err := strconv.ParseInt(loanArgs[2], 10, 64)
		if err != nil {
			return shim.Error("Error in parsing the sanctionedAmt in LoanBalance: " + err.Error())
		}
		disbursedAmt := sanctionedAmt - loanBalance.LoanBal
		repayedAmt, err := strconv.ParseInt(args[5], 10, 64)
		if err != nil {
			return shim.Error("Error in parsing the repayedAmt in LoanBalance: " + err.Error())
		}

		var bankAssetVal int64
		var bankRefundVal int64
		var businessLoanVal int64
		if repayedAmt > disbursedAmt && repayedAmt == invoiceAmt {
			loanBalance.LoanStatus = "collected"
			bankAssetVal = disbursedAmt
			bankRefundVal = repayedAmt - loanBalance.OpenBal
			businessLoanVal = loanBalance.OpenBal
		} else if repayedAmt > disbursedAmt && repayedAmt < invoiceAmt {
			loanBalance.LoanStatus = "partly collected"
			bankAssetVal = loanBalance.OpenBal
			bankRefundVal = repayedAmt - loanBalance.OpenBal
			businessLoanVal = loanBalance.OpenBal
		} else if repayedAmt == loanBalance.LoanBal {
			loanBalance.LoanStatus = "partly collected"
			bankAssetVal = repayedAmt
			bankRefundVal = 0
			businessLoanVal = repayedAmt
		} else if repayedAmt < disbursedAmt {
			loanBalance.LoanStatus = "partly collected"
			bankAssetVal = repayedAmt
			businessLoanVal = repayedAmt
		}
		bankAssetValString := strconv.FormatInt(bankAssetVal, 10)
		bankRefundValString := strconv.FormatInt(bankRefundVal, 10)
		businessLoanValString := strconv.FormatInt(businessLoanVal, 10)

		returnVal := bankAssetValString + "," + bankRefundValString + "," + businessLoanValString

		loanBalance.LoanID = args[1]
		loanBalance.TxnID = args[2]
		loanBalance.TxnDate = timeType
		loanBalance.TxnType = args[4]
		loanBalance.CAmt = 0
		loanBalance.DAmt = repayedAmt
		loanBalance.LoanBal = 0
		loanBalance.OpenBal = 0

		loanBalanceBytes, _ = json.Marshal(loanBalance)
		stub.PutState(args[0], loanBalanceBytes)
		fmt.Println("written into loan balance ledger")

		repayedAmtString := strconv.FormatInt(repayedAmt, 10)
		fmt.Printf("Status:%s\n", "collected")
		chaincodeArgs = toChaincodeArgs("updateLoanInfo", args[1], "collected", repayedAmtString)
		fmt.Println("calling the other chaincode")
		response = stub.InvokeChaincode("loancc", chaincodeArgs, "myc")
		if response.Status != shim.OK {
			return shim.Error(response.Message)
		}

		return shim.Success([]byte(returnVal))
	}
	return shim.Success(nil)
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
		fmt.Printf("Error starting LoanBalance chaincode: %s\n", err)
	}

}
