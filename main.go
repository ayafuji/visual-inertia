package main

import (
	"flag"
	"fmt"
	"github.com/go-vgo/robotgo"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	PORT = 32900
	DEFAULT_HOST = "0.0.0.0"
	BUFSIZE = 1024
)

var (
	record = flag.Bool("record", false, "recoed sensor data")
	play = flag.String("play", "", "play from record")
	file *os.File
)

func getRandomInt(min, max int) int {
	return rand.Intn(max - min) + min
}

func Print() error {
	err := exec.Command("/usr/bin/osascript", "push_lcmd.scpt").Run()
	if err != nil {
		fmt.Printf("failed to exec print jobs %s\n", err)
		return err
	}
	return nil
}

func main() {
	flag.Parse()

	sensorDataCh := make(chan string)

	if *play != "" {
		go func() {
			fmt.Printf("play from record. %s\n", *play)
			bytes, err := ioutil.ReadFile(*play)
			if err != nil {
				panic(err)
			}
			linedBuffer := strings.Split(string(bytes), "\n")
			for _, line := range linedBuffer {
				time.Sleep((1000 / 15) * time.Millisecond)
				if line != "" {
					sensorDataCh <- line
				}
			}
			sensorDataCh <- "done"
		}()
	} else {
		go func() {
			if *record {
				var err error
				fileName := fmt.Sprintf("%d.log", time.Now().Unix())
				file, err = os.Create(fileName)
				if err != nil {
					// Openエラー処理
				}
				defer file.Close()
				fmt.Printf("record flag is enabled. %s", fileName)
			}

			conn, _ := net.ListenPacket("udp", fmt.Sprintf("%s:%d", DEFAULT_HOST, PORT))
			defer conn.Close()
			buffer := make([]byte, 1500)
			fmt.Printf("sensor data recieve control loop\n")
			for {
				length, _, _ := conn.ReadFrom(buffer)
				//spk := strings.Split(string(buffer[:length]), ",")
				//fmt.Printf("Received x: %s, y: %s, z: %s \n", spk[0], spk[1], spk[1])
				if *record {
					file.Write(([]byte)(fmt.Sprintf("%s\n", buffer[:length])))
				}
				sensorDataCh <- string(buffer[:length])
			}
		}()
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		fmt.Printf("start mouse control loop\n")
		for {
			select {
			case str := <-sensorDataCh:
				//fmt.Printf("got data: %s\n", str)
				if str == "done" {
					fmt.Println("record file is finished")
					wg.Done()
					break
				} else {
					spl := strings.Split(str, ",")
					if len(spl) < 5 {
						return
					}
					sx, _ := strconv.ParseFloat(spl[0], 64)
					sy, _ := strconv.ParseFloat(spl[1], 64)
					sz, _ := strconv.ParseFloat(spl[2], 64)
					is_drag := false
					if spl[3] == "1" {
						is_drag = true
					}
					// TODO: photoshop の描画領域に沿って、移動領域を決定できるようにする
					// MEMO: 500 の部分はセンシとして考えられる
					nx := int((1920/2) + (sx * 500))
					ny := int((1080/2) + (-sy * 500))
					if is_drag {
						robotgo.DragSmooth(nx, ny,  0.5)
					} else {
						robotgo.MoveSmooth(nx, ny,  0.5)
					}
					fmt.Printf("to x: %d y: %d, sx: %f sy: %f sz: %f is_drag: %v \n", nx, ny, sx, sy, sz, is_drag)
				}
			}
		}
	}()
	wg.Wait()
}
