package main

import (
	"flag"
	"fmt"
	"github.com/go-vgo/robotgo"
	"github.com/hypebeast/go-osc/osc"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"test-control-mouse/sensor-reciver/control"
	"test-control-mouse/sensor-reciver/network"
	"test-control-mouse/sensor-reciver/processing"
	"time"
)

const (
	PORT         = 32900
	DEFAULT_HOST = "0.0.0.0"
	OSC_PORT     = 32901
	SENSITIVE    = 500

	ALLOWED_DELAY = 100000
)

var (
	record         = flag.Bool("record", false, "recoed sensor data")
	play           = flag.String("play", "", "play from record")
	disableControl = flag.Bool("disable-control", false, "disable disableControl")
	file           *os.File
)

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

			playStartTime := strings.Replace(*play, "data/", "", -1)
			playStartTime = strings.Replace(playStartTime, ".log", "", -1)
			//playStartTimeInt, err := strconv.ParseFloat(playStartTime, 32)

			bytes, err := ioutil.ReadFile(*play)
			if err != nil {
				panic(err)
			}
			linedBuffer := strings.Split(string(bytes), "\n")
			for _, line := range linedBuffer {
				time.Sleep((1000 / 15) * time.Millisecond)
				if line != "" {
					spl := strings.Split(line, ",")
					genTime, _ := strconv.ParseFloat(spl[0], 64)
					genTime /= 1000
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
				fmt.Printf("record flag is enabled. %s\n", fileName)
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
		controller := control.NewPCController()
		xyz := processing.NewXYZ()
		if *disableControl {
			controller.ToggleDisable()
		}
		fmt.Printf("start mouse disableControl loop\n")
		width, height := robotgo.GetScreenSize()
		prevTool := 0
		fmt.Printf("screen size: %d x %d \n", width, height)
		network.SendOSCInt(client, int32(1), "/init")

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
					atTime, _ := strconv.ParseFloat(spl[control.AT_TIME], 64)
					sx, _ := strconv.ParseFloat(spl[control.SX_INDEX], 64)
					sy, _ := strconv.ParseFloat(spl[control.SY_INDEX], 64)
					sz, _ := strconv.ParseFloat(spl[control.SZ_INDEX], 64)
					xyz.AddData(sx, sy, sz)

					eraser, _ := strconv.ParseInt(spl[control.ERASER_INDEX], 10, 32)
					hand, _ := strconv.ParseInt(spl[control.HAND_INDEX], 10, 32)
					merge, _ := strconv.ParseInt(spl[control.MERGE_INDEX], 10, 32)
					ratio, _ := strconv.ParseInt(spl[control.RATIO_INDEX], 10, 32)
					volume, _ := strconv.ParseInt(spl[control.VOLUME_INDEX], 10, 32)

					// OSC data
					network.SendOSCFloat(client, float32(xyz.GetXAX()), "/accell/x")
					network.SendOSCFloat(client, float32(xyz.GetYAX()), "/accell/y")
					network.SendOSCFloat(client, float32(xyz.GetZAX()), "/accell/z")
					network.SendOSCInt(client, int32(eraser), "/eraser")
					network.SendOSCInt(client, int32(hand), "/hand")
					network.SendOSCInt(client, int32(merge), "/merge")
					//network.SendOSCInt(client, int32(print), "/print")
					network.SendOSCInt(client, int32(volume), "/volume")

					cTimemill := time.Now().UTC().UnixNano() / int64(time.Millisecond)
					diff := float64(cTimemill) - atTime

					if diff > ALLOWED_DELAY && *play == "" {
						fmt.Printf("%f millisecond delayed data is ignored\n", diff)
						continue
					} else {
						fmt.Print("----------------------------------\n")
						fmt.Printf("moving average: %f, %f, %f, acceleration: %f, %f, %f \n", xyz.GetXMA(), xyz.GetYMA(), xyz.GetZMA(), xyz.GetXAX(), xyz.GetYAX(), xyz.GetZAX())
						fmt.Printf("diff: %f before, sx: %f sy: %f sz: %f eraser %d, hand: %d, merge: %d, ratio: %d, volume: %d\n", diff, sx, sy, sz, eraser, hand, merge, ratio, volume)
					}

					//if print == 1 {
					//	if err := controller.Print(); err != nil {
					//		fmt.Printf("%s", err)
					//	}
					//}
					if merge == 1 {
						if err := controller.Merge(); err != nil {
							fmt.Printf("%s", err)
						}
					}

					if control.ERASER_INDEX != prevTool && eraser == 1 {
						if err := controller.ChangeTool(control.ERASER_INDEX); err != nil {
							fmt.Printf("%s", err)
						}
					}
					if control.HAND_INDEX != prevTool && hand == 1 {
						if err := controller.ChangeTool(control.HAND_INDEX); err != nil {
							fmt.Printf("%s", err)
						}
					}

					width = control.DRAWABLE_AREA_WIDTH
					height = control.DRAWABLE_AREA_HEIGHT
					nx := int(float64(width/2) + (sx * float64(ratio)))
					ny := int(float64(height/2) + (-sy * float64(ratio)))
					if eraser == 1 || hand == 1 {
						controller.MouseDrag(nx, ny)
					} else {
						controller.MouseMove(nx, ny)
					}

					// keep previous state
					if eraser == 1 {
						prevTool = control.ERASER_INDEX
					}
					if hand == 1 {
						prevTool = control.HAND_INDEX
					}
				}
			}
		}
	}()
	wg.Wait()
}
