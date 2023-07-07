package abiencode

import (
	"bytes"
	"errors"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"strings"
)

var SendFilAbi abi.ABI

func init()  {
	SendFilAbi, _ = abi.JSON(strings.NewReader(`[{"inputs":[{"internalType":"bytes","name":"toAddress","type":"bytes"}],"name":"Send","outputs":[],"stateMutability":"payable","type":"function"}]`))
}

func ABISendFilEncode(addressHex []byte) ([]byte, error) {
	return SendFilAbi.Pack("Send", addressHex)
}

func ABISendFilDecode(data []byte) ([]byte, error) {
	method := SendFilAbi.Methods["Send"]
	methodId := data[:len(method.ID)]

	if bytes.Compare(methodId, method.ID) != 0 {
		return nil, errors.New("method not equal")
	}

	ret, err := SendFilAbi.Methods["Send"].Inputs.Unpack(data[len(methodId):])
	if err != nil {
		return nil, err
	}

	if len(ret) != 1 {
		return nil, errors.New("decode data len not 1")
	}
	return ret[0].([]byte), nil
}