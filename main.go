package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
    "math/rand"

	"github.com/florinsalasan/pokedex/internal/pokecache"
)

type cliCommand struct {
    name string
    description string
    callback func(*Config, *pokecache.Cache, []string) error
}

type Config struct {
    Next *string
    Previous *string
    Pokedex Pokedex
}

type Pokedex struct {
    Caught map[string]Pokemon
}

type Pokemon struct {
    BaseExp int `json:"base_experience"`
    Height int `json:"height"`
    Weight int `json:"weight"`
    Name string `json:"name"`
    Stats []Stat `json:"stats"`
    Types []Type `json:"types"`
}

type Stat struct {
    BaseStat int `json:"base_stat"`
    Stat struct {
        Name string `json:"name"`
    } `json:"stat"`
}

type Type struct {
    Slot int `json:"slot"`
    Type struct {
        Name string `json:"name"`
    } `json:"type"`
}

var commands = map[string]cliCommand {}

func initCommands() {

    commands["help"] = cliCommand {
        name: "help",
        description: "Displays a help message",
        callback: commandHelp,
    }
    commands["exit"] = cliCommand {
        name: "exit",
        description: "Exit the Pokedex",
        callback: commandExit,
    }
    commands["map"] = cliCommand {
        name: "map",
        description: "List 20 location areas in the game",
        callback: commandMap,
    }
    commands["mapb"] = cliCommand {
        name: "mapb",
        description: "List the previous 20 location areas in the game",
        callback: commandMapB,
    }
    commands["explore"] = cliCommand {
        name: "explore",
        description: "List out the pokemon that are found at the given location",
        callback: commandExplore,
    }
    commands["catch"] = cliCommand {
        name: "catch",
        description: "Attempt to catch a given pokemon",
        callback: commandCatch,
    }
    commands["inspect"] = cliCommand {
        name: "inspect",
        description: "Get info about a pokemon that you've caught",
        callback: commandInspect,
    }
    commands["pokedex"] = cliCommand {
        name: "pokedex",
        description: "Prints a list of pokemon that you've caught",
        callback: commandPokedex,
    }
}

func main() {
    config := &Config{}
    cache := pokecache.NewCache(1 * time.Minute)
    initCommands()
    scanner := bufio.NewScanner(os.Stdin)
    for {
        fmt.Print("Pokedex > ")
        _ = scanner.Scan()
        currToken := scanner.Text()
        cleanedInput := cleanInput(currToken)
        commands[cleanedInput[0]].callback(config, cache, cleanedInput[1:])
    }
}

func cleanInput(text string) []string {
    words := strings.Fields(text)
    loweredWords := []string{}
    for _, word := range words {
        loweredWord := strings.ToLower(word)
        loweredWords = append(loweredWords, loweredWord)
    }
    return loweredWords
}

func commandExit(config *Config, cache *pokecache.Cache, args []string) error {
    fmt.Println("Closing the Pokedex... Goodbye!")
    os.Exit(0)
    return nil
}

func commandHelp(config *Config, cache *pokecache.Cache, args []string) error {
    fmt.Println("Welcome to the Pokedex!")
    fmt.Println("Usage:")
    fmt.Println()
    for _, command := range commands {
        fmt.Printf("%s: %s\n", command.name, command.description)
    }
    return nil
}

func commandMap(config *Config, cache *pokecache.Cache, args []string) error {
    // Default if config next doesn't exist yet
    url := "https://pokeapi.co/api/v2/location-area/?offset=0&limit=20"

    if config.Next != nil {
        url = *config.Next
    }

    var body []byte
    var err error

    if cachedData, found := cache.Get(url); found {
        body = cachedData
    } else {
        res, err := http.Get(url)
        if err != nil {
            return fmt.Errorf("Failed to get a response from API call: %w", err)
        }

        defer res.Body.Close()

        body, err = io.ReadAll(res.Body)
        if err != nil {
            fmt.Printf("Failed to read response body: %v", err)
            return fmt.Errorf("Failed to read response body: %w", err)
        }
        cache.Add(url, body)
    }

    var locationResponse struct {
        Results []struct {
            Name string `json:"name"`
            URL string `json:"url"`
        } `json:"results"`
        Next *string `json:"next"`
        Previous *string `json:"previous"`
    }

    err = json.Unmarshal(body, &locationResponse)
    if err != nil {
        fmt.Printf("Failed to unmarshal body: %v\n", err)
        return fmt.Errorf("Failed to unmarshal body: %w", err)
    }

    config.Next = locationResponse.Next
    config.Previous = locationResponse.Previous

    for _, loc := range locationResponse.Results {
        fmt.Println(loc.Name)
    }

    return nil
}

