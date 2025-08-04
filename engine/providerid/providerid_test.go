package providerid

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	for _, unit := range []struct {
		rawPID   string
		expected ProviderID
	}{
		{"FANZA:mdx0109", ProviderID{"FANZA", "mdx0109"}},
		{"FANZA:mdx0109:0.9", ProviderID{"FANZA", "mdx0109"}},
		{"FANZA:mdx0109:0", ProviderID{"FANZA", "mdx0109"}},
		{"FANZA:mdx0109:1", ProviderID{"FANZA", "mdx0109"}},
		{"FANZA:mdx0109:1.2", ProviderID{"FANZA", "mdx0109"}},
		{"AVBASE:dmm:ssis899", ProviderID{"AVBASE", "dmm:ssis899"}},
		{"AVBASE:dmm%3Assis899", ProviderID{"AVBASE", "dmm:ssis899"}},
		{"AVBASE:dmm%3assis899", ProviderID{"AVBASE", "dmm:ssis899"}},
		{"AVBASE:dmm:ssis899:0.99", ProviderID{"AVBASE", "dmm:ssis899"}},
		{"JavBus:HMN-095", ProviderID{"JavBus", "HMN-095"}},
		{"JavBus:HMN-095:0", ProviderID{"JavBus", "HMN-095"}},
		{"JavBus:HMN-095:0.90", ProviderID{"JavBus", "HMN-095"}},
		{"JavBus:HMN%2D095", ProviderID{"JavBus", "HMN-095"}},
	} {
		t.Run(unit.rawPID, func(t *testing.T) {
			pid, err := Parse(unit.rawPID)
			require.NoError(t, err)
			assert.Equal(t, unit.expected, pid)
		})
	}
}

func TestString(t *testing.T) {
	for _, unit := range []struct {
		pid      ProviderID
		expected string
	}{
		{ProviderID{"FANZA", "mdx0109"}, "FANZA:mdx0109"},
		{ProviderID{"AVBASE", "dmm:ssis899"}, "AVBASE:dmm%3Assis899"},
		{ProviderID{"JavBus", "HMN-095"}, "JavBus:HMN-095"},
	} {
		assert.Equal(t, unit.expected, unit.pid.String())
	}
}
