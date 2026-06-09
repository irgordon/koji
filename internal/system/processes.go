package system

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type SystemMetrics struct {
	CPUUsage     float64 `json:"cpuUsage"`
	MemTotal     uint64  `json:"memTotal"`
	MemAvailable uint64  `json:"memAvailable"`
	MemUsed      uint64  `json:"memUsed"`
	MemUsagePct  float64 `json:"memUsagePct"`
	Uptime       uint64  `json:"uptime"`
}

type ProcessInfo struct {
	PID         int
	Name        string
	State       string
	UID         int
	PPID        int
	CPUUser     uint64
	CPUSystem   uint64
	RSS         int64
	MemoryPct   float64
	CommandLine string
}

var pageSize = int64(os.Getpagesize())

func ListProcesses() ([]ProcessInfo, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, fmt.Errorf("failed to scan /proc: %w", err)
	}

	totalMem, err := getTotalMemoryKB()
	if err != nil {
		totalMem = 1
	}

	processes := make([]ProcessInfo, 0, 128)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		proc, err := readProcess(pid, totalMem)
		if err != nil {
			continue
		}

		processes = append(processes, proc)
	}

	return processes, nil
}

func getTotalMemoryKB() (int64, error) {
	total, _, err := parseMemInfo()
	if err != nil {
		return 0, err
	}
	return int64(total), nil
}

func readProcess(pid int, totalMemKB int64) (ProcessInfo, error) {
	statPath := fmt.Sprintf("/proc/%d/stat", pid)
	f, err := os.Open(statPath)
	if err != nil {
		return ProcessInfo{}, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		return ProcessInfo{}, fmt.Errorf("empty stat")
	}
	line := scanner.Text()

	startIdx := strings.IndexByte(line, '(')
	endIdx := strings.LastIndexByte(line, ')')
	if startIdx == -1 || endIdx == -1 || endIdx <= startIdx {
		return ProcessInfo{}, fmt.Errorf("malformed stat block")
	}

	name := line[startIdx+1 : endIdx]
	tail := line[endIdx+1:]
	fields := strings.Fields(tail)
	if len(fields) < 13 {
		return ProcessInfo{}, fmt.Errorf("truncated stat fields")
	}

	state := fields[0]
	ppid, _ := strconv.Atoi(fields[1])
	utime, _ := strconv.ParseUint(fields[11], 10, 64)
	stime, _ := strconv.ParseUint(fields[12], 10, 64)

	rssPages := readRSSPages(pid)
	rssBytes := rssPages * pageSize

	var memPct float64
	if totalMemKB > 0 {
		memPct = (float64(rssBytes) / float64(totalMemKB*1024)) * 100.0
	}

	return ProcessInfo{
		PID:         pid,
		Name:        name,
		State:       state,
		UID:         readProcessUID(pid),
		PPID:        ppid,
		CPUUser:     utime,
		CPUSystem:   stime,
		RSS:         rssBytes,
		MemoryPct:   memPct,
		CommandLine: readProcessCommandLine(pid),
	}, nil
}

func readProcessUID(pid int) int {
	f, err := os.Open(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return -1
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Uid:") {
			return firstUIDValue(line)
		}
	}
	return -1
}

func firstUIDValue(line string) int {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return -1
	}
	uid, err := strconv.Atoi(fields[1])
	if err != nil {
		return -1
	}
	return uid
}

func readProcessCommandLine(pid int) string {
	content, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	if err != nil || len(content) == 0 {
		return ""
	}
	cmdline := strings.ReplaceAll(string(content), "\x00", " ")
	return strings.TrimSpace(cmdline)
}

func readRSSPages(pid int) int64 {
	f, err := os.Open(fmt.Sprintf("/proc/%d/statm", pid))
	if err != nil {
		return 0
	}
	defer f.Close()

	var rssPages int64
	_, _ = fmt.Fscanf(f, "%*d %d", &rssPages)
	return rssPages
}

func parseMemInfo() (uint64, uint64, error) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()

	var total, available uint64
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()
		if total > 0 && available > 0 {
			break
		}
		if strings.HasPrefix(line, "MemTotal:") {
			total = extractMemoryValue(line)
		} else if strings.HasPrefix(line, "MemAvailable:") {
			available = extractMemoryValue(line)
		}
	}

	if total == 0 || available == 0 {
		return 0, 0, fmt.Errorf("could not read memory bounds")
	}

	return total, available, nil
}

func extractMemoryValue(line string) uint64 {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return 0
	}
	val, _ := strconv.ParseUint(fields[1], 10, 64)
	return val
}

func parseUptime() (uint64, error) {
	f, err := os.Open("/proc/uptime")
	if err != nil {
		return 0, err
	}
	defer f.Close()

	var uptimeStr string
	_, err = fmt.Fscanf(f, "%s", &uptimeStr)
	if err != nil {
		return 0, err
	}

	if idx := strings.Index(uptimeStr, "."); idx != -1 {
		uptimeStr = uptimeStr[:idx]
	}

	return strconv.ParseUint(uptimeStr, 10, 64)
}
