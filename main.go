package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

func symbolicateCrash(filePath string) string {
	symFile := filePath + ".sym"
	tmpFile := filePath + ".tmp"
	cmdIcrash := exec.Command("python", "/home/ubuntu/Zecops-Tools/iCrash/icrash_linux.py", "-o", tmpFile, filePath)
	cmdSym := exec.Command("/home/ubuntu/Zecops-Tools/iCrash/symbolicatecrash_linux", "-o", symFile, tmpFile)
	errIcrash := cmdIcrash.Run()
	errSym := cmdSym.Run()
	if errSym != nil || errIcrash != nil {
		fmt.Println(errSym, errIcrash)
		return fmt.Sprint(errSym, errIcrash)
	}
	return symFile
}

func symbolicateCrashNew(filePath string) string {
	symFile := filePath + ".sym"
	tmpFile := filePath + ".tmp"
	cmdSym := exec.Command("/home/ubuntu/Zecops-Tools/iCrash/symbolicatecrash_linux", "-o", tmpFile, filePath)
	cmdIcrash := exec.Command("python", "/home/ubuntu/Zecops-Tools/iCrash/icrash.py", "-o", symFile, tmpFile)
	errIcrash := cmdIcrash.Run()
	errSym := cmdSym.Run()
	if errSym != nil || errIcrash != nil {
		fmt.Println(errSym, errIcrash)
		return fmt.Sprint(errSym, errIcrash)
	}
	return symFile
}

func symbolicatePanic(filePath string) string {
	symFile := filePath + ".symp"
	cmdSym := exec.Command("python", "/home/ubuntu/Zecops-Tools/iCrash/ipanic.py", "-o", symFile, filePath)
	errSym := cmdSym.Run()
	if errSym != nil {
		return fmt.Sprint(errSym)
	}
	return symFile
}

func symbolicate(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method)
	if r.Method == "POST" {
		r.ParseMultipartForm(32 << 20)
		file, _, err := r.FormFile("uploadfile")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()
		f, err := os.OpenFile("./logs/"+getMD5(file), os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()
		file.Seek(0, 0)
		io.Copy(f, file)
		fPath, err := filepath.Abs(f.Name())
		if err != nil {
			fmt.Println(err)
			return
		}

		logType := r.FormValue("type")
		symFile := "static/sym.html"
		if logType == "panic" {
			symFile = symbolicatePanic(fPath)
		} else if logType == "crash" {
			if r.FormValue("v") == "2" {
				symFile = symbolicateCrashNew(fPath)
			} else {
				symFile = symbolicateCrash(fPath)
			}
		}
		http.ServeFile(w, r, symFile)
	}
}

func getMD5(f multipart.File) string {
	var returnMD5String string
	hash := md5.New()
	if _, err := io.Copy(hash, f); err != nil {
		return "no_md5"
	}
	hashInBytes := hash.Sum(nil)

	//Convert the bytes to a string
	returnMD5String = hex.EncodeToString(hashInBytes)

	return returnMD5String
}

func indexPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/sym.html")
}

func main() {
	fmt.Printf("starting web server...")
	http.HandleFunc("/symbolicate", symbolicate)
	http.HandleFunc("/", indexPage)
	fs := http.FileServer(http.Dir("static/download"))
	http.Handle("/download/", http.StripPrefix("/download/", fs))
	log.Fatal(http.ListenAndServe(":80", nil))
}
