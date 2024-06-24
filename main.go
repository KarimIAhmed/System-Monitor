package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"nhooyr.io/websocket"
)

type server struct {
	subscriberMessageBuffer int
	mux                     http.ServeMux
	subscribersMutex        sync.Mutex
	subscribers             map[*subscriber]struct{}
}

type subscriber struct {
	msgs chan []byte
}

func createNewServer() *server {
	fmt.Println("creating new server")
	s := &server{
		subscriberMessageBuffer: 5,
		subscribers:             make(map[*subscriber]struct{}),
	}
	s.mux.Handle("/", http.FileServer(http.Dir("./htmx")))
	s.mux.HandleFunc("/ws", s.subscriberHandler)
	log.Println("starting server")
	return s
}

func (s *server) subscriberHandler(writer http.ResponseWriter, request *http.Request) {
	err := s.subscribe(request.Context(), writer, request)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (s *server) addSubscriber(subscriber *subscriber) {
	s.subscribersMutex.Lock()
	s.subscribers[subscriber] = struct{}{}
	s.subscribersMutex.Unlock()
	fmt.Println("Added subscriber", subscriber)
}

func (s *server) subscribe(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
	c, err := websocket.Accept(writer, request, nil)
	if err != nil {
		log.Println("Failed to connect to websocket: ", err)
		return err
	}
	defer c.Close(websocket.StatusInternalError, "Internal error")

	subscriber := &subscriber{
		msgs: make(chan []byte, s.subscriberMessageBuffer),
	}
	s.addSubscriber(subscriber)

	ctx = c.CloseRead(ctx)
	for {
		select {
		case msg := <-subscriber.msgs:
			ctx, cancel := context.WithTimeout(ctx, time.Second)
			defer cancel()
			err := c.Write(ctx, websocket.MessageText, msg)
			if err != nil {
				log.Println("write error", err)
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func systemInfo() string {
	hostStat, _ := host.Info()
	//fmt.Println("operating system", hostStat.OS)
	//fmt.Println("platform ", hostStat.Platform)
	//fmt.Println("hostname ", hostStat.Hostname)
	//fmt.Println("num of processes", hostStat.Procs)

	memory, _ := mem.VirtualMemory()
	//fmt.Println("memory", memory.Total)
	//fmt.Println("memory ", memory.Free)
	//fmt.Println("used ", memory.UsedPercent, "%")

	html := `<div hx-swap-oob="innerHTML:#update-timestamp">` + "Total: " + hostStat.OS +
		" Platform: " + hostStat.Platform +
		" HostName : " + hostStat.Hostname +
		" Num Of Processes" + strconv.FormatUint(hostStat.Procs, 10) +
		" Memory total " + strconv.FormatUint(memory.Total, 10) +
		" Memory free " + strconv.FormatUint(memory.Free, 10) +
		" Memory Used Percent " + strconv.FormatFloat(memory.UsedPercent, 'f', 2, 64) + "%" +
		`</div>`

	return html
}

func diskInfo() string {
	diskStat, _ := disk.Usage("\\")
	//fmt.Println("mem", diskStat.Total/1073741824)
	//fmt.Println("mem", diskStat.Free/1073741824)
	//fmt.Println("mem", diskStat.UsedPercent, "%")
	html := `<div hx-swap-oob="innerHTML:#update-timestamp">` + "Total: " + strconv.FormatUint(diskStat.Total/1073741824, 10) +
		" Free: " + strconv.FormatUint(diskStat.Free/1073741824, 10) +
		" Used: " + strconv.FormatFloat(diskStat.UsedPercent, 'f', 2, 64) + "%" +
		`</div>`
	return html
}

func cpuInfo() {
	cpuStat, _ := cpu.Info()
	for i, cpuInfo := range cpuStat {
		fmt.Println("CPU [", i, "]", cpuInfo.Cores)
	}
}

func (s *server) broadcast(msg []byte) {
	s.subscribersMutex.Lock()
	for subscriber := range s.subscribers {
		select {
		case subscriber.msgs <- msg:
		default:
			fmt.Println("Dropping message for subscriber")
		}
	}
	s.subscribersMutex.Unlock()
}

func main() {
	fmt.Println("starting")
	srv := createNewServer()

	go func(s *server) {
		for {
			timestamp := time.Now().Format("2006-01-02 15:04:05")
			html := `<div hx-swap-oob="innerHTML:#update-timestamp">` + timestamp + diskInfo() + systemInfo() + `</div>`
			s.broadcast([]byte(html))
			time.Sleep(3 * time.Second)
		}
	}(srv)

	err := http.ListenAndServe(":8080", &srv.mux)
	if err != nil {
		fmt.Println(err)
	}
}
