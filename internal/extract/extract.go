package extract

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"log"
	"os"
	"strings"
	"sync"
)

const SDKPackagePrefix = "github.com/Azure/azure-sdk-for-go/services"

type Extract struct {
	endpointsMap map[Endpoint]struct{}
	mutex        sync.Mutex
}

func NewExtract() *Extract {
	return &Extract{
		endpointsMap: map[Endpoint]struct{}{},
	}
}

func (e *Extract) AddEndpoint(endpoint Endpoint) bool {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	if _, ok := e.endpointsMap[endpoint]; !ok {
		e.endpointsMap[endpoint] = struct{}{}
		return true
	}
	return false
}

func (e *Extract) Run(pass *analysis.Pass) (interface{}, error) {
	invoke := SDKInvoke{
		invokeMap: make(map[string]SDKPackageInvoke),
	}

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		be := n.(*ast.CallExpr)
		beFun, ok := be.Fun.(*ast.SelectorExpr)
		if !ok {
			return
		}
		b, ok := pass.TypesInfo.Types[beFun.X]
		if !ok {
			return
		}
		t, ok := b.Type.(*types.Pointer)
		if !ok {
			return
		}

		packagePath := t.Elem().(*types.Named).Obj().Pkg().Path()
		if !strings.HasPrefix(packagePath, SDKPackagePrefix) {
			return
		}
		typeName := t.Elem().(*types.Named).Obj().Name()
		funcName := beFun.Sel.Name

		invoke.addFuncInvoke(packagePath, typeName, funcName)
	})

	e.searchGoSdkPackage(&invoke)
	return nil, nil
}

func (e *Extract) searchGoSdkPackage(invoke *SDKInvoke) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	for k, v := range invoke.invokeMap {
		dir := cwd + "/vendor/" + k
		fset := token.NewFileSet()
		pkgs, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
		if err != nil {
			panic(err)
		}

		for _, pkg := range pkgs {
			for _, file := range pkg.Files {
				ast.Inspect(file, func(n ast.Node) bool {
					be, ok := n.(*ast.FuncDecl)
					if !ok {
						return true
					}
					if be.Recv == nil || len(be.Recv.List) != 1 {
						return true
					}
					recvType, ok := be.Recv.List[0].Type.(*ast.Ident)
					if !ok {
						return true
					}
					typeInvoke, ok := v.ClientInvokes[recvType.Name]
					if !ok {
						return true
					}
					if !strings.HasSuffix(be.Name.Name, "Preparer") {
						return true
					}
					FuncName := strings.TrimSuffix(be.Name.Name, "Preparer")
					if typeInvoke.FuncNames[FuncName] || typeInvoke.FuncNames[FuncName+"Complete"] {
						endpoint := getEndpointInfoFromGoSdkFunction(be.Body.List)
						if e.AddEndpoint(endpoint) {
							fmt.Println(endpoint.String())
						}
					}
					return true
				})
			}
		}
	}
}
