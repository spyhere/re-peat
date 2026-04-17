package i18n

type Lang int

const (
	En Lang = iota
	Ru
)

func (l Lang) String() string {
	switch l {
	case En:
		return "EN"
	case Ru:
		return "RU"
	default:
		panic("unreachable")
	}
}

func NewI18n() State {
	return State{
		curLang: En,
		Strings: enStr,
	}
}

type State struct {
	curLang Lang
	Strings
}

func (i *State) SetLang(lang Lang) {
	if lang == En {
		i.curLang = En
		i.Strings = enStr
	} else {
		i.curLang = Ru
		i.Strings = ruStr
	}
}
