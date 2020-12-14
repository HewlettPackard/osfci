package main

import (
    "base"
    "os"
    "net/http"
    "strings"
    "log"
    "fmt"
    "io/ioutil"
    "sync"
    "strconv"
    "encoding/base64"
)

var storageRoot = os.Getenv("STORAGE_ROOT")
// write operation must be protected by a Mutex
var file sync.RWMutex

// This is getting a user file entry

func getEntry(username string) (string,int) {
	// The first letter of the username is used as a directory entry
	// if the directory exist we check for the usenarme.conf entry into it
	// if it is there we return the content of the file
	_, err := os.Stat(storageRoot + "/" + string(username[0]))
	if ( ! os.IsNotExist(err) ) {
		// The directory exist we must now check if the file exist
		_,err := os.Stat( storageRoot + "/" + string(username[0]) + "/" + username )
		if ( ! os.IsNotExist(err) ) {
			// We must return the file content into a string
			b,_ := ioutil.ReadFile(storageRoot + "/" + string(username[0]) + "/" + username)
			return string(b),1
		} else
		{
			return "",0
		}
	} else
	{
		return "",0
	}
}

// This is creating a user file entry

func createEntry(username string, content string) (int) {
	_, err := os.Stat(storageRoot + "/" + string(username[0]))
	file.Lock()
	defer file.Unlock()
	if (  os.IsNotExist(err) ) {
		// we must create the directory which will contain the file
		_ = os.Mkdir(storageRoot + "/" + string(username[0]), os.ModePerm)
	}
	_ = ioutil.WriteFile(storageRoot + "/" + string(username[0]) + "/" + username, []byte(content), os.ModePerm)
	return 1
}

func createImage(username string, content string) (int) {
        _, err := os.Stat(storageRoot + "/" + string(username[0]))
	
        file.Lock()
	defer file.Unlock()
        if (  os.IsNotExist(err) ) {
                // we must create the directory which will contain the file
                _ = os.Mkdir(storageRoot + "/" + string(username[0]), os.ModePerm)
        }
	// We have to remove the "base64, stuff"
	coI := strings.Index(content, ",")
        rawImage := string(content)[coI+1:]
        decodedBody, _ := base64.StdEncoding.DecodeString(rawImage)
	_ = ioutil.WriteFile(storageRoot + "/" + string(username[0]) + "/" + username + ".jpg", []byte(decodedBody), os.ModePerm)
	return 1
}

func storeFirmware(username string, r *http.Request, firmware string) (int) {
        _, err := os.Stat(storageRoot + "/" + string(username[0]))

        file.Lock()
        defer file.Unlock()
        if (  os.IsNotExist(err) ) {
                // we must create the directory which will contain the file
                _ = os.Mkdir(storageRoot + "/" + string(username[0]), os.ModePerm)
        }
        _ = ioutil.WriteFile(storageRoot + "/" + string(username[0]) + "/" + firmware + "_" + username + ".rom", base.HTTPGetBody(r), os.ModePerm)
        return 1
}

func getSystemBIOS(username string, w http.ResponseWriter) {
	content,_ := ioutil.ReadFile(storageRoot + "/" + string(username[0]) + "/" + "linuxboot_" + username + ".rom")
	w.Header().Add("Content-Length", strconv.Itoa(len(content)))
	w.Write(content)
}

func getOpenBMC(username string, w http.ResponseWriter) {
	content,_ := ioutil.ReadFile(storageRoot + "/" + string(username[0]) + "/" + "openbmc_" + username + ".rom")
	w.Header().Add("Content-Length", strconv.Itoa(len(content)))
	w.Write(content)
}

func getImage(username string) (string) {
	_, err := os.Stat(storageRoot + "/" + string(username[0]))
        file.Lock()
        defer file.Unlock()
        if (  os.IsNotExist(err) ) {
                // we must create the directory which will contain the file
                _ = os.Mkdir(storageRoot + "/" + string(username[0]), os.ModePerm)
		return ""
        }

        _, err = os.Stat(storageRoot + "/" + string(username[0]) + "/" + username + ".jpg")
	if ( os.IsNotExist(err) ) {
		var staticAssetsDir = os.Getenv("STATIC_ASSETS_DIR")
		content,_ := ioutil.ReadFile(staticAssetsDir + "images/forklift.png")
		encodedContent := base64.StdEncoding.EncodeToString(content)
		return encodedContent
	} else {
		content,_ := ioutil.ReadFile(storageRoot + "/" + string(username[0]) + "/" + username + ".jpg")
		encodedContent := base64.StdEncoding.EncodeToString(content)
		return encodedContent
	}
}

