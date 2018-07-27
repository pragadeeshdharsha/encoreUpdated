package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

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

	if function == "newRepayInfo" {
		return newRepayInfo(stub, args)
	}
	return shim.Error("no function named " + function + " found in Repayment")
}

func newRepayInfo(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) == 1 {
		args = strings.Split(args[0], ",")
	}
	if len(args) != 10 {
		xLenStr := strconv.Itoa(len(args))
		return shim.Error("Invalid number of arguments in newRepayInfo(repayment) (required:10) given:" + xLenStr)
	}

	/*
	 *TxnType string    //args[1]
	 *TxnDate time.Time //args[2]
	 *LoanID  string    //args[3]
	 *InsID   string    //args[4]
	 *Amt     int64     //args[5]
	 *FromID  string    //args[6]  Business
	 *ToID    string    //args[7]  Bank
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

	//#####################################################################################################################
	//Calling for updating Business Main_Wallet
	//####################################################################################################################

	cAmtString := "0"
	dAmtString := args[5]

	walletID, openBalString, txnBalString, err := getWalletInfo(stub, args[6], "main", "businesscc", cAmtString, dAmtString)
	if err != nil {
		return shim.Error(err.Error())
	}

	// STEP-4 generate txn_balance_object and write it to the Txn_Bal_Ledger
	argsList := []string{"5", args[0], args[2], args[3], args[4], walletID, openBalString, args[1], args[5], cAmtString, dAmtString, txnBalString, args[8]}
	argsListStr := strings.Join(argsList, ",")
	chaincodeArgs := toChaincodeArgs("putTxnInfo", argsListStr)
	fmt.Println("calling the other chaincode business main")
	response := stub.InvokeChaincode("txnbalcc", chaincodeArgs, "myc")
	if response.Status != shim.OK {
		return shim.Error(response.Message)
	}
	fmt.Println(string(response.GetPayload()))

	//####################################################################################################################
	//Calling for updating Bank Main_Wallet
	//####################################################################################################################

	cAmtString = args[5]
	dAmtString = "0"

	walletID, openBalString, txnBalString, err = getWalletInfo(stub, args[7], "main", "bankcc", cAmtString, dAmtString)
	if err != nil {
		return shim.Error(err.Error())
	}

	// STEP-4 generate txn_balance_object and write it to the Txn_Bal_Ledger
	argsList = []string{"6", args[0], args[2], args[3], args[4], walletID, openBalString, args[1], args[5], cAmtString, dAmtString, txnBalString, args[8]}
	argsListStr = strings.Join(argsList, ",")
	chaincodeArgs = toChaincodeArgs("putTxnInfo", argsListStr)
	fmt.Println("calling the other chaincode bank main")
	response = stub.InvokeChaincode("txnbalcc", chaincodeArgs, "myc")
	if response.Status != shim.OK {
		return shim.Error(response.Message)
	}
	fmt.Println(string(response.GetPayload()))

	//####################################################################################################################
	//Calling for updating Business Liability_Wallet
	//####################################################################################################################

	cAmtString = "0"
	dAmtString = args[5]

	walletID, openBalString, txnBalString, err = getWalletInfo(stub, args[6], "liability", "businesscc", cAmtString, dAmtString)
	if err != nil {
		return shim.Error(err.Error())
	}

	// STEP-4 generate txn_balance_object and write it to the Txn_Bal_Ledger
	argsList = []string{"7", args[0], args[2], args[3], args[4], walletID, openBalString, args[1], args[5], cAmtString, dAmtString, txnBalString, args[8]}
	argsListStr = strings.Join(argsList, ",")
	chaincodeArgs = toChaincodeArgs("putTxnInfo", argsListStr)
	fmt.Println("calling the other chaincode business liability")
	response = stub.InvokeChaincode("txnbalcc", chaincodeArgs, "myc")
	if response.Status != shim.OK {
		return shim.Error(response.Message)
	}
	fmt.Println(string(response.GetPayload()))

	//####################################################################################################################
	//Calling for Business Loan Balance Update
	//####################################################################################################################
	argsList = []string{"1loanbal", args[3], args[0], args[2], args[1], args[5], args[4], "inst"}
	argsListString := strings.Join(argsList, ",")
	chaincodeArgs = toChaincodeArgs("updateLoanBal", argsListString)
	//sending to loanBalUp chaincode not loanBalance Chaincode
	response = stub.InvokeChaincode("loanbalcc", chaincodeArgs, "myc")
	if response.Status != shim.OK {
		return shim.Error(response.Message)
	}
	fmt.Println("Getting the payload from updateLoan bal (inst)")
	payLoad := strings.Split(string(response.Payload), ",")

	//payload[0] -> bankAssetVal
	//payload[1] -> bankRefundVal
	//payload[2] -> businessLoanVal

	//####################################################################################################################
	//4.Calling for updating Business Loan_Wallet
	//####################################################################################################################

	// Calling getSellerID (instrument) to get the seller ID

	chaincodeArgs = toChaincodeArgs("getSellerID", args[4])
	response = stub.InvokeChaincode("instrumentcc", chaincodeArgs, "myc")
	if response.Status != shim.OK {
		return shim.Error("Error in getting the instrument id:" + response.Message)
	}

	bus2ID := string(response.Payload)

	// STEP-4 generate txn_balance_object and write it to the Txn_Bal_Ledger

	/*chaincodeArgs = toChaincodeArgs("getWalletID", bus2ID, "loan")
	response = stub.InvokeChaincode("businesscc", chaincodeArgs, "myc")
	if response.Status != shim.OK {
		return shim.Error("Retreiving Business loan wallet in repayment " + response.Message)
	}*/
	walletID, err = getWalletIDonly(stub, "businesscc", bus2ID, "loan")
	if err != nil {
		return shim.Error("business loan wallet (repayment) err : " + err.Error())
	}

	walletArgs := toChaincodeArgs("updateWallet", walletID, payLoad[2])
	walletResponse := stub.InvokeChaincode("walletcc", walletArgs, "myc")
	if walletResponse.Status != shim.OK {
		return shim.Error("Wallet updation Business loan wallet " + response.Message)
	}

	argsList = []string{"6", args[0], args[2], args[3], args[4], walletID, openBalString, args[1], args[5], cAmtString, dAmtString, txnBalString, args[8]}
	argsListStr = strings.Join(argsList, ",")
	chaincodeArgs = toChaincodeArgs("putTxnInfo", argsListStr)
	fmt.Println("calling the other chaincode")
	response = stub.InvokeChaincode("txnbalcc", chaincodeArgs, "myc")
	if response.Status != shim.OK {
		return shim.Error(response.Message)
	}
	fmt.Println(string(response.GetPayload()))

	//####################################################################################################################
	//Calling for updating Bank Refund_Wallet
	//####################################################################################################################

	/*cAmtString = "0"
	dAmtString = args[5]*/

	/*walletID, openBalString, txnBalString, err = getWalletInfo(stub, args[6], "refund", "bankcc", cAmtString, dAmtString)
	if err != nil {
		return shim.Error(err.Error())
	}*/

	/*chaincodeArgs = toChaincodeArgs("getWalletID", args[7], "liability")
	response = stub.InvokeChaincode("bankcc", chaincodeArgs, "myc")
	if response.Status != shim.OK {
		return shim.Error("Retreiving bank refund wallet in repayment " + response.Message)
	}*/
	walletID, err = getWalletIDonly(stub, "bankcc", args[7], "liability")
	if err != nil {
		return shim.Error("bank liability wallet (repayment) err : " + err.Error())
	}

	walletArgs = toChaincodeArgs("updateWallet", walletID, payLoad[1])
	walletResponse = stub.InvokeChaincode("walletcc", walletArgs, "myc")
	if walletResponse.Status != shim.OK {
		return shim.Error("Wallet updation bank refund wallet " + response.Message)
	}

	// STEP-4 generate txn_balance_object and write it to the Txn_Bal_Ledger
	argsList = []string{"9", args[0], args[2], args[3], args[4], walletID, openBalString, args[1], args[5], payLoad[1], "0", "500", args[8]}
	argsListStr = strings.Join(argsList, ",")
	chaincodeArgs = toChaincodeArgs("putTxnInfo", argsListStr)
	fmt.Println("calling the other chaincode bank refund")
	response = stub.InvokeChaincode("txnbalcc", chaincodeArgs, "myc")
	if response.Status != shim.OK {
		return shim.Error(response.Message)
	}
	fmt.Println(string(response.GetPayload()))

	//####################################################################################################################
	//Calling for updating Bank Asset Wallet
	//####################################################################################################################

	/*cAmtString = "0"
	dAmtString = args[5]*/

	/*walletID, openBalString, txnBalString, err = getWalletInfo(stub, args[6], "refund", "bankcc", cAmtString, dAmtString)
	if err != nil {
		return shim.Error(err.Error())
	}*/

	/*chaincodeArgs = toChaincodeArgs("getWalletID", args[7], "asset")
	response = stub.InvokeChaincode("bankcc", chaincodeArgs, "myc")
	if response.Status != shim.OK {
		return shim.Error("Retreiving bank asset wallet in repayment " + response.Message)
	}*/
	walletID, err = getWalletIDonly(stub, "bankcc", args[7], "asset")
	if err != nil {
		return shim.Error("bank asset wallet (repayment) err : " + err.Error())
	}
	walletArgs = toChaincodeArgs("getWallet", walletID)
	walletResponse = stub.InvokeChaincode("walletcc", walletArgs, "myc")
	if walletResponse.Status != shim.OK {
		return shim.Error("getting wallet open bal in bank asset")
	}
	openBalString = string(walletResponse.Payload)

	walletArgs = toChaincodeArgs("updateWallet", walletID, payLoad[0])
	walletResponse = stub.InvokeChaincode("walletcc", walletArgs, "myc")
	if walletResponse.Status != shim.OK {
		return shim.Error("Wallet updation bank asset wallet " + response.Message)
	}

	// STEP-4 generate txn_balance_object and write it to the Txn_Bal_Ledger
	argsList = []string{"10", args[0], args[2], args[3], args[4], walletID, openBalString, args[1], args[5], "0", payLoad[0], "500", args[8]}
	argsListStr = strings.Join(argsList, ",")
	chaincodeArgs = toChaincodeArgs("putTxnInfo", argsListStr)
	fmt.Println("calling the other chaincode bank asset")
	response = stub.InvokeChaincode("txnbalcc", chaincodeArgs, "myc")
	if response.Status != shim.OK {
		return shim.Error(response.Message)
	}
	fmt.Println(string(response.GetPayload()))

	//####################################################################################################################

	return shim.Success(nil)
}

