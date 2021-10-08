package main

import (
	"bufio"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"text/template"
)

func makeRandomStr(digit uint32) (string, error) {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// 乱数を生成
	b := make([]byte, digit)
	if _, err := rand.Read(b); err != nil {
		return "", errors.New("unexpected error")
	}

	// letters からランダムに取り出して文字列を生成
	var result string
	for _, v := range b {
		// index が letters の長さに収まるように調整
		result += string(letters[int(v)%len(letters)])
	}
	return result, nil
}

func makeFile(data, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err) //ファイルが開けなかったときエラー出力
	}
	defer file.Close()
	file.Write([]byte(data))
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("index.html"))
		tmpl.Execute(w, nil)
	})
	http.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		data := r.PostFormValue("data")
		randomStr, _ := makeRandomStr(16)
		filename := "./tmp/" + randomStr + ".tex"
		// fmt.Println(data)
		defer os.Remove(filename)
		makeFile(data, filename)
		pdffile := "./tmp/" + randomStr + ".pdf"

		cmd := exec.Command("/usr/bin/cluttex",
			"-e", "platex",
			"-o", pdffile,
			filename)
		defer os.Remove(pdffile)
		stdout, _ := cmd.StdoutPipe()
		cmd.Start()

		result := ""
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			result += scanner.Text() + "\n"
		}

		// err := cmd.Run()
		err := cmd.Wait()
		if err != nil {
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(result))
			w.Write([]byte("Command Exec Error: " + err.Error()))
			return
		}

		w.Header().Set("Content-Type", "application/pdf")
		reader, err := os.Open(pdffile)
		if err != nil {
			http.Error(w, err.Error(),
				http.StatusInternalServerError)
			return
		}

		_, err = io.Copy(w, reader)
		if err != nil {
			http.Error(w, err.Error(),
				http.StatusInternalServerError)
			return
		}
	})

	// このロジックはApp Engine APIから完全脱却した場合のみ
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	addressPort := fmt.Sprintf("0.0.0.0:%s", port)
	log.Printf("Listening on port %s", addressPort)
	err := http.ListenAndServe(addressPort, nil)
	log.Fatal(err, nil)
}
