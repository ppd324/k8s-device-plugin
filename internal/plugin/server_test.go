/**
# Copyright (c) 2022, NVIDIA CORPORATION.  All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
**/

package plugin

import (
	"testing"

	"github.com/stretchr/testify/require"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	v1 "github.com/NVIDIA/k8s-device-plugin/api/config/v1"
	"github.com/NVIDIA/k8s-device-plugin/internal/cdi"
	"github.com/NVIDIA/k8s-device-plugin/internal/imex"
)

func TestCDIAllocateResponse(t *testing.T) {
	testCases := []struct {
		description          string
		deviceIds            []string
		deviceListStrategies []string
		CDIPrefix            string
		GDSEnabled           bool
		MOFEDEnabled         bool
		imexChannels         []*imex.Channel
		expectedResponse     pluginapi.ContainerAllocateResponse
	}{
		{
			description:          "empty device list has empty response",
			deviceListStrategies: []string{"cdi-annotations"},
			CDIPrefix:            "cdi.k8s.io/",
		},
		{
			description:          "single device is added to annotations",
			deviceIds:            []string{"gpu0"},
			deviceListStrategies: []string{"cdi-annotations"},
			CDIPrefix:            "cdi.k8s.io/",
			expectedResponse: pluginapi.ContainerAllocateResponse{
				Annotations: map[string]string{
					"cdi.k8s.io/nvidia-device-plugin_uuid": "nvidia.com/gpu=gpu0",
				},
			},
		},
		{
			description:          "single device is added to annotations with custom prefix",
			deviceIds:            []string{"gpu0"},
			deviceListStrategies: []string{"cdi-annotations"},
			CDIPrefix:            "custom.cdi.k8s.io/",
			expectedResponse: pluginapi.ContainerAllocateResponse{
				Annotations: map[string]string{
					"custom.cdi.k8s.io/nvidia-device-plugin_uuid": "nvidia.com/gpu=gpu0",
				},
			},
		},
		{
			description:          "multiple devices are added to annotations",
			deviceIds:            []string{"gpu0", "gpu1"},
			deviceListStrategies: []string{"cdi-annotations"},
			CDIPrefix:            "cdi.k8s.io/",
			expectedResponse: pluginapi.ContainerAllocateResponse{
				Annotations: map[string]string{
					"cdi.k8s.io/nvidia-device-plugin_uuid": "nvidia.com/gpu=gpu0,nvidia.com/gpu=gpu1",
				},
			},
		},
		{
			description:          "multiple devices are added to annotations with custom prefix",
			deviceIds:            []string{"gpu0", "gpu1"},
			deviceListStrategies: []string{"cdi-annotations"},
			CDIPrefix:            "custom.cdi.k8s.io/",
			expectedResponse: pluginapi.ContainerAllocateResponse{
				Annotations: map[string]string{
					"custom.cdi.k8s.io/nvidia-device-plugin_uuid": "nvidia.com/gpu=gpu0,nvidia.com/gpu=gpu1",
				},
			},
		},
		{
			description:          "mofed devices are selected if configured",
			deviceListStrategies: []string{"cdi-annotations"},
			CDIPrefix:            "cdi.k8s.io/",
			MOFEDEnabled:         true,
			expectedResponse: pluginapi.ContainerAllocateResponse{
				Annotations: map[string]string{
					"cdi.k8s.io/nvidia-device-plugin_uuid": "nvidia.com/mofed=all",
				},
			},
		},
		{
			description:          "gds devices are selected if configured",
			deviceListStrategies: []string{"cdi-annotations"},
			CDIPrefix:            "cdi.k8s.io/",
			GDSEnabled:           true,
			expectedResponse: pluginapi.ContainerAllocateResponse{
				Annotations: map[string]string{
					"cdi.k8s.io/nvidia-device-plugin_uuid": "nvidia.com/gds=all",
				},
			},
		},
		{
			description:          "gds and mofed devices are included with device ids",
			deviceIds:            []string{"gpu0"},
			deviceListStrategies: []string{"cdi-annotations"},
			CDIPrefix:            "cdi.k8s.io/",
			GDSEnabled:           true,
			MOFEDEnabled:         true,
			expectedResponse: pluginapi.ContainerAllocateResponse{
				Annotations: map[string]string{
					"cdi.k8s.io/nvidia-device-plugin_uuid": "nvidia.com/gpu=gpu0,nvidia.com/gds=all,nvidia.com/mofed=all",
				},
			},
		},
		{
			description:          "imex channel is included with devices",
			deviceListStrategies: []string{"cdi-annotations"},
			CDIPrefix:            "cdi.k8s.io/",
			imexChannels:         []*imex.Channel{{ID: "0"}},
			expectedResponse: pluginapi.ContainerAllocateResponse{
				Annotations: map[string]string{
					"cdi.k8s.io/nvidia-device-plugin_uuid": "nvidia.com/imex-channel=0",
				},
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.description, func(t *testing.T) {
			deviceListStrategies, _ := v1.NewDeviceListStrategies(tc.deviceListStrategies)
			plugin := nvidiaDevicePlugin{
				config: &v1.Config{
					Flags: v1.Flags{
						CommandLineFlags: v1.CommandLineFlags{
							GDSEnabled:   &tc.GDSEnabled,
							MOFEDEnabled: &tc.MOFEDEnabled,
						},
					},
				},
				cdiHandler: &cdi.InterfaceMock{
					QualifiedNameFunc: func(c string, s string) string {
						return "nvidia.com/" + c + "=" + s
					},
				},
				deviceListStrategies: deviceListStrategies,
				cdiAnnotationPrefix:  tc.CDIPrefix,
				imexChannels:         tc.imexChannels,
			}

			response := pluginapi.ContainerAllocateResponse{}
			err := plugin.updateResponseForCDI(&response, "uuid", tc.deviceIds...)

			require.Nil(t, err)
			require.EqualValues(t, &tc.expectedResponse, &response)
		})
	}
}
