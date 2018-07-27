package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type chainCode struct {
}

func (c *chainCode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (c *chainCode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()

	if function == "newDisbInfo" {
		return newDisbInfo(stub, args)
	}
	return shim.Error("no function named " + function + " found in disbursement")
}

func newDisbInfo(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) == 1 {
		args = strings.Split(args[0], ",")
	}
	if len(args) != 10 {
		xLenStr := strconv.Itoa(len(args))
		return shim.Error("Invalid number of arguments in newDisbInfo(disbursement) (required:10) given:" + xLenStr)
	}

	/*
	 *TxnType string    //args[1]
	 *TxnDate time.Time //args[2]
	 *LoanID  string    //args[3]
	 *InsID   string    //args[4]
	 *Amt     int64     //args[5]
	 *FromID  string    //args[6]
	 *ToID    string    //args[7]
	 *By      string    //args[8]
	 *PprID   string    //args[9]
	 */

	///////////////////////////////////////////////////////////////////////////////////////////////////
	// 				UPDATING WALLETS																///
	///////////////////////////////////////////////////////////////////////////////////////////////////
	// The transaction object has been created and written into the ledger
	// The JSON file is 'transaction'function
	// Now to create a TXN_Bal_Update obj for 4 times
	// Calling TXN_Balance CC based on TXN_Type {ex: Disbursement}
	/*
	 *	bank main wallet reduced
	 * 	bank asset wallet incresed
	 *	business main wallet increased
	 *	business loan wallet increased
	 */

	//####################################################################################################################
	//Calling for updating Bank Main_Wallet
	//####################################################################################################################

	cAmtString := "0"
	dAmtString := args[5]

	walletID, openBalString, txnBalString, err := getWalletInfo(stub, args[6], "main", "bankcc", cAmtString, dAmtString)
	if err != nil {
		return shim.Error(err.Error())
	}

	// STEP-4 generate txn_balance_object and write it to the Txn_Bal_Ledger
	argsList := []string{"1", args[0], args[2], args[3], args[4], walletID, openBalString, args[1], args[5], cAmtString, dAmtString, txnBalString, args[8]}
	argsListStr := strings.Join(argsList, ",")
	chaincodeArgs := util.ToChaincodeArgs("putTxnInfo", argsListStr)
	fmt.Println("calling the other chaincode")
	response := stub.InvokeChaincode("txnbalcc", chaincodeArgs, "myc")
	if response.Status != shim.OK {
		return shim.Error(response.Message)
	}
	fmt.Println(string(response.GetPayload()))
	//successfully updated Bank's main wallet and written the txn thing to the ledger

	//#####################################################################################################################
	//Calling for updating Business Main_Wallet
	//####################################################################################################################

	cAmtString = args[5]
	dAmtString = "0"

	walletID, openBalString, txnBalString, err = getWalletInfo(stub, args[7], "main", "businesscc", cAmtString, dAmtString)
	if err != nil {
		return shim.Error(err.Error())
	}

	// STEP-4 generate txn_balance_object and write it to the Txn_Bal_Ledger
	argsList = []string{"2", args[0], args[2], args[3], args[4], walletID, openBalString, args[1], args[5], cAmtString, dAmtString, txnBalString, args[8]}
	argsListStr = strings.Join(argsList, ",")
	chaincodeArgs = util.ToChaincodeArgs("putTxnInfo", argsListStr)
	fmt.Println("calling the other chaincode")
	response = stub.InvokeChaincode("txnbalcc", chaincodeArgs, "myc")
	if response.Status != shim.OK {
		return shim.Error(response.Message)
	}
	fmt.Println(string(response.GetPayload()))
	//successfully updated Bank's main wallet and written the txn thing to the ledger

	//####################################################################################################################
	//Calling for updating Bank Refund_Wallet
	//####################################################################################################################

	cAmtString = "0"
	dAmtString = args[5]

	walletID, openBalString, txnBalString, err = getWalletInfo(stub, args[6], "refund", "bankcc", cAmtString, dAmtString)
	if err != nil {
		return shim.Error(err.Error())
	}

	// STEP-4 generate txn_balance_object and write it to the Txn_Bal_Ledger
	argsList = []string{"5", args[0], args[2], args[3], args[4], walletID, openBalString, args[1], args[5], cAmtString, dAmtString, txnBalString, args[8]}
	argsListStr = strings.Join(argsList, ",")
	chaincodeArgs = util.ToChaincodeArgs("putTxnInfo", argsListStr)
	fmt.Println("calling the other chaincode")
	response = stub.InvokeChaincode("txnbalcc", chaincodeArgs, "myc")
	if response.Status != shim.OK {
		return shim.Error(response.Message)
	}
	fmt.Println(string(response.GetPayload()))
	//####################################################################################################################

	//####################################################################################################################
	//Calling for Loan Balance Update
	//####################################################################################################################
	CAmt := "0"
	DAmt := args[5]
	argStrings := []string{"1loanBal", args[3], args[0], args[2], args[1], CAmt, DAmt} // 6 variables for updateLoanBalance
	// args[5] -> DAmt for loanBalance
	argStr := strings.Join(argStrings, ",")
	chaincodeArgs = util.ToChaincodeArgs("updateLoanBal", argStr)
	//sending to loanBalUp chaincode not loanBalance Chaincode
	response = stub.InvokeChaincode("loanbalcc", chaincodeArgs, "myc")
	if response.Status != shim.OK {
		return shim.Error(response.Message)
	}
	//walletID := string(response.GetPayload())

	//####################################################################################################################
	//Calling for updating Bank Asset_Wallet
	//####################################################################################################################

	cAmtString = "0"
	dAmtString = args[5]

	walletID, openBalString, txnBalString, err = getWalletInfo(stub, args[6], "asset", "bankcc", cAmtString, dAmtString)
	if err != nil {
		return shim.Error(err.Error())
	}

	// STEP-4 generate txn_balance_object and write it to the Txn_Bal_Ledger
	argsList = []string{"4", args[0], args[2], args[3], args[4], walletID, openBalString, args[1], args[5], cAmtString, dAmtString, txnBalString, args[8]}
	argsListStr = strings.Join(argsList, ",")
	chaincodeArgs = util.ToChaincodeArgs("putTxnInfo", argsListStr)
	fmt.Println("calling the other chaincode")
	response = stub.InvokeChaincode("txnbalcc", chaincodeArgs, "myc")
	if response.Status != shim.OK {
		return shim.Error(response.Message)
	}
	fmt.Println(string(response.GetPayload()))
	//successfully updated Bank's main wallet and written the txn thing to the ledger

	return shim.Success(nil)
}

