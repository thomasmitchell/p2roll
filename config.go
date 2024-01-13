package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type GameConfig struct {
	Chars []CharConfig `yaml:"players"`
	path  string
}

type ProfBonus int

var (
	ProfUnknown   ProfBonus = -100
	ProfUntrained ProfBonus = 0
	ProfTrained   ProfBonus = 2
	ProfExpert    ProfBonus = 4
	ProfMaster    ProfBonus = 6
	ProfLegendary ProfBonus = 8
)

type CharConfig struct {
	Name          string     `yaml:"name"`
	Player        string     `yaml:"player"`
	Level         int        `yaml:"level"`
	Modifiers     ModConfig  `yaml:"modifiers"`
	Proficiencies ProfConfig `yaml:"proficiencies"`
	ArmorPenalty  int        `yaml:"armor_penalty"`
}

type ModConfig struct {
	Strength     int `yaml:"strength"`
	Dexterity    int `yaml:"dexterity"`
	Constitution int `yaml:"constitution"`
	Intellect    int `yaml:"intellect"`
	Wisdom       int `yaml:"wisdom"`
	Charisma     int `yaml:"charisma"`
}

type ProfConfig struct {
	Perception     ProfBonus      `yaml:"perception"`
	Stealth        ProfBonus      `yaml:"stealth"`
	Saves          SaveConfig     `yaml:"saves"`
	IdentifySkills IdentifyConfig `yaml:"identify"`
}

type SaveConfig struct {
	Reflex    ProfBonus `yaml:"reflex"`
	Fortitude ProfBonus `yaml:"fortitude"`
	Will      ProfBonus `yaml:"will"`
}

type IdentifyConfig struct {
	Arcana    ProfBonus `yaml:"arcana"`
	Nature    ProfBonus `yaml:"nature"`
	Occultism ProfBonus `yaml:"occultism"`
	Religion  ProfBonus `yaml:"religion"`
}

func NewGameConfig(filepath string) *GameConfig {
	return &GameConfig{path: filepath}
}

func (c *GameConfig) Write() error {
	c.sortChars()
	b, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshalling json: %s", err)
	}

	return os.WriteFile(c.path, b, 0640)
}

func LoadConfig(path string) (*GameConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return NewGameConfig(path), nil
		}

		return nil, fmt.Errorf("opening file for loading: %s", err)
	}

	ret := &GameConfig{}
	err = yaml.Unmarshal(b, ret)
	ret.path = path
	return ret, err
}

func (c *GameConfig) AllChars() []CharConfig {
	return c.Chars
}

func (c *GameConfig) AddChar(char *CharConfig) error {
	if _, err := c.CharByName(char.Name); err == nil {
		return fmt.Errorf("character with name already exists in game")
	}

	if _, err := c.CharByPlayerName(char.Player); err == nil {
		return fmt.Errorf("character with player name already exists in game")
	}

	c.Chars = append(c.Chars, *char)
	return nil
}

func (c *GameConfig) RemoveCharByName(name string) error {
	i, err := c.searchChars(nameSearchFn(name))
	if err != nil {
		return err
	}

	c.Chars[i], c.Chars[len(c.Chars)-1] = c.Chars[len(c.Chars)-1], c.Chars[i]
	c.Chars = c.Chars[:len(c.Chars)-1]
	c.sortChars()
	return nil
}

func (c *GameConfig) RemoveCharByPlayerName(playerName string) error {
	i, err := c.searchChars(playerSearchFn(playerName))
	if err != nil {
		return err
	}

	c.Chars[i], c.Chars[len(c.Chars)-1] = c.Chars[len(c.Chars)-1], c.Chars[i]
	c.Chars = c.Chars[:len(c.Chars)-1]
	return nil
}

func (c *GameConfig) sortChars() {
	sort.Slice(c.Chars, func(i, j int) bool { return c.Chars[i].Name < c.Chars[j].Name })
}

type searchFn func(char *CharConfig) bool

func nameSearchFn(name string) searchFn {
	return func(char *CharConfig) bool {
		return strings.EqualFold(char.Name, name)
	}
}

func playerSearchFn(player string) searchFn {
	return func(char *CharConfig) bool {
		return strings.EqualFold(char.Player, player)
	}
}

func (c *GameConfig) CharByName(name string) (*CharConfig, error) {
	i, err := c.searchChars(nameSearchFn(name))
	if err != nil {
		return nil, err
	}

	return &c.Chars[i], nil
}

func (c *GameConfig) CharByPlayerName(playerName string) (*CharConfig, error) {
	i, err := c.searchChars(playerSearchFn(playerName))
	if err != nil {
		return nil, err
	}

	return &c.Chars[i], nil
}

func (c *GameConfig) searchChars(matchFn searchFn) (int, error) {
	for i, char := range c.Chars {
		if !matchFn(&char) {
			continue
		}

		return i, nil
	}

	return -1, fmt.Errorf("character not found")

}

func calcMod(stat int, prof ProfBonus, level int) int {
	return stat + profBonus(prof, level)
}

func profBonus(prof ProfBonus, level int) int {
	if prof == ProfUntrained {
		return 0
	}

	return level + int(prof)
}

type ProfFn func() int

func (c *CharConfig) Perception() int {
	return calcMod(c.Modifiers.Wisdom, c.Proficiencies.Perception, c.Level)
}

func (c *CharConfig) Stealth() int {
	return calcMod(c.Modifiers.Dexterity, c.Proficiencies.Stealth, c.Level) - c.ArmorPenalty
}

func (c *CharConfig) ReflexSave() int {
	return calcMod(c.Modifiers.Dexterity, c.Proficiencies.Saves.Reflex, c.Level)
}

func (c *CharConfig) FortitudeSave() int {
	return calcMod(c.Modifiers.Constitution, c.Proficiencies.Saves.Fortitude, c.Level)
}

func (c *CharConfig) WillSave() int {
	return calcMod(c.Modifiers.Wisdom, c.Proficiencies.Saves.Will, c.Level)
}

func (c *CharConfig) GenericIdentify() int {
	return max(
		c.Arcana(),
		c.Nature(),
		c.Occultism(),
		c.Religion(),
	)
}

func (c *CharConfig) Arcana() int {
	return calcMod(c.Modifiers.Intellect, c.Proficiencies.IdentifySkills.Arcana, c.Level)
}

func (c *CharConfig) Nature() int {
	return calcMod(c.Modifiers.Wisdom, c.Proficiencies.IdentifySkills.Nature, c.Level)
}

func (c *CharConfig) Occultism() int {
	return calcMod(c.Modifiers.Intellect, c.Proficiencies.IdentifySkills.Occultism, c.Level)
}

func (c *CharConfig) Religion() int {
	return calcMod(c.Modifiers.Wisdom, c.Proficiencies.IdentifySkills.Religion, c.Level)
}
