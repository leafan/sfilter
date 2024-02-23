package handler

import (
	"sfilter/utils"
	"time"
)

type HSPair struct {
	set *Setting
}

func (p *HSPair) Run() {
	interval := time.Duration(p.set.Config.HotPairCheckInterval) * time.Second
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
