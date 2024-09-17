package main

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"math"
	"math/rand"
	"os"
	"strings"

	"github.com/google/uuid"
)

var (
	names    []Name
	families []Family
)

type (
	Sex            int
	GenerationType int
)

const (
	GENERATIONS       = 10
	NAME_COUNT        = 200
	FAMILY_COUNT      = 100
	SEED              = 0.2
	COMPETITION_DELTA = 1.5
)

const (
	MALE = iota
	FEMALE
)

const (
	DECLINE = iota - 1
	STABLE
	GROWTH
)

type Family struct {
	Id         string
	Name       string
	Generation int
}

type Person struct {
	Mother     *Person
	Father     *Person
	Generation *Generation
	Id         string
	Name       string
	Family     Family
	Sex        Sex
}
type People []Person

func (people *People) Get(id string) Person {
	for _, p := range *people {
		if p.Id == id {
			return p
		}
	}

	return Person{}
}

func (s Sex) String() string {
	switch s {
	case 0:
		return "MALE"
	case 1:
		return "FEMALE"
	default:
		return "UNKNOWN"
	}
}

type Couple struct {
	Mother *Person
	Father *Person
	Family Family
	Lambda int
}
type Couples []Couple

type Generation struct {
	Competition float64
	Type        GenerationType
	Direction   GenerationType
	Id          int
}

func (gt GenerationType) String() string {
	switch gt {
	case -1:
		return "DECLINE"
	case 0:
		return "STABLE"
	case 1:
		return "GROWTH"
	default:
		return "UNKNOWN"
	}
}

type History []Generation

type Name struct {
	Value string
	Sex   Sex
}

func read_first_names() {
	file, err := os.Open("first_names.csv")
	if err != nil {
		fmt.Printf("failed to open first_names.csv: %v\n", err)
		panic(1)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("failed to read first_names.csv: %v\n", err)
		panic(1)
	}
	rows := strings.Split(string(content), "\n")

	for _, row := range rows {
		if len(row) == 0 {
			break
		}

		values := strings.Split(row, ",")

		var sex Sex
		if values[1] == "MALE" {
			sex = MALE
		} else {
			sex = FEMALE
		}

		names = append(names, Name{values[0], sex})
	}
}

func read_last_names() {
	file, err := os.Open("last_names.csv")
	if err != nil {
		fmt.Printf("failed to open last_names.csv: %v\n", err)
		panic(1)
	}
	content, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("failed to read last_names.csv: %v\n", err)
		panic(1)
	}
	rows := strings.Split(string(content), "\n")

	for _, row := range rows {
		if len(row) == 0 {
			break
		}

		families = append(families, Family{uuid.NewString(), row, 0})
	}
}

func GenerateFirstName(s Sex) string {
	idx := rand.Intn(NAME_COUNT)
	if s == FEMALE {
		idx += NAME_COUNT
	}
	return names[idx].Value
}

func GenerateFamily(g *Generation) Family {
	idx := rand.Intn(FAMILY_COUNT)
	return Family{uuid.NewString(), families[idx].Name, g.Id}
}

func GenerateLambda(g *Generation) int {
	l := 0
	switch g.Type {
	case DECLINE:
		l = 1
	case STABLE:
		l = 2
	case GROWTH:
		l = 4
	}

	return l
}

func create_adam(g Generation) Person {
	father := Person{}
	mother := Person{}
	family := GenerateFamily(&g)

	return Person{
		&mother,
		&father,
		&g,
		uuid.NewString(),
		GenerateFirstName(MALE),
		family,
		MALE,
	}
}

func create_eve(g Generation) Person {
	father := Person{}
	mother := Person{}
	family := GenerateFamily(&g)

	return Person{
		&mother,
		&father,
		&g,
		uuid.NewString(),
		GenerateFirstName(FEMALE),
		family,
		FEMALE,
	}
}

func generatePoissonRV(lambda int) int {
	var n int

	for s := 0.0; s < 1; {
		u := rand.Float64()
		e := -math.Log(u) / float64(lambda)
		n += 1
		s += e
	}

	return n
}

func generateChildren(g *Generation, c Couple) People {
	var children People

	n := generatePoissonRV(c.Lambda)
	for i := 0; i < n; i++ {
		sex := Sex(math.Round(rand.Float64()))
		child := Person{
			c.Mother,
			c.Father,
			g,
			uuid.NewString(),
			GenerateFirstName(sex),
			c.Family,
			sex,
		}
		children = append(children, child)
	}

	return children
}

