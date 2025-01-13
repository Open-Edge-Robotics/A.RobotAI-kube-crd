package v1

type CloudBridgePort struct {
	HostPort      map[string]int32 `json:"basePort,omitempty"`
	ContainerPort map[string]int32 `json:"containerPort,omitempty"`
}

func NewCloudBridgePort(port int32) CloudBridgePort {
	HostPort := map[string]int32{
		"HOST_PORT":     port,
		"HOST_SUB_PORT": port + 1,
		"HOST_PUB_PORT": port + 2,
		"HOST_REQ_PORT": port + 3,
		"HOST_REP_PORT": port + 4,
	}
	ContainerPort := map[string]int32{
		"container-s":  26565,
		"container-p1": 26566,
		"container-p2": 26567,
		"container-p3": 26568,
		"container-p4": 26569,
	}
	return CloudBridgePort{
		HostPort:      HostPort,
		ContainerPort: ContainerPort,
	}
}
