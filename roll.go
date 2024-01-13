package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/jhunt/go-ansi"
)

type RollOpts struct {
	//metadata
	Name   string `name:"name" short:"n" help:"name of character to add" required:"true" xor:"spec"`
	Player string `name:"player" short:"p" help:"name of player to whom character belongs" required:"true" xor:"spec"`
	All    bool   `name:"all" short:"a" help:"roll for all characters" xor:"spec"`

	Target int `name:"target" short:"t" help:"target to match/beat"`
}

type RollCmd struct {
	Perception PerceptionCmd `cmd:""`
	Stealth    StealthCmd    `cmd:""`
	Reflex     ReflexCmd     `cmd:""`
	Fortitude  FortitudeCmd  `cmd:""`
	Will       WillCmd       `cmd:""`
	Identify   IdentifyCmd   `cmd:""`
	Arcana     ArcanaCmd     `cmd:""`
	Nature     NatureCmd     `cmd:""`
	Occultism  OccultismCmd  `cmd:""`
	Religion   ReligionCmd   `cmd:""`
	Flat       FlatCmd       `cmd:""`
}

type PerceptionCmd struct{ RollOpts }

func (c *PerceptionCmd) Run(conf *GameConfig) error { return runRoll(conf, c.RollOpts, Perception) }

type StealthCmd struct{ RollOpts }

func (c *StealthCmd) Run(conf *GameConfig) error { return runRoll(conf, c.RollOpts, Stealth) }

type ReflexCmd struct{ RollOpts }

func (c *ReflexCmd) Run(conf *GameConfig) error { return runRoll(conf, c.RollOpts, ReflexSave) }

type FortitudeCmd struct{ RollOpts }

func (c *FortitudeCmd) Run(conf *GameConfig) error { return runRoll(conf, c.RollOpts, FortitudeSave) }

type WillCmd struct{ RollOpts }

func (c *WillCmd) Run(conf *GameConfig) error { return runRoll(conf, c.RollOpts, WillSave) }

type IdentifyCmd struct{ RollOpts }

func (c *IdentifyCmd) Run(conf *GameConfig) error { return runRoll(conf, c.RollOpts, GenericIdentify) }

type ArcanaCmd struct{ RollOpts }

func (c *ArcanaCmd) Run(conf *GameConfig) error { return runRoll(conf, c.RollOpts, Arcana) }

type NatureCmd struct{ RollOpts }

func (c *NatureCmd) Run(conf *GameConfig) error { return runRoll(conf, c.RollOpts, Nature) }

type OccultismCmd struct{ RollOpts }

func (c *OccultismCmd) Run(conf *GameConfig) error { return runRoll(conf, c.RollOpts, Occultism) }

type ReligionCmd struct{ RollOpts }

func (c *ReligionCmd) Run(conf *GameConfig) error { return runRoll(conf, c.RollOpts, Religion) }

type FlatCmd struct{ RollOpts }

func (c *FlatCmd) Run(conf *GameConfig) error { return runRoll(conf, c.RollOpts, Flat) }

func runRoll(conf *GameConfig, c RollOpts, fn modFn) error {
	var chars []CharConfig
	if c.All {
		chars = conf.AllChars()
	} else {
		foundChar, err := getChar(conf, c.Name, c.Player)
		if err != nil {
			return err
		}

		chars = []CharConfig{*foundChar}
	}

	roll(chars, fn, c.Target)
	return nil
}

type SuccessDegree int

const (
	CriticalFailure SuccessDegree = iota
	Failure
	Success
	CriticalSuccess
)

func (s SuccessDegree) Icon() string {
	switch s {
	case CriticalFailure:
		return ansi.Sprintf(`💥`)
	case Failure:
		return ansi.Sprintf(`❌`)
	case Success:
		return ansi.Sprintf(`✅`)
	case CriticalSuccess:
		return ansi.Sprintf(`🌟`)
	}
	return ansi.Sprintf(`❔`)
}

func roll(chars []CharConfig, fn modFn, target int) {
	if target != 0 {
		ansi.Printf("@C{DC %d}\n", target)
	}
	//consider using a table?
	for _, char := range chars {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		d20 := rng.Intn(20) + 1
		mod := fn(&char)
		result := d20 + mod

		if target != 0 {
			var degree SuccessDegree
			if result <= target-10 {
				degree = CriticalFailure
			} else if result < target {
				degree = Failure
			} else if result >= target+10 {
				degree = CriticalSuccess
			} else {
				degree = Success
			}

			if d20 == 1 {
				degree = max(CriticalFailure, degree-1)
			} else if d20 == 20 {
				degree = min(CriticalSuccess, degree+1)
			}

			fmt.Printf("%s ", degree.Icon())
		}

		fmt.Printf("%s (%s)\t", char.Name, char.Player)
		switch d20 {
		case 1:
			ansi.Printf("@R{<%d>}", d20)
		case 20:
			ansi.Printf("@Y{<%d>}", d20)
		default:
			fmt.Printf("<%d>", d20)
		}

		fmt.Printf(" + %d = %d\n", mod, result)
	}
}

type modFn func(c *CharConfig) int

func Perception(c *CharConfig) int      { return c.Perception() }
func Stealth(c *CharConfig) int         { return c.Stealth() }
func ReflexSave(c *CharConfig) int      { return c.ReflexSave() }
func FortitudeSave(c *CharConfig) int   { return c.FortitudeSave() }
func WillSave(c *CharConfig) int        { return c.WillSave() }
func GenericIdentify(c *CharConfig) int { return c.GenericIdentify() }
func Arcana(c *CharConfig) int          { return c.Arcana() }
func Nature(c *CharConfig) int          { return c.Nature() }
func Occultism(c *CharConfig) int       { return c.Occultism() }
func Religion(c *CharConfig) int        { return c.Religion() }
func Flat(c *CharConfig) int            { return 0 }
