package extract

import (
	"github.com/njucz/terraform-provider-azurerm-analysis/internal/common"
	"go/ast"
	"go/token"
)

const ApiVersionName = "APIVersion"

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

var UrlFuncName = map[string]struct{}{
	"WithPathParameters": {},
	"WithPath":           {},
}

func getEndpointInfoFromGoSdkFunction(funcBody []ast.Stmt) common.Endpoint {
	var endpoint common.Endpoint
	for _, stmt := range funcBody {
		ast.Inspect(stmt, func(n ast.Node) bool {
			if apiVersion, ok := checkAPIVersionAstNode(n); ok {
				endpoint.ApiVersion = apiVersion
				return false
			}
			if url, ok := checkUrlAstNode(n); ok {
				endpoint.Url = formatUrl(url)
				return false
			}
			if httpAction, ok := checkHttpMethodAstNode(n); ok {
				endpoint.HttpMethod = httpAction
				return false
			}
			if isManagementPlane, ok := checkIsManagementPlaneAstNode(n); ok {
				endpoint.IsManagementPlane = isManagementPlane
				return false
			}

			return true
		})
	}

	return endpoint
}

// the source code format should be the following format:
// ```const APIVersion = "2019-09-01"```
func checkAPIVersionAstNode(n ast.Node) (string, bool) {
	if be, ok := n.(*ast.GenDecl); ok && be.Tok == token.CONST {
		if constExpr, ok := be.Specs[0].(*ast.ValueSpec); ok {
			if constExpr.Names[0].Name == ApiVersionName {
				apiVersion := constExpr.Values[0].(*ast.BasicLit).Value
				return apiVersion[1 : len(apiVersion)-1], true
			}
		}
	}
	return "", false
}

// the source code format should be the following format:
// ```autorest.WithPathParameters("url", pathParameters)```
// or
// ```autorest.WithPath("url")```
func checkUrlAstNode(n ast.Node) (string, bool) {
	be, ok := n.(*ast.CallExpr)
	if !ok {
		return "", false
	}
	selectBe, ok := be.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", false
	}
	xIdent, ok := selectBe.X.(*ast.Ident)
	if !ok {
		return "", false
	}
	if xIdent.Name != "autorest" {
		return "", false
	}
	if _, ok := UrlFuncName[selectBe.Sel.Name]; ok {
		if len(be.Args) == 0 {
			return "", false
		}
		args0, ok := be.Args[0].(*ast.BasicLit)
		if !ok {
			return "", false
		}
		return args0.Value[1 : len(args0.Value)-1], true
	}
	return "", false
}

// the source code format should be the following format:
// ```autorest.AsPut()```
func checkHttpMethodAstNode(n ast.Node) (string, bool) {
	be, ok := n.(*ast.CallExpr)
	if !ok {
		return "", false
	}
	selectBe, ok := be.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", false
	}
	xIdent, ok := selectBe.X.(*ast.Ident)
	if !ok {
		return "", false
	}
	if xIdent.Name != "autorest" {
		return "", false
	}
	if action, ok := HttpActionMap[selectBe.Sel.Name]; ok {
		return action, true
	}
	return "", false
}

// the source code format should be the following format:
// ```autorest.WithBaseURL(client.BaseURI)```
// or
// ```autorest.WithCustomBaseURL```
func checkIsManagementPlaneAstNode(n ast.Node) (bool, bool) {
	be, ok := n.(*ast.CallExpr)
	if !ok {
		return false, false
	}
	selectBe, ok := be.Fun.(*ast.SelectorExpr)
	if !ok {
		return false, false
	}
	xIdent, ok := selectBe.X.(*ast.Ident)
	if !ok {
		return false, false
	}
	if xIdent.Name != "autorest" {
		return false, false
	}
	if selectBe.Sel.Name == "WithBaseURL" {
		return true, true
	} else if selectBe.Sel.Name == "WithCustomBaseURL" {
		return false, true
	}
	return false, false
}
