package extract

type ClientInvoke struct {
	TypeName  string
	FuncNames map[string]bool
}

type SDKPackageInvoke struct {
	PackageName   string
	ClientInvokes map[string]ClientInvoke
}

type SDKInvoke struct {
	invokeMap map[string]SDKPackageInvoke
}

func (invoke *SDKInvoke) addFuncInvoke(packagePath, typeName, funcName string) {
	if v, ok := invoke.invokeMap[packagePath]; !ok {
		invoke.invokeMap[packagePath] = SDKPackageInvoke{
			PackageName: packagePath,
			ClientInvokes: map[string]ClientInvoke{
				typeName: {
					TypeName: typeName,
					FuncNames: map[string]bool{
						funcName: true,
					},
				},
			},
		}
	} else {
		if v1, ok := v.ClientInvokes[typeName]; !ok {
			v.ClientInvokes[typeName] = ClientInvoke{
				TypeName: typeName,
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
}
