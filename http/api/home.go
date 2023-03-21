package api

import (
	"fmt"
	addressFil "github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/chain/types/ethtypes"
	"github.com/gin-gonic/gin"
	"github.com/zhp12543/fil-address/config"
	"strings"
)

func Index(c *gin.Context)  {
	data := make(map[string]interface{})
	data["version"] = config.Version
	c.JSON(0, data)
}

func FilAddress(c *gin.Context)  {
	ret := make(map[string]string)
	ret["code"] = "error"

	defer func() {
		c.JSON(200, ret)
	}()

	addr, _ := c.GetQuery("address")
	addr = strings.TrimSpace(addr)
	if strings.EqualFold(addr, "") {
		ret["msg"] = "param address is empty"
		return
	}

	ethAddr := ""
	filAddr := ""
	if strings.HasPrefix(addr, "f0") || strings.HasPrefix(addr, "f4") {
		parseAddr, err := addressFil.NewFromString(addr)
		if err != nil {
			ret["msg"] = fmt.Sprintf("parse fil address:%+v err:%+v", addr, err.Error())
			return
		}
		parseEthAddr, err := ethtypes.EthAddressFromFilecoinAddress(parseAddr)
		if err != nil {
			ret["msg"] = fmt.Sprintf("parse fil address:%+v to eth address err:%+v", addr, err.Error())
			return
		}

		ethAddr = parseEthAddr.String()
		filAddr = addr
	} else if strings.HasPrefix(strings.ToLower(addr), "0x") {
		ethAddr = addr
		parseAddr, err := ethtypes.ParseEthAddress(addr)
		if err != nil {
			ret["msg"] = fmt.Sprintf("parse eth address:%+v err:%+v", addr, err.Error())
			return
		}

		parseFilAddr, err := parseAddr.ToFilecoinAddress()
		if err != nil {
			ret["msg"] = fmt.Sprintf("parse eth address:%+v to fil address err:%+v", addr, err.Error())
			return
		}

		filAddr = parseFilAddr.String()
	} else {
		ret["msg"] = fmt.Sprintf("unsupport address:%+v", addr)
		return
	}

	ret["code"] = "success"
	ret["eth_address"] = ethAddr
	ret["fil_address"] = filAddr
}