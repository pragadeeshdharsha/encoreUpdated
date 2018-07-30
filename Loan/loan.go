package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
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
	InstNum                     string //Instrument Number
	ExposureBusinessID          string //buyer for now
	ProgramID                   string
	SanctionAmt                 int64
	SanctionDate                time.Time //with time
	SanctionAuthority           string
	ROI                         float64
	DueDate                     time.Time
	ValueDate                   time.Time //with time
	LoanStatus                  string
	LoanDisbursedWalletID       string
	LoanChargesWalletID         string
	LoanAccruedInterestWalletID string
	BuyerBusinessID             string
	SellerBusinessID            string
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
	} else if function == "loanIDexists" {
		return loanIDexists(stub, args[0])
	} else if function == "loanStatusSancAmt" {
		return loanStatusSancAmt(stub, args[0])
	} else if function == "getWalletID" {
		return getWalletID(stub, args)
	}
	return shim.Error("No function named " + function + " in Loan")
}

func newLoanInfo(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 14 {
		xLenStr := strconv.Itoa(len(args))
		return shim.Error("Invalid number of arguments in newLoanInfo(loan) (required:14) given: " + xLenStr)
	}
	//Checking existence of loanID
	response := loanIDexists(stub, args[0])
	if response.Status != shim.OK {
		return shim.Error(response.Message)
	}

	//Checking existence of ExposureBusinessID
	chaincodeArgs := toChaincodeArgs("busIDexists", args[2])
	response = stub.InvokeChaincode("businesscc", chaincodeArgs, "myc")
	if response.Status == shim.OK {
		return shim.Error("ExposureBusinessID " + args[2] + " does not exits")
	}

	//Checking if Instrument ID is Instrument Ref. No.
	chaincodeArgs = toChaincodeArgs("getSellerIDnAmt", args[1])
	response = stub.InvokeChaincode("instrumentcc", chaincodeArgs, "myc")
	if response.Status != shim.OK {
		return shim.Error("Instrument refrence no " + args[1] + " does not exits")
	}

	// getting the sanction amount from the instrument
	instAmtStr := strings.Split(string(response.Payload), ",")[1]
	instAmt, err := strconv.ParseInt(instAmtStr, 10, 64)
	if err != nil {
		return shim.Error("Unable to parse instAmt(loan):" + err.Error())
	}

	//Getting the discount percentage
	chaincodeArgs = toChaincodeArgs("discountPercentage", args[3], args[2])
	response = stub.InvokeChaincode("pprcc", chaincodeArgs, "myc")
	if response.Status == shim.OK {
		return shim.Error("PprId " + args[8] + " does not exits")
	}

	discountPercentStr := string(response.Payload)
	discountPercent, _ := strconv.ParseInt(discountPercentStr, 10, 64)
	amt := instAmt - ((discountPercent * instAmt) / 100)

	//SanctionAmt -> sAmt
	sAmt, err := strconv.ParseInt(args[4], 10, 64)
	if err != nil {
		return shim.Error(err.Error())
	}

	if sAmt > amt && sAmt == 0 {
		return shim.Error("Sanction amount exceeds the required value or it is zero : " + args[4])
	}

	//SanctionDate ->sDate
	sDate := time.Now()

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

	hash := sha256.New()

	// Hashing LoanDisbursedWalletID
	LoanDisbursedWalletStr := args[11] + "LoanDisbursedWallet"
	hash.Write([]byte(LoanDisbursedWalletStr))
	md := hash.Sum(nil)
	LoanDisbursedWalletIDsha := hex.EncodeToString(md)
	createWallet(stub, LoanDisbursedWalletIDsha, args[11])

	// Hashing LoanChargesWalletID
	LoanChargesWalletStr := args[12] + "LoanChargesWallet"
	hash.Write([]byte(LoanChargesWalletStr))
	md = hash.Sum(nil)
	LoanChargesWalletIDsha := hex.EncodeToString(md)
	createWallet(stub, LoanChargesWalletIDsha, args[12])

	// Hashing LoanAccruedInterestWalletID
	LoanAccruedInterestWalletStr := args[13] + "LoanAccruedInterestWallet"
	hash.Write([]byte(LoanAccruedInterestWalletStr))
	md = hash.Sum(nil)
	LoanAccruedInterestWalletIDsha := hex.EncodeToString(md)
	createWallet(stub, LoanAccruedInterestWalletIDsha, args[13])

	//Checking existence of BuyerBusinessID
	chaincodeArgs = toChaincodeArgs("busIDexists", args[14])
	response = stub.InvokeChaincode("businesscc", chaincodeArgs, "myc")
	if response.Status == shim.OK {
		return shim.Error("BuyerBusinessID " + args[14] + " does not exits")
	}

	//Checking existence of SellerBusinessID
	chaincodeArgs = toChaincodeArgs("busIDexists", args[15])
	response = stub.InvokeChaincode("businesscc", chaincodeArgs, "myc")
	if response.Status == shim.OK {
		return shim.Error("SellerBusinessID " + args[15] + " does not exits")
	}

	loan := loanInfo{args[1], args[2], args[3], sAmt, sDate, args[6], roi, dDate, vDate, "sanctioned", LoanDisbursedWalletIDsha, LoanChargesWalletIDsha, LoanAccruedInterestWalletIDsha, args[14], args[15]}
	loanBytes, err := json.Marshal(loan)
	if err != nil {
		return shim.Error(err.Error())
	}
	stub.PutState(args[0], loanBytes)
	return shim.Success(nil)
}

