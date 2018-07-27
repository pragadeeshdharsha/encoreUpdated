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

/*
businessWalletId
businessLoanWalletId
businessLiabilityWalletId

bankWalletId
bankAssetWalletId
bankChargesWalletId
bankLiabilityWalletId
TDSreceivableWalletId
*/

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

func disbursement(stub shim.ChaincodeStubInterface, args []string) {
	updateBankWallet(stub, args)
	updateBusinessWallet(stub, args, "0")
	updateBankAssetWallet(stub, args)
	updateBusinessLoanWallet(stub, args)
}

func updateBusinessWallet(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	/*cAmtString := args[5]
	dAmtString := "0"*/

	walletID, openBalString, stub, argstxnBalString, err := getWalletInfo(stub, args[7], "main", "businesscc", cAmtString, dAmtString)
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
	return shim.Success([]byte("Business wallet Updated"))

}

func updateBusinessLoanWallet(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	cAmtString := args[5]
	dAmtString := "0"

	walletID, openBalString, txnBalString, err := getWalletInfo(stub, args[7], "loan", "businesscc", cAmtString, dAmtString)
	if err != nil {
		return shim.Error(err.Error())
	}

	// STEP-4 generate txn_balance_object and write it to the Txn_Bal_Ledger
	argsList := []string{"3", args[0], args[2], args[3], args[4], walletID, openBalString, args[1], args[5], cAmtString, dAmtString, txnBalString, args[8]}
	argsListStr := strings.Join(argsList, ",")
	chaincodeArgs := util.ToChaincodeArgs("putTxnInfo", argsListStr)
	fmt.Println("calling the other chaincode")
	response := stub.InvokeChaincode("txnbalcc", chaincodeArgs, "myc")
	if response.Status != shim.OK {
		return shim.Error(response.Message)
	}
	fmt.Println(string(response.GetPayload()))
	return shim.Success([]byte("Business Loan wallet updated Successfully"))
}

func updateBusinessLiabilityWallet(stub shim.ChaincodeStubInterface, args []string) pb.Response {

}

func updateBankWallet(stub shim.ChaincodeStubInterface, args []string) pb.Response {

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
	return shim.Success([]byte("Bank Wallet Updated Successfully"))

}

func updateBankAssetWallet(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	cAmtString = args[5]
	dAmtString = "0"

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

}

func updateBankChargesWallet(stub shim.ChaincodeStubInterface, args []string) pb.Response {

}

func updateBankLiabilityWallet(stub shim.ChaincodeStubInterface, args []string) pb.Response {

}

func updateTDSreceivableWallet(stub shim.ChaincodeStubInterface, args []string) pb.Response {

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
