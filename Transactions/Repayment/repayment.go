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

	amt, _ := strconv.ParseInt(args[5], 10, 64)

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

	walletID, err := getWalletID(stub, "businesscc", args[6], "main")
	if err != nil {
		return shim.Error("Repayment Business Main WalletID " + err.Error())
	}

	openBalance, err := getWalletValue(stub, walletID)
	if err != nil {
		return shim.Error("Repayment Business Main WalletValue " + err.Error())
	}

	bal := openBalance - amt

	response := walletUpdation(stub, walletID, bal)
	if response.Status != shim.OK {
		return shim.Error(response.Message)
	}

	// STEP-4 generate txn_balance_object and write it to the Txn_Bal_Ledger
	/*
		argsList := []string{"5", args[0], args[2], args[3], args[4], walletID, openBalString, args[1], args[5], cAmtString, dAmtString, txnBalString, args[8]}
		argsListStr := strings.Join(argsList, ",")
		chaincodeArgs := toChaincodeArgs("putTxnInfo", argsListStr)
		fmt.Println("calling the other chaincode business main")
		response := stub.InvokeChaincode("txnbalcc", chaincodeArgs, "myc")
		if response.Status != shim.OK {
			return shim.Error(response.Message)
		}
		fmt.Println(string(response.GetPayload()))
	*/

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

	//amt, _ = strconv.ParseInt(args[5], 10, 64)

	bal = openBalance + amt

	response = walletUpdation(stub, walletID, bal)
	if response.Status != shim.OK {
		return shim.Error(response.Message)
	}

	/*
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
	*/

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

	bal = openBalance - loanChargesWalletValue - loanDisbursedWalletValue

	response = walletUpdation(stub, walletID, bal)
	if response.Status != shim.OK {
		return shim.Error("Repayment Bank Asset Wallet " + response.Message)
	}
	dAmtString = strconv.FormatInt(bal, 10)

	//####################################################################################################################
	//Calling for updating Bank Refund_Wallet
	//####################################################################################################################

	//####################################################################################################################
	//4.Calling for updating Business Loan_Wallet (seller)
	//####################################################################################################################

	// geting seller's ID using instrument no

	cAmtString = "0"
	chaincodeArgs := toChaincodeArgs("getSellerIDnAmt", args[4])
	response = stub.InvokeChaincode("instrumentcc", chaincodeArgs, "myc")
	if response.Status != shim.OK {
		return shim.Error("Instrument refrence no " + args[4] + " does not exits")
	}

	sellerID := strings.Split(string(response.Payload), ",")[0]

	walletID, err = getWalletID(stub, "businesscc", sellerID, "loan")
	if err != nil {
		return shim.Error("Repayment Business Loan_WalletID (seller) " + err.Error())
	}

	openBalance, err = getWalletValue(stub, walletID)
	if err != nil {
		return shim.Error("Repayment Business Loan_WalletValue (seller) " + err.Error())
	}

	var calAmt int64

	if (amt > loanChargesWalletValue+loanDisbursedWalletValue) || (amt == loanChargesWalletValue+loanDisbursedWalletValue) {
		calAmt = loanChargesWalletValue + loanDisbursedWalletValue
	} else if amt < loanChargesWalletValue+loanDisbursedWalletValue {
		calAmt = amt - loanChargesWalletValue
	}

	bal = openBalance - calAmt
	dAmtString = strconv.FormatInt(bal, 10)

	response = walletUpdation(stub, walletID, bal)
	if response.Status != shim.OK {
		return shim.Error("Repayment Business Loan_Wallet (seller) " + response.Message)
	}

	//####################################################################################################################
	//Calling for updating Business Charges O/s Wallet
	//####################################################################################################################

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

	bal = openBalance - loanDisbursedWalletValue

	response = walletUpdation(stub, walletID, bal)
	if response.Status != shim.OK {
		return shim.Error("Repayment Business Principal O/s Wallet " + response.Message)
	}
	dAmtString = strconv.FormatInt(bal, 10)

	//####################################################################################################################
	//Calling for updating Loan Charges Wallet
	//####################################################################################################################

	return shim.Success(nil)
}

/*
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

*/

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
		fmt.Println("Unable to start the chaincode:", err)
	}
}
