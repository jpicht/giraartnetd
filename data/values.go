package data

type ValueBody struct {
	Values ValueList `json:"values"`
}

type ValueList []Value

type Value struct {
	UID   string `json:"uid"`
	Value string `json:"value"`
}
