package handler

import (
	"sfilter/utils"
	"time"
)

type HSPair struct {
	Set *Setting
}

func (p *HSPair) Run() {
	interval := time.Duration(p.Set.Config.HotPairCheckInterval) * time.Second
	timer := time.NewTicker(interval)
	defer timer.Stop()

	p.HotSubnewPairCheck()
	for range timer.C {
		p.HotSubnewPairCheck()
	}
}

func (p *HSPair) HotSubnewPairCheck() {
	utils.Infof("[ HotSubnewPairCheck ] running now....")
}
