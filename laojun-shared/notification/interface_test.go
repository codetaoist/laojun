package notification

import (
	"context"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	config := DefaultConfig()
	
	impl := New(config)
	if impl == nil {
		t.Fatal("New() returned nil")
	}
	
	// TODO: Add more tests
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "invalid timeout",
			config: Config{
				Enabled: true,
				Timeout: -1,
			},
			wantErr: true,
		},
		// TODO: Add more test cases
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TODO: Add more tests for your interface methods
