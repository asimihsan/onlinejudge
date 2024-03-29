package main

import (
	"github.com/stretchr/graceful"

	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

var (
	logger                 = getLogger("logger")
	letters                = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	digits                 = []rune("0123456789")
	maxOutstandingRequests = 1
	handlerSemaphore       = make(chan int, maxOutstandingRequests)
	grecaptchaSecret       = "6LcB8gATAAAAAByLaeJzveuN4_lP_yDdiszVoL60"
	outputLimit            = 10 * 1024
	lxcMutex               = &sync.Mutex{}
	ephemeralImageName     = ""
	lxcPath                = "/home/ubuntu/.local/share/lxc/"
)

func getLogger(prefix string) *log.Logger {
	paddedPrefix := fmt.Sprintf("%-8s: ", prefix)
	return log.New(os.Stdout, paddedPrefix,
		log.Ldate|log.Ltime|log.Lmicroseconds)
}

func getLogPill() string {
	b := make([]rune, 8)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func getEphemeralImageName() string {
	b := make([]rune, 8)
	b[0] = 'v'
	for i := 1; i < len(b); i++ {
		b[i] = digits[rand.Intn(len(digits))]
	}
	return string(b)
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func makeGzipHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			fn(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		gzw := gzipResponseWriter{Writer: gz, ResponseWriter: w}
		fn(gzw, r)
	}
}

type run_handler_struct struct {
	Code      string `json:"code,omitempty"`
	UnitTest  string `json:"unit_test,omitempty"`
	Recaptcha string `json:"recaptcha,omitempty"`
}

type verify_recaptcha_struct struct {
	Success bool
}

func verifyRecaptcha(logger *log.Logger, recaptcha string) (result bool, err error) {
	logger.Println("verifyRecaptcha entry()")
	defer logger.Printf("verifyRecaptcha exit(). result: %s, err: %s\n", result, err)
	result = false
	uri := fmt.Sprintf("https://www.google.com/recaptcha/api/siteverify?secret=%s&response=%s",
		grecaptchaSecret, recaptcha)
	client := &http.Client{}
	request, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		logger.Println("Failed to create HTTP GET to Google reCAPTCHA")
		return
	}
	resp, err := client.Do(request)
	if err != nil {
		logger.Println("Failed during HTTP GET to Google reCAPTCHA")
		return
	}
	if resp.StatusCode != 200 {
		logger.Printf("HTTP GET to Google reCAPTCHA not 200: %s\n", resp)
	}
	decoder := json.NewDecoder(resp.Body)
	var t verify_recaptcha_struct
	err = decoder.Decode(&t)
	if err != nil {
		logger.Panicf("Could not decode Google reCAPTCHA resposne")
	}

	result = t.Success

	return
}

func writeJSONResponse(logger *log.Logger, response map[string]interface{}, w http.ResponseWriter) {
	logger.Println("writeJSONResponse() entry")
	defer logger.Println("writeJSONResponse() exit")
	responseEncoded, _ := json.Marshal(response)
	io.WriteString(w, string(responseEncoded))
}

func runHandler(language string, w http.ResponseWriter, r *http.Request) {
	logger = getLogger(getLogPill())
	logger.Println("runHandler() entry.")
	defer logger.Println("runHandler() exit.")

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	if r.Method == "OPTIONS" {
		return
	}

	response := map[string]interface{}{}
	defer writeJSONResponse(logger, response, w)

	decoder := json.NewDecoder(r.Body)
	var t run_handler_struct
	err := decoder.Decode(&t)
	if err != nil {
		response["success"] = false
		response["output"] = "<could not decode JSON POST request>"
		logger.Panicf("Could not decode JSON POST request")
	}

	// Verify Google reCAPTCHA
	/*
	   result, err := verifyRecaptcha(logger, t.Recaptcha)
	   if err != nil {
	       response["recaptchaSuccess"] = false
	       response["recaptcha"] = "<failed to verify Google reCAPTCHA>"
	       io.WriteString(w, "<failed to verify Google reCAPTCHA>\n")
	       //return
	   } else if result == false {
	       response["recaptchaSuccess"] = false
	       response["recaptcha"] = "<verified Google reCAPTCHA as false, not human. reload page and try again>"
	       //io.WriteString(w, "<verified Google reCAPTCHA as false, not human. reload page and try again.>\n")
	       //return
	   } else {
	       response["recaptchaSuccess"] = true
	   }
	*/

	// Put data into the channel which acts like a semaphore. Once it reaches
	// it's capacity it will block here; hence only "maxOutstandingRequests"
	// permitted.
	handlerSemaphore <- 1
	defer func() {
		<-handlerSemaphore
	}()

	// grab the mutex for the LXC container
	lxcMutex.Lock()
	defer lxcMutex.Unlock()

	codeFile := prepareCodeFile(logger, t.Code)
	defer os.Remove(codeFile.Name())

	unitTestFile := prepareCodeFile(logger, t.UnitTest)
	defer os.Remove(unitTestFile.Name())

	outputFile := prepareOutputFile(logger)
	defer os.Remove(outputFile.Name())

	cmd := runCommand(language, codeFile.Name(), unitTestFile.Name())
	runCode(cmd, outputFile, logger, response)

	if _, ok := response["success"]; !ok {
		response["success"] = true
	}
	go func() {
		restartLxcContainer()
	}()
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	logger = getLogger(getLogPill())
	//logger.Println("pingHandler() entry.")
	//defer logger.Println("pingHandler() exit.")

	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == "OPTIONS" {
		return
	}

	io.WriteString(w, "pong")
}

func runCHandler(w http.ResponseWriter, r *http.Request) {
	runHandler("c", w, r)
}

func runCPPHandler(w http.ResponseWriter, r *http.Request) {
	runHandler("cpp", w, r)
}

func runPythonHandler(w http.ResponseWriter, r *http.Request) {
	runHandler("python", w, r)
}

func runRubyHandler(w http.ResponseWriter, r *http.Request) {
	runHandler("ruby", w, r)
}

func runJavaHandler(w http.ResponseWriter, r *http.Request) {
	runHandler("java", w, r)
}

func runJavaScriptHandler(w http.ResponseWriter, r *http.Request) {
	runHandler("javascript", w, r)
}

func prepareCodeFile(logger *log.Logger, code string) *os.File {
	logger.Println("prepareCodeFile() entry.")
	codeFile, err := ioutil.TempFile("", "run")
	if err != nil {
		logger.Panicf("couldn't create temporary code file: %s", err)
	}
	logger.Println("temporary codeFile: ", codeFile.Name())
	logger.Println("prepareCodeFile() writing code to file...")
	io.WriteString(codeFile, code)
	codeFile.Close()
	logger.Println("prepareCodeFile() finished writing code to file.")
	logger.Println("prepareCodeFile() exit.")
	return codeFile
}

func prepareOutputFile(logger *log.Logger) *os.File {
	logger.Println("prepareOutputFile() entry.")
	outputFile, err := ioutil.TempFile("", "run-output")
	if err != nil {
		logger.Panicf("couldn't create temporary output file: %s", err)
	}
	logger.Println("temporary outputFile: ", outputFile.Name())
	logger.Println("prepareOutputFile() exit.")
	return outputFile
}

func runCode(cmd *exec.Cmd, outputFile *os.File, logger *log.Logger, response map[string]interface{}) {
	logger.Println("runCode() entry.")
	defer logger.Println("runCode() exit.")
	cmd.Stdout = outputFile
	cmd.Stderr = outputFile
	logger.Println("runCode() running file...")
	err := cmd.Start()
	if err != nil {
		logger.Panicf("failed to run command: %s", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	var outputBuffer bytes.Buffer
	select {
	case <-time.After(5 * time.Second):
		if err := cmd.Process.Kill(); err != nil {
			logger.Fatalf("failed to kill: %s", err)
			<-done // allow goroutine to exit
		}
		msg := "<process ran for too long. output is below>\n"
		logger.Println(msg)
		response["success"] = false
		outputBuffer.WriteString(msg)
	case err := <-done:
		if err != nil {
			msg := fmt.Sprintf("<process finished with error: %s. output is below>\n", err)
			logger.Println(msg)
			response["success"] = false
			outputBuffer.WriteString(msg)
		}
	}

	logger.Println("runCode() finished file.")

	logger.Println("runCode() returning output...")
	_, _ = outputFile.Seek(0, 0)
	outputLimitReader := io.LimitReader(outputFile, int64(outputLimit))
	output, err := ioutil.ReadAll(outputLimitReader)
	if err != nil {
		logger.Println(err)
		response["success"] = false
		outputBuffer.WriteString(err.Error())
	}
	outputBuffer.Write(output)
	if len(output) == outputLimit {
		outputBuffer.WriteString("\n<too much output, truncated>\n")
	}
	response["output"] = outputBuffer.String()
	logger.Println("runCode() finished returning output.")
}

func copyPrepareFile(source_filepath string, dest_filepath string) error {
	if err := exec.Command("cp", "-f", source_filepath, dest_filepath).Run(); err != nil {
		logger.Panicf("failed to copy code to %s", dest_filepath)
		return err
	}
	if err := os.Chmod(dest_filepath, 0777); err != nil {
		logger.Panicf("failed to chmod %s", dest_filepath)
		return err
	}
	return nil
}

func runCommand(language string, code_filepath string, unittest_filepath string) *exec.Cmd {
	if err := exec.Command("rm", "-f", "/tmp/foo/*").Run(); err != nil {
		logger.Panicf("failed to clean up old out files in /tmp/foo")
	}
	switch language {
	case "c":
		copyPrepareFile(code_filepath, "/tmp/foo/program.c")
		copyPrepareFile(unittest_filepath, "/tmp/foo/program_test.c")
		return exec.Command("lxc-attach", "-n", ephemeralImageName, "--clear-env", "--keep-var", "TERM", "--",
			"su", "-", "ubuntu", "-c", "/usr/local/bin/sandbox /usr/bin/gcc -Wall -std=c99 /tmp/foo/*.c -o /tmp/foo/a.out && /usr/local/bin/sandbox /tmp/foo/a.out")
	case "cpp":
		copyPrepareFile(code_filepath, "/tmp/foo/program.cpp")
		copyPrepareFile(unittest_filepath, "/tmp/foo/program_test.cpp")
		copyPrepareFile("/home/ubuntu/catch.hpp", "/tmp/foo/catch.hpp")
		return exec.Command("lxc-attach", "-n", ephemeralImageName, "--clear-env", "--keep-var", "TERM", "--",
			"su", "-", "ubuntu", "-c", "/usr/local/bin/sandbox /usr/bin/g++ -Wall -std=c++11 /tmp/foo/*.cpp -o /tmp/foo/a.out && /usr/local/bin/sandbox /tmp/foo/a.out")
	case "python":
		copyPrepareFile(code_filepath, "/tmp/foo/foo.py")
		copyPrepareFile(unittest_filepath, "/tmp/foo/foo_test.py")
		return exec.Command("lxc-attach", "-n", ephemeralImageName, "--clear-env", "--keep-var", "TERM", "--",
			"su", "-", "ubuntu", "-c", "/usr/local/bin/sandbox /usr/bin/python /tmp/foo/foo_test.py")
	case "ruby":
		copyPrepareFile(code_filepath, "/tmp/foo/foo.rb")
		copyPrepareFile(unittest_filepath, "/tmp/foo/foo_test.rb")
		return exec.Command("lxc-attach", "-n", ephemeralImageName, "--clear-env", "--keep-var", "TERM", "--",
			"su", "-", "ubuntu", "-c", "/usr/local/bin/sandbox /usr/bin/ruby /tmp/foo/foo.rb")
	case "java":
		copyPrepareFile(code_filepath, "/tmp/foo/Solution.java")
		copyPrepareFile(unittest_filepath, "/tmp/foo/SolutionTest.java")
		copyPrepareFile("/home/ubuntu/hamcrest-core-1.3.jar", "/tmp/foo/hamcrest-core-1.3.jar")
		copyPrepareFile("/home/ubuntu/junit-4.12.jar", "/tmp/foo/junit-4.12.jar")
		return exec.Command("lxc-attach", "-n", ephemeralImageName, "--clear-env", "--keep-var", "TERM", "--",
			"su", "-", "ubuntu", "-c", "/usr/local/bin/sandbox /usr/bin/javac -J-Xmx350m -cp '/tmp/foo/:/tmp/foo/junit-4.12.jar:/tmp/foo/hamcrest-core-1.3.jar' /tmp/foo/*.java && /usr/local/bin/sandbox /usr/bin/java -cp '/tmp/foo/:/tmp/foo/junit-4.12.jar:/tmp/foo/hamcrest-core-1.3.jar' -Xmx350m SolutionTest")
	// TODO use nodeunit: http://caolanmcmahon.com/posts/unit_testing_in_node_js/
	/* e.g. foo.js

			exports.calculate = function(num) {
		    	return num * 2;
			};

			e.g. foo_test.js

			var foo = require('./foo');

			exports['calculate'] = function(test) {
	    		test.equal(foo.calculate(2), 4);
	    		test.equal(foo.calculate(3), 5);
	    		test.done();
			};
	*/
	case "javascript":
		copyPrepareFile(code_filepath, "/tmp/foo/foo.js")
		copyPrepareFile(unittest_filepath, "/tmp/foo/foo_test.js")
		return exec.Command("lxc-attach", "-n", ephemeralImageName, "--clear-env", "--keep-var", "TERM", "--",
			"su", "-", "ubuntu", "-c", "set -o pipefail && nodeunit --reporter minimal /tmp/foo/foo_test.js | sed 's/\x1b\\[[0-9;]*m//g'")
	}
	return nil
}

func stopLxcContainer() {
	logger.Println("stopLxcContainer() entry.")

	logger.Println("Stopping container...")
	proc := exec.Command("lxc-stop", "--kill", "--name", ephemeralImageName)
	out, _ := proc.Output()
	logger.Println("Stopped container. Output: ", string(out))

	logger.Println("stopLxcContainer() exit.")
}

func startLxcContainer() error {
	logger.Println("startLxcContainer() entry.")
	defer logger.Println("startLxcContainer() exit.")

	logger.Println("Container not running, so restart it.")
	os.Mkdir("/tmp/foo", 0755)
	os.RemoveAll(lxcPath + ephemeralImageName)

	proc := exec.Command("lxc-start-ephemeral", "-d", "-o", "ubase",
		"-n", ephemeralImageName, "-b", "/tmp/foo")
	out, err := proc.Output()
	logger.Println("lxc-start-ephemeral output: ", string(out))
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to start container using lxc-start-ephemeral: %s", string(out)))
	}
	return nil
}

func restartLxcContainer() {
	var (
		err error
		retryCnt = 0
		retryLimit = 5
	)

	logger.Println("restartLxcContainer() entry.")
	lxcMutex.Lock()
	defer lxcMutex.Unlock()

	stopLxcContainer()
	for {
		err = startLxcContainer()
		if err == nil {
			break
		}
		logger.Printf("startLxcContainer failed %d: %s", retryCnt, err)
		time.Sleep(1 * time.Second)
		retryCnt += 1
		if retryCnt >= retryLimit {
			logger.Panicf("startLxcContainer failed too many times: %s", err)
		}
	}
	logger.Println("restartLxcContainer() exit.")
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	ephemeralImageName = getEphemeralImageName()
	logger.Println("main() entry. ephemeralImageName: ", ephemeralImageName)
	startLxcContainer()
	defer func() {
		lxcMutex.Lock()
		defer lxcMutex.Unlock()
		stopLxcContainer()
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/ping", pingHandler)
	mux.HandleFunc("/run/c", makeGzipHandler(runCHandler))
	mux.HandleFunc("/run/cpp", makeGzipHandler(runCPPHandler))
	mux.HandleFunc("/run/python", makeGzipHandler(runPythonHandler))
	mux.HandleFunc("/run/ruby", makeGzipHandler(runRubyHandler))
	mux.HandleFunc("/run/javascript", makeGzipHandler(runJavaScriptHandler))
	mux.HandleFunc("/run/java", makeGzipHandler(runJavaHandler))

	graceful.Run("localhost:8080", 5*time.Second, mux)
}
