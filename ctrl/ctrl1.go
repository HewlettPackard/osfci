package main

import (
	"net/http"
	"strings"
	"path"
	"fmt"
	"log"
	"os/exec"
	"net"
	"time"
	"os"
	"io"
	"base"
	"golang.org/x/sys/unix"
)

var binariesPath = os.Getenv("BINARIES_PATH")
var firmwaresPath = os.Getenv("FIRMWARES_PATH")
var distrosPath = os.Getenv("DISTROS_PATH")
var compileUri = os.Getenv("COMPILE_URI")
var compileTcpPort = os.Getenv("COMPILE_TCPPORT")
var storageUri = os.Getenv("STORAGE_URI")
var storageTcpPort = os.Getenv("STORAGE_TCPPORT")
var isEmulatorsPool = os.Getenv("IS_EMULATORS_POOL")

func ShiftPath(p string) (head, tail string) {
    p = path.Clean("/" + p)
    i := strings.Index(p[1:], "/") + 1
    if i <= 0 {
        return p[1:], "/"
    }
    return p[1:i], p[i:]
}


func home(w http.ResponseWriter, r *http.Request) {
	head,_ := ShiftPath( r.URL.Path)
	switch ( head ) {
		case "getosinstallers":
			_,tail := ShiftPath( r.URL.Path)
			file :=  strings.Split(tail,"/")
			// file[1] does contain the name of the server which is needed
			// The file is seating into the storage server ... We have to transfer it
			// into a local storage (ideally a RAMFS to avoid potential storage impact
			// and accelerating transfer)
			// but all of this is performed within an external script as it needs to
			// be piped to a ttyd as to provide end user feedback
			fmt.Printf("Usb load received\n")
                        args := []string { distrosPath+"/"+file[1] }
                        cmd := exec.Command(binariesPath+"/load_usb", args...)
			cmd.SysProcAttr = &unix.SysProcAttr{
                                                Setsid: true,
                        }
                        cmd.Start()
                        done := make(chan error, 1)
                        go func() {
				done <- cmd.Wait()
			}()
		case "isEmulatorsPool":
			w.Write([]byte('{ "isPool":"'+isEmulatorsPool+'" }'))
		case "resetEmulator":
			_,tail := ShiftPath( r.URL.Path)
			path :=  strings.Split(tail,"/")
                        emulator := path[2]
			if ( emlator == "bmc" ) {
                        } else {
                                if ( emulator == "rom" ) {
                                } else {
                                        w.Write([]byte(emulator))
                                }
                        }
		case "bmcfirmware":
                        switch r.Method {
                                case http.MethodPost:
					_,tail := ShiftPath( r.URL.Path)
                                        path :=  strings.Split(tail,"/")
                                        username := path[1]
                                        r.Body = http.MaxBytesReader(w, r.Body, 64<<20+4096)
                                        err := r.ParseMultipartForm(64<<20+4096)
                                        if ( err != nil ) {
                                                fmt.Printf("Error %s\n",err.Error())
                                        }
                                        file,handler,_ := r.FormFile("fichier")

                                        defer file.Close()
                                        f, err := os.OpenFile(firmwaresPath+"/_"+username+"_"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
                                        if err != nil {
                                               fmt.Println(err)
                                               return
                                        }
                                        defer f.Close()
                                        io.Copy(f, file)
					// we must forward the request to the relevant test server
		                        fmt.Printf("Ilo start received\n")
		                        args := []string { firmwaresPath+"/_"+username+"_"+handler.Filename }
		                        cmd := exec.Command(binariesPath+"/start_bmc", args...)
		                        cmd.Start()
		                        done := make(chan error, 1)
		                        go func() {
		                            done <- cmd.Wait()
		                        }()
		                        // We must hang off after being sure that the console daemon is properly starter
		                        conn, err := net.DialTimeout("tcp", "localhost:7681", 220*time.Millisecond)
		                        max_loop := 5
		                        for ( err != nil && max_loop > 0 ) {
		                                conn, err = net.DialTimeout("tcp", "localhost:7681", 220*time.Millisecond)
		                        }
		                        if ( err != nil ) {
	                                // Daemon has not started
	                                // Let's report an error
		                                w.Write([]byte("Error"))
		                                return
		                        } else {
		                                conn.Close()
				        }
				}
		case "biosfirmware":
			switch r.Method {
                                case http.MethodPost:
					_,tail := ShiftPath( r.URL.Path)
					path :=  strings.Split(tail,"/")
					username := path[1]
                                        r.Body = http.MaxBytesReader(w, r.Body, 64<<20+4096)
                                        err := r.ParseMultipartForm(64<<20+4096)
                                        if ( err != nil ) {
                                                fmt.Printf("Error %s\n",err.Error())
                                        }
                                        file,handler,_ := r.FormFile("fichier")

                                        defer file.Close()
                                        f, err := os.OpenFile(firmwaresPath+"/_"+username+"_"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
                                        if err != nil {
                                               fmt.Println(err)
                                               return
                                        }
                                        defer f.Close()
                                        io.Copy(f, file)
					// we must forward the request to the relevant test server
		                        fmt.Printf("System BIOS start received\n")
		                        args := []string { firmwaresPath+"/_"+username+"_"+handler.Filename }
		                        cmd := exec.Command(binariesPath+"/start_smbios", args...)
		                        cmd.Start()
		                        done := make(chan error, 1)
		                        go func() {
		                            done <- cmd.Wait()
		                        }()
		                        conn, err := net.DialTimeout("tcp", "localhost:7683", 220*time.Millisecond)
		                        max_loop := 5
		                        for ( err != nil && max_loop > 0 ) {
		                                conn, err = net.DialTimeout("tcp", "localhost:7683", 220*time.Millisecond)
		                        }
		                        if ( err != nil ) {
		                                // Daemon has not started
		                                // Let's report an error
		                                w.Write([]byte("Error"))
		                                return
		                        } else {
		                                conn.Close()
		                        }

                        }
		case "loadfromstoragesmbios":
			// We must get the username from the request
			_, tail := ShiftPath( r.URL.Path)
                        login := tail[1:]
			// We have to retreive the BIOS from the compile server
			myfirmware := base.HTTPGetRequest("http://"+storageUri + storageTcpPort + "/user/"+login+"/getFirmware")
                        // f, err := os.Create("firmwares/linuxboot_"+login+".rom", os.O_WRONLY|os.O_CREATE, 0666)
                        f, err := os.Create(firmwaresPath+"/linuxboot_"+login+".rom")
			defer f.Close()
			f.Write([]byte(myfirmware))

					fmt.Printf("System BIOS start received\n")
                                        args := []string { firmwaresPath+"/linuxboot_"+login+".rom" }
                                        cmd := exec.Command(binariesPath+"/start_smbios", args...)
                                        cmd.Start()
                                        done := make(chan error, 1)
                                        go func() {
                                            done <- cmd.Wait()
                                        }()
                                        conn, err := net.DialTimeout("tcp", "localhost:7683", 220*time.Millisecond)
                                        max_loop := 5
                                        for ( err != nil && max_loop > 0 ) {
                                                conn, err = net.DialTimeout("tcp", "localhost:7683", 220*time.Millisecond)
                                        }
                                        if ( err != nil ) {
                                                // Daemon has not started
                                                // Let's report an error
                                                w.Write([]byte("Error"))
                                                return
                                        } else {
                                                conn.Close()
                                        }
		case "loadfromstoragebmc":
                        // We must get the username from the request
                        _, tail := ShiftPath( r.URL.Path)
                        login := tail[1:]
                        // We have to retreive the BIOS from the compile server
                        myfirmware := base.HTTPGetRequest("http://"+storageUri + storageTcpPort + "/user/"+login+"/getBMCFirmware")
                        // f, err := os.Create("firmwares/openbmc_"+login+".rom", os.O_WRONLY|os.O_CREATE, 0666)
                        f, err := os.Create(firmwaresPath+"/openbmc_"+login+".rom")
                        defer f.Close()
                        f.Write([]byte(myfirmware))
                        fmt.Printf("BMC start received\n")
                        args := []string { firmwaresPath+"/openbmc_"+login+".rom" }
                        cmd := exec.Command(binariesPath+"/start_bmc", args...)
                        cmd.Start()
                        done := make(chan error, 1)
                        go func() {
                               done <- cmd.Wait()
                        }()
                        conn, err := net.DialTimeout("tcp", "localhost:7681", 220*time.Millisecond)
                        max_loop := 5
                        for ( err != nil && max_loop > 0 ) {
                                             conn, err = net.DialTimeout("tcp", "localhost:7681", 220*time.Millisecond)
                        }
                        if ( err != nil ) {
                                // Daemon has not started
                                // Let's report an error
                               w.Write([]byte("Error"))
                                return
                        } else {
                              conn.Close()
                        }	
		case "startbmc":
			fmt.Printf("BMC start received\n")
			args := []string { firmwaresPath+"/ilo_dl360_OpenBMC.rom" }
                        cmd := exec.Command(binariesPath+"/start_bmc", args...)
                        cmd.Start()
			done := make(chan error, 1)
                        go func() {
                            done <- cmd.Wait()
                        }()
			// We must hang off after being sure that the console daemon is properly starter
			conn, err := net.DialTimeout("tcp", "localhost:7681", 220*time.Millisecond)
			max_loop := 5
			for ( err != nil && max_loop > 0 ) {
				conn, err = net.DialTimeout("tcp", "localhost:7681", 220*time.Millisecond)
			}
			if ( err != nil ) {
				// Daemon has not started
				// Let's report an error
				w.Write([]byte("Error"))
				return
			} else {
				conn.Close()
			}
		case "startsmbios":
			// we must forward the request to the relevant test server
                        fmt.Printf("System BIOS start received\n")
                        args := []string { firmwaresPath+"/SBIOS_OpenBMC.rom" }
                        cmd := exec.Command(binariesPath+"/start_smbios", args...)
                        cmd.Start()
                        done := make(chan error, 1)
                        go func() {
                            done <- cmd.Wait()
                        }()
			conn, err := net.DialTimeout("tcp", "localhost:7683", 220*time.Millisecond)
                        max_loop := 5
                        for ( err != nil && max_loop > 0 ) {
                                conn, err = net.DialTimeout("tcp", "localhost:7683", 220*time.Millisecond)
                        }
                        if ( err != nil ) {
                                // Daemon has not started
                                // Let's report an error
                                w.Write([]byte("Error"))
                                return
                        } else {
                                conn.Close()
                        }
		case "poweron":
			fmt.Printf("start power\n")
                        args := []string { "on" }
                        cmd := exec.Command(binariesPath+"/iPDUpower", args...)
                        cmd.Start()
                        done := make(chan error, 1)
                        go func() {
                            done <- cmd.Wait()
                        }()
		case "bmcup":
			
		case "poweroff":
			// We need to cleanup the em100
			// We also need to clean up the screen command
			// and free the USB->Serial
			fmt.Printf("stop power\n")
                        args := []string { "off" }
                        cmd := exec.Command(binariesPath+"/iPDUpower", args...)
                        cmd.Start()
                        cmd.Wait()
			args = []string { "" }
                        cmd = exec.Command(binariesPath+"/cleanUP", args...)
                        cmd.Start()
                        cmd.Wait()
		default:
	}
}

func main() {
    print("=============================== \n")
    print("| Starting frontend           |\n")
    print("| Development version -       |\n")
    print("=============================== \n")

    var ctrlTcpPort = os.Getenv("CTRL_TCPPORT")
    mux := http.NewServeMux()

    // Highest priority must be set to the signed request
    mux.HandleFunc("/",home)


    log.Fatal(http.ListenAndServe(ctrlTcpPort, mux))
}
