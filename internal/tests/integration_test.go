package tests

import (
	"bytes"
	"os/exec"
	"testing"
)

const (
	testRedisHost = "localhost"
	testRedisPort = "6379"
	testAPIPort   = "8082"
	testSvcPort   = "8083"
)

func TestBonsaiCLIIntegration(t *testing.T) {
	// Start bonsaid instance
	instance, err := StartBonsai(testRedisHost, testRedisPort, testAPIPort, testSvcPort)
	if err != nil {
		t.Fatalf("Failed to start bonsaid: %v", err)
	}
	defer instance.Stop()

	// Test cases
	t.Run("Add URL", testAddURL)
	t.Run("Get URL", testGetURL)
	t.Run("List URLs", testListURLs)
	t.Run("Delete URL", testDeleteURL)
}

func testAddURL(t *testing.T) {
	cmd := exec.Command("bonsai-cli", "--api_host=localhost", "--api_port=8082", "add", "testkey", "http://example.com")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add URL: %v", out.String())
	}

	if !bytes.Contains(out.Bytes(), []byte("Successfully added")) {
		t.Fatalf("Unexpected output: %s", out.String())
	}
}

func testGetURL(t *testing.T) {
	cmd := exec.Command("bonsai-cli", "--api_host=localhost", "--api_port=8082", "get", "testkey")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to get URL: %v", out.String())
	}

	expected := "testkey -> http://example.com"
	if !bytes.Contains(out.Bytes(), []byte(expected)) {
		t.Fatalf("Unexpected output: %s", out.String())
	}
}

func testListURLs(t *testing.T) {
	cmd := exec.Command("bonsai-cli", "--api_host=localhost", "--api_port=8082", "list")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to list URLs: %v", out.String())
	}

	if !bytes.Contains(out.Bytes(), []byte("testkey -> http://example.com")) {
		t.Fatalf("Unexpected output: %s", out.String())
	}
}

func testDeleteURL(t *testing.T) {
	cmd := exec.Command("bonsai-cli", "--api_host=localhost", "--api_port=8082", "delete", "testkey")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to delete URL: %v", out.String())
	}

	if !bytes.Contains(out.Bytes(), []byte("Successfully deleted")) {
		t.Fatalf("Unexpected output: %s", out.String())
	}
}
