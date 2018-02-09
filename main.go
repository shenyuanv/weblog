package main

import (
	"time"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

func symbolicate(filePath string) string {
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

func upload(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method)
	if r.Method == "POST" {
		r.ParseMultipartForm(32 << 20)
		file, _, err := r.FormFile("uploadfile")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()
		f, err := os.OpenFile("./logs/"+time.Now().String(), os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()
		io.Copy(f, file)
		fPath, err := filepath.Abs(f.Name())
		if err != nil {
			fmt.Println(err)
			return
		}
		symFile := symbolicate(fPath)
		fmt.Println(symFile)
		http.ServeFile(w, r, symFile)
	}
}

func indexPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/sym.html")
}

func main() {
	fmt.Printf("starting web server...")
	http.HandleFunc("/symbolicate", upload)
	http.HandleFunc("/", indexPage)
	log.Fatal(http.ListenAndServe(":80", nil))
}
