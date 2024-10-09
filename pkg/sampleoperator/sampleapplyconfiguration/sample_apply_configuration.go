package sampleapplyconfiguration

import (
	"context"
	"time"

	"github.com/openshift/multi-operator-manager/pkg/library/libraryapplyconfiguration"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

func SampleRunApplyConfiguration(ctx context.Context, inputDirectory string, now time.Time, streams genericiooptions.IOStreams) (*libraryapplyconfiguration.ApplyConfiguration, error) {
	// TODO wire up the must-gather reading client
	ret := &libraryapplyconfiguration.ApplyConfiguration{
		DesiredConfigurationCluster: &libraryapplyconfiguration.SimpleClusterApplyResult{
			ClusterType: libraryapplyconfiguration.ClusterTypeConfiguration,
		},
		DesiredManagementCluster: &libraryapplyconfiguration.SimpleClusterApplyResult{
			ClusterType: libraryapplyconfiguration.ClusterTypeManagement,
		},
		DesiredUserWorkloadCluster: &libraryapplyconfiguration.SimpleClusterApplyResult{
			ClusterType: libraryapplyconfiguration.ClusterTypeUserWorkload,
		},
	}

	return ret, nil
}
