package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	// "code.uni-ledger.com/liuhangyu/fabric/core/chaincode/shim"
	// pb "code.uni-ledger.com/liuhangyu/fabric/protos/peer"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

const (
	STANDARD_HEAD = "##STANDARD##"
	PUYANDGET_KEY = "STANDARD_KEY_"
)

var (
	logger = shim.NewLogger("DigitalAssets")
)

type DigitalAssets struct {
}

//链码测试
func (s *DigitalAssets) IsAlive(APIstub shim.ChaincodeStubInterface, args []string) pb.Response {
	if args[0] == "sysinfo" {
		sysinfo, err := GetSysInfo()
		if err != nil {
			return shim.Error(err.Error())
		}

		valByte, err := json.Marshal(sysinfo)
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(valByte)

	} else if args[0] == "isalive" {
		val, getErr := APIstub.GetState("isalive")
		if getErr != nil {
			return shim.Error(fmt.Sprintf("Failed to get state: %s", getErr.Error()))
		}
		return shim.Success(val)
	}

	return shim.Success([]byte("OK"))
}

func (s *DigitalAssets) PutStandard(APIstub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments, expecting 2")
	}

	name := args[0]
	valStr := args[1]

	compositeKey, compositeErr := APIstub.CreateCompositeKey(PUYANDGET_KEY, []string{STANDARD_HEAD, name})
	if compositeErr != nil {
		return shim.Error(fmt.Sprintf("Could not create a composite key for %s: %s",
			compositeKey, compositeErr.Error()))
	}

	putErr := APIstub.PutState(compositeKey, []byte(valStr))
	if putErr != nil {
		return shim.Error(fmt.Sprintf("Failed to put state: %s", putErr.Error()))
	}

	eventName := "put." + APIstub.GetTxID()
	err := APIstub.SetEvent(eventName, []byte(APIstub.GetTxID()))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte("from cc hello world"))
}

func (s *DigitalAssets) GetStandard(APIstub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments, expecting 1")
	}

	name := args[0]

	compositeKey, compositeErr := APIstub.CreateCompositeKey(PUYANDGET_KEY, []string{STANDARD_HEAD, name})
	if compositeErr != nil {
		return shim.Error(fmt.Sprintf("Could not create a composite key for %s: %s",
			compositeKey, compositeErr.Error()))
	}

	val, getErr := APIstub.GetState(compositeKey)
	if getErr != nil {
		return shim.Error(fmt.Sprintf("Failed to get state: %s", getErr.Error()))
	}

	return shim.Success(val)
}

//建立资源权限
type ResInfo struct {
	ResType string `json:"restype"`
	ResID   string `json:"resid"`
}

func (s *DigitalAssets) PutStandardUint(APIstub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments, expecting 2")
	}

	traderAttr, err := GetTrader(APIstub)
	if err != nil {
		return shim.Error("get trader err:" + err.Error())
	}
	logger.Debug("put", traderAttr)

	name := args[0]
	valStr := args[1]

	Avalbytes, err := APIstub.GetState(name)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + name + "\"}"
		return shim.Error(jsonResp)
	}

	var iOldVal uint64
	if Avalbytes != nil {
		iOldVal, err = strconv.ParseUint(string(Avalbytes), 10, 64)
		if err != nil {
			jsonResp := fmt.Sprintf("1c strconv %s to int64  %s", string(Avalbytes), err.Error())
			return shim.Error(jsonResp)
		}
	}

	iNewVal, err := strconv.ParseUint(valStr, 10, 64)
	if err != nil {
		jsonResp := fmt.Sprintf("2c strconv %s to int64  %s", valStr, err.Error())
		return shim.Error(jsonResp)
	}
	iNewVal += iOldVal

	sVal := strconv.FormatUint(iNewVal, 10)

	putErr := APIstub.PutState(name, []byte(sVal))
	if putErr != nil {
		return shim.Error(fmt.Sprintf("Failed to put state: %s", putErr.Error()))
	}

	eventName := "put." + APIstub.GetTxID()
	err = APIstub.SetEvent(eventName, []byte(APIstub.GetTxID()))
	if err != nil {
		return shim.Error(err.Error())
	}

	eventName = "put.liuhy" //覆盖上一个event
	err = APIstub.SetEvent(eventName, []byte("my test"))
	if err != nil {
		return shim.Error(err.Error())
	}

	resInfo := &ResInfo{ResType: "catalog", ResID: name}
	bytes, err := json.Marshal(resInfo)
	if err != nil {
		return shim.Error(err.Error())
	}

	trans := [][]byte{[]byte("ResAuth"), []byte("enrollResAuth"), bytes}
	response := APIstub.InvokeChaincode("mycc1", trans, "mychannel")
	if response.Status != shim.OK {
		fmt.Println(response.Message)
		return shim.Error(response.Message)
	}
	fmt.Println(response.Payload)

	// return shim.Success([]byte("from cc hello world"))
	return shim.Success([]byte(response.Payload))
}

func (s *DigitalAssets) GetStandardUint(APIstub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments, expecting 1")
	}

	traderAttr, err := GetTrader(APIstub)
	if err != nil {
		return shim.Error("get trader err:" + err.Error())
	}
	logger.Debug("get", traderAttr)

	name := args[0]
	val, getErr := APIstub.GetState(name)
	if getErr != nil {
		return shim.Error(fmt.Sprintf("Failed to get state: %s", getErr.Error()))
	}

	return shim.Success(val)
}

//链码初始化
func (s *DigitalAssets) Init(APIstub shim.ChaincodeStubInterface) pb.Response {
	putErr := APIstub.PutState("isalive", []byte("true"))
	if putErr != nil {
		return shim.Error(fmt.Sprintf("Failed to put state: %s", putErr.Error()))
	}
	return shim.Success(nil)
}

func (s *DigitalAssets) Invoke(APIstub shim.ChaincodeStubInterface) pb.Response {
	function, args := APIstub.GetFunctionAndParameters()
	switch function {
	case "isalive":
		return s.IsAlive(APIstub, args) //监控chaincode
	case "putstandard":
		return s.PutStandard(APIstub, args) //测试写
	case "getstandard":
		return s.GetStandard(APIstub, args) //测试读
	case "putstandardint":
		return s.PutStandardUint(APIstub, args) //测试写
	case "getstandardint":
		return s.GetStandardUint(APIstub, args) //测试读
	}
	return shim.Error("Invalid Smart Contract function name.")
}

func main() {
	err := shim.Start(new(DigitalAssets))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
