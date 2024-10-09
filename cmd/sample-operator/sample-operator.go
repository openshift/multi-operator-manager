package main

import (
	"fmt"
	sample_operator "github.com/openshift/multi-operator-manager/pkg/cmd/multi-operator-manager/sample-operator"
	"k8s.io/component-base/cli/globalflag"
	"k8s.io/component-base/version/verflag"
	"math/rand"
	"os"
	"time"

	"github.com/openshift/library-go/pkg/serviceability"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	utilflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/logs"
	_ "k8s.io/klog/v2"
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	rand.Seed(time.Now().UTC().UnixNano())

	pflag.CommandLine.SetNormalizeFunc(utilflag.WordSepNormalizeFunc)
	//pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)

	ioStreams := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	root := sample_operator.NewSampleOperatorCommand(ioStreams)
	verflag.AddFlags(root.Flags())
	globalflag.AddGlobalFlags(root.Flags(), root.Name(), logs.SkipLoggingConfigurationFlags())

	if err := func() error {
		defer serviceability.Profile(os.Getenv("OPENSHIFT_PROFILE")).Stop()
		return root.Execute()
	}(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
