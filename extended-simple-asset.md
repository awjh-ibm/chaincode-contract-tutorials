# Using further features of the Contract API

## Calling functions every time a request is made
Sometimes functions in a contract may all have to repeat the same task. The Contract API provides provision for you to set a function to be called before and after each time a contract used in the chaincode is called.

For example each function in the [simple asset chaincode](samples/simple_asset_contract/simple-asset.go) performs the same action at the start, reading the world state using the passed asset ID. We could therefore define a function to be called before each call to perform this action. We can then use the returned asset in our already defined functions by storing it in the transaction context. To do this we must define our own custom transaction context which contains space to store this value:

```
// CustomTransactionContext - extends contractapi.TransactionContext with a field to store retrieved simple assets
type CustomTransactionContext struct {
	contractapi.TransactionContext
	callData []byte
}
```

The chaincode must be told to use our updated transaction context which is done by setting it in the main function BEFORE `contractapi.CreateNewChaincode` is called. Update the main function to include `sac.SetTransactionContextHandler` which sets the type of context to be sent.

```
func main() {
	sac := new(SimpleAsset)
	sac.SetTransactionContextHandler(new(CustomTransactionContext))

	if err := contractapi.CreateNewChaincode(sac); err != nil {
		fmt.Printf("Error starting SimpleAsset chaincode: %s", err)
	}
}
```

Now that the transaction context being sent is no longer `*contractapi.TransactionContext` we must update each of our functions to take it in its new type by changing `ctx *contractapi.TransactionContext` to `ctx *CustomTransactionContext`.

Once we have our transaction context fully set up we can then define the following function to get the data:

```
func getAsset(ctx *CustomTransactionContext, assetID string) error {

	existing, err := ctx.GetStub().GetState(assetID)

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	ctx.callData = existing

	return nil
}
```

The function sets the call data to be the value from the world state of the asset with the passed ID. Notice that the passed ID parameter is in the same position in the function declaration as it is in the other function declarations of the simple asset chaincode. This is key as the before function receives the same data as is passed in to the named function. There are functions named such as Update in the simple asset chaincode which take in more parameters, the above function as it has fewer parameters will not be passed the extra data when called. If the above function returns a non nil error then the response to the peer will be that error and the named function will not be called.

We must tell the chaincode to use this function before each function call and therefore we must call `SetBeforeTransaction` in the main function BEFORE we call `contractapi.CreateNewChaincode` but AFTER we set the transaction context handler. Notice that the function above is not linked to our simple asset struct nor is it public however it can still be used as a set function, it however cannot be called by itself by a user initialising, invoking or querying. It is perfectly possible however to use a function that is available for a user making such calls as a set function. The rule being for set functions that they must match the format of what is an allowed function in our chaincode outlined in [simple-asset.md](simple-asset.md#adding-functions-to-manage-our-asset).

Update the main function to include `sac.SetBeforeTransaction` to set the above function to be called every time a user makes a call to the simple asset contract:

```
func main() {
	sac := new(SimpleAsset)
	sac.SetTransactionContextHandler(new(CustomTransactionContext))
	sac.SetBeforeTransaction(getAsset)

	if err := contractapi.CreateNewChaincode(sac); err != nil {
		fmt.Printf("Error starting SimpleAsset chaincode: %s", err)
	}
}
```

You can then replace the repeated code in the other functions:

```
existing, err := ctx.GetStub().GetState(assetID)

if err != nil {
    return errors.New("Unable to interact with world state")
}

```

with

```
existing := ctx.callData
```

Note that the Read function implements the repeated code slightly differently due to its alternate return type.

## Performing a custom action when a user passes an unknown function name
By default when a user passes an unknown function an error will be returned to them telling them that a function of that name doesn't exist. It is possible to use the contract API to set a custom function to handle this occurrence, throwing a custom error or even returning a success message. Like with the before function above it is not necessary for the unknown function to be public or a method of the struct used in creating the chaincode, it merely needs to match the format of what is an allowed function in our chaincode as outlined in [simple-asset.md](simple-asset.md#adding-functions-to-manage-our-asset). Again it is possible to use a public (or private function) of the struct used in creating the chaincode.

Here is a function for custom handling of an unknown function name being passed:

```
func handleUnknown(ctx *CustomTransactionContext) error {
	fn, args := ctx.GetStub().GetFunctionAndParameters()

    return fmt.Errorf("Unknown function name %s passed with args %v", fn, args)
}
```

The above function takes in the context and uses the stub to get the details of what was passed by the call. The function then returns an error which will be returned as the peer's response. 

Update the main function to include `sac.SetUnknownTransaction` to set the above function to be called every time a user makes a call to the simple asset contract with an unknown function name:

```
func main() {
	sac := new(SimpleAsset)
	sac.SetTransactionContextHandler(new(CustomTransactionContext))
	sac.SetBeforeTransaction(getAsset)
	sac.SetUnknownTransaction(handleUnknown)

	if err := contractapi.CreateNewChaincode(sac); err != nil {
		fmt.Printf("Error starting SimpleAsset chaincode: %s", err)
	}
}
```

### How it all looks

```
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

func handleUnknown(ctx *CustomTransactionContext) error {
	fn, args := ctx.GetStub().GetFunctionAndParameters()

    return fmt.Errorf("Unknown function name %s passed with args %v", fn, args)
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
```

You can test your chaincode by [testing using dev mode](simple-asset.md#testing-using-dev-mode). It should perform in the same way as the non-extended chaincode you developed earlier.
