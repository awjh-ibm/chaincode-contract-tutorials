# Using multiple namespaces in a single chaincode
It may be that you wish to have a single chaincode to regulate logically seperate entities such as assets. This can be done using multiple structs embedding `contractapi.Contract` and provides benefits such as allowing shared generic functions to be used. In this example we will further extend the [extended simple asset contract](samples/simple_asset_contract_extended.md) by adding a second asset.

## Defining the second asset
First lets start by adding a type definition for the second asset. As we want this to become part of our chaincode it will need to embed the contractapi.Contract class. We will also make this a more complex asset than the initial simple asset by adding some fields and using JSON descriptors to allow us to store is as a byte array in the world state. Add this definition below the `Read` function of your simple asset:

```
type ComplexAsset struct {
    contractapi.Contract
    Owner string `json:"owner"`
	Value int    `json:"value"`
}
```

## Defining functions for the second asset
We can define functions in the same way as we did for the simple asset. We will add the same three basic operations of Create, Update and Read.

### Create
The first function we describe will be to create an instance of our asset in the world state:

```
// Create - Initialises a complex asset with the given ID in the world state
func (ca *ComplexAsset) Create(ctx *contractapi.TransactionContext, assetID string) error {
	existing, err := ctx.GetStub().GetState(assetID)

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	if existing != nil {
		return fmt.Errorf("Cannot create asset. Asset with id %s already exists", assetID)
	}

	ca.Owner = "Regulator"
	ca.Value = 0

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
```

This functions works in a similar way as the create function for the simple asset however as the asset has fields with JSON descriptors we can convert the more complex asset into JSON and store that in the world state.

### Update
The complex asset has multiple fields so we will define an update function to update each individually, the first of these will update the owner. This function will simple replace the value of the owner property of the asset:

```
// UpdateOwner - Updates a complex asset with given ID in the world state to have a new owner
func (ca *ComplexAsset) UpdateOwner(ctx *contractapi.TransactionContext, assetID string, newOwner string) error {
	existing, err := ctx.GetStub().GetState(assetID)

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	if existing == nil {
		return fmt.Errorf("Cannot update asset. Asset with id %s does not exist", assetID)
	}

	err = json.Unmarshal(existing, ca)

	if err != nil {
		return fmt.Errorf("Asset with id %s is not a ComplexAsset", assetID)
	}

	ca.Owner = newOwner

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
```

This function takes the existing asset found at assetID and checks that it is a complex asset by attempting to convert the JSON stored into a complex asset. It then updates the owner field and writes the new JSON version into the world state.

The second function we will write will update the value by adding the passed value to the existing value stored in the asset. As the contract functions can only take in strings we will need to convert the passed in value to an integer to do this:

```
// UpdateValue - Updates a complex asset with given ID in the world state to have a new value by adding the passed value to its existing value
func (ca *ComplexAsset) UpdateValue(ctx *contractapi.TransactionContext, assetID string, additionalValue string) error {
	additionalValueInt, err := strconv.Atoi(additionalValue)

	if err != nil {
		return fmt.Errorf("Cannot use passed value %s as value. It is not an integer", additionalValue)
	}

	existing, err := ctx.GetStub().GetState(assetID)

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	if existing == nil {
		return fmt.Errorf("Cannot update asset. Asset with id %s does not exist", assetID)
	}

	err = json.Unmarshal(existing, ca)

	if err != nil {
		return fmt.Errorf("Asset with id %s is not a ComplexAsset", assetID)
	}

	ca.Value += additionalValueInt

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
```

The function works in the same way as the previous update function by retrieving from the world state the value at the ID converting that into a complex asset and then updating the value and writing the changes to the world state.

### Read
The final function we will write will return to the user the value of complex asset. As we are storing the value as JSON we can return the JSON formatted value:


```
// Read - Returns the JSON value of a complex asset with given ID from world state as string
func (ca *ComplexAsset) Read(ctx *contractapi.TransactionContext, assetID string) (string, error) {
	existing, err := ctx.GetStub().GetState(assetID)

	if err != nil {
		return "", errors.New("Unable to interact with world state")
	}

	if existing == nil {
		return "", fmt.Errorf("Cannot read asset. Asset with id %s does not exist", assetID)
	}

	err = json.Unmarshal(existing, ca)

	if err != nil {
		return "", fmt.Errorf("Asset with id %s is not a ComplexAsset", assetID)
	}

	return string(existing), nil
}
```

## Adding the second asset to the chaincode
When we added our simple asset to the chaincode we created an instance of it in our main function and then passed that instance to `contractapi.CreateNewChaincode`. To add the complex asset functions to our chaincode we do the same, creating an instance and passing it to the same `contractapi.CreateNewChaincode` function. Because our unknown function handler is generic we will use that again for our new asset type although we could specify custom handlers for each. As discussed in the simple asset tutorial because we didn't set a custom namespace for the simple asset the default namespace was used to talk to its functions via the chaincode. Now that we have two sets of functions we now want to be able to talk to through the chaincode we need to have unique namespaces for each meaning at least one must have a custom namespace. Here we will elect to use two:

