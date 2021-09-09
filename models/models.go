package models


const configFileName = "config.json"

type TraitType int

const (
	TraitNormal TraitType = iota
	TraitRare
	TraitSuperRare
)
