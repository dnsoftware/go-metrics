// Package app реализует свой статический анализатор
package app

import (
	"golang.org/x/tools/go/analysis/multichecker"
)

func StaticCheckerRun() {

	mychecks := Analyzers()

	multichecker.Main(
		mychecks...,
	)

}