```
func main() {
    sac := new(SimpleAsset)
    sac.SetBeforeFn(getAsset)
    sac.SetUnknownFn(handleUnknown)
    sac.SetNamespace("simpleasset")

    cac := new(ComplexAsset)
    sac.SetUnknownFn(handleUnknown)
    cac.SetNamespace("complexasset")

	if err := contractapi.CreateNewChaincode(sac, cac); err != nil {
		fmt.Printf("Error starting SimpleAsset chaincode: %s", err)
	}
}
```

The main function now is generating a chaincode that has two namespaces within it, simpleasset and complexasset. Now when issuing commands to initialise, invoke or query the chaincode we can refer to the separate functions using namespace_functionname, for example if we wish to call our new Create function we can pass the first parameter `complexasset_Create` in our call. Note that although there are two namespaces you can only initialise a chaincode once.

## Simplifying our second asset
### Using a before function
You may have noticed whilst writing the above code that, like in the simple asset, every function performs the same action of getting data from the world state. This means that like we did in the extended simple asset tutorial we can remove that code and add a before function. As the code is the same for getting the simple asset as it is the complex we can use the same function. Add the before function to the complex asset by setting it in the main:

```
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
```

We can then replace the repeated get code in the complex asset's Create, UpdateOwner, UpdateValue and Read functions with a call to get the transaction context call data. Replace:

```
existing, err := ctx.GetStub().GetState(assetID)

if err != nil {
    return errors.New("Unable to interact with world state")
}

```

with

```
existing := ctx.GetCallData().([]byte)
```

Note that the Read function implements the repeated code slightly differently due to its alternate return type and that as the `err` initialisation is removed you may have to update a later `err =` to use `:=`. You then can update the `if` to check the length of existing as it will no longer return using nil. Replace:

```
if existing == nil {
```

with

```
if len(existing) == 0
```

in the UpdateOwner, UpdateValue and Read functions and replace:

```
if existing != nil {
```

with

```
if len(existing) > 0
```

in the Create function.

### Using a private function
Three of the four functions perform the same action to put the complex asset data in the world state. As not all the functions perform the same action it is not worth setting a function to run after each call and write this data but we can create a generic put function for complex assets. By making this function private we can prevent users calling this function by itself via the chaincode and also the function defintion does not have to match the limitations of public functions, although in this case our function does. We can describe the put function as such:

```
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
```

We can then replace the repeated code in Create, UpdateOwner and UpdateValue:

```
caJSON, err := json.Marshal(&ca)

if err != nil {
    return nil, errors.New("Error converting asset to JSON")
}

err = ctx.GetStub().PutState(assetID, caJSON)

if err != nil {
    return errors.New("Unable to interact with world state")
}
```

with

```
err = ca.put(ctx, assetID)

if err != nil {
    return err
}
```

Note that as `err` is not already defined in `Create` you should use `err :=` rather than `err =`.

### Further simplification
We could also split the file we have generated into multiple files for the same package, having a file for each of the simple asset, complex asset and helper functions. In this tutorial we will not do that.

## Putting it all together
Our final code should look like the below. We can run it using the [chaincode dev mode](simple-asset.md#testing-using-dev-mode) environment and making calls using our new namespaces and function names.

```
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
	Value int    `json:"value"`
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
```

### Example calls:
Create simple asset:
```
peer chaincode invoke -n mycc -c '{"Args":["simpleasset_Create", "SIMPLE_ASSET_1"]}' -C myc
```

Create complex asset:
```
peer chaincode invoke -n mycc -c '{"Args":["complexasset_Create", "COMPLEX_ASSET_1"]}' -C myc
```

Update simple asset:
```
peer chaincode invoke -n mycc -c '{"Args":["simpleasset_Update", "SIMPLE_ASSET_1", "Updated"]}' -C myc
```

Update complex asset owner:
```
peer chaincode invoke -n mycc -c '{"Args":["complexasset_UpdateOwner", "COMPLEX_ASSET_1", "Andy"]}' -C myc
```

Update complex asset owner:
```
peer chaincode invoke -n mycc -c '{"Args":["complexasset_UpdateValue", "COMPLEX_ASSET_1", "50"]}' -C myc
```

Read simple asset:
```
peer chaincode query -n mycc -c '{"Args":["simpleasset_Read","SIMPLE_ASSET_1"]}' -C myc
```

Read complex asset:
```
peer chaincode query -n mycc -c '{"Args":["complexasset_Read","COMPLEX_ASSET_1"]}' -C myc
```