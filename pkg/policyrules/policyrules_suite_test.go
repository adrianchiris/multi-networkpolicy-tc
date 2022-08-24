package policyrules_test

import (
	"flag"
	"testing"

	"k8s.io/klog/v2"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "policyrules")
}

var _ = BeforeSuite(func() {
	fs := flag.NewFlagSet("test-flag-set", flag.PanicOnError)
	klog.InitFlags(fs)
	Expect(fs.Set("v", "8")).ToNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	klog.Flush()
})
