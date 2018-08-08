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
	ProgramID                         string  //[1]
	BusinessID                        string  //[2]
	Relationship                      string  //[3]
	ProgramBusinessLimit              int64   //[4]
	ProgramBusinessROI                float64 //[5]
	ProgramBusinessDiscountPeriod     int     //[6]
	ProgramBusinessDiscountPercentage string  //[7]//use float64 for parsing
	StaleDays                         int     //[8]
	RepaymentAcNo                     string  //[9]
	RepaymentWalletID                 string  //will be taken from business Id
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
		//Creates a new PPR Information
		return createPPR(stub, args)
	} else if function == "seePPR" {
		//Retrieves the PPR Information
		return seePPR(stub, args)
	} else if function == "pprIDexists" {
		//Checks the existence of PprID
		return pprIDexists(stub, args[0])
	} else if function == "discountPercentage" {
		//Returns DiscountPercentage
		return discountPercentage(stub, args)
	} else if function == "updatePPR" {
		/*
			Parameters for Value Calculation
			    a. Program Business Limit
			    b. Program Business ROI
			    c. Program Business Discount Percentage
				d. Program Business Discount Period
			are updated
		*/
		return updatePPR(stub, args)
	}
	return shim.Error("No function named " + function + " in PPRsssssss")
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

	//Wallet ID for repayment
	chaincodeArgs = toChaincodeArgs("getWalletID", args[2], "main")
	response = stub.InvokeChaincode("businesscc", chaincodeArgs, "myc")
	if response.Status != shim.OK {
		return shim.Error(response.Message)
	}
	repayWalletID := string(response.GetPayload())

	ppr := pprInfo{args[1], args[2], relationshipLower, PBLimit, PBroi, PBDperiod, args[7], sDays, args[9], repayWalletID}
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

func updatePPR(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 3 {
		xLenStr := strconv.Itoa(len(args))
		return shim.Error("Invalid number of arguments in updatePPR(PPR) (required:3) given:" + xLenStr)
	}
	/*
		args[0] -> pprID
		args[1] -> Program Business Limit, Program Business ROI,
				   Program Business Discount Percentage, Program Business Discount Period
		args[2] -> values
	*/
	pprObject := pprInfo{}
	pprBytes, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error("updatePPR(PPR)" + err.Error())
	}

	err = json.Unmarshal(pprBytes, &pprObject)
	lowerStr := strings.ToLower(args[1])

	if lowerStr == "program business limit" {
		//Changing Program Business Limit
		PBLimit, err := strconv.ParseInt(args[2], 10, 64)
		if err != nil {
			return shim.Error("updatePPR(PPR) Program Business Limit" + err.Error())
		}
		pprObject.ProgramBusinessLimit = PBLimit
	} else if lowerStr == "program business roi" {
		//Changing Program Business ROI
		PBroi, err := strconv.ParseFloat(args[2], 64)
		if err != nil {
			return shim.Error("updatePPR(PPR) Program Business ROI" + err.Error())
		}
		pprObject.ProgramBusinessROI = PBroi
	} else if lowerStr == "program business discount percentage" {
		//Changing Program Business Discount Percentage
		_, err = strconv.ParseFloat(args[2], 64)
		if err != nil {
			return shim.Error("updatePPR(PPR) Program Business Discount Percentage" + err.Error())
		}
		pprObject.ProgramBusinessDiscountPercentage = args[2]
	} else if lowerStr == "program business discount period" {
		//Chainging Program Business Discount Period
		PBDperiod, err := strconv.Atoi(args[2])
		if err != nil {
			return shim.Error("updatePPR(PPR) Program Business Discount Period" + err.Error())
		}
		pprObject.ProgramBusinessDiscountPeriod = PBDperiod
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
	defer prgrmBusPercentageIte.Close()
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
