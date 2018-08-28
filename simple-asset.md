# Writing chaincode using the Contract API
Chaincode running on Hyperledger Fabric must implement the chaincode interface. This can be done by manually adding the required functions or by using the Contract API. This tutorial will cover how you can develop chaincode using the Contract API.

## Simple Asset Chaincode
Our application is a basic sample chaincode to create assets (key-value pairs) on the ledger.

### Choosing a Location for the Code
If you haven’t been doing programming in Go, you may want to make sure that you have [Go Programming Language](https://hyperledger-fabric.readthedocs.io/en/release-1.2/prereqs.html#golang) installed and your system properly configured.

Now, you will want to create a directory for your chaincode application as a child directory of `$GOPATH/src/`.

To keep things simple, let’s use the following command:

```
mkdir -p $GOPATH/src/sacc && cd $GOPATH/src/sacc
```

Now, let’s create the source file that we’ll fill in with code:

```
touch sacc.go
```

### Housekeeping
First lets start with some housekeeping. As with every chaincode ours must implement the [Chaincode Interface](https://godoc.org/github.com/hyperledger/fabric/core/chaincode/shim#Chaincode), fortunately the Contract API has a Struct call Contract which implements this and saves us having to write our own implementation. So lets start with importing the Contract API package. Next, let's add a struct `SimpleAsset` which will provide the functions to manage our asset and embed `contractapi.Contract` to ensure it can be used in chaincode.

```
package main

import (
	"github.com/hyperledger/fabric/core/chaincode/contractapi"
)

type SimpleAsset struct {
    contractapi.Contract
}
```

### Adding functions to manage our asset
Functions that are to be exported to our chaincode must act against a struct embedding contractapi.Contract and must be made public. They also must fit a specific format, they may take in zero or more arguments and return zero, one or two values. The arguments taken in can be any of the following:
- *contractapi.TransactionContext
- string
- []string

Any number of string parameters may be taken in however only zero or one parameter may be of types *contractapi.TransactionContext and []string. There is also the limitation that if a parameter of type *contractapi.TransactionContext is taken in then it must be the first parameter in the definition. If a parameter of type []string is taken in this must be the last listed parameter in the definition.

Functions can be defined to return zero, one or two values. These can be of types:
- string
- error

At most one string and one error can be returned. If no error is returned the value of the string will be returned to the chaincode interaction request.

#### Create
The first function we will write will be a create function for the asset. The function will take in an ID for the asset and initialise it in the world state:

```
// Create - Initialises a simple asset with the given ID in the world state
func (sa *SimpleAsset) Create(ctx *contractapi.TransactionContext, assetID string) error {
	existing, err := ctx.GetStub().GetState(assetID)

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	if existing != nil {
		return fmt.Errorf("Cannot create asset. Asset with id %s already exists", assetID)
	}

	err = ctx.GetStub().PutState(assetID, []byte("Initialised"))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}
```

The above function uses the stub from the transaction context to read and write data to the world state. It checks whether it can connect to the world state, if an asset with that ID exists already and then writes an initial value for the asset. As the function only returns an error type the peer response to the request will be a blank string success unless an error is specified.

#### Update
The second function we write will take in an ID of an asset to update the value for and the value to update it to:

```
// Update - Updates a simple asset with given ID in the world state
func (sa *SimpleAsset) Update(ctx *contractapi.TransactionContext, assetID string, value string) error {
	existing, err := ctx.GetStub().GetState(assetID)

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	if existing == nil {
		return fmt.Errorf("Cannot update asset. Asset with id %s does not exist", assetID)
	}

	err = ctx.GetStub().PutState(assetID, []byte(value))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}
```

The above function uses the stub from the transaction context to read and write data to the world state. It checks whether it can connect to the world state, if an asset with that ID exists already and then writes the updated value for the asset. As the function only returns an error type the message back will be a blank string unless an error is specified.

#### Read
The third function will allow users to query the value of an asset:

```
// Read - Returns value of a simple asset with given ID from world state as string
func (sa *SimpleAsset) Read(ctx *contractapi.TransactionContext, assetID string) (string, error) {
	existing, err := ctx.GetStub().GetState(assetID)

	if err != nil {
		return "", errors.New("Unable to interact with world state")
	}

	if existing == nil {
		return "", fmt.Errorf("Cannot read asset. Asset with id %s does not exist", assetID)
	}

	return string(existing), nil
}
```

The above function reads the asset from the world state and returns it as a string (it is stored in the world state as an array of bytes). The function returns both string and error types. If the error is set to nil then the user's query will be answered with the string value returned otherwise they will receive the error.

### Putting it All Together
Finally we need to add the main function to start up a chaincode instance using our functions declared for managing a simple asset. This will require us to pass a new instance of `SimpleAsset` to `contractapi.CreateNewChaincode`. Here is the full contract source:

```
package main

import (
	"errors"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/contractapi"
)

type SimpleAsset struct {
	contractapi.Contract
}

// Create - Initialises a simple asset with the given ID in the world state
func (sa *SimpleAsset) Create(ctx *contractapi.TransactionContext, assetID string) error {
	existing, err := ctx.GetStub().GetState(assetID)

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	if existing != nil {
		return fmt.Errorf("Cannot create asset. Asset with id %s already exists", assetID)
	}

	err = ctx.GetStub().PutState(assetID, []byte("Initialised"))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// Update - Updates a simple asset with given ID in the world state
func (sa *SimpleAsset) Update(ctx *contractapi.TransactionContext, assetID string, value string) error {
	existing, err := ctx.GetStub().GetState(assetID)

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	if existing == nil {
		return fmt.Errorf("Cannot update asset. Asset with id %s does not exist", assetID)
	}

	err = ctx.GetStub().PutState(assetID, []byte(value))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// Read - Returns value of a simple asset with given ID from world state as string
func (sa *SimpleAsset) Read(ctx *contractapi.TransactionContext, assetID string) (string, error) {
	existing, err := ctx.GetStub().GetState(assetID)

	if err != nil {
		return "", errors.New("Unable to interact with world state")
	}

	if existing == nil {
		return "", fmt.Errorf("Cannot read asset. Asset with id %s does not exist", assetID)
	}

	return string(existing), nil
}

func main() {
	sac := new(SimpleAsset)

	if err := contractapi.CreateNewChaincode(sac); err != nil {
		fmt.Printf("Error starting SimpleAsset chaincode: %s", err)
	}
}
```

### Building Chaincode
Now let's compile your chaincode.

```
go get -u github.com/hyperledger/fabric/core/chaincode/contractapi
go build
```

Assuming there are new errors, now you can proceed to the next step, testing your chaincode. 

### Testing Using dev mode
Normally chaincodes are started and maintained by peer. However in “dev mode”, chaincode is built and started by the user. This mode is useful during chaincode development phase for rapid code/build/run/debug cycle turnaround.

We start “dev mode” by leveraging pre-generated orderer and channel artifacts for a sample dev network. As such, the user can immediately jump into the process of compiling chaincode and driving calls.

#### Install Hyperledger Fabric Samples
If you haven’t already done so, please Install Samples, Binaries and Docker Images.

Create a new folder in the chaincode directory of fabric-samples named `simple_asset_contract`. Into this folder place the code you created above.

Navigate to the chaincode-docker-devmode directory of the fabric-samples clone:

```
cd chaincode-docker-devmode
```

Now open three terminals and navigate to your chaincode-docker-devmode directory in each.

#### Terminal 1 - Start the network
```
docker-compose -f docker-compose-simple.yaml up
```

#### Terminal 2 - Build & start the chaincode
```
docker exec -it chaincode bash
```

You should see the following:

```
root@d2629980e76b:/opt/gopath/src/chaincode#
```

Now, compile your chaincode:

```
cd simple_asset_contract
go build
```

Now run the chaincode:

```
CORE_PEER_ADDRESS=peer:7052 CORE_CHAINCODE_ID_NAME=mycc:0 ./simple_asset_contract
```

The chaincode is started with peer and chaincode logs indicating successful registration with the peer. Note that at this stage the chaincode is not associated with any channel. This is done in subsequent steps using the instantiate command.

#### Terminal 3 - Use the chaincode
Even though you are in --peer-chaincodedev mode, you still have to install the chaincode so the life-cycle system chaincode can go through its checks normally. This requirement may be removed in future when in --peer-chaincodedev mode.

We’ll leverage the CLI container to drive these calls.

```
docker exec -it cli bash
```

```
peer chaincode install -p chaincodedev/chaincode/simple_asset_contract -n mycc -v 0
```

When instantiating, invoking or querying a chaincode created using the contractapi we must pass as the first argument the namespace and function name of the contract we wish to call and then the arguments to pass in to the function. In the code we have set up there is only one contract used in the chaincode and we did not set a namespace therefore we use the default namespace `contract`. When instantiating we shall create a first instance of an asset in the world state with the ID `ASSET_1` using the create function.

```
peer chaincode instantiate -n mycc -v 0 -c '{"Args":["contract_Create","ASSET_1"]}' -C myc
```

Now we can issue an invoke to update the value of “ASSET_1” to “Updated”. Notice that when passing arguments they are evaluated left to right such that the first argument is the namespace and function to call, the second is then the first string parameter, the third the second string parameter and so on for further string parameters.

```
peer chaincode invoke -n mycc -c '{"Args":["contract_Update", "ASSET_1", "Updated"]}' -C myc
```

Finally, query ASSET_1. We should see a value of Updated.

```
peer chaincode query -n mycc -c '{"Args":["contract_Read","ASSET_1"]}' -C myc
```