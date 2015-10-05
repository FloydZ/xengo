package main

import (
    "os/exec"
    "os"
    "strconv"
    "bufio"
    "io"
    "strings"
    "github.com/op/go-logging"
	"bytes"
	"fmt"
)


var NormalHddSaveDir = "/hdd/"
var NormalIsoDir = "/home/duda/xen/"
var log = logging.MustGetLogger("test")
var format = logging.MustStringFormatter("%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}", )

func closeFile(f *os.File, s string){
    log.Debug("debug %s", s)
    f.Close()
}

type xen_config interface {
    generate() bool;
    generate_from() bool;
    create_VM() bool;
    pause_VM() bool;
    unpause_VM() bool;
    restart_VM() bool;
}

type xen_config_struct struct {
    Template_OS string

    Mac string
    Vcores int
    Mem int64


    KVM bool							//paravm oder nich?
    Name string							//name der vm


    Vnc bool							//vnc enable?
    Vnc_ip string
    Vnc_nr int

	//%TODO multiple HDD
    Dir_disk string							//path to hdd to boot
	Sizeof_disk uint64
	Numof_disk int

    Dir_cfg string							//dir der cfg + name als id
    Id int 							//uid der vm und gleichzeitig name der der cfg
}

type xen_domU struct {
    Domain string
    Id int
    Mem int64
    vcores int
    state string
    x xen_config_struct
}

func (x xen_config_struct) Generate_cfg() bool {
    if (x.Dir_cfg == "") {
		log.Error("XEN CONFIG Not Set")
		return false
	}

    f, e := os.Create(x.Dir_cfg)
    if e != nil {
        log.Error("OPEN XEN CONFIG")
        return false
    }


    defer closeFile(f, "generate")


    f.WriteString("name='" + (x.Name) + "'\n")
    if (x.KVM == true) {
        f.WriteString("builder='hvm'\n")
    }else {

    }
    f.WriteString("vcpus=" + strconv.Itoa(x.Vcores) + "\n")
    f.WriteString("vif=[ 'mac=" + (x.Mac) + ",bridge=xenbr0' ]\n")

    str := strconv.FormatInt(x.Mem, 10)
    f.WriteString("memory=" + str + "\n")


	if (x.Numof_disk == 0){
		x.Numof_disk = 1
	}


	var disks bytes.Buffer
	disks.WriteString(fmt.Sprint("disk = [ 'file:"))

	for i := 0; i < x.Numof_disk; i++{
		disks.WriteString(fmt.Sprint(x.Dir_disk ,",sda1,w' "))
	}

    disks.WriteString(fmt.Sprint(",'file:", NormalIsoDir, x.Template_OS,".iso,hdc:cdrom,r' ]"))
	f.WriteString(disks.String() + "\n")


    if (x.Vnc == true){
        f.WriteString("vnc=1\n")
        f.WriteString("vnclisten='" + x.Vnc_ip + "'\n")
        f.WriteString("vndisplay=" + strconv.Itoa(x.Vnc_nr) + "\n")
    }

    f.WriteString("on_poweroff = 'destroy'\n")
    f.WriteString("on_reboot = 'restart'\n")
    f.WriteString("on_crash = 'restart'\n")
    f.Sync()
    return true
}

func (x xen_config_struct) Generate_from_cfg() bool{
	if (x.Dir_cfg == "") {
		log.Error("XEN CONFIG from Not Set")
		return false
	}

    f, e := os.Open(x.Dir_cfg)
    if e != nil {
        log.Error("OPEN XEN CONFIG FROM")
        return false
    }

    defer closeFile(f, "generate_from")

    bf := bufio.NewReader(f)

    for {
        line, isPrefix, e := bf.ReadLine()
        if e == io.EOF {
            break
        }
        if e != nil {
            log.Error("READ LINE")
            return false
        }
        if isPrefix {
            log.Error("READ LINE PREFIX")
            return false
        }

        if strings.Index(string(line), "=") != -1 {
            arr := strings.Split(string(line), "=")
            switch arr[0] {
                case "name":
                    {x.Name = arr[1]}
                case "builder":
                    x.KVM = true
                case "memory":
                    i, e := strconv.ParseInt(arr[1], 10, 64)
                    if e != nil {
                        log.Error("MEM Convert StoI")
                    }
                    x.Mem = i
                case "vnc":
                    x.Vnc = true
                case "vncdisplay":
                    i, e := strconv.Atoi(arr[1])
                    if e != nil {
                        log.Error("VNCDISPLAY Convert StoI")
                    }
                    x.Vnc_nr = i
                case "vnclisten":
                    x.Vnc_ip = arr[1]
                case "vcpus":
                    i, e := strconv.Atoi(arr[1])
                    if e != nil {
                        log.Error("VCPUS Convert StoI")
                    }
                    x.Vcores = i
                case "vif":
                    x.Mac = arr[2]
            }
        }


    }
    return true
}

