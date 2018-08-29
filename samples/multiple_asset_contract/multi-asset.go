package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/contractapi"
)

type SimpleAsset struct {
	contractapi.Contract
}

// Create - Initialises a simple asset with the given ID in the world state
func (sa *SimpleAsset) Create(ctx *contractapi.TransactionContext, assetID string) error {
	existing := ctx.GetCallData().([]byte)

	if len(existing) > 0 {
		return fmt.Errorf("Cannot create asset. Asset with id %s already exists", assetID)
	}

	err := ctx.GetStub().PutState(assetID, []byte("Initialised"))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// Update - Updates a simple asset with given ID in the world state
func (sa *SimpleAsset) Update(ctx *contractapi.TransactionContext, assetID string, value string) error {
	existing := ctx.GetCallData().([]byte)

	if len(existing) == 0 {
		return fmt.Errorf("Cannot update asset. Asset with id %s does not exist", assetID)
	}

	err := ctx.GetStub().PutState(assetID, []byte(value))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// Read - Returns value of a simple asset with given ID from world state as string
func (sa *SimpleAsset) Read(ctx *contractapi.TransactionContext, assetID string) (string, error) {
	existing := ctx.GetCallData().([]byte)

	if len(existing) == 0 {
		return "", fmt.Errorf("Cannot read asset. Asset with id %s does not exist", assetID)
	}

	return string(string(existing)), nil
}

type ComplexAsset struct {
	contractapi.Contract
	Owner string `json:"owner"`
	Value int    `json:value`
}

// Create - Initialises a complex asset with the given ID in the world state
func (ca *ComplexAsset) Create(ctx *contractapi.TransactionContext, assetID string) error {
	existing := ctx.GetCallData().([]byte)

	if len(existing) > 0 {
		return fmt.Errorf("Cannot create asset. Asset with id %s already exists", assetID)
	}

	ca.Owner = "Regulator"
	ca.Value = 0

	err := ca.put(ctx, assetID)

	if err != nil {
		return err
	}

	return nil
}

// UpdateOwner - Updates a complex asset with given ID in the world state to have a new owner
func (ca *ComplexAsset) UpdateOwner(ctx *contractapi.TransactionContext, assetID string, newOwner string) error {
	existing := ctx.GetCallData().([]byte)

	if len(existing) == 0 {
		return fmt.Errorf("Cannot update asset. Asset with id %s does not exist", assetID)
	}

	err := json.Unmarshal(existing, ca)

	if err != nil {
		return fmt.Errorf("Asset with id %s is not a ComplexAsset", assetID)
	}

	ca.Owner = newOwner

	err = ca.put(ctx, assetID)

	if err != nil {
		return err
	}

	return nil
}

// UpdateValue - Updates a complex asset with given ID in the world state to have a new value by adding the passed value to its existing value
func (ca *ComplexAsset) UpdateValue(ctx *contractapi.TransactionContext, assetID string, additionalValue string) error {
	additionalValueInt, err := strconv.Atoi(additionalValue)

	if err != nil {
		return fmt.Errorf("Cannot use passed value %s as value. It is not an integer", additionalValue)
	}

	existing := ctx.GetCallData().([]byte)

	if len(existing) == 0 {
		return fmt.Errorf("Cannot update asset. Asset with id %s does not exist", assetID)
	}

	err = json.Unmarshal(existing, ca)

	if err != nil {
		return fmt.Errorf("Asset with id %s is not a ComplexAsset", assetID)
	}

	ca.Value += additionalValueInt

	err = ca.put(ctx, assetID)

	if err != nil {
		return err
	}

	return nil
}

// Read - Returns the JSON value of a complex asset with given ID from world state as string
func (ca *ComplexAsset) Read(ctx *contractapi.TransactionContext, assetID string) (string, error) {
	existing := ctx.GetCallData().([]byte)

	if len(existing) == 0 {
		return "", fmt.Errorf("Cannot read asset. Asset with id %s does not exist", assetID)
	}

	err := json.Unmarshal(existing, ca)

	if err != nil {
		return "", fmt.Errorf("Asset with id %s is not a ComplexAsset", assetID)
	}

	return string(existing), nil
}

func (ca *ComplexAsset) put(ctx *contractapi.TransactionContext, assetID string) error {
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

func getAsset(ctx *contractapi.TransactionContext, assetID string) error {

	existing, err := ctx.GetStub().GetState(assetID)

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	ctx.SetCallData(existing)

	return nil
}

func handleUnknown(args []string) error {
	return fmt.Errorf("Unknown function name passed with args %v", args)
}

func main() {
	sac := new(SimpleAsset)
	sac.SetBeforeFn(getAsset)
	sac.SetUnknownFn(handleUnknown)
	sac.SetNamespace("simpleasset")

	cac := new(ComplexAsset)
	cac.SetBeforeFn(getAsset)
	cac.SetUnknownFn(handleUnknown)
	cac.SetNamespace("complexasset")

	if err := contractapi.CreateNewChaincode(sac, cac); err != nil {
		fmt.Printf("Error starting SimpleAsset chaincode: %s", err)
	}
}
