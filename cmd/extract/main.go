package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/singlechecker"
	"golang.org/x/tools/go/ast/inspector"
	"log"
	"os"
	"strings"

	"github.com/njucz/terraform-provider-azurerm-analysis/internal/extract"
)

type ClientInvoke struct {
	TypeName  string
	FuncNames map[string]bool
}

type SDKPackageInvoke struct {
	PackageName   string
	ClientInvokes map[string]ClientInvoke
}

func main() {
	myAnalyzer := &analysis.Analyzer{
		Name:     "endpointExtractor",
		Doc:      "extract all endpoints in azure go sdk invoked by azurerm provider",
		Run:      run,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}
	singlechecker.Main(myAnalyzer)
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	packageInvokesMap := map[string]SDKPackageInvoke{}
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
		if !strings.HasPrefix(packagePath, "github.com/Azure/azure-sdk-for-go/services") {
			return
		}
		TypeName := t.Elem().(*types.Named).Obj().Name()
		funcName := beFun.Sel.Name

		if v, ok := packageInvokesMap[packagePath]; !ok {
			packageInvokesMap[packagePath] = SDKPackageInvoke{
				PackageName: packagePath,
				ClientInvokes: map[string]ClientInvoke{
					TypeName: {
						TypeName: TypeName,
						FuncNames: map[string]bool{
							funcName: true,
						},
					},
				},
			}
		} else {
			if v1, ok := v.ClientInvokes[TypeName]; !ok {
				v.ClientInvokes[TypeName] = ClientInvoke{
					TypeName: TypeName,
					FuncNames: map[string]bool{
						funcName: true,
					},
				}
			} else {
				if !v1.FuncNames[funcName] {
					v1.FuncNames[funcName] = true
				}
			}
		}
	})

	cwd, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	for k, v := range packageInvokesMap {
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
						if endpoint := extract.EndpointInfoFromGoSdkFunction(be.Body.List); endpoint != nil {
							fmt.Println(endpoint.String())
						}
					}
					return true
				})
			}
		}
	}

	return nil, nil
}
