package models

type Team struct {
	Id   int
	Name string
}

type TeamStatsPR struct {
	Name     string
	TotalPr  int
	OpenPr   int
	MergedPr int
}
