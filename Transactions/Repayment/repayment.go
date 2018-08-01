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
	// Now to create a TXN_Bal_Update obj for 10 times
	// Calling TXN_Balance CC based on TXN_Type {ex: Disbursement}
	/*
			    a. Debiting (decreasing) Business Wallet (Buyer)
		            i. Txn amt
		        b. Crediting (Increasing) Bank Wallet
		            i. Txn amt
		        c. Debiting (decreasing) Bank Asset Wallet
		            i. Loan Disbursed Wallet Balance + Loan Charges Wallet Balance
		        d. Crediting (Increasing) Bank Refund Wallet (if applicable)
		            i. Txn Amt – Loan Disbursed Wallet balance – Loan Charges Wallet Balance
		        e. Debiting (decreasing) Business Loan Wallet (Seller)
		            i. If Txn Amt is >/= (Loan Charges Wallet Balance + Loan Disbursed Wallet Balance)
		                1. Loan disbursed Wallet Balance + Loan Charges Wallet Balance
		            ii. If Txn Amt is  < (Loan Charges Wallet Balance + Loan Disbursed Wallet Balance)
		                1. Txn Amt – Loan Charges Wallet Balance
		        f. Debiting (decreasing) Business Charges O/s Wallet
		            i. Loan Charges Wallet Balance
		        g. Debiting (decreasing) Business Principal O/s Wallet
		            i. Loan Disbursed Wallet Balance
		        h. Debiting (Decreasing) Loan Charges Wallet
		            i. Loan Charges Wallet Balance
		        i. Debiting (Decreasing) Loan Disbursed Wallet
		            i. If Txn Amt is >/= (Loan Charges Wallet Balance + Loan Disbursed Wallet Balance)
		                1. Loan Disbursed Wallet is reduced to Zero
		                2. Loan Status is updated to Collected
		            ii. If Txn Amt is  < (Loan Charges Wallet Balance + Loan Disbursed Wallet Balance)
		                1. Txn Amt – Loan Charges Wallet Balance
		                2. Loan Status is updated to Part Collected
		        j. Debiting (Decreasing) Business Liability Wallet (Buyer)
		            i. Txn Amt
	*/

	amt, _ := strconv.ParseInt(args[5], 10, 64)

	//#####################################################################################################################
	//Calling for updating Business Main_Wallet
	//####################################################################################################################

	cAmtString := "0"
	dAmtString := args[5]

	walletID, err := getWalletID(stub, "businesscc", args[6], "main")
	if err != nil {
		return shim.Error("Repayment Business Main WalletID " + err.Error())
	}

	openBalance, err := getWalletValue(stub, walletID)
	if err != nil {
		return shim.Error("Repayment Business Main WalletValue " + err.Error())
	}
	openBalString := strconv.FormatInt(openBalance, 10)
	bal := openBalance - amt

	response := walletUpdation(stub, walletID, bal)
	if response.Status != shim.OK {
		return shim.Error(response.Message)
	}
	txnBalString := strconv.FormatInt(bal, 10)
	// STEP-4 generate txn_balance_object and write it to the Txn_Bal_Ledger

	argsList := []string{"1rep", args[0], args[2], args[3], args[4], walletID, openBalString, args[1], args[5], cAmtString, dAmtString, txnBalString, args[8]}
	argsListStr := strings.Join(argsList, ",")
	txnResponse := putInTxnBal(stub, argsListStr)
	if txnResponse.Status != shim.OK {
		return shim.Error(txnResponse.Message)
	}

	//####################################################################################################################
	//Calling for updating Bank Main_Wallet
	//####################################################################################################################

	cAmtString = args[5]
	dAmtString = "0"

	walletID, err = getWalletID(stub, "bankcc", args[7], "main")
	if err != nil {
		return shim.Error("Repayment Bank Main WalletID " + err.Error())
	}

	openBalance, err = getWalletValue(stub, walletID)
	if err != nil {
		return shim.Error("Repayment Bank Main WalletValue " + err.Error())
	}
	openBalString = strconv.FormatInt(openBalance, 10)

	//amt, _ = strconv.ParseInt(args[5], 10, 64)

	bal = openBalance + amt
	txnBalString = strconv.FormatInt(bal, 10)

	response = walletUpdation(stub, walletID, bal)
	if response.Status != shim.OK {
		return shim.Error(response.Message)
	}
	argsList = []string{"2rep", args[0], args[2], args[3], args[4], walletID, openBalString, args[1], args[5], cAmtString, dAmtString, txnBalString, args[8]}
	argsListStr = strings.Join(argsList, ",")
	txnResponse = putInTxnBal(stub, argsListStr)
	if txnResponse.Status != shim.OK {
		return shim.Error(txnResponse.Message)
	}

	//####################################################################################################################
	//Calling for updating Bank Asset Wallet
	//####################################################################################################################

	cAmtString = "0"

	//Loan Disbursed Wallet balance
	loanDisbursedWalletID, err := getWalletID(stub, "loancc", args[3], "disbursed")
	if err != nil {
		return shim.Error("Repayment loanDisbursedWalletID " + err.Error())
	}
	loanDisbursedWalletValue, err := getWalletValue(stub, loanDisbursedWalletID)
	if err != nil {
		return shim.Error("Repayment loanDisbursedWalletValue " + err.Error())
	}

	//Loan Charges Wallet Balance
	loanChargesWalletID, err := getWalletID(stub, "loancc", args[3], "charges")
	if err != nil {
		return shim.Error("Repayment loanChargesWalletID " + err.Error())
	}
	loanChargesWalletValue, err := getWalletValue(stub, loanChargesWalletID)
	if err != nil {
		return shim.Error("Repayment loanChargesWalletValue " + err.Error())
	}

	//Bank Asset Wallet
	walletID, err = getWalletID(stub, "bankcc", args[7], "asset")
	if err != nil {
		return shim.Error("Repayment Bank Asset WalletID " + err.Error())
	}

	openBalance, err = getWalletValue(stub, walletID)
	if err != nil {
		return shim.Error("Repayment Bank Asset WalletValue " + err.Error())
	}
	openBalString = strconv.FormatInt(openBalance, 10)

	bal = openBalance - loanChargesWalletValue - loanDisbursedWalletValue
	txnBalString = strconv.FormatInt(bal, 10)

	response = walletUpdation(stub, walletID, bal)
	if response.Status != shim.OK {
		return shim.Error("Repayment Bank Asset Wallet " + response.Message)
	}
	dAmt := loanChargesWalletValue + loanDisbursedWalletValue
	dAmtString = strconv.FormatInt(dAmt, 10)
	argsList = []string{"3rep", args[0], args[2], args[3], args[4], walletID, openBalString, args[1], args[5], cAmtString, dAmtString, txnBalString, args[8]}
	argsListStr = strings.Join(argsList, ",")
	txnResponse = putInTxnBal(stub, argsListStr)
	if txnResponse.Status != shim.OK {
		return shim.Error(txnResponse.Message)
	}

	//####################################################################################################################
	//Calling for updating Bank Refund_Wallet
	//####################################################################################################################

	dAmtString = "0"
	walletID, err = getWalletID(stub, "bankcc", args[7], "liability")
	if err != nil {
		return shim.Error("Repayment Bank Liability WalletID " + err.Error())
	}

	openBalance, err = getWalletValue(stub, walletID)
	if err != nil {
		return shim.Error("Repayment Bank Liability WalletValue " + err.Error())
	}
	openBalString = strconv.FormatInt(openBalance, 10)

	cAmt := amt - loanChargesWalletValue - loanDisbursedWalletValue
	if cAmt > 0 {
		bal = openBalance + cAmt
	} else {
		bal = openBalance
	}
	response = walletUpdation(stub, walletID, bal)
	if response.Status != shim.OK {
		return shim.Error("Repayment Bank Liability Wallet " + response.Message)
	}
	txnBalString = strconv.FormatInt(bal, 10)
	cAmtString = strconv.FormatInt(cAmt, 10)

	argsList = []string{"4rep", args[0], args[2], args[3], args[4], walletID, openBalString, args[1], args[5], cAmtString, dAmtString, txnBalString, args[8]}
	argsListStr = strings.Join(argsList, ",")
	txnResponse = putInTxnBal(stub, argsListStr)
	if txnResponse.Status != shim.OK {
		return shim.Error(txnResponse.Message)
	}

	//####################################################################################################################
	//Calling for updating Business Loan_Wallet (seller)
	//####################################################################################################################

	// geting seller's ID using loan ID

	cAmtString = "0"
	chaincodeArgs := toChaincodeArgs("getSellerID", args[3])
	response = stub.InvokeChaincode("loancc", chaincodeArgs, "myc")
	if response.Status != shim.OK {
		return shim.Error(response.Message)
	}

	sellerID := string(response.Payload)
	walletID, err = getWalletID(stub, "businesscc", sellerID, "loan")
	if err != nil {
		return shim.Error("Repayment Business Loan_WalletID (seller) " + err.Error())
	}

	openBalance, err = getWalletValue(stub, walletID)
	if err != nil {
		return shim.Error("Repayment Business Loan_WalletValue (seller) " + err.Error())
	}
	openBalString = strconv.FormatInt(openBalance, 10)

	if (amt > loanChargesWalletValue+loanDisbursedWalletValue) || (amt == loanChargesWalletValue+loanDisbursedWalletValue) {
		dAmt = loanChargesWalletValue + loanDisbursedWalletValue
	} else if amt < loanChargesWalletValue+loanDisbursedWalletValue {
		dAmt = amt - loanChargesWalletValue
	}

	bal = openBalance - dAmt
	txnBalString = strconv.FormatInt(bal, 10)

	response = walletUpdation(stub, walletID, bal)
	if response.Status != shim.OK {
		return shim.Error("Repayment Business Loan_Wallet (seller) " + response.Message)
	}
	dAmtString = strconv.FormatInt(dAmt, 10)

	argsList = []string{"5rep", args[0], args[2], args[3], args[4], walletID, openBalString, args[1], args[5], cAmtString, dAmtString, txnBalString, args[8]}
	argsListStr = strings.Join(argsList, ",")
	txnResponse = putInTxnBal(stub, argsListStr)
	if txnResponse.Status != shim.OK {
		return shim.Error(txnResponse.Message)
	}

	//####################################################################################################################
	//Calling for updating Business Charges/Interest O/s Wallet
	//####################################################################################################################

	cAmtString = "0"
	walletID, err = getWalletID(stub, "businesscc", args[6], "interestOut")
	if err != nil {
		return shim.Error("Repayment Business Charges/Interest O/s WalletID " + err.Error())
	}

	openBalance, err = getWalletValue(stub, walletID)
	if err != nil {
		return shim.Error("Repayment Business Charges/Interest O/s WalletValue " + err.Error())
	}
	openBalString = strconv.FormatInt(openBalance, 10)
	dAmt = loanChargesWalletValue
	bal = openBalance - loanChargesWalletValue
	txnBalString = strconv.FormatInt(bal, 10)

	response = walletUpdation(stub, walletID, bal)
	if response.Status != shim.OK {
		return shim.Error("Repayment Business Charges/Interest O/s Wallet " + response.Message)
	}
	dAmtString = strconv.FormatInt(dAmt, 10)

	argsList = []string{"6rep", args[0], args[2], args[3], args[4], walletID, openBalString, args[1], args[5], cAmtString, dAmtString, txnBalString, args[8]}
	argsListStr = strings.Join(argsList, ",")
	txnResponse = putInTxnBal(stub, argsListStr)
	if txnResponse.Status != shim.OK {
		return shim.Error(txnResponse.Message)
	}

	//####################################################################################################################
	//Calling for updating Business Principal O/s Wallet
	//####################################################################################################################

	cAmtString = "0"
	walletID, err = getWalletID(stub, "businesscc", args[6], "principalOut")
	if err != nil {
		return shim.Error("Repayment Business Principal O/s WalletID " + err.Error())
	}

	openBalance, err = getWalletValue(stub, walletID)
	if err != nil {
		return shim.Error("Repayment Business Principal O/s WalletValue " + err.Error())
	}
	openBalString = strconv.FormatInt(openBalance, 10)

	bal = openBalance - loanDisbursedWalletValue
	txnBalString = strconv.FormatInt(bal, 10)

	response = walletUpdation(stub, walletID, bal)
	if response.Status != shim.OK {
		return shim.Error("Repayment Business Principal O/s Wallet " + response.Message)
	}
	dAmt = loanDisbursedWalletValue
	dAmtString = strconv.FormatInt(dAmt, 10)

	argsList = []string{"7rep", args[0], args[2], args[3], args[4], walletID, openBalString, args[1], args[5], cAmtString, dAmtString, txnBalString, args[8]}
	argsListStr = strings.Join(argsList, ",")
	txnResponse = putInTxnBal(stub, argsListStr)
	if txnResponse.Status != shim.OK {
		return shim.Error(txnResponse.Message)
	}

	//####################################################################################################################
	//Calling for updating Loan Charges Wallet
	//####################################################################################################################

	cAmtString = "0"
	walletID, err = getWalletID(stub, "loan", args[3], "charges")
	if err != nil {
		return shim.Error("Repayment Loan Charges WalletID " + err.Error())
	}

	openBalance, err = getWalletValue(stub, walletID)
	if err != nil {
		return shim.Error("Repayment Loan Charges WalletValue " + err.Error())
	}
	openBalString = strconv.FormatInt(openBalance, 10)

	bal = openBalance - loanChargesWalletValue
	txnBalString = strconv.FormatInt(bal, 10)

	response = walletUpdation(stub, walletID, bal)
	if response.Status != shim.OK {
		return shim.Error("Repayment Loan Charges Wallet " + response.Message)
	}
	dAmt = loanChargesWalletValue
	dAmtString = strconv.FormatInt(dAmt, 10)

	argsList = []string{"8rep", args[0], args[2], args[3], args[4], walletID, openBalString, args[1], args[5], cAmtString, dAmtString, txnBalString, args[8]}
	argsListStr = strings.Join(argsList, ",")
	txnResponse = putInTxnBal(stub, argsListStr)
	if txnResponse.Status != shim.OK {
		return shim.Error(txnResponse.Message)
	}

	//####################################################################################################################
	//Calling for updating Loan Disbursed Wallet
	//####################################################################################################################

	cAmtString = "0"
	walletID, err = getWalletID(stub, "loan", args[3], "disbursed")
	if err != nil {
		return shim.Error("Repayment Loan Disbursed WalletID " + err.Error())
	}

	openBalance, err = getWalletValue(stub, walletID)
	if err != nil {
		return shim.Error("Repayment Loan Disbursed WalletValue " + err.Error())
	}
	openBalString = strconv.FormatInt(openBalance, 10)

	if (amt > loanChargesWalletValue+loanDisbursedWalletValue) || (amt == loanChargesWalletValue+loanDisbursedWalletValue) {
		bal = 0
		dAmt = openBalance
		chaincodeArgs := toChaincodeArgs("updateLoanInfo", args[3], "repayment", "collected")
		response := stub.InvokeChaincode("loancc", chaincodeArgs, "myc")
		if response.Status != shim.OK {
			return shim.Error(response.Message)
		}
	} else if amt < loanChargesWalletValue+loanDisbursedWalletValue {
		dAmt = amt - loanChargesWalletValue
		bal = openBalance - dAmt
		chaincodeArgs := toChaincodeArgs("updateLoanInfo", args[3], "repayment", "part collected")
		response := stub.InvokeChaincode("loancc", chaincodeArgs, "myc")
		if response.Status != shim.OK {
			return shim.Error(response.Message)
		}
	}
	txnBalString = strconv.FormatInt(bal, 10)
	response = walletUpdation(stub, walletID, bal)
	if response.Status != shim.OK {
		return shim.Error("Repayment Loan Charges Wallet " + response.Message)
	}
	dAmtString = strconv.FormatInt(dAmt, 10)

	argsList = []string{"9rep", args[0], args[2], args[3], args[4], walletID, openBalString, args[1], args[5], cAmtString, dAmtString, txnBalString, args[8]}
	argsListStr = strings.Join(argsList, ",")
	txnResponse = putInTxnBal(stub, argsListStr)
	if txnResponse.Status != shim.OK {
		return shim.Error(txnResponse.Message)
	}

	//####################################################################################################################
	//Calling for updating Business Liability Wallet (Buyer)
	//####################################################################################################################

	cAmtString = "0"
	dAmtString = args[5]

	walletID, err = getWalletID(stub, "business", args[6], "liability")
	if err != nil {
		return shim.Error("Repayment Bank Liability WalletID " + err.Error())
	}

	openBalance, err = getWalletValue(stub, walletID)
	if err != nil {
		return shim.Error("Repayment Bank Liability WalletValue " + err.Error())
	}
	openBalString = strconv.FormatInt(openBalance, 10)

	bal = openBalance - amt
	txnBalString = strconv.FormatInt(bal, 10)

	response = walletUpdation(stub, walletID, bal)
	if response.Status != shim.OK {
		return shim.Error(response.Message)
	}

	argsList = []string{"10rep", args[0], args[2], args[3], args[4], walletID, openBalString, args[1], args[5], cAmtString, dAmtString, txnBalString, args[8]}
	argsListStr = strings.Join(argsList, ",")
	txnResponse = putInTxnBal(stub, argsListStr)
	if txnResponse.Status != shim.OK {
		return shim.Error(txnResponse.Message)
	}

	//####################################################################################################################

	return shim.Success(nil)
}

