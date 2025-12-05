package ui

import (
	"bufio"
	"cdn-edge-server/internal/config"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
)

type CLI struct {
	reader *bufio.Reader
}

func Run() {
	cli := &CLI{
		reader: bufio.NewReader(os.Stdin),
	}

	cli.printBanner()
	cli.mainMenu()
}

func (c *CLI) printBanner() {
	fmt.Print(
		"   ____ ____  _   _   _____    _              ____                           \n" +
			"  / ___|  _ \\| \\ | | | ____|__| | __ _  ___  / ___|  ___ _ ____   _____ _ __ \n" +
			" | |   | | | |  \\| | |  _| / _` |/ _` |/ _ \\ \\___ \\ / _ \\ '__\\ \\ / / _ \\ '__|\n" +
			" | |___| |_| | |\\  | | |__| (_| | (_| |  __/  ___) |  __/ |   \\ V /  __/ |   \n" +
			"  \\____|____/|_| \\_| |_____\\__,_|\\__, |\\___| |____/ \\___|_|    \\_/ \\___|_|   \n" +
			"                                 |___/                                        \n\n")
}

func (c *CLI) mainMenu() {
	for {
		fmt.Println("\n┌─ Main Menu ─────────────────────────────┐")
		fmt.Println("│ 1. Check Server Status                  │")
		fmt.Println("│ 2. Send requests to edge server         │")
		fmt.Println("│ 3. View Configuration                   │")
		fmt.Println("│ 4. Exit                                 │")
		fmt.Println("└─────────────────────────────────────────┘")
		fmt.Print("\nSelect option: ")

		choice := c.readInput()

		switch choice {
		case "1":
			c.checkServerStatus()
		case "2":
			c.clientMenu()
		case "3":
			c.viewConfiguration()
		case "4":
			fmt.Println("\nCLI Exited")
			os.Exit(0)
		default:
			fmt.Println("Invalid option. Please try again.")
		}
	}
}

func (c *CLI) checkServerStatus() {
	fmt.Println("\nChecking Server Status...")
	fmt.Println("═══════════════════════════════════════")

	// Check edge server
	if conn, err := net.DialTimeout("tcp", config.EdgeHost+":"+config.EdgePort, time.Second); err == nil {
		conn.Close()
		fmt.Printf("Edge Server: Running on %s:%s\n", config.EdgeHost, config.EdgePort)
	} else {
		fmt.Printf("Edge Server: Not running\n")
		fmt.Println("   Start with: go run cmd/edge/main.go")
	}

	// Check origin server
	if conn, err := net.DialTimeout("tcp", config.OriginHost+":"+config.OriginPort, time.Second); err == nil {
		conn.Close()
		fmt.Printf("Origin Server: Running on %s:%s\n", config.OriginHost, config.OriginPort)
	} else {
		fmt.Printf("Origin Server: Not running\n")
		fmt.Println("   Start with: go run cmd/origin/main.go")
	}

	fmt.Println("═══════════════════════════════════════")
}

func (c *CLI) clientMenu() {
	// Check if edge server is running first
	if conn, err := net.DialTimeout("tcp", config.EdgeHost+":"+config.EdgePort, time.Second); err != nil {
		fmt.Println("\n  Edge server is not running!")
		fmt.Println("   Start with: go run cmd/edge/main.go")
		return
	} else {
		conn.Close()
	}

	for {
		fmt.Println("\n┌─ Send Requests ─────────────────────────┐")
		fmt.Println("│ 1. GET request                          │")
		fmt.Println("│ 2. HEAD request                         │")
		fmt.Println("│ 3. POST request (create file)           │")
		fmt.Println("│ 4. PUT request (update file)            │")
		fmt.Println("│ 5. Back to main menu                    │")
		fmt.Println("└─────────────────────────────────────────┘")
		fmt.Print("\nSelect option: ")

		choice := c.readInput()

		switch choice {
		case "1":
			c.sendGET()
		case "2":
			c.sendHEAD()
		case "3":
			c.sendPOST()
		case "4":
			c.sendPUT()
		case "5":
			return
		default:
			fmt.Println("Invalid option. Please try again.")
		}
	}
}

func (c *CLI) sendGET() {
	fmt.Print("\nEnter filename: ")
	filename := c.readInput()

	if filename == "" {
		fmt.Println("Filename cannot be empty")
		return
	}

	fmt.Printf("\n Sending GET request for '%s'...\n", filename)
	c.sendRequest("GET", filename, "")
}

func (c *CLI) sendHEAD() {
	fmt.Print("\nEnter filename: ")
	filename := c.readInput()

	if filename == "" {
		fmt.Println("Filename cannot be empty")
		return
	}

	fmt.Printf("\n Sending HEAD request for '%s'...\n", filename)
	c.sendRequest("HEAD", filename, "")
}

func (c *CLI) sendPOST() {
	fmt.Print("\nEnter filename: ")
	filename := c.readInput()

	if filename == "" {
		fmt.Println("Filename cannot be empty")
		return
	}

	fmt.Print("Enter body content (optional): ")
	body := c.readInput()

	fmt.Printf("\n Sending POST request to create '%s'...\n", filename)
	c.sendRequest("POST", filename, body)
}

func (c *CLI) sendPUT() {
	fmt.Print("\nEnter filename: ")
	filename := c.readInput()

	if filename == "" {
		fmt.Println("Filename cannot be empty")
		return
	}

	fmt.Print("Enter body content (optional): ")
	body := c.readInput()

	fmt.Printf("\n📤 Sending PUT request to update '%s'...\n", filename)
	c.sendRequest("PUT", filename, body)
}

func (c *CLI) sendRequest(method, filename, body string) {
	conn, err := net.Dial("tcp", config.EdgeHost+":"+config.EdgePort)
	if err != nil {
		fmt.Println("Error connecting to edge server:", err)
		return
	}
	defer conn.Close()

	// Construct HTTP request (headers)
	req := fmt.Sprintf("%s /%s HTTP/1.0\r\nHost: localhost\r\nContent-Length: %d\r\n\r\n",
		method, filename, len(body))

	// Send headers
	_, err = conn.Write([]byte(req))
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}

	// Send body if present
	if len(body) > 0 {
		_, err = conn.Write([]byte(body))
		if err != nil {
			fmt.Println("Error sending body:", err)
			return
		}
	}

	// Read response (read until connection closes)
	var response strings.Builder
	buf := make([]byte, 4096)

	for {
		n, err := conn.Read(buf)
		if n > 0 {
			response.Write(buf[:n])
		}
		if err != nil {
			if err == io.EOF {
				break // Normal connection close
			}
			fmt.Println("Error reading response:", err)
			return
		}
	}

	// Display response
	fmt.Println("\n Response:")
	fmt.Println("═══════════════════════════════════════")
	fmt.Println(response.String())
	fmt.Println("═══════════════════════════════════════")
}

func (c *CLI) viewConfiguration() {
	fmt.Println("\n  Configuration")
	fmt.Println("═══════════════════════════════════════")
	fmt.Printf("Edge Server:   %s:%s\n", config.EdgeHost, config.EdgePort)
	fmt.Printf("Origin Server: %s:%s\n", config.OriginHost, config.OriginPort)
	fmt.Printf("Cache Dir:     %s\n", config.CacheDir)
	fmt.Printf("Storage Dir:   %s\n", config.StorageDir)
}

func (c *CLI) readInput() string {
	input, _ := c.reader.ReadString('\n')
	return strings.TrimSpace(input)
}
