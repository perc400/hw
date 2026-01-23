//go:build integration
// +build integration

package integration

import (
	"testing"

	. "github.com/onsi/ginkgo/v2" //nolint:revive,depguard
	. "github.com/onsi/gomega"    //nolint:revive,depguard
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}
