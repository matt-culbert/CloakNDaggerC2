//go:build withLua

package shared

import (
	si "github.com/yuin/gopher-lua"
)

func DoLua(LuaStr string) bool {
	L := si.NewState()
	L.OpenLibs()
	defer L.Close()
	if err := L.DoString(LuaStr); err != nil {
		//fmt.Println(err.Error())
		return false
	}
	return true
}
