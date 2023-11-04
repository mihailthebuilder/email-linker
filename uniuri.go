package main

import "github.com/dchest/uniuri"

type UniUriStringGenerator struct {
	Length int
}

func newUniUriStringGenerator() *UniUriStringGenerator {
	return &UniUriStringGenerator{Length: 20}
}

func (u *UniUriStringGenerator) Generate() string {
	return uniuri.NewLen(u.Length)
}
