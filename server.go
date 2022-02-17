package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
)

type Ret struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data interface{} `json:"data"`
}

func getTheMuteTime(url string) (muteTime float64, integratedLoudness float64, err error) {
	cmdArguments := []string{"-i",url, "-filter_complex", "ebur128", "-c:v", "copy", "-t","10", "-f", "null", "/dev/null"}
	command := exec.Command("ffmpeg", cmdArguments...)
	var out bytes.Buffer
	var errOut bytes.Buffer
	command.Stdout = &out
	command.Stderr = &errOut

	err = command.Start()

	if err != nil {
		fmt.Println(err)
	}

	if err = command.Wait(); err != nil {
		fmt.Println(err.Error())
	}

	result := errOut.String()

	reg1 := regexp.MustCompile(`t:[\s]*(\d+[\.\d+]*)\s+TARGET:(-\d+)\sLUFS\s+M:([-\s]*\d+\.\d)\s+S:([-\s]*\d+\.\d)\s+I:\s+([-\s]*\d+\.\d)\sLUFS\s+LRA:\s+(\d+\.\d)\sLU`)
	if reg1 == nil { //解释失败，返回nil
		fmt.Println("regexp err")
		return
	}

	result1 := reg1.FindAllStringSubmatch(result, -1)

	// fmt.Println("result1 = ", result1)

	if len(result1) == 0 {
		fmt.Println("result1 is empty")
		return
	}

	for _, v := range result1 {
		time, _ := strconv.ParseFloat(v[1], 10)
		volume, _ := strconv.ParseFloat(v[5], 10)

		//fmt.Println(time, volume)

		if volume > -70.0 {
			muteTime = time
			break
		}
	}

	reg2 := regexp.MustCompile(`\s+I:\s+([-+\s]\d+\.\d) LUFS`)
	if reg2 == nil { //解释失败，返回nil
		fmt.Println("regexp err")
		return
	}

	result2 := reg2.FindAllStringSubmatch(result, -1)

	loudness := result2[len(result2)-1][1]

	integratedLoudness, _ = strconv.ParseFloat(loudness, 10)
	return muteTime, integratedLoudness, nil
}

func handler(rw http.ResponseWriter, req *http.Request) {
	var err error

	rw.Header().Set("Content-Type", "application/json")


	defer func() {
		if err != nil {
			ret := new(Ret)
			ret.Code = 500
			ret.Msg = err.Error()
			ret_json, _ := json.Marshal(ret)
			io.WriteString(rw, string(ret_json))
		}
	}()

	defer req.Body.Close()

	url := req.URL.Query().Get("url")

	var muteTime float64
	var integratedLoudness float64

	ret := new(Ret)
	if url == "" {
		ret.Code = 500
		ret.Msg = "error"
		ret.Data = "url is empty"
	}

	muteTime, integratedLoudness, err = getTheMuteTime(url)

	fmt.Println("loudness = ", integratedLoudness)

	fmt.Println("muteTime = ", muteTime)

	if err == nil {
		ret.Code = 0
		ret.Msg = "success"
		ret.Data = map[string]float64{"mute_time": muteTime, "integrated_loudness": integratedLoudness}
		ret_json, _ := json.Marshal(ret)
		io.WriteString(rw, string(ret_json))
	}
}

func health(rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte("ok"))
}

func main() {
	port := os.Getenv("PORT_HTTP")
	if port == "" {
		port = "9100"
	}

	http.HandleFunc("/handler", handler)
	http.HandleFunc("/health", health)
	log.Fatalln(http.ListenAndServe("0.0.0.0:" + port, nil))

}
