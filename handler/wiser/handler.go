package handler

type Handler struct {
	wiser *Wiser
	hpair *HPair
}

// deal or wiser 表示分析deal和wiser, 任意一个开启均表示打开 wiser 服务
func NewHandler(account, db string, debug bool, deal, wiser bool) *Handler {
	set := NewSetting(account, db, debug)

	hndl := &Handler{}

	if deal || wiser { // 是否开启 wiser searcher
		hndl.wiser = &Wiser{
			set:          set,
			dealInspect:  deal,
			wiserInspect: wiser,
		}
	}

	hndl.hpair = &HPair{
		set: set,
	}

	return hndl
}

func (h *Handler) Run() {
	if h.wiser != nil {
		h.wiser.set.doWiserPreparation()
		h.wiser.Run()
	}

	h.hpair.Run()
}