func getWalletInfo(stub shim.ChaincodeStubInterface, participantID string, walletType string, ccName string, cAmtStr string, dAmtStr string) (string, string, string, error) {

	//STEP-1
	// Getting wallet id from the chaincode
	walletID, err := getWalletIDonly(stub, ccName, participantID, walletType)
	if err != nil {
		return "", "", "", err
	}

	// STEP-2
	// getting Balance from walletID
	// walletFcn := "getWallet"
	walletArgs := toChaincodeArgs("getWallet", walletID)
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

	walletArgs = toChaincodeArgs("updateWallet", walletID, txnBalString)
	walletResponse = stub.InvokeChaincode("walletcc", walletArgs, "myc")
	if walletResponse.Status != shim.OK {
		return "", "", "", errors.New(walletResponse.Message)
	}

	return walletID, openBalString, txnBalString, nil
}

func getWalletIDonly(stub shim.ChaincodeStubInterface, ccName string, id string, walletType string) (string, error) {

	// STEP-1
	// using FromID, get a walletID from bank structure

	chaincodeArgs := toChaincodeArgs("getWalletID", id, walletType)
	response := stub.InvokeChaincode(ccName, chaincodeArgs, "myc")
	if response.Status != shim.OK {
		return "", errors.New(response.Message)
	}
	walletID := string(response.GetPayload())
	return walletID, nil
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
		fmt.Println("Unable to start the chaincode:", err)
	}
}
