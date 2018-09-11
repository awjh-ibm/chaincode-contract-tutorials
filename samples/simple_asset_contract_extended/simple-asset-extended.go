/*
Copyright IBM Corp All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"errors"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/contractapi"
)

// CustomTransactionContext - extends contractapi.TransactionContext with a field to store retrieved simple assets
type CustomTransactionContext struct {
	contractapi.TransactionContext
	callData []byte
}

type SimpleAsset struct {
	contractapi.Contract
}

// Create - Initialises a simple asset with the given ID in the world state
func (sa *SimpleAsset) Create(ctx *CustomTransactionContext, assetID string) error {
	existing := ctx.callData

	if existing != nil {
		return fmt.Errorf("Cannot create asset. Asset with id %s already exists", assetID)
	}

	err := ctx.GetStub().PutState(assetID, []byte("Initialised"))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// Update - Updates a simple asset with given ID in the world state
func (sa *SimpleAsset) Update(ctx *CustomTransactionContext, assetID string, value string) error {
	existing := ctx.callData

	if existing == nil {
		return fmt.Errorf("Cannot update asset. Asset with id %s does not exist", assetID)
	}

	err := ctx.GetStub().PutState(assetID, []byte(value))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// Read - Returns value of a simple asset with given ID from world state as string
func (sa *SimpleAsset) Read(ctx *CustomTransactionContext, assetID string) (string, error) {
	existing := ctx.callData

	if existing == nil {
		return "", fmt.Errorf("Cannot read asset. Asset with id %s does not exist", assetID)
	}

	return string(existing), nil
}

func getAsset(ctx *CustomTransactionContext, assetID string) error {

	existing, err := ctx.GetStub().GetState(assetID)

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	ctx.callData = existing

	return nil
}

func handleUnknown(args []string) error {
	return fmt.Errorf("Unknown function name passed with args %v", args)
}

func main() {
	sac := new(SimpleAsset)
	sac.SetTransactionContextHandler(new(CustomTransactionContext))
	sac.SetBeforeTransaction(getAsset)
	sac.SetUnknownTransaction(handleUnknown)

	if err := contractapi.CreateNewChaincode(sac); err != nil {
		fmt.Printf("Error starting SimpleAsset chaincode: %s", err)
	}
}
