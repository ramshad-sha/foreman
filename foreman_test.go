package foreman

import (
	"sync"
	"testing"

	"github.com/Eslam-Nawara/foreman/internal/depgraph"
	parser "github.com/Eslam-Nawara/foreman/internal/procparser"
)

const procfile = "./test-procfiles/Procfile"
const testProcfile = "./test-procfiles/Procfile-test"
const testBadProcfile = "./test-procfiles/Procfile-bad-test"
const testCyclicProcfile = "./test-procfiles/Procfile-cyclic-test"

func TestNew(t *testing.T) {
	t.Run("Parse existing procfile with correct syntax", func(t *testing.T) {
		want := Foreman{
			services:      map[string]parser.Service{},
			servicesMutex: sync.Mutex{},
		}
		sleeper := parser.Service{
			Name:    "sleeper",
			Process: nil,
			Cmd:     "sleep infinity",
			RunOnce: true,
			Deps:    []string{"hello"},
			Checks: parser.ServiceChecks{
				Cmd:      "ls",
				TcpPorts: []string{"4759", "1865"},
				UdpPorts: []string{"4500", "3957"},
			},
		}
		want.services["sleeper"] = sleeper

		hello := parser.Service{
			Name:    "hello",
			Process: nil,
			Cmd:     `echo "hello"`,
			RunOnce: true,
			Deps:    []string{},
		}
		want.services["hello"] = hello

		got, _ := New(testProcfile, false)

		assertForeman(t, got, &want)
	})

	t.Run("Run existing file with bad yml syntax", func(t *testing.T) {
		_, err := New(testBadProcfile, false)
		if err == nil {
			t.Error("Expected error: yaml: unmarshal errors")
		}
	})

	t.Run("Run non-existing file", func(t *testing.T) {
		_, err := New("unknown_file", false)
		want := "open unknown_file: no such file or directory"
		assertError(t, err, want)
	})
}

func TestBuildDepGraph(t *testing.T) {
	foreman, _ := New(procfile, false)

	got := foreman.BuildDepGraph()
	want := make(map[string][]string)
	want["service_ping"] = []string{"service_redis"}
	want["service_sleep"] = []string{"service_ping"}

	assertGraph(t, got, want)
}

func TestIsCyclic(t *testing.T) {
	t.Run("run cyclic graph", func(t *testing.T) {
		foreman, _ := New(testCyclicProcfile, false)
		graph := foreman.BuildDepGraph()
		got := depgraph.IsCyclic(graph)
		if !got {
			t.Error("got:true, want:false")
		}
	})

	t.Run("run acyclic graph", func(t *testing.T) {
		foreman, _ := New(testProcfile, false)
		graph := foreman.BuildDepGraph()
		got := depgraph.IsCyclic(graph)
		if got {
			t.Error("got:false, want:true")
		}
	})
}

func TestTopSort(t *testing.T) {
	foreman, _ := New(procfile, false)
	depGraph := foreman.BuildDepGraph()
	got := depgraph.TopSort(depGraph)
	want := []string{"service_redis", "service_ping", "service_sleep"}
	assertList(t, got, want)
}

// Helpers
func assertForeman(t *testing.T, got, want *Foreman) {
	t.Helper()

	for serviceName, service := range got.services {
		assertService(t, service, want.services[serviceName])
	}
}

func assertService(t *testing.T, got, want parser.Service) {
	t.Helper()

	if got.Name != want.Name {
		t.Errorf("got:\n%q\nwant:\n%q", got.Name, want.Name)
	}

	if got.Process != want.Process {
		t.Errorf("got:\n%v\nwant:\n%v", got.Process, want.Process)
	}

	if got.Cmd != want.Cmd {
		t.Errorf("got:\n%q\nwant:\n%q", got.Cmd, want.Cmd)
	}

	if got.RunOnce != want.RunOnce {
		t.Errorf("got:\n%t\nwant:\n%t", got.RunOnce, want.RunOnce)
	}

	assertList(t, got.Deps, want.Deps)
}

func assertChecks(t *testing.T, got, want parser.ServiceChecks) {
	t.Helper()

	if got.Cmd != want.Cmd {
		t.Errorf("got:\n%q\nwant:\n%q", got.Cmd, want.Cmd)
	}

	assertList(t, got.TcpPorts, want.TcpPorts)
	assertList(t, got.UdpPorts, want.UdpPorts)
}

func assertList(t *testing.T, got, want []string) {
	t.Helper()

	if len(want) != len(got) {
		t.Errorf("got:\n%v\nwant:\n%v", got, want)
	}

	n := len(want)
	for i := 0; i < n; i++ {
		if got[i] != want[i] {
			t.Errorf("got:\n%v\nwant:\n%v", got, want)
		}
	}
}

func assertError(t *testing.T, err error, want string) {
	t.Helper()

	if err == nil {
		t.Fatal("Expected Error: open unknown_file: no such file or directory")
	}

	if err.Error() != want {
		t.Errorf("got:\n%q\nwant:\n%q", err.Error(), want)
	}
}

func assertGraph(t *testing.T, got, want map[string][]string) {
	t.Helper()

	for key, value := range got {
		assertList(t, value, want[key])
	}
}
