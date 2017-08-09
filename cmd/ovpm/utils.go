package main

import (
	"fmt"
	"log"
	"os"

	"google.golang.org/grpc"
)

func emitToFile(filePath, content string, mode uint) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("Cannot create file %s: %v", filePath, err)

	}
	if mode != 0 {
		file.Chmod(os.FileMode(mode))
	}
	defer file.Close()
	fmt.Fprintf(file, content)
	return nil
}

func getConn(port string) *grpc.ClientConn {
	if port == "" {
		port = "9090"
	}

	conn, err := grpc.Dial(fmt.Sprintf(":%s", port), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	return conn

}
