package main

import (
	"github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
	importer "github.com/luiz290788/covid19-brazil-importer"
)

func main() {
	funcframework.RegisterEventFunction("/", importer.ImportData)
	funcframework.Start("8080")
}
