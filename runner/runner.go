package main

/*
curl -X POST -H "Content-Type: text/plain" --compressed \
    --data-binary @foo.py http://localhost:8080/run/python
 */

// lxc-start-ephemeral -d -o ubase -n u1 -b /tmp/foo

import (
    "bytes"
    "compress/gzip"
    "encoding/json"
    "fmt"
    "io"
    "io/ioutil"
    "log"
    "math/rand"
    "net/http"
    "os"
    "os/exec"
    "strings"
    "time"
)

var (
    logger = getLogger("logger")    
    letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
    maxOutstandingRequests = 1
    handlerSemaphore = make(chan int, maxOutstandingRequests)
    grecaptchaSecret = "6LcB8gATAAAAAByLaeJzveuN4_lP_yDdiszVoL60"
    outputLimit = 10 * 1024
)

func getLogger(prefix string) *log.Logger {
    paddedPrefix := fmt.Sprintf("%-8s: ", prefix)
    return log.New(os.Stdout, paddedPrefix,
        log.Ldate | log.Ltime | log.Lmicroseconds)    
}

func getLogPill() string {
    b := make([]rune, 8)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
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
    Code string
    Recaptcha string
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
    w.Header().Set("Content-Type", "application/json")

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

    codeFile := prepareCodeFile(logger, t.Code)
    defer os.Remove(codeFile.Name())

    outputFile := prepareOutputFile(logger)
    defer os.Remove(outputFile.Name())

    cmd := runCommand(language, codeFile.Name())
    runCode(cmd, codeFile, outputFile, logger, w, response)

    if _, ok := response["success"]; !ok {
        response["success"] = true
    }
    if val, _ := response["success"]; val != true {
        // Something went wrong while running the code. There's a bug in Java
        // where doing an infinite print means we can't run Java any more. By
        // this I mean you run '/usr/local/bin/sandbox /usr/bin/javac' or java
        // and it just quits with return code 0. This is odd, and strace isn't
        // revealing. So for now let's always cycle the LXC container on any
        // sort of failure.
        logger.Println("code failure, so cycle the LXC container.")
        ensureLxcContainerIsRunning()
    }

    <-handlerSemaphore    
}

func runPythonHandler (w http.ResponseWriter, r *http.Request) {
    runHandler("python", w, r)
}

func runRubyHandler (w http.ResponseWriter, r *http.Request) {
    runHandler("ruby", w, r)
}

func runJavaHandler (w http.ResponseWriter, r *http.Request) {
    runHandler("java", w, r)
}

func prepareCodeFile(logger *log.Logger, code string) *os.File {
    logger.Println("prepareCodeFile() entry.")
    codeFile, err := ioutil.TempFile("", "run")
    if err != nil {
        logger.Panicf("couldn't create temporary code file", err)
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
        logger.Panicf("couldn't create temporary output file", err)
    }
    logger.Println("temporary outputFile: ", outputFile.Name())
    logger.Println("prepareOutputFile() exit.")
    return outputFile
}

