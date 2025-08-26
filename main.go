package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"

	godigauth "github.com/AYM1607/godig/pkg/auth"
	"github.com/AYM1607/godig/pkg/tunnel"
	godigtypes "github.com/AYM1607/godig/types"
	"github.com/mdp/qrterminal"
)

const (
	tunnelServerAddr = "godig.xyz:8080"
	localhostAddr    = "127.0.0.1"
)

func main() {

	// Create a channel to receive OS signals
	sigChan := make(chan os.Signal, 1)
	// Register the channel to receive specific signals
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	port, err := getFreePort()
	if err != nil {
		log.Fatalf("Failed to get free port: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	err = serveOpenCode(ctx, port)
	if err != nil {
		log.Fatal("Failed to start opencode")
	}

	link, auth, err := runProxy(ctx, port)
	if err != nil {
		log.Fatal(err)
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
	cmd := exec.CommandContext(ctx, "opencode", []string{"serve", "--port", strconv.Itoa(port)}...)

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
	apikey, err := godigauth.GetServerKey()
	if err != nil {
		return "", "", err
	}

	cli, err := tunnel.NewTunnelClient(
		tunnelServerAddr,
		localhostAddr+":"+strconv.Itoa(port),
		apikey,
		godigtypes.TunnelClientConfig{PersistConfig: true},
	)
	if err != nil {
		return "", "", err
	}

	go func() {
		cli.Run(ctx)
		log.Println("Tunnel exited")
	}()

	return fmt.Sprintf("https://%s.godig.xyz", cli.TunnelID), cli.Bearer, nil
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
