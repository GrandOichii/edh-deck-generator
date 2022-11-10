package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/GrandOichii/colorwrapper"
	"github.com/GrandOichii/mtgsdk"
)

var (
	cardName     string
	outPath      string
	logF         bool
	forceOffline bool
)

type printFlag struct {
	Use bool

	Name  string
	Usage string
	Do    func(d *mtgsdk.Deck) error
}

var printFlags = []*printFlag{
	{
		Name:  "stats",
		Usage: "(optional) Prints out the statistics of the deck",
		Do: func(d *mtgsdk.Deck) error {
			stats, err := d.GetStats()
			if err != nil {
				return err
			}
			return stats.Print()
		},
	},
	{
		Name:  "print",
		Usage: "(optional) Prints out the deck",
		Do: func(d *mtgsdk.Deck) error {
			err := d.Print()
			return err
		},
	},
}

type deckGenFlag struct {
	Value int

	Name    string
	Usage   string
	Default int
}

var deckGenFlags = []*deckGenFlag{
	{
		Name:    mtgsdk.DeckGenBoardWipeCountKey,
		Usage:   "(optional) The preferred amount of board wipes in the deck",
		Default: mtgsdk.BoardWipeCountDefault,
	},
	{
		Name:    mtgsdk.DeckGenCardDrawCount,
		Usage:   "(optional) The preffered amount of card draw in the deck",
		Default: mtgsdk.CardDrawCountDefault,
	},
	{
		Name:    mtgsdk.DeckGenLandCount,
		Usage:   "(optional) The preffered amount of lands in the deck",
		Default: mtgsdk.LandCountDefault,
	},
	{
		Name:    mtgsdk.DeckGenRampCountKey,
		Usage:   "(optional) The preffered amount of ramp in the deck",
		Default: mtgsdk.RampCountDefault,
	},
	{
		Name:    mtgsdk.DeckGenRemovalCount,
		Usage:   "(optional) The preffered amount of removal in the deck",
		Default: mtgsdk.RemovalCountDefault,
	},
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func init() {
	flag.StringVar(&cardName, "card", "", "The name of the commander")
	flag.StringVar(&outPath, "out", "", "The resulting deck file name")
	flag.BoolVar(&logF, "log", false, "(optional) Log the messages")
	flag.BoolVar(&forceOffline, "offline", false, "(optional) Force to use the local data")

	for _, pflag := range printFlags {
		flag.BoolVar(&pflag.Use, pflag.Name, false, pflag.Usage)
	}
	for _, dgflag := range deckGenFlags {
		flag.IntVar(&dgflag.Value, dgflag.Name, dgflag.Default, dgflag.Usage)
	}
}

func main() {
	// mtgsdk.UpdateBulkData()
	flag.Parse()
	if cardName == "" {
		flag.PrintDefaults()
		return
	}
	if outPath == "" {
		flag.PrintDefaults()
		return
	}
	if !logF {
		log.SetOutput(ioutil.Discard)
	}
	cards, err := mtgsdk.GetCards(map[string]string{mtgsdk.CardNameKey: cardName})
	if len(cards) == 0 {
		fmt.Printf("Couldn't find card with name: %s, check your internat connection", cardName)
		return
	}
	checkErr(err)
	ci := 0
	for {
		if cards[ci].IsCreature() && cards[ci].IsLegendary() {
			break
		}
		ci++
		if ci == len(cards) {
			fmt.Printf("No commander with name %s\n", cardName)
		}
	}
	fmt.Printf("Generating deck for %s...\n", cards[ci].Name)
	params := map[string]int{}
	for _, dgflag := range deckGenFlags {
		params[dgflag.Name] = dgflag.Value
	}
	deck, err := cards[ci].GenerateCommanderDeck(params)
	checkErr(err)
	fmt.Println("Generated!")
	err = deck.Save(outPath)
	checkErr(err)
	fmt.Printf("Deck saved to %s\n", outPath)
	for _, pflag := range printFlags {
		if pflag.Use {
			colored, err := colorwrapper.GetColored("normal-red", pflag.Name)
			checkErr(err)
			fmt.Println("\t" + colored)
			err = pflag.Do(deck)
			checkErr(err)
		}
	}
}
