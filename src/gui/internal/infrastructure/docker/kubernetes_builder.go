package docker

import (
	"fmt"
	"strings"

	"ui/internal/ports"
)

type kubernetesBuilder struct {
	manifests strings.Builder
}

func NewKubernetesBuilder() ports.KubernetesBuilder {
	return &kubernetesBuilder{}
}

// AddConfigMap creates a ConfigMap for all files (JSON, params, KML, etc.)
func (b *kubernetesBuilder) AddConfigMap(name string, files map[string]string) {
	fmt.Fprintf(&b.manifests, "---\napiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: %s\ndata:\n", name)
	for fileName, content := range files {
		fmt.Fprintf(&b.manifests, "  %s: |-\n", fileName)
		// Ensure proper indentation for multiline content
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			line = strings.TrimRight(line, "\r")
			if line == "" {
				fmt.Fprintf(&b.manifests, "\n")
			} else {
				fmt.Fprintf(&b.manifests, "    %s\n", line)
			}
		}
	}
	fmt.Fprintf(&b.manifests, "\n")
}

func (b *kubernetesBuilder) AddDeployment(name string, containers []ports.KubeContainer, volumes []ports.KubeVolume, nodeLabel string, hostNetwork bool) {
	fmt.Fprintf(&b.manifests, "---\napiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: %s\nspec:\n  replicas: 1\n  selector:\n    matchLabels:\n      app: %s\n  template:\n    metadata:\n      labels:\n        app: %s\n    spec:\n", name, name, name)

	if hostNetwork {
		fmt.Fprintf(&b.manifests, "      hostNetwork: true\n")
	}

	if nodeLabel != "" {
		fmt.Fprintf(&b.manifests, "      nodeSelector:\n        node: %s\n", nodeLabel)
	}
	fmt.Fprintf(&b.manifests, "      containers:\n")

	for _, c := range containers {
		fmt.Fprintf(&b.manifests, "      - name: %s\n        image: %s\n", c.Name, c.Image)
		
		if len(c.Command) > 0 {
			fmt.Fprintf(&b.manifests, "        command:\n")
			for _, cmd := range c.Command {
				fmt.Fprintf(&b.manifests, "        - %q\n", cmd)
			}
		}

		if len(c.Ports) > 0 {
			fmt.Fprintf(&b.manifests, "        ports:\n")
			for _, p := range c.Ports {
				protocol := "TCP"
				if p.Protocol != "" {
					protocol = p.Protocol
				}
				fmt.Fprintf(&b.manifests, "        - containerPort: %d\n          protocol: %s\n", p.ContainerPort, protocol)
			}
		}

		if len(c.Env) > 0 {
			fmt.Fprintf(&b.manifests, "        env:\n")
			for k, v := range c.Env {
				// We don't quote blindly if it's already structured, but for simple strings it's safe.
				// However, if the value contains newlines, %q handles it.
				fmt.Fprintf(&b.manifests, "        - name: %s\n          value: %q\n", k, v)
			}
		}

		if len(c.VolumeMounts) > 0 {
			fmt.Fprintf(&b.manifests, "        volumeMounts:\n")
			for _, vm := range c.VolumeMounts {
				fmt.Fprintf(&b.manifests, "        - name: %s\n          mountPath: %s\n", vm.Name, vm.MountPath)
				if vm.SubPath != "" {
					fmt.Fprintf(&b.manifests, "          subPath: %s\n", vm.SubPath)
				}
			}
		}
	}

	if len(volumes) > 0 {
		fmt.Fprintf(&b.manifests, "      volumes:\n")
		for _, v := range volumes {
			fmt.Fprintf(&b.manifests, "      - name: %s\n", v.Name)
			if v.ConfigMap != "" {
				fmt.Fprintf(&b.manifests, "        configMap:\n          name: %s\n", v.ConfigMap)
			} else if v.HostPath != "" {
				fmt.Fprintf(&b.manifests, "        hostPath:\n          path: %s\n", v.HostPath)
				if v.Type != "" {
					fmt.Fprintf(&b.manifests, "          type: %s\n", v.Type)
				}
			}
		}
	}

	fmt.Fprintf(&b.manifests, "\n")
}

func (b *kubernetesBuilder) AddService(name string, selector map[string]string, ports []ports.KubeServicePort, isLoadBalancer bool) {
	fmt.Fprintf(&b.manifests, "---\napiVersion: v1\nkind: Service\nmetadata:\n  name: %s\nspec:\n", name)
	if isLoadBalancer {
		fmt.Fprintf(&b.manifests, "  type: LoadBalancer\n")
	}
	fmt.Fprintf(&b.manifests, "  selector:\n")
	for k, v := range selector {
		fmt.Fprintf(&b.manifests, "    %s: %s\n", k, v)
	}
	fmt.Fprintf(&b.manifests, "  ports:\n")
	for _, p := range ports {
		protocol := "TCP"
		if p.Protocol != "" {
			protocol = p.Protocol
		}
		fmt.Fprintf(&b.manifests, "    - name: %s\n      port: %d\n      targetPort: %d\n      protocol: %s\n", p.Name, p.Port, p.TargetPort, protocol)
	}
	fmt.Fprintf(&b.manifests, "\n")
}

func (b *kubernetesBuilder) Build() string {
	return b.manifests.String()
}
