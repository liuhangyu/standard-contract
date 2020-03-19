package main

import (
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	msp "github.com/hyperledger/fabric/protos/msp"
)

//获取CC环境参数
type SystemInfo struct {
	TotalMem int64
	FreeMem  int64
	UsedMem  int64
	Time     time.Time
}

func GetSysInfo() (*SystemInfo, error) {
	var (
		total   string
		freeMem string
	)

	contents, err := ioutil.ReadFile("/proc/meminfo")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(contents), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "MemTotal") {
			total = strings.Split(strings.Trim(strings.Split(line, ":")[1], " "), " ")[0]
		} else if strings.HasPrefix(line, "MemFree") {
			freeMem = strings.Split(strings.Trim(strings.Split(line, ":")[1], " "), " ")[0]
		}
		continue
	}

	iTotal, err := strconv.ParseInt(total, 10, 64)
	if err != nil {
		return nil, err
	}
	iTotal = iTotal / 1024

	iFreeMem, err := strconv.ParseInt(freeMem, 10, 64)
	if err != nil {
		return nil, err
	}
	iFreeMem = iFreeMem / 1024

	iUsed := iTotal - iFreeMem

	sysInfo := &SystemInfo{
		TotalMem: iTotal,
		FreeMem:  iFreeMem,
		UsedMem:  iUsed,
		Time:     time.Now().UTC(),
	}
	return sysInfo, err
}

//用户身份注册以及查询
type UserInfo struct {
	OrgID    string `json:"orgid"`
	IssuerCN string `json:"issuercn"`
	UserName string `json:"username"`
	PubKey   []byte `json:"pubkey,omitempty"`
	X509     []byte `json:"x509,omitempty"`
}

func GetTrader(stub shim.ChaincodeStubInterface) (*UserInfo, error) {
	creatorCert, err := stub.GetCreator()
	if err != nil {
		return nil, err
	}

	identity := msp.SerializedIdentity{}
	err = proto.Unmarshal(creatorCert, &identity)
	if err != nil {
		return nil, err
	}

	cert, _ := pem.Decode(identity.IdBytes)
	cert509, err := x509.ParseCertificate(cert.Bytes)
	if err != nil {
		return nil, err
	}

	publicKeyDer, err := x509.MarshalPKIXPublicKey(cert509.PublicKey)
	if err != nil {
		return nil, err
	}

	publicKeyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyDer,
	}
	pubKeyPem := pem.EncodeToMemory(publicKeyBlock)

	userAttr := &UserInfo{
		OrgID:    identity.Mspid,
		IssuerCN: cert509.Issuer.CommonName,
		UserName: cert509.Subject.CommonName,
		PubKey:   pubKeyPem,
		X509:     identity.IdBytes,
	}

	return userAttr, nil
}