func getWalletInfo(stub shim.ChaincodeStubInterface, participantID string, walletType string, ccName string, cAmtStr string, dAmtStr string) (string, string, string, error) {

	// STEP-1
	// using FromID, get a walletID from bank structure
	// bankID = bankID

	chaincodeArgs := util.ToChaincodeArgs("getWalletID", participantID, walletType)
	response := stub.InvokeChaincode(ccName, chaincodeArgs, "myc")
	if response.Status != shim.OK {
		return "", "", "", errors.New(response.Message)
	}
	walletID := string(response.GetPayload())

	// STEP-2
	// getting Balance from walletID
	// walletFcn := "getWallet"
	walletArgs := util.ToChaincodeArgs("getWallet", walletID)
	walletResponse := stub.InvokeChaincode("walletcc", walletArgs, "myc")
	if walletResponse.Status != shim.OK {
		return "", "", "", errors.New(walletResponse.Message)
	}
	openBalString := string(walletResponse.Payload)

	openBal, err := strconv.ParseInt(openBalString, 10, 64)
	if err != nil {
		return "", "", "", errors.New("Error in converting the openBalance")
	}
	cAmt, err := strconv.ParseInt(cAmtStr, 10, 64)
	if err != nil {
		return "", "", "", errors.New("Error in converting the cAmt")
	}
	dAmt, err := strconv.ParseInt(dAmtStr, 10, 64)
	if err != nil {
		return "", "", "", errors.New("Error in converting the dAmt")
	}

	txnBal := openBal - dAmt + cAmt
	txnBalString := strconv.FormatInt(txnBal, 10)

	// STEP-3
	// update wallet of ID walletID here, and write it to the wallet_ledger
	// walletFcn := "updateWallet"

	walletArgs = util.ToChaincodeArgs("updateWallet", walletID, txnBalString)
	walletResponse = stub.InvokeChaincode("walletcc", walletArgs, "myc")
	if walletResponse.Status != shim.OK {
		return "", "", "", errors.New(walletResponse.Message)
	}

	return walletID, openBalString, txnBalString, nil
}

func main() {
	err := shim.Start(new(chainCode))
	if err != nil {
		fmt.Println("Unable to start the chaincode")
	}
}
