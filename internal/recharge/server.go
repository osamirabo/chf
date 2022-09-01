package recharge

import (
	"encoding/binary"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	chf_context "github.com/free5gc/chf/internal/context"
	"github.com/free5gc/chf/internal/logger"
	"github.com/free5gc/chf/internal/sbi/producer"
	"github.com/fsnotify/fsnotify"
)

var ServerStartTime time.Time

type Server struct {
	// Watcher watches a set of files, delivering events to a channel.
	QuotaWatcher *fsnotify.Watcher
}

func OpenServer() *Server {
	s := new(Server)
	go s.Serve()
	logger.RechargingLog.Infof("Recharging server started")

	return s
}

func (s *Server) Serve() {
	watcher, err := fsnotify.NewWatcher()
	ctx := chf_context.CHF_Self()
	ctx.QuotaWatcher = &watcher

	if err != nil {
		logger.RechargingLog.Warnf("create NewWatcher err: %+v", err)
	}

	go func() {
		for {
			defer func() {
				logger.RechargingLog.Infof("Recharging server stopped")
			}()
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op == fsnotify.Write {
					data, err := ioutil.ReadFile(event.Name)
					logger.RechargingLog.Warnf("Notify quota file change")
					if err != nil {
						logger.RechargingLog.Warnf("Read Events err: %+v", err)
					}

					qouta := binary.BigEndian.Uint32(data[0:4])

					// fileName = /tmp/quota/:ratinggroup.quota
					rg := strings.Split(event.Name, "/")[3]
					rg = strings.Split(rg, ".")[0]
					ratinggroup, _ := strconv.Atoi(rg)
					producer.NotifyRecharge(qouta, int32(ratinggroup))
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logger.RechargingLog.Warnf("watcher Events err: %+v", err)
			}
		}
	}()

	ServerStartTime = time.Now()
}
