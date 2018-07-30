package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type chainCode struct {
}

type instrumentInfo struct {
	InstrumentRefNo string
	InstrumenDate   time.Time
	SellBusinessID  string
	BuyBusinsessID  string
	InsAmount       string // use int64 for convertion
	InsStatus       string // not required
	InsDueDate      time.Time
	ProgramID       string
	PPRid           string
	UploadBatchNo   string
	ValueDate       time.Time
}

func (c *chainCode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	indexName := "InstrumentRefNo~SellBusinessID~InsAmount"
	inst := instrumentInfo{}

	refNoSellIDkey, err := stub.CreateCompositeKey(indexName, []string{inst.InstrumentRefNo, inst.SellBusinessID, inst.InsAmount})
	if err != nil {
		return shim.Error("Composite key InstrumentRefNo~SellBusinessID can not be created (instrument)")
	}
	value := []byte{0x00}
	stub.PutState(refNoSellIDkey, value)
	return shim.Success(nil)
}

func (c *chainCode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	if function == "enterInstrument" {
		return enterInstrument(stub, args)
	} else if function == "getInstrument" {
		return getInstrument(stub, args)
	} else if function == "getSellerIDnAmt" {
		return getSellerIDnAmt(stub, args)
	}

	return shim.Error("No function named " + function + " in Instrument")

}

func enterInstrument(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 10 {
		xLenStr := strconv.Itoa(len(args))
		return shim.Error("Invalid number of arguments in enterInstrument (required:10) given:" + xLenStr)

	}

	// Checking existence of Instrument Reference No. – Supplier ID pair
	refNoSellIDiterator, _ := stub.GetStateByPartialCompositeKey("InstrumentRefNo~SellBusinessID", []string{args[0]})
	refNoSellIDdata, _ := refNoSellIDiterator.Next()
	_, refNoSellIDvalues, err := stub.SplitCompositeKey(refNoSellIDdata.Key)
	if err != nil {
		return shim.Error("Unable to split composite key InstrumentRefNo~SellBusinessID")
	}
	if refNoSellIDvalues[1] == args[0] {
		return shim.Error("Instrument Reference No. – Supplier ID pair already exists")
	}

	// Hashing for key to store in ledger
	hash := sha256.New()
	instID := args[0] + args[2]
	hash.Write([]byte(instID))
	md := hash.Sum(nil)
	instIDsha := hex.EncodeToString(md)

	//Checking existence of ProgramID
	chaincodeArgs := toChaincodeArgs("programIDexists", args[7])
	response := stub.InvokeChaincode("programcc", chaincodeArgs, "myc")
	if response.Status == shim.OK {
		return shim.Error("ProgramId " + args[1] + " does not exits")
	}

	//Checking existence of pprID
	chaincodeArgs = toChaincodeArgs("pprIDexists", args[8])
	response = stub.InvokeChaincode("pprcc", chaincodeArgs, "myc")
	if response.Status == shim.OK {
		return shim.Error("PprId " + args[8] + " does not exits")
	}

	//Checking existence of SellerBusinessID
	chaincodeArgs = toChaincodeArgs("busIDexists", args[2])
	response = stub.InvokeChaincode("businesscc", chaincodeArgs, "myc")
	if response.Status == shim.OK {
		return shim.Error("BusinessId " + args[2] + " does not exits")
	}

	//Checking existence of BuyerBusinessID
	chaincodeArgs = toChaincodeArgs("busIDexists", args[3])
	response = stub.InvokeChaincode("businesscc", chaincodeArgs, "myc")
	if response.Status == shim.OK {
		return shim.Error("BusinessId " + args[3] + " does not exits")
	}

	//InstrumentDate -> instDate
	instDate, err := time.Parse("02/01/2006", args[1])
	if err != nil {
		return shim.Error(err.Error())
	}

	_, err = strconv.ParseInt(args[4], 10, 64)
	if err != nil {
		return shim.Error(err.Error())
	}

	/*		insStatusValues := map[string]bool{
					"open":              true,
					"sanctioned":        true,
					"part disbursed":    true,
					"disbursed":         true,
					"part collected":    true,
					"collected/settled": true,
					"overdue":           true,
				}


			insStatusValuesLower := strings.ToLower(args[5])
			if !insStatusValues[insStatusValuesLower] {
				return shim.Error("Invalid Instrument Status " + args[5])
			}
	*/

	//InsDueDate -> insDate
	insDueDate, err := time.Parse("02/01/2006", args[6])
	if err != nil {
		return shim.Error(err.Error())
	}

	//Converting the incoming date from Dd/mm/yy:hh:mm:ss to Dd/mm/yyThh:mm:ss for parsing
	vString := args[9][:10] + "T" + args[9][11:] //removing the ":" part from the string

	//ValueDate -> vDate
	vDate, err := time.Parse("02/01/2006T15:04:05", vString)
	if err != nil {
		return shim.Error("error in parsing the date and time (instrument)" + err.Error())
	}

	inst := instrumentInfo{args[0], instDate, args[2], args[3], args[4], "open", insDueDate, args[7], args[8], args[9], vDate}
	instBytes, err := json.Marshal(inst)
	if err != nil {
		return shim.Error(err.Error())
	}
	stub.PutState(instIDsha, instBytes)
	return shim.Success(nil)
}

func toChaincodeArgs(args ...string) [][]byte {
	bargs := make([][]byte, len(args))
	for i, arg := range args {
		bargs[i] = []byte(arg)
	}
	return bargs
}

func getInstrument(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		xLenStr := strconv.Itoa(len(args))
		return shim.Error("Invalid number of arguments in getInstrument (required:1) given:" + xLenStr)

	}

	insBytes, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error(err.Error())
	} else if insBytes == nil {
		return shim.Error("No data exists on this InstrumentID: " + args[0])
	}

	ins := instrumentInfo{}
	err = json.Unmarshal(insBytes, &ins)
	insString := fmt.Sprintf("%+v", ins)
	return shim.Success([]byte(insString))
}

func getSellerIDnAmt(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		xLenStr := strconv.Itoa(len(args))
		return shim.Error("Invalid number of arguments in getSellerID(instrument) (required:1) given: " + xLenStr)
	}

	instID := args[0]
	instRefNoSellIDiterator, err := stub.GetStateByPartialCompositeKey("InstrumentRefNo~SellBusinessID", []string{instID})
	if err != nil {
		return shim.Error("Unable to get the result for composite key : InstrumentRefNo~SellBusinessID")
	}
	var sellBusID string
	instRefNoSellIData, err := instRefNoSellIDiterator.Next()
	if err != nil {
		return shim.Error("Unable to iterate instRefNoSellIDiterator:" + err.Error())
	}
	_, requiredArgs, err := stub.SplitCompositeKey(instRefNoSellIData.Key)
	if err != nil {
		return shim.Error("error spliting the composite key instRefNoSellIDiterator:" + err.Error())
	}
	sellBusID = requiredArgs[1] + "," + requiredArgs[2]

	return shim.Success([]byte(sellBusID))
}

func main() {
	err := shim.Start(new(chainCode))
	if err != nil {
		fmt.Printf("Error starting Instrument chaincode: %s\n", err)
	}
}
