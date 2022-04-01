package main

import (
	"fmt"
	"net/http"
	"os"
)

const (
	listenAddr = ":8080"
	macPath    = "/latest/meta-data/mac"
	cidrPath   = "/latest/meta-data/network/interfaces/macs/ff:ff:ff:ff:ff:ff/vpc-ipv4-cidr-block"
	fakeMac    = "ff:ff:ff:ff:ff:ff"
	fakeCIDR   = "10.0.0.0/16"
)

func main() {
	http.HandleFunc(macPath, func(writer http.ResponseWriter, req *http.Request) {
		fmt.Println("Received request for mac address")
		_, err := writer.Write([]byte(fakeMac))
		if err != nil {
			fmt.Fprintf(os.Stderr,
				"Failed to write response body: %s", err)
			writer.WriteHeader(http.StatusServiceUnavailable)
		}
	})

	http.HandleFunc(cidrPath, func(writer http.ResponseWriter, req *http.Request) {
		fmt.Println("Received request for vpc cidr")
		_, err := writer.Write([]byte(fakeCIDR))
		if err != nil {
			fmt.Fprintf(os.Stderr,
				"Failed to write response body: %s", err)
			writer.WriteHeader(http.StatusServiceUnavailable)
		}
	})

	fmt.Println("Starting fake IMDS v1 server...")
	http.ListenAndServe(listenAddr, nil)
}