func deleteEntry(username string, content string) (int) {
	_, err := os.Stat(storageRoot + "/" + string(username[0]) + "/" + username)
	file.Lock()
	defer file.Unlock()
	if (  !os.IsNotExist(err) ) {
		_=os.Remove(storageRoot + "/" + string(username[0]) + "/" + username)
	}
	_, err = os.Stat(storageRoot + "/" + string(username[0]) + "/" + username + ".jpg")
	if (  !os.IsNotExist(err) ) {
                _=os.Remove(storageRoot + "/" + string(username[0]) + "/" + username + ".jpg")
        }
	return 1
}

func distrosCallback(w http.ResponseWriter, r *http.Request) {
	// We must breakdown the words, because directory filename is the last word
	path := strings.Split( r.URL.Path, "/" )
	if ( len(path) < 3 ) {
                http.Error(w, "401 Malformed URI", 401)
	}
	if ( path[2] == "" ) {
		// We must provide the directory content from distros
		files,_ := ioutil.ReadDir(storageRoot+"/distros")
		var answer string
		var count int
		if len(files) > 0 {
			answer = "{ \"files\": ["
			count = 0
			for _, file := range files {
			    if count == 1 {
				answer = answer + ","
			    }
		            answer = answer + "\"" + file.Name() + "\""
			    count = 1
                	}
			answer = answer + "] }"
		}
		w.Write([]byte(answer))
	} else {
		// We must serve the file
		http.ServeFile(w,r, storageRoot+"/distros/"+path[2])
	}
}

func userCallback(w http.ResponseWriter, r *http.Request) {
        var username string
	var filecontent string
	var return_value int
	// We must breakdown the words, because username is not always the last word
	path := strings.Split( r.URL.Path, "/" )
        if ( len(path) < 3 ) {
                http.Error(w, "401 Malformed URI", 401)
                return
        }
        username = path[2]
	var command string
	if ( len(path) > 3 ) {
		command = path[3]
	}
        switch r.Method {
                case http.MethodGet:
			// Serve the resource.
			// I must return the content of the user file if it does exist otherwise
			// an error
			switch command {
			case "avatar":
				w.Write([]byte(getImage(username)))
			case "getFirmware":
				getSystemBIOS(username, w)
			case "getBMCFirmware":
				getOpenBMC(username, w)
			default:
				filecontent, return_value=getEntry(username)
				if ( return_value != 0) {
					fmt.Fprint(w,filecontent)			
				} else {
					fmt.Fprintf(w,"Error")
				}
			}
                case http.MethodPut:
			// Update an existing record.
			if (r.Header.Get("Content-Type") != "image/jpg" ) {
				if ( r.Header.Get("Content-Type") == "application/octet-stream" ) {
					// We got a firmware
					if ( command == "linuxboot" ) {
						storeFirmware(username,r, "linuxboot")
					} else {	
						if ( command == "openbmc" ) {
							storeFirmware(username,r, "openbmc")
						}
					}
				} else {
					createEntry(username,string(base.HTTPGetBody(r)))	
				}
			} else {
				createImage(username,string(base.HTTPGetBody(r)))
			}
		case http.MethodDelete:
			deleteEntry(username,string(base.HTTPGetBody(r)))
                default:
        }
}

func main() {
    print("=============================== \n")
    print("| Starting storage backend    |\n")
    print("| Development version -       |\n")
    print("| Private use only            |\n")
    print("=============================== \n")

    mux := http.NewServeMux()
    var StorageURI = os.Getenv("STORAGE_URI")
    var StorageTCPPORT = os.Getenv("STORAGE_TCPPORT")

    mux.HandleFunc("/user/", userCallback)
    mux.HandleFunc("/distros/", distrosCallback)

    log.Fatal(http.ListenAndServe(StorageURI+StorageTCPPORT, mux))
}
