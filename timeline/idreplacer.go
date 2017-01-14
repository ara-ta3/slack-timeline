package timeline

import (
	"fmt"
	"strings"
)

type IDReplacer struct {
	replacer  *strings.Replacer
	IDToNames []string
}

func (r IDReplacer) Replace(s string) string {
	return r.replacer.Replace(s)
}

func NewIDReplacerFactory(r UserRepository) IDReplacerFactory {
	return IDReplacerFactory{
		userRepository: r,
	}
}

type IDReplacerFactory struct {
	userRepository UserRepository
}

func (f IDReplacerFactory) NewReplacer() (IDReplacer, error) {
	us, e := f.userRepository.GetAll()
	if e != nil {
		return IDReplacer{}, e
	}

	ns := make([]string, 2*len(us))
	for i := 0; i < len(us); i++ {
		u := us[i]
		ns[2*i] = fmt.Sprintf("<@%s>", u.ID)
		ns[2*i+1] = fmt.Sprintf("@%s", u.Name)
	}
	ms := []string{
		"<!here|@here>",
		"@here",
		"<!channel>",
		"@channel",
	}
	replace := append(ns, ms...)

	r := strings.NewReplacer(replace...)
	return IDReplacer{
		replacer:  r,
		IDToNames: ns,
	}, nil
}
