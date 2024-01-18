package handler

type Handler struct {
	wiser *Wiser
}

func NewHandler(account, db string, debug bool, wiser bool) *Handler {
	set := NewSetting(account, db, debug)

	hndl := &Handler{}

	if wiser { // 是否开启 wiser searcher
		hndl.wiser = &Wiser{
			set: set,
		}
	}

	return hndl
}

func (h *Handler) Run() {
	if h.wiser != nil {
		h.wiser.Run()
	}
}
