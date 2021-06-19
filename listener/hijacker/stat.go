package hijacker

import (
	"bufio"
	"fmt"
	C "github.com/Dreamacro/clash/constant"
	"github.com/Dreamacro/clash/log"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

var statMap = make(map[string]int64)
var fileName = "stat.csv"
var lockAddNum sync.Mutex
var lockSaveStat sync.Mutex

// add stat
func AddStat(metadata *C.Metadata, remoteConn C.Conn) {
	key := remoteConn.Chains()[0] + "," + metadata.String()
	addNum(key, 1)
}
func addNum(key string, value int64) {
	lockAddNum.Lock()
	defer lockAddNum.Unlock()
	if v, ok := statMap[key]; ok {
		statMap[key] = v + value
	} else {
		statMap[key] = value
	}
	if len(statMap)%1024 == 0 {
		go SaveStat()
	}
}

func SaveStat() {
	lockSaveStat.Lock()
	defer lockSaveStat.Unlock()
	// open file
	filePath := filepath.Join(C.Path.HomeDir(), fileName)
	log.Infoln("SaveStat:" + filePath)
	fi, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Errorln("SaveStat file error", err)
		return
	}
	defer fi.Close()
	// read file
	br := bufio.NewReader(fi)
	for {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		line := string(a)
		words := strings.Split(line, ",")
		if len(words) >= 2 {
			v, _ := strconv.ParseInt(words[len(words)-1], 10, 64)
			addNum(strings.Join(words[0:(len(words)-1)], ","), v)
		}
	}

	// write file
	fi.Seek(0, io.SeekStart)
	fi.Truncate(0)
	w := bufio.NewWriter(fi)
	for k, v := range statMap {
		line := fmt.Sprintf("%s,%d", k, v)
		fmt.Fprintln(w, line)
	}
	w.Flush()
	// clear map
	statMap = make(map[string]int64)
}
