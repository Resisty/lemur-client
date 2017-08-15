package main

import (
    "io"
    "os/exec"
    "bytes"
    "fmt"
)

func CommandPipe(exe string, cmd []string, yml string) (string) {
    var output  bytes.Buffer
    var errput     bytes.Buffer
    done := make(chan bool)
    subProcess := exec.Command(exe, cmd...)
    stdin, err := subProcess.StdinPipe()
    if err != nil {
        panic(err)
    }
    stdout, err := subProcess.StdoutPipe()
    if err != nil {
        panic(err)
    }
    stderr, err := subProcess.StderrPipe()
    if err != nil {
        panic(err)
    }
    if err := subProcess.Start(); err != nil {
        panic(err)
    }
    go func() {
        _, err := io.WriteString(stdin, yml)
        if err != nil {
            panic(err)
        }
        stdin.Close()
    }()
    go func() {
        _, err := io.Copy(&output, stdout)
        if err != nil {
            panic(err)
        }
        io.Copy(&errput, stderr)
        stdout.Close()
        stderr.Close()
        done <- true
    }()
    <-done
    if err := subProcess.Wait(); err != nil {
        fmt.Printf("Error?: %v", string(errput.Bytes()))
        panic(err)
    }
    return string(output.Bytes())
}

