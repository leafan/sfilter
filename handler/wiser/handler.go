package handler

type Handler struct {
	Wiser   *Wiser
	Hbpair  *HBPair
	Hspair  *HSPair
	Hnpair  *HNPair
	HspPair *HSPPair
}

// deal or wiser 表示分析deal和wiser, 任意一个开启均表示打开 wiser 服务
func NewHandler(account, db string, debug bool, deal, wiser bool, hx string) *Handler {
	set := NewSetting(account, db, debug)

	hndl := &Handler{}

	if deal || wiser { // 是否开启 wiser searcher
		hndl.Wiser = &Wiser{
			set:          set,
			dealInspect:  deal,
			wiserInspect: wiser,
		}
	}

	if hx == "hb" {
		hndl.Hbpair = &HBPair{
			Set: set,
		}
	}

	if hx == "hs" {
		hndl.Hspair = &HSPair{
			Set: set,
		}
	}

	if hx == "hn" {
		hndl.Hnpair = &HNPair{
			Set: set,
		}
	}

	if hx == "hsp" {
		hndl.HspPair = &HSPPair{
			Set: set,
		}
	}

	return hndl
}

func (h *Handler) Run() {
	if h.Wiser != nil {
		h.Wiser.set.doWiserPreparation()
		h.Wiser.Run()
	}

	if h.Hbpair != nil {
		h.Hbpair.Run()
	}

	if h.Hspair != nil {
		h.Hspair.Run()
	}

	if h.Hnpair != nil {
		h.Hnpair.Run()
	}

	if h.HspPair != nil {
		h.HspPair.Run()
	}
}
