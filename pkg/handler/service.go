package handler

import (
	"errors"
	"fmt"
	"net"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

var availableIps []string

var index int
var totalIps int

var IpToHostMap map[string]string
var HostToIndexMap map[string]int

var logger *zap.SugaredLogger

type SelectedAlgorithm string

const (
	ALG_RoundRobin SelectedAlgorithm = "roundrobin"
	ALG_StickyIP   SelectedAlgorithm = "stickyip"
)

type T struct {
	Ips []string `yaml:"ips"`
}

func getCurrentIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		os.Stderr.WriteString("Oops: " + err.Error() + "\n")
		return "", err
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				// os.Stdout.WriteString(ipnet.IP.String() + "\n")
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", errors.New("error getting ip addr")
}

/*
   gets list of ips from yaml file and initializes maps
*/
func Init() error {
	dat, err := os.ReadFile("ips.yaml")
	if err != nil {
		return err
	}
	t := T{}

	err = yaml.Unmarshal(dat, &t)
	if err != nil {
		return err
	}

	fmt.Println("got yaml: ", t)
	availableIps = t.Ips
	totalIps = len(t.Ips)

	IpToHostMap = make(map[string]string)
	HostToIndexMap = make(map[string]int)

	prodLogger, _ := zap.NewProduction()
	defer prodLogger.Sync() // flushes buffer, if any
	logger = prodLogger.Sugar()

	logger.Infow("initialized yaml, maps, and logger")

	return nil
}

// back to index 0 when out of bounds (%)
func RoundRobin() int {
	return index % totalIps
}

/*
   stick IP to a particular host
   inorder to maintain sessions etc...

   maps:
    - ip addr to host
    - index of ip's host

*/
func StickyIP(selectedIp string) int {
	ip, _ := getCurrentIP()
	if IpToHostMap[ip] == "" {
		IpToHostMap[ip] = selectedIp
	}

	// no entry found
	if HostToIndexMap[selectedIp] < 0 {
		HostToIndexMap[selectedIp] = index
	}

	return HostToIndexMap[selectedIp]
}

func Balance(ctx *gin.Context) {
	// availableIps := []string{"http://localhost:5002", "http:localhost:5003"}

	// TODO: get availableIps from yaml and selectedIp from algorithm (rr or stickyIP)
	var selectedIp string
	selectedAlgorithm := ALG_RoundRobin
	logger.Infow("selected algorithm", "algorithm", selectedAlgorithm)

	// selectedAlgorithm := ALG_StickyIP

	if selectedAlgorithm == ALG_RoundRobin {
		currentIndex := RoundRobin()
		selectedIp = availableIps[currentIndex]
		index++
	} else {
		currentIndex := StickyIP(selectedIp)
		selectedIp = availableIps[currentIndex]
	}
	logger.Infow("selected host", "host", selectedIp)

	url, err := url.Parse(selectedIp)
	if err != nil {
		logger.Errorw("error parsing selected host", "error", err)
		// return err
	}

	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.ServeHTTP(ctx.Writer, ctx.Request)
}
