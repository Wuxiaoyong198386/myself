package cpu

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"
)

var (
	clkTck float64 = -1 // write once before read
	// pageSize float64 = -1 // write once before read
)

// procStat is the field of /proc/${pid}/stat，jiffies 除以 CLK_TCK 即为 秒
type procStat struct {
	// cutime    float64 // 该任务的所有处于 waited-for 状态的子进程在用户态运行的时间，单位为 jiffies
	// cstime    float64 // 该任务的所有处于 waited-for 状态的子进程在内核态运行的时间，单位为 jiffies
	utime     float64 // 该任务在用户态运行的时间，单位为 jiffies
	stime     float64 // 该任务在内核态运行的时间，单位为 jiffies
	startTime float64 // 从系统启动开始到该任务启动的时间间隔，单位为 jiffies
	rss       float64 // 该任务当前驻留物理地址空间的页数
	uptime    float64 // 系统启动时间，两次作差，可作为进程的运行时间
}

func init() {
	getClkTck()
}

func getClkTck() error {
	clkTckStr, _, err := runShellCmd(3, "/usr/bin/getconf", "CLK_TCK")
	if err != nil {
		return err
	}
	clkTckFloat64, err := strconv.ParseFloat(strings.TrimSpace(clkTckStr), 64)
	if err != nil {
		return err
	}
	clkTck = clkTckFloat64
	return nil
}

// GetProcessCPUPercent gets the cpu used percent of the process
// @pid: the pid of the process
// @interval: compare system CPU times elapsed before and after the interval (blocking),
// it must be greater than 0, unit: second
// reference: http://git.intra.weibo.com/adx/prometheus/-/blob/master/proc.md
func GetProcessCPUPercent(pid, interval int) (float64, error) {
	if interval <= 0 {
		return -1, errors.New("invalid interval, it must be a positive number")
	}
	if clkTck < 0 {
		return -1, errors.New("failed to run 'getconf CLK_TCK'")
	}
	statLast, err := parseProcStat(pid)
	if err != nil {
		return -1, err
	}
	time.Sleep(time.Duration(interval) * time.Second)
	statNow, err := parseProcStat(pid)
	if err != nil {
		return -1, err
	}

	deltaCPUTime := (statNow.utime - statLast.utime + statNow.stime - statLast.stime) / clkTck
	deltaRunTime := statNow.uptime - statLast.uptime
	if math.Abs(deltaRunTime) < 1e-5 {
		return -1, errors.New("invalid delta running time of the current process")
	}

	return deltaCPUTime / deltaRunTime * 100, nil
}

func parseProcStat(pid int) (procStat, error) {
	var stat procStat

	// read file
	uptimeFileBytes, err := ioutil.ReadFile(path.Join("/proc", "uptime"))
	if err != nil {
		return stat, fmt.Errorf("failed to read uptime file, err: %s", err.Error())
	}
	procStatFileBytes, err := ioutil.ReadFile(path.Join("/proc", strconv.Itoa(pid), "stat"))
	if err != nil {
		return stat, fmt.Errorf("failed to read proc file, err: %s", err.Error())
	}

	// parse /proc/uptime
	uptimeFields := strings.Split(string(uptimeFileBytes), " ")
	if len(uptimeFields) < 2 {
		return stat, errors.New("number of filed in uptime can not be less than 2")
	}
	uptimeFloat64, err := strconv.ParseFloat(uptimeFields[0], 64)
	if err != nil {
		return stat, fmt.Errorf("failed to parse uptime, err: %s", err.Error())
	}

	// parse /proc/[pid]/stat
	fields := strings.Split(string(procStatFileBytes), " ")
	if len(fields) < 52 {
		return stat, errors.New("number of filed in stat can not be less than 52")
	}
	utimeFloat64, err := strconv.ParseFloat(fields[13], 64)
	if err != nil {
		return stat, fmt.Errorf("failed to parse utime, err: %s", err.Error())
	}
	stimeFloat64, err := strconv.ParseFloat(fields[14], 64)
	if err != nil {
		return stat, fmt.Errorf("failed to parse stime, err: %s", err.Error())
	}
	startTimeFloat64, err := strconv.ParseFloat(fields[21], 64)
	if err != nil {
		return stat, fmt.Errorf("failed to parse start time, err: %s", err.Error())
	}
	rssFloat64, err := strconv.ParseFloat(fields[23], 64)
	if err != nil {
		return stat, fmt.Errorf("failed to parse rss, err: %s", err.Error())
	}

	stat = procStat{
		utime:     utimeFloat64,
		stime:     stimeFloat64,
		startTime: startTimeFloat64,
		rss:       rssFloat64,
		uptime:    uptimeFloat64,
	}
	return stat, nil
}

// runShellCmd runs shell command
func runShellCmd(timeout int, cmd string, arg ...string) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	cmdGo := exec.CommandContext(ctx, cmd, arg...)
	var stdOut, stdErr bytes.Buffer
	cmdGo.Stdout = &stdOut
	cmdGo.Stderr = &stdErr
	if err := cmdGo.Run(); err != nil {
		return "", "", fmt.Errorf("cmd string: %s, err: %s", cmdGo.String(), err.Error())
	}
	return stdOut.String(), stdErr.String(), nil
}
