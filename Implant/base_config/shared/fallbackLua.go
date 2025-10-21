//go:build !withLua

package shared

func DoLua(LuaStr string) bool {
	return false
}
