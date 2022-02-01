package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"

	fib "github.com/t-pwk/go-fibonacci"
)

func main() {
	colonyID := os.Getenv("COLONYID")
	runtimePrvKey := os.Getenv("RUNTIME_PRVKEY")
	host := os.Getenv("COLONIES_SERVER_HOST")
	portStr := os.Getenv("COLONIES_SERVER_PORT")

	port, err := strconv.Atoi(portStr)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Ask the Colonies server to assign a process to this Runtime
	client := client.CreateColoniesClient(host, port, true)
	for {
		assignedProcess, err := client.AssignProcess(colonyID, runtimePrvKey)
		if err != nil {
			fmt.Println(err)
			time.Sleep(1000 * time.Millisecond)
			continue
		}

		// Parse env attribute and calculate the given Fibonacci number
		for _, attribute := range assignedProcess.Attributes {
			if attribute.Key == "fibonacciNum" {
				nr, _ := strconv.Atoi(attribute.Value)
				fibonacci := fib.FibonacciBig(uint(nr))

				min := 100   // 0.1 s
				max := 40000 // 40s
				sleepTime := rand.Intn(max-min+1) + min

				fmt.Printf("sleeping for %d\n", sleepTime)

				time.Sleep(time.Duration(sleepTime) * time.Millisecond)

				attribute := core.CreateAttribute(assignedProcess.ID, core.OUT, "result", fibonacci.String())
				client.AddAttribute(attribute, runtimePrvKey)

				// Close the process as Successful
				client.CloseSuccessful(assignedProcess.ID, runtimePrvKey)
			}
		}
	}
}