func commandMapB(config *Config, cache *pokecache.Cache, args []string) error {
    // Default if config next doesn't exist yet
    url := "https://pokeapi.co/api/v2/location-area/?offset=0&limit=20"

    if config.Previous != nil {
        url = *config.Previous
    }

    if config.Previous == nil {
        fmt.Println("you're on the first page")
        return nil
    }

    var body []byte
    var err error

    if data, found := cache.Get(url); found {
        body = data
    } else {
        res, err := http.Get(url)
        if err != nil {
            return fmt.Errorf("Failed to get a response from API call: %w", err)
        }
        defer res.Body.Close()
        body, err = io.ReadAll(res.Body)
        if err != nil {
            return fmt.Errorf("Failed to read response body: %w", err)
        }
        cache.Add(url, body)
    }

    var locationResponse struct {
        Results []struct {
            Name string `json:"name"`
            URL string `json:"url"`
        } `json:"results"`
        Next *string `json:"next"`
        Previous *string `json:"previous"`
    }

    err = json.Unmarshal(body, &locationResponse)
    if err != nil {
        return fmt.Errorf("Failed to unmarshal body: %w", err)
    }

    config.Next = locationResponse.Next
    config.Previous = locationResponse.Previous

    for _, loc := range locationResponse.Results {
        fmt.Println(loc.Name)
    }

    return nil
}

func commandExplore(config *Config, cache *pokecache.Cache, args []string) error {
    if len(args) > 1 {
        fmt.Println("Please explore one location at a time")
        return nil
    }

    cleanedArgs := cleanInput(args[0])[0]
    url := fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%v", cleanedArgs)
    fmt.Printf("Exploring %v\n", cleanedArgs)

    var body []byte
    var err error

    if data, found := cache.Get(url); found {
        body = data
    } else {
        res, err := http.Get(url)
        if err != nil {
            return fmt.Errorf("Failed to get a response from API call: %w", err)
        }
        defer res.Body.Close()
        body, err = io.ReadAll(res.Body)
        if err != nil {
            return fmt.Errorf("Failed to read response body: %w", err)
        }
        cache.Add(url, body)
    }

    var Results struct {
        PokemonEncounters []struct {
            PokemonInfo struct {
                Name string `json:"name"`
                URL string `json:"url"`
            } `json:"pokemon"`
        } `json:"pokemon_encounters"`
    }

    err = json.Unmarshal(body, &Results)
    if err != nil {
        return fmt.Errorf("Failed to unmarshal body: %w", err)
    }

    for _, encounter := range Results.PokemonEncounters {
        fmt.Println(encounter.PokemonInfo.Name)
    }

    return nil
}

func commandCatch(config *Config, cache *pokecache.Cache, args []string) error {
    if len(args) > 1 {
        fmt.Println("Please explore one location at a time")
        return nil
    }

    cleanedArgs := cleanInput(args[0])[0]
    url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%v", cleanedArgs)
    fmt.Printf("Exploring %v\n", cleanedArgs)

    var body []byte
    var err error

    if data, found := cache.Get(url); found {
        body = data
    } else {
        res, err := http.Get(url)
        if err != nil {
            return fmt.Errorf("Failed to get a response from API call: %w", err)
        }
        defer res.Body.Close()
        body, err = io.ReadAll(res.Body)
        if err != nil {
            return fmt.Errorf("Failed to read response body: %w", err)
        }
        cache.Add(url, body)
    }

    var Results Pokemon

    err = json.Unmarshal(body, &Results)
    if err != nil {
        return fmt.Errorf("Failed to unmarshal body: %w", err)
    }

    fmt.Printf("Throwing a Pokeball at %v...\n", cleanedArgs)

    // gen a random int from 0 to base_exp, then if less than 50 pokemon is caught
    val := rand.Intn(Results.BaseExp)
    if val < 50 {
        fmt.Printf("%v was caught!\n", cleanedArgs)
        if config.Pokedex.Caught == nil {
            config.Pokedex.Caught = make(map[string]Pokemon)
        }
        config.Pokedex.Caught[Results.Name] = Results
    } else {
        fmt.Printf("%v escaped!\n", cleanedArgs)
    }

    return nil
}

func commandInspect(config *Config, cache *pokecache.Cache, args []string) error {
    if len(args) > 1 {
        fmt.Println("Please only inspect one pokemon at a time")
        return nil
    }

    targetPokemon := cleanInput(args[0])[0]
    mon, ok := config.Pokedex.Caught[targetPokemon]
    if !ok {
        fmt.Println("you have not caught that pokemon")
        return nil
    }

    fmt.Println("Name:", mon.Name)
    fmt.Println("Height:", mon.Height)
    fmt.Println("Weight:", mon.Weight)
    fmt.Println("Stats:")
    for _, stat := range mon.Stats {
        fmt.Printf("-%v: %v\n", stat.Stat.Name, stat.BaseStat)
    }
    fmt.Println("Types:")
    for _, type_ := range mon.Types {
        fmt.Printf("- %v\n", type_.Type.Name)
    }

    return nil
}

func commandPokedex(config *Config, cache *pokecache.Cache, args []string) error {

    if len(config.Pokedex.Caught) <= 0 {
        fmt.Println("You haven't caught any pokemon yet! Try using 'catch pikachu' or some other pokemon!")
        return nil
    }

    fmt.Println("Your Pokedex:")
    for _, mon := range config.Pokedex.Caught {
        fmt.Println("-", mon.Name)
    }

    return nil
}
