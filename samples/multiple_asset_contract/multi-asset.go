/*
Copyright IBM Corp All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/contractapi"
)

// CustomTransactionContext - extends contractapi.TransactionContext with a field to store retrieved simple assets
type CustomTransactionContext struct {
	contractapi.TransactionContext
	callData []byte
}

// PutComplexAsset - writes a complex asset to the world state
func (ctx *CustomTransactionContext) PutComplexAsset(assetID string, ca *ComplexAsset) error {
	caJSON, err := json.Marshal(&ca)

	if err != nil {
		return errors.New("Error converting asset to JSON")
	}

	err = ctx.GetStub().PutState(assetID, caJSON)

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
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

	return string(string(existing)), nil
}

type ComplexAsset struct {
	contractapi.Contract
	Owner string `json:"owner"`
	Value int    `json:"value"`
}

// Create - Initialises a complex asset with the given ID in the world state
func (ca *ComplexAsset) Create(ctx *CustomTransactionContext, assetID string) error {
	existing := ctx.callData

	if existing != nil {
		return fmt.Errorf("Cannot create asset. Asset with id %s already exists", assetID)
	}

	ca.Owner = "Regulator"
	ca.Value = 0

	err := ctx.PutComplexAsset(assetID, ca)

	if err != nil {
		return err
	}

	return nil
}

// UpdateOwner - Updates a complex asset with given ID in the world state to have a new owner
func (ca *ComplexAsset) UpdateOwner(ctx *CustomTransactionContext, assetID string, newOwner string) error {
	existing := ctx.callData

	if existing == nil {
		return fmt.Errorf("Cannot update asset. Asset with id %s does not exist", assetID)
	}

	err := json.Unmarshal(existing, ca)

	if err != nil {
		return fmt.Errorf("Asset with id %s is not a ComplexAsset", assetID)
	}

	ca.Owner = newOwner

	err = ctx.PutComplexAsset(assetID, ca)

	if err != nil {
		return err
	}

	return nil
}

// UpdateValue - Updates a complex asset with given ID in the world state to have a new value by adding the passed value to its existing value
func (ca *ComplexAsset) UpdateValue(ctx *CustomTransactionContext, assetID string, additionalValue string) error {
	additionalValueInt, err := strconv.Atoi(additionalValue)

	if err != nil {
		return fmt.Errorf("Cannot use passed value %s as value. It is not an integer", additionalValue)
	}

	existing := ctx.callData

	if existing == nil {
		return fmt.Errorf("Cannot update asset. Asset with id %s does not exist", assetID)
	}

	err = json.Unmarshal(existing, ca)

	if err != nil {
		return fmt.Errorf("Asset with id %s is not a ComplexAsset", assetID)
	}

	ca.Value += additionalValueInt

	err = ctx.PutComplexAsset(assetID, ca)

	if err != nil {
		return err
	}

	return nil
}

// Read - Returns the JSON value of a complex asset with given ID from world state as string
func (ca *ComplexAsset) Read(ctx *CustomTransactionContext, assetID string) (string, error) {
	existing := ctx.callData

	if existing == nil {
		return "", fmt.Errorf("Cannot read asset. Asset with id %s does not exist", assetID)
	}

	err := json.Unmarshal(existing, ca)

	if err != nil {
		return "", fmt.Errorf("Asset with id %s is not a ComplexAsset", assetID)
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
	sac.SetNamespace("org.example.assets.simple")

	cac := new(ComplexAsset)
	cac.SetTransactionContextHandler(new(CustomTransactionContext))
	cac.SetBeforeTransaction(getAsset)
	cac.SetUnknownTransaction(handleUnknown)
	cac.SetNamespace("org.example.assets.complex")

	if err := contractapi.CreateNewChaincode(sac, cac); err != nil {
		fmt.Printf("Error starting SimpleAsset chaincode: %s", err)
	}
}
