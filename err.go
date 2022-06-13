package main

type State uint8

const (
	PARSER_ERROR State = iota
	PARSER_SUCCESS
)
