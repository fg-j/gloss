package main_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestCommands(t *testing.T) {
	suite := spec.New("gloss", spec.Report(report.Terminal{}))
	suite("TestResponseTimes", testResponseTimes)
	suite.Run(t)
}
