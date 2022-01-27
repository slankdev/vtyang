package vtyang

import (
	"encoding/json"
	"fmt"

	"github.com/k0kubun/pp"
)

var (
	x   = NewXPathOrDie
	mod = "account"
	_   = x
	_   = mod
	_   = pp.Println
)

func InitVTYang() {
	js := `
	{
		"users": {
			"user": [
				{
					"age": 26,
					"name": "hiroki",
					"projects": [
						{
							"finished": true,
							"name": "tennis"
						}
					]
				},
				{
					"age": 36,
					"name": "slankdev",
					"projects": [
						{
							"finished": false,
							"name": "kloudnfv"
						}
					]
				}
			]
		}
	}
	`

	m := map[string]interface{}{}
	err := json.Unmarshal([]byte(js), &m)
	ErrorOnDie(err)
	//pp.Println(m)
	n, err := Interface2DBNode(m)
	ErrorOnDie(err)
	fmt.Println(n.String())

	// ExecuteCommand("show running-config")
	// pp.Println(dbm.SetNode(mod, x(mod, "/users/user['name'='hiroki']/age"), "10"))
	// pp.Println(dbm.SetNode(mod, x(mod, "/users/user['name'='hiroki']/age"), "10"))
	// ExecuteCommand("set account users user taro age 10")
	// ExecuteCommand("show operational-data account users user taro")
	// err := dbm.DeleteNode(mod, x(mod, "/users/user['name'='taro']/age"))
	// ErrorOnDie(err)
	// ExecuteCommand("show operational-data account users user taro")

	// pp.Println(dbm.SetNode(mod, x(mod, "/users/user['name'='hoge']/age"), "10"))
	// pp.Println(dbm.SetNode(mod, x(mod, "/users/user['name'='hiroki']/age"), "10"))
	// pp.Println(dbm.SetNode(mod, x(mod, "/users/user['name'='yuta']/age"), "100"))
	// pp.Println(dbm.SetNode(mod, x(mod,
	// 	"/users/user['name'='hoge']/projects['name'='p1']/finished",
	// ), "true"))

	//dbm.SetNode("accounting", NewXPathOrDie("/users/user['name'='hoge']"), nil)
	//ExecuteCommand("show operational-data account users user hoge")

	//ExecuteCommand("show operational-data account users user hiroki")
	//ExecuteCommand("set account users user hoge")
	// ExecuteCommand("show operational-data account users user hoge")
	// ExecuteCommand("show operational-data account users user hoge")
	//ExecuteCommand("show operational-data account users user hiroki")
	// ExecuteCommand("show operational-data account users user hiroki age")

	// ExecuteCommand("set account users user hoge")
	// ExecuteCommand("show operational-data account users user hoge")

	// node, err := dbm.GetNode("account", "/users/user['name'='eva']")
	// if err != nil {
	// 	panic(err)
	// }
	// pp.Println(node)
}
