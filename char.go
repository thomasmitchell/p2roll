package main

import (
	"fmt"
	"os"
	"reflect"
)

type CharacterCmd struct {
	Add    CharacterAddCmd    `cmd:"" help:"add a new character to the game"`
	Remove CharacterRemoveCmd `cmd:"" help:"remove character from the game"`
	Edit   CharacterEditCmd   `cmd:"" help:"edit a character"`
	List   CharacterListCmd   `cmd:"" help:"list characters"`
}

type CharacterAddCmd struct {
	//metadata
	Name   string `name:"name" short:"n" help:"name of character to add" required:"true"`
	Player string `name:"player" short:"p" help:"name of player to whom character belongs" required:"true"`

	Level int `name:"level" help:"character level" default:"1"`
	//attribute modifiers
	Strength     int `name:"strength" help:"str mod" required:"true"`
	Dexterity    int `name:"dexterity" help:"dex mod" required:"true"`
	Constitution int `name:"constitution" help:"con mod" required:"true"`
	Intellect    int `name:"intelligence" help:"int mod" required:"true"`
	Wisdom       int `name:"wisdom" help:"wis mod" required:"true"`
	Charisma     int `name:"charisma" help:"cha mod" required:"true"`
	//proficiencies
	Perception   string `name:"perception" help:"perception prof for character" default:"U" enum:"U,T,E,M,L"`
	Stealth      string `name:"stealth" help:"stealth prof for character" default:"U" enum:"U,T,E,M,L"`
	Reflex       string `name:"reflex" help:"reflex save prof for character" default:"U" enum:"U,T,E,M,L"`
	Fortitude    string `name:"fortitude" help:"fortitude save prof for character" default:"U" enum:"U,T,E,M,L"`
	Will         string `name:"will" help:"will save prof for character" default:"U" enum:"U,T,E,M,L"`
	Arcana       string `name:"arcana" help:"arcana prof for character" default:"U" enum:"U,T,E,M,L"`
	Nature       string `name:"nature" help:"nature prof for character" default:"U" enum:"U,T,E,M,L"`
	Occultism    string `name:"occultism" help:"occultism prof for character" default:"U" enum:"U,T,E,M,L"`
	Religion     string `name:"religion" help:"religion prof for character" default:"U" enum:"U,T,E,M,L"`
	ArmorPenalty int    `name:"armor-penalty" help:"reduction to stealth from armor" default:"0"`
}

func (c *CharacterAddCmd) Run(conf *GameConfig) error {
	char := CharConfig{
		Name:   c.Name,
		Player: c.Player,
		Level:  c.Level,
		Modifiers: ModConfig{
			Strength:     c.Strength,
			Dexterity:    c.Dexterity,
			Constitution: c.Constitution,
			Intellect:    c.Intellect,
			Wisdom:       c.Wisdom,
			Charisma:     c.Charisma,
		},
		Proficiencies: ProfConfig{
			Perception: parseProf(c.Perception),
			Stealth:    parseProf(c.Stealth),
			Saves: SaveConfig{
				Reflex:    parseProf(c.Reflex),
				Fortitude: parseProf(c.Fortitude),
				Will:      parseProf(c.Will),
			},
			IdentifySkills: IdentifyConfig{
				Arcana:    parseProf(c.Arcana),
				Nature:    parseProf(c.Nature),
				Occultism: parseProf(c.Occultism),
				Religion:  parseProf(c.Religion),
			},
		},
		ArmorPenalty: c.ArmorPenalty,
	}

	err := conf.AddChar(&char)
	if err != nil {
		return err
	}

	err = conf.Write()
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "created character '%s' (%s)\n", char.Name, char.Player)
	return nil
}

type CharacterRemoveCmd struct {
	Name   string `name:"name" short:"n" help:"name of character to remove" required:"true" xor:"spec"`
	Player string `name:"player" short:"p" help:"name of player to whom character belongs to remove" required:"true" xor:"spec"`
}

func (c *CharacterRemoveCmd) Run(conf *GameConfig) error {
	var err error
	if c.Name != "" {
		err = conf.RemoveCharByName(c.Name)
	} else {
		err = conf.RemoveCharByPlayerName(c.Player)
	}

	if err != nil {
		return err
	}

	err = conf.Write()
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "removed character\n")
	return nil
}

