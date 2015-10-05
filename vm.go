package models

type VM struct {
	Name string
	Vcores int
	Owner User


	Mem uint64
	MemUsed uint64
	MemUsedProzent float64

	Cluster Cluster
	Status string
	OSName string





}
