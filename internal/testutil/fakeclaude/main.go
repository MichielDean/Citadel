// fakeclaude is a minimal fake agent binary used in integration tests to verify
// that the proc-based liveness check (proc.AgentAliveUnderPIDIn) correctly
// detects a running agent process. It simply sleeps until killed, producing a
// /proc/<pid>/cmdline whose argv[0] base is "claude".
package main

import "time"

func main() {
	time.Sleep(60 * time.Second)
}
