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
	ProgramBusinessDiscountPercentage float64
	StaleDays                         int
	RepaymentAcNo                     string
	RepaymentWalletID                 string
}

func (c *chainCode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (c *chainCode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()

	if function == "createPPR" {
		return createPPR(stub, args)
	} else if function == "seePPR" {
		return seePPR(stub, args)
	}
	return shim.Error("No function named " + function + " in PPR")
}

func createPPR(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 10 {
		xLenStr := strconv.Itoa(len(args))
		return shim.Error("Invalid number of arguments in createPPR (required:10) given:" + xLenStr)
	}

	relationship := map[string]bool{
		"Seller / Vendor": true,
		"Buyer / Dealer":  true,
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
	PBDpercentange, err := strconv.ParseFloat(args[7], 64)
	if err != nil {
		return shim.Error(err.Error())
	}

	//StaleDays -> sDays
	sDays, err := strconv.Atoi(args[8])
	if err != nil {
		return shim.Error(err.Error())
	}

	ppr := pprInfo{args[1], args[2], relationshipLower, PBLimit, PBroi, PBDperiod, PBDpercentange, sDays, args[9], args[10]}
	pprBytes, err := json.Marshal(ppr)
	err = stub.PutState(args[0], pprBytes)

	return shim.Success(nil)
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

func main() {
	err := shim.Start(new(chainCode))
	if err != nil {
		fmt.Printf("Error starting PPR chaincode: %s\n", err)
	}
}