func loanIDexists(stub shim.ChaincodeStubInterface, loanID string) pb.Response {
	ifExists, _ := stub.GetState(loanID)
	if ifExists != nil {
		fmt.Println(ifExists)
		return shim.Error("LoanId " + loanID + " exits. Cannot create new ID")
	}
	return shim.Success(nil)
}

func loanStatusSancAmt(stub shim.ChaincodeStubInterface, loanID string) pb.Response {
	loanBytes, err := stub.GetState(loanID)
	if err != nil {
		return shim.Error(err.Error())
	} else if loanBytes == nil {
		return shim.Error("No data exists on this loanID: " + loanID)
	}

	loan := loanInfo{}
	err = json.Unmarshal(loanBytes, &loan)
	if err != nil {
		return shim.Error("Error unmarshiling in loanstatus(loan):" + err.Error())
	}

	sancAmtString := strconv.FormatInt(loan.SanctionAmt, 64)
	return shim.Success([]byte(loan.LoanStatus + "," + sancAmtString))
}

func getWalletID(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	loanBytes, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error(err.Error())
	} else if loanBytes == nil {
		return shim.Error("No data exists on this loanID: " + args[0])
	}

	loan := loanInfo{}
	err = json.Unmarshal(loanBytes, &loan)
	if err != nil {
		return shim.Error("Unable to parse into loan the structure (loanWalletValues)" + err.Error())
	}

	walletID := ""

	switch args[1] {
	case "accrued":
		walletID = loan.LoanAccruedInterestWalletID
	case "charges":
		walletID = loan.LoanChargesWalletID
	case "disbursed":
		walletID = loan.LoanDisbursedWalletID
	default:
		return shim.Error("There is no wallet of this type in Loan :" + args[1])
	}

	return shim.Success([]byte(walletID))
}
func createWallet(stub shim.ChaincodeStubInterface, walletID string, amt string) pb.Response {
	chaincodeArgs := toChaincodeArgs("newWallet", walletID, amt)
	response := stub.InvokeChaincode("walletcc", chaincodeArgs, "myc")
	if response.Status != shim.OK {
		return shim.Error("Unable to create new wallet from business")
	}
	return shim.Success([]byte("created new wallet from business"))
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
	if args[2] == "disbursed" {
		//Updating Loan status for disbursement
		loan.LoanStatus = args[1]
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
