package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type chainCode struct {
}

type transactionInfo struct {
	TxnType string
	TxnDate string
	LoanID  string
	InsID   string
	Amt     int64
	FromID  string
	ToID    string
	By      string
	PprID   string
}

func (c *chainCode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (c *chainCode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()

	if function == "newTxnInfo" {
		return newTxnInfo(stub, args)
	} else if function == "getTxnInfo" {
		return getTxnInfo(stub, args)
	}
	return shim.Success(nil)
}

func newTxnInfo(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 10 {
		return shim.Error("Invalid number of arguments in txn")
	}

	tTypeValues := map[string]bool{
		"disbursement": true,
		"collection":   true,
		"refund":       true,
	}

	//Converting into lower case for comparison
	tTypeLower := strings.ToLower(args[1])
	if !tTypeValues[tTypeLower] {
		return shim.Error("Invalid transaction type " + args[1])
	}

	//TxnDate -> tDate
	/*tDate, err := time.Parse("02/01/2006", args[2])
	if err != nil {
		return shim.Error(err.Error())
	}*/

	amt, err := strconv.ParseInt(args[5], 10, 64)
	if err != nil {
		return shim.Error(err.Error())
	}

	//TODO: put it at last for redability
	transaction := transactionInfo{tTypeLower, args[2], args[3], args[4], amt, args[6], args[7], args[8], args[9]}
	fmt.Println("transaction:", transaction)
	txnBytes, err := json.Marshal(transaction)
	err = stub.PutState(args[0], txnBytes)
	if err != nil {
		return shim.Error("in put state " + err.Error())
	}
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// 													UPDATING WALLETS																///
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
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

	// STEP-1
	// using FromID, get a walletID from bank structure
	// bankID = bankID
	/*	bankID := args[6] // of Bank
		//bank, err := stub.getState(bankID)
		//bankFcn := "getWalletID"
		chaincodeBankArgs := util.ToChaincodeArgs("getWalletID", bankID, "main")
		bankResponse := stub.InvokeChaincode("bankcc", chaincodeBankArgs, "myc")
		if bankResponse.Status != shim.OK {
			return shim.Error(bankResponse.Message)
		}
		walletID := string(bankResponse.Payload)
		fmt.Println("Bank Main walletID : ", walletID)

		// STEP-2
		// getting Balance from walletID
		// walletFcn := "getWallet"
		walletArgs := util.ToChaincodeArgs("getWallet", walletID)
		walletResponse := stub.InvokeChaincode("walletcc", walletArgs, "myc")
		if walletResponse.Status != shim.OK {
			return shim.Error(walletResponse.Message)
		}
		openBalString := string(walletResponse.Payload)
		fmt.Printf("Open balance of %s : %s\n", walletID, openBalString)
		openBal, err := strconv.ParseInt(openBalString, 10, 64)
		if err != nil {
			return shim.Error("Error in converting the bank main wallet balance")
		}
		cAmt := int64(0)
		cAmtString := strconv.FormatInt(cAmt, 10)
		dAmt := amt
		dAmtString := strconv.FormatInt(dAmt, 10)
		txnBal := openBal - dAmt + cAmt
		fmt.Printf("Txn bal for bank main wallet : %d\n", txnBal)
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
		fmt.Println("Sneding to transaction balance")
		argsList := []string{"1", args[0], args[2], args[3], args[4], walletID, openBalString, args[1], args[5], cAmtString, dAmtString, txnBalString, args[8]}
		argsListStr := strings.Join(argsList, ",")
		fmt.Printf("Args list : %s \n", argsListStr)
		chaincodeArgs := util.ToChaincodeArgs("putTxnInfo", argsListStr)
		fmt.Println("calling the other chaincode")
		response := stub.InvokeChaincode("txnbalcc", chaincodeArgs, "myc")
		fmt.Println(string(response.Payload))
		if response.Status != shim.OK {
			return shim.Error(response.Message)
		}

		fmt.Println("successfully updated Bank's main wallet and written the txn thing to the ledger")
		//+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

		//#####################################################################################################################
		//Calling for updating Business Main_Wallet
		//####################################################################################################################

		// STEP-1
		// using businessID, get a walletID from business structure
		// businessID = busienssID
		businessID := args[7] // of Business

		chaincodeBissArgs := util.ToChaincodeArgs("getWalletID", businessID, "main")
		bissResponse := stub.InvokeChaincode("businesscc", chaincodeBissArgs, "myc")
		if bissResponse.Status != shim.OK {
			return shim.Error(bissResponse.Message)
		}
		walletID = string(bissResponse.Payload)
		fmt.Println("Business Main walletID : ", walletID)
		// STEP-2
		// getting Balance from walletID
		// walletFcn := "getWallet"
		walletArgs = util.ToChaincodeArgs("getWallet", walletID)
		walletResponse = stub.InvokeChaincode("walletcc", walletArgs, "myc")
		if walletResponse.Status != shim.OK {
			return shim.Error(walletResponse.Message)
		}
		openBalString = string(walletResponse.Payload)
		fmt.Printf("Open balance of %s : %s\n", walletID, openBalString)
		openBal, err = strconv.ParseInt(openBalString, 10, 64)
		if err != nil {
			return shim.Error("Error in converting business main wallet the balance")
		}
		cAmt = amt
		cAmtString = strconv.FormatInt(cAmt, 10)
		dAmt = 0
		dAmtString = strconv.FormatInt(dAmt, 10)
		txnBal = openBal - dAmt + cAmt
		fmt.Printf("Txn bal for business main wallet : %d\n", txnBal)
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
		argsList = []string{"2", args[0], args[2], args[3], args[4], walletID, openBalString, args[1], args[5], cAmtString, dAmtString, txnBalString, args[8]}
		argsListStr = strings.Join(argsList, ",")
		chaincodeArgs = util.ToChaincodeArgs("putTxnInfo", argsListStr)
		fmt.Println("calling the other chaincode")
		response = stub.InvokeChaincode("txnbalcc", chaincodeArgs, "myc")
		fmt.Println(string(response.Payload))
		if response.Status != shim.OK {
			return shim.Error(response.Message)
		}

		fmt.Println("successfully updated Business main wallet and written the txn thing to the ledger")
		//+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

		//####################################################################################################################
		//Calling for updating Business Loan_Wallet
		//####################################################################################################################

		// STEP-1
		// using businessID, get a walletID from biss structure
		// bankID = businessID
		businessID = args[7] // of business
		chaincodeBissArgs = util.ToChaincodeArgs("getWalletID", businessID, "loan")
		bissResponse = stub.InvokeChaincode("businesscc", chaincodeBissArgs, "myc")
		if bissResponse.Status != shim.OK {
			return shim.Error(bissResponse.Message)
		}
		walletID = string(bissResponse.Payload)
		fmt.Println("Business Loan walletID : ", walletID)
		// STEP-2
		// getting Balance from walletID
		// walletFcn := "getWallet"
		walletArgs = util.ToChaincodeArgs("getWallet", walletID)
		walletResponse = stub.InvokeChaincode("walletcc", walletArgs, "myc")
		if walletResponse.Status != shim.OK {
			return shim.Error(walletResponse.Message)
		}
		openBalString = string(walletResponse.Payload)
		fmt.Printf("Open balance of %s : %s\n", walletID, openBalString)
		openBal, err = strconv.ParseInt(openBalString, 10, 64)
		if err != nil {
			return shim.Error("Error in converting business loan wallet the balance")
		}
		cAmt = amt
		cAmtString = strconv.FormatInt(cAmt, 10)
		dAmt = 0
		dAmtString = strconv.FormatInt(dAmt, 10)
		txnBal = openBal - dAmt + cAmt
		fmt.Printf("Txn bal for business loan wallet : %d\n", txnBal)
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
		argsList = []string{"3", args[0], args[2], args[3], args[4], walletID, openBalString, args[1], args[5], cAmtString, dAmtString, txnBalString, args[8]}
		argsListStr = strings.Join(argsList, ",")
		chaincodeArgs = util.ToChaincodeArgs("putTxnInfo", argsListStr)
		fmt.Println("calling the other chaincode")
		response = stub.InvokeChaincode("txnbalcc", chaincodeArgs, "myc")
		fmt.Println(string(response.Payload))
		if response.Status != shim.OK {
			return shim.Error(response.Message)
		}
		fmt.Println("successfully updated Business loan wallet and written the txn thing to the ledger")
		//+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

		//####################################################################################################################
		//Calling for updating Bank Asset_Wallet
		//####################################################################################################################

		// STEP-1
		// using FromID, get a walletID from bank structure
		// bankID = bankID
		bankID = args[6] // of Bank
		//bank, err := stub.getState(bankID)
		//bankFcn := "getWalletID"
		chaincodeBankArgs = util.ToChaincodeArgs("getWalletID", bankID, "asset")
		bankResponse = stub.InvokeChaincode("bankcc", chaincodeBankArgs, "myc")
		if bankResponse.Status != shim.OK {
			return shim.Error(bankResponse.Message)
		}
		walletID = string(bankResponse.Payload)
		fmt.Println("Bank Asset walletID : ", walletID)
		// STEP-2
		// getting Balance from walletID
		// walletFcn := "getWallet"
		walletArgs = util.ToChaincodeArgs("getWallet", walletID)
		walletResponse = stub.InvokeChaincode("walletcc", walletArgs, "myc")
		if walletResponse.Status != shim.OK {
			return shim.Error(walletResponse.Message)
		}
		openBalString = string(walletResponse.Payload)
		fmt.Printf("Open balance of %s : %s\n", walletID, openBalString)
		openBal, err = strconv.ParseInt(openBalString, 10, 64)
		if err != nil {
			return shim.Error("Error in converting bank asset wallet the balance")
		}
		cAmt = amt
		cAmtString = strconv.FormatInt(cAmt, 10)
		dAmt = 0
		dAmtString = strconv.FormatInt(dAmt, 10)
		txnBal = openBal - dAmt + cAmt
		fmt.Printf("Txn bal for bank asset wallet : %d\n", txnBal)
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
		argsList = []string{"4", args[0], args[2], args[3], args[4], walletID, openBalString, args[1], args[5], cAmtString, dAmtString, txnBalString, args[8]}
		argsListStr = strings.Join(argsList, ",")
		chaincodeArgs = util.ToChaincodeArgs("putTxnInfo", argsListStr)
		fmt.Println("calling the other chaincode")
		response = stub.InvokeChaincode("txnbalcc", chaincodeArgs, "myc")
		if response.Status != shim.OK {
			return shim.Error(response.Message)
		}
		fmt.Println("successfully updated Bank Asset wallet and written the txn thing to the ledger")
		//+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

		//return on successful transaction updation*/
	return shim.Success(nil)
}

func getTxnInfo(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Invalid number of arguments")
	}

	txnBytes, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error(err.Error())
	} else if txnBytes == nil {
		return shim.Error("No data exists on this loanID: " + args[0])
	}

	transaction := transactionInfo{}
	fmt.Println(transaction)
	err = json.Unmarshal(txnBytes, &transaction)
	if err != nil {
		return shim.Error(err.Error())
	}

	tString := fmt.Sprintf("%+v", transaction)
	return shim.Success([]byte(tString))

}

func main() {
	err := shim.Start(new(chainCode))
	if err != nil {
		fmt.Println("Unable to start the chaincode")
	}
}
