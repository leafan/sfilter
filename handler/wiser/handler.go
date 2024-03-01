package handler

type Handler struct {
	Wiser  *Wiser
	Hbpair *HBPair
	Hspair *HSPair
}

// deal or wiser 表示分析deal和wiser, 任意一个开启均表示打开 wiser 服务
func NewHandler(account, db string, debug bool, deal, wiser, hb, hs bool) *Handler {
	set := NewSetting(account, db, debug)

	hndl := &Handler{}

	if deal || wiser { // 是否开启 wiser searcher
		hndl.Wiser = &Wiser{
			set:          set,
			dealInspect:  deal,
			wiserInspect: wiser,
		}
	}

	if hb {
		hndl.Hbpair = &HBPair{
			Set: set,
		}
	}

	if hs {
		hndl.Hspair = &HSPair{
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
}
