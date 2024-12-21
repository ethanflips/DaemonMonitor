package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
)

const dataURL = "http://10.101.20.10:3000/game-servers/daemon-states"
const fetchInterval = 25 * time.Second
const ntfyEstopURL = "https://ntfy.sh/ethandaemonalerts555"
const ntfyErrURL = "https://ntfy.sh/ethandaemonalerts556"

type DaemonState struct {
	Number     int    `json:"number"`
	Hostname   string `json:"hostname"`
	Status     string `json:"status"`
	SessionID  string `json:"session_id"`
	UIPath     string `json:"ui_path"`
	Server     string `json:"server"`
	Client     string `json:"client"`
	RFactor    string `json:"r_factor"`
	Difficulty string `json:"difficulty"`
	Track      string `json:"track"`
	Phase      string `json:"phase"`
	Applied    string `json:"applied"`
	Error      string `json:"error"`
	State      string `json:"state"`
	Last       string `json:"last"`
	Timestamp  string `json:"timestamp"`
}

var latestStates []DaemonState

var estopMap = make(map[string]bool)
var errorMap = make(map[string]string)

var m = sync.RWMutex{}
var wg = sync.WaitGroup{}

func DataFetch(url string) ([]DaemonState, error) {
	fmt.Printf("Fetching New Data... %s\n", time.Now())
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	tableSelector := `table tbody tr`
	var rows []string

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(tableSelector),
		chromedp.Evaluate(`
			Array.from(document.querySelectorAll('table tbody tr')).map(row => 
				Array.from(row.cells).map(cell => cell.textContent.trim()).join('|')
			)
		`, &rows),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch table: %v", err)
	}

	currentTime := time.Now().Format("2006-01-02 15:04:05")
	var states []DaemonState
	for _, row := range rows {
		cols := strings.Split(row, "|")
		if len(cols) < 14 {
			continue
		}

		state := DaemonState{
			Number:     ParseInt(cols[0]),
			Hostname:   TrimHostname(cols[1]),
			Status:     cols[2],
			SessionID:  cols[3],
			UIPath:     cols[4],
			Server:     cols[5],
			Client:     cols[6],
			RFactor:    cols[7],
			Difficulty: cols[8],
			Track:      cols[9],
			Phase:      cols[10],
			Applied:    cols[11],
			Error:      cols[12],
			State:      cols[13],
			Last:       cols[14],
			Timestamp:  currentTime,
		}
		m.Lock()
		CheckEstop(state)
		CheckSimErrors(state)
		states = append(states, state)
		m.Unlock()

	}
	fmt.Printf("Success!\n")
	latestStates = states
	return states, nil
}

// only sends notification
func CheckEstop(sim DaemonState) {
	checkTerm := "estop"
	host := sim.Hostname
	hasEstop := strings.Contains(sim.State, checkTerm)
	if hasEstop && !estopMap[host] {
		noti := fmt.Sprintf("Sim %d | ESTOP", sim.Number)
		SendSimNoti(noti, ntfyEstopURL)
		estopMap[host] = true
	} else if !hasEstop && estopMap[host] {
		delete(estopMap, host)
	}
}

func CheckSimErrors(sim DaemonState) {
	// check for Looping
	host := sim.Hostname
	serverStatus := strings.ToLower(sim.Server)
	rf2Status := strings.ToLower(sim.RFactor)
	serverFailed := strings.Contains(serverStatus, "failedtostart")
	alreadyFailed := strings.Contains(errorMap[host], "fail")
	rf2Crashed := strings.Contains(rf2Status, "crashed")
	alreadyCrashed := strings.Contains(errorMap[host], "crashed")

	if serverFailed && !alreadyFailed {
		SendSimNoti(fmt.Sprintf("%s | Server Failed", host), ntfyErrURL)
		errorMap[host] = "server failed"
	} else if !serverFailed && alreadyFailed {
		delete(errorMap, host)
	}
	if rf2Crashed && !alreadyCrashed {
		SendSimNoti(fmt.Sprintf("%s | Crashed", host), ntfyErrURL)
		errorMap[host] = "crashed"
	} else if !rf2Crashed && alreadyCrashed {
		delete(errorMap, host)
	}

}

func SendSimNoti(message string, url string) {
	http.Post(url, "text/plain",
		strings.NewReader(message))
	fmt.Printf("\nSending Notification:\n%s\nTO: %s\n\n", message, url)
}
func TrimHostname(hostname string) string {
	snips := strings.SplitN(hostname, ",", 2)
	return snips[1]
}
func ParseInt(value string) int {
	num, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return num
}

func main() {
	fmt.Printf("Daemon Monitor Starting...\n")
	go StartWebService()
	for {
		go DataFetch(dataURL)
		time.Sleep(20 * time.Second)

	}
}