func (x xen_domU) Create_VM_Wrapper() bool {
	y := xen_config_struct{
		Sizeof_disk: 1000,	//grob n gig
		Vcores: 1,
		Mem: 2000, //2GB
		Vnc:true,
		Vnc_nr:1,
		Vnc_ip:"0.0.0.0",
		Mac:"00:16:3e:XX:XX:XX",
		KVM:true,
		Template_OS:"debian",
		Name:"nottest",
		Dir_cfg:"/home/duda/xen/arch.cfg",
		Numof_disk:1}

	hdd_dir := CreateFileAsHdd(y.Name, y.Sizeof_disk)
	y.Dir_disk = hdd_dir

	if (y.Generate_cfg() == false){
		return false
	}
	return  x.Create_VM(y)
}

func (x xen_domU) Create_VM(y xen_config_struct) bool {
    ret1, err := exec.Command("xl", "create", y.Dir_cfg).CombinedOutput()
    if err != nil {
        log.Error("Error create VM")
        println(string(ret1))
        return false
    }

    ret := y.GetIdFromName()
    if ret == false {
        log.Error("Error create VM get Id from name")
        return false
    }

    x.Domain = y.Name
    x.Mem = y.Mem
    x.vcores = y.Vcores
    x.x = y
    x.Id = y.Id
    return true
}

func (x xen_domU) Pause_VM() bool{

    _, err:= exec.Command("xl", "pause", x.Domain).CombinedOutput()
    if err != nil {
        println("error occured pause vm")
        return  false
    }
    return true
}

func (x xen_domU) Unpause_VM() bool{

    _, err:= exec.Command("xl", "pause", x.Domain).CombinedOutput()
    if err != nil {
        println("error occured unpause vm")
        return  false
    }
    return true
}

func (x xen_domU) Restart_VM() bool{

    _, err:= exec.Command("xl", "reboot", x.Domain).CombinedOutput()
    if err != nil {
        println("error occured pause vm")
        return  false
    }
    return true
}

func (x xen_domU) Shutdown_VM() bool{

    _, err:= exec.Command("xl", "shutdown", x.Domain).CombinedOutput()
    if err != nil {
        println("error occured pause vm")
        return  false
    }
    return true
}

func (x xen_domU) GetNameFromId() (bool){
    a := strconv.Itoa(x.Id)
    ret, err:= exec.Command("xl", "domname", a).Output()
    if err != nil {
        println("error get name from id vm")
        return  false
    }

    b := (string(ret))
    x.Domain = b
    return true
}

func (x xen_config_struct) GetNameFromId() (bool){
    a := strconv.Itoa(x.Id)
    ret, err:= exec.Command("xl", "domname", a).Output()
    if err != nil {
        println("error get idfrom name vm")
        return  false
    }

    b := (string(ret))
    x.Name = b
    return true
}

func CreateFileAsHdd(name string, size uint64) (string){

	ssize := "seek=" + strconv.FormatUint(size, 10)
	dir := NormalHddSaveDir + name + ".img"
	cmd := "of=" + dir

	ret1, err := exec.Command("dd", "if=/dev/zero", cmd, ssize,  "count=1" ).CombinedOutput()
	if err != nil {
		log.Error("Error create .img")
		println(string(ret1))
		return "Error"
	}
	return  dir
}

//%TODO
func (x xen_domU) GetIdFromName() (bool){
    ret, err:= exec.Command("xl", "domid", x.Domain).Output()
    if err != nil {
        println("error get name from id vm")
        return  false
    }

    b := (string(ret))
    a, err := strconv.Atoi(b)
    if err != nil {
        println("Error Get Id from name vm convert")
        return false
    }
    x.Id = a
    return true
}
func (x xen_config_struct) GetIdFromName() (bool){
    ret, err:= exec.Command("xl", "domid", x.Name).Output()
    if err != nil {
        println("error get idfrom name vm")
        return  false
    }

    b := (string(ret))
    a, err := strconv.Atoi(b)
    if err != nil {
        println("Error Get Id from name vm convert")
        return false
    }
    x.Id = a
    return true
}

type xen_lvm struct  {
    name string
    size int
    //...
}
func (x xen_lvm) Create_lvm() (bool) {
    return true
}
func (x xen_lvm) Delete_lvm() (bool) {
    return true
}
func (x xen_lvm) statistics_lvm() (bool) {
    return true
}
//ENDTODO

func main() {
    backend1 := logging.NewLogBackend(os.Stderr, "", 0)
    backend2 := logging.NewLogBackend(os.Stderr, "", 0)

    // For messages written to backend2 we want to add some additional
    // information to the output, including the used log level and the name of
    // the function.
    backend2Formatter := logging.NewBackendFormatter(backend2, format)

    // Only errors and more severe messages should be sent to backend1
    backend1Leveled := logging.AddModuleLevel(backend1)
    backend1Leveled.SetLevel(logging.ERROR, "")

    // Set the backends to be used.
    logging.SetBackend(backend1Leveled, backend2Formatter)
    y := xen_domU{}

	y.Create_VM_Wrapper()

	/*
    y.Domain = "arch11"
    y.Id = 6
    y.GetNameFromId()
	*/


}
