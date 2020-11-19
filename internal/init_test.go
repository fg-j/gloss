package internal_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestInternal(t *testing.T) {
	suite := spec.New("gloss/internal", spec.Report(report.Terminal{}))
	suite("TestIssue", testIssue)
	suite("TestAPIClient", testAPIClient)
	suite("TestRepository", testRepository)
	suite.Run(t)
}
