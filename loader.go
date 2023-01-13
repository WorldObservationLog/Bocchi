package main

import (
	"github.com/dop251/goja"
	"path/filepath"
)

var packages = make(map[string]goja.Value)

func moduleTemplate(c string) string {
	return "(function(module, exports) {" + c + "\n})"
}

func createModule(c *goja.Runtime) *goja.Object {
	m := c.NewObject()
	e := c.NewObject()
	err := m.Set("exports", e)
	CheckErr(err)
	return m
}

func compileModule(p string) *goja.Program {
	code := ReadScript(p)
	text := moduleTemplate(code)
	prg, err := goja.Compile(p, text, true)
	CheckErr(err)
	return prg
}

func loadModule(c *Script, p string) goja.Value {
	p = filepath.Clean(p)
	pkg := packages[c.Name]
	if pkg != nil {
		return pkg
	}

	prg := compileModule(p)

	f, err := c.VM.RunProgram(prg)
	CheckErr(err)
	g, _ := goja.AssertFunction(f)

	m := createModule(c.VM)
	jsExports := m.Get("exports")
	_, err = g(jsExports, m, jsExports)
	CheckErr(err)

	return m.Get("exports")
}

// RegisterLoader register a simple commonjs style loader to runtime
func RegisterLoader(c *Script) {
	r := c.VM

	err := r.Set("require", func(call goja.FunctionCall) goja.Value {
		p := call.Argument(0).String()
		return loadModule(c, p)
	})
	CheckErr(err)
}
