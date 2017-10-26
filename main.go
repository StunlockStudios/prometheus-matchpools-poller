package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type poolsResponse struct {
	Pools []matchpoolInfo `json:""`
}
type matchpoolInfo struct {
	PoolId              string   `json:"poolId"`
	Pid                 int      `json:"pid"`
	Name                string   `json:"name"`
	Region              string   `json:"region"`
	RegionPingHost      string   `json:"regionPingHost"`
	LoadAvailable       int      `json:"loadAvailable"`
	LoadTimestamp       uint64   `json:"loadTimestamp"`
	MatchConnectHost    string   `json:"matchConnectHost"`
	ApiUri              string   `json:"apiUri"`
	OriginalApiUri      string   `json:"originalApiUri"`
	State               string   `json:"state"`
	SupportedMatchTypes []string `json:"supportedMatchTypes"`
	GameplayVersion     int      `json:"gameplayVersion"`
	Revision            int      `json:"revision"`
	PooledMatches       int      `json:"pooledMatches"`
	ActiveMatches       int      `json:"activeMatches"`
	LoadMOdifier        float32  `json:"loadModifier"`
	LoadLimit           int      `json:"loadLimit"`
	TargetFps           float32  `json:"targetFps"`
	Uptime              uint     `json:"uptime"`
	UsersReserved       int      `json:"usersReserved"`
	UsersConnectetd     int      `json:"usersConnected"`
	UsersDisconnected   int      `json:"usersDisconnected"`
	UsersKicked         int      `json:"usersKicked"`
	BotsActive          int      `json:"botsActive"`
	CoreCount           int      `json:"coreCount"`
	ThreadCount         int      `json:"threadCount"`
	Exceptions          int      `json:"exceptions"`
	Errors              int      `json:"errors"`
	Warnings            int      `json:"warnings"`
	IgnoreSignals       bool     `json:"ignoreSignals"`
	AverageFps          float32  `json:"averageFps"`
	AverageFrameTimeMs  float32  `json:"averageFrameTimeMs"`
	CurrentLoadPercent  float32  `json:"currentLoadPercent"`
}

type config struct {
	Url   string `json:"url"`
	Token string `json:"token"`
}

func getConfig(file string) (config, error) {
	var s config

	raw, fileErr := ioutil.ReadFile(file)
	if fileErr != nil {
		return s, fileErr
	}

	jsonErr := json.Unmarshal(raw, &s)
	if jsonErr != nil {
		return s, jsonErr
	}
	return s, nil
}

func main() {

	var version bool
	var help bool
	var configPath string
	var configUrl string
	var configToken string

	flag.BoolVar(&version, "version", false, "Print current version and exit.")
	flag.BoolVar(&help, "help", false, "Print help and exit.")
	flag.StringVar(&configPath, "config", "./config.json", "Path to configuration (json) file.")
	flag.StringVar(&configUrl, "url", "", "URL to read source data from.")
	flag.StringVar(&configToken, "token", "", "Authorization token.")
	flag.Parse()

	if version {
		fmt.Println("version: 0.1")
		return
	}
	if help {
		return
	}

	config, err := getConfig(configPath)
	if err != nil {
		log.Fatal(err)
		return
	}

	var url = config.Url
	var token = config.Token

	if configUrl != "" {
		url = configUrl
	}
	if configToken != "" {
		token = configToken
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "Token "+token)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
		return
	}

	var m []matchpoolInfo
	jsonErr := json.Unmarshal(body, &m)
	if jsonErr != nil {
		log.Fatal(jsonErr)
		return
	}
	//fmt.Printf("%#v", m[34])

	type NodeLabels struct {
		Instance string `json:"instance"`
	}
	type NodeEntry struct {
		Targets []string   `json:"targets"`
		Labels  NodeLabels `json:"labels"`
	}

	re := regexp.MustCompile(`http://(.+)/match-pool`)

	var list []NodeEntry
	for _, element := range m {
		// http://149.202.162.66:43675/match-pool
		var match = re.FindStringSubmatch(element.ApiUri)
		var target = match[1]

		var instance = element.Name
		instance = strings.Replace(instance, " ", "", -1)
		instance = strings.Replace(instance, ":", "-", -1)

		entry := NodeEntry{
			Targets: []string{target},
			Labels: NodeLabels{
				Instance: instance,
			},
		}
		list = append(list, entry)
	}

	//b2, err := json.Marshal(list)
	b2, err := json.MarshalIndent(list, "", "    ")
	if err != nil {
		panic(err)
	}
	if b2 != nil {
		fmt.Println(string(b2))
	}
}
