package main

import (
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
	cmdIcrash := exec.Command("python", "/home/shenyuanv/Work/Zecops-Tools/iCrash/icrash_linux.py", "-o", tmpFile, filePath)
	cmdSym := exec.Command("/home/shenyuanv/Work/Zecops-Tools/iCrash/symbolicatecrash_linux", "-o", symFile, tmpFile)
	errIcrash := cmdIcrash.Run()
	errSym := cmdSym.Run()
	if errSym != nil || errIcrash != nil {
		fmt.Println(errSym, errIcrash)
		return fmt.Sprint(errSym, errIcrash)
	}
	return symFile
}

func upload(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method) //获取请求的方法
	if r.Method == "POST" {
		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("uploadfile")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()
		fmt.Fprintf(w, "%v", handler.Header)
		f, err := os.OpenFile("./logs/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666) // 此处假设当前目录下已存在test目录
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
