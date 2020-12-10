package main

import (
	"github.com/njucz/terraform-provider-azurerm-analysis/internal/extract"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	e := extract.NewExtract()
	myAnalyzer := &analysis.Analyzer{
		Name:     "endpointExtractor",
		Doc:      "extract all endpoints in azure go sdk invoked by azurerm provider",
		Run:      e.Run,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}
	singlechecker.Main(myAnalyzer)
}
