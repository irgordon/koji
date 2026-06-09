package system

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type cpuSample struct {
	idle  uint64
	total uint64
}

type Probe struct {
	mu            sync.Mutex
	lastCPUSample cpuSample
	lastUpdate    time.Time
}

func NewProbe() (*Probe, error) {
	p := &Probe{}
	initialSample, err := p.readCPUSample()
	if err != nil {
		return nil, fmt.Errorf("failed to prime CPU sample: %w", err)
	}
	p.lastCPUSample = initialSample
	p.lastUpdate = time.Now()
	return p, nil
}

func (p *Probe) Collect() (SystemMetrics, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	metrics := SystemMetrics{}

	if err := p.collectMemory(&metrics); err != nil {
		return metrics, err
	}

	_ = p.collectCPU(&metrics)
	_ = p.collectUptime(&metrics)

	return metrics, nil
}

func (p *Probe) collectMemory(m *SystemMetrics) error {
	total, available, err := parseMemInfo()
	if err != nil {
		return fmt.Errorf("memory parse failed: %w", err)
	}

	m.MemTotal = total
	m.MemAvailable = available
	m.MemUsed = total - available

	if total > 0 {
		m.MemUsagePct = float64(m.MemUsed) * 100.0 / float64(total)
	}

	return nil
}

func (p *Probe) collectCPU(m *SystemMetrics) error {
	current, err := p.readCPUSample()
	if err != nil {
		return err
	}

	idleDelta := current.idle - p.lastCPUSample.idle
	totalDelta := current.total - p.lastCPUSample.total

	if totalDelta > 0 {
		m.CPUUsage = (1.0 - float64(idleDelta)/float64(totalDelta)) * 100.0
	}

	p.lastCPUSample = current
	return nil
}

func (p *Probe) collectUptime(m *SystemMetrics) error {
	uptime, err := parseUptime()
	if err != nil {
		return err
	}
	m.Uptime = uptime
	return nil
}

func (p *Probe) readCPUSample() (cpuSample, error) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return cpuSample{}, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		return cpuSample{}, fmt.Errorf("empty /proc/stat")
	}

	line := scanner.Text()
	if !strings.HasPrefix(line, "cpu ") {
		return cpuSample{}, fmt.Errorf("unexpected /proc/stat format")
	}

	fields := strings.Fields(line)
	if len(fields) < 5 {
		return cpuSample{}, fmt.Errorf("insufficient cpu fields")
	}

	var total, idle uint64
	for i := 1; i < len(fields); i++ {
		val, err := strconv.ParseUint(fields[i], 10, 64)
		if err != nil {
			continue
		}
		total += val
		if i == 4 {
			idle = val
		}
	}

	return cpuSample{idle: idle, total: total}, nil
}
