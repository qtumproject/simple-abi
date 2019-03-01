## Qtum Simple ABI

This project is a CLI made for the purpose of generating non Solidity smart contracts via templates. Current supported languages are C with more to come.

### Example:
In order to create our smart contracts we need to create a `.abi` file. We'll create our own called `Coins.abi`. `Coins.abi` looks like the following:

```
# the name of the contract class/interface
:name=Coins
# addCoins is a function with inputs for number of coins received, a from address and a to address
# the output returned is the sum of coins
numCoins:uint64 to:uniaddress from:uniaddress addCoins:fn -> sumCoins:uint64
```

After that we call our CLI command from the same directory as `Coins.abi`. We call:

`simpleabi --abi Coins.abi --decode --encode`

This will generate a pair of files for decoding contract interactions and a pair of files for encodinng contract interactions specifically designed to interact with the qtum library. 