func putInTxnBal(stub shim.ChaincodeStubInterface, argsListStr string) pb.Response {

	chaincodeArgs := toChaincodeArgs("putTxnInfo", argsListStr)
	fmt.Println("calling the txnbalcc chaincode from repayment")
	response := stub.InvokeChaincode("txnbalcc", chaincodeArgs, "myc")
	if response.Status != shim.OK {
		return shim.Error(response.Message)
	}
	fmt.Println(string(response.Payload))
	return shim.Success(nil)
}

func getWalletID(stub shim.ChaincodeStubInterface, ccName string, id string, walletType string) (string, error) {

	// STEP-1
	// using FromID, get a walletID from bank structure

	chaincodeArgs := toChaincodeArgs("getWalletID", id, walletType)
	response := stub.InvokeChaincode(ccName, chaincodeArgs, "myc")
	if response.Status != shim.OK {
		return "0", errors.New(response.Message)
	}
	walletID := string(response.GetPayload())
	return walletID, nil

}

func getWalletValue(stub shim.ChaincodeStubInterface, walletID string) (int64, error) {

	walletArgs := toChaincodeArgs("getWallet", walletID)
	walletResponse := stub.InvokeChaincode("walletcc", walletArgs, "myc")
	if walletResponse.Status != shim.OK {
		return 0, errors.New(walletResponse.Message)
	}
	balString := string(walletResponse.Payload)
	balance, _ := strconv.ParseInt(balString, 10, 64)
	return balance, nil
}

func walletUpdation(stub shim.ChaincodeStubInterface, walletID string, amt int64) pb.Response {

	txnBalString := strconv.FormatInt(amt, 10)
	walletArgs := toChaincodeArgs("updateWallet", walletID, txnBalString)
	walletResponse := stub.InvokeChaincode("walletcc", walletArgs, "myc")
	if walletResponse.Status != shim.OK {
		return shim.Error(walletResponse.Message)
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
		fmt.Println("Unable to start Repayment chaincode:", err)
	}
}
