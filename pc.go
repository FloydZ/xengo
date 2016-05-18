package models
//package main

import (
	//"fmt"
	//"github.com/revel/revel"
	linuxproc "github.com/c9s/goprocinfo/linux"							//zum Auslesen von /proc/...
	"github.com/pivotal-golang/bytefmt"
	"github.com/revel/revel"
	//"website/app/models"
	//"website/app/controllers"
	"errors"
	"strconv"
)

var tempDir = "/tmp"

// struktur die alle wichtigen dynamischen Werte beinhaltet, dH dauernd initalisieren
type PCDynamic struct {
	Id int						`id`

	MemFree string				`memfree`
	MemAvailable string			`memavail`
	MemUsed string				`memused`
	MemUsedProzent string		`memusedprozent`

	LoadAvg float64				`loadavg`
	NumProcess uint64			`numprocess`

}
// strukur die alle wichtigen statischen Werte beinhaltet, dH nur einmal initalisieren
type PCStatic struct {
	Id int					`id`

	NumCPUS	int
	NumVCORES int
	NumVCORESUsed int				//%TODO wird noch nicht richitg berechnent
	NumPhysicalCores int
	NumVCORESUsedProzent string

	MemTotal string
	MemUsed string

	NetDevices string
}
//Main Struktur, beinhaltet alles
type PCAll struct {
	Id int				`id`
	Cpuinfo *linuxproc.CPUInfo
	Cpustat *linuxproc.CPUStat
	Meminfo *linuxproc.MemInfo
	Diskstat []linuxproc.DiskStat
	Netstat *linuxproc.NetStat
	Netdev []linuxproc.NetworkStat

}
//Initalisiert Statischen Pc
func (y *PCStatic) PCInitStatic(s *SSH) error{
	/*
	NumCPUS	int
	NumVCORES int
	NumVCORESUsed int
	NumPhysicalCores int
	NumVCORESUsedProzent float64

	MemTotal string
	MemUsed string

	NetDevices string
	*/

	cpuinfo := &linuxproc.CPUInfo{}
	meminfo := &linuxproc.MemInfo{}
	netdev := []linuxproc.NetworkStat{}

	err := errors.New("")


	if s.Client == nil {
		err := errors.New("Could not find a ssh Client ")
		revel.ERROR.Print(err)
		return err
	}else {
		s.Download("/proc/cpuinfo", tempDir + "/cpuinfo" + strconv.Itoa(y.Id))
		cpuinfo, err = linuxproc.ReadCPUInfo(tempDir + "/cpuinfo" + strconv.Itoa(y.Id))
		if err != nil {
			revel.ERROR.Print(err)
			return err
		}
		s.Download("/proc/meminfo", tempDir + "/meminfo" + strconv.Itoa(y.Id))
		meminfo, err = linuxproc.ReadMemInfo( tempDir + "/meminfo" + strconv.Itoa(y.Id))
		if err != nil {
			revel.ERROR.Print(err)
			return err
		}
		s.Download("/proc/net/dev", tempDir + "/net" + strconv.Itoa(y.Id))
		netdev, err = linuxproc.ReadNetworkStat(tempDir + "/net" + strconv.Itoa(y.Id))
		if err != nil {
			revel.ERROR.Print(err)
			return err
		}
	}

	y.NumCPUS =  cpuinfo.NumCPU()
	y.NumPhysicalCores = cpuinfo.NumPhysicalCPU()
	y.NumVCORES = (cpuinfo.NumCore() +1) * 2
	y.NumVCORESUsed = y.NumVCORES/2 //%TODO fake und so


	a := float64(y.NumVCORESUsed)
	b := float64(y.NumVCORES)
	y.NumVCORESUsedProzent = strconv.FormatFloat(float64(a / b) * 100, 'f', -1, 64)

	y.NetDevices = netdev[0].Iface

	y.MemTotal  = bytefmt.ByteSize(meminfo.MemTotal * 1024)
	y.MemUsed = bytefmt.ByteSize((meminfo.MemTotal - meminfo.MemFree) * 1024)

	return nil
}
//Initalisiert Dynamischen Pc
func (y *PCDynamic) PCInitDynamic(s *SSH) error{
/*
	Id int						`id`

	MemFree string				`memfree`
	MemAvailable string			`memavail`
	MemUsed uint64				`memused`
	MemUsedProzent float64		`memusedprozent`

	LoadAvg float64				`loadavg`
	NumProcess uint64			`numprocess`
	*/
	loadavg := &linuxproc.LoadAvg{}
	meminfo := &linuxproc.MemInfo{}
	err := errors.New("")

	if s.Client == nil {
		err := errors.New("Could not find a ssh Client ")
		revel.ERROR.Print(err)
		return err
	}else{
		s.Download("/proc/loadavg", tempDir + "/loadavg" + strconv.Itoa(y.Id))
		loadavg, err = linuxproc.ReadLoadAvg("/proc/loadavg")
		if err != nil {
			revel.ERROR.Print(err)
			return err
		}

		y.LoadAvg = loadavg.Last1Min
		y.NumProcess = loadavg.ProcessRunning

		s.Download("/proc/meminfo", tempDir + "/meminfo" + strconv.Itoa(y.Id))
		meminfo, err = linuxproc.ReadMemInfo( "/home/pr0gramming/Dokumente/meminfo" + strconv.Itoa(y.Id))
		if err != nil {
			revel.ERROR.Print(err)
			return err
		}

	}

	y.MemUsed = bytefmt.ByteSize((meminfo.MemTotal - meminfo.MemFree) * 1024)
	y.MemFree = bytefmt.ByteSize(meminfo.MemFree * 1024)
	y.MemAvailable = bytefmt.ByteSize(meminfo.MemAvailable * 1024)

	a := float64(meminfo.MemTotal)
	b := float64(meminfo.MemFree)
	y.MemUsedProzent = strconv.FormatFloat(float64(((a - b) / a) * 100), 'f', -1, 64)

	//println("test: " + (y.MemUsedProzent))
	return nil
}

//Initalisiert Pc
func (x *PCAll) Pc_Init_All() error{
	cpuinfo, err := linuxproc.ReadCPUInfo("/proc/cpuinfo")
	if err != nil {
		revel.ERROR.Print(err)
		return err
	}
	meminfo, err := linuxproc.ReadMemInfo("/proc/meminfo")
	if err != nil {
		revel.ERROR.Print(err)
		return err
	}
	diskinfo, err := linuxproc.ReadDiskStats("/proc/diskstats")
	if err != nil {
		revel.ERROR.Print(err)
		return err
	}

	netstat, err := linuxproc.ReadNetStat("/proc/net/netstat")
	if err != nil {
		revel.ERROR.Print(err)
		return err
	}

	netdev, err := linuxproc.ReadNetworkStat("/proc/net/dev")
	if err != nil {
		revel.ERROR.Print(err)
		return err
	}

	x.Cpuinfo = cpuinfo
	x.Meminfo = meminfo
	x.Diskstat = diskinfo
	x.Netstat = netstat
	x.Netdev = netdev

	return nil
}

/*
// Aus Testzwecken noch drin
func main(){
	x := Pc{}
	x.read()
	println(x.cpuinfo.NumCPU())
	println(x.meminfo.Buffers)
}*/
