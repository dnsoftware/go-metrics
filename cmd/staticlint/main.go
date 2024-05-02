// Статический анализатор staticlint.
//
// Запускается из директории проекта командой: make staticlink.
//
// # Анализаторы
//
// • стандартные статические анализаторы пакета golang.org/x/tools/go/analysis/passes.
// • анализаторы staticcheck.io класса SA - выявляют ошибки и проблемы с производительностью.
// • анализаторы staticcheck.io класса quickfix - рефакторинг кода.
// • анализаторы staticcheck.io класса simple - упрощение кода.
// • анализаторы staticcheck.io класса tylecheck - 	содержит анализы, обеспечивающие соблюдение правил стиля.
//
// • анализатор bodyclose - проверяет, правильно ли закрыт res.Body.
// • анализатор ginkgolinter - Обеспечивает соблюдение стандартов использования ginkgo и gomega.
//
// • собственный анализатор noosexit - анализирует вызовы функции os.Exit в пакете main функции main(). При наличии выдает предупреждение.
package main

import (
	"github.com/dnsoftware/go-metrics/internal/staticlint/app"
)

func main() {

	app.StaticCheckerRun()

}
