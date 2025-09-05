package utils

import (
	"testing"
	"url-manager-system/backend/internal/db/models"
)

func TestValidateContainerConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  models.ContainerConfig
		wantErr bool
	}{
		{
			name: "valid container config",
			config: models.ContainerConfig{
				ContainerName: "my-app",
				WorkingDir:    "/app",
				Devices: []models.DeviceMapping{
					{
						HostPath:      "/dev/kvm",
						ContainerPath: "/dev/kvm",
						Permissions:   "rw",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid container name - uppercase",
			config: models.ContainerConfig{
				ContainerName: "My-App",
			},
			wantErr: true,
		},
		{
			name: "invalid container name - starts with dash",
			config: models.ContainerConfig{
				ContainerName: "-myapp",
			},
			wantErr: true,
		},
		{
			name: "invalid device permissions",
			config: models.ContainerConfig{
				Devices: []models.DeviceMapping{
					{
						HostPath:      "/dev/kvm",
						ContainerPath: "/dev/kvm",
						Permissions:   "invalid",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid working directory",
			config: models.ContainerConfig{
				WorkingDir: "relative/path",
			},
			wantErr: true,
		},
		{
			name: "sensitive device mapping",
			config: models.ContainerConfig{
				Devices: []models.DeviceMapping{
					{
						HostPath:      "/dev/mem",
						ContainerPath: "/dev/mem",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateContainerConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateContainerConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDeviceMapping(t *testing.T) {
	tests := []struct {
		name    string
		device  models.DeviceMapping
		wantErr bool
	}{
		{
			name: "valid device mapping",
			device: models.DeviceMapping{
				HostPath:      "/dev/kvm",
				ContainerPath: "/dev/kvm",
				Permissions:   "rw",
			},
			wantErr: false,
		},
		{
			name:    "empty host path",
			device:  models.DeviceMapping{},
			wantErr: true,
		},
		{
			name: "relative host path",
			device: models.DeviceMapping{
				HostPath:      "dev/kvm",
				ContainerPath: "/dev/kvm",
			},
			wantErr: true,
		},
		{
			name: "relative container path",
			device: models.DeviceMapping{
				HostPath:      "/dev/kvm",
				ContainerPath: "dev/kvm",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDeviceMapping(tt.device)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDeviceMapping() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
