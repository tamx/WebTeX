package main

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"text/template"

	jpegstructure "github.com/dsoprea/go-jpeg-image-structure/v2"
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

func makeFile(data, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		// log.Fatal(err) //ファイルが開けなかったときエラー出力
		return err
	}
	defer file.Close()
	file.Write([]byte(data))
	return nil
}

func post(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	data := r.PostFormValue("data")
	randomStr, _ := makeRandomStr(16)
	filename := "./tmp/" + randomStr + ".tex"
	// fmt.Println(data)
	defer os.Remove(filename)
	if err := makeFile(data, filename); err != nil {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("File Error: " + err.Error()))
		return
	}
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
}

func exifjpeg(w http.ResponseWriter, r *http.Request) {
	jpegBase64, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	file, err := base64.StdEncoding.DecodeString(string(jpegBase64))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// パーサーを作る
	jmp := jpegstructure.NewJpegMediaParser()
	// JPEGファイルを読み取ってセグメントリストを得る
	ec, err := jmp.ParseBytes(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// タグ（Exifに含まれる情報）の一覧を得る
	sl := ec.(*jpegstructure.SegmentList)
	_, _, tags, err := sl.DumpExif()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	// タグの一覧を表示する
	for _, tag := range tags {
		fmt.Printf("%s: %s: %#v\n", tag.IfdPath, tag.TagName, tag.Value)
		if tag.TagName == "ImageDescription" {
			w.Write(tag.ValueBytes)
		}
	}
}

func main() {
	http.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("static"))))
	http.HandleFunc("/",
		func(w http.ResponseWriter, r *http.Request) {
			tmpl := template.Must(template.
				ParseFiles("static/vscode.html"))
			tmpl.Execute(w, nil)
		})
	http.HandleFunc("/post", post)
	http.HandleFunc("/jpeg", exifjpeg)

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
