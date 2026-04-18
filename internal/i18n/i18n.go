package i18n

import "strings"

type Lang int

const (
	En Lang = iota
	Ru
)

func Parse(s string) Lang {
	parts := strings.SplitN(strings.ToLower(s), "-", 2)
	switch parts[0] {
	case "en":
		return En
	case "ru":
		return Ru
	default:
		return En
	}
}

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

func (l Lang) Tag() string {
	switch l {
	case En:
		return "en-US"
	case Ru:
		return "ru-RU"
	default:
		return "en-US"
	}
}

func NewI18n(l Lang) State {
	if l == En {
		return State{
			Cur:     En,
			Strings: enStr,
		}
	}
	return State{
		Cur:     Ru,
		Strings: ruStr,
	}
}

type State struct {
	Cur Lang
	Strings
}

func (i *State) SetLang(lang Lang) {
	if lang == En {
		i.Cur = En
		i.Strings = enStr
	} else {
		i.Cur = Ru
		i.Strings = ruStr
	}
}
