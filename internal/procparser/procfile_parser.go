package procparser

import (
	"fmt"
	"os"
)

type (
	Service struct {
		Name    string
		Status  bool
		Process *os.Process
		Cmd     string
		RunOnce bool
		Checks  ServiceChecks
		Deps    []string
	}

	ServiceChecks struct {
		Cmd      string
		TcpPorts []string
		UdpPorts []string
	}
)

func ParseService(serviceMap map[string]any) Service {
	service := Service{}

	for key, value := range serviceMap {
		switch key {
		case "cmd":
			service.cmd = value.(string)
		case "run_once":
			service.RunOnce = value.(bool)
		case "deps":
			for _, dep := range value.([]any) {
				service.Deps = append(service.Deps, dep.(string))
			}
		case "checks":
			service.Checks = parseChecks(value)
		}
	}
	return service
}

func parseChecks(serviceChecks any) ServiceChecks {
	checksMap := ServiceChecks{}

	for key, value := range serviceChecks.(map[string]any) {
		switch key {
		case "cmd":
			checksMap.Cmd = value.(string)
		case "tcp_ports":
			checksMap.TcpPorts = parsePorts(value)
		case "udp_ports":
			checksMap.UdpPorts = parsePorts(value)
		}
	}
	return checksMap
}

func parsePorts(ports any) []string {
	var parsedPorts []string
	for _, port := range ports.([]any) {
		parsedPorts = append(parsedPorts, fmt.Sprint(port.(int)))
	}

	return parsedPorts
}
