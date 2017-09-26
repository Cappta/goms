package goms

import "github.com/Cappta/gohelpgabs"

var (
	inputPath  = "input"
	outputPath = "output"
)

func inputOutputHandler(container *gohelpgabs.Container) {
	inputContainer := container.PopPath(inputPath)
	container.SetP(inputContainer.Data(), outputPath)
}