type CharacterEditCmd struct {
	//metadata
	Name      string `name:"name" short:"n" help:"name of character to edit" required:"true" xor:"spec"`
	Player    string `name:"player" short:"p" help:"name of player to whom character belongs" required:"true" xor:"spec"`
	NewName   string `name:"new-name" help:"name to change to"`
	NewPlayer string `name:"new-player" help:"player name to change to"`

	Level int `name:"level" help:"character level"`

	//attribute modifiers
	Strength     int `name:"strength" help:"str mod" `
	Dexterity    int `name:"dexterity" help:"dex mod" `
	Constitution int `name:"constitution" help:"con mod" `
	Intellect    int `name:"intelligence" help:"int mod" `
	Wisdom       int `name:"wisdom" help:"wis mod" `
	Charisma     int `name:"charisma" help:"cha mod" `
	//proficiencies
	Perception string `name:"perception" help:"perception prof for character" default:"" enum:",U,T,E,M,L"`
	Stealth    string `name:"stealth" help:"stealth prof for character" default:""  enum:",U,T,E,M,L"`

	//saves
	Reflex    string `name:"reflex" help:"reflex save prof for character" default:"" enum:",U,T,E,M,L"`
	Fortitude string `name:"fortitude" help:"fortitude save prof for character" default:"" enum:",U,T,E,M,L"`
	Will      string `name:"will" help:"will save prof for character" default:"" enum:",U,T,E,M,L"`

	Arcana       string `name:"arcana" help:"arcana prof for character" default:"" enum:",U,T,E,M,L"`
	Nature       string `name:"nature" help:"nature prof for character" default:"" enum:",U,T,E,M,L"`
	Occultism    string `name:"occultism" help:"occultism prof for character" default:"" enum:",U,T,E,M,L"`
	Religion     string `name:"religion" help:"religion prof for character" default:"" enum:",U,T,E,M,L"`
	ArmorPenalty int    `name:"armor-penalty" help:"reduction to stealth from armor" default:"-100"`
}

func (c *CharacterEditCmd) Run(conf *GameConfig) error {
	char, err := getChar(conf, c.Name, c.Player)
	if err != nil {
		return err
	}

	setIfNotEmpty(&char.Name, c.NewName)
	setIfNotEmpty(&char.Player, c.NewPlayer)
	setIfNotEmpty(&char.Level, c.Level)

	setIfNotEmpty(&char.Modifiers.Strength, c.Strength)
	setIfNotEmpty(&char.Modifiers.Dexterity, c.Dexterity)
	setIfNotEmpty(&char.Modifiers.Constitution, c.Constitution)
	setIfNotEmpty(&char.Modifiers.Intellect, c.Intellect)
	setIfNotEmpty(&char.Modifiers.Wisdom, c.Wisdom)
	setIfNotEmpty(&char.Modifiers.Charisma, c.Charisma)

	setIfNotMatching(&char.ArmorPenalty, c.ArmorPenalty, -100)
	setIfNotEmpty(&char.Proficiencies.Saves.Reflex, parseProf(c.Reflex))
	setIfNotEmpty(&char.Proficiencies.Saves.Fortitude, parseProf(c.Fortitude))
	setIfNotEmpty(&char.Proficiencies.Saves.Will, parseProf(c.Will))

	setIfNotMatching(&char.Proficiencies.Perception, parseProf(c.Perception), ProfUnknown)
	setIfNotMatching(&char.Proficiencies.Stealth, parseProf(c.Stealth), ProfUnknown)

	setIfNotMatching(&char.Proficiencies.IdentifySkills.Arcana, parseProf(c.Arcana), ProfUnknown)
	setIfNotMatching(&char.Proficiencies.IdentifySkills.Nature, parseProf(c.Nature), ProfUnknown)
	setIfNotMatching(&char.Proficiencies.IdentifySkills.Occultism, parseProf(c.Occultism), ProfUnknown)
	setIfNotMatching(&char.Proficiencies.IdentifySkills.Religion, parseProf(c.Religion), ProfUnknown)

	err = conf.Write()
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "edited character '%s' (%s)\n", char.Name, char.Player)
	return nil
}

func setIfNotEmpty[T any](dest *T, val T) {
	if reflect.ValueOf(val).IsZero() {
		return
	}

	*dest = val
}

func setIfNotMatching[T comparable](dest *T, val T, comp T) {
	if val == comp {
		return
	}

	*dest = val
}

func parseProf(p string) ProfBonus {
	ret := ProfUnknown
	switch p {
	case "U":
		ret = ProfUntrained
	case "T":
		ret = ProfTrained
	case "E":
		ret = ProfExpert
	case "M":
		ret = ProfMaster
	case "L":
		ret = ProfLegendary
	}

	return ret
}

func getChar(conf *GameConfig, name, player string) (*CharConfig, error) {
	if name != "" {
		return conf.CharByName(name)
	}

	return conf.CharByPlayerName(player)
}

type CharacterListCmd struct{}

func (c *CharacterListCmd) Run(conf *GameConfig) error {
	chars := conf.AllChars()
	for _, char := range chars {
		fmt.Printf("%s (%s)\n", char.Name, char.Player)
	}

	return nil
}
