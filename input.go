package main

import (
    "net"
    "bufio"
    "io"
    "encoding/json"
    "encoding/csv"
    "errors"
    "strings"
)

func MakeJsonExtractor(field string) Extractor {
    return func(s string) (net.IP, error) {
        var f interface{}
        b := []byte(s)
        if err := json.Unmarshal(b, &f); err != nil {
            return nil, err
        }
        m := f.(map[string]interface{})
        v, ok := m[field]; 
        if !ok {
            return nil, errors.New("missing field: " + field)
        }
        switch v.(type) {
        case string:
            ipString := v.(string)
            if ip := net.ParseIP(ipString); ip == nil {
                return nil, errors.New("invalid IP address: " + ipString)
            } else {
                return ip, nil
            }
        default:
            return nil, errors.New("invalid field: " + field)
        }
    }
}

func MakeCsvExtractor(column int) Extractor {
    return func(s string) (net.IP, error) {
        sr := strings.NewReader(s)
        cr := csv.NewReader(sr)
        records, err := cr.Read() 
        if err != nil {
            return nil, err
        }
        ipString := records[column]
        ip := net.ParseIP(ipString)
        if ip == nil {
            return nil, errors.New("invalid IP address: " + ipString)
        }
        return ip, nil
    }
}

func MakeLineSplitter(extract Extractor) func(lineChan chan string, ipChan chan net.IP, rewriteChan chan string) {
    return func(lineChan chan string, ipChan chan net.IP, rewriteChan chan string) {
        for line := range lineChan {
            // Push the line to the output file rewriter
            rewriteChan <- line
            // Grab the IP from the line, send it to stdout writer
            if (config.SuccessOnly) {
                if (strings.Contains(line, "synack")) {
                    if ip, err := extract(line); err == nil {
                        ipChan <- ip
                    }
                }
            } else {
                if ip, err := extract(line); err == nil {
                    ipChan <- ip
                }
            }
        }
        close(ipChan)
        close(rewriteChan)
    }
}

func GobbleInput(inputFile io.Reader, lineChan chan string) {
    scanner := bufio.NewScanner(inputFile)
    for scanner.Scan() {
        line := scanner.Text()
        lineChan <- line
    }
    close(lineChan)
}

type Extractor func(string) (net.IP, error)
