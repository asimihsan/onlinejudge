package main

/*
curl -X POST -H "Content-Type: text/plain" --compressed \
    --data-binary @foo.py http://localhost:8080/run/python
 */

// lxc-start-ephemeral -d -o ubase -n u1 -b /tmp/foo

import (
    "compress/gzip"
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

func runHandler(language string, w http.ResponseWriter, r *http.Request) {
    logger = getLogger(getLogPill())
    logger.Println("runHandler() entry.")

    // Put data into the channel which acts like a semaphore. Once it reaches
    // it's capacity it will block here; hence only "maxOutstandingRequests"
    // permitted.
    handlerSemaphore <- 1
    codeFile := prepareCodeFile(logger, w, r)
    defer os.Remove(codeFile.Name())
    outputFile := prepareOutputFile(logger, w, r)
    defer os.Remove(outputFile.Name())
    cmd := runCommand(language, codeFile.Name())
    runCode(cmd, codeFile, outputFile, logger, w, r)
    <-handlerSemaphore

    logger.Println("runHandler() exit.")
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

func prepareCodeFile(logger *log.Logger, w http.ResponseWriter, r *http.Request) *os.File {
    logger.Println("prepareCodeFile() entry.")
    codeFile, err := ioutil.TempFile("", "run")
    if err != nil {
        logger.Panicf("couldn't create temporary code file", err)
    }
    logger.Println("temporary codeFile: ", codeFile.Name())
    logger.Println("prepareCodeFile() writing code to file...")
    io.Copy(codeFile, r.Body)
    codeFile.Close()
    logger.Println("prepareCodeFile() finished writing code to file.")
    logger.Println("prepareCodeFile() exit.")
    return codeFile
}

func prepareOutputFile(logger *log.Logger, w http.ResponseWriter, r *http.Request) *os.File {
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
             logger *log.Logger, w http.ResponseWriter, r *http.Request) {
    logger.Println("runCode() entry.")
    cmd.Stdout = outputFile
    cmd.Stderr = outputFile
    logger.Println("runCode() running file...")
    err := cmd.Start()
    if err != nil {
        logger.Panicf("failed to run command", err)
    }

    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "POST")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
    w.Header().Set("Content-Type", "text/plain")

    done := make(chan error, 1)
    go func() {
        done <- cmd.Wait()
    }()
    select {
        case <- time.After(5 * time.Second):
            if err := cmd.Process.Kill(); err != nil {
                logger.Fatalf("failed to kill: ", err)
                <- done  // allow goroutine to exit
            }
            msg := "<process ran for too long. output is below>\n"
            logger.Println(msg)
            io.WriteString(w, msg)
        case err := <- done:
            if err != nil {
                msg := fmt.Sprintf("<process finished with error: %s. output is below>\n", err)
                logger.Println(msg)
                io.WriteString(w, msg)
            }
    }

    logger.Println("runCode() finished file.")

    logger.Println("runCode() returning output...")
    _, _ = outputFile.Seek(0, 0)
    // more than this causes gzip encoder to crash with null pointer exception
    written, err := io.CopyN(w, outputFile, 1 * 1024 * 1024)
    if written == 1 * 1024 * 1024 {
        io.WriteString(w, "\n<too much output, truncated>\n")
    }

    logger.Println("runCode() finished returning output.")

    logger.Println("runCode() exit.")
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
        return exec.Command("lxc-attach", "-n", "u1", "--",
            "su", "-", "ubuntu", "-c", "/home/ubuntu/sandbox/sandbox /usr/bin/python /tmp/foo/foo.py")
    case "ruby":
        if err := exec.Command("cp", "-f", filepath, "/tmp/foo/foo.rb").Run(); err != nil {
            logger.Panicf("failed to copy code to /tmp/foo/foo.py")
        }
        if err := os.Chmod("/tmp/foo/foo.rb", 0755); err != nil {
            logger.Panicf("failed to chmod /tmp/foo/foo.rb")
        }
        return exec.Command("lxc-attach", "-n", "u1", "--",
            "/home/ubuntu/sandbox/sandbox", "/usr/bin/ruby", "/tmp/foo/foo.rb")
    case "java":
        if err := exec.Command("cp", "-f", filepath, "/tmp/foo/Solution.java").Run(); err != nil {
            logger.Panicf("failed to copy code to /tmp/foo/Solution.java")
        }
        if err := os.Chmod("/tmp/foo/Solution.java", 0755); err != nil {
            logger.Panicf("failed to chmod /tmp/foo/Solution.java")
        }
        return exec.Command("lxc-attach", "-n", "u1", "--",
            "/bin/bash", "-c", "rm -f /tmp/foo/*.class && /home/ubuntu/sandbox/sandbox /usr/bin/javac -J-Xmx350m /tmp/foo/Solution.java && /home/ubuntu/sandbox/sandbox /usr/bin/java -Xmx350m -classpath /tmp/foo Solution")
    }
    return nil
}

func ensureLxcContainerIsRunning() {
    logger.Println("ensureLxcContainerIsRunning() entry.")

    logger.Println("Stopping container...")
    proc := exec.Command("lxc-stop", "--timeout", "5", "--name", "u1")
    proc.Start()
    proc.Wait()
    logger.Println("Stopped container.")

    proc = exec.Command("lxc-info", "-n", "u1")
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
    proc2 := exec.Command("lxc-start-ephemeral", "-d", "-o", "ubase",
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
