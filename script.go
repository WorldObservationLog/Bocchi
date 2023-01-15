package main

import (
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/dop251/goja"
)

var scripts []Script

type Script struct {
	Name        string
	Author      string
	Version     string
	Description string
	Matches     []string
	Path        string
	Priority    int
	VM          *goja.Runtime
}

func LoadScripts() {
	root := "./scripts/"
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if strings.HasSuffix(GetFilename(path), ".js") {
			scripts = append(scripts, InitialScript(path))
		}
		return nil
	})
	sort.Slice(scripts, func(i, j int) bool {
		return scripts[i].Priority > scripts[j].Priority
	})
	CheckErr(err)
}

func HandleResponse(resp http.Response) http.Response {
	returnResp := make(chan Response, 1)
	returnResp <- FromHttpResponse(resp)
	for _, script := range scripts {
		matched := MatchURL(script.Matches, resp.Request.URL.String())
		if matched {
			ExecResponseScript(script, <-returnResp, returnResp)
		}
	}
	return (<-returnResp).ToHttpResponse()
}

func ExecResponseScript(script Script, resp Response, returnResp chan Response) {
	var fn func(response Response) Response
	err := script.VM.ExportTo(script.VM.Get("OnResponse"), &fn)
	CheckErr(err)
	returnResp <- fn(resp)
}

func GetVM(name string, path string) *goja.Runtime {
	vm := goja.New()
	program, err := goja.Compile(name, ReadScript(path), true)
	CheckErr(err)
	_, err = vm.RunProgram(program)
	CheckErr(err)
	return vm
}

func ReadScript(path string) string {
	file, err := os.ReadFile(path)
	CheckErr(err)
	return string(file)
}

func GetFilename(path string) string {
	filename := filepath.Base(path)
	return filename
}

func HandleRequest(req http.Request) http.Request {
	returnReq := make(chan Request, 1)
	returnReq <- FromHttpRequest(req)
	for _, script := range scripts {
		matched := MatchURL(script.Matches, req.URL.String())
		if matched {
			ExecRequestScript(script, <-returnReq, returnReq)
		}
	}
	return (<-returnReq).ToHttpRequest()
}

func ExecRequestScript(script Script, req Request, returnReq chan Request) {
	var fn func(request Request) Request
	err := script.VM.ExportTo(script.VM.Get("OnRequest"), &fn)
	CheckErr(err)
	returnReq <- fn(req)
}

func MatchMetadata(pattern string, s string) []string {
	var matchStrings []string
	for _, i := range regexp.MustCompile(pattern).FindAllStringSubmatch(s, -1) {
		matchStrings = append(matchStrings, strings.TrimSuffix(i[1], "\r"))
	}
	return matchStrings
}

func InitialScript(path string) Script {
	content := ReadScript(path)
	startMatched, err := regexp.MatchString("// ==BocchiScript==", content)
	CheckErr(err)
	endMatched, err := regexp.MatchString("// ==/BocchiScript==", content)
	CheckErr(err)
	if startMatched && endMatched {
		name := MatchMetadata("//\\s*@name\\s*(.*)", content)[0]
		version := MatchMetadata("//\\s*@version\\s*(.*)", content)[0]
		description := MatchMetadata("//\\s*@description\\s*(.*)", content)[0]
		author := MatchMetadata("//\\s*@author\\s*(.*)", content)[0]
		matches := MatchMetadata("(?m)//\\s*@match\\s*(.*)", content)
		priority, err := strconv.Atoi(MatchMetadata("//\\s*@priority\\s*(.*)", content)[0])
		CheckErr(err)
		script := Script{
			Name:        name,
			Author:      author,
			Version:     version,
			Description: description,
			Path:        path,
			Matches:     matches,
			Priority:    priority,
		}
		script.VM = GetVM(name, path)
		InjectFunctions(&script)
		log.Printf("[INFo] LOoaded script: %s", name)
		return script
	}
	return Script{}
}

func MatchURL(matches []string, url string) bool {
	for _, i := range matches {
		if regexp.MustCompile(i).MatchString(url) {
			return true
		}
	}
	return false
}

func InjectFunctions(s *Script) {
	vm := s.VM
	err := vm.Set("ConventStringToBytes", ConventStringToBytes)
	CheckErr(err)
	err = vm.Set("LogInfo", log.Println)
	CheckErr(err)
	err = vm.Set("ConventBytesToString", ConventBytesToString)
	CheckErr(err)
	RegisterLoader(s)
}
