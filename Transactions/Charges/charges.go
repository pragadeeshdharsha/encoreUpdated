package main

import (
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
	if function == "putTxnBalInfo" { //Inserting a New Business information
		return putTxnBalInfo(stub, args)
	} else {
		return shim.Error("incorrect function")
	}
}

func putTxnBalInfo(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	/*
	 * arg[0]	:	Key
	 * arg[1]	:	txnID	(0)
	 * arg[2]	:	date					//given	args[0]
	 * arg[3]	:	LoanID					//given	args[1]
	 * arg[4]	:	insID					//given args[2]
	 * arg[5]	:	bankID	=>	walletID	//given args[3]
	 *			:	bissID	=>	walletID	//given args[4]
	 * arg[6]	:	openBal
	 * arg[7]	:	txnType : charges		//given args[7]
	 * arg[8]	:	amt						//given args[5]
	 * arg[9]	:	cAmt (calc)
	 * arg[10]	:	dAmt (calc)
	 * arg[11]	:	txnBal
	 * arg[12]	:	by						//given args[6]
	 */

	if len(args) != 8 {
		return shim.Error("incorrcect number of arguments")
	}

	///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// 													UPDATING WALLETS																///
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	/*
	 *	business load wallet increased
	 * 	bank revenue wallet incresed
	 *	bank asset wallet increased
	 */

	//####################################################################################################################
	//Calling for updating Business_loan_wallet
	//####################################################################################################################

	// STEP-1
	// using bankID, get a walletID from bank structure
	// toID = bissID
	bissID := args[4] // of Biss
	//bank, err := stub.getState(bankID)
	//bankFcn := "getWalletID"
	chaincodeArgs := util.ToChaincodeArgs("getWalletID", bissID, "loan")
	response := stub.InvokeChaincode("businesscc", chaincodeArgs, "myc")
	if response.Status != shim.OK {
		return shim.Error(response.Message)
	}
	walletID := string(response.Payload)

	// STEP-2
	// getting Balance from walletID
	// walletFcn := "getWallet"
	walletArgs := util.ToChaincodeArgs("getWallet", walletID)
	walletResponse := stub.InvokeChaincode("walletcc", walletArgs, "myc")
	if walletResponse.Status != shim.OK {
		return shim.Error(walletResponse.Message)
	}
	openBalString := string(walletResponse.Payload)
	openBal, err := strconv.ParseInt(openBalString, 10, 64)
	if err != nil {
		return shim.Error("Error in converting the balance")
	}

	cAmtString := args[5]
	cAmt, _ := strconv.ParseInt(cAmtString, 10, 64)
	dAmtString := "0"
	dAmt, _ := strconv.ParseInt(dAmtString, 10, 64)
	txnBal := openBal - dAmt + cAmt
	txnBalString := strconv.FormatInt(txnBal, 10)

	// STEP-3
	// update wallet of ID walletID here, and write it to the wallet_ledger
	// walletFcn := "updateWallet"
	walletArgs = util.ToChaincodeArgs("updateWallet", walletID, txnBalString)
	walletResponse = stub.InvokeChaincode("walletcc", walletArgs, "myc")
	if walletResponse.Status != shim.OK {
		return shim.Error(walletResponse.Message)
	}

	// STEP-4 generate txn_balance_object and write it to the Txn_Bal_Ledger
	argsList := []string{"1", "0", args[0], args[1], args[2], walletID, openBalString, args[7], args[5], cAmtString, dAmtString, txnBalString, args[6]}
	argsListStr := strings.Join(argsList, ",")
	chaincodeArgs = util.ToChaincodeArgs("putTxnInfo", argsListStr)
	fmt.Println("calling the other chaincode")
	response = stub.InvokeChaincode("txnbalcc", chaincodeArgs, "myc")
	if response.Status != shim.OK {
		return shim.Error(response.Message)
	}
	fmt.Println(response.GetPayload())
	//successfully updated Bank's main wallet and written the txn thing to the ledger
	//+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

	//####################################################################################################################
	//Calling for updating Bank Revenue_Wallet
	//####################################################################################################################

	// STEP-1
	// using bankID, get a walletID from bank structure
	// bankID = bankID
	bankID := args[3] // of Bank
	//bank, err := stub.getState(bankID)
	//bankFcn := "getWalletID"
	chaincodeArgs = util.ToChaincodeArgs("getWalletID", bankID, "main")
	response = stub.InvokeChaincode("bankcc", chaincodeArgs, "myc")
	if response.Status != shim.OK {
		return shim.Error(response.Message)
	}
	walletID = string(response.Payload[:])

	// STEP-2
	// getting Balance from walletID
	// walletFcn := "getWallet"
	walletArgs = util.ToChaincodeArgs("getWallet", walletID)
	walletResponse = stub.InvokeChaincode("walletcc", walletArgs, "myc")
	if walletResponse.Status != shim.OK {
		return shim.Error(walletResponse.Message)
	}
	openBalString = string(walletResponse.Payload[:])
	openBal, err = strconv.ParseInt(openBalString, 10, 64)
	if err != nil {
		return shim.Error("Error in converting the balance")
	}

	cAmtString = args[4]
	cAmt, _ = strconv.ParseInt(cAmtString, 10, 64)
	dAmtString = "0"
	dAmt, _ = strconv.ParseInt(dAmtString, 10, 64)
	txnBal = openBal - dAmt + cAmt
	txnBalString = strconv.FormatInt(txnBal, 10)

	// STEP-3
	// update wallet of ID walletID here, and write it to the wallet_ledger
	// walletFcn := "updateWallet"
	walletArgs = util.ToChaincodeArgs("updateWallet", walletID, txnBalString)
	walletResponse = stub.InvokeChaincode("walletcc", walletArgs, "myc")
	if walletResponse.Status != shim.OK {
		return shim.Error(walletResponse.Message)
	}

	// STEP-4 generate txn_balance_object and write it to the Txn_Bal_Ledger
	argsList = []string{"2", "0", args[0], args[1], args[2], walletID, openBalString, args[7], args[4], cAmtString, dAmtString, txnBalString, args[5]}
	argsListStr = strings.Join(argsList, ",")
	chaincodeArgs = util.ToChaincodeArgs("putTxnInfo", argsListStr)
	fmt.Println("calling the other chaincode")
	response = stub.InvokeChaincode("txnbalcc", chaincodeArgs, "myc")
	if response.Status != shim.OK {
		return shim.Error(response.Message)
	}
	fmt.Println(response.GetPayload())
	//successfully updated Bank's main wallet and written the txn thing to the ledger
	//+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

	//####################################################################################################################
	//Calling for updating Bank Asset_Wallet
	//####################################################################################################################

	// STEP-1
	// using bankID, get a walletID from bank structure
	// bankID = bankID
	bankID = args[3] // of Bank
	//bank, err := stub.getState(bankID)
	//bankFcn := "getWalletID"
	chaincodeArgs = util.ToChaincodeArgs("getWalletID", bankID, "main")
	response = stub.InvokeChaincode("bankcc", chaincodeArgs, "myc")
	if response.Status != shim.OK {
		return shim.Error(response.Message)
	}
	walletID = string(response.Payload[:])

	// STEP-2
	// getting Balance from walletID
	// walletFcn := "getWallet"
	walletArgs = util.ToChaincodeArgs("getWallet", walletID)
	walletResponse = stub.InvokeChaincode("walletcc", walletArgs, "myc")
	if walletResponse.Status != shim.OK {
		return shim.Error(walletResponse.Message)
	}
	openBalString = string(walletResponse.Payload[:])
	openBal, err = strconv.ParseInt(openBalString, 10, 64)
	if err != nil {
		return shim.Error("Error in converting the balance")
	}

	cAmtString = args[4]
	cAmt, _ = strconv.ParseInt(cAmtString, 10, 64)
	dAmtString = "0"
	dAmt, _ = strconv.ParseInt(dAmtString, 10, 64)
	txnBal = openBal - dAmt + cAmt
	txnBalString = strconv.FormatInt(txnBal, 10)

	// STEP-3
	// update wallet of ID walletID here, and write it to the wallet_ledger
	// walletFcn := "updateWallet"
	walletArgs = util.ToChaincodeArgs("updateWallet", walletID, txnBalString)
	walletResponse = stub.InvokeChaincode("walletcc", walletArgs, "myc")
	if walletResponse.Status != shim.OK {
		return shim.Error(walletResponse.Message)
	}

	// STEP-4 generate txn_balance_object and write it to the Txn_Bal_Ledger
	argsList = []string{"3", "0", args[0], args[1], args[2], walletID, openBalString, args[7], args[4], cAmtString, dAmtString, txnBalString, args[5]}
	argsListStr = strings.Join(argsList, ",")
	chaincodeArgs = util.ToChaincodeArgs("putTxnInfo", argsListStr)
	fmt.Println("calling the other chaincode")
	response = stub.InvokeChaincode("txnbalcc", chaincodeArgs, "myc")
	if response.Status != shim.OK {
		return shim.Error(response.Message)
	}
	fmt.Println(response.GetPayload())
	//successfully updated Bank's main wallet and written the txn thing to the ledger
	//+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

	return shim.Success(nil)
}

func main() {
	err := shim.Start(new(chainCode))
	if err != nil {
		fmt.Println("Unable to start the chaincode")
	}
}
