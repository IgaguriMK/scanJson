package scanner

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"sort"
)

type ValueType string

const (
	TypeList    ValueType = "list"
	TypeDict              = "dict"
	TypeBoolean           = "bool"
	TypeInt               = "int"
	TypeFloat             = "float"
	TypeString            = "string"
	TypeNil               = "<nil>"
)

type Value struct {
	Types      []ValueType
	IsNullable bool
	Child      *Value
	Variables  map[string]*Value
}

func (v *Value) HasType(t ValueType) bool {
	for _, tt := range v.Types {
		if tt == t {
			return true
		}
	}
	return false
}

func (v *Value) Print(w io.Writer) {
	v.printPrefixed(w, "")
}

func (v *Value) printPrefixed(w io.Writer, prefix string) {
	var nullable string

	if v.IsNullable {
		nullable = "nullable "
	}

	for _, t := range v.Types {
		switch t {
		case TypeList:
			v.printList(w, prefix)
		case TypeDict:
			v.printDict(w, prefix)
		case TypeNil:
			fmt.Fprintf(w, "%s<nil>\n", prefix)
		default:
			fmt.Fprintf(w, "%s(%s%s)\n", prefix, nullable, string(t))
		}
	}
}

func (v *Value) printList(w io.Writer, prefix string) {
	if v.Child == nil || len(v.Child.Types) == 0 {
		fmt.Fprintf(w, "[]\n")
		return
	}
	v.Child.printPrefixed(w, prefix+"[]")
}

func (v *Value) printDict(w io.Writer, prefix string) {
	keys := make([]string, 0)

	for k := range v.Variables {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		p := fmt.Sprintf("%s.%s", prefix, k)
		v.Variables[k].printPrefixed(w, p)
	}
}

func (v *Value) normalize() {
	var isFloat bool
	var nullable bool

	for _, t := range v.Types {
		switch t {
		case TypeFloat:
			isFloat = true
		case TypeNil:
			nullable = true
		}
	}

	if isFloat {
		v.removeType(TypeInt)
	}

	if nullable && len(v.Types) > 1 {
		v.removeType(TypeNil)
	}
}

func (v *Value) removeType(typ ValueType) {
	tt := make([]ValueType, 0)

	for _, t := range v.Types {
		if t != typ {
			tt = append(tt, t)
		}
	}

	v.Types = tt
}

func ParseValue(dec *json.Decoder) (*Value, error) {
	token, err := dec.Token()
	if err != nil {
		return nil, err
	}

	switch tval := token.(type) {
	case json.Delim:
		return parseListOrObject(tval, dec)
	case bool:
		return &Value{
			Types: []ValueType{TypeBoolean},
		}, nil
	case float64:
		if tval == math.Trunc(tval) {
			return &Value{
				Types: []ValueType{TypeInt},
			}, nil
		} else {
			return &Value{
				Types: []ValueType{TypeFloat},
			}, nil
		}
	case string:
		return &Value{
			Types: []ValueType{TypeString},
		}, nil
	case nil:
		return &Value{
			Types:      []ValueType{TypeNil},
			IsNullable: true,
		}, nil
	}

	return nil, fmt.Errorf("unknown token '%v'", token)
}

func parseListOrObject(token json.Delim, dec *json.Decoder) (*Value, error) {
	switch token {
	case '[':
		return parseList(dec)
	case '{':
		return parseObject(dec)
	}

	return nil, fmt.Errorf("invalid token %q", token.String())
}

func parseList(dec *json.Decoder) (*Value, error) {
	res := &Value{
		Types: []ValueType{TypeList},
	}

	if !dec.More() {
		err := parseEnd(dec, ']')
		if err != nil {
			return nil, err
		}
		return res, nil
	}

	child, err := ParseValue(dec)
	if err != nil {
		return nil, err
	}

	for dec.More() {
		c, err := ParseValue(dec)
		if err != nil {
			return nil, err
		}
		child = mergeValue(child, c)
	}

	res.Child = child

	err = parseEnd(dec, ']')
	if err != nil {
		return nil, err
	}

	return res, nil
}

func parseObject(dec *json.Decoder) (*Value, error) {
	res := &Value{
		Types: []ValueType{TypeDict},
	}

	dict := make(map[string]*Value)

	for dec.More() {
		nameTok, err := dec.Token()
		if err != nil {
			return nil, err
		}

		name, ok := nameTok.(string)
		if !ok {
			return nil, fmt.Errorf("unexpected token '%v', expected key", nameTok)
		}

		val, err := ParseValue(dec)
		if err != nil {
			return nil, err
		}

		dict[name] = val
	}

	err := parseEnd(dec, '}')
	if err != nil {
		return nil, err
	}

	res.Variables = dict
	return res, nil
}

func parseEnd(dec *json.Decoder, r rune) error {
	token, err := dec.Token()
	if err != nil {
		return err
	}

	if d, ok := token.(json.Delim); ok {
		if rune(d) == r {
			return nil
		}
		return fmt.Errorf("invalid Delim %q", d.String())
	}
	return fmt.Errorf("invalid token '%v'", token)
}

func mergeValue(x, y *Value) *Value {
	if x == nil {
		return y
	}
	if y == nil {
		return x
	}

	types := make([]ValueType, 0)
	types = append(types, x.Types...)

	z := &Value{
		Types:      types,
		IsNullable: x.IsNullable || y.IsNullable,
		Child:      mergeValue(x.Child, y.Child),
		Variables:  make(map[string]*Value),
	}

	for _, t := range y.Types {
		if !z.HasType(t) {
			z.Types = append(z.Types, t)
		}
	}

	for k, v := range x.Variables {
		z.Variables[k] = v
	}

	for k, v := range y.Variables {
		if old, ok := z.Variables[k]; ok {
			z.Variables[k] = mergeValue(old, v)
		} else {
			z.Variables[k] = v
		}
	}

	z.normalize()

	return z
}
