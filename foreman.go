package foreman

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Eslam-Nawara/foreman/internal/depgraph"
	parser "github.com/Eslam-Nawara/foreman/internal/procparser"
	"gopkg.in/yaml.v3"
)

const interval = 50 * time.Millisecond

type Foreman struct {
	services      map[string]parser.Service
	verbose       bool
	logger        *log.Logger
	servicesMutex sync.Mutex
}

func New(filePath string, verbose bool) (*Foreman, error) {
	foreman := &Foreman{
		services:      map[string]parser.Service{},
		servicesMutex: sync.Mutex{},
		logger:        log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime),
		verbose:       verbose,
	}

	yamlMap := make(map[string]map[string]any)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal([]byte(data), yamlMap)
	if err != nil {
		return nil, err
	}

	for serviceName, serviceInfo := range yamlMap {
		service := parser.ParseService(serviceInfo)
		service.Name = serviceName
		foreman.services[serviceName] = service
	}
	return foreman, nil
}

// start all services from yaml file
func (foreman *Foreman) Start() error {
	var wg sync.WaitGroup
	graph := foreman.buildDepGraph()

	if depgraph.IsCyclic(graph) {
		return errors.New("Cyclic dependencies detected")
	}

	services := depgraph.TopSort(graph)

	for _, serviceName := range services {
		wg.Add(1)
		go func(serviceName string) {
			defer wg.Done()
			foreman.runService(serviceName)
		}(serviceName)
	}

	go func() {
		sigChan := make(chan os.Signal)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		foreman.Exit(1)
	}()
	wg.Wait()
	return nil
}

// start one service and wait for it
func (foreman *Foreman) runService(serviceName string) {
	stopChecker := make(chan bool)
	for {
		foreman.servicesMutex.Lock()
		service := foreman.services[serviceName]
		foreman.servicesMutex.Unlock()

		cmd := exec.Command("bash", "-c", service.Cmd)
		cmd.Start()

		foreman.vLog(fmt.Sprintf("%s service has been started\n", serviceName))

		service.Process = cmd.Process

		foreman.servicesMutex.Lock()
		foreman.services[serviceName] = service
		foreman.servicesMutex.Unlock()

		go foreman.runChecker(service, stopChecker)

		cmd.Wait()

		stopChecker <- true
		service.Process = nil

		foreman.servicesMutex.Lock()
		foreman.services[serviceName] = service
		foreman.servicesMutex.Unlock()

		foreman.vLog(fmt.Sprintf("%s service exited with status code %s\n", serviceName, cmd.ProcessState.String()))
		if service.RunOnce {
			break
		}
	}
}

func (foreman *Foreman) runChecker(service parser.Service, stopChecker chan bool) {
	ticker := time.NewTicker(interval)
	foreman.vLog(fmt.Sprintf("%s checks has been started\n", service.Name))

	for {
		select {
		case <-stopChecker:
			foreman.vLog(fmt.Sprintf("%s checks stopped\n", service.Name))
			return

		case <-ticker.C:

			err := foreman.checkDeps(service)
			if err != nil {
				syscall.Kill(service.Process.Pid, syscall.SIGINT)
				foreman.vLog(fmt.Sprintf("dependencies check for %s failed, services has been restarted\n", service.Name))
			}

			checkCmd := exec.Command("bash", "-c", service.Checks.Cmd)
			err = checkCmd.Run()
			if err != nil {
				syscall.Kill(service.Process.Pid, syscall.SIGINT)
				foreman.vLog(fmt.Sprintf("check for process %s failed, services has been restarted\n", service.Name))
			}

			err = foreman.checkPort(service, "tcp")
			if err != nil {
				syscall.Kill(service.Process.Pid, syscall.SIGINT)
				foreman.vLog(fmt.Sprintf("tcp ports checking for process %s failed, services has been restarted\n", service.Name))
			}

			err = foreman.checkPort(service, "udp")
			if err != nil {
				syscall.Kill(service.Process.Pid, syscall.SIGINT)
				foreman.vLog(fmt.Sprintf("udp ports checking for process %s failed, services has been restarted\n", service.Name))
			}

		}
	}
}

func (foreman *Foreman) buildDepGraph() map[string][]string {
	graph := make(map[string][]string)
	for serviceName, service := range foreman.services {
		graph[serviceName] = service.Deps
	}
	return graph
}

func (f *Foreman) vLog(msg string) {
	if f.verbose {
		f.logger.Print(msg)
	}
}

// ُExit kills all the running services and checkers.
// exits foreman with the given exit status.
func (f *Foreman) Exit(exitStatus int) {
	f.servicesMutex.Lock()
	for _, service := range f.services {
		if service.Process != nil {
			syscall.Kill(-service.Process.Pid, syscall.SIGINT)
		}
	}
	f.servicesMutex.Unlock()
	os.Exit(exitStatus)
}

func (foreman *Foreman) checkDeps(service parser.Service) error {

	return nil
}
func (foreman *Foreman) checkPort(service parser.Service, portType string) error {

	return nil
}
