package main

import (
	"fmt"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
)

func systemInfo() {
	//runTimeOS := runtime.GOOS
	hostStat, _ := host.Info()
	fmt.Println("operating system", hostStat.OS)
	fmt.Println("platform ", hostStat.Platform)
	fmt.Println("hostname ", hostStat.Hostname)
	fmt.Println("num of processes", hostStat.Procs)

	memory, _ := mem.VirtualMemory()
	fmt.Println("memory", memory.Total)
	fmt.Println("memory ", memory.Free)
	fmt.Println("used ", memory.UsedPercent, "%")
}
func diskInfo() {
	diskStat, _ := disk.Usage("\\")
	fmt.Println("mem", diskStat.Total)
	fmt.Println("mem", diskStat.Free)
	fmt.Println("mem", diskStat.UsedPercent, "%")
}
func cpuInfo() {
	cpuStat, _ := cpu.Info()
	var index = 0
	for index < len(cpuStat) {
		fmt.Println("CPU [", index, "]", cpuStat[index].Cores)
		index++
	}
}
func main() {
	//hostStat, _ := host.Info()
	//a := Test1{1}
	//fmt.Println(hostStat.Hostname)
	systemInfo()
	diskInfo()
	cpuInfo()
}
