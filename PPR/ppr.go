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

type pprInfo struct {
	ProgramID                         string
	BusinessID                        string
	Relationship                      string
	ProgramBusinessLimit              int64
	ProgramBusinessROI                float64
	ProgramBusinessDiscountPeriod     int
	ProgramBusinessDiscountPercentage string //use float64 for parsing
	StaleDays                         int
	RepaymentAcNo                     string
	RepaymentWalletID                 string
}

func (c *chainCode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	indexName := "ProgramID~BusinessID~DiscountPercentage"
	ppr := pprInfo{}
	prgrmBusPercentageKey, err := stub.CreateCompositeKey(indexName, []string{ppr.ProgramID, ppr.BusinessID, ppr.ProgramBusinessDiscountPercentage})
	if err != nil {
		return shim.Error("Unableto create composite key ProgramID~BusinessID~DiscountPercentage :" + err.Error())

	}
	value := []byte{0x00}
	stub.PutState(prgrmBusPercentageKey, value)
	return shim.Success(nil)
}

func (c *chainCode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()

	if function == "createPPR" {
		return createPPR(stub, args)
	} else if function == "seePPR" {
		return seePPR(stub, args)
	} else if function == "pprIDexists" {
		return pprIDexists(stub, args[0])
	} else if function == "discountPercentage" {
		return discountPercentage(stub, args)
	}
	return shim.Error("No function named " + function + " in PPR")
}

func createPPR(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 10 {
		xLenStr := strconv.Itoa(len(args))
		return shim.Error("Invalid number of arguments in createPPR (required:10) given:" + xLenStr)
	}

	//Checking existence of PprID
	response := pprIDexists(stub, args[0])
	if response.Status != shim.OK {
		return shim.Error(response.Message)
	}

	//Checking existence of businessID
	chaincodeArgs := toChaincodeArgs("busIDexists", args[2])
	response = stub.InvokeChaincode("businesscc", chaincodeArgs, "myc")
	if response.Status == shim.OK {
		return shim.Error("BusinessId " + args[2] + " does not exits")
	}

	//Checking existence of ProgramID
	chaincodeArgs = toChaincodeArgs("programIDexists", args[1])
	response = stub.InvokeChaincode("programcc", chaincodeArgs, "myc")
	if response.Status == shim.OK {
		return shim.Error("ProgramId " + args[1] + " does not exits")
	}

	relationship := map[string]bool{
		"Seller": true,
		"Vendor": true,
		"Buyer":  true,
		"Dealer": true,
	}

	relationshipLower := strings.ToLower(args[3])

	if !relationship[relationshipLower] {
		return shim.Error("Invalid relationship " + relationshipLower)
	}

	// ProgramBusinessLimit -> PBLimit
	PBLimit, err := strconv.ParseInt(args[4], 10, 64)
	if err != nil {
		return shim.Error(err.Error())
	}

	//ProgramBusinessROI -> PBroi
	PBroi, err := strconv.ParseFloat(args[5], 64)
	if err != nil {
		return shim.Error(err.Error())
	}

	//ProgramBusinessDiscountPeriod -> PBDperiod
	PBDperiod, err := strconv.Atoi(args[6])
	if err != nil {
		return shim.Error(err.Error())
	}

	//ProgramBusinessDiscountPercentage -> PBDpercentange
	_, err = strconv.ParseFloat(args[7], 64)
	if err != nil {
		return shim.Error(err.Error())
	}

	//StaleDays -> sDays
	sDays, err := strconv.Atoi(args[8])
	if err != nil {
		return shim.Error(err.Error())
	}

	ppr := pprInfo{args[1], args[2], relationshipLower, PBLimit, PBroi, PBDperiod, args[7], sDays, args[9], args[10]}
	pprBytes, err := json.Marshal(ppr)
	err = stub.PutState(args[0], pprBytes)

	return shim.Success(nil)
}

func pprIDexists(stub shim.ChaincodeStubInterface, pprID string) pb.Response {
	ifExists, _ := stub.GetState(pprID)
	if ifExists != nil {
		fmt.Println(ifExists)
		return shim.Error("PprId " + pprID + " exits. Cannot create new ID")
	}
	return shim.Success(nil)
}

func discountPercentage(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	prgrmBusPercentageIte, err := stub.GetStateByPartialCompositeKey("ProgramID~BusinessID~DiscountPercentage", []string{args[0], args[1]})
	prgrmBusPercentageData, _ := prgrmBusPercentageIte.Next()
	_, data, err := stub.SplitCompositeKey(prgrmBusPercentageData.Key)
	if err != nil {
		return shim.Error("Error spliting composite key ProgramID~BusinessID~DiscountPercentage (ppr):" + err.Error())
	}
	return shim.Success([]byte(data[2]))
}

func seePPR(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		xLenStr := strconv.Itoa(len(args))
		return shim.Error("Invalid number of arguments in seePPR (required:1) given:" + xLenStr)
	}

	pprObject := pprInfo{}
	pprArray, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error(err.Error())
	}

	err = json.Unmarshal(pprArray, &pprObject)
	pprString := fmt.Sprintf("%+v", pprObject)

	return shim.Success([]byte(pprString))

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
		fmt.Printf("Error starting PPR chaincode: %s\n", err)
	}
}