func runCode(cmd *exec.Cmd, codeFile *os.File, outputFile *os.File,
             logger *log.Logger, w http.ResponseWriter, response map[string]interface{}) {
    logger.Println("runCode() entry.")
    defer logger.Println("runCode() exit.")
    cmd.Stdout = outputFile
    cmd.Stderr = outputFile
    logger.Println("runCode() running file...")
    err := cmd.Start()
    if err != nil {
        logger.Panicf("failed to run command", err)
    }

    done := make(chan error, 1)
    go func() {
        done <- cmd.Wait()
    }()
    var outputBuffer bytes.Buffer
    select {
        case <- time.After(5 * time.Second):
            if err := cmd.Process.Kill(); err != nil {
                logger.Fatalf("failed to kill: ", err)
                <- done  // allow goroutine to exit
            }
            msg := "<process ran for too long. output is below>\n"
            logger.Println(msg)
            response["success"] = false
            outputBuffer.WriteString(msg)
        case err := <- done:
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

func runCommand(language string, filepath string) *exec.Cmd {
    os.Mkdir("/tmp/foo", 0755)
    switch language {
    case "python":
        if err := exec.Command("cp", "-f", filepath, "/tmp/foo/foo.py").Run(); err != nil {
            logger.Panicf("failed to copy code to /tmp/foo/foo.py")
        }
        if err := os.Chmod("/tmp/foo/foo.py", 0755); err != nil {
            logger.Panicf("failed to chmod /tmp/foo/foo.py")
        }
        return exec.Command("ssh", "ubuntu@localhost", "lxc-attach", "-n", "u1", "--clear-env", "--keep-var", "TERM", "--",
            "/usr/local/bin/sandbox", "/usr/bin/python", "/tmp/foo/foo.py")
    case "ruby":
        if err := exec.Command("cp", "-f", filepath, "/tmp/foo/foo.rb").Run(); err != nil {
            logger.Panicf("failed to copy code to /tmp/foo/foo.py")
        }
        if err := os.Chmod("/tmp/foo/foo.rb", 0755); err != nil {
            logger.Panicf("failed to chmod /tmp/foo/foo.rb")
        }
        return exec.Command("ssh", "ubuntu@localhost", "lxc-attach", "-n", "u1", "--clear-env", "--keep-var", "TERM", "--",
            "/usr/local/bin/sandbox", "/usr/bin/ruby", "/tmp/foo/foo.rb")
    case "java":
        if err := exec.Command("cp", "-f", filepath, "/tmp/foo/Solution.java").Run(); err != nil {
            logger.Panicf("failed to copy code to /tmp/foo/Solution.java")
        }
        if err := os.Chmod("/tmp/foo/Solution.java", 0755); err != nil {
            logger.Panicf("failed to chmod /tmp/foo/Solution.java")
        }
        if err := exec.Command("rm", "-f", "/tmp/foo/*.class").Run(); err != nil {
            logger.Panicf("failed to clean up old class files in /tmp/foo")
        }
        return exec.Command("ssh", "ubuntu@localhost", "lxc-attach", "-n", "u1", "--clear-env", "--keep-var", "TERM", "--",
            "/bin/sh", "-c", "/usr/local/bin/sandbox /usr/bin/javac -J-Xmx350m /tmp/foo/Solution.java && /usr/local/bin/sandbox /usr/bin/java -Xmx350m -classpath /tmp/foo Solution")
    }
    return nil
}

func ensureLxcContainerIsRunning() {
    logger.Println("ensureLxcContainerIsRunning() entry.")

    logger.Println("Stopping container...")
    proc := exec.Command("ssh", "ubuntu@localhost", "lxc-stop", "--timeout", "1", "--name", "u1")
    proc.Start()
    proc.Wait()
    logger.Println("Stopped container.")

    proc = exec.Command("ssh", "ubuntu@localhost", "lxc-info", "-n", "u1")
    err := proc.Start()
    if err != nil {
        logger.Panicf("Failed to start lxc-info", err)
    }
    err = proc.Wait()
    if err == nil {
        logger.Println("Container is already running.")
        logger.Println("ensureLxcContainerIsRunning() exit.")
        return
    }
    logger.Println("Container not running, so restart it.")
    os.Mkdir("/tmp/foo", 0755)
    proc2 := exec.Command("ssh", "ubuntu@localhost", "lxc-start-ephemeral", "-d", "-o", "ubase",
        "-n", "u1", "-b", "/tmp/foo")
    err2 := proc2.Start() 
    if err2 != nil {
        logger.Panicf("Failed to start container using lxc-start-ephemeral", err)
    }
    err2 = proc2.Wait()
    if err2 != nil {
        logger.Panic(err)
    }
    logger.Println("ensureLxcContainerIsRunning() exit.")
}

func main() {
    logger.Println("main() entry.")
    ensureLxcContainerIsRunning()
    rand.Seed(time.Now().UTC().UnixNano())
    http.HandleFunc("/run/python", makeGzipHandler(runPythonHandler))
    http.HandleFunc("/run/ruby", makeGzipHandler(runRubyHandler))
    http.HandleFunc("/run/java", makeGzipHandler(runJavaHandler))
    err := http.ListenAndServe("localhost:8080", nil)
    if err != nil {
        logger.Panic(err)
    }
}
