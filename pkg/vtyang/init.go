package vtyang

import (
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
	// CMD := func(l string) { GetCommandNodeCurrent().ExecuteCommand(l) }
	// CMD("configure")
	// CMD("delete users user hiroki age")
	// CMD("commit")
	// CMD("set users user hiroki age 20")
	// CMD("commit")
	// CMD("set users user hiroki age 30")
	// CMD("commit")
	// CMD("do show configuration commit list")
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
