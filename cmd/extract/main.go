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
	"regexp"
	"strings"
	"sync"
)

var HttpActionMap = map[string]string{
	"AsDelete":  "DELETE",
	"AsGet":     "GET",
	"AsHead":    "HEAD",
	"AsMerge":   "MERGE",
	"AsOptions": "OPTIONS",
	"AsPatch":   "PATCH",
	"AsPost":    "POST",
	"AsPut":     "PUT",
}

var re = regexp.MustCompile(`/\{.*?\}`)

func formatUrl(url string) string {
	url = re.ReplaceAllString(url, "")
	if !strings.HasSuffix(url, "/") {
		url = url + "/"
	}
	return strings.ToUpper(url)
}

var urlUsage = map[string]struct{}{}
var urlUsageMutex = sync.RWMutex{}

type ClientInvoke struct {
	TypeName  string
	FuncNames map[string]bool
}

type SDKPackageInvoke struct {
	PackageName   string
	ClientInvokes map[string]ClientInvoke
}

var myAnalyzer = &analysis.Analyzer{
	Name:     "resourceUrlAnalyzer",
	Doc:      "get all azure sdk urls invoked by azurerm provider",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
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
						if url, action := walk(be.Body.List); url != "" && action != "" {
							fmt.Printf("%s %s\n", url, action)
						}
					}
					return true
				})
			}
		}
	}

	return nil, nil
}

func walk(body []ast.Stmt) (string, string) {
	httpActionPlusPath := make([]string, 0, len(HttpActionMap)+1)
	for k := range HttpActionMap {
		httpActionPlusPath = append(httpActionPlusPath, k)
	}
	httpActionPlusPath = append(httpActionPlusPath, "WithPathParameters", "WithPath")

	var url, action, apiVersion string
	for _, stmt := range body {
		ast.Inspect(stmt, func(n ast.Node) bool {
			if be, ok := n.(*ast.GenDecl); ok && be.Tok == token.CONST {
				if constExpr, ok := be.Specs[0].(*ast.ValueSpec); ok {
					if constExpr.Names[0].Name == "APIVersion" {
						apiVersion = constExpr.Values[0].(*ast.BasicLit).Value
						apiVersion = apiVersion[1 : len(apiVersion)-1]
					}
				}
				return true
			}

			be, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			selectBe, ok := be.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if !foundInList(httpActionPlusPath, selectBe.Sel.Name) {
				return true
			}
			xIdent, ok := selectBe.X.(*ast.Ident)
			if !ok {
				return true
			}
			if xIdent.Name != "autorest" {
				return true
			}
			if selectBe.Sel.Name == "WithPathParameters" || selectBe.Sel.Name == "WithPath" {
				if len(be.Args) == 0 {
					return false
				}
				args0, ok := be.Args[0].(*ast.BasicLit)
				if !ok {
					return false
				}
				url = args0.Value[1 : len(args0.Value)-1]
			} else {
				action = HttpActionMap[selectBe.Sel.Name]
			}
			return false
		})
	}

	url = formatUrl(url)
	key := fmt.Sprintf("%s-%s", url, action)
	urlUsageMutex.RLock()
	_, ok := urlUsage[key]
	urlUsageMutex.RUnlock()
	if ok {
		return "", ""
	} else {
		urlUsageMutex.Lock()
		urlUsage[key] = struct{}{}
		urlUsageMutex.Unlock()
	}

	return url, action
}

func foundInList(arr []string, ele string) bool {
	for _, v := range arr {
		if v == ele {
			return true
		}
	}
	return false
}

func main() {
	singlechecker.Main(myAnalyzer)
}
