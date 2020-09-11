package main

import "fmt"

func abc() {
	txt := 3 + 2
	fmt.Printf("Txt: %s\n", txt)

	//dataSourceMap := make(map[string]string)
	//resourceMap := make(map[string]string)
	//inspect.Preorder(nodeFilter, func(n ast.Node) {
	//	be := n.(*ast.FuncDecl)
	//	if be.Name.Name != "SupportedDataSources" && be.Name.Name != "SupportedResources" {
	//		return
	//	}
	//	if len(be.Recv.List) != 1 {
	//		return
	//	}
	//	if be.Recv.List[0].Type.(*ast.Ident).Name != "Registration" {
	//		return
	//	}
	//	if len(be.Body.List) == 0 {
	//		return
	//	}
	//	returnStmt, ok := be.Body.List[len(be.Body.List) - 1].(*ast.ReturnStmt)
	//	if !ok {
	//		return
	//	}
	//	resultList := returnStmt.Results[0].(*ast.CompositeLit).Elts
	//	isResource := be.Name.Name == "SupportedResources"
	//	for _, v := range resultList {
	//		kv := v.(*ast.KeyValueExpr)
	//		key := kv.Key.(*ast.BasicLit).Value
	//		value := kv.Value.(*ast.CallExpr).Fun.(*ast.Ident).Name
	//		if isResource {
	//			resourceMap[key] = value
	//		} else {
	//			dataSourceMap[key] = value
	//		}
	//	}
	//	//pass.Reportf(be.Pos(), "integer addition found %q",
	//	//	render(pass.Fset, be))
	//})
	//
	//dataSourceFunc := make([]string, 0)
	//for _, v := range dataSourceMap {
	//	nodeFilter := []ast.Node{
	//		(*ast.FuncDecl)(nil),
	//	}
	//	inspect.Preorder(nodeFilter, func(n ast.Node) {
	//		be := n.(*ast.FuncDecl)
	//		if be.Name.Name != v {
	//			return
	//		}
	//		if len(be.Body.List) == 0 {
	//			return
	//		}
	//		returnStmt, ok := be.Body.List[len(be.Body.List) - 1].(*ast.ReturnStmt)
	//		if !ok {
	//			return
	//		}
	//		elements := returnStmt.Results[0].(*ast.UnaryExpr).X.(*ast.CompositeLit).Elts
	//		for _, v := range elements {
	//			kv := v.(*ast.KeyValueExpr)
	//			if kv.Key.(*ast.Ident).Name == "Read" {
	//				dataSourceFunc = append(dataSourceFunc, kv.Value.(*ast.Ident).Name)
	//				break
	//			}
	//		}
	//	})
	//}
	//
	//for _, v := range dataSourceFunc {
	//	nodeFilter := []ast.Node{
	//		(*ast.FuncDecl)(nil),
	//	}
	//	inspect.Preorder(nodeFilter, func(n ast.Node) {
	//		be := n.(*ast.FuncDecl)
	//		if be.Name.Name != v {
	//			return
	//		}
	//		if len(be.Body.List) == 0 {
	//			return
	//		}
	//	})
	//}
}
