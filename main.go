package main

import (
	"flag"
	"fmt"
	"github.com/go-vgo/robotgo"
	"github.com/hypebeast/go-osc/osc"
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
	OSC_PORT = 32901
	BUFSIZE = 1024
	AS_BIN_PATH = "/usr/bin/osascript"
	PRINT_SCRIPT_PATH = "as/print.scpt"
	HAND_SCRIPT_PATH = "as/hand.scpt"
	ERASER_SCRIPT_PATH = "as/eraser.scpt"
	MERGE_SCRIPT_PATH = "as/merge.scpt"
	SENSITIVE = 500
)

const (
	SX_INDEX = 0
	SY_INDEX = 1
	SZ_INDEX = 2
	ERASER_INDEX = 3
	HAND_INDEX = 4
	MERGE_INDEX = 5
	PRINT_INDEX = 6
)

var (
	record         = flag.Bool("record", false, "recoed sensor data")
	play           = flag.String("play", "", "play from record")
	disableControl = flag.Bool("disable-control", false, "disable disableControl")
	file           *os.File
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

func SendOSCFloat(client *osc.Client, value float32, path string) error {
	message := osc.NewMessage(path)
	message.Append(value)
	if err := client.Send(message); err != nil {
		return err
	}
	return nil
}

func SendOSCInt(client *osc.Client, value int32, path string) error {
	message := osc.NewMessage(path)
	message.Append(value)
	if err := client.Send(message); err != nil {
		return err
	}
	return nil
}

func SendOSCString(client *osc.Client, value string, path string) error {
	message := osc.NewMessage(path)
	message.Append(value)
	if err := client.Send(message); err != nil {
		return err
	}
	return nil
}

func main() {
	flag.Parse()

	sensorDataCh := make(chan string)

	client := osc.NewClient("127.0.0.1", OSC_PORT)
	if client == nil {
		fmt.Printf("failed to create osc client")
		return
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("failed to get ip address")
		return
	}
	for _, i := range ifaces {
		addrs, _ := i.Addrs()
		for _, a := range addrs {
			if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					fmt.Printf("enter ip address [%s]\n", ipnet.IP.String())
				}
			}
		}
	}

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
				fileName := fmt.Sprintf("data/%d.log", time.Now().Unix())
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
			fmt.Printf("sensor data recieve disableControl loop\n")
			for {
				length, addr, _ := conn.ReadFrom(buffer)
				//spk := strings.Split(string(buffer[:length]), ",")
				if *record {
					file.Write(([]byte)(fmt.Sprintf("%s\n", buffer[:length])))
				}
				sensorDataCh <- string(buffer[:length])

				if string(buffer[:length]) == "hello" {
					conn.WriteTo([]byte("fine"), addr)
				}
			}
		}()
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		fmt.Printf("start mouse disableControl loop\n")
		width, height := robotgo.GetScreenSize()
		prevTool := 0
		fmt.Printf("screen size: %d x %d \n", width, height)
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
					// sx, sy, sz, eraser, hand, merge, print
					sx, _ := strconv.ParseFloat(spl[SX_INDEX], 64)
					sy, _ := strconv.ParseFloat(spl[SY_INDEX], 64)
					sz, _ := strconv.ParseFloat(spl[SZ_INDEX], 64)
					eraser, _ := strconv.ParseInt(spl[ERASER_INDEX], 10,32)
					hand, _ := strconv.ParseInt(spl[HAND_INDEX], 10,32)
					merge, _ := strconv.ParseInt(spl[MERGE_INDEX], 10,32)
					print, _ := strconv.ParseInt(spl[PRINT_INDEX], 10, 32)

					// OSC data
					SendOSCFloat(client, float32(sx), "/accell/x")
					SendOSCFloat(client, float32(sy), "/accell/y")
					SendOSCFloat(client, float32(sz), "/accell/z")
					SendOSCInt(client, int32(eraser), "/eraser")
					SendOSCInt(client, int32(hand), "/hand")
					SendOSCInt(client, int32(merge), "/merge")
					SendOSCInt(client, int32(print), "/print")

					//fmt.Printf("sx: %f sy: %f sz: %f eraser %d, hand: %d, merge: %d, print: %d: \n",  sx, sy, sz, eraser, hand, merge, print)

					if !*disableControl {
						// TODO: photoshop の描画領域に沿って、移動領域を決定できるようにする
						// MEMO: 500 の部分はセンシとして考えられる

						if print == 1 {
							//TODO: consider this process is blocking
							err := exec.Command(AS_BIN_PATH, PRINT_SCRIPT_PATH).Start()
							if err != nil {
								fmt.Printf("failed to exec print command %s\n", err)
							}
						}
						if merge == 1 {
							err := exec.Command(AS_BIN_PATH, MERGE_SCRIPT_PATH).Start()
							if err != nil {
								fmt.Printf("failed to exec merge command %s\n", err)
							}
						}

						if ERASER_INDEX != prevTool && eraser == 1 {
							fmt.Println("change to ERASER tool")
							// exec eraser script
							err := exec.Command(AS_BIN_PATH, ERASER_SCRIPT_PATH).Run()
							if err != nil {
								fmt.Printf("failed to exec eraser command %s\n", err)
							}
						}
						if HAND_INDEX != prevTool && hand == 1 {
							fmt.Println("change to HAND tool")
							err := exec.Command(AS_BIN_PATH, HAND_SCRIPT_PATH).Run()
							if err != nil {
								fmt.Printf("failed to exec hand command %s\n", err)
							}
						}
						nx := int(float64(width/2) + (sx * SENSITIVE))
						ny := int(float64(height/2) + (-sy * SENSITIVE))
						if eraser == 1 || hand == 1 {
							robotgo.DragSmooth(nx, ny,  0.5)
						} else {
							robotgo.MoveSmooth(nx, ny,  0.5)
						}
					}

					// keep previous state
					if eraser == 1 {
						prevTool = ERASER_INDEX
					}
					if hand == 1 {
						prevTool = HAND_INDEX
					}
				}
			}
		}
	}()
	wg.Wait()
}