func simulate(gen Generation, males People, females People, everybody *People) (People, People) {
	var couples Couples
	var maxCouples int
	malesCount, femalesCount := len(males), len(females)
	if malesCount <= femalesCount {
		maxCouples = malesCount
	} else {
		maxCouples = femalesCount
	}

	rand.Shuffle(len(males), func(i int, j int) {
		males[i], males[j] = males[j], males[i]
	})
	rand.Shuffle(len(females), func(i int, j int) {
		females[i], females[j] = females[j], females[i]
	})
	for i := 0; i < maxCouples; i++ {
		couple := Couple{
			&females[i],
			&males[i],
			males[i].Family,
			GenerateLambda(&gen),
		}
		couples = append(couples, couple)
	}

	var m, f People
	for _, c := range couples {
		children := generateChildren(&gen, c)
		for _, child := range children {
			*everybody = append(*everybody, child)
			switch child.Sex {
			case MALE:
				m = append(m, child)
			case FEMALE:
				f = append(f, child)
			}
		}
	}

	return m, f
}

func main() {
	read_first_names()
	read_last_names()

	generation := Generation{SEED, GROWTH, GROWTH, 0}
	adam := create_adam(generation)
	eve := create_eve(generation)

	history := History{generation}
	males := People{adam}
	females := People{eve}
	everybody := People{adam, eve}

	for i := 0; i < GENERATIONS; i++ {
		generation.Id += 1
		males, females = simulate(generation, males, females, &everybody)

		var comp float64
		var dir GenerationType
		switch generation.Direction {
		case DECLINE:
			comp = math.Pow(generation.Competition, 1/COMPETITION_DELTA)
		case GROWTH:
			comp = math.Pow(generation.Competition, COMPETITION_DELTA)
		}

		var genType GenerationType
		if comp > 0.6 {
			genType = DECLINE
			dir = GROWTH
		} else if comp < 0.2 {
			genType = GROWTH
			dir = DECLINE
		} else {
			genType = STABLE
			dir = generation.Direction // previous direction
		}
		generation = Generation{comp, genType, dir, generation.Id}
		history = append(history, generation)

		fmt.Printf(
			"GENERATION: %d - TYPE: %s\nDir: %s - Competition Factor: %f\n",
			generation.Id,
			generation.Type,
			generation.Direction,
			generation.Competition,
		)
		fmt.Printf(
			"Population: %d\nActive: %d\n\n",
			len(everybody),
			len(males)+len(females),
		)

		if len(males)+len(females) == 0 {
			fmt.Println("Family has died out :(")
			break
		}
	}

	os.Remove("../imports")
	err := os.Mkdir("../imports", fs.FileMode(fs.ModeDir))

	// nodes
	fmt.Println("writing people.csv...")
	people, err := os.Create("../imports/people.csv")
	if err != nil {
		fmt.Printf("failed to create people.csv: %v\n", err)
		panic(1)
	}
	defer people.Close()

	buf := bufio.NewWriter(people)
	fmt.Fprintf(buf, "personId:ID,name:STRING,sex:STRING,motherId:STRING,fatherId:STRING,familyId:STRING,:LABEL\n")
	for _, p := range everybody {
		fmt.Fprintf(buf, "%s,%s,%s,%s,%s,%s,%s\n", p.Id, p.Name, p.Sex, p.Mother.Id, p.Father.Id, p.Family.Id, "Person")
		buf.Flush()
	}

	fmt.Println("writing generations.csv...")
	generations, err := os.Create("../imports/generations.csv")
	if err != nil {
		fmt.Printf("failed to create generations.csv: %v\n", err)
		panic(1)
	}
	defer generations.Close()

	buf = bufio.NewWriter(generations)
	fmt.Fprintf(buf, "generationId:ID,Type:STRING,Direction:STRING,competitionFactor:FLOAT,:LABEL\n")
	for _, g := range history {
		fmt.Fprintf(buf, "%d,%s,%s,%f,%s\n", g.Id, g.Type, g.Direction, g.Competition, "Generation")
		buf.Flush()
	}

	// edges
	fmt.Println("writing edges.csv...")
	edges, err := os.Create("../imports/edges.csv")
	if err != nil {
		fmt.Printf("failed to create edges.csv: %v\n", err)
		panic(1)
	}
	defer edges.Close()

	buf = bufio.NewWriter(edges)
	fmt.Fprintf(buf, ":START_ID,:END_ID,:TYPE\n")
	for _, p := range everybody {
		if p.Father.Id != "" {
			fmt.Fprintf(buf, "%s,%s,%s\n", p.Id, p.Father.Id, "FATHER")
			fmt.Fprintf(buf, "%s,%s,%s\n", p.Id, p.Father.Id, "PARENT")
			fmt.Fprintf(buf, "%s,%s,%s\n", p.Father.Id, p.Id, "CHILD")
		}
		if p.Mother.Id != "" {
			fmt.Fprintf(buf, "%s,%s,%s\n", p.Id, p.Mother.Id, "MOTHER")
			fmt.Fprintf(buf, "%s,%s,%s\n", p.Id, p.Mother.Id, "PARENT")
			fmt.Fprintf(buf, "%s,%s,%s\n", p.Mother.Id, p.Id, "CHILD")
		}
		fmt.Fprintf(buf, "%s,%d,%s\n", p.Id, p.Generation.Id, "IN_GENERATION")
		buf.Flush()
	}
}
