package main

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/mdp/qrterminal"
	"golang.ngrok.com/ngrok/v2"
)

func main() {

	// Create a channel to receive OS signals
	sigChan := make(chan os.Signal, 1)
	// Register the channel to receive specific signals
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	port, err := getFreePort()
	if err != nil {
		log.Fatal("Failed to get free port", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	err = serveOpenCode(ctx, port)
	if err != nil {
		log.Fatal("Failed to start opencode")
	}
	time.Sleep(time.Second * 5)

	link, auth, err := runProxy(ctx, port)
	if err != nil {
		log.Fatal("")
	}
	log.Printf("Listening on %q, with auth %q", link, auth)
	qrterminal.GenerateHalfBlock(fmt.Sprintf(`{"link": "%s", "auth": "%s"}`, link, auth), qrterminal.L, os.Stdout)

	go func() {
		for range sigChan {
			cancel()
			break
		}
	}()

	<-ctx.Done()
}

func serveOpenCode(ctx context.Context, port int) error {
	cmd := exec.CommandContext(ctx, "opencode", []string{"serve", "--port", strconv.FormatInt(int64(port), 10)}...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start program: %w", err)
	}

	go func() {
		err := cmd.Wait()
		if err != nil && ctx.Err() == nil {
			// Process exited with error (not due to cancellation)
			fmt.Printf("opencode exited with error: %v\n", err)
		}
	}()

	return nil
}

func runProxy(ctx context.Context, port int) (string, string, error) {
	username, password, err := generateCredentials()
	if err != nil {
		return "", "", err
	}

	agent, err := ngrok.NewAgent(ngrok.WithAuthtoken(os.Getenv("NGROK_AUTHTOKEN")))
	if err != nil {
		return "", "", err
	}

	trafficPolicy := fmt.Sprintf(`
inbound:
- name: "basic_auth"
  actions:
  - type: "basic-auth"
    config:
      credentials:
      - "%s:%s"
`, username, password)

	ln, err := agent.Forward(ctx,
		ngrok.WithUpstream("http://127.0.0.1:"+strconv.FormatInt(int64(port), 10)),
		ngrok.WithTrafficPolicy(trafficPolicy),
	)

	if err != nil {
		fmt.Println("Error", err)
		return "", "", err
	}

	go func() {
		<-ln.Done()
		fmt.Println("Done forwarding")
	}()

	fmt.Println("Endpoint online: forwarding from", ln.URL(), "to", port)

	return ln.URL().String(), fmt.Sprintf("%s:%s", username, password), nil
}

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()

	return l.Addr().(*net.TCPAddr).Port, nil
}

func generateCredentials() (string, string, error) {
	// Generate random bytes for username and password
	usernameBytes := make([]byte, 5) // 8 chars when base32 encoded
	passwordBytes := make([]byte, 8) // 13 chars when base32 encoded

	if _, err := rand.Read(usernameBytes); err != nil {
		return "", "", err
	}

	if _, err := rand.Read(passwordBytes); err != nil {
		return "", "", err
	}

	// Encode to base32 and remove padding
	username := strings.TrimRight(base32.StdEncoding.EncodeToString(usernameBytes), "=")
	password := strings.TrimRight(base32.StdEncoding.EncodeToString(passwordBytes), "=")

	return strings.ToLower(username), strings.ToLower(password), nil
}
