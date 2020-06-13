package main

// #include <stdio.h>
// void testEcho() {
// 	fprintf(stdout, "stdout\n");
//  fprintf(stderr, "stderr\n");
// }
import "C"
import (
	"fmt"
	"go-capturer/capture"
)

func main() {
	t, err := capture.CaptureCgo(func() {
		C.testEcho()
	}, capture.CaptureFlag{
		CaptureStdout: true,
		CaptureStderr: true,
	})
	if err != nil {
		fmt.Printf("err: %s", err)
	} else {
		fmt.Printf("output: %s", string(t))
	}
	t2, err2 := capture.CaptureCgo(func() {
		C.testEcho()
	}, capture.CaptureFlag{
		CaptureStdout: true,
		CaptureStderr: false,
	})
	if err2 != nil {
		fmt.Printf("err: %s", err2)
	} else {
		fmt.Printf("output: %s", string(t2))
	}
	t3, err3 := capture.CaptureCgo(func() {
		C.testEcho()
	}, capture.CaptureFlag{
		CaptureStdout: false,
		CaptureStderr: false,
	})
	if err3 != nil {
		fmt.Printf("err: %s", err3)
	} else {
		fmt.Printf("output: %s", string(t3))
	}
}
