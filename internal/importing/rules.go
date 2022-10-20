package importing

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strings"
)

const DS = "|"

func CreateBm(d [][]string, fields []string, prefixes string, rules []Rules) []bson.M {
	var res []bson.M
	for _, row := range d {
		m := bson.M{"type": prefixes}
		l := len(row)
		for index, field := range fields {
			if index >= l {
				break
			}

			m[field] = row[index]
		}
		res = append(res, m)
	}

	if len(rules) != 0 {
		res = ExecRules(res, rules)
	}

	return res
}

func ExecRules(d []bson.M, rules []Rules) []bson.M {
	if len(rules) != 0 {
		for i, v := range d {
			for _, p := range rules {
				d[i] = p.Exec(v)
			}
		}
	}

	return d
}

type Rules interface {
	Exec(m bson.M) bson.M
}

func NewRules(s ...string) []Rules {
	var p []Rules
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case "hid":
			p = append(p, &Hid{})
		case "splitting":
			p = append(p, &Splitting{})
		case "revHid":
			p = append(p, &RevHid{})
		case "revSplitting":
			p = append(p, &RevSplitting{})
		}
	}

	return p
}

type Hid struct {
}

func (h Hid) Exec(m bson.M) bson.M {
	m["_hid"] = HashString(m["pat_number"].(string), m["brand"].(string))
	return m
}

type Splitting struct {
}

func (s Splitting) Exec(m bson.M) bson.M {
	for field, value := range m {
		if vStr, ok := value.(string); ok {
			if strings.Index(vStr, DS) != -1 {
				m[field] = strings.Split(strings.Trim(vStr, DS), DS)
			}
		}
	}
	return m
}

type RevHid struct {
}

func (h RevHid) Exec(m bson.M) bson.M {
	m["id"] = m["_hid"]
	return m
}

type RevSplitting struct {
}

func (s RevSplitting) Exec(m bson.M) bson.M {
	for field, value := range m {
		if v, ok := value.(primitive.A); ok {
			m[field] = strings.Join(ToStringSlice(v), DS)
		}
	}
	return m
}
