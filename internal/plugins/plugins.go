package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"bakshell/internal/config"

	lua "github.com/yuin/gopher-lua"
)

type Manager struct {
	L        *lua.LState
	execFn   *lua.LFunction
	promptFn *lua.LFunction
}

func New() *Manager {
	return &Manager{L: lua.NewState()}
}

func (m *Manager) Close() {
	m.L.Close()
}

func (m *Manager) LoadConfig(path string) (*config.Config, error) {
	cfg, err := config.LoadFromLua(m.L, path)
	if err != nil {
		return cfg, fmt.Errorf("loading config: %w", err)
	}
	return cfg, nil
}

func (m *Manager) LoadPlugins(pluginDir string, active []string) {
	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not read plugins directory: %v\n", err)
		return
	}

	activeSet := make(map[string]bool, len(active))
	for _, p := range active {
		activeSet[p] = true
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".lua") {
			continue
		}
		if !activeSet[entry.Name()] {
			continue
		}

		path := filepath.Join(pluginDir, entry.Name())
		if err := m.L.DoFile(path); err != nil {
			fmt.Fprintf(os.Stderr, "Error loading plugin %s: %v\n", entry.Name(), err)
			continue
		}
		fmt.Printf("Loaded plugin: %s\n", entry.Name())
	}

	if fn := m.L.GetGlobal("execute_command"); fn != lua.LNil {
		if f, ok := fn.(*lua.LFunction); ok {
			m.execFn = f
		}
	}
	if fn := m.L.GetGlobal("get_prompt"); fn != lua.LNil {
		if f, ok := fn.(*lua.LFunction); ok {
			m.promptFn = f
		}
	}
}

func (m *Manager) ExecuteCommand(args []string) bool {
	if m.execFn == nil {
		return false
	}

	tbl := m.L.NewTable()
	for i, a := range args {
		tbl.RawSetInt(i+1, lua.LString(a))
	}

	if err := m.L.CallByParam(lua.P{
		Fn:      m.execFn,
		NRet:    1,
		Protect: true,
	}, tbl); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing Lua command handler: %v\n", err)
		return false
	}

	ret := m.L.Get(-1)
	m.L.Pop(1)
	return ret == lua.LTrue
}

func (m *Manager) GetPrompt() string {
	if m.promptFn == nil {
		return ""
	}

	if err := m.L.CallByParam(lua.P{
		Fn:      m.promptFn,
		NRet:    1,
		Protect: true,
	}); err != nil {
		return ""
	}

	ret := m.L.Get(-1)
	m.L.Pop(1)
	if s, ok := ret.(lua.LString); ok {
		return string(s)
	}
	return ""
}

func (m *Manager) SetExitCode(code int) {
	var fn *lua.LFunction
	if v := m.L.GetGlobal("set_exit_code"); v != lua.LNil {
		fn, _ = v.(*lua.LFunction)
	}
	if fn == nil {
		return
	}
	_ = m.L.CallByParam(lua.P{Fn: fn, NRet: 0, Protect: true}, lua.LNumber(code))
}
