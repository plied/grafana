package datasources

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDataSource_IsTeamAllowed(t *testing.T) {
	tests := []struct {
		name         string
		allowedTeams string
		testTeamID   int64
		expected     bool
	}{
		{
			name:         "empty allowed teams allows all",
			allowedTeams: "",
			testTeamID:   1,
			expected:     true,
		},
		{
			name:         "single team match",
			allowedTeams: "1",
			testTeamID:   1,
			expected:     true,
		},
		{
			name:         "single team no match",
			allowedTeams: "1",
			testTeamID:   2,
			expected:     false,
		},
		{
			name:         "multiple teams match first",
			allowedTeams: "1,2",
			testTeamID:   1,
			expected:     true,
		},
		{
			name:         "multiple teams match second",
			allowedTeams: "1,2",
			testTeamID:   2,
			expected:     true,
		},
		{
			name:         "multiple teams no match",
			allowedTeams: "1,2",
			testTeamID:   3,
			expected:     false,
		},
		{
			name:         "teams with spaces",
			allowedTeams: "1, 2, 3",
			testTeamID:   2,
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := &DataSource{
				AllowedTeams: tt.allowedTeams,
			}
			
			result := ds.IsTeamAllowed(tt.testTeamID)
			assert.Equal(t, tt.expected, result)
		})
	}
}