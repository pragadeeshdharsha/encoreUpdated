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

type instrumentInfo struct {
	InstrumentRefNo string
	InstrumenDate   time.Time
	SellBusinessID  string
	BuyBusinsessID  string
	InsAmount       int64
	InsStatus       string
	InsDueDate      time.Time
	ProgramID       string
	UploadBatchNo   string
	ValueDate       time.Time
}

func (c *chainCode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (c *chainCode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	if function == "enterInstrument" {
		return enterInstrument(stub, args)
	} else if function == "getInstrument" {
		return getInstrument(stub, args)
	} else if function == "getSellerID" {
		return getSellerID(stub, args)
	}

	return shim.Error("No function named " + function + " in Instrument")

}

func enterInstrument(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 11 {
		xLenStr := strconv.Itoa(len(args))
		return shim.Error("Invalid number of arguments in enterInstrument (required:11) given:" + xLenStr)

	}

	//InstrumentDate -> instDate
	instDate, err := time.Parse("02/01/2006", args[2])
	if err != nil {
		return shim.Error(err.Error())
	}

	insAmt, err := strconv.ParseInt(args[5], 10, 64)
	if err != nil {
		return shim.Error(err.Error())
	}

	insStatusValues := map[string]bool{
		"open":              true,
		"sanctioned":        true,
		"part disbursed":    true,
		"disbursed":         true,
		"part collected":    true,
		"collected/settled": true,
		"overdue":           true,
	}

	insStatusValuesLower := strings.ToLower(args[6])
	if !insStatusValues[insStatusValuesLower] {
		return shim.Error("Invalid Instrument Status " + args[6])
	}

	//InsDueDate -> insDate
	insDueDate, err := time.Parse("02/01/2006", args[7])
	if err != nil {
		return shim.Error(err.Error())
	}

	//Checking if the programID exist or not
	//commented just for testing
	/*chk, err := stub.GetState(args[8])
	if chk == nil {
		return shim.Error("There is no program in this ID (checked from instrument)" + args[8])
	}*/

	//Converting the incoming date from Dd/mm/yy:hh:mm:ss to Dd/mm/yyThh:mm:ss for parsing
	vString := args[10][:10] + "T" + args[10][11:] //removing the ":" part from the string

	//ValueDate -> vDate
	vDate, err := time.Parse("02/01/2006T15:04:05", vString)
	if err != nil {
		return shim.Error("error in parsing the date and time (instrument)" + err.Error())
	}

	inst := instrumentInfo{args[1], instDate, args[3], args[4], insAmt, insStatusValuesLower, insDueDate, args[8], args[9], vDate}
	instBytes, err := json.Marshal(inst)
	if err != nil {
		return shim.Error(err.Error())
	}
	stub.PutState(args[0], instBytes)

	indexName := "instRefNo~sellBusID"
	refNoBusIDkey, err := stub.CreateCompositeKey(indexName, []string{inst.InstrumentRefNo, inst.SellBusinessID})
	if err != nil {
		return shim.Error("Unable to create instRefNo~sellBusID composite key:" + err.Error())
	}
	value := []byte{0x00}
	stub.PutState(refNoBusIDkey, value)
	return shim.Success(nil)
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

func getSellerID(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		xLenStr := strconv.Itoa(len(args))
		return shim.Error("Invalid number of arguments in getSellerID(instrument) (required:1) given: " + xLenStr)
	}

	instID := args[0]
	instRefNoSellIDiterator, err := stub.GetStateByPartialCompositeKey("instRefNo~sellBusID", []string{instID})
	if err != nil {
		return shim.Error("Unable to get the result for composite key : instRefNo~sellBusID")
	}
	var busID string
	for i := 0; instRefNoSellIDiterator.HasNext(); i++ {
		instRefNoSellIData, err := instRefNoSellIDiterator.Next()
		if err != nil {
			return shim.Error("Unable to iterate instRefNoSellIDiterator:" + err.Error())
		}
		_, requiredArgs, err := stub.SplitCompositeKey(instRefNoSellIData.Key)
		if err != nil {
			return shim.Error("error spliting the composite key instRefNoSellIDiterator:" + err.Error())
		}
		busID = requiredArgs[1]
	}
	return shim.Success([]byte(busID))
}

func main() {
	err := shim.Start(new(chainCode))
	if err != nil {
		fmt.Printf("Error starting Instrument chaincode: %s\n", err)
	}
}
