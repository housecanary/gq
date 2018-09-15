// Copyright 2018 HouseCanary, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/housecanary/gq/schema/structschema/gen"
)

func main() {
	var types = flag.String("types", "", "Types to include.  Comma separated.")
	var outputPackageName = flag.String("outputPackage", "", "Output package name")
	var outputPath = flag.String("outputPath", "", "Output path")
	var outputFileName = flag.String("outputFileName", "gen.go", "Output file name")
	flag.Parse()
	var packageNames = flag.Args()

	if types == nil || *types == "" {
		fmt.Println("types is required")
		return
	}
	if outputPackageName == nil || *outputPackageName == "" {
		fmt.Println("outputPackage is required")
		return
	}
	if outputPath == nil || *outputPath == "" {
		fmt.Println("outputPath is required")
		return
	}

	must(gen.Generate(
		*outputPath,
		*outputFileName,
		*outputPackageName,
		strings.Split(*types, ","),
		packageNames,
	))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
