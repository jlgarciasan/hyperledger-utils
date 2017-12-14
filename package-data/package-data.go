package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"log"
	"os"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/core/common/ccpackage"
	"github.com/hyperledger/fabric/core/common/ccprovider"

	"github.com/hyperledger/fabric/protos/utils"

	"fmt"
	"io/ioutil"

	pcommon "github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric/protos/peer"
)

func getPackageFromFile(ccpackfile string) ([]byte, error) {

	//fmt.Printf("Chaincode package file : " + ccpackfile)

	b, err := ioutil.ReadFile(ccpackfile)
	if err != nil {
		fmt.Printf("Error (%s)", err)
		return nil, err
	}
	return b, nil
}

//getPackageFromFile get the chaincode package from file and the extracted ChaincodeDeploymentSpec
func getDataFromPackage(b []byte) (proto.Message, *pb.ChaincodeDeploymentSpec, *pb.SignedChaincodeDeploymentSpec, error) {

	//fmt.Printf("\nObtaining CDS or SignedCDS package ...\n")

	//the bytes should be a valid package (CDS or SigedCDS)
	ccpack, err := ccprovider.GetCCPackage(b)
	if err != nil {
		fmt.Printf("Error (%s)", err)
		return nil, nil, nil, err
	}

	//fmt.Printf("\nObtaining package Object...\n")

	//either CDS or Envelope
	o, err := ccpack.GetPackageObject(), nil
	if err != nil {
		fmt.Printf("Error (%s)", err)
		return nil, nil, nil, err
	}

	// //try CDS first
	cds, ok := o.(*pb.ChaincodeDeploymentSpec)
	//sCDS, ok := o.(*pb.SignedChaincodeDeploymentSpec)

	if !ok || cds == nil {
		//try Envelope next
		env, ok := o.(*pcommon.Envelope)
		if !ok || env == nil {
			return nil, nil, nil, fmt.Errorf("error extracting valid chaincode package")
		}

		//this will check for a valid package Envelope
		_, sCDS, err := ccpackage.ExtractSignedCCDepSpec(env)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("error extracting valid signed chaincode package(%s)", err)
		}
		//Get the CDS from SignedCDS
		cds, err = utils.GetChaincodeDeploymentSpec(sCDS.ChaincodeDeploymentSpec)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("error extracting chaincode deployment spec(%s)", err)
		}

		return o, cds, sCDS, nil
	}

	return o, cds, nil, nil

}

func printChainCode(cds *pb.ChaincodeDeploymentSpec) {

	//Reading code package
	readerCP := bytes.NewReader(cds.CodePackage)
	gzipCP, err := gzip.NewReader(readerCP)

	if err != nil {
		log.Fatal(err)
	}

	defer gzipCP.Close()

	scanner := bufio.NewScanner(gzipCP)

	//We need to indicate the begining of CC

	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "package main") {
			fmt.Println(scanner.Text()) // Println will add back the final '\n'
			for scanner.Scan() {
				fmt.Println(scanner.Text())
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Reading standard input:", err)
	}
}

func printEndorsements(sCDS *pb.SignedChaincodeDeploymentSpec) {

	size := len(sCDS.OwnerEndorsements)

	for i := 0; i < size; i++ {
		//fmt.Printf("ENDORSER %d", i+1)
		fmt.Println(BytesToString(sCDS.OwnerEndorsements[i].Endorser))
		//fmt.Println(BytesToString(sCDS.OwnerEndorsements[i].Signature))
	}
}

//BytesToString Transform a byte[] to string
func BytesToString(data []byte) string {
	return string(data[:])
}

func main() {

	pathPackage := os.Args[1]

	//fmt.Printf("\nObtaining package ...\n")
	packageFile, _ := getPackageFromFile(pathPackage)

	//fmt.Println("Obtaining data from package ...")
	_, cds, sCDS, _ := getDataFromPackage(packageFile)

	fmt.Printf("\nChaincode package data")
	fmt.Printf("\n======================\n")
	fmt.Printf("\nPATH: " + cds.ChaincodeSpec.ChaincodeId.Path)
	fmt.Printf("\nNAME: " + cds.ChaincodeSpec.ChaincodeId.Name)
	fmt.Printf("\nVERSION: " + cds.ChaincodeSpec.ChaincodeId.Version)
	fmt.Printf("\nCHAINCODE:\n\n")
	printChainCode(cds)
	printEndorsements(sCDS)
}
