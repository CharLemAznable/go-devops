package main

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/mitchellh/go-homedir"
	"github.com/valyala/fasttemplate"
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type LogFileArg struct {
	Options string
	LogPath string
	Ps      string
	Home    string
	Kill    string
	Start   string
	LogSeq  int
}

type LogFileInfoResult struct {
	MachineName  string
	Error        string
	LastModified string
	FileSize     string
	TailContent  string
	TailNextSeq  int
	CostTime     string
	ProcessInfo  string
}

type LogFileCommand int

func (t *LogFileCommand) TailFLog(args *LogFileArg, result *LogFileInfoResult) error {
	start := time.Now()
	tailContent, nextSeq := tail(args.LogPath, args.LogSeq)
	result.TailContent = string(tailContent)
	result.TailNextSeq = nextSeq

	result.CostTime = time.Since(start).String()
	return nil
}

func (t *LogFileCommand) RestartProcess(args *LogFileArg, result *LogFileInfoResult) error {
	start := time.Now()

	killTemplate := fasttemplate.New(args.Kill, "${", "}")
	killCommand := killTemplate.ExecuteString(map[string]interface{}{"ps": args.Ps})
	ExecuteCommands(killCommand, 500*time.Millisecond, true)

	argsHome, _ := homedir.Expand(args.Home)
	ExecuteCommands("cd "+argsHome+";"+args.Start, 1000*time.Millisecond, false)
	//randomShellName := RandStringBytesMaskImpr(16) + ".sh"
	//ExecuteCommands("cd "+args.Home+"\n"+
	//	"echo \""+args.Start+"\">"+randomShellName+"\n"+
	//	"chmod +x "+randomShellName+"\n"+
	//	"./"+randomShellName+"\n"+
	//	"rm "+randomShellName, 500*time.Millisecond)

	err := ""
	result.ProcessInfo, err = ExecuteCommands(args.Ps, 500*time.Millisecond, true)
	if err != "" {
		result.Error = err
	}

	result.CostTime = time.Since(start).String()
	return nil
}

func (t *LogFileCommand) TailLogFile(args *LogFileArg, result *LogFileInfoResult) error {
	start := time.Now()

	logPath, _ := homedir.Expand(args.LogPath)
	_, err := os.Stat(logPath)
	if err == nil {
		stdout, stderr := ExecuteCommands("tail "+args.Options+" "+logPath, 500*time.Millisecond, true)
		result.TailContent = stdout
		if stderr != "" {
			result.Error = stderr
		}
	} else {
		if os.IsNotExist(err) {
			result.Error = "Log file does not exist"
		} else {
			result.Error = err.Error()
		}
	}

	result.CostTime = time.Since(start).String()
	return nil
}

func (t *LogFileCommand) TruncateLogFile(args *LogFileArg, result *LogFileInfoResult) error {
	start := time.Now()

	logPath, _ := homedir.Expand(args.LogPath)
	_, err := os.Stat(logPath)
	if err == nil {
		ExecuteCommands("> "+logPath, 500*time.Millisecond, true)
		info, _ := os.Stat(logPath)

		result.FileSize = humanize.IBytes(uint64(info.Size()))
		result.LastModified = humanize.Time(info.ModTime())
	} else {
		if os.IsNotExist(err) {
			result.Error = "Log file does not exist"
		} else {
			result.Error = err.Error()
		}
	}

	result.CostTime = time.Since(start).String()
	return nil
}

func (t *LogFileCommand) LogFileInfo(args *LogFileArg, result *LogFileInfoResult) error {
	start := time.Now()

	if args.Ps != "" {
		result.ProcessInfo, _ = ExecuteCommands(args.Ps, 500*time.Millisecond, true)
		humanizedPsOutput(result)
	}

	logPath, _ := homedir.Expand(args.LogPath)
	info, err := os.Stat(logPath)
	if err == nil {
		result.FileSize = humanize.IBytes(uint64(info.Size()))
		result.LastModified = humanize.Time(info.ModTime())
	} else {
		if os.IsNotExist(err) {
			result.Error = "Log file does not exist"
		} else {
			result.Error = err.Error()
		}
	}

	result.CostTime = time.Since(start).String()
	return nil
}

func humanizedPsOutput(result *LogFileInfoResult) {
	fields := strings.Fields(result.ProcessInfo)
	if len(fields) < 6 {
		return
	}

	vszKib, _ := strconv.ParseUint(fields[4], 10, 64)
	vsz := humanize.IBytes(1024 * vszKib) // virtual memory usage of entire process (in KiB)
	vsz = strings.Replace(vsz, " ", "", 1)
	result.ProcessInfo = strings.Replace(result.ProcessInfo, fields[4], vsz, 1)

	rssKib, _ := strconv.ParseUint(fields[5], 10, 64)
	rss := humanize.IBytes(1024 * rssKib) // resident set size, the non-swapped physical memory that a task has used (in KiB)
	rss = strings.Replace(rss, " ", "", 1)
	result.ProcessInfo = strings.Replace(result.ProcessInfo, fields[5], rss, 1)
}

func CallLogFileCommand(wg *sync.WaitGroup, logMachineName string, log Log, resultChan chan LogFileInfoResult,
	funcName string, processConfigRequired bool, options string, logSeq int) {
	defer wg.Done()

	found := fullFindLogMachineName(log, logMachineName)

	if !found {
		logMachineName, found = prefixFindLogMachineName(log, logMachineName)
	}

	if !found {
		fmt.Println(logMachineName, "is unknown")
		return

	}

	machineName, machineAddress, errorMsg := parseLogMachineNameAndAddress(logMachineName)
	fmt.Println("funcName:", funcName, "machineName:", machineName, ",machineAddress:", machineAddress, ",errorMsg:", errorMsg)

	reply := LogFileInfoResult{
		MachineName: machineName,
		Error:       errorMsg,
	}

	if errorMsg != "" {
		resultChan <- reply
		return
	}

	process, ok := devopsConf.Processes[log.Process]
	if !ok {
		process = Process{Ps: log.Process}
	}

	if processConfigRequired && (process.Home == "" || process.Kill == "" || process.Start == "") {
		reply.Error = log.Path + " is not well configured"
		resultChan <- reply
		return
	}

	c := make(chan LogFileInfoResult, 1)

	go func() {
		err := DialAndCall(machineAddress, func(client *rpc.Client) error {
			arg := &LogFileArg{
				LogPath: log.Path,
				Ps:      process.Ps,
				Home:    process.Home,
				Kill:    process.Kill,
				Start:   process.Start,
				Options: options,
				LogSeq:  logSeq,
			}
			fmt.Println("machineAddress:", machineAddress, "call func: LogFileCommand.", funcName, ",arg:", arg)
			return client.Call("LogFileCommand."+funcName, arg, &reply)
		})
		if err != nil {
			reply.Error = err.Error()
		}

		fmt.Println("reply:", reply)
		c <- reply
	}()

	select {
	case result := <-c:
		resultChan <- result
	case <-time.After(1 * time.Second):
		reply.Error = "timeout"
		resultChan <- reply
	}
}
func prefixFindLogMachineName(log Log, logMachineName string) (string, bool) {
	for _, configLogMachineName := range log.Machines {
		if strings.Index(configLogMachineName, logMachineName+":") == 0 {
			return configLogMachineName, true
		}
	}

	return "", false
}

func fullFindLogMachineName(log Log, logMachineName string) bool {
	for _, configLogMachineName := range log.Machines {
		if configLogMachineName == logMachineName {
			return true
		}
	}
	return false
}
